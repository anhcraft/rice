// Package json provides JSON encoding and decoding for Rice values.
//
//	encode(value) -> String        — encode any Rice value to minified JSON
//	encode(value, indent) -> String — encode any Rice value to pretty-printed JSON with given indent
//	decode(raw) -> any             — decode a JSON string into Rice values
package json

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/anhcraft/rice/exec/fun"
	"github.com/anhcraft/rice/exec/stdlib"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
)

// Functions is the public function package for the json stdlib module.
var Functions = fun.FunctionPackage{
	"encode": {
		stdlib.Define(Encode),
		stdlib.Define(EncodePretty),
	},
	"decode": {
		stdlib.Define(Decode),
	},
}

// Encode converts a Rice value into a minified JSON string.
// Returns an error if the value cannot be encoded (e.g., functions).
func Encode(val types.Value) (types.Value, error) {
	var buf strings.Builder
	if err := encodeValue(val, &buf, "", 0, false); err != nil {
		return nil, err
	}
	return values.String(buf.String()), nil
}

// encodeValue recursively writes the JSON representation of a Rice value to buf.
// When pretty is true, newlines and indentation are added for readability.
func encodeValue(v types.Value, buf *strings.Builder, indent string, level int, pretty bool) error {
	if v == nil {
		buf.WriteString("null")
		return nil
	}

	switch v.Type() {
	case types.Int:
		fmt.Fprint(buf, int64(v.(values.Int)))

	case types.Float:
		f := float64(v.(values.Float))
		if math.IsNaN(f) || math.IsInf(f, 0) {
			buf.WriteString("null")
		} else {
			fmt.Fprint(buf, f)
		}

	case types.Bool:
		if v.(values.Bool) {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}

	case types.String:
		buf.WriteString(strconv.Quote(string(v.(values.String))))

	case types.List:
		list := v.(*values.List)
		if err := encodeCollection(list, buf, indent, level, pretty); err != nil {
			return err
		}

	case types.Set:
		set := v.(*values.Set)
		if err := encodeCollection(set, buf, indent, level, pretty); err != nil {
			return err
		}

	case types.Map:
		m := v.(*values.Map)
		if err := encodeMap(m, buf, indent, level, pretty); err != nil {
			return err
		}

	default:
		return fmt.Errorf("cannot encode value of type %s to JSON", v.Type())
	}

	return nil
}

// writeIndent writes a newline followed by indent * count spaces (when pretty).
func writeIndent(buf *strings.Builder, indent string, count int, pretty bool) {
	if !pretty {
		return
	}
	buf.WriteByte('\n')
	for i := 0; i < count; i++ {
		buf.WriteString(indent)
	}
}

// encodeCollection writes a JSON array: [elem, elem, ...]
func encodeCollection(col values.Collection, buf *strings.Builder, indent string, level int, pretty bool) error {
	buf.WriteByte('[')
	first := true
	for elem := range col.Iterate() {
		if first {
			first = false
		} else {
			buf.WriteByte(',')
		}
		writeIndent(buf, indent, level+1, pretty)
		if err := encodeValue(elem, buf, indent, level+1, pretty); err != nil {
			return err
		}
	}
	if !first && pretty {
		writeIndent(buf, indent, level, pretty)
	}
	buf.WriteByte(']')
	return nil
}

// encodeMap writes a JSON object: {"key": value, ...}
func encodeMap(m *values.Map, buf *strings.Builder, indent string, level int, pretty bool) error {
	buf.WriteByte('{')
	keys := m.Keys()
	first := true
	for _, key := range keys {
		if first {
			first = false
		} else {
			buf.WriteByte(',')
		}
		writeIndent(buf, indent, level+1, pretty)

		// Map keys are stringified via fmt.Sprint (consistent with util.UnwrapIndexedCollection)
		keyStr := strconv.Quote(fmt.Sprint(key))
		buf.WriteString(keyStr)
		buf.WriteByte(':')
		if pretty {
			buf.WriteByte(' ')
		}

		elem, err := m.Element(key)
		if err != nil {
			return err
		}
		if err := encodeValue(elem, buf, indent, level+1, pretty); err != nil {
			return err
		}
	}
	if !first && pretty {
		writeIndent(buf, indent, level, pretty)
	}
	buf.WriteByte('}')
	return nil
}

// Decode parses a JSON string into corresponding Rice types.
//   - JSON objects become *values.Map
//   - JSON arrays become *values.List
//   - JSON numbers become values.Int (if integer) or values.Float
//   - JSON strings become values.String
//   - JSON bools become values.Bool
//   - JSON null becomes nil
func Decode(raw values.String) (types.Value, error) {
	dec := json.NewDecoder(strings.NewReader(string(raw)))

	// Read the first token (skip leading whitespace)
	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}

	val, err := decodeToken(tok, dec)
	if err != nil {
		return nil, err
	}

	// Ensure no trailing non-whitespace data exists
	if dec.More() {
		return nil, fmt.Errorf("json decode: unexpected trailing data")
	}

	// Final read should be the closing delimiter or io.EOF
	if _, err := dec.Token(); err != nil {
		// Expected: io.EOF or nil after complete parse
		// json.Decoder.Token() returns io.EOF at end
		return val, nil
	}

	return val, nil
}

// decodeToken converts a single JSON token into a Rice value.
// For delimiters ([, {), it consumes the rest of the array/object from dec.
func decodeToken(tok json.Token, dec *json.Decoder) (types.Value, error) {
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '[':
			return decodeArray(dec)
		case '{':
			return decodeObject(dec)
		default:
			return nil, fmt.Errorf("json decode: unexpected delimiter %q", t)
		}

	case float64:
		// JSON numbers come as float64; disambiguate Int vs Float
		return numberToValue(t), nil

	case string:
		return values.String(t), nil

	case bool:
		return values.Bool(t), nil

	case nil:
		return nil, nil

	default:
		return nil, fmt.Errorf("json decode: unexpected token type %T", tok)
	}
}

// numberToValue converts a float64 to Int if it has no fractional part
// and fits within int64 range; otherwise returns Float.
func numberToValue(f float64) types.Value {
	if f >= math.MinInt64 && f <= math.MaxInt64 && f == math.Trunc(f) {
		return values.Int(f)
	}
	return values.Float(f)
}

// decodeArray reads a JSON array from dec until ']', returning a *values.List.
func decodeArray(dec *json.Decoder) (types.Value, error) {
	list := values.NewList()

	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("json decode: %w", err)
		}

		// Check for end of array
		if delim, ok := tok.(json.Delim); ok && delim == ']' {
			return list, nil
		}

		val, err := decodeToken(tok, dec)
		if err != nil {
			return nil, err
		}
		list.Append(val)
	}

	// Consume the closing ']'
	if _, err := dec.Token(); err != nil {
		return nil, fmt.Errorf("json decode: expected ']': %w", err)
	}
	return list, nil
}

// decodeObject reads a JSON object from dec until '}', returning a *values.Map.
func decodeObject(dec *json.Decoder) (types.Value, error) {
	m := values.NewMap()

	for dec.More() {
		// Read the key (must be a string)
		tok, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("json decode: %w", err)
		}

		// Check for end of object
		if delim, ok := tok.(json.Delim); ok && delim == '}' {
			return m, nil
		}

		keyStr, ok := tok.(string)
		if !ok {
			return nil, fmt.Errorf("json decode: object key must be string, got %T", tok)
		}

		// Read the value
		valTok, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("json decode: %w", err)
		}

		val, err := decodeToken(valTok, dec)
		if err != nil {
			return nil, err
		}
		m.Put(values.String(keyStr), val)
	}

	// Consume the closing '}'
	if _, err := dec.Token(); err != nil {
		return nil, fmt.Errorf("json decode: expected '}': %w", err)
	}
	return m, nil
}

// EncodePretty converts a Rice value into a pretty-printed JSON string
// using the given indent string for each nesting level.
func EncodePretty(val types.Value, indent values.String) (types.Value, error) {
	var buf strings.Builder
	if err := encodeValue(val, &buf, string(indent), 0, true); err != nil {
		return nil, err
	}
	return values.String(buf.String()), nil
}
