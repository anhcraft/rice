# 08 — Standard Library

This is a comprehensive reference of all built-in functions.

## Conventions

- `any` — accepts any type
- `Int|Float|Bool` — union type, accepts any of those types
- `...any` — variadic, accepts zero or more arguments of any type
- Functions marked with `...` in the parameter list support varargs
- All native functions support overloading

---

## Type & Conversion

Located in [`exec/stdlib/type/`](../exec/stdlib/type/type.go).

| Function | Description |
|----------|-------------|
| `typeof(value any)` | Returns the type name: `"Int"`, `"Float"`, `"Bool"`, `"String"`, `"List"`, `"Set"`, `"Map"`, `"Func"`, `"null"` |
| `float(value any)` | Converts value to Float |
| `bool(value any)` | Converts value to Bool |
| `int(value any)` | Converts value to Int |
| `string(value any)` | Converts value to String |
| `isNumberLike(value any)` | Returns `true` if the type is Int, Float, or Bool |
| `isNumber(value any)` | Returns `true` if the type is Int or Float |
| `len(value any)` | Returns the length of a collection (string, list, set, map) |

```rice
typeof(42)              # "Int"
int("123")              # 123
float("3.14")           # 3.14
string(true)            # "true"
bool(1)                 # true
len("hello")            # 5
len(list.of(1, 2, 3))   # 3
```

---

## I/O

Located in [`exec/stdlib/io/`](../exec/stdlib/io/io.go).

| Function | Description |
|----------|-------------|
| `print(args ...any)` | Prints arguments with no separator |
| `println(args ...any)` | Prints arguments with a trailing newline |
| `printf(format String, args ...any)` | Prints formatted string (no trailing newline) |
| `printlnf(format String, args ...any)` | Prints formatted string with trailing newline |

---

## Error Handling

Located in [`exec/stdlib/error/`](../exec/stdlib/error/error.go).

| Function | Description |
|----------|-------------|
| `assert(cond Bool, msg String)` | Throws an exception with `msg` if `cond` is `false` |
| `throw(msg String)` | Throws an exception with the given message |

```rice
assert(x > 0, "x must be positive");
# throw("Something went wrong");
```

---

## Math

Located in [`exec/stdlib/math/`](../exec/stdlib/math/math.go).

| Function | Description |
|----------|-------------|
| `math.sqrt(num Int\|Float\|Bool)` | Square root |
| `math.floor(num Int\|Float\|Bool)` | Floor |
| `math.ceil(num Int\|Float\|Bool)` | Ceiling |
| `math.pow(base Int\|Float\|Bool, exp Int\|Float\|Bool)` | Power (base^exp) |
| `math.max(args ...any)` | Maximum of arguments |
| `math.min(args ...any)` | Minimum of arguments |
| `math.abs(num Int\|Float\|Bool)` | Absolute value |

```rice
math.sqrt(16)           # 4
math.floor(3.7)         # 3
math.ceil(3.2)          # 4
math.pow(2, 3)          # 8
math.max(3, 7, 1, 9)    # 9
math.abs(-5)            # 5
```

---

## Strings

Located in [`exec/stdlib/string/`](../exec/stdlib/string/string.go).

| Function | Description |
|----------|-------------|
| `strings.join(separator String, values ...any)` | Joins values with separator, converting each to String |
| `strings.toUpper(str String)` | Uppercase |
| `strings.toLower(str String)` | Lowercase |
| `strings.index(str String, substr String)` | First index of substr, or -1 |
| `strings.lastIndex(str String, substr String)` | Last index of substr, or -1 |
| `strings.substr(str String, offset Int, exclusiveEnd Int)` | Substring from offset to exclusiveEnd |
| `strings.substr(str String, offset Int)` | Substring from offset to end |
| `strings.format(fmt String, args ...any)` | Format a string |
| `strings.trim(str String)` | Trim whitespace |
| `strings.include(str String, substr String)` | Check if substr is contained |
| `strings.split(str String, separator String)` | Split into list by separator |

```rice
strings.join(", ", "a", "b", "c")  # "a, b, c"
"hello".toUpper()                   # "HELLO"
"HELLO".toLower()                   # "hello"
"hello".include("ell")              # true
"a,b,c".split(",")                  # ["a", "b", "c"]
"  hi  ".trim()                     # "hi"
```

---

## List

Located in [`exec/stdlib/list/`](../exec/stdlib/list/list.go).

| Function | Description |
|----------|-------------|
| `list.new()` | Creates an empty list |
| `list.of(items ...any)` | Creates a list from the given items |
| `list.prepend(list List, items ...any)` | Prepends items in-place, returns the list |
| `list.append(list List, items ...any)` | Appends items in-place, returns the list |
| `list.include(list List, item any)` | Checks if item exists |
| `list.index(list List, item any)` | First index of item, or -1 |
| `list.lastIndex(list List, item any)` | Last index of item, or -1 |
| `list.sort(list List, lambda Func)` | Sorts in-place using comparator `(a, b) → true if a before b` |
| `list.map(list List, lambda Func)` | Returns a new list from mapping function |
| `list.reverse(list List)` | Reverses in-place, returns the list |
| `list.filter(list List, lambda Func)` | Returns a new list from filter function |
| `list.removeAt(list List, index Int)` | Removes item at index in-place |
| `list.removeAll(list List, item any)` | Removes all occurrences of item, returns count removed |
| `list.slice(list List, offset Int, exclusiveEnd Int)` | Returns a new sliced list |
| `list.slice(list List, offset Int)` | Returns a new list sliced from offset to end |

```rice
list.of(1, 2, 3).append(4)                # [1, 2, 3, 4]
list.of(2, 3).prepend(1)                   # [1, 2, 3]
list.of(1, 3, 2).sort(func(a,b){a < b})    # [1, 2, 3]
list.of(1, 2, 3, 4, 5).slice(1, 4)         # [2, 3, 4]
```

---

## Set

Located in [`exec/stdlib/set/`](../exec/stdlib/set/set.go).

| Function | Description |
|----------|-------------|
| `set.new()` | Creates an empty set |
| `set.of(items ...any)` | Creates a set from the given items |
| `set.add(set Set, items ...any)` | Adds items in-place, returns the set |
| `set.include(set Set, item any)` | Checks if item exists |
| `set.map(set Set, lambda Func)` | Returns a new set from mapping function |
| `set.filter(set Set, lambda Func)` | Returns a new set from filter function |
| `set.remove(set Set, item any)` | Removes item in-place, returns the set |

```rice
set.of(1, 2, 3).add(4)              # {1, 2, 3, 4}
set.of(1, 2, 3).include(2)          # true
set.of(1, 2, 3).remove(2)           # {1, 3}
```

---

## Map

Located in [`exec/stdlib/map/`](../exec/stdlib/map/map.go).

| Function | Description |
|----------|-------------|
| `map.new()` | Creates an empty map |
| `map.of(entries ...any)` | Creates a map from alternating K-V pairs |
| `map.put(map Map, entries ...any)` | Puts K-V entries in order, returns the map |
| `map.remove(map Map, keys ...any)` | Removes entries by keys, returns the map |
| `map.keys(map Map)` | Returns a **set** of keys |
| `map.values(map Map)` | Returns a **list** of values |
| `map.entries(map Map)` | Returns a **list** of `list(key, value)` |
| `map.includeKey(map Map, key any)` | Checks if a key exists |
| `map.map(map Map, lambda Func)` | Returns a new map from mapping `list(key, value)` |
| `map.filter(map Map, lambda Func)` | Returns a new map from filter `list(key, value)` |

```rice
map.of("a", 1, "b", 2).keys()       # {"a", "b"}
map.of("a", 1, "b", 2).values()     # [1, 2]
map.of("a", 1).includeKey("a")      # true
map.of("a", 1).put("b", 2)          # {"a": 1, "b": 2}
```

---

## Datetime

Located in [`exec/stdlib/datetime/`](../exec/stdlib/datetime/datetime.go).

| Function | Description |
|----------|-------------|
| `datetime.now()` | Returns the current Unix timestamp in milliseconds |
| `datetime.parse(s String)` | Parses an ISO 8601 / RFC 3339 date string and returns a Unix timestamp in ms |
| `datetime.format(ts Int, fmt String)` | Formats a Unix millisecond timestamp in UTC. Supported formats: `"rfc3339"`, `"date"`, `"time"`, `"datetime"` |

```rice
datetime.now()                       # 1700000000000 (example)
datetime.parse("2024-01-15T10:30:00Z")  # 1705314600000
datetime.parse("2024-01-15")            # 1705276800000 (midnight UTC)
datetime.format(1705314600000, "rfc3339")  # "2024-01-15T10:30:00Z"
datetime.format(1705314600000, "date")     # "2024-01-15"
datetime.format(1705314600000, "datetime") # "2024-01-15 10:30:00"
```

---

## Duration

Located in [`exec/stdlib/duration/`](../exec/stdlib/duration/duration.go).

| Function | Description |
|----------|-------------|
| `duration.parse(s String)` | Parses a duration string and returns milliseconds. Delegates to Go's `time.ParseDuration` — supports `ns`, `us`/`µs`, `ms`, `s`, `m`, `h` |
| `duration.days(n Int\|Float\|Bool)` | Converts n days to milliseconds |
| `duration.hours(n Int\|Float\|Bool)` | Converts n hours to milliseconds |
| `duration.minutes(n Int\|Float\|Bool)` | Converts n minutes to milliseconds |
| `duration.seconds(n Int\|Float\|Bool)` | Converts n seconds to milliseconds |
| `duration.millis(n Int\|Float\|Bool)` | Converts n milliseconds to milliseconds (identity) |

```rice
duration.parse("2h30m")    # 9000000
duration.parse("-1h")      # -3600000
duration.days(1)           # 86400000
duration.hours(3.5)        # 12600000
duration.minutes(10)       # 600000
```

---

## JSON

Located in [`exec/stdlib/json/`](../exec/stdlib/json/json.go).

| Function | Description |
|----------|-------------|
| `json.encode(val any)` | Converts a Rice value to minified JSON |
| `json.encode(val any, indent String)` | Converts a Rice value to prettified JSON |
| `json.decode(val String)` | Parses JSON into corresponding Rice values |

```rice
json.encode(map.of("key", "value"))         # {"key":"value"}
json.encode(map.of("key", "value"), "  ")   # prettified
json.decode('{"key":"value"}')              # {"key": "value"}
```

---

## Type-Bound Functions Quick Reference

These are the same functions called via dot syntax:

| Type | Available Methods |
|------|-------------------|
| **String** | `toUpper()`, `toLower()`, `index()`, `substr()`, `format()`, `trim()`, `include()`, `lastIndex()`, `split()`, `join()` |
| **Int** | `max()`, `min()`, `abs()`, `sqrt()`, `floor()`, `ceil()`, `pow()` |
| **Float** | `max()`, `min()`, `abs()`, `sqrt()`, `floor()`, `ceil()`, `pow()` |
| **List** | `append()`, `include()`, `index()`, `lastIndex()`, `sort()`, `map()`, `reverse()`, `filter()`, `removeAt()`, `removeAll()`, `slice()`, `prepend()` |
| **Set** | `map()`, `filter()`, `remove()`, `add()`, `include()` |
| **Map** | `map()`, `filter()`, `put()`, `remove()`, `keys()`, `entries()`, `includeKey()`, `values()` |

---

**Previous:** [07 — Functional Programming](07-functional-programming.md)  
**Next:** [09 — Embedding Rice in Go](09-integration.md)
