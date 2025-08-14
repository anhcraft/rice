package mem

import "rice/exec/types"

type Variable struct {
	value    types.Value
	constant bool
}
