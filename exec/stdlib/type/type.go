package _type

import (
	"fmt"
	"github.com/anhcraft/rice/exec/fun"
	"github.com/anhcraft/rice/exec/stdlib"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
)

var Functions = fun.FunctionPackage{
	"typeof":       {stdlib.Define(Typeof)},
	"isNumber":     {stdlib.Define(IsNumber)},
	"isNumberLike": {stdlib.Define(IsNumberLike)},
	"len":          {stdlib.Define(Len)},
	"int":          {stdlib.Define(Int)},
	"float":        {stdlib.Define(Float)},
	"bool":         {stdlib.Define(Bool)},
	"string":       {stdlib.Define(String)},
}

// Typeof finds the type name of the given value; return array for vector.
func Typeof(val types.Value) (types.Value, error) {
	if val == nil {
		return values.String("null"), nil
	}
	return values.String(val.Type().String()), nil
}

// IsNumber checks if the value is either int or float.
func IsNumber(val types.Value) (types.Value, error) {
	return values.Bool(val.Type().IsNumeric()), nil
}

// IsNumberLike checks if the value is one of int, float or bool
func IsNumberLike(val types.Value) (types.Value, error) {
	return values.Bool(val.Type().IsNumericLike()), nil
}

// Len finds the length of the given string or array.
func Len(arg types.Value) (types.Value, error) {
	if v, ok := arg.(values.Collection); ok {
		return v.Size(), nil
	}
	return nil, fmt.Errorf("%T is not a collection", arg)
}

// Int casts the given value to Int.
func Int(arg types.Value) (types.Value, error) {
	return values.AsInt(arg)
}

// Float casts the given value to Float.
func Float(arg types.Value) (types.Value, error) {
	return values.AsFloat(arg)
}

// Bool casts the given value to Bool.
func Bool(arg types.Value) (types.Value, error) {
	return values.AsBool(arg)
}

// String casts the given value to String.
func String(arg types.Value) (types.Value, error) {
	return values.AsString(arg), nil
}
