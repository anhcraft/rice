package queue

import (
	"testing"
)

func TestQueue(t *testing.T) {
	t.Run("New queue is empty", func(t *testing.T) {
		q := New[int](4)
		if !q.IsEmpty() {
			t.Error("New queue should be empty")
		}
		if q.Size() != 0 {
			t.Errorf("New queue size should be 0, got %d", q.Size())
		}
	})

	t.Run("Enqueue and Dequeue single item", func(t *testing.T) {
		q := New[string](4)
		q.Enqueue("test")

		if q.Size() != 1 {
			t.Errorf("Expected size 1, got %d", q.Size())
		}

		item, ok := q.Dequeue()
		if !ok {
			t.Error("Dequeue should return true for non-empty queue")
		}
		if item != "test" {
			t.Errorf("Expected item 'test', got '%s'", item)
		}
		if !q.IsEmpty() {
			t.Error("Queue should be empty after dequeue")
		}
	})

	t.Run("Dequeue empty queue", func(t *testing.T) {
		q := New[int](4)
		_, ok := q.Dequeue()
		if ok {
			t.Error("Dequeue on empty queue should return false")
		}
	})

	t.Run("Auto-resize on enqueue", func(t *testing.T) {
		q := New[int](2)
		q.Enqueue(1)
		q.Enqueue(2)
		q.Enqueue(3)

		if q.capacity != 4 {
			t.Errorf("Expected capacity 4 after resize, got %d", q.capacity)
		}
		if q.Size() != 3 {
			t.Errorf("Expected size 3, got %d", q.Size())
		}

		for i := 1; i <= 3; i++ {
			item, ok := q.Dequeue()
			if !ok || item != i {
				t.Errorf("Expected item %d, got %d (ok=%v)", i, item, ok)
			}
		}
	})

	t.Run("Circular buffer behavior", func(t *testing.T) {
		q := New[int](3)
		q.Enqueue(1)
		q.Enqueue(2)
		q.Dequeue()
		q.Enqueue(3)
		q.Enqueue(4)

		expected := []int{2, 3, 4}
		for i, expectedVal := range expected {
			item, ok := q.Dequeue()
			if !ok || item != expectedVal {
				t.Errorf("At index %d: expected %d, got %d (ok=%v)", i, expectedVal, item, ok)
			}
		}
	})

	t.Run("Zero value initialization", func(t *testing.T) {
		q := New[int](0)
		if q.capacity != 16 {
			t.Errorf("Expected default capacity 16, got %d", q.capacity)
		}
	})

	t.Run("Different types", func(t *testing.T) {
		type TestStruct struct {
			value int
		}

		q := New[TestStruct](2)
		q.Enqueue(TestStruct{value: 42})
		item, ok := q.Dequeue()
		if !ok || item.value != 42 {
			t.Errorf("Expected struct with value 42, got %v (ok=%v)", item, ok)
		}
	})

	t.Run("Large number of operations", func(t *testing.T) {
		q := New[int](4)
		for i := 0; i < 100; i++ {
			q.Enqueue(i)
		}
		for i := 0; i < 100; i++ {
			item, ok := q.Dequeue()
			if !ok || item != i {
				t.Errorf("At index %d: expected %d, got %d (ok=%v)", i, i, item, ok)
			}
		}
	})
}
