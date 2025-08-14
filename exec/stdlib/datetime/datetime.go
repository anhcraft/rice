package datetime

import (
	"github.com/anhcraft/rice/exec/fun"
	"github.com/anhcraft/rice/exec/stdlib"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
	"time"
)

var Functions = fun.FunctionPackage{
	"now": {stdlib.Define(Now)},
}

// Now returns the current Unix timestamp in milliseconds.
func Now() (types.Value, error) {
	return values.Int(time.Now().UnixMilli()), nil
}
