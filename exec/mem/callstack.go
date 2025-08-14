package mem

import (
	"rice/exec/types/values"
)

type StackFrame struct {
	activeScope *LexicalScope
	callSite    values.CallSite
}

func NewStackFrame(scope *LexicalScope, callSite values.CallSite) *StackFrame {
	return &StackFrame{activeScope: scope, callSite: callSite}
}

func (sf *StackFrame) EnterScope() *LexicalScope {
	scope := NewLexicalScope(sf.activeScope)
	sf.activeScope = scope
	return scope
}

func (sf *StackFrame) CurrentScope() *LexicalScope {
	return sf.activeScope
}

func (sf *StackFrame) ExitScope() {
	if sf.activeScope == nil {
		panic("no scope left to exit")
	}
	curr := sf.activeScope
	sf.activeScope = curr.parent
	releaseLexicalScope(curr)
}

func (sf *StackFrame) CallSite() values.CallSite {
	return sf.callSite
}
