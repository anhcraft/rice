# 04 — Collections

Rice provides four collection types: [`string`](../exec/types/values/string.go), [`list`](../exec/types/values/list.go), [`set`](../exec/types/values/set.go), and [`map`](../exec/types/values/map.go). Each has different traits.

## Collection Traits

| Type | Indexed | Mutable | Iterable |
|------|---------|---------|----------|
| `string` | ✅ | ❌ | ✅ |
| `list` | ✅ | ✅ | ✅ |
| `set` | ❌ | ✅ | ✅ |
| `map` | ✅ | ✅ | ✅ |

- **Indexed**: Supports random access via element access syntax, e.g. `list[0]`, `map[key]`, `string[0]`
- **Mutable**: Supports assigning new values to elements (for indexed collections) or adding/removing elements
- **Iterable**: Can be used with `for..in`

## String

Strings are indexed (by rune, not byte) and immutable:

```rice
"Hello"[0]       # "H" — first Unicode character
"Hello"[1]       # "e"
```

All string operations work at the rune level. Unicode is fully supported.

```rice
var s = "Café";
s[3]             # "é" (correct rune indexing, not byte indexing)
len(s)           # 4 (characters, not bytes)
```

## List

Lists are ordered, indexed, and mutable collections:

```rice
var items = list.of(1, 2, 3);
items[0]         # 1 — zero-based indexing
items[0] = 99;   # mutation via element access
items.append(4); # in-place append
items.prepend(0);# in-place prepend
```

### List Iteration

`for..in` on a list yields each **element value**:

```rice
for elem in list.of(1, 2, 3) {
    print(elem, " ");
}
```

## Set

Sets are unordered, mutable collections of unique elements. They are **not** indexed:

```rice
var s = set.of(1, 2, 3);
s.add(4);            # add an element
s.include(2);        # true
# s[0]               # ERROR — set is not indexed
```

### Set Iteration

`for..in` on a set yields each **element value**:

```rice
for elem in set.of(1, 2, 3) {
    print(elem, " ");
}
```

## Map

Maps are key-value collections that are both indexed and mutable:

```rice
var m = map.of("key", "value");
m["key"]           # "value" — access by key
m["key"] = "new";  # mutation via element access
m.put("k2", "v2"); # add entries
```

Keys and values can be of any type:

```rice
map.of(0, "value")[0]     # "value" — Int key
map.of(true, "yes")[true] # "yes"  — Bool key
```

### Selector Syntax

The dot selector (`.`) provides shorthand for string-key map access:

```rice
map.of("key", "value").key  # "value"
# equivalent to:
map.of("key", "value")["key"]
```

### Map Iteration

`for..in` on a map yields each **entry** as a `list(key, value)`:

```rice
for entry in map.of("k1", "v1", "k2", "v2") {
    const key = entry[0];
    const value = entry[1];
    println(key + ": " + value);
}
```

## String Iteration

Strings iterate character-by-character, yielding `String` values of length 1:

```rice
for ch in "hello world" {
    print(ch);     # prints each character as a String
}
```

There is no dedicated Character/Rune type — each character is a `String` of length 1.

## Element Assignment

For mutable indexed collections, you can assign new values to specific elements:

```rice
var items = list.of(1, 2, 3);
items[0] = 10;        # list mutation

var m = map.of("a", 1);
m["a"] = 2;           # map mutation
m["b"] = 3;           # map insertion via assignment
```

---

**Previous:** [03 — Operators](03-operators.md)  
**Next:** [05 — Control Flow](05-control-flow.md)
