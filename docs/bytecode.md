# Rice bytecode (RICE 1.0)

This is the interchange format for Rice modules. A TypeScript `DataView` and a Java `DataInput` must accept the same bytes with no Go-specific step. The Go types in `exec/vm` are a decode result, not the format.

There are no stack-map frames, exception tables, field/method descriptors, or attributes. Indexes are 0-based. Multi-byte integers are **big-endian**. `0xFFFF` means “none” for an optional `u16`. An unknown major/minor, opcode, constant tag, or flag bit fails at load.

## File layout

The preamble is frozen across every major: a VM that does not implement this module's version can still read it and report a clear error.

```
u32 magic          0x52494345 ('RICE')
u16 major
u16 minor
u16 const_count
    constants      tag u8 + payload
u16 span_count
    spans          line u32, index u32, end_line u32, end_index u32, parent u16, label u16
u16 hotspot_count
    hotspots       label u16 (string const), span u16, flags u8
u16 func_count
    functions      see below
u16 entry          function index of the script body
```

The first release is **1.0**.

### Constants

| Tag | Payload | Since |
|-----|---------|-------|
| 1   | null (no payload) | 1.0 |
| 2   | bool as `u8` (0 or 1) | 1.0 |
| 3   | int as `i64` | 1.0 |
| 4   | float as IEEE-754 binary64 | 1.0 |
| 5   | string: `u32` byte length + standard UTF-8 (not Java modified UTF-8) | 1.0 |

### Spans

`line` and `end_line` are 1-based. `index` and `end_index` are 0-based rune offsets (`end` exclusive), matching `ast.Pos`. `parent` is a span index or `0xFFFF`. `label` is a string-constant index or `0xFFFF`.

### Hotspots

`flags` bit 0: the region opens a lexical scope (`OpEnter` increments scope depth and checks `LexicalScopeLimit`). Since 1.0. Unused bits must be 0.

### Function

```
u8  flags          bit 0 = variadic (since 1.0)
u16 arity
u16 max_stack
u16 slot_count
    slot_names     slot_count × u16 string const (params first; `0xFFFF` for compiler temps)
u16 upvalue_count
    upvalues       upvalue_count × (is_local u8, index u16)
u32 code_len
    code           code_len × 5 bytes: opcode u8, argument i32
    spans          code_len × u16 span index
```

Instruction `i` starts at byte offset `i * 5` within `code`. After `ip++`, a jump adds (or subtracts) `argument` instruction slots.

Limits: 65535 constants, spans, hotspots, functions, and slots; 255 upvalues per function; arguments fit in `i32`. Exceeding a limit is a compile error.

A new opcode takes a free number in the groups below and bumps the **minor**.

## Versioning

The version is `major.minor`. A VM built for `M.n` loads a module `X.y` when `MinMajor <= X <= M` and, if `X == M`, `y <= n`. Anything newer fails at load (`module needs RICE 1.3, this VM supports up to 1.2`). This is the same guarantee as a Java 17 JVM running class files from Java 8.

`MinMajor` starts at `1`. It is raised only by an explicit, documented decision to drop an old major.

The encoder stamps the **lowest** version that covers every opcode, constant tag, and flag the module actually uses. A blob from a 1.4 compiler that uses only 1.0 features is marked 1.0 and runs on a 1.0 VM.

`Decode` rejects a module that uses an opcode, tag, or flag newer than its declared version.

Stdlib is outside this versioning. Host names resolve at run time. A missing native fails when its instruction runs.

### Minor bump (additive)

Every byte that was valid in `M.(n-1)` keeps exactly the same meaning.

| Allowed | Example |
|---------|---------|
| New opcode in a free number | `OpFoo = 6` |
| New constant tag | tag 6 |
| New flag bit that older modules leave as 0 | function flags bit 1 |

### Major bump (breaking)

Any change that alters existing bytes.

| Breaking | Example |
|----------|---------|
| Header or section layout | extra table before functions |
| Instruction width | 6-byte instructions |
| Opcode stack effect or semantics | `OpAdd` starts concatenating maps |
| Remove or renumber an opcode, tag, or flag | `OpPop` becomes 99 |

**Opcode meaning is frozen within a major.** To change behavior (for example a different `++` rule), add a new opcode and bump the minor. Never redefine an existing one.

**Upgrade on decode.** A VM keeps one decoder per supported major. Each older-major decoder rewrites the module into the current in-memory `Module`. `VM.Run` only knows the latest instruction set. Ports do the same.

## Opcodes

Unused argument is `0`.

### Stack (1–15)

| # | Name | Arg | Stack | Since |
|---|------|-----|-------|-------|
| 1 | `OpPush` | const | → c | 1.0 |
| 2 | `OpPop` | — | x → | 1.0 |
| 3 | `OpDup` | — | x → x x | 1.0 |
| 4 | `OpDup2` | — | a b → a b a b | 1.0 |
| 5 | `OpRot3` | — | a b c → b c a | 1.0 |

### Slots (16–31)

| # | Name | Arg | Since |
|---|------|-----|-------|
| 16 | `OpLoadLocal` | slot | 1.0 |
| 17 | `OpStoreLocal` | slot (pops, does not push) | 1.0 |
| 18 | `OpDefineLocal` | slot (pops) | 1.0 |
| 19 | `OpDefineConstLocal` | slot (pops) | 1.0 |
| 20 | `OpLoadUpvalue` | upvalue | 1.0 |
| 21 | `OpStoreUpvalue` | upvalue (pops) | 1.0 |
| 22 | `OpDefineUpvalue` | upvalue (pops) | 1.0 |
| 23 | `OpDefineConstUpvalue` | upvalue (pops) | 1.0 |
| 24 | `OpCloseUpvalue` | slot (close the open upvalue for that local) | 1.0 |

`OpStore*` on a const is a runtime error. `OpLoad*` of an undefined slot is `unresolved reference`. `OpDefineLocal` / `OpDefineConstLocal` overwrite the slot so a loop body can re-enter; redeclaring a name in the same block is a compile error. `OpDefineGlobal` / `OpDefineConstGlobal` fail if the name is already defined. Every `OpDefine*` fails if the name collides with a host namespace entry.

### Globals (32–47)

| # | Name | Arg | Since |
|---|------|-----|-------|
| 32 | `OpLoadGlobal` | string const | 1.0 |
| 33 | `OpStoreGlobal` | string const (pops) | 1.0 |
| 34 | `OpDefineGlobal` | string const (pops) | 1.0 |
| 35 | `OpDefineConstGlobal` | string const (pops) | 1.0 |

`OpLoadGlobal` tries the host namespace, then the root environment.

### Jumps (48–63)

Forward jumps use `OpJump`; loop backs use `OpJumpBackward`. Conditional jumps **do not pop**.

| # | Name | When it jumps | Since |
|---|------|----------------|-------|
| 48 | `OpJump` | always (`ip += arg`) | 1.0 |
| 49 | `OpJumpBackward` | always (`ip -= arg`); checks cancellation | 1.0 |
| 50 | `OpJumpIfFalse` | top is `Bool` false; error if not `Bool` | 1.0 |
| 51 | `OpJumpIfFalsey` | `AsBool(top)` is false | 1.0 |
| 52 | `OpJumpIfBoolFalse` | top is `Bool` and false (no jump if not `Bool`) | 1.0 |
| 53 | `OpJumpIfBoolTrue` | top is `Bool` and true | 1.0 |

### Operators (64–95)

Binary ops pop right, then left, and push the result.

| # | Name | Since |
|---|------|-------|
| 64 | `OpNegate` | 1.0 |
| 65 | `OpNot` | 1.0 |
| 66 | `OpAdd` | 1.0 |
| 67 | `OpSubtract` | 1.0 |
| 68 | `OpMultiply` | 1.0 |
| 69 | `OpDivide` | 1.0 |
| 70 | `OpModulo` | 1.0 |
| 71 | `OpEqual` | 1.0 |
| 72 | `OpLess` | 1.0 |
| 73 | `OpMore` | 1.0 |
| 74 | `OpLessOrEqual` | 1.0 |
| 75 | `OpMoreOrEqual` | 1.0 |
| 76 | `OpAnd` | 1.0 |
| 77 | `OpOr` | 1.0 |
| 78 | `OpIncrement` (arg is `+1` or `-1`; Int/Float only) | 1.0 |

`!=` is `OpEqual` then `OpNot`.

Implicit coercion order: String, Bool, Float, Int. Ordering on strings runs only when **both operands were strings before coercion**. Non-primitives (including `null`) compare by identity for `==` / `!=` and error otherwise.

`&&` / `||` compile as `left; JumpIfBoolFalse/True end; right; OpAnd/OpOr; end:` so a left `Bool` false/true skips the right side; otherwise both sides run and `OpAnd`/`OpOr` apply the binary rules.

Integer `/` and `%` by zero are runtime errors. Integer arithmetic wraps in 64-bit two’s complement.

### Data (96–111)

| # | Name | Arg | Stack | Since |
|---|------|-----|-------|-------|
| 96 | `OpArray` | n | v1…vn → list | 1.0 |
| 97 | `OpMap` | n | (k v)×n → map | 1.0 |
| 98 | `OpGetIndex` | — | obj idx → val | 1.0 |
| 99 | `OpSetIndex` | — | obj idx val → val | 1.0 |
| 100 | `OpLoadField` | string const | obj → val | 1.0 |
| 101 | `OpStoreField` | string const | obj val → val | 1.0 |

Field load: host type-bound table for `value.Type()` + name, else indexed access. The name is an identifier. Missing map keys yield `null`.

### Calls (112–127)

| # | Name | Arg | Stack | Since |
|---|------|-----|-------|-------|
| 112 | `OpClosure` | function index | → closure | 1.0 |
| 113 | `OpCall` | argc | callee a1…an → result | 1.0 |
| 114 | `OpArgs` | — | start spread buffer | 1.0 |
| 115 | `OpArg` | — | x → (buffer) | 1.0 |
| 116 | `OpSpread` | — | collection → (buffer) | 1.0 |
| 117 | `OpCallArgs` | — | callee → result | 1.0 |
| 118 | `OpReturn` | — | x → (return x) | 1.0 |

Spread accepts any collection. `OpCall` / `OpCallArgs` require a callable.

### Iteration (128–143)

| # | Name | Since |
|---|------|-------|
| 128 | `OpBegin` (pop collection, push iterator; length fixed at start) | 1.0 |
| 129 | `OpJumpIfEnd` | 1.0 |
| 130 | `OpPointer` (push current element) | 1.0 |
| 131 | `OpIncrementIndex` | 1.0 |
| 132 | `OpEnd` (pop iterator) | 1.0 |

Portable yields: string code points, list elements, set elements, map entries as `[key, value]` lists.

### Regions (144–159)

| # | Name | Arg | Since |
|---|------|-----|-------|
| 144 | `OpEnter` | hotspot | 1.0 |
| 145 | `OpLeave` | — | 1.0 |

`OpEnter` checks cancellation; if the hotspot opens a scope, increments scope depth and checks the limit; then starts a profiler region (no-op when profiling is off). `OpLeave` ends the profiler region and decrements scope depth when applicable. User-function and native calls also add one scope level. Each frame restores scope and profiler depth on return or fault.

## Port rules

These are where Go, Java, and TypeScript disagree. Ports must match these vectors.

- **Int:** signed 64-bit two’s complement, wrapping. TypeScript uses `bigint` with `BigInt.asIntN(64)`; Java uses `long`. Division truncates toward zero. Remainder takes the sign of the dividend.
- **Float:** IEEE-754 binary64. Remainder follows Go `math.Mod` (sign of the dividend). Truthiness is `|x| > 1e-9`.
- **Float to string** (concatenation, `string()`): Go `strconv.FormatFloat(x, 'g', -1, 64)`. `1000000` prints `1e+06`; `1.0` prints `1`.
- **String to Int / Float / Bool:** Go `ParseInt` base 10, `ParseFloat`, and `ParseBool` (exactly `1`, `t`, `T`, `TRUE`, `true`, `True`, `0`, `f`, `F`, `FALSE`, `false`, `False`).
- **Strings:** UTF-8 on the wire. Indexing, `len`, `substr`, and `for-in` count Unicode code points. Ordering compares code points (same as Go byte order on well-formed UTF-8). Do not use Java `String.compareTo` or JS `<` (UTF-16 units).
- **Equality and keys:** primitives by type and value after coercion; list, map, set, and function by identity. Map/set keys: primitives by type and value, composites by identity. Iteration order of map and set is unspecified.
- **Stdlib is not in the blob.** Each host provides the same global and type-bound names. A missing name is a runtime error when the instruction runs.

## Conformance corpus

`exec/conformance` holds compiled `RICE` modules and an expected-result file. `go test ./exec/conformance -update` rebuilds `testdata/*.ricebc` and `expected.txt` from `exec/testdata/*.rice`. Port-rule vectors (float format, parse, string order) live in `exec/conformance/vectors/` and are also checked by the Go tests in that package.

`exec/conformance/compat/<major.minor>/` is a frozen snapshot. `-update` never rewrites it. A newer VM must still load and run every folder at or below its version.
