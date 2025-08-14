package values

import (
	"fmt"
	"math"
	"rice/exec/types"
	"strconv"
)

var _ = types.Bool.DefineType(Bool(false))
var _ Primitive = Bool(false)

type Bool bool

func (b Bool) Type() types.Type {
	return types.Bool
}

func (b Bool) ToInt() (Int, error) {
	return AsInt(b)
}

func (b Bool) ToFloat() (Float, error) {
	return AsFloat(b)
}

func (b Bool) ToBool() (Bool, error) {
	return b, nil
}

func AsBool(val any) (Bool, error) {
	switch v := val.(type) {
	case Int:
		return v != 0, nil
	case Float:
		return math.Abs(float64(v)) > epsilon, nil
	case Bool:
		return v, nil
	case String:
		b, err := strconv.ParseBool(string(v))
		if err != nil {
			return false, fmt.Errorf("cannot convert string %q to Bool: %w", v, err)
		}
		return Bool(b), nil
	case int:
		return v != 0, nil
	case int8:
		return v != 0, nil
	case int16:
		return v != 0, nil
	case int32:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case uint:
		return v != 0, nil
	case uint8:
		return v != 0, nil
	case uint16:
		return v != 0, nil
	case uint32:
		return v != 0, nil
	case uint64:
		return v != 0, nil
	case float32:
		return math.Abs(float64(v)) > epsilon, nil
	case float64:
		return math.Abs(v) > epsilon, nil
	case bool:
		return Bool(v), nil
	case string:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return false, fmt.Errorf("cannot convert string %q to Bool: %w", v, err)
		}
		return Bool(b), nil
	case nil:
		return false, nil
	default:
		return false, fmt.Errorf("cannot cast %T to Bool", val)
	}
}

func IsBool(val any) bool {
	_, ok := val.(Bool)
	return ok
}
