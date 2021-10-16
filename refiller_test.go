package refiller

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGenerate(t *testing.T) {
	dstPath := filepath.Join("testdata", "ex1", "ex1.go")
	dstName := "Ex1"
	srcPath := filepath.Join("testdata", "ex2", "ex2.go")
	srcName := "Ex2"
	w := &bytes.Buffer{}

	if err := Generate(w, dstPath, dstName, srcPath, srcName); err != nil {
		t.Errorf("Generate() error = %v", err)
		return
	}
	t.Log(w.String())
}

func TestInspectPairs(t *testing.T) {
	dstPath := filepath.Join("testdata", "ex1", "ex1.go")
	dstName := "Ex1"
	srcPath := filepath.Join("testdata", "ex2", "ex2.go")
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
	}

	got, err := InspectPairs(dstPath, dstName, srcPath, srcName)
	if err != nil {
		t.Errorf("InspectPairs() error = %v", err)
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("InspectPairs() = %v, want %v", got, want)
	}
}

func TestMakePairs(t *testing.T) {
	dst := &ast.StructType{
		Fields: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{ast.NewIdent("ID")},
				},
				{
					Names: []*ast.Ident{ast.NewIdent("Name")},
				},
				{
					Names: []*ast.Ident{ast.NewIdent("ShortName")},
				},
				{
					Names: []*ast.Ident{ast.NewIdent("variant")},
				},
			},
		},
	}
	src := &ast.StructType{
		Fields: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{ast.NewIdent("Id")},
				},
				{
					Names: []*ast.Ident{ast.NewIdent("Name")},
				},
				{
					Names: []*ast.Ident{ast.NewIdent("shortName")},
				},
				{
					Names: []*ast.Ident{ast.NewIdent("Variant")},
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

func TestFindStruct(t *testing.T) {
	path := filepath.Join("testdata", "ex1", "ex1.go")
	name := "Ex1"

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.Mode(0))
	if err != nil {
		t.Fatal(err)
	}

	_, err = FindStruct(file, name)
	if err != nil {
		t.Errorf("FindStruct() error = %v", err)
		return
	}
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

func Test_getPackageName(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{
				path: "ex1.go",
			},
			want: "",
		},
		{
			args: args{
				path: filepath.Join("ex1", "ex1.go"),
			},
			want: "ex1",
		},
		{
			args: args{
				path: filepath.Join("testdata", "ex1", "ex1.go"),
			},
			want: "ex1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPackageName(tt.args.path); got != tt.want {
				t.Errorf("getPackageName() = %v, want %v", got, tt.want)
			}
		})
	}
}
