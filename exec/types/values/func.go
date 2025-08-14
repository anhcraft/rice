package values

import (
	"context"
	"errors"
	"rice/exec/types"
	"strings"
)

type FuncDelegate func(ctx context.Context, self *Func, site CallSite, args []types.Value) (types.Value, error)

var _ = types.Func.DefineType((*Func)(nil))
var _ Callable = (*Func)(nil)

type Func struct {
	params   []Identifier
	variadic bool
	closure  any // *mem.LexicalScope
	delegate FuncDelegate
}

func NewFunc(params []Identifier, variadic bool, closure any, delegate FuncDelegate) *Func {
	if variadic && len(params) == 0 {
		panic(errors.New("variadic parameter cannot be empty"))
	}
	return &Func{params, variadic, closure, delegate}
}

func (f *Func) Type() types.Type {
	return types.Func
}

func (f *Func) Arity() int {
	return len(f.params)
}

func (f *Func) Param(i int) Identifier {
	return f.params[i]
}

func (f *Func) Variadic() bool {
	return f.variadic
}

func (f *Func) Closure() any {
	return f.closure
}

func (f *Func) Call(ctx context.Context, site CallSite, args []types.Value) (types.Value, error) {
	return f.delegate(ctx, f, site, args)
}

func (f *Func) ToInt() (Int, error) {
	return 0, errors.New("cannot cast Func to Int")
}

func (f *Func) ToFloat() (Float, error) {
	return 0, errors.New("cannot cast Func to Float")
}

func (f *Func) ToBool() (Bool, error) {
	return false, errors.New("cannot cast Func to Bool")
}

func (f *Func) String() string {
	sb := strings.Builder{}
	sb.WriteString("func(")
	for i, param := range f.params {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(string(param))
	}
	if f.variadic {
		sb.WriteString("...")
	}
	sb.WriteString(")")
	return sb.String()
}

func IsFunc(val any) bool {
	_, ok := val.(Func)
	return ok
}
