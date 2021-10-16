package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/spf13/pflag"
	"github.com/uhey22e/refiller"
)

var (
	output = pflag.StringP("output", "o", "refill", "Output directory name.")
	dest   = pflag.StringP("destination", "d", "testdata/ex1/ex1.Ex1", "Destination struct.")
	src    = pflag.StringP("source", "s", "testdata/ex2/ex2.Ex2", "Source struct.")
)

func main() {
	pflag.Parse()

	sp := strings.Split(*dest, ".")
	dstPath := sp[0] + ".go"
	dstName := sp[1]
	sp = strings.Split(*src, ".")
	srcPath := sp[0] + ".go"
	srcName := sp[1]

	if err := os.MkdirAll(*output, os.ModePerm); err != nil {
		log.Fatal(err)
	}
	o, err := os.Create(filepath.Join(*output, strcase.ToSnake(dstName)+".go"))
	if err != nil {
		log.Fatal(err)
	}
	defer o.Close()

	err = refiller.Generate(o, dstPath, dstName, srcPath, srcName)
	if err != nil {
		log.Fatal(err)
	}
}
