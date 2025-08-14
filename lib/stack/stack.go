package stack

type Stack[T any] struct {
	data []T
}

// New creates a new stack with the given initial capacity.
func New[T any](initialCapacity int) *Stack[T] {
	if initialCapacity <= 0 {
		initialCapacity = 16
	}
	return &Stack[T]{
		data: make([]T, 0, initialCapacity),
	}
}

// Push adds an element to the top of the stack.
func (s *Stack[T]) Push(val T) {
	s.data = append(s.data, val)
}

// Pop removes and returns the top element of the stack.
// Returns false if the stack is empty.
func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.data) == 0 {
		return zero, false
	}
	top := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	return top, true
}

// Peek returns the top element without removing it.
// Returns false if the stack is empty.
func (s *Stack[T]) Peek() (T, bool) {
	var zero T
	if len(s.data) == 0 {
		return zero, false
	}
	return s.data[len(s.data)-1], true
}

// Size returns the number of elements in the stack.
func (s *Stack[T]) Size() int {
	return len(s.data)
}

// IsEmpty checks if the stack is empty
func (s *Stack[T]) IsEmpty() bool {
	return len(s.data) == 0
}

// PopAll pops all the elements
func (s *Stack[T]) PopAll() []T {
	result := make([]T, 0, s.Size())
	for !s.IsEmpty() {
		val, _ := s.Pop()
		result = append(result, val)
	}
	return result
}

// Clear removes all elements from the stack, preserving the capacity
func (s *Stack[T]) Clear() {
	s.data = s.data[:0]
}
