// Package threadsafe implements thread-safe operations.
package threadsafe

import "sync"

// RWMutexPriorityQueue is a thread-safe binary min-heap implementation parameterized by a Less comparator.
// It maintains O(log n) push/pop/fix/removeAt and O(1) peek. It optionally notifies a caller-supplied
// onSwap callback whenever two indices swap, which can be used to maintain external index fields.
//
// The zero value is not ready to use; construct via NewRWMutexPriorityQueue.
// This mirrors the style of NewRWMutexQueue in this repository.

type RWMutexPriorityQueue[T any] struct {
	mu     sync.RWMutex
	items  []T
	cmp    func(a, b T) bool
	onSwap func(i, j int, items []T)
}

// NewRWMutexPriorityQueue creates a new heap with the provided comparator.
// less(a,b) should return true when a has higher priority than b (i.e., a comes before b).
// onSwap is optional; if non-nil it is called under the write lock whenever two items swap indices.
func NewRWMutexPriorityQueue[T any](less func(a, b T) bool, onSwap func(i, j int, items []T)) *RWMutexPriorityQueue[T] {
	return &RWMutexPriorityQueue[T]{cmp: less, onSwap: onSwap}
}

// Push inserts one or more items into the heap.
func (h *RWMutexPriorityQueue[T]) Push(items ...T) {
	if len(items) == 0 {
		return
	}
	h.mu.Lock()
	for _, x := range items {
		h.items = append(h.items, x)
		h.up(len(h.items) - 1)
	}
	h.mu.Unlock()
}

// Pop removes and returns the minimum item.
func (h *RWMutexPriorityQueue[T]) Pop() (item T, ok bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.items) == 0 {
		return item, false
	}
	last := len(h.items) - 1
	h.swap(0, last)
	item = h.items[last]
	h.items = h.items[:last]
	if len(h.items) > 0 {
		h.down(0)
	}
	return item, true
}

// Peek returns the minimum item without removing it.
func (h *RWMutexPriorityQueue[T]) Peek() (item T, ok bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.items) == 0 {
		return item, false
	}
	return h.items[0], true
}

// Len returns number of items.
func (h *RWMutexPriorityQueue[T]) Len() int {
	h.mu.RLock()
	l := len(h.items)
	h.mu.RUnlock()
	return l
}

// Clear removes all items.
func (h *RWMutexPriorityQueue[T]) Clear() {
	h.mu.Lock()
	h.items = nil
	h.mu.Unlock()
}

// Slice returns a copy of the internal array (arbitrary order).
func (h *RWMutexPriorityQueue[T]) Slice() []T {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.items) == 0 {
		return nil
	}
	cp := make([]T, len(h.items))
	copy(cp, h.items)
	return cp
}

// Range iterates over the current snapshot in arbitrary order.
func (h *RWMutexPriorityQueue[T]) Range(f func(item T) bool) {
	h.mu.RLock()
	snap := make([]T, len(h.items))
	copy(snap, h.items)
	h.mu.RUnlock()
	for _, it := range snap {
		if !f(it) {
			break
		}
	}
}

// Fix restores heap order after the item at index i may have changed.
func (h *RWMutexPriorityQueue[T]) Fix(i int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if i < 0 || i >= len(h.items) {
		return
	}
	if !h.down(i) {
		h.up(i)
	}
}

// RemoveAt removes and returns the item at index i, if valid.
func (h *RWMutexPriorityQueue[T]) RemoveAt(i int) (item T, ok bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if i < 0 || i >= len(h.items) {
		return item, false
	}
	last := len(h.items) - 1
	if i != last {
		h.swap(i, last)
	}
	item = h.items[last]
	h.items = h.items[:last]
	if i < len(h.items) {
		if !h.down(i) {
			h.up(i)
		}
	}
	return item, true
}

// UpdateAt replaces the element at index i and restores invariants.
func (h *RWMutexPriorityQueue[T]) UpdateAt(i int, x T) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if i < 0 || i >= len(h.items) {
		return false
	}
	h.items[i] = x
	if !h.down(i) {
		h.up(i)
	}
	return true
}

// Internal helpers (callers must hold write lock)

func (h *RWMutexPriorityQueue[T]) lessIdx(i, j int) bool { return h.cmp(h.items[i], h.items[j]) }

func (h *RWMutexPriorityQueue[T]) swap(i, j int) {
	if i == j {
		return
	}
	h.items[i], h.items[j] = h.items[j], h.items[i]
	if h.onSwap != nil {
		h.onSwap(i, j, h.items)
	}
}

func (h *RWMutexPriorityQueue[T]) up(i int) {
	for {
		p := (i - 1) / 2
		if i == 0 || !h.lessIdx(i, p) {
			break
		}
		h.swap(i, p)
		i = p
	}
}

// down moves item at i down; returns true if moved down.
func (h *RWMutexPriorityQueue[T]) down(i int) bool {
	n := len(h.items)
	moved := false
	for {
		l := 2*i + 1
		if l >= n {
			break
		}
		smallest := l
		r := l + 1
		if r < n && h.lessIdx(r, l) {
			smallest = r
		}
		if !h.lessIdx(smallest, i) {
			break
		}
		h.swap(i, smallest)
		i = smallest
		moved = true
	}
	return moved
}

// Ensure RWMutexPriorityQueue implements PriorityQueue.
var _ PriorityQueue[any] = (*RWMutexPriorityQueue[any])(nil)
