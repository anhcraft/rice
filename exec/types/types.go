//go:generate stringer -type Type

package types

import "reflect"

type Type uint8

const (
	Int Type = iota
	Float
	Bool
	String
	List
	Set
	Map
	Func
	Identifier
	Namespace
	NativeFuncSet
	Selector
	_ // unused
	_ // unused
	_ // special: Null
	_ // special: Any
)

func (t Type) IsNumeric() bool {
	return t == Int || t == Float
}

func (t Type) IsNumericLike() bool {
	return t == Int || t == Float || t == Bool
}

func (t Type) IsPrimitive() bool {
	return t == Int || t == Float || t == Bool || t == String
}

func (t Type) IsMeta() bool {
	return t == Identifier || t == Namespace || t == NativeFuncSet || t == Selector
}

var typeOfs = make(map[reflect.Type]Type)

func (t Type) DefineType(dummy Value) Value {
	x := reflect.TypeOf(dummy)
	if _, ok := typeOfs[x]; ok {
		panic("type already registered")
	}
	typeOfs[x] = t
	return dummy
}

func OfReflect(p reflect.Type) (Type, bool) {
	v, o := typeOfs[p]
	return v, o
}
