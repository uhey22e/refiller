package refiller

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/jackc/pgtype"
)

type fieldDef struct {
	FieldIndex   int
	FieldType    reflect.Type
	ParserFunc   parserFunc
	DefaultValue string
	*parseOption
}

type parseOption struct {
	TimeFormat string
}

type parserFunc func(s string, opt *parseOption) (interface{}, error)

var (
	csvColNameTagKey     = "scan"
	defaultValueTagKey   = "scanDefault"
	scanTimeFormatTagKey = "scanTimeFormat"
)

var builtInTypeParsers = map[reflect.Kind]parserFunc{
	reflect.String: func(s string, opt *parseOption) (interface{}, error) {
		return s, nil
	},
	reflect.Int: func(s string, opt *parseOption) (interface{}, error) {
		x, err := strconv.ParseInt(s, 10, 64)
		return int(x), err
	},
	reflect.Float32: func(s string, opt *parseOption) (interface{}, error) {
		x, err := strconv.ParseFloat(s, 64)
		return float32(x), err
	},
}

var structTypeParsers = map[reflect.Type]parserFunc{
	reflect.TypeOf(pgtype.UUID{}): func(s string, opt *parseOption) (interface{}, error) {
		x := pgtype.UUID{}
		err := x.Set(s)
		return x, err
	},
	reflect.TypeOf(time.Time{}): func(s string, opt *parseOption) (interface{}, error) {
		return time.Parse(opt.TimeFormat, s)
	},
}

func ParseRow(dest interface{}, header, row []string) error {
	hm := headerAsMap(header)
	defs := inspectStruct(dest)
	err := parseRowWithDefinitions(dest, row, hm, defs)
	if err != nil {
		return err
	}
	return nil
}

func parseRowWithDefinitions(dest interface{}, row []string, header map[string]int, defs map[string]*fieldDef) error {
	rv := reflect.ValueOf(dest).Elem()
	if dest == nil {
		rv = reflect.New(reflect.TypeOf(dest))
	}
	for name, def := range defs {
		var v interface{}
		var err error
		f := rv.Field(def.FieldIndex)
		if !f.CanSet() {
			continue
		}

		if ci, ok := header[name]; ok {
			v, err = def.ParserFunc(row[ci], def.parseOption)
			if err == nil {
				f.Set(reflect.ValueOf(v))
				continue
			}
		}
		if s := def.DefaultValue; s != "" {
			v, err = def.ParserFunc(s, def.parseOption)
			if err == nil {
				f.Set(reflect.ValueOf(v))
				continue
			}
			panic(fmt.Errorf("invalid default value: %s as %v", s, f.Type()))
		}
		if err == nil {
			err = fmt.Errorf("missing column: %s", name)
		}
		return err
	}
	return nil
}

func inspectStruct(dest interface{}) map[string]*fieldDef {
	t := reflect.TypeOf(dest)
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		t = t.Elem()
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		panic("type of dest must be any of struct, []struct, []*struct")
	}

	m := make(map[string]*fieldDef, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name, ok := f.Tag.Lookup(csvColNameTagKey)
		if !ok {
			continue
		}

		def := &fieldDef{
			FieldIndex:  i,
			ParserFunc:  getParserFunc(f.Type),
			parseOption: &parseOption{},
		}
		if v, ok := f.Tag.Lookup(defaultValueTagKey); ok {
			def.DefaultValue = v
		}
		if v, ok := f.Tag.Lookup(scanTimeFormatTagKey); ok {
			def.TimeFormat = v
		}
		m[name] = def
	}
	return m
}

func headerAsMap(header []string) map[string]int {
	m := make(map[string]int, len(header))
	for i := range header {
		m[header[i]] = i
	}
	return m
}

func getParserFunc(t reflect.Type) parserFunc {
	k := t.Kind()
	fn, ok := builtInTypeParsers[k]
	if !ok {
		fn, ok = structTypeParsers[t]
	}
	if ok && fn != nil {
		return fn
	}
	panic(fmt.Errorf("%w: %v", ErrUnsupportedType, t))
}
