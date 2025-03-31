package utils

import "sync"

// Set is a generic set, similar to the std::set in C++
type Set[T comparable] struct {
	data  map[T]struct{}
	mutex sync.Mutex
}

// NewSet creates a new generic Set
func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		data: make(map[T]struct{}),
	}
}

// Add adds an item to the set
func (s *Set[T]) Add(item T) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data[item] = struct{}{}
}

// Remove removes an item from the set
func (s *Set[T]) Remove(item T) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.data, item)
}

// Contains checks if an item exists in the set
func (s *Set[T]) Contains(item T) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, exists := s.data[item]
	return exists
}

// Size returns the number of items in the set
func (s *Set[T]) Size() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return len(s.data)
}

func (s *Set[T]) GetItems() []T {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	keys := make([]T, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

func (s *Set[T]) GetItemRefs() []*T {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	keys := make([]*T, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, &k)
	}
	return keys
}

func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	result := NewSet[T]()

	for _, item := range s.GetItems() {
		result.Add(item)
	}
	for _, item := range other.GetItems() {
		result.Add(item)
	}

	return result
}

func (s *Set[T]) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data = make(map[T]struct{})
}
