# 00 — Introduction

## What is Rice?

Rice is a minimal embeddable scripting language designed for Go applications. It provides a familiar C-like syntax with dynamic typing, first-class functions, and a rich standard library — all while remaining small and easy to embed.

Script files use the [`.rice`](../examples/tutorial.rice) extension.

## Hello World

```rice
print("Hello World!");
```

## Comments

Rice supports **single-line comments** only. Prefix a line with `#`:

```rice
# This is a comment
var x = 10; # inline comment after code
```

There are no multi-line comment blocks (no `/* ... */`). Use multiple `#` lines when you need longer explanations.

## Statements and Expressions

Rice distinguishes between **statements** (no return value) and **expressions** (return a value). A key difference from many languages is that control-flow constructs like `if` are **expressions** — they produce a value.

```rice
# if-as-expression: the whole block evaluates to "pass" or "fail"
var label = if score > 50 { "pass" } else { "fail" };

# block expression: the last value in a block is its result
var x = {
    var a = 2;
    var b = 3;
    a * b  # this is the block's value (6)
};
```

## Semicolons

Statements are separated by semicolons. A trailing semicolon after the last statement in a block is optional:

```rice
var a = 1;
var b = 2  # trailing semicolon optional in blocks
```

## Keywords

Rice has exactly **10 keywords**:

```
var, const, if, else, func, for, in, continue, break, return
```

## Execution Model

Rice scripts are parsed, compiled to an AST, and interpreted by the Go-based runtime. There is no JIT or bytecode compiler — execution is pure tree-walking interpretation, which keeps the embeddable footprint small.

---

**Next:** [01 — Data Types](01-data-types.md)
