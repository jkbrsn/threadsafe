// Package threadsafe implements thread-safe operations.
package threadsafe

import (
	"iter"
	"sync"
)

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
func (q *IndexedPriorityQueue[T]) Push(items ...T) {
	if len(items) == 0 {
		return
	}
	q.mu.Lock()
	for _, x := range items {
		q.items = append(q.items, x)
		q.up(len(q.items) - 1)
	}
	q.mu.Unlock()
}

// Pop removes and returns the minimum item.
func (q *IndexedPriorityQueue[T]) Pop() (item T, ok bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		return item, false
	}
	last := len(q.items) - 1
	q.swap(0, last)
	item = q.items[last]
	q.items = q.items[:last]
	if len(q.items) > 0 {
		q.down(0)
	}
	return item, true
}

// Peek returns the minimum item without removing it.
func (q *IndexedPriorityQueue[T]) Peek() (item T, ok bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if len(q.items) == 0 {
		return item, false
	}
	return q.items[0], true
}

// Len returns number of items.
func (q *IndexedPriorityQueue[T]) Len() int {
	q.mu.RLock()
	l := len(q.items)
	q.mu.RUnlock()
	return l
}

// Clear removes all items.
func (q *IndexedPriorityQueue[T]) Clear() {
	q.mu.Lock()
	q.items = nil
	q.mu.Unlock()
}

// Range iterates over the current snapshot in arbitrary order. Mutations during range does not
// affect the current iteration.
func (q *IndexedPriorityQueue[T]) Range(f func(item T) bool) {
	q.mu.RLock()
	snap := make([]T, len(q.items))
	copy(snap, q.items)
	q.mu.RUnlock()
	for _, it := range snap {
		if !f(it) {
			break
		}
	}
}

// All returns an iterator over items in the queue in internal heap order (not sorted).
// The iteration order is implementation-defined and not guaranteed to be priority-sorted.
func (q *IndexedPriorityQueue[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		q.mu.RLock()
		snapshot := make([]T, len(q.items))
		copy(snapshot, q.items)
		q.mu.RUnlock()

		for _, item := range snapshot {
			if !yield(item) {
				return
			}
		}
	}
}

// Fix restores heap order after the item at index i may have changed.
func (q *IndexedPriorityQueue[T]) Fix(i int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if i < 0 || i >= len(q.items) {
		return
	}
	if !q.down(i) {
		q.up(i)
	}
}

// RemoveAt removes and returns the item at index i, if valid.
func (q *IndexedPriorityQueue[T]) RemoveAt(i int) (item T, ok bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if i < 0 || i >= len(q.items) {
		return item, false
	}
	last := len(q.items) - 1
	if i != last {
		q.swap(i, last)
	}
	item = q.items[last]
	q.items = q.items[:last]
	if i < len(q.items) {
		if !q.down(i) {
			q.up(i)
		}
	}
	return item, true
}

// UpdateAt replaces the element at index i and restores invariants.
func (q *IndexedPriorityQueue[T]) UpdateAt(i int, x T) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if i < 0 || i >= len(q.items) {
		return false
	}
	q.items[i] = x
	if !q.down(i) {
		q.up(i)
	}
	return true
}

// Internal helpers (callers must hold write lock)

func (q *IndexedPriorityQueue[T]) lessIdx(i, j int) bool { return q.cmp(q.items[i], q.items[j]) }

func (q *IndexedPriorityQueue[T]) swap(i, j int) {
	if i == j {
		return
	}
	q.items[i], q.items[j] = q.items[j], q.items[i]
	if q.onSwap != nil {
		q.onSwap(i, j, q.items)
	}
}

func (q *IndexedPriorityQueue[T]) up(i int) {
	idx := i
	for {
		p := (idx - 1) / 2
		if idx == 0 || !q.lessIdx(idx, p) {
			break
		}
		q.swap(idx, p)
		idx = p
	}
}

// down moves item at i down; returns true if moved down.
func (q *IndexedPriorityQueue[T]) down(i int) bool {
	idx := i
	n := len(q.items)
	moved := false
	for {
		l := 2*idx + 1
		if l >= n {
			break
		}
		smallest := l
		r := l + 1
		if r < n && q.lessIdx(r, l) {
			smallest = r
		}
		if !q.lessIdx(smallest, idx) {
			break
		}
		q.swap(idx, smallest)
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
