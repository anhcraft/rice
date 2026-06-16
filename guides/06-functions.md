# 06 — Functions

Rice has two kinds of functions: **native/built-in** functions and **user-defined** function literals.

## Native / Built-in Functions

- Have names (e.g., [`println`](../exec/stdlib/io/io.go), [`strings.toLower`](../exec/stdlib/string/string.go))
- **Static-typing**: parameters accept specific types, union types (e.g., `Int|Float|Bool`), or `any`
- Support **overloading** (multiple signatures for the same function name)
- Support **varargs** (indicated by `...` in documentation)
- **Namespaced**: grouped in packages — e.g., [`strings`](../exec/stdlib/string/string.go), [`math`](../exec/stdlib/math/math.go), [`list`](../exec/stdlib/list/list.go)

```rice
strings.toLower("RICE")   # "rice"
math.sqrt(16)              # 4
list.of(1, 2, 3)           # [1, 2, 3]
```

## Type-Bound Functions

Native functions whose first parameter matches a specific type can be called using dot syntax on values of that type:

```
V.f(...)   is equivalent to   f(V, ...)
```

```rice
"RICE".toLower()            # "rice"
# equivalent to:
strings.toLower("RICE")
```

```rice
list.of(3, 1, 2).sort(func(a,b){a < b})   # [1, 2, 3]
# equivalent to:
list.sort(list.of(3, 1, 2), func(a,b){a < b})
```

The selector (`.`) is the primary way to access type-bound functions.

## User-Defined Functions (Function Literals)

Function literals are defined using the [`func`](../frontend/token_type.go) keyword:

```rice
# Anonymous function, immediately invoked
func(){ print("Hello World") }();

# Assigned to a variable
const greet = func(name) {
    return "Hello, " + name + "!";
};
greet("Alice");     # "Hello, Alice!"
```

### Characteristics

- **Anonymous**: The function itself has no name; variables that reference it have names
- **Dynamic-typing**: Parameters and return values are dynamically typed
- **Varargs**: Supported via `...` in the parameter list
- **No overloading**: Each function literal has a single signature
- **Closures**: Function literals capture the innermost lexical scope at definition time and can see changes to that scope over time

### Parameters and Varargs

```rice
# Fixed parameters
const add = func(a, b) {
    return a + b;
};

# Variadic function
const sum = func(nums...) {
    var total = 0;
    for n in nums {
        total = total + n;
    }
    return total;
};
sum(1, 2, 3, 4);    # 10
```

### Return Values

A function returns the value of the last expression in its body, or `null` if there is an explicit `return;` without a value:

```rice
const calc = func(x) {
    if x < 0 { return 0; }
    x * 2           # implicit return
};
calc(5);            # 10
calc(-1);           # 0
```

## Closures

Function literals capture the scope where they are **defined**, not where they are called. They see changes to captured variables:

```rice
const makeCounter = func() {
    var count = 0;
    return func() {
        count++;
        return count;
    };
};

const counter = makeCounter();
counter();      # 1
counter();      # 2
counter();      # 3
```

## Call Syntax

Function calls use parentheses and support the spread operator:

```rice
myFunc(1, 2, 3);
myFunc(...list, 5, "...", ...set);
```

---

**Previous:** [05 — Control Flow](05-control-flow.md)  
**Next:** [07 — Functional Programming](07-functional-programming.md)
