// Package threadsafe implements thread-safe operations.
package threadsafe

import (
	"sync"
)

// Slice is a generic interface for stores with any type T.
type Slice[T any] interface {
	// Append appends an item to the buffer in a thread-safe way.
	Append(item ...T)
	// Flush atomically retrieves all items and clears the buffer.
	// Returns a slice with the previous contents.
	Flush() []T
	// Peek returns a copy of the current buffer contents without clearing.
	// The returned slice is safe to read but may be stale if new items are added concurrently.
	Peek() []T
	// Len returns the current number of items in the buffer.
	Len() int
}

// MutexSlice is a thread-safe buffer for any type T.
// It allows concurrent appends and atomic flushes.
type MutexSlice[T any] struct {
	mu   sync.Mutex
	data []T
}

// NewMutexSlice creates a new MutexSlice with an optional initial capacity.
func NewMutexSlice[T any](initialCap int) *MutexSlice[T] {
	return &MutexSlice[T]{
		data: make([]T, 0, initialCap),
	}
}

// Append appends items to the buffer in a thread-safe way.
func (b *MutexSlice[T]) Append(item ...T) {
	b.mu.Lock()
	b.data = append(b.data, item...)
	b.mu.Unlock()
}

// Flush atomically retrieves all items and clears the buffer.
// Returns a slice with the previous contents.
func (b *MutexSlice[T]) Flush() []T {
	b.mu.Lock()
	defer b.mu.Unlock()
	flushed := b.data
	b.data = make([]T, 0, cap(flushed))
	return flushed
}

// Peek returns a copy of the current buffer contents without clearing.
// The returned slice is safe to read but may be stale if new items are added concurrently.
func (b *MutexSlice[T]) Peek() []T {
	b.mu.Lock()
	defer b.mu.Unlock()
	copied := make([]T, len(b.data))
	copy(copied, b.data)
	return copied
}

// Len returns the current number of items in the buffer.
func (b *MutexSlice[T]) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.data)
}
