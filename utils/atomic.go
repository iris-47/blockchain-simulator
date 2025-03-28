// similar to std::atomic in C++, AtomicValue is a thread-safe value holder that can be updated atomically.
// It provides a way to update the value with a condition, and a way to update the value with a function.
// The value can be read with Get() method.
// The value can be set with Set() method, which takes a new value and a condition function.
// The value can be updated with Update() method, which takes an updater function.
package utils

import "sync"

type AtomicValue[T any] struct {
	mu    sync.RWMutex
	value T

	// Optional, a callback function that will be called after the value is updated
	onUpdate func(oldVal, newVal T) (updated bool)
}

func NewAtomicValue[T any](initial T) *AtomicValue[T] {
	return &AtomicValue[T]{value: initial}
}

// Get returns the current value of the atomic value.
func (s *AtomicValue[T]) Get() T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

// Set sets the value of the atomic value to newVal if the condition is met.
// If the condition is not met, the value is not updated.
// If the condition is nil, the value is always updated.
// The callback function onUpdate is called after the value is updated.
func (s *AtomicValue[T]) Set(newVal T, cond func(current T) bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cond == nil || cond(s.value) {
		if s.onUpdate != nil {
			if !s.onUpdate(s.value, newVal) {
				return false
			}
		}
		s.value = newVal
		return true
	}
	return false
}

// Update updates the value of the atomic value using the updater function.
// The callback function onUpdate is called after the value is updated.
func (s *AtomicValue[T]) Update(updater func(current T) T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	old := s.value
	s.value = updater(old)
	if s.onUpdate != nil {
		s.onUpdate(old, s.value)
	}
}

// SetUpdateFunc sets the callback function that will be called after the value is updated.
func (s *AtomicValue[T]) SetUpdateFunc(fn func(old, new T) bool) *AtomicValue[T] {
	s.onUpdate = fn
	return s
}
