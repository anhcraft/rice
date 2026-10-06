package vm

import (
	"fmt"

	"github.com/anhcraft/rice/exec/ast"
	"github.com/anhcraft/rice/exec/types/values"
)

// Fault is a VM runtime failure. The interpreter converts it to exec.RuntimeError.
type Fault struct {
	Message string
	Source  values.CallSite
	Start   ast.Pos
	End     ast.Pos
	Cause   error
}

func (f Fault) Error() string {
	if f.Source.Internal {
		return fmt.Sprintf("%s (internal)", f.Message)
	}
	return fmt.Sprintf("%s (L%d-L%d)", f.Message, f.Start.Line, f.End.Line)
}

func (f Fault) Unwrap() error { return f.Cause }

func (f Fault) causedBy(err error) Fault {
	f.Cause = err
	return f
}
