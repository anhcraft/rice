package exec

import (
	"fmt"
	"github.com/anhcraft/rice/exec/ast"
	"github.com/anhcraft/rice/exec/types/values"
	"strings"
)

type RuntimeError struct {
	message string

	// source the caller of the stack frame where the error occurred
	// this is a reliable key to distinct stack frame without keeping references to mem.StackFrame
	source values.CallSite

	// start the starting position of the relevant AST node
	start ast.Pos
	// end the starting position of the relevant AST node
	end ast.Pos

	cause error
}

func (re RuntimeError) causedBy(err error) RuntimeError {
	re.cause = err
	return re
}

func (re RuntimeError) Error() string {
	if re.source.Internal {
		return fmt.Sprintf("%s (internal)", re.message)
	}
	return fmt.Sprintf("%v (at %s-%s)", re.message, re.start, re.end)
}

func (re RuntimeError) Stacktrace() string {
	var sb strings.Builder
	sb.WriteString("RuntimeError:\n")
	buildErrorStacktrace(&sb, &re, 0)
	return sb.String()
}
