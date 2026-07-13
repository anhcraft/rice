package duration

import (
	"fmt"
	"time"

	"github.com/anhcraft/rice/exec/fun"
	"github.com/anhcraft/rice/exec/stdlib"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
)

var Functions = fun.FunctionPackage{
	"parse":   {stdlib.Define(Parse)},
	"days":    {stdlib.DefineAndMap(Days, widenNumeric)},
	"hours":   {stdlib.DefineAndMap(Hours, widenNumeric)},
	"minutes": {stdlib.DefineAndMap(Minutes, widenNumeric)},
	"seconds": {stdlib.DefineAndMap(Seconds, widenNumeric)},
	"millis":  {stdlib.DefineAndMap(Millis, widenNumeric)},
}

func widenNumeric(def *fun.FunctionDef) {
	def.DefineArg(0,
		fun.NewArgType(0, types.Int),
		fun.NewArgType(0, types.Float),
		fun.NewArgType(0, types.Bool),
	)
}

// Parse parses a duration string and returns the duration in milliseconds.
// Delegates to Go's time.ParseDuration, which supports ns, us/µs, ms, s, m, h.
func Parse(s values.String) (types.Value, error) {
	d, err := time.ParseDuration(string(s))
	if err != nil {
		return nil, fmt.Errorf("duration.parse: %w", err)
	}
	return values.Int(d.Milliseconds()), nil
}

func Days(n types.Value) (types.Value, error) {
	v, err := values.AsFloat(n)
	if err != nil {
		return nil, fmt.Errorf("duration.days expects a numeric value: %w", err)
	}
	return values.Int(int64(float64(v) * 86400000)), nil
}

func Hours(n types.Value) (types.Value, error) {
	v, err := values.AsFloat(n)
	if err != nil {
		return nil, fmt.Errorf("duration.hours expects a numeric value: %w", err)
	}
	return values.Int(int64(float64(v) * 3600000)), nil
}

func Minutes(n types.Value) (types.Value, error) {
	v, err := values.AsFloat(n)
	if err != nil {
		return nil, fmt.Errorf("duration.minutes expects a numeric value: %w", err)
	}
	return values.Int(int64(float64(v) * 60000)), nil
}

func Seconds(n types.Value) (types.Value, error) {
	v, err := values.AsFloat(n)
	if err != nil {
		return nil, fmt.Errorf("duration.seconds expects a numeric value: %w", err)
	}
	return values.Int(int64(float64(v) * 1000)), nil
}

func Millis(n types.Value) (types.Value, error) {
	v, err := values.AsFloat(n)
	if err != nil {
		return nil, fmt.Errorf("duration.millis expects a numeric value: %w", err)
	}
	return values.Int(int64(float64(v))), nil
}
