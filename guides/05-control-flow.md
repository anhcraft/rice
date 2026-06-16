# 05 — Control Flow

## If Expression

The [`if`](../frontend/token_type.go) construct is an **expression** — it returns the value of the chosen branch:

```rice
var choiceCode =
    if choice == "Apple" {0}
    else if choice == "Banana" {1}
    else if choice == "Coconut" {2}
    else if choice == "Durian" {3}
    else {4};
```

Each branch must be a block expression. The final `else` is optional (if omitted, the `if` expression evaluates to `null` when no condition matches).

```rice
var result = if x > 0 { "positive" } else { "non-positive" };

# Without else — produces null when condition is false
var maybe = if x > 0 { "positive" };
print(maybe);   # null if x <= 0
```

## For Loop (C-Style)

C-style [`for`](../frontend/token_type.go) loops accept three optional parts: `init; cond; post`:

```rice
for (var i = 0; i < n; i++) {
    print(i);
}
```

Each part is optional:

```rice
# Infinite loop
for (;;) {
    if done { break; }
}
```

- `init`: optional simple statement (declaration, increment/decrement, or expression)
- `cond`: optional expression (loop continues while truthy)
- `post`: optional simple statement (executed after each iteration)

## For Loop (Short Form / While Loop)

Omit the parentheses and semicolons for a while-style loop:

```rice
for i < n {
    print(i);
    i++;
}
```

This is equivalent to `for (; i < n;)`.

## For-In Loop

Iterate over any collection using [`for..in`](../frontend/token_type.go):

```rice
for elem in list.of(1, 2, 3) {
    print(elem);
}
```

The loop variable (`elem` above) is declared in the loop's scope and receives each element in order.

| Collection | What each iteration yields |
|------------|---------------------------|
| `string` | Each character as a `String` (length 1) |
| `list` | Each element value |
| `set` | Each element value |
| `map` | Each entry as `list(key, value)` |

```rice
# Map iteration
for entry in map.of("a", 1, "b", 2) {
    const key = entry[0];
    const value = entry[1];
}

# String iteration
for ch in "hello" {
    print(ch);     # "h", "e", "l", "l", "o"
}
```

## Break & Continue

[`break`](../frontend/token_type.go) exits the innermost loop; [`continue`](../frontend/token_type.go) skips to the next iteration:

```rice
for (var i = 0; i < 10; i++) {
    if i == 3 { continue; }  # skip 3
    if i == 7 { break; }     # stop at 7
    print(i);                 # prints 0, 1, 2, 4, 5, 6
}
```

## Return

[`return`](../frontend/token_type.go) exits the current function, optionally with a value:

```rice
return;          # return with no value (returns null)
return "OK";     # return with value
return res;      # return a variable
```

Functions that don't have an explicit `return` implicitly return `null`.

## Nested Loop Control

`break` and `continue` only affect the **innermost** loop:

```rice
for (var i = 0; i < 3; i++) {
    for (var j = 0; j < 3; j++) {
        if j == 1 { break; }    # breaks inner j-loop only
        print(i + "," + j);
    }
}
# prints: 0,0  1,0  2,0
```

There is no labeled break/continue — to exit an outer loop, use a flag variable.

---

**Previous:** [04 — Collections](04-collections.md)  
**Next:** [06 — Functions](06-functions.md)
