// Package threadsafe implements thread-safe operations.
package threadsafe

import "sync"

// RWMutexSlice is a thread-safe buffer for any type T, backed by a slice and protected by a
// sync.RWMutex. It allows concurrent appends and atomic flushes.
type RWMutexSlice[T any] struct {
	mu   sync.RWMutex
	data []T
}

// NewRWMutexSlice creates a new RWMutexSlice with an optional initial capacity.
func NewRWMutexSlice[T any](initialCap int) *RWMutexSlice[T] {
	return &RWMutexSlice[T]{
		data: make([]T, 0, initialCap),
	}
}

// Append appends items to the buffer in a thread-safe way.
func (b *RWMutexSlice[T]) Append(item ...T) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.data = append(b.data, item...)
}

// Flush atomically retrieves all items and clears the buffer.
// Returns a slice with the previous contents.
func (b *RWMutexSlice[T]) Flush() []T {
	b.mu.Lock()
	defer b.mu.Unlock()

	flushed := b.data
	b.data = make([]T, 0, cap(flushed))
	return flushed
}

// Peek returns a copy of the current buffer contents without clearing.
// The returned slice is safe to read but may be stale if new items are added concurrently.
func (b *RWMutexSlice[T]) Peek() []T {
	b.mu.RLock()
	defer b.mu.RUnlock()

	copied := make([]T, len(b.data))
	copy(copied, b.data)
	return copied
}

// Len returns the current number of items in the buffer.
func (b *RWMutexSlice[T]) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.data)
}
