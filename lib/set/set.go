package set

import (
	"iter"
)

type Set[E comparable] struct {
	elements map[E]uint8
}

// NewSet creates and returns a new empty Set.
func NewSet[E comparable]() *Set[E] {
	return &Set[E]{
		elements: make(map[E]uint8),
	}
}

// Size returns the number of elements in the set.
func (s *Set[E]) Size() int {
	return len(s.elements)
}

// IsEmpty returns true if the set is empty, false otherwise.
func (s *Set[E]) IsEmpty() bool {
	return len(s.elements) == 0
}

// Add adds an element to the set.
// If the element already exists, the set remains unchanged.
func (s *Set[E]) Add(element E) {
	s.elements[element] = 1
}

// Has returns true if the element exists in the set, false otherwise.
func (s *Set[E]) Has(element E) bool {
	_, exists := s.elements[element]
	return exists
}

// Remove removes an element from the set.
// If the element does not exist, the set remains unchanged.
func (s *Set[E]) Remove(element E) {
	delete(s.elements, element)
}

// Slice returns a slice containing all elements in the set.
// The order of elements is not guaranteed.
func (s *Set[E]) Slice() []E {
	elements := make([]E, 0, len(s.elements))
	for element := range s.elements {
		elements = append(elements, element)
	}
	return elements
}

func (s *Set[E]) Iterate() iter.Seq[E] {
	return func(yield func(E) bool) {
		for v := range s.elements {
			if !yield(v) {
				return
			}
		}
	}
}
