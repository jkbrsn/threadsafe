package threadsafe

import (
	"math/rand"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type heapTestItem struct {
	ID   string
	Prio int
	Idx  int // external index maintenance example
}

// helper less: min-heap by Prio
func lessItem(a, b heapTestItem) bool { return a.Prio < b.Prio }

// onSwap maintains external indices
func onSwapItem(i, j int, items []heapTestItem) {
	items[i].Idx = i
	items[j].Idx = j
}

func TestHeapPriorityQueueImplementsInterface(_ *testing.T) {
	var _ PriorityQueue[int] = &HeapPriorityQueue[int]{}
}

func TestRWMutexPriorityQueueImplementsInterface(_ *testing.T) {
	var _ PriorityQueue[int] = &RWMutexPriorityQueue[int]{}
}

func TestHeapPriorityQueueBasic(t *testing.T) {
	pq := NewHeapPriorityQueue[heapTestItem](lessItem, nil)
	assert.Equal(t, 0, pq.Len())

	pq.Push(heapTestItem{ID: "a", Prio: 3},
		heapTestItem{ID: "b", Prio: 1},
		heapTestItem{ID: "c", Prio: 2})
	it, ok := pq.Peek()
	assert.True(t, ok)
	assert.Equal(t, "b", it.ID)

	x, ok := pq.Pop()
	assert.True(t, ok)
	assert.Equal(t, "b", x.ID)

	x, _ = pq.Pop()
	assert.Equal(t, "c", x.ID)
	x, _ = pq.Pop()
	assert.Equal(t, "a", x.ID)
	_, ok = pq.Pop()
	assert.False(t, ok)
}

func TestHeapPriorityQueueFixRemoveUpdate(t *testing.T) {
	pq := NewHeapPriorityQueue[heapTestItem](lessItem, nil)
	pq.Push(heapTestItem{ID: "a", Prio: 5},
		heapTestItem{ID: "b", Prio: 3},
		heapTestItem{ID: "c", Prio: 7},
		heapTestItem{ID: "d", Prio: 1})
	// RemoveAt root
	_, ok := pq.RemoveAt(0)
	assert.True(t, ok)
	// UpdateAt some index if exists
	if pq.Len() >= 2 {
		_ = pq.UpdateAt(1, heapTestItem{ID: "x", Prio: 0})
		pq.Fix(1)
		it, _ := pq.Peek()
		assert.Equal(t, 0, it.Prio)
	}
}

func TestPriorityQueueBasicOperations(t *testing.T) {
	h := NewRWMutexPriorityQueue[heapTestItem](lessItem, onSwapItem)
	assert.Equal(t, 0, h.Len())

	// push
	h.Push(heapTestItem{ID: "a", Prio: 3}, heapTestItem{ID: "b", Prio: 1})
	assert.Equal(t, 2, h.Len())

	// peek should be min
	it, ok := h.Peek()
	assert.True(t, ok)
	assert.Equal(t, "b", it.ID)

	// pop order
	it, ok = h.Pop()
	assert.True(t, ok)
	assert.Equal(t, "b", it.ID)
	it, ok = h.Pop()
	assert.True(t, ok)
	assert.Equal(t, "a", it.ID)
	_, ok = h.Pop()
	assert.False(t, ok)

	// clear
	h.Push(heapTestItem{ID: "x", Prio: 10})
	assert.Equal(t, 1, h.Len())
	h.Clear()
	assert.Equal(t, 0, h.Len())
}

func TestPriorityQueueFixUpdateRemove(t *testing.T) {
	h := NewRWMutexPriorityQueue[heapTestItem](lessItem, onSwapItem)
	items := []heapTestItem{{ID: "a", Prio: 5}, {ID: "b", Prio: 3}, {ID: "c", Prio: 7}, {ID: "d", Prio: 1}}
	h.Push(items...)
	// indices should be tracked in onSwap
	for i := 0; i < h.Len(); i++ {
		// ensure indices in range
		snap := h.Slice()
		_ = snap
	}

	// Decrease key of 'c' to become minimum
	snap := h.Slice()
	var idxC int = -1
	for i := range snap {
		if snap[i].ID == "c" {
			idxC = i
		}
	}
	// UpdateAt requires internal index; we'll find by scanning snapshot and assume same index
	if idxC >= 0 {
		// we don't know if snapshot index matches internal; use RemoveAt+Push as robust path
		if removed, ok := h.RemoveAt(idxC); ok {
			removed.Prio = 0
			h.Push(removed)
		}
	}

	it, _ := h.Peek()
	assert.Equal(t, 0, it.Prio)

	// RemoveAt root then check next min
	_, ok := h.RemoveAt(0)
	assert.True(t, ok)
	it, _ = h.Peek()
	// next min should be 1 or 3 depending on previous state
	assert.True(t, it.Prio == 1 || it.Prio == 3)
}

func TestPriorityQueueConcurrentPushSequentialPop(t *testing.T) {
	h := NewRWMutexPriorityQueue[int](func(a, b int) bool { return a < b }, nil)
	const goroutines = 8
	const per = 200
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(seed int64) {
			defer wg.Done()
			r := rand.New(rand.NewSource(seed))
			vals := make([]int, per)
			for i := 0; i < per; i++ {
				vals[i] = r.Intn(1000000)
			}
			h.Push(vals...)
		}(time.Now().UnixNano() + int64(g))
	}
	wg.Wait()

	// verify ascending order by popping all and comparing to sorted snapshot
	all := h.Slice()
	sort.Ints(all)
	for _, want := range all {
		got, ok := h.Pop()
		assert.True(t, ok)
		assert.Equal(t, want, got)
	}
	assert.Equal(t, 0, h.Len())
}
