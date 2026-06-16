# 03 — Operators

## Operator List

Rice supports these operators:

| Category | Operators |
|----------|-----------|
| Assignment | `=` |
| Logical | `&&`, `\|\|` |
| Comparison | `==`, `!=`, `>=`, `<=`, `>`, `<` |
| Arithmetic | `+`, `-`, `*`, `/`, `%` |
| Unary | `!` (not), `-` (negate) |
| Increment/Decrement | `++`, `--` |
| Spread | `...` |

## Precedence

Operators follow standard precedence rules (highest to lowest):

1. Unary: `!`, `-`
2. Multiplicative: `*`, `/`, `%`
3. Additive: `+`, `-`
4. Comparison: `==`, `!=`, `>=`, `<=`, `>`, `<`
5. Logical AND: `&&`
6. Logical OR: `||`
7. Assignment: `=`

Use parentheses to override precedence:

```rice
1 + 2 * (3 + a);
a && (b || c) && (d > 10);
!check;
-(num / 2);
```

## Short-Circuit Evaluation

Logical operators use short-circuit evaluation:

```rice
# && stops at the first falsy value
false && expensive_call();  # expensive_call() is never called

# || stops at the first truthy value
true || expensive_call();   # expensive_call() is never called
```

## Implicit Type Conversion

Binary operators perform implicit conversion following this priority order:

1. **String**: If either operand is a `String`, both are converted to `String`
2. **Bool**: If either operand is a `Bool`, both are converted to `Bool`
3. **Float**: If either operand is a `Float`, both are converted to `Float`
4. **Int**: Otherwise, if either operand is an `Int`, both are converted to `Int`

```rice
"value: " + 42      # "value: 42" (Int → String)
true + 0            # 1 (Bool → Int conversion)
5 + 3.14            # 8.14 (Int → Float)
```

## Modulo Operator

The `%` operator computes the remainder:

```rice
10 % 3      # 1
15.5 % 3    # 0.5 (works on floats too)
```

## Spread Operator (`...`)

The spread operator expands a collection into individual arguments. It can only be used **within an argument list** and can appear at any position:

```rice
myFunction(...list, 5, ":", ...set)
```

It is not restricted to the last argument position, unlike many other languages.

---

**Previous:** [02 — Variables & Declarations](02-variables-and-declarations.md)  
**Next:** [04 — Collections](04-collections.md)
