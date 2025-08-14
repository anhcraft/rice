package queue

type Queue[T any] struct {
	items    []T
	head     int
	tail     int
	size     int
	capacity int
}

// New creates a new queue with the given initial capacity
func New[T any](initialCapacity int) *Queue[T] {
	if initialCapacity <= 0 {
		initialCapacity = 16
	}
	return &Queue[T]{
		items:    make([]T, initialCapacity),
		capacity: initialCapacity,
	}
}

// Enqueue adds an item to the end of the queue
func (q *Queue[T]) Enqueue(item T) {
	if q.size == q.capacity {
		newCapacity := q.capacity * 2
		newItems := make([]T, newCapacity)

		for i := 0; i < q.size; i++ {
			newItems[i] = q.items[(q.head+i)%q.capacity]
		}

		q.items = newItems
		q.head = 0
		q.tail = q.size
		q.capacity = newCapacity
	}

	q.items[q.tail] = item
	q.tail = (q.tail + 1) % q.capacity
	q.size++
}

// Dequeue removes and returns the first item from the queue
// Returns the zero value and false if the queue is empty
func (q *Queue[T]) Dequeue() (T, bool) {
	if q.size == 0 {
		var zero T
		return zero, false
	}

	item := q.items[q.head]
	var zero T
	q.items[q.head] = zero
	q.head = (q.head + 1) % q.capacity
	q.size--
	return item, true
}

// Size returns the number of items in the queue
func (q *Queue[T]) Size() int {
	return q.size
}

// IsEmpty checks if the queue is empty
func (q *Queue[T]) IsEmpty() bool {
	return q.size == 0
}
