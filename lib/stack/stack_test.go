package stack

import (
	"testing"
)

func TestStack_PushPopPeek(t *testing.T) {
	s := Stack[int]{}

	if size := s.Size(); size != 0 {
		t.Errorf("expected size 0, got %d", size)
	}

	s.Push(1)
	s.Push(2)
	s.Push(3)

	if size := s.Size(); size != 3 {
		t.Errorf("expected size 3, got %d", size)
	}

	// Peek
	if val, ok := s.Peek(); !ok || val != 3 {
		t.Errorf("expected peek 3, got %v (ok: %v)", val, ok)
	}

	// Pop 3
	if val, ok := s.Pop(); !ok || val != 3 {
		t.Errorf("expected pop 3, got %v (ok: %v)", val, ok)
	}

	// Pop 2
	if val, ok := s.Pop(); !ok || val != 2 {
		t.Errorf("expected pop 2, got %v (ok: %v)", val, ok)
	}

	// Pop 1
	if val, ok := s.Pop(); !ok || val != 1 {
		t.Errorf("expected pop 1, got %v (ok: %v)", val, ok)
	}

	// Now empty
	if _, ok := s.Pop(); ok {
		t.Errorf("expected pop to fail on empty stack")
	}

	if _, ok := s.Peek(); ok {
		t.Errorf("expected peek to fail on empty stack")
	}

	if size := s.Size(); size != 0 {
		t.Errorf("expected size 0 after popping all, got %d", size)
	}
}
