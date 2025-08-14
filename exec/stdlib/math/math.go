package math

import (
	"fmt"
	"github.com/anhcraft/rice/exec/fun"
	"github.com/anhcraft/rice/exec/stdlib"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
	"math"
)

var Functions = fun.FunctionPackage{
	"max": {stdlib.Define(Max)},
	"min": {stdlib.Define(Min)},
	"abs": {stdlib.DefineAndMap(Abs, func(def *fun.FunctionDef) {
		def.DefineArg(0,
			fun.NewArgType(0, types.Int),
			fun.NewArgType(0, types.Float),
			fun.NewArgType(0, types.Bool),
		)
	})},
	"sqrt": {stdlib.DefineAndMap(Sqrt, func(def *fun.FunctionDef) {
		def.DefineArg(0,
			fun.NewArgType(0, types.Int),
			fun.NewArgType(0, types.Float),
			fun.NewArgType(0, types.Bool),
		)
	})},
	"floor": {stdlib.DefineAndMap(Floor, func(def *fun.FunctionDef) {
		def.DefineArg(0,
			fun.NewArgType(0, types.Int),
			fun.NewArgType(0, types.Float),
			fun.NewArgType(0, types.Bool),
		)
	})},
	"ceil": {stdlib.DefineAndMap(Ceil, func(def *fun.FunctionDef) {
		def.DefineArg(0,
			fun.NewArgType(0, types.Int),
			fun.NewArgType(0, types.Float),
			fun.NewArgType(0, types.Bool),
		)
	})},
	"pow": {stdlib.DefineAndMap(Pow, func(def *fun.FunctionDef) {
		def.DefineArg(0,
			fun.NewArgType(0, types.Int),
			fun.NewArgType(0, types.Float),
			fun.NewArgType(0, types.Bool),
		)
		def.DefineArg(1,
			fun.NewArgType(0, types.Int),
			fun.NewArgType(0, types.Float),
			fun.NewArgType(0, types.Bool),
		)
	})},
}

// Max selects numeric items and returns the maximum value.
// It returns `NaN` if any of the values is `NaN`, and `nil` if no numeric values are provided.
func Max(args ...types.Value) (types.Value, error) {
	var (
		maxInt   values.Int
		haveInt  bool
		maxFloat values.Float
		haveF    bool
	)

	for _, value := range args {
		switch v := value.(type) {
		case values.Int:
			if !haveInt {
				maxInt, haveInt = v, true
			} else if v > maxInt {
				maxInt = v
			}
		case values.Float:
			if math.IsNaN(float64(v)) {
				return values.Float(math.NaN()), nil
			}
			if !haveF {
				maxFloat, haveF = v, true
			} else if v > maxFloat {
				maxFloat = v
			}
		}
	}

	switch {
	case !haveInt && !haveF:
		return nil, nil
	case !haveF:
		return maxInt, nil
	case !haveInt:
		return maxFloat, nil
	default:
		if float64(maxInt) > float64(maxFloat) {
			return maxInt, nil
		}
		return maxFloat, nil
	}
}

// Min selects numeric items and returns the minimum value.
// It returns `NaN` if any of the values is `NaN`, and `nil` if no numeric values are provided.
func Min(args ...types.Value) (types.Value, error) {
	var (
		minInt   values.Int
		haveInt  bool
		minFloat values.Float
		haveF    bool
	)

	for _, value := range args {
		switch v := value.(type) {
		case values.Int:
			if !haveInt {
				minInt, haveInt = v, true
			} else if v < minInt {
				minInt = v
			}
		case values.Float:
			if math.IsNaN(float64(v)) {
				return values.Float(math.NaN()), nil
			}
			if !haveF {
				minFloat, haveF = v, true
			} else if v < minFloat {
				minFloat = v
			}
		}
	}

	switch {
	case !haveInt && !haveF:
		return nil, nil
	case !haveF:
		return minInt, nil
	case !haveInt:
		return minFloat, nil
	default:
		if float64(minInt) < float64(minFloat) {
			return minInt, nil
		}
		return minFloat, nil
	}
}

// Abs returns the absolute of a numeric value.
func Abs(value types.Value) (types.Value, error) {
	switch v := value.(type) {
	case values.Int:
		if v < 0 {
			return -v, nil
		}
		return v, nil
	case values.Float:
		return values.Float(math.Abs(float64(v))), nil
	default:
		return nil, fmt.Errorf("abs expects Int or Float, but got %T", value)
	}
}

// Sqrt returns the square root of a numeric value. The result is always a Float.
func Sqrt(number types.Value) (types.Value, error) {
	v, err := values.AsFloat(number)
	if err != nil {
		return nil, fmt.Errorf("sqrt expects a numeric value: %w", err)
	}
	if v < 0 {
		return values.Float(math.NaN()), nil
	}
	return values.Float(math.Sqrt(float64(v))), nil
}

// Floor returns the greatest integer value less than or equal to x.
func Floor(number types.Value) (types.Value, error) {
	if v, ok := number.(values.Float); ok {
		return values.Float(math.Floor(float64(v))), nil
	}

	v, err := values.AsFloat(number)
	if err != nil {
		return nil, fmt.Errorf("floor expects a numeric value: %w", err)
	}
	return values.Int(math.Floor(float64(v))), nil
}

// Ceil returns the least integer value greater than or equal to x.
func Ceil(number types.Value) (types.Value, error) {
	if v, ok := number.(values.Float); ok {
		return values.Float(math.Ceil(float64(v))), nil
	}

	v, err := values.AsFloat(number)
	if err != nil {
		return nil, fmt.Errorf("ceil expects a numeric value: %w", err)
	}
	return values.Int(math.Ceil(float64(v))), nil
}

// Pow returns base to the power of the exponent. The result is always a Float.
func Pow(base types.Value, exponent types.Value) (types.Value, error) {
	baseFloat, err := values.AsFloat(base)
	if err != nil {
		return nil, fmt.Errorf("base must be a numeric value: %w", err)
	}

	exponentFloat, err := values.AsFloat(exponent)
	if err != nil {
		return nil, fmt.Errorf("exponent must be a numeric value: %w", err)
	}

	result := math.Pow(float64(baseFloat), float64(exponentFloat))
	return values.Float(result), nil
}
