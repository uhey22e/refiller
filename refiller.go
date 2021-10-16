package refiller

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"path/filepath"
	"text/template"
	"unicode"

	"github.com/iancoleman/strcase"
	"golang.org/x/mod/modfile"
)

type RenderArgs struct {
	Package string
	Imports []string
	Dest    *Name
	Src     *Name
	FwPairs []*Pair
	BwPairs []*Pair
}

type Name struct {
	Package string
	Name    string
}

type Pair struct {
	Dest string
	Src  string
}

var (
	ErrStructNotFound = errors.New("struct is not found")
	ErrNotStruct      = errors.New("not a struct")
)

var (
	//go:embed templates
	tmpls    embed.FS
	fillTmpl = template.Must(template.ParseFS(tmpls, "templates/refiller.go.tmpl"))
)

func Generate(w io.Writer, dstPath, dstName, srcPath, srcName string) error {
	b, err := ioutil.ReadFile("go.mod")
	if err != nil {
		return err
	}
	mod, err := modfile.Parse("go.mod", b, nil)
	if err != nil {
		return err
	}
	root := mod.Module.Mod.Path

	fw, err := InspectPairs(dstPath, dstName, srcPath, srcName)
	if err != nil {
		return err
	}
	bw, err := InspectPairs(srcPath, srcName, dstPath, dstName)
	if err != nil {
		return err
	}
	args := RenderArgs{
		Package: "demo",
		Imports: []string{
			filepath.Join(root, filepath.Dir(dstPath)),
			filepath.Join(root, filepath.Dir(srcPath)),
		},
		Dest: &Name{
			Package: getPackageName(dstPath),
			Name:    dstName,
		},
		Src: &Name{
			Package: getPackageName(srcPath),
			Name:    srcName,
		},
		FwPairs: fw,
		BwPairs: bw,
	}
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	if err := fillTmpl.Execute(buf, args); err != nil {
		return err
	}
	res, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}
	w.Write(res)
	return nil
}

func InspectPairs(dstPath, dstName, srcPath, srcName string) ([]*Pair, error) {
	fset := token.NewFileSet()
	dstFile, err := parser.ParseFile(fset, dstPath, nil, parser.Mode(0))
	if err != nil {
		return nil, err
	}
	srcFile, err := parser.ParseFile(fset, srcPath, nil, parser.Mode(0))
	if err != nil {
		return nil, err
	}

	dst, err := FindStruct(dstFile, dstName)
	if err != nil {
		return nil, err
	}
	src, err := FindStruct(srcFile, srcName)
	if err != nil {
		return nil, err
	}

	pairs := MakePairs(dst, src)
	return pairs, nil
}

func MakePairs(dst, src *ast.StructType) []*Pair {
	m := make(map[string]string, len(dst.Fields.List))
	for _, f := range src.Fields.List {
		name := getFieldName(f)
		if isPrivate(name) {
			continue
		}
		m[getKey(name)] = name
	}
	res := make([]*Pair, 0, len(dst.Fields.List))
	for _, f := range dst.Fields.List {
		name := getFieldName(f)
		if isPrivate(name) {
			continue
		}
		if s, ok := m[getKey(name)]; ok {
			res = append(res, &Pair{
				Dest: name,
				Src:  s,
			})
		}
	}
	return res
}

func FindStruct(file *ast.File, name string) (*ast.StructType, error) {
	for _, d := range file.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, s := range gd.Specs {
			ts, ok := s.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if ts.Name.Name != name {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				return nil, fmt.Errorf("%w: %s.%s", ErrNotStruct, file.Name.String(), name)
			}
			return st, nil
		}
	}
	return nil, fmt.Errorf("%w: %s.%s", ErrStructNotFound, file.Name.String(), name)
}

func getFieldName(f *ast.Field) string {
	if len(f.Names) == 0 {
		panic(fmt.Sprintf("unexpected names: %v", f.Names))
	}
	return f.Names[0].Name
}

func getKey(s string) string {
	return strcase.ToSnake(s)
}

func isPrivate(name string) bool {
	if len(name) == 0 {
		return false
	}
	return unicode.IsLower(rune(name[0]))
}

func getPackageName(path string) string {
	d := filepath.Base(filepath.Dir(path))
	if d == "." {
		d = ""
	}
	return d
}
