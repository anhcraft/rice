package datetime

import (
	"rice/exec/fun"
	"rice/exec/stdlib"
	"rice/exec/types"
	"rice/exec/types/values"
	"time"
)

var Functions = fun.FunctionPackage{
	"now": {stdlib.Define(Now)},
}

// Now returns the current Unix timestamp in milliseconds.
func Now() (types.Value, error) {
	return values.Int(time.Now().UnixMilli()), nil
}
