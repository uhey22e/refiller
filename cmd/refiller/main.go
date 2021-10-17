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

func getOutputFilename(t string) string {
	s := strings.Split(t, ".")
	if len(s) != 2 {
		panic("Invalid target name: " + t)
	}
	return strcase.ToSnake(s[1]) + ".go"
}

func main() {
	pflag.Parse()

	var err error
	if err := os.MkdirAll(*output, os.ModePerm); err != nil {
		log.Fatal(err)
	}
	o, err := os.Create(filepath.Join(*output, getOutputFilename(*dest)))
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

	err = refiller.Generate(o, filepath.Base(*output), *dest, *src)
	if err != nil {
		log.Println(err)
	}
}
