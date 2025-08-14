package values

import (
	"context"
	"rice/exec/types"
)

type NativeFunctionSetDelegate func(ctx context.Context, self NativeFunctionSet, site CallSite, args []types.Value) (types.Value, error)

var _ = types.NativeFuncSet.DefineType(NativeFunctionSet{})
var _ Callable = NativeFunctionSet{}

type NativeFunctionSet struct {
	boundValue types.Value
	delegate   NativeFunctionSetDelegate
}

func NewNativeFunctionSet(val types.Value, delegate NativeFunctionSetDelegate) NativeFunctionSet {
	return NativeFunctionSet{boundValue: val, delegate: delegate}
}

func (f NativeFunctionSet) String() string {
	return "NativeFunctionSet"
}

func (f NativeFunctionSet) Type() types.Type {
	return types.NativeFuncSet
}

func (f NativeFunctionSet) Call(ctx context.Context, site CallSite, args []types.Value) (types.Value, error) {
	return f.delegate(ctx, f, site, args)
}
