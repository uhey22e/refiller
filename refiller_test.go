package refiller

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGenerate(t *testing.T) {
	dest := "testdata/ex1.Ex1"
	src := "testdata/ex2.Ex2"
	w := &bytes.Buffer{}

	if err := Generate(w, "test", dest, src); err != nil {
		t.Errorf("Generate() error = %v", err)
		t.Log(w.String())
		return
	}
	t.Log(w.String())
}

func TestInspectPairs(t *testing.T) {
	destPath := filepath.Join("testdata", "ex1")
	destName := "Ex1"
	srcPath := filepath.Join("testdata", "ex2")
	srcName := "Ex2"

	want := []*Pair{
		{
			Dest: "ID",
			Src:  "Id",
		},
		{
			Dest: "Name",
			Src:  "Name",
		},
		{
			Dest: "Timestamp",
			Src:  "Timestamp",
		},
	}

	got, err := InspectPairs(destPath, destName, srcPath, srcName)
	if err != nil {
		t.Errorf("InspectPairs() error = %v", err)
		return
	}
	if !cmp.Equal(got, want) {
		t.Errorf("InspectPairs() = %v", cmp.Diff(got, want))
	}
}

func TestMakePairs(t *testing.T) {
	dst := &ast.StructType{
		Fields: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{ast.NewIdent("ID")},
					Type:  ast.NewIdent("string"),
				},
				{
					Names: []*ast.Ident{ast.NewIdent("Name")},
					Type:  ast.NewIdent("string"),
				},
				{
					Names: []*ast.Ident{ast.NewIdent("ShortName")},
					Type:  ast.NewIdent("string"),
				},
				{
					Names: []*ast.Ident{ast.NewIdent("variant")},
					Type:  ast.NewIdent("string"),
				},
				{
					Names: []*ast.Ident{ast.NewIdent("Date")},
					Type:  ast.NewIdent("time.Time"),
				},
			},
		},
	}
	src := &ast.StructType{
		Fields: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{ast.NewIdent("Id")},
					Type:  ast.NewIdent("string"),
				},
				{
					Names: []*ast.Ident{ast.NewIdent("Name")},
					Type:  ast.NewIdent("string"),
				},
				{
					Names: []*ast.Ident{ast.NewIdent("shortName")},
					Type:  ast.NewIdent("string"),
				},
				{
					Names: []*ast.Ident{ast.NewIdent("Variant")},
					Type:  ast.NewIdent("string"),
				},
				{
					Names: []*ast.Ident{ast.NewIdent("Date")},
					Type:  ast.NewIdent("string"),
				},
			},
		},
	}
	want := []*Pair{
		{
			Dest: "ID",
			Src:  "Id",
		},
		{
			Dest: "Name",
			Src:  "Name",
		},
	}
	if got := MakePairs(dst, src); !cmp.Equal(got, want) {
		t.Errorf("MakePairs() = %v", cmp.Diff(got, want))
	}
}

func TestFindStructFromFile(t *testing.T) {
	path := filepath.Join("testdata", "ex1", "ex1.go")
	name := "Ex1"

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.Mode(0))
	if err != nil {
		t.Fatal(err)
	}

	s, err := FindStructFromFile(file, name)
	if err != nil {
		t.Errorf("FindStruct() error = %v", err)
		return
	}
	ast.Print(fset, s)
}

func Test_isPrivate(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			args: args{
				name: "member",
			},
			want: true,
		},
		{
			args: args{
				name: "Member",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPrivate(tt.args.name); got != tt.want {
				t.Errorf("isPrivate() = %v, want %v", got, tt.want)
			}
		})
	}
}
