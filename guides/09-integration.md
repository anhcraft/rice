# 09 — Embedding Rice in Go

This guide covers how to embed the Rice scripting engine into your Go application: parsing scripts, configuring the interpreter, executing code, and handling errors.

## Overview

The integration pipeline has three stages:

```
Script (.rice) ──► Tokenize ──► Parse ──► Interpret
```

| Stage | Package | Input | Output |
|-------|---------|-------|--------|
| Tokenize | [`frontend.Tokenize()`](../frontend/tokenizer.go) | `string` | `[]Token, error` |
| Parse | [`frontend.NewParser()`](../frontend/parser.go) + `.Parse()` | `[]Token` | `[]ast.Stmt` |
| Interpret | [`(*Interpreter).Interpret()`](../exec/interpreter_api.go) | `[]ast.Stmt` + config | `types.Value, error` |

## Minimal Example

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/anhcraft/rice/exec"
    "github.com/anhcraft/rice/exec/conf"
    "github.com/anhcraft/rice/frontend"
)

func main() {
    script := `print("Hello from Rice!")`

    // Step 1: Tokenize
    tokens, err := frontend.Tokenize(script)
    if err != nil {
        panic(err)
    }

    // Step 2: Parse
    parser := frontend.NewParser(tokens)
    ast := parser.Parse()
    if len(parser.Errors()) > 0 {
        panic(parser.Errors()[0])
    }

    // Step 3: Create interpreter
    it := exec.NewInterpreter(conf.NewDefaultEnvConfig())

    // Step 4: Run
    result, err := it.Interpret(
        context.Background(),
        ast,
        conf.NewDefaultRunConfig(),
    )
    if err != nil {
        panic(err)
    }

    fmt.Printf("Result: %v\n", result)
}
```

## Reading from a File

The typical pattern reads a [`.rice`](../examples/tutorial.rice) file from disk:

```go
scriptBytes, err := os.ReadFile("./script.rice")
if err != nil {
    panic(err)
}
script := string(scriptBytes)

tokens, err := frontend.Tokenize(script)
// ... parse and interpret as above
```

---

## EnvConfig — Interpreter-Level Configuration

[`EnvConfig`](../exec/conf/env.go) is passed to [`NewInterpreter()`](../exec/interpreter_api.go) and applies globally across all executions on that interpreter instance.

### Default Configuration

```go
cfg := conf.NewDefaultEnvConfig()
// Equivalent to:
// &EnvConfig{
//     LoggingOutput:     os.Stdout,
//     NativeFuncTimeout: 1 * time.Second,
//     PreAllocatedFrames: 8,
// }
```

### Available Options

| Method | Description |
|--------|-------------|
| `SetNativeFuncTimeout(d)` | Timeout for native/built-in function calls (default: 1s) |
| `SetPreAllocatedFrames(n)` | Number of call-stack frames pre-allocated (default: 8) |
| `EnableProfiling()` | Enable the built-in profiler |
| `SetLoggingOutput(w)` | Where warnings and logs are written (default: `os.Stdout`) |

```go
cfg := conf.NewDefaultEnvConfig().
    SetNativeFuncTimeout(5 * time.Second).
    SetPreAllocatedFrames(16).
    EnableProfiling().
    SetLoggingOutput(os.Stderr)
```

### Profiling

When profiling is enabled via [`EnableProfiling()`](../exec/conf/env.go), you can retrieve timing data after execution:

```go
it := exec.NewInterpreter(conf.NewDefaultEnvConfig().EnableProfiling())
// ... run scripts ...
fmt.Println(it.Profiler().Report())
```

---

## RunConfig — Per-Execution Configuration

[`RunConfig`](../exec/conf/run.go) is passed to each [`Interpret()`](../exec/interpreter_api.go) call and applies only to that single execution.

### Default Configuration

```go
cfg := conf.NewDefaultRunConfig()
// Equivalent to:
// &RunConfig{
//     UserFuncTimeout:   2 * time.Second,
//     LexicalScopeLimit: 256,
// }
```

### Available Options

| Method | Description |
|--------|-------------|
| `SetUserFuncTimeout(d)` | Timeout for user-defined function calls (default: 2s) |
| `SetLexicalScopeLimit(n)` | Maximum lexical scope depth (default: 256) |
| `DefineVariable(key, value)` | Pre-define a mutable variable available in the script |
| `DefineConstant(key, value)` | Pre-define an immutable constant available in the script |

### Injecting Values from Go

You can pass Go values into the Rice script before execution:

```go
runCfg := conf.NewDefaultRunConfig().
    DefineConstant("NUM_RECORDS", values.Int(10000)).
    DefineConstant("CATEGORIES", values.ListOf([]values.String{
        "Electronics", "Books", "Home Goods",
    })).
    DefineVariable("counter", values.Int(0))

result, err := it.Interpret(ctx, ast, runCfg)
```

The script can then reference these:

```rice
print(NUM_RECORDS);       # 10000
counter = counter + 1;    # mutable variable
```

**Go-to-Rice type mapping:**

| Go type | [`values`](../exec/types/values/) constructor |
|---------|------------------|
| `int64` | `values.Int(42)` |
| `float64` | `values.Float(3.14)` |
| `bool` | `values.Bool(true)` |
| `string` | `values.String("hello")` |
| `[]T` (list) | `values.ListOf([]T{...})` |
| Map | `values.NewMap()` + `.Put()` |
| Set | `values.NewSet()` + `.Add()` |

---

## Error Handling

Interpretation errors are returned as [`RuntimeError`](../exec/runtime_error.go), which includes a stack trace:

```go
result, err := it.Interpret(ctx, ast, runCfg)
if err != nil {
    var re exec.RuntimeError
    if errors.As(err, &re) {
        fmt.Println("Stack trace:")
        fmt.Println(re.Stacktrace())
    } else {
        fmt.Println("Error:", err)
    }
    return
}
```

### Error Recovery in Stream Mode

[`InterpretStream()`](../exec/interpreter_api.go) accepts an `errorRecover` callback that lets you decide whether to continue after each statement fails:

```go
result, err := it.InterpretStream(ctx, seq, runCfg,
    func(val types.Value, err error) bool {
        if err != nil {
            fmt.Println("Recoverable error:", err)
            return true // continue to next statement
        }
        return true
    },
)
```

---

## Type Conversion Between Go and Rice

Rice values live in the [`values`](../exec/types/values/) package and implement [`types.Value`](../exec/types/types.go). You need to convert between Go native types and Rice values at both boundaries: injecting data into scripts and extracting results.

### Go → Rice: Constructors

| Go type | Rice constructor | Example |
|---------|-----------------|---------|
| `int64` | [`values.Int(42)`](../exec/types/values/int.go) | `values.Int(10000)` |
| `float64` | [`values.Float(3.14)`](../exec/types/values/float.go) | `values.Float(2.5)` |
| `bool` | [`values.Bool(true)`](../exec/types/values/bool.go) | `values.Bool(false)` |
| `string` | [`values.String("hi")`](../exec/types/values/string.go) | `values.String("hello")` |
| `[]T` (list) | [`values.ListOf()`](../exec/types/values/list.go) | `values.ListOf([]values.String{"a","b"})` |
| Map (from entries) | [`values.NewMap()`](../exec/types/values/map.go) + `.Put()` | Construct and populate |
| Set (from entries) | [`values.NewSet()`](../exec/types/values/set.go) + `.Add()` | Construct and populate |

All the primitive types ([`Int`](../exec/types/values/int.go), [`Float`](../exec/types/values/float.go), [`Bool`](../exec/types/values/bool.go), [`String`](../exec/types/values/string.go)) are simple type aliases — just cast the Go native type.

### Rice → Go: Manual Type-Switch

When a script returns a value, type-switch to extract the Go native:

```go
switch v := result.(type) {
case values.Int:
    goInt := int64(v)
case values.Float:
    goFloat := float64(v)
case values.String:
    goStr := string(v)
case values.Bool:
    goBool := bool(v)
case *values.List:
    // iterate via v.Iterate() or access via v.Element(idx)
case *values.Map:
    // iterate via v.Iterate() or access via v.Element(key)
case *values.Set:
    // iterate via v.Iterate()
}
```

### Rice → Go: The `util.Unwrap()` Helper

For convenience, the [`util`](../util/unwrapper.go) package provides [`util.Unwrap()`](../util/unwrapper.go) which recursively converts any Rice value to a plain Go `any`:

```go
import "github.com/anhcraft/rice/util"

raw, err := util.Unwrap(result)
if err != nil {
    // handle
}
fmt.Printf("%#v\n", raw)
```

Conversion rules:

| Rice type | Go type |
|-----------|---------|
| `Int` | `int64` |
| `Float` | `float64` |
| `Bool` | `bool` |
| `String` | `string` |
| `List` / `Set` | `[]any` (recursively unwrapped) |
| `Map` / `Namespace` | `map[string]any` (keys stringified) |
| `Func` / `NativeFuncSet` | `nil` |
| `Identifier` | `string` |
| `Selector` | Unwrapped value of the selector target |

```go
# Rice script returns: map.of("name", "Alice", "scores", list.of(85, 92))
raw, _ := util.Unwrap(result)
// raw = map[string]any{"name": "Alice", "scores": []any{int64(85), int64(92)}}
```

This is useful for passing results directly to JSON serialization, logging, or other Go APIs that accept `interface{}`.

---

**Previous:** [08 — Standard Library](08-standard-library.md)  
**Next:** [10 — Custom Packages](10-custom-packages.md)
