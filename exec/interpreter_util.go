package exec

import (
	"fmt"

	"github.com/anhcraft/rice/exec/ast"
	"github.com/anhcraft/rice/exec/profiler"
	"github.com/anhcraft/rice/exec/vm"
)

func (i *Interpreter) Profiler() profiler.Profiler {
	return i.profiler
}

// Debug enables single-stepping on the interpreter VM. Stepping only waits
// when the binary is built with -tags rice_debug. Call it before Interpret.
func (i *Interpreter) Debug() {
	i.vm.Debug()
}

// Step allows one opcode to run. Safe to call from another goroutine while Interpret is running.
func (i *Interpreter) Step() {
	i.vm.Step()
}

// Position yields VM locations after each opcode. Range it from another goroutine.
func (i *Interpreter) Position() <-chan vm.DebugPos {
	return i.vm.Position()
}

// CallStack copies the current call frames. Read it after a Position event and before the next Step.
func (i *Interpreter) CallStack() []vm.CallFrame {
	return i.vm.CallStack()
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
