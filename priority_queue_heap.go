// Package threadsafe implements thread-safe operations.
package threadsafe

import (
	"container/heap"
	"sync"
)

// HeapPriorityQueue is a thread-safe priority queue built on container/heap.
// It is a generic min-heap parameterized by a Less comparator.
// Optional onSwap callback is invoked under write lock when indices swap.
// The zero value is not ready; construct via NewHeapPriorityQueue.
//
// Semantics match PriorityQueue[T].
//
// Note: Like container/heap, internal indices are unstable across operations.
// If you store indices outside, use onSwap to update them.
//
// Complexity: Push/Pop/Fix/RemoveAt/UpdateAt are O(log n); Peek O(1).
// Range/Slice do not mutate the heap.
type HeapPriorityQueue[T any] struct {
	mu     sync.RWMutex
	items  []T
	less   func(a, b T) bool
	onSwap func(i, j int, items []T)
}

// NewHeapPriorityQueue creates a new heap-backed priority queue.
// less(a,b) should return true if a has higher priority (i.e., should come before b).
// onSwap is optional; if non-nil it is called whenever two elements swap.
func NewHeapPriorityQueue[T any](less func(a, b T) bool, onSwap func(i, j int, items []T)) *HeapPriorityQueue[T] {
	return &HeapPriorityQueue[T]{less: less, onSwap: onSwap}
}

// Push inserts one or more items into the queue.
func (h *HeapPriorityQueue[T]) Push(items ...T) {
	if len(items) == 0 {
		return
	}
	h.mu.Lock()
	ad := heapAdapter[T]{h: h}
	for _, x := range items {
		heap.Push(&ad, x)
	}
	h.mu.Unlock()
}

// Pop removes and returns the minimum item.
func (h *HeapPriorityQueue[T]) Pop() (item T, ok bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.items) == 0 {
		return item, false
	}
	ad := heapAdapter[T]{h: h}
	v := heap.Pop(&ad)
	return v.(T), true
}

// Peek returns the minimum item without removing it.
func (h *HeapPriorityQueue[T]) Peek() (item T, ok bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.items) == 0 {
		return item, false
	}
	return h.items[0], true
}

// Len returns the number of items.
func (h *HeapPriorityQueue[T]) Len() int {
	h.mu.RLock()
	l := len(h.items)
	h.mu.RUnlock()
	return l
}

// Clear removes all items.
func (h *HeapPriorityQueue[T]) Clear() {
	h.mu.Lock()
	h.items = nil
	h.mu.Unlock()
}

// Slice returns a copy of items in arbitrary internal order.
func (h *HeapPriorityQueue[T]) Slice() []T {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.items) == 0 {
		return nil
	}
	cp := make([]T, len(h.items))
	copy(cp, h.items)
	return cp
}

// Range iterates a snapshot in arbitrary order.
func (h *HeapPriorityQueue[T]) Range(f func(item T) bool) {
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
func (h *HeapPriorityQueue[T]) Fix(i int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if i < 0 || i >= len(h.items) {
		return
	}
	ad := heapAdapter[T]{h: h}
	heap.Fix(&ad, i)
}

// RemoveAt removes and returns the item at index i.
func (h *HeapPriorityQueue[T]) RemoveAt(i int) (item T, ok bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if i < 0 || i >= len(h.items) {
		return item, false
	}
	ad := heapAdapter[T]{h: h}
	v := ad.removeNoCheck(i)
	return v, true
}

// UpdateAt replaces the element at index i with x and restores invariants.
func (h *HeapPriorityQueue[T]) UpdateAt(i int, x T) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if i < 0 || i >= len(h.items) {
		return false
	}
	h.items[i] = x
	ad := heapAdapter[T]{h: h}
	heap.Fix(&ad, i)
	return true
}

// heapAdapter implements heap.Interface over HeapPriorityQueue's storage.
type heapAdapter[T any] struct{ h *HeapPriorityQueue[T] }

func (a heapAdapter[T]) Len() int { return len(a.h.items) }

func (a heapAdapter[T]) Less(i, j int) bool { return a.h.less(a.h.items[i], a.h.items[j]) }

func (a heapAdapter[T]) Swap(i, j int) {
	if i == j {
		return
	}
	a.h.items[i], a.h.items[j] = a.h.items[j], a.h.items[i]
	if a.h.onSwap != nil {
		a.h.onSwap(i, j, a.h.items)
	}
}

func (a *heapAdapter[T]) Push(x any) {
	v := x.(T)
	a.h.items = append(a.h.items, v)
}

func (a *heapAdapter[T]) Pop() any {
	n := len(a.h.items)
	v := a.h.items[n-1]
	a.h.items[n-1] = *new(T) // avoid keeping reference; zero it
	a.h.items = a.h.items[:n-1]
	return v
}

// removeNoCheck removes at i, assuming bounds have been checked. Mirrors heap.Remove.
func (a *heapAdapter[T]) removeNoCheck(i int) T {
	// Fast path: remove last
	n := len(a.h.items)
	last := n - 1
	if i == last {
		v := a.h.items[last]
		a.h.items[last] = *new(T)
		a.h.items = a.h.items[:last]
		return v
	}
	// Swap i with last, pop last, then Fix at i.
	a.Swap(i, last)
	v := a.h.items[last]
	a.h.items[last] = *new(T)
	a.h.items = a.h.items[:last]
	// After replacing position i with former last, restore order.
	heap.Fix(a, i)
	return v
}
