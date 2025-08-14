package values

import (
	"fmt"
	"github.com/anhcraft/rice/exec/types"
	"strconv"
)

var _ = types.Float.DefineType(Float(0))
var _ Primitive = Float(0)

type Float float64

func (f Float) Type() types.Type {
	return types.Float
}

func (f Float) ToInt() (Int, error) {
	return AsInt(f)
}

func (f Float) ToFloat() (Float, error) {
	return f, nil
}

func (f Float) ToBool() (Bool, error) {
	return AsBool(f)
}

func AsFloat(val any) (Float, error) {
	switch v := val.(type) {
	case Int:
		return Float(v), nil
	case Float:
		return v, nil
	case Bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	case String:
		f, err := strconv.ParseFloat(string(v), 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string %q to Float: %w", v, err)
		}
		return Float(f), nil
	case int:
		return Float(v), nil
	case int8:
		return Float(v), nil
	case int16:
		return Float(v), nil
	case int32:
		return Float(v), nil
	case int64:
		return Float(v), nil
	case uint:
		return Float(v), nil
	case uint8:
		return Float(v), nil
	case uint16:
		return Float(v), nil
	case uint32:
		return Float(v), nil
	case uint64:
		return Float(v), nil
	case float32:
		return Float(v), nil
	case float64:
		return Float(v), nil
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string %q to Float: %w", v, err)
		}
		return Float(f), nil
	case nil:
		return 0.0, nil
	default:
		return 0.0, fmt.Errorf("cannot cast %T to Float", val)
	}
}

func IsFloat(val any) bool {
	_, ok := val.(Float)
	return ok
}
