package set

import (
	"testing"
)

func TestSet(t *testing.T) {
	t.Run("IntSet", func(t *testing.T) {
		set := NewSet[int]()

		if set.Size() != 0 {
			t.Errorf("Expected size to be 0, got %d", set.Size())
		}
		if !set.IsEmpty() {
			t.Errorf("Expected set to be empty")
		}

		set.Add(1)
		set.Add(2)
		set.Add(3)

		if set.Size() != 3 {
			t.Errorf("Expected size to be 3, got %d", set.Size())
		}
		if set.IsEmpty() {
			t.Errorf("Expected set not to be empty")
		}

		if !set.Has(1) {
			t.Errorf("Expected set to have element 1")
		}
		if !set.Has(2) {
			t.Errorf("Expected set to have element 2")
		}
		if !set.Has(3) {
			t.Errorf("Expected set to have element 3")
		}
		if set.Has(4) {
			t.Errorf("Expected set not to have element 4")
		}

		set.Add(2) // Adding existing element
		if set.Size() != 3 {
			t.Errorf("Expected size to remain 3 after adding duplicate, got %d", set.Size())
		}

		set.Remove(2)
		if set.Size() != 2 {
			t.Errorf("Expected size to be 2 after removing 2, got %d", set.Size())
		}
		if set.Has(2) {
			t.Errorf("Expected set not to have element 2 after removal")
		}

		set.Remove(4) // Removing non-existing element
		if set.Size() != 2 {
			t.Errorf("Expected size to remain 2 after removing non-existing 4, got %d", set.Size())
		}

		found := make(map[int]bool)
		for e := range set.Iterate() {
			found[e] = true
		}
		if !found[1] {
			t.Errorf("Expected element 1 to be found in iteration")
		}
		if !found[3] {
			t.Errorf("Expected element 3 to be found in iteration")
		}
	})

	t.Run("StringSet", func(t *testing.T) {
		set := NewSet[string]()

		set.Add("apple")
		set.Add("banana")
		set.Add("cherry")

		if set.Size() != 3 {
			t.Errorf("Expected size to be 3, got %d", set.Size())
		}

		if !set.Has("apple") {
			t.Errorf("Expected set to have element 'apple'")
		}
		if set.Has("grape") {
			t.Errorf("Expected set not to have element 'grape'")
		}

		set.Remove("banana")
		if set.Size() != 2 {
			t.Errorf("Expected size to be 2 after removing 'banana', got %d", set.Size())
		}
		if set.Has("banana") {
			t.Errorf("Expected set not to have element 'banana' after removal")
		}

		found := make(map[string]bool)
		for e := range set.Iterate() {
			found[e] = true
		}
		if !found["apple"] {
			t.Errorf("Expected element 'apple' to be found in iteration")
		}
		if !found["cherry"] {
			t.Errorf("Expected element 'cherry' to be found in iteration")
		}
	})
}
