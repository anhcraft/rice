# Rice Language Guide

Rice is a minimal embeddable scripting language for Go applications. The language file extension is `.rice`.

## Table of Contents

| Lesson | Topic |
|--------|-------|
| [00](00-introduction.md) | Introduction — What is Rice, file extension, comments, basic syntax |
| [01](01-data-types.md) | Data Types — Primitives, composites, Unicode codepoints |
| [02](02-variables-and-declarations.md) | Variables & Declarations — `var`, `const`, assignment, block scope, shadowing |
| [03](03-operators.md) | Operators — Unary, binary, comparison, logical, implicit conversion, spread |
| [04](04-collections.md) | Collections — `string`, `list`, `set`, `map`; traits, element access, iteration, selector |
| [05](05-control-flow.md) | Control Flow — `if`/`else` expression, `for` loops, `for..in`, `break`/`continue`/`return` |
| [06](06-functions.md) | Functions — Native vs user-defined, type-bound functions, closures, varargs |
| [07](07-functional-programming.md) | Functional Programming — `map`, `filter`, `sort`, chaining, lambdas |
| [08](08-standard-library.md) | Standard Library — All built-in namespaced and type-bound functions |
| [09](09-integration.md) | Embedding in Go — Tokenize, parse, interpret, EnvConfig, RunConfig, value passing |
| [10](10-custom-packages.md) | Custom Packages — Writing native functions, package registration, disable/strict/override |
| [11](11-production.md) | Production Guide — Security hardening, resource control, concurrency, error handling, performance |

## Quick Start

Create a file ending in [`.rice`](../examples/tutorial.rice):

```rice
# This is a comment
const name = "Rice";
println("Hello from " + name + "!");

# Everything is an expression
var result = if 5 > 3 { "yes" } else { "no" };
print(result); # yes
```

## Examples & Tests

| Directory | Description |
|-----------|-------------|
| [`examples/`](../examples/) | Complete Rice scripts: [tutorial](../examples/tutorial.rice), [functional programming](../examples/functional.rice), [data processing](../examples/process-dataset.rice), [insert interval](../examples/insert-interval.rice) |
| [`examples/example.go`](../examples/example.go) | Full Go integration example (tokenize → parse → interpret with profiler) |
| [`exec/testdata/`](../exec/testdata/) | 50+ language feature test scripts — one per concept (e.g. [closures](../exec/testdata/closures.rice), [spread](../exec/testdata/spread.rice), [varargs](../exec/testdata/varargs.rice)) |
| [`exec/testdata/grind75/`](../exec/testdata/grind75/) | 21 algorithmic problems solved in Rice (Two Sum, Valid Parentheses, Binary Search, Flood Fill, etc.) |
| [`exec/interpreter_api_test.go`](../exec/interpreter_api_test.go) | Go test suite including custom package integration tests |
