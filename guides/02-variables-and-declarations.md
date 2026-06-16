# 02 — Variables & Declarations

## Declaration

Variables are declared with [`var`](../frontend/token_type.go) (mutable) or [`const`](../frontend/token_type.go) (immutable):

```rice
var age = 19;       # mutable — can be reassigned
const name = "Bob"; # immutable — cannot be reassigned
```

- Declarations are **statements**, not expressions — they do not produce a value.
- The type is inferred from the assigned value; there is no explicit type annotation syntax.
- A `const` binding prevents reassignment, but for composite types (list, set, map) the *contents* can still be mutated:

```rice
const items = list.of(1, 2);
items.append(3);     # OK — mutating contents
# items = list.of(); # ERROR — reassigning the variable
```

## Assignment

Assignment returns the **new value** and cannot be chained:

```rice
var a = 5;
print(a = 10);  # prints "10" — assignment returns the new value
# a = 5 = b     # ERROR — chaining is disallowed
```

## Block Scope

A block (`{ ... }`) creates a new lexical scope. Variables declared inside a block **shadow** variables with the same name from outer scopes:

```rice
var a = 10;
{
    var a = 5;
    println(a);   # 5 (inner a shadows outer a)
}
println(a);       # 10 (original a is restored)
```

Shadowing works at any nesting depth:

```rice
var x = "outer";
{
    var x = "middle";
    {
        var x = "inner";
        println(x);   # "inner"
    }
    println(x);       # "middle"
}
println(x);           # "outer"
```

## Increment / Decrement

Increment and decrement are **statements**, not expressions — they do not return a value:

```rice
var i = 0;
++i;    # pre-increment: increases i, then reads i
i++;    # post-increment: reads i, then increases i
--i;    # pre-decrement
i--;    # post-decrement
```

You cannot use them inline in expressions:

```rice
# var x = i++;  # ERROR — i++ does not return a value
```

---

**Previous:** [01 — Data Types](01-data-types.md)  
**Next:** [03 — Operators](03-operators.md)
