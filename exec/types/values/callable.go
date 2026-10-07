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

	return fmt.Sprintf("%s (%s-%s)", call.Caller, call.StartPos.String(), call.EndPos.String())
}

// Equal compares call sites by identity fields, ignoring Pos display caches.
func (call CallSite) Equal(other CallSite) bool {
	return call.Caller == other.Caller &&
		call.Internal == other.Internal &&
		call.StartPos.Line == other.StartPos.Line &&
		call.StartPos.Index == other.StartPos.Index &&
		call.EndPos.Line == other.EndPos.Line &&
		call.EndPos.Index == other.EndPos.Index
}

type callSiteKey struct{}

// WithCallSite stores the script call site on ctx for native functions and callbacks.
func WithCallSite(ctx context.Context, site CallSite) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, callSiteKey{}, site)
}

// CallSiteFrom returns the CallSite previously stored with WithCallSite.
func CallSiteFrom(ctx context.Context) (CallSite, bool) {
	if ctx == nil {
		return CallSite{}, false
	}
	site, ok := ctx.Value(callSiteKey{}).(CallSite)
	return site, ok
}

// CallbackCallSite is the site used when stdlib invokes a user lambda.
// It stays distinct from the outer native call while keeping that call's span.
func CallbackCallSite(ctx context.Context) CallSite {
	parent, ok := CallSiteFrom(ctx)
	if !ok {
		return CallSite{Internal: true, Caller: "<callback>"}
	}
	return CallSite{
		Caller:   "<callback>",
		Internal: false,
		StartPos: parent.StartPos,
		EndPos:   parent.EndPos,
	}
}

// InternalCallSite reusable internal context
var InternalCallSite = CallSite{Internal: true, Caller: "<internal>"}

// RootCallSite reusable root context
var RootCallSite = CallSite{Internal: false, Caller: "<root>"}
