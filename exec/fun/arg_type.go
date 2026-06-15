package fun

import (
	"fmt"
	"github.com/anhcraft/rice/exec/types"
)

// ArgType the first 4 bits is dimension; the last 4 bits is type ID
type ArgType uint8

const maxArgType = ArgType((1 << 8) - 1)

// Note: sort the type by precedence from the most specific

func NewArgType(dim uint8, tp types.Type) ArgType {
	return ArgType((dim << 4) | uint8(tp))
}

func NewArgTypeAny(dim uint8) ArgType {
	return NewArgType(dim, 15)
}

// NewNullArgType returns an ArgType representing a null argument value.
// null is a value, not a types.Type — this uses slot 14 which is unused in types.Type.
func NewNullArgType() ArgType {
	return NewArgType(0, 14)
}

func (t ArgType) Dim() uint8 {
	return uint8((t >> 4) & 0xF)
}

func (t ArgType) Type() uint8 {
	return uint8(t & 0xF)
}

func (t ArgType) IsAny() bool {
	return t.Type() == 15
}

func (t ArgType) IsNull() bool {
	// null is a value, not a types.Type — uses unused type-ID slot 14
	return t.Type() == 14 && t.Dim() == 0
}

func (t ArgType) MatchAnyType() bool {
	// Go allows `any` to match any kind of value
	// But if dim > 0, then `any` can only match `any` (e.g. `[]any != []int`)
	return t.IsAny() && t.Dim() == 0
}

func (t ArgType) CanAccept(other ArgType) bool {
	if other.IsNull() {
		// null is compatible with non-primitive types and 'any', but not primitives
		return t.MatchAnyType() || (t.Dim() == 0 && !types.Type(t.Type()).IsPrimitive())
	}
	return t.MatchAnyType() || (t.Type() == other.Type() && t.Dim() == other.Dim())
}

func (t ArgType) CanContainMultiOf(other ArgType) bool {
	if other.IsNull() {
		// ...any can hold null
		if t.IsAny() {
			return true
		}
		// ...T where T is non-primitive can hold null
		return t.Dim() == 1 && !types.Type(t.Type()).IsPrimitive()
	}
	return t.IsAny() && t.Dim() == 1 || // `...any` can hold `<any>, <any>, ...`
		(t.Type() == other.Type() && t.Dim() == other.Dim()+1)
}

func (t ArgType) String() string {
	return fmt.Sprintf("ArgType(t=%d,d=%d)", t.Type(), t.Dim())
}
