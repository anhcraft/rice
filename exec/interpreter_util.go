package exec

import (
	"fmt"
	"github.com/anhcraft/rice/exec/ast"
	"github.com/anhcraft/rice/exec/profiler"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
)

func (i *Interpreter) Profiler() profiler.Profiler {
	return i.profiler
}

func (i *Interpreter) cleanUp() {
	i.profiler.Reset()
	i.env.Reset()
	i.functionDepth = 0
	i.loopDepth = 0
	i.dirty = false
}

// throw returns a formatted RuntimeError caused at the given statement
func (i *Interpreter) throw(stmt ast.Stmt, msg string, args ...any) RuntimeError {
	return RuntimeError{
		message: fmt.Sprintf(msg, args...),
		source:  i.env.CurrentFrame().CallSite(),
		start:   stmt.StartPos(),
		end:     stmt.EndPos(),
	}
}

// throwCall returns a formatted RuntimeError caused at the given values.CallSite
func (i *Interpreter) throwCall(site values.CallSite, msg string, args ...any) RuntimeError {
	return RuntimeError{
		message: fmt.Sprintf(msg, args...),
		source:  i.env.CurrentFrame().CallSite(), // this could be different from site
		start:   site.StartPos,
		end:     site.EndPos,
	}
}

// astNilCheck requires an AST node to be non-nil
// When panics, recheck the parser to complete the validation
func astNilCheck(v ast.Node) {
	if v == nil {
		panic("nil AST: incomplete validation from parser")
	}
}

type EvalFlag uint8

const (
	ExceptId  EvalFlag = 1
	ExceptSel EvalFlag = 2
)

// eval evaluates an statement or value
func (i *Interpreter) eval(arg any) (types.Value, error) {
	return i.evalc(arg, 0)
}

// evalc evaluates an statement or value conditionally
func (i *Interpreter) evalc(arg any, flags EvalFlag) (types.Value, error) {
	if arg == nil {
		return nil, nil
	}

	if id, ok := arg.(ast.Stmt); ok {
		val, err := id.Accept(i)
		if err != nil {
			return nil, err
		}
		arg = val
	}

	if arg == nil {
		return nil, nil
	}

	if flags&ExceptSel == 0 {
		if sel, ok := arg.(values.Selector); ok {
			val, err := sel.Get()
			if err != nil {
				return nil, err
			}
			arg = val
		}
	}

	if arg == nil {
		return nil, nil
	}

	if flags&ExceptId == 0 {
		if id, ok := arg.(values.Identifier); ok {
			if ns, err := i.env.Namespace().Element(id); ns != nil && err == nil {
				return ns, nil
			}

			if val, ok := i.env.Get(id); ok {
				return val, nil
			}

			return nil, fmt.Errorf("unresolved reference %q", id)
		}
	}

	if _, ok := arg.(types.Value); ok {
		return arg.(types.Value), nil
	}

	panic(fmt.Errorf("unsupported evaluation of %T", arg))
}

// checkScopeThrottle checks if the lexical scope exceeds the defined limit
func (i *Interpreter) checkScopeThrottle() error {
	if i.env.LexicalScopeDepth() > i.lexicalScopeLimit {
		return fmt.Errorf("reached lexical scope limit of %d", i.lexicalScopeLimit)
	}
	return nil
}

// checkContext checks if the context requires termination
func (i *Interpreter) checkContext() error {
	select {
	case <-i.ctx.Done():
		return i.ctx.Err()
	default:
		return nil
	}
}
