package main

import (
	"fmt"
	"os"
	"runtime"

	"golang.org/x/tools/go/packages"
)

func main() {
	conf := &packages.Config{
		Mode:  packages.LoadAllSyntax,
		Tests: true,
	}

	paths := os.Args[1:]

	if len(paths) == 0 {
		paths = []string{"."}
	}
	pkgs, err := packages.Load(conf, paths...)
	if err != nil {
		panic(err)
	}
	runtime.GC()

	for _, pkg := range pkgs {
		if pkg.IllTyped {
			for _, err := range pkg.Errors {
				fmt.Printf("error: %v\n", err.Error())
			}
			return
		}
	}
}
