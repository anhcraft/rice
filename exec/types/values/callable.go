package values

import (
	"context"
	"fmt"
	"github.com/anhcraft/rice/exec/ast"
	"github.com/anhcraft/rice/exec/types"
)

// Callable denotes a value that could be called with a list of arguments
type Callable interface {
	Call(ctx context.Context, site CallSite, args []types.Value) (types.Value, error)
}

// CallSite informs where the call was started
type CallSite struct {
	// Caller name of the caller
	Caller string

	// Internal if the call is from internal/native code
	Internal bool

	// StartPos the start position of ast.CallExpr (for Internal=false)
	StartPos ast.Pos

	// EndPos the end position of ast.CallExpr (for Internal=false)
	EndPos ast.Pos
}

func (call CallSite) String() string {
	if call.Internal {
		return fmt.Sprintf("%s (internal)", call.Caller)
	}

	return fmt.Sprintf("%s (at %s-%s)", call.Caller, call.StartPos, call.EndPos)
}

// InternalCallSite reusable internal context
var InternalCallSite = CallSite{Internal: true, Caller: "<internal>"}

// RootCallSite reusable root context
var RootCallSite = CallSite{Internal: false, Caller: "<root>"}
