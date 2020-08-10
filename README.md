# Static type checker for interface{}

This is an experiment.

* I started this experiment because I was looking for a way to implement something like OCaml's variant or Rust's enum in a way that is idiomatic Go.
* No code generation, works with Go 1 compiler.
* You need to add a special comment, eg `// #type T1, T2`
* It is a tool that performs a **static type check** on interface{} values.
* Internally, the implementation is based on a new type in go2go's (the go2go/types.Sum type).
  Go2go and this experiment have different concerns: go2go is about generic functions, this experiment is about sum types and sum variables.
  There doesn't seem to be any incompatibilities.

## Download

* [windows-amd64]()
* [darwin-amd64]()
* [linux-amd64]()

## Build

```bash
go get -d github.com/siadat/interface-type-check
cd interface-type-check

cd ./go/src && ./all.bash
GOROOT=./go ./go/bin/go install .
```

## How

Without a type list / type constrain:

```go
/*  1     */  package main
/*  2     */
/*  3     */  type Numeric interface{}
/*  4     */  
/*  5     */  func main() {
/*  6     */  	var n1 Numeric = 5
/*  7     */  	var n2 Numeric = "bad value"
/*  8     */  	_ = n1
/*  9     */  	_ = n2
/* 10     */  }
```

With a type list / type constraint (L4):

```go
/*  1     */  package main
/*  2     */
/*  3     */  type Numeric interface{
/*  4 --> */  	// #type int, float64
/*  5     */  }
/*  6     */  
/*  7     */  func main() {
/*  8 OK  */  	var n1 Numeric = 5
/*  9 ERR */  	var n2 Numeric = "bad value"
/* 10     */  	_ = n1
/* 11     */  	_ = n2
/* 12     */  }
```

Now perform a check to get the error:

```bash
$ interface-type-check .
testfile.go:10:6: cannot use "bad value" (constant of type string) as Numeric value in variable declaration: mismatching sum type (have string, want a type in interface{type int, float64})
```

## Checks

Given the declaration:

```go
type Numeric interface{
	// #type int, float64
}
```

The following checks are performed:

- asign a value of an unlisted type
  ```go
var number Numeric = "abc" // CHECK ERR: expected int or float
  ```
- assert to an unlisted type
  ```go
_, _ = number.(string) // CHECK ERR: string not allowed
  ```
- include an unlisted type in a type switch
  ```go
switch number.(type) {
case string: // CHECK ERR: string not allowed
case float:
}
  ```
- neglect a type in a type switch
  ```go
switch number.(type) { // CHECK ERR: missing case for int
case float:
}
  ```


<!--
### src/cmd/compile/internal/ssa/gen/rulegen.go Node and Statement
### database/sql/driver.Value
### plugin.Symbol
### xml.Token
-->

## Example (json.Token)

All supported types of [encoding/json.Token](https://pkg.go.dev/encoding/json?tab=doc#Token) are known,
as documented here:

```go
// A Token holds a value of one of these types:
//
//	Delim, for the four JSON delimiters [ ] { }
//	bool, for JSON booleans
//	float64, for JSON numbers
//	Number, for JSON numbers
//	string, for JSON string literals
//	nil, for JSON null
//
type Token interface{}
```

Adding the #type comment, it would look like this:

```go
// A Token holds a value of one of these types:
//
//	Delim, for the four JSON delimiters [ ] { }
//	bool, for JSON booleans
//	float64, for JSON numbers
//	Number, for JSON numbers
//	string, for JSON string literals
//	nil, for JSON null
//
type Token interface {
	// #type Delim, bool, float64, Number, string, nil
}
```

That's all we need to be able to use the checker.

## Example (sql.Scanner)

[database/sql.Scanner](https://pkg.go.dev/database/sql?tab=doc#Scanner) is also defined as an empty interface whose possible types are known.

Before:

```go
// Scanner is an interface used by Scan.
type Scanner interface {
	// Scan assigns a value from a database driver.
	//
	// The src value will be of one of the following types:
	//
	//    int64
	//    float64
	//    bool
	//    []byte
	//    string
	//    time.Time
	//    nil - for NULL values
	//
	Scan(src interface{}) error
}
```

After:

```go
// Scanner is an interface used by Scan.
type Scanner interface {
	Scan(src SourceType) error
}

type SourceType interface {
	// #type int64, float64, bool, []byte, string, time.Time, nil
}
```

## Example (Connection type)

The Connecting type has a retry field:

```go
type Connected    struct{}
type Disconnected struct{}
type Connecting   struct{ rety int }

type Connection interface {
	// #type Connected, Disconnected, Connecting
}

func log(conn Connection) int {
	switch c := conn.(type) {
	case Connected:    fmt.Println("Connected")
	case Disconnected: fmt.Println("Disconnected")
	case Connecting:   fmt.Println("Connecting, retry:", c.retry)
	case nil:          panic("conn is nil")
	}
}
```

## Example (error)

```go
type RequestOrError interface {
	// #type *http.Request, error
}

func H(re RequestOrError) {
	switch x := re.(type) {
	case *http.Request: fmt.Println("received request", x)
	case error:         fmt.Println("received error", x)
	case nil:           fmt.Println("received nil")
	}
}
```


## Example (net.IP)

The standard library defines one [net.IP](https://pkg.go.dev/net#IP) type for both IPv4 and IPv6 IPs:

```go
// An IP is a single IP address, a slice of bytes.
// Functions in this package accept either 4-byte (IPv4)
// or 16-byte (IPv6) slices as input.
type IP []byte
```

This type has a String() function, which relies on runtime checks to detect the version of the IP [here](https://github.com/golang/go/blob/edfd6f28486017dcb136cd3f3ec252706d4b326e/src/net/ip.go#L299):

```go
if p4 := p.To4(); len(p4) == IPv4len { ...
```

There are very good reasons to use a simple []byte data structure for the IPs.
I am *not* suggesting that this code should change.
I am only running tiny hypothetical experiments. With that in mind, let's write it using `// #type`:

```go
type IPv4 [4]byte
type IPv6 [16]byte

type IP interface {
	// #type IPv4, IPv6
}

func version(ip IP) int {
	switch ip.(type) {
	case IPv4: return 4
	case IPv6: return 6
	case nil:  panic("ip is nil")
	}
}
```

## When to use / When not to use

Empty interfaces are used when we want to store variables of different types
which don't implement a common interface.

There are two general use cases of an empty interface:

1. supported types are unknown (eg json.Marshal)
2. supported types are known (eg json.Token)

### Don't use if...

You should not use this checker for 1.
Sometimes we do not have prior knowledge about the expected types.
For example, json.Marshal(v interface{}) is designed to accept
structs of any type. This function uses reflect to gather information
it needs about the type of v.
In this case, it is not possible to list all the supported types.

### Use if...

You could consider using it, when all the types you support
are known at the type of writing your code.

This is particularly useful when the types are primitives (eg int),
where we have to create a new wrapper type (eg type Int int) and
implement a non-empty interface on it.

## Go2's type list

This tool is designed to work with code written in the current versions of Go (ie Go1).
The current design draft of Go2 includes the type list:

```go
type Numeric interface {
	type int, float64
}
```

At the moment, the type list is intended for function type parameters only.

```
interface type for variable cannot contain type constraints
```

The draft [notes](https://go.googlesource.com/proposal/+/refs/heads/master/design/go2draft-type-parameters.md#type-lists-in-interface-types):

> Interface types with type lists may only be used as constraints on type
> parameters. They may not be used as ordinary interface types. The same is true
> of the predeclared interface type comparable.
> 
> **This restriction may be lifted in future language versions. An interface type
> with a type list may be useful as a form of sum type, albeit one that can have
> the value nil**. Some alternative syntax would likely be required to match on
> identical types rather than on underlying types; perhaps type ==. For now, this
> is not permitted.

The highlight section is what this experiment addresses via an external type checking tool.

You might think of this tool as an experiment to see whether a sum type would be a valuable addition to the language.

## Implementation

This experiment is built on top of the dev.go2go branch and uses types.Sum.

## Limitations

- Only single line comments are implemented, ie `// #type T1, T2`
- Zero-value for an interface is nil.
  Several approaches come to mind:
  - allow nil values.
  - allow nil values, but fail if type switch statements don't include nil (what we do in this checker).
  - track all initializations/assignments/etc of the interfaces with types and fail if they are nil.
  - change the zero-value of an interface with a type list to be the zero-value of its first type (or some type chosen by the programmer).

## Discussion

You feedback about this experiment is appreciated.
Just ping me here or [@sinasiadat](https://twitter.com/sinasiadat) with a link to where you post it.
