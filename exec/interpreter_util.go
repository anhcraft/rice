package exec

import (
	"fmt"

	"github.com/anhcraft/rice/exec/ast"
	"github.com/anhcraft/rice/exec/profiler"
)

func (i *Interpreter) Profiler() profiler.Profiler {
	return i.profiler
}

func (i *Interpreter) cleanUp() {
	i.profiler.Reset()
	i.env.Reset()
	i.dirty = false
}

func (i *Interpreter) throw(stmt ast.Stmt, msg string, args ...any) RuntimeError {
	return RuntimeError{
		message: fmt.Sprintf(msg, args...),
		source:  i.env.CurrentFrame().CallSite(),
		start:   stmt.StartPos(),
		end:     stmt.EndPos(),
	}
}

func (i *Interpreter) checkContext() error {
	select {
	case <-i.ctx.Done():
		return i.ctx.Err()
	default:
		return nil
	}
}
