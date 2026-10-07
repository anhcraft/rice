package values

import (
	"context"
	"github.com/anhcraft/rice/exec/types"
)

type NativeFunctionSetDelegate func(ctx context.Context, self NativeFunctionSet, site CallSite, args []types.Value) (types.Value, error)

var _ = types.NativeFuncSet.DefineType(NativeFunctionSet{})
var _ Callable = NativeFunctionSet{}

type NativeFunctionSet struct {
	boundValue types.Value
	name       Identifier
	delegate   NativeFunctionSetDelegate
}

func NewNativeFunctionSet(val types.Value, name Identifier, delegate NativeFunctionSetDelegate) NativeFunctionSet {
	return NativeFunctionSet{boundValue: val, name: name, delegate: delegate}
}

func (f NativeFunctionSet) String() string {
	return "NativeFunctionSet"
}

func (f NativeFunctionSet) Name() Identifier {
	return f.name
}

func (f NativeFunctionSet) Type() types.Type {
	return types.NativeFuncSet
}

func (f NativeFunctionSet) Call(ctx context.Context, site CallSite, args []types.Value) (types.Value, error) {
	return f.delegate(ctx, f, site, args)
}
