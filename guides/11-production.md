# 11 — Using Rice in Production

This guide covers operational concerns for running Rice scripts in production environments: security hardening, resource control, concurrency, error handling, performance, and monitoring.

---

## Security Hardening

### Strict Mode — Whitelist Only What You Need

The safest configuration starts with an empty environment and adds only the packages your scripts require:

```go
cfg := conf.NewDefaultEnvConfig().
    SetStrictStdlibMode(true).
    EnableNamespacedPackage("math").
    EnableNamespacedPackage("strings").
    AddGlobalFunctionPackage(myApproved.Functions)

it := exec.NewInterpreter(cfg)
```

In strict mode, even `print`/`println` are absent unless you explicitly whitelist the `"io"` package. You can add your own audited I/O functions via [`AddGlobalFunctionPackage()`](../exec/conf/env.go).

### Disable Dangerous Packages

If you don't need full strict mode, selectively remove packages:

```go
// No I/O — scripts cannot print or log
cfg := conf.NewDefaultEnvConfig().
    DisableNamespacedPackage("io")

// No error throwing — scripts cannot call throw() or assert()
cfg.DisableNamespacedPackage("error")

// No type-bound string operations
cfg.DisableTypeBoundPackage(types.String)
```

**Security checklist** — consider disabling:

| Package | Risk if left enabled |
|---------|---------------------|
| `"io"` | Scripts can write to `LoggingOutput` (default: stdout) |
| `"error"` | Scripts can intentionally crash execution via `throw()` |
| `"json"` | Scripts can parse arbitrary JSON (CPU/memory for large input) |
| `"datetime"` | Low risk, but leaks wall-clock time |

### Removing All I/O

To completely silence script output, direct [`LoggingOutput`](../exec/conf/env.go) to `io.Discard` and disable the `"io"` package:

```go
cfg := conf.NewDefaultEnvConfig().
    SetLoggingOutput(io.Discard).
    DisableNamespacedPackage("io")
```

This prevents both script-requested output (`print`) and interpreter warnings from reaching the outside world.

---

## Concurrency & Thread Safety

The [`Interpreter`](../exec/interpreter_api.go) is **single-threaded** and guarded by a [`sync.Mutex`](../exec/interpreter_api.go:31). A single interpreter instance must not be shared across goroutines concurrently.

### Pattern 1: One Interpreter Per Goroutine

```go
go func() {
    it := exec.NewInterpreter(cfg)
    result, err := it.Interpret(ctx, ast, runCfg)
    // ...
}()
```

Each goroutine gets its own interpreter. This is the simplest approach and avoids contention.

### Pattern 2: Sync Pool (Many Scripts, Few Interpreters)

Reuse interpreter instances sequentially to amortize the cost of function compilation:

```go
type ScriptEngine struct {
    mu  sync.Mutex
    it  *exec.Interpreter
    cfg *conf.EnvConfig
}

func (e *ScriptEngine) Run(ctx context.Context, src string, runCfg *conf.RunConfig) (types.Value, error) {
    // Tokenize + Parse (could be cached separately)
    tokens, err := frontend.Tokenize(src)
    if err != nil {
        return nil, err
    }
    parser := frontend.NewParser(tokens)
    ast := parser.Parse()
    if len(parser.Errors()) > 0 {
        return nil, fmt.Errorf("parse: %v", parser.Errors()[0])
    }

    e.mu.Lock()
    defer e.mu.Unlock()
    return e.it.Interpret(ctx, ast, runCfg)
}
```

The mutex serializes access. For higher throughput, use a pool of interpreters behind a channel.

---

## Resource Control

### Timeouts

Two independent timeout layers protect against runaway scripts:

| Config | Scope | Default | Controls |
|--------|-------|---------|----------|
| [`NativeFuncTimeout`](../exec/conf/env.go) | Per native/built-in call | 1s | `SetNativeFuncTimeout()` |
| [`UserFuncTimeout`](../exec/conf/run.go) | Per user-defined function call | 2s | `SetUserFuncTimeout()` |

```go
// Tight timeouts for a public-facing service
cfg := conf.NewDefaultEnvConfig().
    SetNativeFuncTimeout(500 * time.Millisecond)  // native calls must finish fast

runCfg := conf.NewDefaultRunConfig().
    SetUserFuncTimeout(1 * time.Second)           // user funcs get 1s max
```

Timeouts are enforced per-call, not per-script. A single script execution can span many function calls, each subject to its own timeout.

### Lexical Scope Limit

The [`LexicalScopeLimit`](../exec/conf/run.go) prevents deeply nested blocks from exhausting memory:

```go
runCfg := conf.NewDefaultRunConfig().
    SetLexicalScopeLimit(256)  // default; lower for stricter control
```

Scripts that exceed the limit produce an error.

### Call-Stack Frames

[`PreAllocatedFrames`](../exec/conf/env.go) controls initial capacity, not a hard limit. Set it based on expected recursion depth:

```go
cfg := conf.NewDefaultEnvConfig().
    SetPreAllocatedFrames(64)  // for scripts with deeper call chains
```

### Context Cancellation

Pass a [`context.Context`](../exec/interpreter_api.go) with a deadline to enforce an overall execution budget:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := it.Interpret(ctx, ast, runCfg)
// Context deadline exceeded returns an error
```

---

## Error Handling in Production

### Structured Error Extraction

[`RuntimeError`](../exec/runtime_error.go) contains position information and a causal chain. Log structured data for observability:

```go
result, err := it.Interpret(ctx, ast, runCfg)
if err != nil {
    var re exec.RuntimeError
    if errors.As(err, &re) {
        log.Printf("script error: %s", re.Error())
        // For debugging (not production logging by default):
        if os.Getenv("DEBUG") != "" {
            log.Print(re.Stacktrace())
        }
    }
    return nil, err
}
```

### Don't Leak Stack Traces to Users

Stack traces include internal details. Only return the error message to callers and log the full trace server-side.

### Distinguish Script Errors from System Errors

Script-level errors (syntax, type mismatch, assertion failures) return `RuntimeError`. System-level errors (context deadline exceeded, nil pointer in native functions) may return different error types. Handle both:

```go
if err != nil {
    var re exec.RuntimeError
    if errors.As(err, &re) {
        // script error — may be user-caused, return as-is or sanitize
        return nil, fmt.Errorf("script error: %w", err)
    }
    // system error — log at higher severity
    log.Printf("internal error: %v", err)
    return nil, fmt.Errorf("internal error")
}
```

### Stream Mode for Partial Execution

If a script has multiple independent statements and one fails, stream mode lets you continue:

```go
it.InterpretStream(ctx, seq, runCfg,
    func(val types.Value, err error) bool {
        if err != nil {
            log.Printf("statement %d failed: %v", idx, err)
            return true  // continue to next statement
        }
        return true
    },
)
```

---

## Performance

### Cache the AST

Tokenizing and parsing are pure functions — cache the AST and reuse it across executions:

```go
type CachedScript struct {
    ast      []ast.Stmt
    runCfg   *conf.RunConfig
    parseErr error
}

var cache sync.Map // map[string]*CachedScript

func LoadScript(name, src string) (*CachedScript, error) {
    if v, ok := cache.Load(name); ok {
        return v.(*CachedScript), nil
    }
    tokens, err := frontend.Tokenize(src)
    if err != nil {
        return nil, err
    }
    parser := frontend.NewParser(tokens)
    ast := parser.Parse()
    if len(parser.Errors()) > 0 {
        return nil, fmt.Errorf("parse: %v", parser.Errors()[0])
    }
    cs := &CachedScript{ast: ast, runCfg: conf.NewDefaultRunConfig()}
    cache.Store(name, cs)
    return cs, nil
}
```

Parse once, interpret many times. This is the biggest performance win.

### Reuse a Single Interpreter

Creating an interpreter compiles all native functions into lookup tables — an O(n) cost that pays off across many executions. Hold one interpreter for the lifetime of your application:

```go
var engine = exec.NewInterpreter(conf.NewDefaultEnvConfig())

func HandleRequest(w http.ResponseWriter, r *http.Request) {
    cs, _ := LoadScript("greet", `return "Hello, " + name + "!";`)
    cs.runCfg.DefineVariable("name", values.String(r.URL.Query().Get("name")))
    engine.mu.Lock()
    result, err := engine.it.Interpret(r.Context(), cs.ast, cs.runCfg)
    engine.mu.Unlock()
    // ...
}
```

### Profiling in Development

Enable the profiler during load testing to identify slow functions:

```go
it := exec.NewInterpreter(conf.NewDefaultEnvConfig().EnableProfiling())
// ... run several scripts ...
fmt.Println(it.Profiler().Report())
// Prints per-frame timing: each function call's cumulative and self time
```

Reset between measurements: `it.Profiler().Reset()`

### Performance Characteristics

Based on the benchmark suite in [`exec/bench_test.go`](../exec/bench_test.go):

- Simple arithmetic and comparisons: sub-microsecond
- Recursive calls (fibonacci): ~50µs for fib(20)
- Collection operations (list/map/set): proportional to element count
- Native function calls: overhead of reflection dispatch (~1-2µs)

The interpreter is a tree-walking evaluator — it is not JIT-compiled. For throughput-sensitive workloads, push heavy logic into native Go functions and keep Rice scripts lightweight (configuration, DSL rules, simple transformations).

---

## Monitoring & Observability

### Log Interpreter Warnings

The [`LoggingOutput`](../exec/conf/env.go) writer receives conflict warnings when custom packages override stdlib functions. In production, capture this:

```go
var buf bytes.Buffer
cfg := conf.NewDefaultEnvConfig().
    SetLoggingOutput(&buf)
// ... after execution, check buf for warnings
if buf.Len() > 0 {
    log.Printf("interpreter warnings: %s", buf.String())
}
```

### Track Execution Count & Errors

Wrap the interpreter call with metrics:

```go
var (
    scriptExecutions = prometheus.NewCounter(...)
    scriptErrors     = prometheus.NewCounter(...)
    scriptDuration   = prometheus.NewHistogram(...)
)

start := time.Now()
result, err := it.Interpret(ctx, ast, runCfg)
scriptDuration.Observe(time.Since(start).Seconds())
scriptExecutions.Inc()
if err != nil {
    scriptErrors.Inc()
}
```

### Avoid Logging Script Output in Production

Scripts can call `print()` which writes to `LoggingOutput`. If you allow `"io"` in production, redirect it:

```go
cfg := conf.NewDefaultEnvConfig().
    SetLoggingOutput(buf)  // capture, don't let scripts pollute stdout
```

Or disable `"io"` entirely and provide your own audited output functions via custom packages.

---

## Deployment Checklist

- [ ] **Use `StrictStdlibMode`** or explicitly disable unused packages
- [ ] **Set timeouts**: `NativeFuncTimeout` ≤ 1s, `UserFuncTimeout` ≤ 2s
- [ ] **Set a context deadline** on every `Interpret()` call
- [ ] **Set `LexicalScopeLimit`** to 128–256
- [ ] **Direct `LoggingOutput`** away from stdout in production
- [ ] **Cache parsed ASTs** — parse once, run many times
- [ ] **Use one interpreter per goroutine** or serialize with a mutex
- [ ] **Log stack traces server-side only**, never return them to callers
- [ ] **Distinguish script errors from system errors**
- [ ] **Benchmark with `EnableProfiling()`** before deploying
- [ ] **Push heavy logic into native Go functions**, keep scripts lightweight

---

**Previous:** [10 — Custom Packages](10-custom-packages.md)
