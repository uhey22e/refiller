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
	output = pflag.StringP("output", "o", "refiller", "Output directory name.")
	dest   = pflag.StringP("destination", "d", "", "Destination struct.")
	src    = pflag.StringP("source", "s", "", "Source struct.")
)

func main() {
	pflag.Parse()

	sp := strings.Split(*dest, ".")
	dstPath := sp[0]
	dstName := sp[1]
	sp = strings.Split(*src, ".")
	srcPath := sp[0]
	srcName := sp[1]

	if err := os.MkdirAll(*output, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	var err error
	o, err := os.Create(filepath.Join(*output, strcase.ToSnake(dstName)+".go"))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		o.Close()
		if err != nil {
			if err := os.RemoveAll(*output); err != nil {
				log.Print(err)
			}
		}
	}()

	err = refiller.Generate(o, filepath.Base(*output), dstPath, dstName, srcPath, srcName)
	if err != nil {
		log.Println(err)
	}
}
