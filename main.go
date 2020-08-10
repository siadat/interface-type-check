// vim: noexpandtab
// vim: tabstop=8
// vim: shiftwidth=8
package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"runtime"

	"golang.org/x/tools/go/packages"
)

// func main2() {
// 	var importerConf loader.Config
//
// 	// Use the command-line arguments to specify
// 	// a set of initial packages to load from source.
// 	// See FromArgsUsage for help.
// 	_, err := importerConf.FromArgs(os.Args[1:], false)
// 	if err != nil {
// 		panic(fmt.Sprintf("args error: %s", err))
// 	}
//
// 	conf := types.Config{Importer: importerConf}
//
// 	importerConf.TypeChecker = conf
// 	// Finally, load all the packages specified by the configuration.
// 	prog, err := importerConf.Load()
// 	if err != nil {
// 		panic(fmt.Sprintf("load error: %s", err))
// 	}
//
// 	for pkgName, pkg := range prog.Created {
// 		for fileName, f := range pkg.Files {
// 			_, err = conf.Check("fib", prog.Fset, []*ast.File{f}, &info)
// 			if err != nil {
// 				panic(fmt.Sprintf("pkg: %s, filename: %s, err: %v", pkgName, fileName, err))
// 			}
// 		}
// 	}
//
// }

func main_ok() {
	dirPath := os.Args[1]

	fset := token.NewFileSet()
	// f, err := parser.ParseFile(fset, "fib.go", src, parser.ParseComments)
	pkgs, err := parser.ParseDir(fset, dirPath, nil, parser.ParseComments)
	if err != nil {
		log.Fatal("parse failed:", err)
	}

	// Type-check the package.
	// We create an empty map for each kind of input
	// we're interested in, and Check populates them.
	info := types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}
	conf := types.Config{Importer: importer.Default()}

	// pkg, err := conf.Check("fib", fset, []*ast.File{f}, &info)
	for pkgName, pkg := range pkgs {
		fmt.Println("pkg", pkgName, pkg)
		for fileName, f := range pkg.Files {
			_, err = conf.Check("fib", fset, []*ast.File{f}, &info)
			if err != nil {
				panic(fmt.Sprintf("pkg: %s, filename: %s, err: %v", pkgName, fileName, err))
			}
		}
	}
}

func main() {
	conf := &packages.Config{
		Mode:  packages.LoadAllSyntax,
		Tests: true,
		// BuildFlags: []string{
		// 	"-tags=" + strings.Join(opt.Tags, " "),
		// },
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
		// fmt.Println(pkg)
		// fmt.Printf("== %#v\n", pkg)
		if pkg.IllTyped {
			for _, err := range pkg.Errors {
				fmt.Printf("error: %v\n", err.Error())
			}
			return
		}
	}
}
