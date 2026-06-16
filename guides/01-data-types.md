# 01 — Data Types

Rice is dynamically typed. Every value belongs to one of the following types.

## Primitive Types

| Type | Examples | Notes |
|------|----------|-------|
| **Int** | `123` | 64-bit signed integer |
| **Float** | `-123.456`, `1e-9` | 64-bit floating-point (supports scientific notation) |
| **Bool** | `true`, `false` | Boolean literals |
| **String** | `"Xin chao!"` | Unicode-aware; all operations work at rune level, not byte level |
| **Null** | `null` | Represents absence of value |

```rice
123          # Int
-123.456     # Float
1e-9         # Float (scientific notation)
true         # Bool
"Xin chao!"  # String
null         # Null
```

## Composite Types

| Type | Description |
|------|-------------|
| **List** | Dynamic-typed, ordered, indexed collection — created via [`list.new()`](../exec/stdlib/list/list.go) or [`list.of()`](../exec/stdlib/list/list.go) |
| **Set** | Dynamic-typed, unordered collection of unique elements — created via [`set.new()`](../exec/stdlib/set/set.go) or [`set.of()`](../exec/stdlib/set/set.go) |
| **Map** | Dynamic-typed, key-value collection — created via [`map.new()`](../exec/stdlib/map/map.go) or [`map.of()`](../exec/stdlib/map/map.go) |
| **Func** | Function literal — created with `func() { ... }` |

```rice
list.new()   # empty list
set.new()    # empty set
map.new()    # empty map
func() {     # function literal
    return
}
```

## Important Notes

- **No array type**: There is no fixed-length array or `[...]` array literal syntax. Use [`list`](../exec/stdlib/list/list.go) instead.
- **No dedicated Character/Rune type**: Individual characters from a string are themselves `String` values of length 1.
- **Unicode**: Strings are fully Unicode-aware. All string operations (indexing, length, slicing) work at the Unicode codepoint (rune) level — not at the byte level like Go's native strings.

## Checking Types

Use [`typeof()`](../exec/stdlib/type/type.go) to get the type name of any value:

```rice
typeof(123)          # "Int"
typeof(3.14)         # "Float"
typeof(true)         # "Bool"
typeof("hello")      # "String"
typeof(list.new())   # "List"
typeof(set.new())    # "Set"
typeof(map.new())    # "Map"
typeof(func(){})     # "Func"

# it is intended that type-of null is "null" (lowercase) while others are Title case
# because null is not a real data type in Rice
typeof(null)         # "null"
```

## Type Conversion

Built-in functions convert between types:

```rice
string(123)       # "123"
int("456")        # 456
float("3.14")     # 3.14
bool(1)           # true
bool(0)           # false
```

## Unicode Codepoints

String literals support Unicode escape sequences:

```rice
"\u20AC"          # € (4-digit hex)
"\U0001F60E"      # 😎 (8-digit hex)
```

---

**Previous:** [00 — Introduction](00-introduction.md)  
**Next:** [02 — Variables & Declarations](02-variables-and-declarations.md)
