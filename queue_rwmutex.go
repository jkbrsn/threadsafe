// Package threadsafe implements thread-safe operations.
package threadsafe

import "sync"

// RWMutexQueue is a thread-safe FIFO queue implementation backed by a slice and protected
// by a sync.RWMutex.
//
// The implementation aims for amortized O(1) Push and Pop by keeping a head index instead
// of shifting the slice on every Pop. When the internal slice has too much unused prefix,
// it is resliced to reclaim memory.
//
// The zero value of RWMutexQueue is ready to use.
const shrinkThreshold = 64 // when head exceeds this and half the slice is unused, shrink

type RWMutexQueue[T any] struct {
	mu    sync.RWMutex
	items []T
	head  int // index of the current front element in items slice
}

// NewRWMutexQueue creates a new instance of RWMutexQueue.
func NewRWMutexQueue[T any]() *RWMutexQueue[T] {
	return &RWMutexQueue[T]{}
}

// Push adds one or more items to the back of the queue.
func (q *RWMutexQueue[T]) Push(items ...T) {
	if len(items) == 0 {
		return
	}
	q.mu.Lock()
	q.items = append(q.items, items...)
	q.mu.Unlock()
}

// Pop removes and returns the item at the front of the queue.
// If the queue is empty it returns ok == false and the zero value of T.
func (q *RWMutexQueue[T]) Pop() (item T, ok bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.head >= len(q.items) {
		return item, false
	}

	item = q.items[q.head]
	ok = true
	q.head++

	// Periodically reclaim memory when head grows large.
	if q.head > shrinkThreshold && q.head*2 >= len(q.items) {
		// copy the active items to a new slice and reset head.
		newItems := make([]T, len(q.items)-q.head)
		copy(newItems, q.items[q.head:])
		q.items = newItems
		q.head = 0
	}

	return item, ok
}

// Peek returns the item at the front without removing it.
func (q *RWMutexQueue[T]) Peek() (item T, ok bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.head >= len(q.items) {
		return item, false
	}
	return q.items[q.head], true
}

// Len returns the current number of items.
func (q *RWMutexQueue[T]) Len() int {
	q.mu.RLock()
	l := len(q.items) - q.head
	q.mu.RUnlock()
	return l
}

// Clear removes all items from the queue.
func (q *RWMutexQueue[T]) Clear() {
	q.mu.Lock()
	q.items = nil
	q.head = 0
	q.mu.Unlock()
}

// Slice returns a copy of the queue contents from front to back.
func (q *RWMutexQueue[T]) Slice() []T {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.head >= len(q.items) {
		return nil
	}
	result := make([]T, len(q.items)-q.head)
	copy(result, q.items[q.head:])
	return result
}

// Range calls f sequentially for each item from front to back. This action does not modify
// the queue or its items.
func (q *RWMutexQueue[T]) Range(f func(item T) bool) {
	q.mu.RLock()
	items := q.items[q.head:]
	q.mu.RUnlock()

	for _, it := range items {
		if !f(it) {
			break
		}
	}
}

// Ensure RWMutexQueue implements Queue.
var _ Queue[any] = (*RWMutexQueue[any])(nil)
