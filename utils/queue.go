package utils

import (
	"fmt"
	"sync"
)

// Queue is similar to the std::queue in C++
type Queue[T any] struct {
	items []T
	mutex sync.Mutex
}

func NewQueue[T any]() *Queue[T] {
	q := &Queue[T]{
		items: make([]T, 0),
	}
	return q
}

// Add an element to the end of the queue
func (q *Queue[T]) Enqueue(item T) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.items = append(q.items, item)
}

// Get the first element in the queue, and remove it from the queue, block the goroutine when the queue is empty
func (q *Queue[T]) Dequeue() (T, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	for len(q.items) == 0 {
		var zero T
		return zero, fmt.Errorf("Queue is empty")
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item, nil
}

func (q *Queue[T]) IsEmpty() bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.items) == 0
}

func (q *Queue[T]) Size() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.items)
}
