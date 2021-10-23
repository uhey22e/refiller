package refiller

import (
	"errors"
	"reflect"
)

type Row struct {
	ID   string `refiller:"id"`
	Name string `refiller:"name"`
}

var (
	ErrUnsupportedType = errors.New("unsupported type")

	dbColNameTagKey = "column"
)

func Columns(v interface{}) []string {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		t = t.Elem()
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		panic("type of v must be any of struct, []struct or []*struct")
	}

	cols := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if v, ok := f.Tag.Lookup(dbColNameTagKey); ok {
			cols = append(cols, v)
		}
	}
	return cols
}

func Values(v interface{}) []interface{} {
	t := reflect.TypeOf(v)
	rv := reflect.ValueOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		rv = rv.Elem()
	}
	if t.Kind() != reflect.Struct {
		panic("type of v must be any of struct or *struct")
	}

	values := make([]interface{}, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		if _, ok := t.Field(i).Tag.Lookup(dbColNameTagKey); ok {
			x := rv.Field(i).Interface()
			values = append(values, x)
		}
	}
	return values
}
