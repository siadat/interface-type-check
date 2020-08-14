# Static type checker for interface{} with a type list

This is an experiment.

* This is a tool that performs a static type check on values of type interface{}.
* You specify the types in a special comment, eg `// #type T1, T2`
* Internally, the implementation is based on go2go's Sum type.
  Go2go and this experiment have different concerns: go2go is about generic functions and type parameters,
  this experiment is about sum types.
* For an alternative solution see the [siadat/group-interface](https://github.com/siadat/group-interface) pattern.

## Demo

Without type checking:

```go
/*    */  package main
/*    */
/*    */  type Numeric interface{}
/*    */  
/*    */  func main() {
/*    */  	var n Numeric
/* OK */  	n = 3
/* OK */  	n = 3.14
/* OK */  	n = "abcd"
/*    */  	_ = n
/*    */  }
```

With type checking:

```go
/*     */  package main
/*     */
/*     */  type Numeric interface{
/* --> */  	// #type int, float64
/*     */  }
/*     */  
/*     */  func main() {
/*     */  	var n Numeric
/* OK  */  	n = 3
/* OK  */  	n = 3.14
/* ERR */  	n = "abcd"
/*     */  	_ = n
/*     */  }
```

Execute the checker to get the error:

```bash
$ interface-type-check .
testfile.go:10:6: cannot use "bad value" (constant of type string) as Numeric value in variable declaration: mismatching sum type (have string, want a type in interface{type int, float64})
```
## Download

Prebuilt binaries are available here as well as in the [release page](https://github.com/siadat/interface-type-check/releases/tag/v0.0.0).

* Darwin: [download](https://github.com/siadat/interface-type-check/releases/download/v0.0.0/interface-type-check.darwin-amd64.tar.gz) (2.9 MB)
* Linux: [download](https://github.com/siadat/interface-type-check/releases/download/v0.0.0/interface-type-check.linux-amd64.tar.gz) (2.96 MB)
* Windows: [download](https://github.com/siadat/interface-type-check/releases/download/v0.0.0/interface-type-check.windows-amd64.tar.gz) (3.01 MB)

## Build

```bash
git clone https://github.com/siadat/interface-type-check
cd interface-type-check
make test build
```

## Checks

Given the declaration:

```go
type Numeric interface{
	// #type int, float64
}
```

The following checks are performed:

```go
var number Numeric = "abc" // CHECK ERR: expected int or float
```


```go
_, _ = number.(string)     // CHECK ERR: string not allowed
```


```go
switch number.(type) {
case string:               // CHECK ERR: string not allowed
case float:
}
```


```go
switch number.(type) {     // CHECK ERR: missing case for int
case float:
}
```

```go
switch number.(type) {     // CHECK ERR: missing case for nil
case float:
case int:
}
```

More examples: fork/[src/types/examples/sum.go2](https://github.com/siadat/go/blob/interface-type-check/src/go/types/examples/sum.go2)

<!--
### src/cmd/compile/internal/ssa/gen/rulegen.go Node and Statement
### database/sql/driver.Value
### plugin.Symbol
### xml.Token
-->

## Experiment: json.Token

All supported types of encoding/json.Token are known,
as documented [here](https://pkg.go.dev/encoding/json?tab=doc#Token):

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
type Token interface {
	// #type Delim, bool, float64, Number, string
}
```

That's all we need to be able to use the checker.

## Experiment: sql.Scanner

database/sql.Scanner is also [defined](https://pkg.go.dev/database/sql?tab=doc#Scanner)
as an empty interface whose possible types are known.

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
	// #type int64, float64, bool, []byte, string, time.Time
}
```


<!--
## Experiment: error handling

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
-->


## Experiment: net.IP

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

## Experiment: a hypothetical connection object

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

## When to use / When not to use

Empty interfaces are used when we want to store variables of different types
which don't implement a common interface.

There are two general use cases of an empty interface:

1. supported types are unknown (eg json.Marshal)
2. supported types are known (eg json.Token)

### Don't use if:

You should not use this checker for 1.
Sometimes we do not have prior knowledge about the expected types.
For example, json.Marshal(v interface{}) is designed to accept
structs of any type. This function uses reflect to gather information
it needs about the type of v.
In this case, it is not possible to list all the supported types.

### Use if:

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

- This experiment is built on top of the dev.go2go branch and uses types.Sum. See [diff](https://github.com/siadat/go/commit/af8a19e4de0c689be9d898d7ca3b0b5fd51767cb).
- A few more test examples are added to [types/examples](https://github.com/siadat/go/commit/af8a19e4de0c689be9d898d7ca3b0b5fd51767cb#diff-4204251ae72f5797f67f5f4393ab8c10).

## Limitations

- Only single line comments are implemented, ie `// #type T1, T2`
- Zero-value for an interface is nil.
  Several approaches come to mind:
  - allow nil values.
  - allow nil values, but fail if type switch statements don't include nil (what we do in this checker).
  - track all initializations/assignments/etc of the interfaces with types and fail if they are nil.
  - change the zero-value of an interface with a type list to be the zero-value of its first type (or some type chosen by the programmer).

## Contribute

Do any of these:

- Download a binary, or build from source.
- Report issues. You will most likely run into problems, because this is a new project.
- Use it! Let me know what you use it for.
- Search for TODOs in the code.
- Implement missing features.

<!-- https://github.com/golang/go/compare/dev.go2go...siadat:interface-type-check -->

