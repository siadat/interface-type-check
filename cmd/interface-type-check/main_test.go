package main_test

import (
	"errors"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func addLinenumbers(src string) string {
	lines := strings.Split(src, "\n")
	numberedLines := make([]string, len(lines))
	for i := range lines {
		numberedLines[i] = fmt.Sprintf("%3d %s", i+1, lines[i])
	}
	return strings.Join(numberedLines, "\n")
}

func TestParserAndTypes(t *testing.T) {
	testCases := []struct {
		src      string
		typeErr  error
		parseErr error
	}{
		{
			src: `
				package whatever
				
				type MyC interface {
					// #type string
				}
			`,
			typeErr: nil,
		},
		{
			src: `
				package whatever
				
				type MyC interface {
					// #type string, int
				}
			`,
			typeErr: nil,
		},
		{
			src: `
				package whatever
				
				type MyC interface {
					// #type string, int, T
				}
			`,
			typeErr: errors.New(`testfile.go:5:24: undeclared name: T`),
		},
		{
			src: `
				package whatever
				
				type T struct {
					Name string
				}
				
				type MyC interface {
					// #type string, int, T
				}
			`,
			typeErr: nil,
		},
		{
			src: `
				package whatever
				
				type T struct {
					Name string
				}
				
				type MyC interface {
					// #type T, string, int, MyC
				}
			`,
			typeErr: nil,
		},
		{
			src: `
				package whatever
				
				type T struct {
					Name string
				}
				
				type MyC2 interface {
					// #type bool, struct{}
				}
				
				type MyC1 interface {
					// #type T, string, int, MyC2
				}
			`,
			typeErr: nil,
		},
		{
			src: `
				package whatever
				
				type MyI interface {
				}
				
				func Handle(c MyI) {
					switch c.(type) {
					case string:
					}
				}
			`,
			typeErr: nil,
		},
		{
			src: `
				package whatever
				
				type MyC interface {/* #type string, int */}
				
				func Handle(c MyC) {
					switch c.(type) {
					case string, int, nil:
					}
				}
			`,
			typeErr: nil,
		},
		{
			src: `
				package whatever
				
				type MyC interface {
					// #type string, int
				}
				
				func Handle(c MyC) {
					switch c.(type) {
					default:
					}
				}
			`,
			typeErr: nil,
		},
		{
			src: `
				package whatever
				
				type MyC interface { /* #type string, int */ }
				
				func Handle(c MyC) {
					switch c.(type) {
					case string, int:
					}
				}
			`,
			typeErr: nil,
		},
		{
			src: `
				package whatever
				
				type MyC interface {
					// #type string
				}
				
				func Handle(c MyC) {
					switch c.(type) {
					case bool:
					}
				}
			`,
			typeErr: errors.New("testfile.go:10:7: c (variable of type MyC) cannot have dynamic type bool (mismatching sum assertion)"),
		},
		{
			src: `
				package whatever
				
				type MyC interface {
					// #type string
				}
				
				func Handle(c MyC) {
					c = 5
				}
			`,
			typeErr: errors.New("testfile.go:9:6: cannot use 5 (constant of type int) as MyC value in assignment: mismatching sum type (have int, want a type in interface{type string})"),
		},
		{
			src: `
				package whatever
				
				type MyC interface {
					// #type string
				}
				
				func Handle(c MyC) {
					switch c.(type) {
					case string:
					}
				}
			`,
			typeErr: errors.New("testfile.go:9:9: missing sum case nil in type switch"),
		},
		{
			src: `
				package whatever
				
				type MyC interface {
					// #type string, int
				}
				
				func Handle(c MyC) {
					switch c.(type) {
					case string, int:
					}
				}
			`,
			typeErr: errors.New("testfile.go:9:9: missing sum case nil in type switch"),
		},
		{
			src: `
				package whatever
				
				type MyC interface {
					// #type string, int
				}
				
				func Handle(c MyC) {
					switch c.(type) {
					case string, nil:
					}
				}
			`,
			typeErr: errors.New("testfile.go:9:9: missing sum case int in type switch"),
		},
		{
			src: `
				package whatever
				
				type MyC interface {
					// #type string, int
				}
				
				func Handle(c MyC) {
					switch c.(type) {
					case nil:
					}
				}
			`,
			typeErr: errors.New("testfile.go:9:9: missing sum case string in type switch"),
		},
		{
			src: `
				package whatever
				
				type MyC interface {
					// #type string, int
				}
				
				func Handle(c MyC) {
					switch c.(type) {
					case string, int, nil, bool:
					}
				}
			`,
			typeErr: errors.New("testfile.go:10:25: c (variable of type MyC) cannot have dynamic type bool (mismatching sum assertion)"),
		},
		{
			src: `
				package whatever
				
				type MyC interface {
					// #type string, int
				}
				
				func Handle(c MyC) {
					_ = c.(bool)
				}
			`,
			typeErr: errors.New("testfile.go:9:6: c (variable of type MyC) cannot have dynamic type bool (mismatching sum assertion)"),
		},
		{
			src: `
				package whatever
				
				type MyC interface {
					// #type string, int
				}
				
				func Handle(c MyC) {
					_ = c.(int)
				}
			`,
			typeErr: nil,
		},
		{
			src: `
				package whatever
				
				type MyC interface {
					// #type string, int
				}
				
				func main() {
					var x MyC = "hello"
					_ = x

					var y MyC = 5
					_ = y
				}
			`,
			typeErr: nil,
		},
		{
			src: `
				package whatever
				
				type MyC interface {
					// #type string, int
				}
				
				func main() {
					var x MyC = 1.1
					_ = x
				}
			`,
			typeErr: errors.New("testfile.go:9:14: cannot use 1.1 (constant of type float64) as MyC value in variable declaration: mismatching sum type (have float64, want a type in interface{type string, int})"),
		},
		{
			src: `
				package whatever
				
				type MyS struct {
					Name string
				}
				type MyC interface {
					// #type bool, MyS
				}
				
				func main() {
					var x MyC = MyS{Name: "hi"}
					switch xx := x.(type) {
					case bool: _ = xx
					case MyS:  _ = xx.Name
					case nil:
					}
				}
			`,
			typeErr: nil,
		},
		{
			src: `
				package whatever
				
				//#type string, int,,,,
				// # type string, int,,,,
				// #typestring, int,,,,

				type MyC1 interface {
					//#type string, int,,,,
				}

				type MyC2 interface {
					// # type string, int,,,,
				}

				type MyC3 interface {
					// #typestring, int,,,,
				}
			`,
			parseErr: nil,
			typeErr:  nil,
		},
		{
			src: `
				package whatever
				// #type string, int
			`,
			parseErr: errors.New(`testfile.go:3:16: expected type, found ','`),
			typeErr:  nil,
		},
		{
			src: `
				package whatever
				
				type MyC interface {
					// #type int,
					//       string
				}
			`,
			parseErr: errors.New(`testfile.go:7:1: expected type, found '}'`),
		},
		{
			src: `
				package whatever
				
				type T struct {
					Name string
				}
				
				type MyC interface {
					// #type T, string, int
					// #type bool
				}
			`,
			typeErr: errors.New("testfile.go:10:6: cannot have multiple type lists in an interface"),
		},
		{
			src: `
				package whatever
				
				// import "fmt"
				
				var x = 3.14
				var xx MyC // should fail
				
				func main() {
					// switch (interface{}(nil)).(type) {
					// case int, bool, string:
					// }
				
					var a int
					// var c MjijijyC // should fail
					// var c MyC // should fail
				
					var c = "hola" // ok
				
					// var c interface{string; int} // should fail
					// c := T{Name: "sina"} // ok
					// c := T2{Name: "sina"} // should fail
					// c := MyC{} // should fail
				
					_ = a
					_ = c
				
					DoC(c)
				}
				
				type MyI interface {
					Method1()
					 Method2()
				}
				
				type T struct {
					Name string
				}
				
				type T2 struct {
					Name string
				}
				
				type MyC interface {
					// #type T, string, int
				}
				
				type MyC2 interface{}
				
				func DoC(c MyC) (x MyC) {
					// c = 3   // ok fail
					// c = 3.2 // should fail
				
					// err := fmt.Errorf("ok?")
					// switch err.(type) {
					// case int:
					// 	fmt.Println("INT")
					// }
				
					switch c.(type) {
					case int:
					case string:
					case T:
					case nil:
					// case bool: // err
					// default: // ok even when missing a case
					}
				
				
					return
				}
				
				func DoI(c MyI) {
				}
			`,
			typeErr: nil,
		},
	}

	for ti, tt := range testCases {
		// src := tt.src
		src := func() string {
			lines := strings.Split(tt.src, "\n")
			indent := ""

			for i := range lines {
				if strings.TrimSpace(lines[i]) == "package whatever" {
					indent = strings.TrimSuffix(lines[i], "package whatever")
					break
				}
			}

			for i := range lines {
				lines[i] = strings.TrimPrefix(lines[i], indent)
			}

			return strings.Join(lines, "\n")
		}()

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "testfile.go", src, parser.ParseComments)
		if tt.parseErr == nil {
			require.NoError(t, err, fmt.Sprintf("case %d\n", ti)+addLinenumbers(src))
		} else {
			require.Error(t, err, fmt.Sprintf("case %d\n", ti)+addLinenumbers(src))
			require.Equal(t, tt.parseErr.Error(), err.Error(), fmt.Sprintf("case %d\n", ti)+addLinenumbers(src))
		}

		info := types.Info{
			Types: make(map[ast.Expr]types.TypeAndValue),
			Defs:  make(map[*ast.Ident]types.Object),
			Uses:  make(map[*ast.Ident]types.Object),
		}
		conf := types.Config{Importer: importer.Default()}
		_, err = conf.Check("testfile", fset, []*ast.File{f}, &info)
		if tt.typeErr == nil {
			require.NoError(t, err, fmt.Sprintf("case %d\n", ti)+addLinenumbers(src))
		} else {
			require.Error(t, err, fmt.Sprintf("case %d\n", ti)+addLinenumbers(src))
			require.Equal(t, tt.typeErr.Error(), err.Error(), fmt.Sprintf("case %d\n", ti)+addLinenumbers(src))
		}

	}
}
