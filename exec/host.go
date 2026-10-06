package exec

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/anhcraft/rice/exec/ast"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
	"github.com/anhcraft/rice/exec/vm"
)

var _ vm.Host = (*Interpreter)(nil)

func (i *Interpreter) LoadGlobal(name string) (types.Value, error) {
	id := values.Identifier(name)
	if ns, err := i.env.Namespace().Element(id); ns != nil && err == nil {
		return ns, nil
	}
	if val, ok := i.env.Get(id); ok {
		return val, nil
	}
	return nil, fmt.Errorf("unresolved reference %q", id)
}

func (i *Interpreter) StoreGlobal(name string, value types.Value) error {
	return i.env.Assign(values.Identifier(name), value)
}

func (i *Interpreter) DefineGlobal(name string, value types.Value, constant bool) error {
	id := values.Identifier(name)
	if i.env.IsNamespaceEntry(id) {
		return fmt.Errorf("cannot declare %q because it conflicts with a built-in function or namespace", id)
	}
	if !i.env.Define(id, value, constant) {
		return fmt.Errorf("cannot redeclare %q", id)
	}
	return nil
}

func (i *Interpreter) IsNamespaceEntry(name string) bool {
	return i.env.IsNamespaceEntry(values.Identifier(name))
}

func (i *Interpreter) TypeBound(obj types.Value, name values.Identifier) (types.Value, bool) {
	if tries, ok := i.typeBoundFuncPkg[obj.Type()]; ok {
		if trie, ok := tries[name]; ok {
			return buildNativeFuncSet(obj, name, trie, i.nativeFuncTimeout), true
		}
	}
	return nil, false
}

func (i *Interpreter) ProfilerStart(label string, pos ast.Pos) {
	i.profiler.StartAt(label, pos)
}

func (i *Interpreter) ProfilerEnd() {
	i.profiler.End()
}

func (i *Interpreter) ProfilerDepth() int {
	return i.profiler.Depth()
}

func (i *Interpreter) ProfilerUnwind(depth int) {
	i.profiler.Unwind(depth)
}

func (i *Interpreter) Context() context.Context {
	return i.ctx
}

func (i *Interpreter) SetContext(ctx context.Context) {
	i.ctx = ctx
}

func (i *Interpreter) CheckContext() error {
	return i.checkContext()
}

func (i *Interpreter) ScopeLimit() uint32 {
	return i.lexicalScopeLimit
}

func (i *Interpreter) UserFuncTimeout() time.Duration {
	return i.userFuncTimeout
}

func faultToRuntime(err error) error {
	if err == nil {
		return nil
	}
	var f vm.Fault
	if !errors.As(err, &f) {
		return err
	}
	re := RuntimeError{
		message: f.Message,
		source:  f.Source,
		start:   f.Start,
		end:     f.End,
	}
	if f.Cause != nil {
		re.cause = faultToRuntime(f.Cause)
	}
	return re
}
