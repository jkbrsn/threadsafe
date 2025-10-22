package threadsafe

import (
	"reflect"
	"slices"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// queueTestSuite is a generic test suite for the Queue interface.
// It can be instantiated with different item types.
type queueTestSuite[T any] struct {
	newQueue func() Queue[T]
	item1    T
	item2    T
	item3    T
}

func TestRWMutexQueueImplementsQueue(_ *testing.T) {
	var _ Queue[string] = &RWMutexQueue[string]{}
}

// TestBasicOperations verifies Push, Pop, Peek, Len, Clear.
func (s *queueTestSuite[T]) TestBasicOperations(t *testing.T) {
	q := s.newQueue()
	assert.Equal(t, 0, q.Len())

	// Push items
	q.Push(s.item1, s.item2)
	assert.Equal(t, 2, q.Len())

	// Peek should return first item without removal
	item, ok := q.Peek()
	assert.True(t, ok)
	assert.Equal(t, s.item1, item)
	assert.Equal(t, 2, q.Len())

	// Pop items in FIFO order
	item, ok = q.Pop()
	assert.True(t, ok)
	assert.Equal(t, s.item1, item)
	assert.Equal(t, 1, q.Len())

	item, ok = q.Pop()
	assert.True(t, ok)
	assert.Equal(t, s.item2, item)
	assert.Equal(t, 0, q.Len())

	// Pop from empty
	_, ok = q.Pop()
	assert.False(t, ok)

	// Clear should be idempotent
	q.Clear()
	assert.Equal(t, 0, q.Len())
}

func (s *queueTestSuite[T]) TestSlice(t *testing.T) {
	q := s.newQueue()

	// Empty slice
	assert.Empty(t, q.Slice())

	// Push items
	q.Push(s.item1, s.item2, s.item3)
	sl := q.Slice()
	expected := []T{s.item1, s.item2, s.item3}
	assert.True(t, slices.EqualFunc(sl, expected, func(a, b T) bool {
		return reflect.DeepEqual(a, b)
	}))
}

func (s *queueTestSuite[T]) TestRange(t *testing.T) {
	q := s.newQueue()
	// Add items
	q.Push(s.item1, s.item2, s.item3)

	visited := []T{}
	q.Range(func(it T) bool {
		visited = append(visited, it)
		return true
	})
	assert.Equal(t, 3, len(visited))
	assert.Equal(t, s.item1, visited[0])
	assert.Equal(t, s.item2, visited[1])
	assert.Equal(t, s.item3, visited[2])

	// Early stop
	count := 0
	q.Range(func(_ T) bool {
		count++
		return false
	})
	assert.Equal(t, 1, count)
}

func (s *queueTestSuite[T]) TestAllIterator(t *testing.T) {
	q := s.newQueue()
	q.Push(s.item1, s.item2, s.item3)

	items := collectSeq(q.All())
	assert.Equal(t, []T{s.item1, s.item2, s.item3}, items)

	var calls int
	q.All()(func(_ T) bool {
		calls++
		return false
	})
	assert.Equal(t, 1, calls)

	var observed []T
	q.All()(func(item T) bool {
		observed = append(observed, item)
		if len(observed) == 1 {
			q.Push(s.item1)
		}
		return true
	})
	assert.Equal(t, []T{s.item1, s.item2, s.item3}, observed)
	assert.Equal(t, 4, q.Len())
}

func runQueueTestSuite[T any](t *testing.T, s *queueTestSuite[T]) {
	t.Run("BasicOperations", s.TestBasicOperations)
	t.Run("Slice", s.TestSlice)
	t.Run("Range", s.TestRange)
	t.Run("AllIterator", s.TestAllIterator)
}

func TestQueueImplementations(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		t.Run("RWMutexQueue", func(t *testing.T) {
			suite := &queueTestSuite[string]{
				newQueue: func() Queue[string] { return NewRWMutexQueue[string]() },
				item1:    "a",
				item2:    "b",
				item3:    "c",
			}
			runQueueTestSuite(t, suite)
		})
	})

	t.Run("int", func(t *testing.T) {
		t.Run("RWMutexQueue", func(t *testing.T) {
			suite := &queueTestSuite[int]{
				newQueue: func() Queue[int] { return NewRWMutexQueue[int]() },
				item1:    1,
				item2:    2,
				item3:    3,
			}
			runQueueTestSuite(t, suite)
		})
	})

	t.Run("struct", func(t *testing.T) {
		type testStruct struct{ ID int }
		t.Run("RWMutexQueue", func(t *testing.T) {
			suite := &queueTestSuite[testStruct]{
				newQueue: func() Queue[testStruct] { return NewRWMutexQueue[testStruct]() },
				item1:    testStruct{1},
				item2:    testStruct{2},
				item3:    testStruct{3},
			}
			runQueueTestSuite(t, suite)
		})
	})
}

// testConcurrentQueueAccess tests that the queue remains consistent under
// concurrent enqueues while dequeues happen sequentially afterwards. This keeps
// the test deterministic while still exercising thread-safety code paths.
func testConcurrentQueueAccess(t *testing.T, q Queue[string]) {
	const goroutines = 10
	const perGoroutine = 100

	var wg sync.WaitGroup

	// Concurrent enqueues
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				q.Push(strconv.Itoa(id*perGoroutine + j))
			}
		}(i)
	}

	// Wait for all writers to finish
	wg.Wait()

	// Now dequeue everything sequentially
	total := goroutines * perGoroutine
	for i := 0; i < total; i++ {
		item, ok := q.Pop()
		assert.True(t, ok)
		_ = item // value not important for this test
	}

	// Queue should now be empty
	assert.Equal(t, 0, q.Len())
}

func TestQueueConcurrentAccess(t *testing.T) {
	q := NewRWMutexQueue[string]()
	testConcurrentQueueAccess(t, q)
}
