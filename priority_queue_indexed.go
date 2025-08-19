// Package threadsafe implements thread-safe operations.
package threadsafe

import "sync"

// IndexedPriorityQueue is a thread-safe binary min-heap implementation parameterized by a Less
// comparator. It optionally notifies a caller-supplied onSwap callback whenever two indices swap,
// which can be used to maintain external index fields.
//
// The zero value is not ready to use; construct via NewIndexedPriorityQueue. The less(a,b)
// comparator must define a strict weak ordering (irreflexive, transitive, consistent).
//
// Semantics mirror container/heap where applicable; indices are stable only for the lifetime
// between operations that may move elements. For external index maintenance (e.g., storing an
// "index" field inside elements), implementations may provide a swap-callback to notify callers
// when indices change. Note that index values refer to internal heap storage and are unstable
// across operations.
//
// Complexity: Push/Pop/Fix/RemoveAt O(log n), Peek O(1); Range does not mutate the heap.
type IndexedPriorityQueue[T any] struct {
	mu     sync.RWMutex
	items  []T
	cmp    func(a, b T) bool
	onSwap func(i, j int, items []T)
}

// Push inserts one or more items into the heap.
func (h *IndexedPriorityQueue[T]) Push(items ...T) {
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
func (h *IndexedPriorityQueue[T]) Pop() (item T, ok bool) {
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
func (h *IndexedPriorityQueue[T]) Peek() (item T, ok bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.items) == 0 {
		return item, false
	}
	return h.items[0], true
}

// Len returns number of items.
func (h *IndexedPriorityQueue[T]) Len() int {
	h.mu.RLock()
	l := len(h.items)
	h.mu.RUnlock()
	return l
}

// Clear removes all items.
func (h *IndexedPriorityQueue[T]) Clear() {
	h.mu.Lock()
	h.items = nil
	h.mu.Unlock()
}

// Range iterates over the current snapshot in arbitrary order. Mutations during range does not
// affect the current iteration.
func (h *IndexedPriorityQueue[T]) Range(f func(item T) bool) {
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
func (h *IndexedPriorityQueue[T]) Fix(i int) {
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
func (h *IndexedPriorityQueue[T]) RemoveAt(i int) (item T, ok bool) {
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
func (h *IndexedPriorityQueue[T]) UpdateAt(i int, x T) bool {
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

func (h *IndexedPriorityQueue[T]) lessIdx(i, j int) bool { return h.cmp(h.items[i], h.items[j]) }

func (h *IndexedPriorityQueue[T]) swap(i, j int) {
	if i == j {
		return
	}
	h.items[i], h.items[j] = h.items[j], h.items[i]
	if h.onSwap != nil {
		h.onSwap(i, j, h.items)
	}
}

func (h *IndexedPriorityQueue[T]) up(i int) {
	idx := i
	for {
		p := (idx - 1) / 2
		if idx == 0 || !h.lessIdx(idx, p) {
			break
		}
		h.swap(idx, p)
		idx = p
	}
}

// down moves item at i down; returns true if moved down.
func (h *IndexedPriorityQueue[T]) down(i int) bool {
	idx := i
	n := len(h.items)
	moved := false
	for {
		l := 2*idx + 1
		if l >= n {
			break
		}
		smallest := l
		r := l + 1
		if r < n && h.lessIdx(r, l) {
			smallest = r
		}
		if !h.lessIdx(smallest, idx) {
			break
		}
		h.swap(idx, smallest)
		idx = smallest
		moved = true
	}
	return moved
}

// NewIndexedPriorityQueue creates a new heap with the provided comparator.
// less(a,b) should return true when a has higher priority than b (i.e., a comes before b).
// onSwap is optional; if non-nil it's called under the write lock whenever two items swap indices
// and as such must not block or call back into the queue.
func NewIndexedPriorityQueue[T any](
	less func(a, b T) bool,
	onSwap func(i, j int, items []T),
) *IndexedPriorityQueue[T] {
	return &IndexedPriorityQueue[T]{cmp: less, onSwap: onSwap}
}
