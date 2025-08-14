package values

import (
	"fmt"
	"github.com/anhcraft/rice/exec/types"
	"strconv"
)

var _ = types.Int.DefineType(Int(0))
var _ Primitive = Int(0)

type Int int64

func (i Int) Type() types.Type {
	return types.Int
}

func (i Int) ToInt() (Int, error) {
	return i, nil
}

func (i Int) ToFloat() (Float, error) {
	return AsFloat(i)
}

func (i Int) ToBool() (Bool, error) {
	return AsBool(i)
}

func AsInt(val any) (Int, error) {
	switch v := val.(type) {
	case Int:
		return v, nil
	case Float:
		return Int(v), nil
	case Bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case String:
		i, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string %q to Int: %w", v, err)
		}
		return Int(i), nil
	case int:
		return Int(v), nil
	case int8:
		return Int(v), nil
	case int16:
		return Int(v), nil
	case int32:
		return Int(v), nil
	case int64:
		return Int(v), nil
	case uint:
		return Int(v), nil
	case uint8:
		return Int(v), nil
	case uint16:
		return Int(v), nil
	case uint32:
		return Int(v), nil
	case uint64:
		return Int(v), nil
	case float32:
		return Int(v), nil
	case float64:
		return Int(v), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string %q to Int: %w", v, err)
		}
		return Int(i), nil
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot cast %T to Int", val)
	}
}

func IsInt(val any) bool {
	_, ok := val.(Int)
	return ok
}
