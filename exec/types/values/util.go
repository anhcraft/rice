package values

import (
	"errors"
)

const epsilon = 1e-9

var indexOutOfBoundErr = errors.New("index out of bound")
var elementNotIntErr = errors.New("element type is not Int")
var elementNotIdErr = errors.New("element type is not Identifier")

func init() {
	//types.Map.DefineType((*Map)(nil))
}
