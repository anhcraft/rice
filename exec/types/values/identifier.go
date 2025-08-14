package values

import "rice/exec/types"

var _ = types.Identifier.DefineType(Identifier(""))

type Identifier string

func (i Identifier) Type() types.Type {
	return types.Identifier
}
