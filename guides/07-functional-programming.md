# 07 — Functional Programming

Rice supports a functional programming style through higher-order functions on collections and first-class function literals. Collections have methods like [`map`](../exec/stdlib/list/list.go), [`filter`](../exec/stdlib/list/list.go), and [`sort`](../exec/stdlib/list/list.go) that accept lambda functions.

## Map

Transform each element using a mapping function. Returns a **new** collection:

```rice
# List mapping
list.of(1, 2, 3).map(func(x){ x * 2 })        # [2, 4, 6]

# Set mapping — returns a new set
set.of(1, 2, 3).map(func(x){ x * 10 })        # {10, 20, 30}

# Map mapping — lambda receives list(key, value)
map.of("a", 1, "b", 2).map(func(entry){
    return list.of(entry[0], entry[1] * 10);
})                                             # {"a": 10, "b": 20}
```

## Filter

Keep only elements for which the lambda returns `true`. Returns a **new** collection:

```rice
# List filter
list.of(1, 2, 3, 4, 5).filter(func(x){ x > 2 })   # [3, 4, 5]

# Set filter
set.of(1, 2, 3, 4, 5).filter(func(x){ x % 2 == 0 })  # {2, 4}

# Map filter — lambda receives list(key, value)
map.of("a", 1, "b", 2, "c", 3).filter(func(entry){
    return entry[1] > 1;
})                                                   # {"b": 2, "c": 3}
```

## Sort

Sort a list using a comparator function. The comparator takes two elements `(a, b)` and returns `true` if `a` should come before `b`. Sorting is **in-place**:

```rice
var items = list.of(3, 1, 4, 1, 5);
items.sort(func(a, b){ a < b });    # ascending — [1, 1, 3, 4, 5]

# Sort by a derived value
var words = list.of("apple", "banana", "cherry");
words.sort(func(a, b){ len(a) < len(b) });  # ["apple", "banana", "cherry"]
```

## Reverse

Reverse a list **in-place**:

```rice
list.of(1, 2, 3).reverse()    # [3, 2, 1]
```

## Chaining

Because many methods return the collection (or a new one), you can chain operations:

```rice
strings.join("\n", ...list.of(3, 1, 4, 1, 5)
    .map(func(x){ x * 2 })
    .filter(func(x){ x > 5 })
    .sort(func(a, b){ a < b }))
# prints: 6, 8, 10
```

## Real-World Example

```rice
strings.join("\n", ...map.new()
    .put(
        "Fruit 1", "Apple",
        "Fruit 2", "Banana",
        "Fruit 3", "Cherry",
        "Fruit 4", "Durian",
        "Fruit 5", "Elderberry"
    )
    .filter(func(entry){
        typeof(entry[1]) == "String" && entry[1].toLower().include("a")
    })
    .entries()
    .sort(func(a, b){ a[1] < b[1] })
    .map(func(entry){ entry[0] + ": " + entry[1] }))

# Output:
# Fruit 2: Banana
# Fruit 5: Elderberry
# Fruit 1: Apple
# Fruit 4: Durian
```

## Lambda Functions

Function literals used as arguments (lambdas) are regular functions. They capture the surrounding scope:

```rice
var threshold = 5;
var big = list.of(3, 6, 1, 8, 2)
    .filter(func(x){ x > threshold });   # [6, 8] — captures threshold
```

---

**Previous:** [06 — Functions](06-functions.md)  
**Next:** [08 — Standard Library](08-standard-library.md)
