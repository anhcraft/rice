package vm

import (
	"context"
	"time"

	"github.com/anhcraft/rice/exec/ast"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
)

// Host is implemented by the interpreter. exec/vm must not import package exec.
type Host interface {
	LoadGlobal(name string) (types.Value, error)
	StoreGlobal(name string, value types.Value) error
	DefineGlobal(name string, value types.Value, constant bool) error
	IsNamespaceEntry(name string) bool

	TypeBound(obj types.Value, name values.Identifier) (types.Value, bool)

	ProfilerStart(label string, pos ast.Pos)
	ProfilerEnd()
	ProfilerDepth() int
	ProfilerUnwind(depth int)

	Context() context.Context
	SetContext(ctx context.Context)
	CheckContext() error
	ScopeLimit() uint32
	UserFuncTimeout() time.Duration
}
