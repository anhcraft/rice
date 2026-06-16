# 10 — Custom Packages

This guide covers defining your own native functions in Go and registering them with the Rice interpreter, as well as controlling which standard library packages are available.

## Core Types

The key types for defining functions are in [`exec/fun/`](../exec/fun/fun_index.go):

| Type | Definition |
|------|------------|
| [`FunctionImpl`](../exec/fun/fun_index.go) | `[]*FunctionDef` — a list of overloaded definitions for one function name |
| [`FunctionPackage`](../exec/fun/fun_index.go) | `map[values.Identifier]FunctionImpl` — a map from function name to its overloads |

A package is simply a map of function names to implementations.

## Defining a Simple Namespaced Package

Create a Go package that exports a `var Functions fun.FunctionPackage`:

```go
package tools

import (
    "github.com/anhcraft/rice/exec/fun"
    "github.com/anhcraft/rice/exec/stdlib"
    "github.com/anhcraft/rice/exec/types"
    "github.com/anhcraft/rice/exec/types/values"
)

var Functions = fun.FunctionPackage{
    "greet": {stdlib.Define(Greet)},
    "add":   {stdlib.Define(Add)},
}

func Greet(name values.String) (types.Value, error) {
    return values.String("Hello, " + string(name) + "!"), nil
}

func Add(a, b values.Int) (types.Value, error) {
    return values.Int(int(a) + int(b)), nil
}
```

### Function Signature Rules

Every native function must follow this pattern:

```go
func MyFunc(arg1 ArgType1, arg2 ArgType2, ...) (types.Value, error)
```

- **Parameters**: Use the concrete value types from [`values`](../exec/types/values/) — e.g., [`values.Int`](../exec/types/values/int.go), [`values.String`](../exec/types/values/string.go), [`values.Float`](../exec/types/values/float.go), [`values.Bool`](../exec/types/values/bool.go), `*values.List`, `*values.Map`, `*values.Set`, or `types.Value` for `any`
- **Return**: Always `(types.Value, error)`. Return `nil, nil` for functions with no meaningful return value
- **Varargs**: Use Go's native variadic syntax — the last parameter can be `...T`

### [`stdlib.Define()`](../exec/stdlib/util.go)

Wraps [`fun.ScanFunction()`](../exec/fun/function_def.go) which uses reflection to inspect the Go function signature and build a [`FunctionDef`](../exec/fun/function_def.go). It panics on error (appropriate for init-time registration).

### [`stdlib.DefineAndMap()`](../exec/stdlib/util.go)

Same as `Define()`, but accepts a callback to modify the [`FunctionDef`](../exec/fun/function_def.go) after scanning:

```go
"greet": {stdlib.DefineAndMap(Greet, func(def *fun.FunctionDef) {
    // Modify def here, e.g., widen parameter types
})},
```

## Registering the Package

Register your package when creating the interpreter:

```go
it := exec.NewInterpreter(
    conf.NewDefaultEnvConfig().
        AddNamespacedFunctionPackage("tools", tools.Functions),
)
```

The script can then call:

```rice
tools.greet("World")   # "Hello, World!"
tools.add(1, 2)        # 3
```

### Adding Functions to the Global Namespace

Use [`AddGlobalFunctionPackage()`](../exec/conf/env.go) (which maps to namespace `""`) to add functions without a prefix:

```go
conf.NewDefaultEnvConfig().
    AddGlobalFunctionPackage(tools.Functions)
```

```rice
greet("World")   # "Hello, World!" — no prefix needed
```

---

## Varargs

Variadic functions accept zero or more trailing arguments. In Rice, variadic Go functions use Go's native `...T` syntax.

### Simple Varargs with `stdlib.Define()`

For the common case — variadic `any` (`...types.Value`) or a single concrete type (`...values.Int`) — [`stdlib.Define()`](../exec/stdlib/util.go) handles everything automatically via reflection:

```go
var Functions = fun.FunctionPackage{
    // Variadic any: accepts any number of any-type arguments
    "sum": {stdlib.Define(Sum)},
}

func Sum(nums ...types.Value) (types.Value, error) {
    var total values.Int
    for _, n := range nums {
        if v, ok := n.(values.Int); ok {
            total += v
        }
    }
    return total, nil
}
```

```rice
tools.sum(1, 2, 3, 4)   # 10
tools.sum()               # 0
```

This also works with a single concrete variadic type:

```go
"join": {stdlib.Define(Join)},

func Join(strs ...values.String) (types.Value, error) { ... }
```

```rice
tools.join("a", "b", "c")  # works — all args must be String
```

**How it works**: [`fun.ScanFunction()`](../exec/fun/function_def.go) uses Go reflection to detect `tp.IsVariadic()`. For the last parameter `...T`, it reads `[]T` from `tp.In()`, determines the base type and dimension via [`getArgType()`](../exec/fun/util.go), and sets the variadic flag.

| Go parameter | Inferred Rice type |
|---|---|
| `...types.Value` | Variadic `any` |
| `...values.Int` | Variadic `Int` |
| `...values.String` | Variadic `String` |

### Custom Varargs with `stdlib.DefineAndMap()`

When you need a **union type** for the variadic parameter — e.g., accepting `...Int|Float|Bool` instead of just `...any` — use [`stdlib.DefineAndMap()`](../exec/stdlib/util.go) to redefine the last argument's accepted types:

```go
"printNums": {stdlib.DefineAndMap(PrintNums, func(def *fun.FunctionDef) {
    // The variadic parameter is the last one (index = SizeOfArgs() - 1)
    lastIdx := def.SizeOfArgs() - 1
    def.DefineArg(lastIdx,
        fun.NewArgType(0, types.Int),
        fun.NewArgType(0, types.Float),
        fun.NewArgType(0, types.Bool),
    )
})},

func PrintNums(vals ...types.Value) (types.Value, error) {
    for _, v := range vals {
        // Only Int, Float, Bool will be passed due to the type guard above
        println(v)
    }
    return nil, nil
}
```

```rice
tools.printNums(1, 3.14, true)
# TypeError if you pass a String — the union excludes it
```

The key difference from `Define`:
- `Define` relies entirely on reflection — the Rice signature mirrors the Go types exactly
- `DefineAndMap` lets you **broaden** (widen to union) or **narrow** (restrict from `any`) each parameter's accepted types after reflection

This is the same pattern used throughout the standard library, e.g. [`math.abs`](../exec/stdlib/math/math.go) redefines its parameter from `any` to `Int|Float|Bool`.

---

## Overloading

Multiple Go functions can be registered under the same Rice function name. The interpreter dispatches to the correct one based on argument types at call time:

```go
var Functions = fun.FunctionPackage{
    "describe": {
        stdlib.Define(DescribeInt),     // for Int arguments
        stdlib.Define(DescribeString),  // for String arguments
    },
}

func DescribeInt(v values.Int) (types.Value, error) {
    return values.String("an integer"), nil
}

func DescribeString(v values.String) (types.Value, error) {
    return values.String("a string"), nil
}
```

```rice
tools.describe(42)       # "an integer"
tools.describe("hello")  # "a string"
```

---

## Contextual Functions

If your function's first parameter is `context.Context`, it receives the interpreter's context (including timeout):

```go
func FetchData(ctx context.Context, url values.String) (types.Value, error) {
    // ctx carries the native function timeout from EnvConfig.NativeFuncTimeout
    // use it for HTTP calls, DB queries, etc.
    return values.String("data from " + string(url)), nil
}
```

---

## Type-Bound Custom Functions

To add methods callable via dot syntax (e.g., `myValue.myMethod()`), use [`AddTypeBoundFunctionPackage()`](../exec/conf/env.go):

```go
// Define functions that take the bound type as their first argument
var MyTypeBoundFuncs = fun.FunctionPackage{
    "double": {stdlib.Define(func(v values.Int) (types.Value, error) {
        return values.Int(int(v) * 2), nil
    })},
}

cfg := conf.NewDefaultEnvConfig().
    AddTypeBoundFunctionPackage(types.Int, MyTypeBoundFuncs)
```

```rice
var x = 21;
x.double()   # 42
```

The transformation rule is: `V.f(...)` ⟺ `f(V, ...)`. The first parameter of the Go function receives the bound value.

---

## Package Management

The standard library is organized as individually-addressable sub-packages. You can disable, override, or whitelist them.

### Disabling Standard Packages

Remove specific stdlib packages from the environment:

```go
// Remove I/O functions (print, printf, println, printlnf)
cfg := conf.NewDefaultEnvConfig().
    DisableNamespacedPackage("io")

// Remove error functions (assert, throw)
cfg.DisableNamespacedPackage("error")

// Remove type-bound string methods (.toUpper, .toLower, etc.)
cfg.DisableTypeBoundPackage(types.String)
```

| Standard Package IDs | Functions |
|----------------------|-----------|
| `"io"` | `print`, `printf`, `println`, `printlnf` (global namespace) |
| `"error"` | `assert`, `throw` (global namespace) |
| `"type"` | `typeof`, `float`, `bool`, `int`, `string`, `isNumberLike`, `isNumber`, `len` (global namespace) |
| `"strings"` | All `strings.*` functions |
| `"math"` | All `math.*` functions |
| `"list"` | All `list.*` functions |
| `"set"` | All `set.*` functions |
| `"map"` | All `map.*` functions |
| `"datetime"` | `datetime.now` |
| `"json"` | `json.encode`, `json.decode` |

**Decomposed global namespace**: The `""` (global) namespace is split into tracked sub-packages `"io"`, `"error"`, and `"type"`. Disabling `"io"` does not affect `"error"` or `"type"`.

### Strict Mode (Whitelist)

For security-sensitive or DSL environments, enable strict mode to allow only explicitly-listed packages:

```go
cfg := conf.NewDefaultEnvConfig().
    SetStrictStdlibMode(true).
    EnableNamespacedPackage("math").
    EnableNamespacedPackage("list").
    AddGlobalFunctionPackage(myCore.Functions)
```

Only `math.*`, `list.*`, and your custom functions are available. Everything else (including `io`, `strings`, etc.) is absent.

### Override with Conflict Detection

When a custom package defines a function that collides with a standard one, a warning is logged by default. Use [`OverrideNamespacedPackage()`](../exec/conf/env.go) to suppress the warning and explicitly replace:

```go
cfg := conf.NewDefaultEnvConfig().
    DisableNamespacedPackage("io").
    AddNamespacedFunctionPackage("", customIO.Functions).
    OverrideNamespacedPackage("io")
```

---

## Complete Example: Adding a "tools" Package

See [`examples/example.go`](../examples/example.go) for the full integration pattern. Here's a self-contained example:

```go
package main

import (
    "context"
    "fmt"

    "github.com/anhcraft/rice/exec"
    "github.com/anhcraft/rice/exec/conf"
    "github.com/anhcraft/rice/exec/fun"
    "github.com/anhcraft/rice/exec/stdlib"
    "github.com/anhcraft/rice/exec/types"
    "github.com/anhcraft/rice/exec/types/values"
    "github.com/anhcraft/rice/frontend"
)

// Define your custom package
var toolsPkg = fun.FunctionPackage{
    "greet": {stdlib.Define(func(name values.String) (types.Value, error) {
        return values.String("Hello, " + string(name) + "!"), nil
    })},
    "multiply": {stdlib.Define(func(a, b values.Float) (types.Value, error) {
        return values.Float(float64(a) * float64(b)), nil
    })},
}

func main() {
    script := `print(tools.greet("Rice")); print(tools.multiply(3, 7))`

    tokens, _ := frontend.Tokenize(script)
    parser := frontend.NewParser(tokens)
    ast := parser.Parse()

    it := exec.NewInterpreter(
        conf.NewDefaultEnvConfig().
            AddNamespacedFunctionPackage("tools", toolsPkg),
    )

    result, err := it.Interpret(context.Background(), ast, conf.NewDefaultRunConfig())
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Printf("Result: %v\n", result)
}
```

Output:
```
Hello, Rice!
21
Result: 21
```

---

**Previous:** [09 — Embedding Rice in Go](09-integration.md)
**Next:** [11 — Production Guide](11-production.md)
