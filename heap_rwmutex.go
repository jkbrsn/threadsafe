// Package threadsafe implements thread-safe operations.
package threadsafe

import (
	"iter"
	"slices"
	"sync"
)

// RWMutexHeap is a thread-safe binary heap implementation protected by a sync.RWMutex.
// The ordering is determined by the provided less function: less(a, b) == true means
// a has higher priority than b (i.e., a comes out before b). This makes it a min-heap
// when less(a, b) is a < b, and a max-heap when less(a, b) is a > b.
//
// The zero value is not ready to use; use NewRWMutexHeap to construct with a comparator.
type RWMutexHeap[T any] struct {
	mu   sync.RWMutex
	data []T
	less func(a, b T) bool
}

// NewRWMutexHeap creates a new RWMutexHeap with the provided less function.
func NewRWMutexHeap[T any](less func(a, b T) bool) *RWMutexHeap[T] {
	return &RWMutexHeap[T]{
		data: make([]T, 0),
		less: less,
	}
}

// Push adds one or more items to the heap.
func (h *RWMutexHeap[T]) Push(items ...T) {
	if len(items) == 0 {
		return
	}
	h.mu.Lock()
	for _, x := range items {
		h.data = append(h.data, x)
		h.up(len(h.data) - 1)
	}
	h.mu.Unlock()
}

// Pop removes and returns the top-priority item.
// If the heap is empty it returns ok == false and the zero value of T.
func (h *RWMutexHeap[T]) Pop() (item T, ok bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	n := len(h.data)
	if n == 0 {
		return item, false
	}
	// Swap first and last, pop last, then down from root.
	item = h.data[0]
	last := h.data[n-1]
	h.data = h.data[:n-1]
	if n-1 > 0 {
		h.data[0] = last
		h.down(0)
	}
	return item, true
}

// Peek returns the top-priority item without removing it.
func (h *RWMutexHeap[T]) Peek() (item T, ok bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.data) == 0 {
		return item, false
	}
	return h.data[0], true
}

// Len returns the current number of items.
func (h *RWMutexHeap[T]) Len() int {
	h.mu.RLock()
	l := len(h.data)
	h.mu.RUnlock()
	return l
}

// Clear removes all items from the heap.
func (h *RWMutexHeap[T]) Clear() {
	h.mu.Lock()
	h.data = nil
	h.mu.Unlock()
}

// Slice returns a copy of the heap contents in internal heap order.
func (h *RWMutexHeap[T]) Slice() []T {
	return slices.Collect(h.All())
}

// Range calls f sequentially for each item in internal heap order. This action does not modify
// the heap or its items.
func (h *RWMutexHeap[T]) Range(f func(item T) bool) {
	h.mu.RLock()
	items := make([]T, len(h.data))
	copy(items, h.data)
	h.mu.RUnlock()
	for _, it := range items {
		if !f(it) {
			break
		}
	}
}

// All returns an iterator over items in the heap in internal heap order (not sorted).
// The iteration order is implementation-defined and not guaranteed to be priority-sorted.
func (h *RWMutexHeap[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		h.mu.RLock()
		snapshot := make([]T, len(h.data))
		copy(snapshot, h.data)
		h.mu.RUnlock()

		for _, item := range snapshot {
			if !yield(item) {
				return
			}
		}
	}
}

// up restores the heap property by sifting up the element at index i.
func (h *RWMutexHeap[T]) up(i int) {
	idx := i
	for idx > 0 {
		p := (idx - 1) / 2
		if !h.less(h.data[idx], h.data[p]) {
			break
		}
		h.data[idx], h.data[p] = h.data[p], h.data[idx]
		idx = p
	}
}

// down restores the heap property by sifting down the element at index i.
func (h *RWMutexHeap[T]) down(i int) {
	idx := i
	n := len(h.data)
	for {
		left := 2*idx + 1
		if left >= n {
			break
		}
		smallest := left
		right := left + 1
		if right < n && h.less(h.data[right], h.data[left]) {
			smallest = right
		}
		if !h.less(h.data[smallest], h.data[idx]) {
			break
		}
		h.data[idx], h.data[smallest] = h.data[smallest], h.data[idx]
		idx = smallest
	}
}

// Ensure RWMutexHeap implements Heap.
var _ Heap[any] = (*RWMutexHeap[any])(nil)
