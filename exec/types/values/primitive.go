package values

import (
	"errors"
	"rice/exec/types"
)

type Primitive interface {
	types.Value
	ToInt() (Int, error)
	ToFloat() (Float, error)
	ToBool() (Bool, error)
}

// ConvertPrimitiveImplicitly converts the two primitives values into a uniform type
func ConvertPrimitiveImplicitly(a Primitive, b Primitive) (Primitive, Primitive, error) {
	// follows top to bottom precedence

	if IsString(a) || IsString(b) {
		r1 := AsString(a)
		r2 := AsString(b)
		return r1, r2, nil
	}

	if IsFunc(a) || IsFunc(b) {
		return nil, nil, errors.New("func cannot be implicitly converted")
	}

	if IsBool(a) || IsBool(b) {
		r1, e1 := a.ToBool()
		if e1 != nil {
			return nil, nil, e1
		}

		r2, e2 := b.ToBool()
		if e2 != nil {
			return nil, nil, e2
		}

		return r1, r2, nil
	}

	if IsFloat(a) || IsFloat(b) {
		r1, e1 := a.ToFloat()
		if e1 != nil {
			return nil, nil, e1
		}

		r2, e2 := b.ToFloat()
		if e2 != nil {
			return nil, nil, e2
		}

		return r1, r2, nil
	}

	if IsInt(a) || IsInt(b) {
		r1, e1 := a.ToInt()
		if e1 != nil {
			return nil, nil, e1
		}

		r2, e2 := b.ToInt()
		if e2 != nil {
			return nil, nil, e2
		}

		return r1, r2, nil
	}

	return a, b, nil
}
