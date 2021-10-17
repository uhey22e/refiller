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
	"strings"
	"text/template"
	"unicode"

	"github.com/iancoleman/strcase"
	"golang.org/x/mod/modfile"
)

type RenderArgs struct {
	Package  string
	Imports  []string
	DestType string
	DestName string
	SrcType  string
	SrcName  string
	Pairs    []*Pair
}

type Pair struct {
	Dest string
	Src  string
}

type Target struct {
	Package    string
	StructName string
}

var (
	ErrStructNotFound = errors.New("struct is not found")
	ErrNotStruct      = errors.New("not a struct")
	ErrInvalidTarget  = errors.New("invalid target")
)

var (
	//go:embed templates
	tmpls    embed.FS
	fillTmpl = template.Must(template.ParseFS(tmpls, "templates/refiller.go.tmpl"))
)

func Generate(w io.Writer, packageName, dest, src string) error {
	b, err := ioutil.ReadFile("go.mod")
	if err != nil {
		return err
	}
	mod, err := modfile.Parse("go.mod", b, nil)
	if err != nil {
		return err
	}
	root := mod.Module.Mod.Path

	d, err := parseTarget(dest)
	if err != nil {
		return err
	}
	s, err := parseTarget(src)
	if err != nil {
		return err
	}
	pairs, err := InspectPairs(d.Package, d.StructName, s.Package, s.StructName)
	if err != nil {
		return err
	}

	args := RenderArgs{
		Package: packageName,
		Imports: []string{
			filepath.Join(root, d.Package),
			filepath.Join(root, s.Package),
		},
		DestType: filepath.Base(d.Package) + "." + d.StructName,
		DestName: d.StructName,
		SrcType:  filepath.Base(s.Package) + "." + s.StructName,
		SrcName:  s.StructName,
		Pairs:    pairs,
	}
	if args.DestName == args.SrcName {
		args.DestName = strcase.ToCamel(d.Package) + args.DestName
		args.SrcName = strcase.ToCamel(s.Package) + args.SrcName
	}

	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	if err := fillTmpl.Execute(buf, args); err != nil {
		return err
	}
	res, err := format.Source(buf.Bytes())
	if err != nil {
		buf.WriteTo(w)
		return err
	}
	w.Write(res)
	return nil
}

func InspectPairs(destPath, destName, srcPath, srcName string) ([]*Pair, error) {
	fset := token.NewFileSet()
	dstPkgs, err := parser.ParseDir(fset, destPath, nil, parser.Mode(0))
	if err != nil {
		return nil, err
	}
	srcPkgs := dstPkgs
	if srcPath != destPath {
		srcPkgs, err = parser.ParseDir(fset, srcPath, nil, parser.Mode(0))
		if err != nil {
			return nil, err
		}
	}

	d, err := FindStructFromPackages(dstPkgs, destName)
	if err != nil {
		return nil, err
	}
	s, err := FindStructFromPackages(srcPkgs, srcName)
	if err != nil {
		return nil, err
	}

	pairs := MakePairs(d, s)
	return pairs, nil
}

func MakePairs(dst, src *ast.StructType) []*Pair {
	m := make(map[string]*ast.Field, len(dst.Fields.List))
	for _, f := range src.Fields.List {
		name := getFieldName(f)
		if isPrivate(name) {
			continue
		}
		m[getKey(name)] = f
	}
	res := make([]*Pair, 0, len(dst.Fields.List))
	for _, f := range dst.Fields.List {
		name := getFieldName(f)
		if isPrivate(name) {
			continue
		}
		if s, ok := m[getKey(name)]; ok {
			if getTypeName(f) != getTypeName(s) {
				continue
			}
			res = append(res, &Pair{
				Dest: name,
				Src:  getFieldName(s),
			})
		}
	}
	return res
}

func FindStructFromPackages(pkgs map[string]*ast.Package, name string) (*ast.StructType, error) {
	for _, pkg := range pkgs {
		s, err := FindStructFromPackage(pkg, name)
		if errors.Is(err, ErrStructNotFound) {
			continue
		} else if err != nil {
			return nil, err
		}
		return s, nil
	}
	return nil, fmt.Errorf("%w: %s", ErrStructNotFound, name)
}

func FindStructFromPackage(pkg *ast.Package, name string) (*ast.StructType, error) {
	for _, file := range pkg.Files {
		s, err := FindStructFromFile(file, name)
		if errors.Is(err, ErrStructNotFound) {
			continue
		} else if err != nil {
			return nil, err
		}
		return s, nil
	}
	return nil, fmt.Errorf("%w: %s.%s", ErrStructNotFound, pkg.Name, name)
}

func FindStructFromFile(file *ast.File, name string) (*ast.StructType, error) {
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
				return nil, fmt.Errorf("%w: %s %s", ErrNotStruct, file.Name, name)
			}
			return st, nil
		}
	}
	return nil, fmt.Errorf("%w: %s %s", ErrStructNotFound, file.Name, name)
}

func getFieldName(f *ast.Field) string {
	if len(f.Names) == 0 {
		panic(fmt.Sprintf("unexpected names: %v", f.Names))
	}
	return f.Names[0].Name
}

func getTypeName(f *ast.Field) string {
	switch v := f.Type.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		x := v.X.(*ast.Ident).Name
		sel := v.Sel.Name
		return x + "." + sel
	}
	if t, ok := f.Type.(*ast.Ident); ok {
		return t.Name
	}
	panic(fmt.Sprintf("unknown type: %#v", f.Type))
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

func parseTarget(t string) (*Target, error) {
	s := strings.Split(t, ".")
	if len(s) != 2 {
		return nil, fmt.Errorf("%w: %s", ErrInvalidTarget, t)
	}
	return &Target{
		Package:    s[0],
		StructName: s[1],
	}, nil
}
