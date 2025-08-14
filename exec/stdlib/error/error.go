package error

import (
	"fmt"
	"rice/exec/fun"
	"rice/exec/stdlib"
	"rice/exec/types"
	"rice/exec/types/values"
)

var Functions = fun.FunctionPackage{
	"throw":  {stdlib.Define(Throw)},
	"assert": {stdlib.Define(Assert)},
}

// Throw throws a custom error.
func Throw(err values.String) (types.Value, error) {
	return values.Bool(false), fmt.Errorf("throw: %s", err)
}

// Assert throws a custom error when the condition is false
func Assert(cond values.Bool, err values.String) (types.Value, error) {
	if !cond {
		return cond, fmt.Errorf("assert: %s", err)
	}
	return cond, nil
}
