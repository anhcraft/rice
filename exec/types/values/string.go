package values

import (
	"errors"
	"fmt"
	"github.com/anhcraft/rice/exec/types"
	"iter"
	"unicode/utf8"
)

var _ = types.String.DefineType(String(""))
var _ Primitive = String("")
var _ IndexedCollection = String("")

// NOTE: All operations should support UTF8; no byte-level operation

type String string

func (s String) Type() types.Type {
	return types.String
}

func (s String) Element(id types.Value) (types.Value, error) {
	if i, ok := id.(Int); ok {
		runes := []rune(s)
		if i < 0 || int(i) >= len(runes) {
			return nil, indexOutOfBoundErr
		}
		return String(runes[i]), nil
	} else {
		return nil, elementNotIntErr
	}
}

func (s String) PutElement(id types.Value, item types.Value) error {
	return errors.New("string is immutable")
}

func (s String) Size() Int {
	return Int(utf8.RuneCountInString(string(s)))
}

func (s String) Keys() []types.Value {
	k := make([]types.Value, s.Size())
	for i := 0; i < len(k); i++ {
		k[i] = Int(i)
	}
	return k
}

func (s String) Iterate() iter.Seq[types.Value] {
	return func(yield func(types.Value) bool) {
		runes := []rune(s)

		for i := 0; i < len(runes); i++ {
			if !yield(String(runes[i])) {
				return
			}
		}
	}
}

func (s String) ToInt() (Int, error) {
	return AsInt(s)
}

func (s String) ToFloat() (Float, error) {
	return AsFloat(s)
}

func (s String) ToBool() (Bool, error) {
	return AsBool(s)
}

func (s String) ToString() (String, error) {
	return s, nil
}

func AsString(val any) String {
	if val == nil {
		return ""
	}
	if v, ok := val.(String); ok {
		return v
	}
	if v, ok := val.(string); ok {
		return String(v)
	}
	return String(fmt.Sprint(val))
}

func IsString(val any) bool {
	_, ok := val.(String)
	return ok
}
