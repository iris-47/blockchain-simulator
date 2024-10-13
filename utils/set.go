package utils

// Set is a generic set, similar to the std::set in C++
type Set[T comparable] struct {
	data map[T]struct{}
}

// NewSet creates a new generic Set
func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		data: make(map[T]struct{}),
	}
}

// Add adds an item to the set
func (s *Set[T]) Add(item T) {
	s.data[item] = struct{}{}
}

// Remove removes an item from the set
func (s *Set[T]) Remove(item T) {
	delete(s.data, item)
}

// Contains checks if an item exists in the set
func (s *Set[T]) Contains(item T) bool {
	_, exists := s.data[item]
	return exists
}

// Size returns the number of items in the set
func (s *Set[T]) Size() int {
	return len(s.data)
}
