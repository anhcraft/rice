package mem

import "github.com/anhcraft/rice/exec/types"

type Variable struct {
	value    types.Value
	constant bool
}
