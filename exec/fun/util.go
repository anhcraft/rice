package fun

import (
	"context"
	"fmt"
	"github.com/anhcraft/rice/exec/types"
	"reflect"
)

var (
	typeOfValue   = reflect.TypeOf((*types.Value)(nil)).Elem()
	typeOfContext = reflect.TypeOf((*context.Context)(nil)).Elem()
)

func getBaseTypeAndDimension(t reflect.Type) (reflect.Type, uint8) {
	dim := 0
	for t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		dim++
		t = t.Elem()
	}
	return t, uint8(dim)
}

func isAny(t reflect.Type) bool {
	return t.Kind() == reflect.Interface && t.NumMethod() == 0
}

func getArgType(t reflect.Type) (ArgType, error) {
	if t == nil {
		return NewNullArgType(), nil
	}

	baseType, dim := getBaseTypeAndDimension(t)

	if baseType == typeOfValue {
		return NewArgTypeAny(dim), nil
	}

	if baseType.Implements(typeOfValue) {
		if tp, ok := types.OfReflect(baseType); ok {
			return NewArgType(dim, tp), nil
		}
	}

	if isAny(baseType) {
		return NewArgTypeAny(dim), nil
	}

	return 0, fmt.Errorf("unsupported arg type %v", baseType)
}
