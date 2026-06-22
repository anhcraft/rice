package mem

import (
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
	"github.com/anhcraft/rice/lib/stack"
)

type Environment struct {
	stackFrames       *stack.Stack[*StackFrame]
	namespace         *values.Namespace
	lexicalScopeDepth uint32
}

func NewEnvironment(preAllocatedFrames uint16) *Environment {
	return &Environment{
		stackFrames: stack.New[*StackFrame](int(preAllocatedFrames)),
		namespace:   values.NewNamespace(),
	}
}

// Namespace gets the Namespace
func (e *Environment) Namespace() *values.Namespace {
	return e.namespace
}

// LexicalScopeDepth current nesting depth of mem.LexicalScope
func (e *Environment) LexicalScopeDepth() uint32 {
	return e.lexicalScopeDepth
}

// PushFrame pushes a fresh frame to the call stack.
func (e *Environment) PushFrame(callSite values.CallSite) {
	frame := NewStackFrame(nil, callSite)
	frame.EnterScope()
	e.stackFrames.Push(frame)
	e.lexicalScopeDepth++
}

// PushFrameWithScope pushes a fresh frame with the given scope to the call stack.
func (e *Environment) PushFrameWithScope(scope *LexicalScope, callSite values.CallSite) {
	frame := NewStackFrame(scope, callSite)
	e.stackFrames.Push(frame)
	e.lexicalScopeDepth++
}

// PopFrame pops the frame from the call stack.
func (e *Environment) PopFrame() (*StackFrame, bool) {
	v, ok := e.stackFrames.Pop()
	if ok {
		e.lexicalScopeDepth--
		v.ExitScope()
	}
	return v, ok
}

// CurrentFrame retrieves the current frame in the call stack.
func (e *Environment) CurrentFrame() *StackFrame {
	frame, ok := e.stackFrames.Peek()
	if !ok {
		panic("no stack frame")
	}
	return frame
}

// EnterScope creates a new, nested lexical scope.
func (e *Environment) EnterScope() *LexicalScope {
	e.lexicalScopeDepth++
	return e.CurrentFrame().EnterScope()
}

// ExitScope leaves the current lexical scope.
func (e *Environment) ExitScope() {
	e.lexicalScopeDepth--
	e.CurrentFrame().ExitScope()
}

// IsNamespaceEntry checks whether the given identifier collides with an entry in the namespace tree
// (either a root-level native function or a sub-namespace).
func (e *Environment) IsNamespaceEntry(key values.Identifier) bool {
	v, err := e.namespace.Element(key)
	return v != nil && err == nil
}

// Define defines a variable; return false if it already exists
func (e *Environment) Define(key values.Identifier, value types.Value, constant bool) bool {
	return e.CurrentFrame().CurrentScope().Define(key, value, constant)
}

// Get retrieves a variable, searching from the current scope up to its parent/ancestor.
func (e *Environment) Get(key values.Identifier) (types.Value, bool) {
	return e.CurrentFrame().CurrentScope().Get(key)
}

// Assign updates an existing variable. It searches up the scope chain to find
// the variable and updates it in the scope where it was found. It returns false
// if the existing variable is constant (so reassignment is disallowed)
// If it doesn't exist, it defines it in the current (innermost) scope.
func (e *Environment) Assign(key values.Identifier, value types.Value) error {
	return e.CurrentFrame().CurrentScope().Assign(key, value)
}

func (e *Environment) Reset() {
	for !e.stackFrames.IsEmpty() {
		e.PopFrame()
	}
	e.lexicalScopeDepth = 0
}
