package mem

import (
	"fmt"
	"rice/exec/types"
	"rice/exec/types/values"
	"sync"
)

var scopePool = sync.Pool{
	New: func() interface{} {
		return &LexicalScope{
			store: make(map[values.Identifier]Variable),
		}
	},
}

type LexicalScope struct {
	parent *LexicalScope
	store  map[values.Identifier]Variable
}

func NewLexicalScope(parent *LexicalScope) *LexicalScope {
	scope := scopePool.Get().(*LexicalScope)
	scope.parent = parent
	// scope store gets cleared on release
	return scope
}

func releaseLexicalScope(s *LexicalScope) {
	for k := range s.store {
		delete(s.store, k)
	}
	s.parent = nil
	scopePool.Put(s)
}

// Define defines a variable; return false if it already exists
func (s *LexicalScope) Define(key values.Identifier, value types.Value, constant bool) bool {
	if _, exists := s.store[key]; exists {
		return false
	}
	s.store[key] = Variable{value, constant}
	return true
}

// Get retrieves a variable, searching from the current scope up to its parent/ancestor.
func (s *LexicalScope) Get(key values.Identifier) (types.Value, bool) {
	current := s
	for current != nil {
		if vr, ok := current.store[key]; ok {
			return vr.value, true
		}
		current = current.parent
	}

	return nil, false
}

// Assign updates an existing variable. It searches up the scope chain to find
// the variable and updates it in the scope where it was found. It returns false
// if the existing variable is constant (so reassignment is disallowed)
// If it doesn't exist, it defines it in the current (innermost) scope.
func (s *LexicalScope) Assign(key values.Identifier, value types.Value) error {
	current := s
	for current != nil {
		if vr, ok := current.store[key]; ok {
			if vr.constant {
				return fmt.Errorf("cannot assign to constant %q", key)
			}
			current.store[key] = Variable{value, false}
			return nil
		}
		current = current.parent
	}
	return fmt.Errorf("unknown variable %q", key)
}
