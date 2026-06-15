package json

import (
	"math"
	"strings"
	"testing"

	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
)

// ---- Helpers ----

func mustEncode(t *testing.T, val types.Value) values.String {
	t.Helper()
	res, err := Encode(val)
	if err != nil {
		t.Fatalf("Encode(%v) returned error: %v", val, err)
	}
	return res.(values.String)
}

func mustEncodePretty(t *testing.T, val types.Value, indent string) values.String {
	t.Helper()
	res, err := EncodePretty(val, values.String(indent))
	if err != nil {
		t.Fatalf("EncodePretty(%v, %q) returned error: %v", val, indent, err)
	}
	return res.(values.String)
}

func mustDecode(t *testing.T, raw string) types.Value {
	t.Helper()
	res, err := Decode(values.String(raw))
	if err != nil {
		t.Fatalf("Decode(%q) returned error: %v", raw, err)
	}
	return res
}

func assertJSON(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("JSON mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

// ---- Primitives ----

func TestEncodeNil(t *testing.T) {
	res := mustEncode(t, nil)
	assertJSON(t, string(res), "null")
}

func TestEncodeInt(t *testing.T) {
	res := mustEncode(t, values.Int(42))
	assertJSON(t, string(res), "42")

	res = mustEncode(t, values.Int(-7))
	assertJSON(t, string(res), "-7")

	res = mustEncode(t, values.Int(0))
	assertJSON(t, string(res), "0")
}

func TestEncodeFloat(t *testing.T) {
	res := mustEncode(t, values.Float(3.14))
	assertJSON(t, string(res), "3.14")

	res = mustEncode(t, values.Float(-0.5))
	assertJSON(t, string(res), "-0.5")
}

func TestEncodeFloatSpecial(t *testing.T) {
	res := mustEncode(t, values.Float(math.NaN()))
	assertJSON(t, string(res), "null")

	res = mustEncode(t, values.Float(math.Inf(1)))
	assertJSON(t, string(res), "null")

	res = mustEncode(t, values.Float(math.Inf(-1)))
	assertJSON(t, string(res), "null")
}

func TestEncodeBool(t *testing.T) {
	res := mustEncode(t, values.Bool(true))
	assertJSON(t, string(res), "true")

	res = mustEncode(t, values.Bool(false))
	assertJSON(t, string(res), "false")
}

func TestEncodeString(t *testing.T) {
	res := mustEncode(t, values.String("hello"))
	assertJSON(t, string(res), `"hello"`)

	res = mustEncode(t, values.String("line\nbreak"))
	assertJSON(t, string(res), `"line\nbreak"`)

	res = mustEncode(t, values.String("tab\there"))
	assertJSON(t, string(res), `"tab\there"`)

	res = mustEncode(t, values.String(`quote"inside`))
	assertJSON(t, string(res), `"quote\"inside"`)
}

func TestEncodeStringUnicode(t *testing.T) {
	res := mustEncode(t, values.String("こんにちは"))
	assertJSON(t, string(res), `"こんにちは"`)

	res = mustEncode(t, values.String("🚀"))
	assertJSON(t, string(res), `"🚀"`)

	res = mustEncode(t, values.String("\\backslash"))
	assertJSON(t, string(res), `"\\backslash"`)
}

// ---- Collections ----

func TestEncodeEmptyList(t *testing.T) {
	res := mustEncode(t, values.NewList())
	assertJSON(t, string(res), "[]")
}

func TestEncodeList(t *testing.T) {
	l := values.ListOf([]types.Value{
		values.Int(1), values.String("two"), values.Bool(true),
	})
	res := mustEncode(t, l)
	assertJSON(t, string(res), `[1,"two",true]`)
}

func TestEncodeNestedList(t *testing.T) {
	inner := values.ListOf([]types.Value{values.Int(1), values.Int(2)})
	outer := values.ListOf([]types.Value{inner, values.Int(3)})
	res := mustEncode(t, outer)
	assertJSON(t, string(res), `[[1,2],3]`)
}

func TestEncodeEmptySet(t *testing.T) {
	res := mustEncode(t, values.NewSet())
	assertJSON(t, string(res), "[]")
}

func TestEncodeSet(t *testing.T) {
	s := values.SetOf([]values.String{"a", "b"})
	res := mustEncode(t, s)
	// Set iteration order is not guaranteed, check structure
	raw := string(res)
	if !strings.HasPrefix(raw, "[") || !strings.HasSuffix(raw, "]") {
		t.Errorf("Set JSON should be an array, got: %s", raw)
	}
}

func TestEncodeEmptyMap(t *testing.T) {
	res := mustEncode(t, values.NewMap())
	assertJSON(t, string(res), "{}")
}

func TestEncodeMap(t *testing.T) {
	m := values.NewMap()
	m.Put(values.String("name"), values.String("Alice"))
	m.Put(values.String("age"), values.Int(30))
	res := mustEncode(t, m)
	raw := string(res)
	if !strings.HasPrefix(raw, "{") || !strings.HasSuffix(raw, "}") {
		t.Errorf("Map JSON should be an object, got: %s", raw)
	}
	if !strings.Contains(raw, `"name"`) || !strings.Contains(raw, `"Alice"`) {
		t.Errorf("Map JSON should contain name/Alice, got: %s", raw)
	}
	if !strings.Contains(raw, `"age"`) || !strings.Contains(raw, `30`) {
		t.Errorf("Map JSON should contain age/30, got: %s", raw)
	}
}

func TestEncodeMapNonStringKeys(t *testing.T) {
	m := values.NewMap()
	m.Put(values.Int(1), values.String("one"))
	m.Put(values.Float(2.5), values.String("two-point-five"))
	res := mustEncode(t, m)
	raw := string(res)
	// Keys are stringified via fmt.Sprint
	if !strings.HasPrefix(raw, "{") || !strings.HasSuffix(raw, "}") {
		t.Errorf("Map JSON should be an object, got: %s", raw)
	}
}

// ---- Pretty Print ----

func TestEncodePretty(t *testing.T) {
	m := values.NewMap()
	m.Put(values.String("a"), values.Int(1))
	m.Put(values.String("b"), values.Int(2))

	res := mustEncodePretty(t, m, "  ")
	raw := string(res)
	// Should contain newlines and indentation
	if !strings.Contains(raw, "\n") {
		t.Errorf("Pretty JSON should contain newlines, got: %s", raw)
	}
	if !strings.Contains(raw, "  ") {
		t.Errorf("Pretty JSON should contain indentation, got: %s", raw)
	}
}

func TestEncodePrettyCustomIndent(t *testing.T) {
	l := values.ListOf([]types.Value{values.Int(1), values.Int(2)})
	res := mustEncodePretty(t, l, "\t")
	raw := string(res)
	if !strings.Contains(raw, "\n") {
		t.Errorf("Pretty JSON should contain newlines, got: %s", raw)
	}
	if !strings.Contains(raw, "\t") {
		t.Errorf("Pretty JSON should contain tab indentation, got: %s", raw)
	}
}

func TestEncodePrettyMinified(t *testing.T) {
	// When no indent is passed via Encode, it should be minified
	m := values.NewMap()
	m.Put(values.String("key"), values.String("value"))
	res := mustEncode(t, m)
	if strings.Contains(string(res), "\n") {
		t.Errorf("Minified JSON should not contain newlines, got: %s", string(res))
	}
}

// ---- Encode Error Cases ----

func TestEncodeFuncErrors(t *testing.T) {
	fn := values.NewFunc(
		[]values.Identifier{"x"},
		false,
		nil,
		nil,
	)
	_, err := Encode(fn)
	if err == nil {
		t.Error("Encode of a function value should return an error")
	}
}

func TestEncodeNativeFuncSetErrors(t *testing.T) {
	nf := values.NewNativeFunctionSet(nil, nil)
	_, err := Encode(nf)
	if err == nil {
		t.Error("Encode of a NativeFunctionSet should return an error")
	}
}

// ---- Decode ----

func TestDecodeNull(t *testing.T) {
	res := mustDecode(t, "null")
	if res != nil {
		t.Errorf("Decode(null) should return nil, got %v", res)
	}
}

func TestDecodeBool(t *testing.T) {
	res := mustDecode(t, "true")
	if res != values.Bool(true) {
		t.Errorf("Decode(true) = %v, want Bool(true)", res)
	}

	res = mustDecode(t, "false")
	if res != values.Bool(false) {
		t.Errorf("Decode(false) = %v, want Bool(false)", res)
	}
}

func TestDecodeInt(t *testing.T) {
	res := mustDecode(t, "42")
	if res != values.Int(42) {
		t.Errorf("Decode(42) = %v (type %T), want Int(42)", res, res)
	}

	res = mustDecode(t, "-100")
	if res != values.Int(-100) {
		t.Errorf("Decode(-100) = %v (type %T), want Int(-100)", res, res)
	}

	res = mustDecode(t, "0")
	if res != values.Int(0) {
		t.Errorf("Decode(0) = %v (type %T), want Int(0)", res, res)
	}
}

func TestDecodeFloat(t *testing.T) {
	res := mustDecode(t, "3.14")
	if res != values.Float(3.14) {
		t.Errorf("Decode(3.14) = %v (type %T), want Float(3.14)", res, res)
	}

	res = mustDecode(t, "-0.5")
	if res != values.Float(-0.5) {
		t.Errorf("Decode(-0.5) = %v (type %T), want Float(-0.5)", res, res)
	}

	res = mustDecode(t, "1e10")
	// 1e10 = 10000000000, which fits in int64 and is a whole number -> Int
	if res != values.Int(10000000000) {
		t.Errorf("Decode(1e10) = %v (type %T), want Int(10000000000)", res, res)
	}

	res = mustDecode(t, "1.5e3")
	// 1.5e3 = 1500, whole number, fits int64 -> Int
	if res != values.Int(1500) {
		t.Errorf("Decode(1.5e3) = %v (type %T), want Int(1500)", res, res)
	}
}

func TestDecodeString(t *testing.T) {
	res := mustDecode(t, `"hello"`)
	if res != values.String("hello") {
		t.Errorf("Decode(\"hello\") = %v, want String(hello)", res)
	}

	res = mustDecode(t, `"line\nbreak"`)
	if res != values.String("line\nbreak") {
		t.Errorf("Decode(\"line\\nbreak\") = %v, want String(line\\nbreak)", res)
	}
}

func TestDecodeEmptyArray(t *testing.T) {
	res := mustDecode(t, "[]")
	l, ok := res.(*values.List)
	if !ok {
		t.Fatalf("Decode([]) should return *values.List, got %T", res)
	}
	if l.Size() != 0 {
		t.Errorf("Expected empty list, got size %d", l.Size())
	}
}

func TestDecodeArray(t *testing.T) {
	res := mustDecode(t, `[1,"two",true,3.14]`)
	l, ok := res.(*values.List)
	if !ok {
		t.Fatalf("Expected *values.List, got %T", res)
	}
	if l.Size() != 4 {
		t.Fatalf("Expected list size 4, got %d", l.Size())
	}
	if l.At(0) != values.Int(1) {
		t.Errorf("l[0] = %v, want Int(1)", l.At(0))
	}
	if l.At(1) != values.String("two") {
		t.Errorf("l[1] = %v, want String(two)", l.At(1))
	}
	if l.At(2) != values.Bool(true) {
		t.Errorf("l[2] = %v, want Bool(true)", l.At(2))
	}
	if l.At(3) != values.Float(3.14) {
		t.Errorf("l[3] = %v, want Float(3.14)", l.At(3))
	}
}

func TestDecodeNestedArray(t *testing.T) {
	res := mustDecode(t, `[[1,2],[3,4]]`)
	l, _ := res.(*values.List)
	inner0, _ := l.At(0).(*values.List)
	inner1, _ := l.At(1).(*values.List)
	if inner0.At(0) != values.Int(1) || inner0.At(1) != values.Int(2) {
		t.Errorf("inner[0] = %v, %v, want 1, 2", inner0.At(0), inner0.At(1))
	}
	if inner1.At(0) != values.Int(3) || inner1.At(1) != values.Int(4) {
		t.Errorf("inner[1] = %v, %v, want 3, 4", inner1.At(0), inner1.At(1))
	}
}

func TestDecodeEmptyObject(t *testing.T) {
	res := mustDecode(t, "{}")
	m, ok := res.(*values.Map)
	if !ok {
		t.Fatalf("Decode({}) should return *values.Map, got %T", res)
	}
	if m.Size() != 0 {
		t.Errorf("Expected empty map, got size %d", m.Size())
	}
}

func TestDecodeObject(t *testing.T) {
	res := mustDecode(t, `{"name":"Alice","age":30}`)
	m, ok := res.(*values.Map)
	if !ok {
		t.Fatalf("Expected *values.Map, got %T", res)
	}
	nameVal, found := m.Get(values.String("name"))
	if !found {
		t.Error("Map missing key 'name'")
	}
	if nameVal != values.String("Alice") {
		t.Errorf("name = %v, want String(Alice)", nameVal)
	}
	ageVal, found := m.Get(values.String("age"))
	if !found {
		t.Error("Map missing key 'age'")
	}
	if ageVal != values.Int(30) {
		t.Errorf("age = %v (type %T), want Int(30)", ageVal, ageVal)
	}
}

func TestDecodeNestedObject(t *testing.T) {
	res := mustDecode(t, `{"user":{"name":"Bob","scores":[90,85]}}`)
	m, ok := res.(*values.Map)
	if !ok {
		t.Fatalf("Expected *values.Map, got %T", res)
	}
	userVal, _ := m.Get(values.String("user"))
	user, _ := userVal.(*values.Map)
	nameVal, _ := user.Get(values.String("name"))
	if nameVal != values.String("Bob") {
		t.Errorf("user.name = %v, want String(Bob)", nameVal)
	}
	scoresVal, _ := user.Get(values.String("scores"))
	scores, _ := scoresVal.(*values.List)
	if scores.At(0) != values.Int(90) || scores.At(1) != values.Int(85) {
		t.Errorf("user.scores = [%v, %v], want [90, 85]", scores.At(0), scores.At(1))
	}
}

// ---- Decode Error Cases ----

func TestDecodeInvalidJSON(t *testing.T) {
	_, err := Decode(values.String(`{invalid}`))
	if err == nil {
		t.Error("Decode of invalid JSON should return an error")
	}
}

func TestDecodeTrailingData(t *testing.T) {
	_, err := Decode(values.String(`42 extra`))
	if err == nil {
		t.Error("Decode with trailing data should return an error")
	}
}

func TestDecodeEmptyString(t *testing.T) {
	_, err := Decode(values.String(``))
	if err == nil {
		t.Error("Decode of empty string should return an error")
	}
}

// ---- Round-Trip ----

func TestRoundTripInt(t *testing.T) {
	original := values.Int(12345)
	encoded := mustEncode(t, original)
	decoded := mustDecode(t, string(encoded))
	if decoded != original {
		t.Errorf("Round-trip failed: %v -> %s -> %v", original, encoded, decoded)
	}
}

func TestRoundTripFloat(t *testing.T) {
	original := values.Float(3.14159)
	encoded := mustEncode(t, original)
	decoded := mustDecode(t, string(encoded))
	if decoded != original {
		t.Errorf("Round-trip failed: %v -> %s -> %v", original, encoded, decoded)
	}
}

func TestRoundTripBool(t *testing.T) {
	for _, b := range []values.Bool{true, false} {
		encoded := mustEncode(t, b)
		decoded := mustDecode(t, string(encoded))
		if decoded != b {
			t.Errorf("Round-trip failed: %v -> %s -> %v", b, encoded, decoded)
		}
	}
}

func TestRoundTripString(t *testing.T) {
	original := values.String("hello world")
	encoded := mustEncode(t, original)
	decoded := mustDecode(t, string(encoded))
	if decoded != original {
		t.Errorf("Round-trip failed: %v -> %s -> %v", original, encoded, decoded)
	}
}

func TestRoundTripStringSpecial(t *testing.T) {
	tests := []string{
		"line\nbreak",
		"tab\there",
		`quote"here`,
		"unicode🚀",
		"\\backslash",
	}
	for _, s := range tests {
		original := values.String(s)
		encoded := mustEncode(t, original)
		decoded := mustDecode(t, string(encoded))
		if decoded != original {
			t.Errorf("Round-trip failed for %q: %v -> %s -> %v", s, original, encoded, decoded)
		}
	}
}

func TestRoundTripList(t *testing.T) {
	original := values.ListOf([]types.Value{
		values.Int(1),
		values.String("two"),
		values.Bool(true),
		values.Float(3.14),
	})
	encoded := mustEncode(t, original)
	decoded := mustDecode(t, string(encoded))
	dl, ok := decoded.(*values.List)
	if !ok {
		t.Fatalf("Round-trip list: expected *values.List, got %T", decoded)
	}
	if dl.Size() != original.Size() {
		t.Fatalf("Round-trip list: size mismatch %d vs %d", dl.Size(), original.Size())
	}
	for i := values.Int(0); i < original.Size(); i++ {
		if dl.At(i) != original.At(i) {
			t.Errorf("Round-trip list[%d]: %v != %v", i, dl.At(i), original.At(i))
		}
	}
}

func TestRoundTripMap(t *testing.T) {
	original := values.NewMap()
	original.Put(values.String("a"), values.Int(1))
	original.Put(values.String("b"), values.String("two"))
	original.Put(values.String("c"), values.Bool(true))

	encoded := mustEncode(t, original)
	decoded := mustDecode(t, string(encoded))
	dm, ok := decoded.(*values.Map)
	if !ok {
		t.Fatalf("Round-trip map: expected *values.Map, got %T", decoded)
	}
	if dm.Size() != original.Size() {
		t.Fatalf("Round-trip map: size mismatch %d vs %d", dm.Size(), original.Size())
	}
	for _, key := range original.Keys() {
		origVal, _ := original.Get(key)
		decVal, found := dm.Get(key)
		if !found {
			t.Errorf("Round-trip map: missing key %v", key)
		}
		if origVal != decVal {
			t.Errorf("Round-trip map[%v]: %v != %v", key, origVal, decVal)
		}
	}
}

func TestRoundTripNested(t *testing.T) {
	inner := values.NewMap()
	inner.Put(values.String("x"), values.Int(10))
	inner.Put(values.String("y"), values.Int(20))

	list := values.ListOf([]types.Value{
		inner,
		values.String("end"),
	})

	encoded := mustEncode(t, list)
	decoded := mustDecode(t, string(encoded))

	dl, _ := decoded.(*values.List)
	if dl.Size() != 2 {
		t.Fatalf("Nested round-trip: expected list size 2, got %d", dl.Size())
	}
	if dl.At(1) != values.String("end") {
		t.Errorf("Nested round-trip: dl[1] = %v, want String(end)", dl.At(1))
	}
	dm, _ := dl.At(0).(*values.Map)
	xVal, _ := dm.Get(values.String("x"))
	if xVal != values.Int(10) {
		t.Errorf("Nested round-trip: inner.x = %v, want Int(10)", xVal)
	}
}

func TestRoundTripNull(t *testing.T) {
	encoded := mustEncode(t, nil)
	decoded := mustDecode(t, string(encoded))
	if decoded != nil {
		t.Errorf("Round-trip null: expected nil, got %v", decoded)
	}
}
