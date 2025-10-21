# Go 1.23+ Improvements - Product Requirements Document

## Overview

This document outlines improvements to the `threadsafe` package to leverage features introduced in Go 1.23 and later versions. The improvements are organized into bite-sized, actionable tasks that can be implemented incrementally.

**Note:** This PRD has been reviewed against Go 1.24 and Go 1.25 release notes to ensure recommendations remain current and take advantage of any relevant new features.

## Goals

1. Modernize the codebase to use Go 1.23+ features
2. Improve performance and code clarity where newer Go versions offer better primitives
3. Provide idiomatic iterator support for all data structures
4. Maintain backward compatibility while adding new functionality
5. Ensure comprehensive test coverage for all new features

## Release: Go 1.23+ Feature Adoption

### Compatibility Notes

**Go Version Requirements:**
- Minimum: Go 1.23 (for iterator support and sync.Map.Clear())
- Current: Go 1.25.3 (as specified in go.mod)
- All recommendations are compatible with Go 1.24 and Go 1.25

**Go 1.24/1.25 Relevant Changes:**
- **sync.Map performance improvement (Go 1.24)**: The sync.Map implementation changed to improve performance, particularly for modifications. Our SyncMap and SyncMapSet implementations will automatically benefit.
- **Improved sync.Map (Go 1.25)**: Further improvements to sync.Map with better concurrency. No code changes needed.
- **hash.Cloner interface (Go 1.24)**: New interface for hash cloning. Not directly applicable to threadsafe types.
- **testing.B.Loop (Go 1.24)**: New benchmark method. Can be adopted in benchmark updates (see Phase 9).
- **weak package (Go 1.24)**: Could be evaluated for weak reference implementations (see Phase 7 additions).

### Breaking Changes to Avoid

**Go 1.24 Changes:**
- None that affect our recommendations

**Go 1.25 Changes:**
- **nil pointer bug fix**: A compiler bug fix (Go 1.21-1.24) may cause some previously-working code to panic. Our recommendations don't involve patterns that would trigger this.
- No breaking changes affect our implementation plans

---

## Phase 1: Quick Wins (Immediate Impact, Low Effort)

### Task 1.1: Update sync.Map.Clear() in SyncMap
**Priority:** HIGH
**Effort:** LOW
**Files:** `map_sync.go:47-51`

**Description:**
Replace the manual Range + Delete loop with Go 1.23's native `sync.Map.Clear()` method.

**Current Implementation:**
```go
func (s *SyncMap[K, V]) Clear() {
    s.values.Range(func(key, _ any) bool {
        s.values.Delete(key)
        return true
    })
}
```

**New Implementation:**
```go
func (s *SyncMap[K, V]) Clear() {
    s.values.Clear()
}
```

**Acceptance Criteria:**
- [ ] Implementation updated to use `sync.Map.Clear()`
- [ ] Existing tests pass
- [ ] Benchmark shows performance improvement (expected)

---

### Task 1.2: Update sync.Map.Clear() in SyncMapSet
**Priority:** HIGH
**Effort:** LOW
**Files:** `set_syncmap.go:52-56`

**Description:**
Replace the manual Range + Delete loop with Go 1.23's native `sync.Map.Clear()` method.

**Current Implementation:**
```go
func (s *SyncMapSet[T]) Clear() {
    s.items.Range(func(key, _ any) bool {
        s.items.Delete(key)
        return true
    })
}
```

**New Implementation:**
```go
func (s *SyncMapSet[T]) Clear() {
    s.items.Clear()
}
```

**Acceptance Criteria:**
- [ ] Implementation updated to use `sync.Map.Clear()`
- [ ] Existing tests pass
- [ ] Benchmark shows performance improvement (expected)

---

## Phase 2: Iterator Support - Interface Design

### Task 2.1: Add Iterator Methods to Map Interface
**Priority:** HIGH
**Effort:** MEDIUM
**Files:** `map.go:6-42`

**Description:**
Add iterator-returning methods to the `Map[K, V]` interface to support Go 1.23's range-over-function feature.

**New Methods to Add:**
```go
type Map[K comparable, V any] interface {
    // ... existing methods ...

    // All returns an iterator over key-value pairs in the map.
    // The iteration order is not guaranteed to be consistent.
    All() iter.Seq2[K, V]

    // Keys returns an iterator over keys in the map.
    // The iteration order is not guaranteed to be consistent.
    Keys() iter.Seq[K]

    // Values returns an iterator over values in the map.
    // The iteration order is not guaranteed to be consistent.
    Values() iter.Seq[V]
}
```

**Acceptance Criteria:**
- [ ] Import `iter` package added to file
- [ ] Three new methods added to interface with clear documentation
- [ ] Documentation specifies no guaranteed iteration order
- [ ] Code compiles (implementations will fail until Phase 3)

---

### Task 2.2: Add Iterator Methods to Set Interface
**Priority:** HIGH
**Effort:** LOW
**Files:** `set.go:4-21`

**Description:**
Add iterator-returning method to the `Set[T]` interface.

**New Method to Add:**
```go
type Set[T comparable] interface {
    // ... existing methods ...

    // All returns an iterator over all items in the set.
    // The iteration order is not guaranteed to be consistent.
    All() iter.Seq[T]
}
```

**Acceptance Criteria:**
- [ ] Import `iter` package added to file
- [ ] New method added to interface with clear documentation
- [ ] Documentation specifies no guaranteed iteration order
- [ ] Code compiles (implementations will fail until Phase 3)

---

### Task 2.3: Add Iterator Methods to Queue Interface
**Priority:** HIGH
**Effort:** LOW
**Files:** `queue.go:9-35`

**Description:**
Add iterator-returning method to the `Queue[T]` interface.

**New Method to Add:**
```go
type Queue[T any] interface {
    // ... existing methods ...

    // All returns an iterator over items in the queue from front to back.
    // The iteration order matches the queue order (FIFO).
    All() iter.Seq[T]
}
```

**Acceptance Criteria:**
- [ ] Import `iter` package added to file
- [ ] New method added to interface with clear documentation
- [ ] Documentation specifies FIFO iteration order
- [ ] Code compiles (implementations will fail until Phase 3)

---

### Task 2.4: Add Iterator Methods to Heap Interface
**Priority:** HIGH
**Effort:** LOW
**Files:** `heap.go:7-33`

**Description:**
Add iterator-returning method to the `Heap[T]` interface.

**New Method to Add:**
```go
type Heap[T any] interface {
    // ... existing methods ...

    // All returns an iterator over items in the heap in internal heap order (not sorted).
    // The iteration order is implementation-defined and not guaranteed to be priority-sorted.
    All() iter.Seq[T]
}
```

**Acceptance Criteria:**
- [ ] Import `iter` package added to file
- [ ] New method added to interface with clear documentation
- [ ] Documentation clarifies iteration is NOT priority-sorted
- [ ] Code compiles (implementations will fail until Phase 3)

---

### Task 2.5: Add Iterator Methods to PriorityQueue Interface
**Priority:** HIGH
**Effort:** LOW
**Files:** `priority_queue.go:7-27`

**Description:**
Add iterator-returning method to the `PriorityQueue[T]` interface.

**New Method to Add:**
```go
type PriorityQueue[T any] interface {
    // ... existing methods ...

    // All returns an iterator over items in the queue in internal heap order (not sorted).
    // The iteration order is implementation-defined and not guaranteed to be priority-sorted.
    All() iter.Seq[T]
}
```

**Acceptance Criteria:**
- [ ] Import `iter` package added to file
- [ ] New method added to interface with clear documentation
- [ ] Documentation clarifies iteration is NOT priority-sorted
- [ ] Code compiles (implementations will fail until Phase 3)

---

### Task 2.6: Add Iterator Methods to Slice Interface
**Priority:** MEDIUM
**Effort:** LOW
**Files:** `slice.go:4-16`

**Description:**
Add iterator-returning method to the `Slice[T]` interface for consistency.

**New Method to Add:**
```go
type Slice[T any] interface {
    // ... existing methods ...

    // All returns an iterator over all items in the slice in order.
    All() iter.Seq[T]
}
```

**Acceptance Criteria:**
- [ ] Import `iter` package added to file
- [ ] New method added to interface with clear documentation
- [ ] Code compiles (implementations will fail until Phase 3)

---

## Phase 3: Iterator Support - Map Implementations

### Task 3.1: Implement Map.All() for MutexMap
**Priority:** HIGH
**Effort:** MEDIUM
**Files:** `map_mutex.go`

**Description:**
Implement the `All()` iterator method for `MutexMap`.

**Implementation Approach:**
```go
func (m *MutexMap[K, V]) All() iter.Seq2[K, V] {
    return func(yield func(K, V) bool) {
        m.mu.RLock()
        snapshot := maps.Clone(m.values)
        m.mu.RUnlock()

        for k, v := range snapshot {
            if !yield(k, v) {
                return
            }
        }
    }
}
```

**Acceptance Criteria:**
- [ ] Import `iter` package
- [ ] Method implemented with proper locking
- [ ] Creates snapshot to avoid holding lock during iteration
- [ ] Respects yield return value for early termination
- [ ] Unit tests added
- [ ] Example usage in tests: `for k, v := range m.All() { ... }`

---

### Task 3.2: Implement Map.Keys() for MutexMap
**Priority:** HIGH
**Effort:** LOW
**Files:** `map_mutex.go`

**Description:**
Implement the `Keys()` iterator method for `MutexMap`.

**Implementation Approach:**
```go
func (m *MutexMap[K, V]) Keys() iter.Seq[K] {
    return func(yield func(K) bool) {
        m.mu.RLock()
        keys := make([]K, 0, len(m.values))
        for k := range m.values {
            keys = append(keys, k)
        }
        m.mu.RUnlock()

        for _, k := range keys {
            if !yield(k) {
                return
            }
        }
    }
}
```

**Acceptance Criteria:**
- [ ] Method implemented with proper locking
- [ ] Creates snapshot to avoid holding lock during iteration
- [ ] Respects yield return value
- [ ] Unit tests added

---

### Task 3.3: Implement Map.Values() for MutexMap
**Priority:** HIGH
**Effort:** LOW
**Files:** `map_mutex.go`

**Description:**
Implement the `Values()` iterator method for `MutexMap`.

**Implementation Approach:**
```go
func (m *MutexMap[K, V]) Values() iter.Seq[V] {
    return func(yield func(V) bool) {
        m.mu.RLock()
        values := make([]V, 0, len(m.values))
        for _, v := range m.values {
            values = append(values, v)
        }
        m.mu.RUnlock()

        for _, v := range values {
            if !yield(v) {
                return
            }
        }
    }
}
```

**Acceptance Criteria:**
- [ ] Method implemented with proper locking
- [ ] Creates snapshot to avoid holding lock during iteration
- [ ] Respects yield return value
- [ ] Unit tests added

---

### Task 3.4: Implement Map Iterators for RWMutexMap
**Priority:** HIGH
**Effort:** MEDIUM
**Files:** `map_rwmutex.go`

**Description:**
Implement `All()`, `Keys()`, and `Values()` iterator methods for `RWMutexMap`. Implementation should be nearly identical to MutexMap but using RWMutex.

**Acceptance Criteria:**
- [ ] All three iterator methods implemented
- [ ] Proper read locking used
- [ ] Snapshot strategy for safe iteration
- [ ] Unit tests added

---

### Task 3.5: Implement Map Iterators for SyncMap
**Priority:** HIGH
**Effort:** MEDIUM
**Files:** `map_sync.go`

**Description:**
Implement `All()`, `Keys()`, and `Values()` iterator methods for `SyncMap`.

**Implementation Approach for All():**
```go
func (s *SyncMap[K, V]) All() iter.Seq2[K, V] {
    return func(yield func(K, V) bool) {
        s.values.Range(func(k, v any) bool {
            return yield(k.(K), v.(V))
        })
    }
}
```

**Note:** SyncMap can use Range directly without snapshotting since sync.Map handles concurrent iteration.

**Acceptance Criteria:**
- [ ] All three iterator methods implemented
- [ ] Leverages sync.Map.Range for efficient iteration
- [ ] Proper type assertions
- [ ] Unit tests added

---

## Phase 4: Iterator Support - Other Data Structures

### Task 4.1: Implement Set.All() for RWMutexSet
**Priority:** HIGH
**Effort:** MEDIUM
**Files:** `set_rwmutex.go`

**Description:**
Implement the `All()` iterator method for `RWMutexSet`.

**Implementation Approach:**
```go
func (s *RWMutexSet[T]) All() iter.Seq[T] {
    return func(yield func(T) bool) {
        s.mu.RLock()
        items := make([]T, 0, len(s.items))
        for item := range s.items {
            items = append(items, item)
        }
        s.mu.RUnlock()

        for _, item := range items {
            if !yield(item) {
                return
            }
        }
    }
}
```

**Acceptance Criteria:**
- [ ] Import `iter` package
- [ ] Method implemented with proper locking
- [ ] Creates snapshot
- [ ] Unit tests added

---

### Task 4.2: Implement Set.All() for SyncMapSet
**Priority:** HIGH
**Effort:** MEDIUM
**Files:** `set_syncmap.go`

**Description:**
Implement the `All()` iterator method for `SyncMapSet`.

**Implementation Approach:**
```go
func (s *SyncMapSet[T]) All() iter.Seq[T] {
    return func(yield func(T) bool) {
        s.items.Range(func(key, _ any) bool {
            return yield(key.(T))
        })
    }
}
```

**Acceptance Criteria:**
- [ ] Import `iter` package
- [ ] Method implemented using Range
- [ ] Unit tests added

---

### Task 4.3: Implement Queue.All() for RWMutexQueue
**Priority:** HIGH
**Effort:** MEDIUM
**Files:** `queue_rwmutex.go`

**Description:**
Implement the `All()` iterator method for `RWMutexQueue`.

**Implementation Approach:**
```go
func (q *RWMutexQueue[T]) All() iter.Seq[T] {
    return func(yield func(T) bool) {
        q.mu.RLock()
        snapshot := make([]T, len(q.items)-q.head)
        copy(snapshot, q.items[q.head:])
        q.mu.RUnlock()

        for _, item := range snapshot {
            if !yield(item) {
                return
            }
        }
    }
}
```

**Acceptance Criteria:**
- [ ] Import `iter` package
- [ ] Method implemented with proper locking
- [ ] Iterates from front to back (FIFO order)
- [ ] Unit tests verify order

---

### Task 4.4: Implement Heap.All() for RWMutexHeap
**Priority:** HIGH
**Effort:** MEDIUM
**Files:** `heap_rwmutex.go`

**Description:**
Implement the `All()` iterator method for `RWMutexHeap`.

**Implementation Approach:**
```go
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
```

**Acceptance Criteria:**
- [ ] Import `iter` package
- [ ] Method implemented with proper locking
- [ ] Documentation clarifies not priority-sorted
- [ ] Unit tests added

---

### Task 4.5: Implement PriorityQueue.All() for CorePriorityQueue
**Priority:** HIGH
**Effort:** MEDIUM
**Files:** `priority_queue_core.go`

**Description:**
Implement the `All()` iterator method for `CorePriorityQueue`.

**Acceptance Criteria:**
- [ ] Import `iter` package
- [ ] Method implemented with proper locking
- [ ] Creates snapshot
- [ ] Unit tests added

---

### Task 4.6: Implement PriorityQueue.All() for IndexedPriorityQueue
**Priority:** HIGH
**Effort:** MEDIUM
**Files:** `priority_queue_indexed.go`

**Description:**
Implement the `All()` iterator method for `IndexedPriorityQueue`.

**Acceptance Criteria:**
- [ ] Import `iter` package
- [ ] Method implemented with proper locking
- [ ] Creates snapshot
- [ ] Unit tests added

---

### Task 4.7: Implement Slice.All() for Implementations
**Priority:** MEDIUM
**Effort:** MEDIUM
**Files:** `slice_mutex.go`, `slice_rwmutex.go`, `slice_sharded.go`

**Description:**
Implement the `All()` iterator method for all Slice implementations.

**Acceptance Criteria:**
- [ ] Method implemented for MutexSlice
- [ ] Method implemented for RWMutexSlice
- [ ] Method implemented for ShardedSlice
- [ ] All use snapshot strategy
- [ ] Unit tests added

---

## Phase 5: Leverage slices Package

### Task 5.1: Use slices.Collect() in Set.Slice()
**Priority:** MEDIUM
**Effort:** LOW
**Files:** `set_rwmutex.go`, `set_syncmap.go`

**Description:**
Refactor `Set.Slice()` implementations to use the new iterator + `slices.Collect()`.

**Example Refactor:**
```go
// Before
func (s *RWMutexSet[T]) Slice() []T {
    s.mu.RLock()
    defer s.mu.RUnlock()
    result := make([]T, 0, len(s.items))
    for item := range s.items {
        result = append(result, item)
    }
    return result
}

// After
func (s *RWMutexSet[T]) Slice() []T {
    return slices.Collect(s.All())
}
```

**Acceptance Criteria:**
- [ ] Import `slices` package
- [ ] Refactored to use `slices.Collect()`
- [ ] Existing tests pass
- [ ] Code is simpler and more idiomatic

---

### Task 5.2: Use slices.Collect() in Queue.Slice()
**Priority:** MEDIUM
**Effort:** LOW
**Files:** `queue_rwmutex.go`

**Description:**
Refactor `Queue.Slice()` implementation to use iterator + `slices.Collect()`.

**Acceptance Criteria:**
- [ ] Import `slices` package
- [ ] Refactored to use `slices.Collect()`
- [ ] Maintains FIFO order
- [ ] Existing tests pass

---

### Task 5.3: Use slices.Collect() in Heap.Slice()
**Priority:** MEDIUM
**Effort:** LOW
**Files:** `heap_rwmutex.go`

**Description:**
Refactor `Heap.Slice()` implementation to use iterator + `slices.Collect()`.

**Acceptance Criteria:**
- [ ] Import `slices` package
- [ ] Refactored to use `slices.Collect()`
- [ ] Existing tests pass

---

### Task 5.4: Add slices Package to Test Utilities
**Priority:** MEDIUM
**Effort:** LOW
**Files:** `*_test.go` files

**Description:**
Update test files to leverage `slices` package functions where applicable.

**Examples:**
- Use `slices.Contains()` instead of manual loops
- Use `slices.Equal()` for slice comparisons
- Use `slices.Sort()` for sorting test data

**Acceptance Criteria:**
- [ ] Test code simplified where applicable
- [ ] All tests pass
- [ ] Benchmark impact measured (if any)

---

## Phase 6: Leverage maps Package

### Task 6.1: Use maps.Collect() in Map Implementations
**Priority:** MEDIUM
**Effort:** LOW
**Files:** Map implementation files

**Description:**
Evaluate whether `maps.Collect()` can improve any existing methods. Likely candidates include helper functions or initialization paths.

**Acceptance Criteria:**
- [ ] Code review completed
- [ ] Opportunities identified and documented
- [ ] Implementations updated if beneficial
- [ ] Benchmarks show no performance regression

---

### Task 6.2: Explore maps.Insert() for SetMany()
**Priority:** LOW
**Effort:** MEDIUM
**Files:** `map_mutex.go:151`, `map_rwmutex.go:151`

**Description:**
Investigate whether `maps.Insert()` provides benefits for `SetMany()` when used with iterators.

**Current Implementation (map_mutex.go:151):**
```go
func (m *MutexMap[K, V]) SetMany(entries map[K]V) {
    m.mu.Lock()
    defer m.mu.Unlock()
    maps.Copy(m.values, entries)
}
```

**Potential Alternative:**
Consider if iterator-based approach offers advantages for specific use cases.

**Acceptance Criteria:**
- [ ] Analysis completed
- [ ] Benchmark comparison if changes proposed
- [ ] Decision documented
- [ ] Implementation updated if beneficial

---

### Task 6.3: Consider maps Functions in CalculateMapDiff()
**Priority:** LOW
**Effort:** MEDIUM
**Files:** `map.go:50-77`

**Description:**
Evaluate whether Go 1.23 `maps` package functions could simplify or improve `CalculateMapDiff()`.

**Current Implementation:**
Uses manual loops over both maps.

**Exploration:**
- Can `maps.All()` simplify iteration?
- Can any other maps functions reduce code complexity?

**Acceptance Criteria:**
- [ ] Analysis completed
- [ ] Decision documented with rationale
- [ ] Implementation updated if beneficial
- [ ] All tests pass

---

## Phase 7: Consider unique Package (Lower Priority)

### Task 7.1: Evaluate unique Package for Set Implementations
**Priority:** LOW
**Effort:** HIGH
**Files:** `set_*.go`

**Description:**
Investigate whether the `unique` package could provide memory savings for Set implementations through value canonicalization.

**Considerations:**
- Memory footprint reduction vs performance overhead
- API compatibility (would require breaking changes or new type)
- Use cases where canonicalization makes sense

**Acceptance Criteria:**
- [ ] Spike/POC completed
- [ ] Performance benchmarks run
- [ ] Memory benchmarks run
- [ ] Decision documented with data
- [ ] If beneficial: design proposal created for separate implementation

**Note:** This likely requires a new type (e.g., `UniqueSet[T]`) rather than modifying existing implementations.

---

### Task 7.2: Evaluate unique Package for Map Keys
**Priority:** LOW
**Effort:** HIGH
**Files:** `map_*.go`

**Description:**
Investigate whether the `unique` package could provide benefits for Map implementations through key canonicalization.

**Go 1.25 Note:** The `unique` package received performance improvements for more eager reclamation of interned values. This makes it more viable for heavy use cases.

**Acceptance Criteria:**
- [ ] Analysis completed
- [ ] Benchmarks if promising
- [ ] Decision documented
- [ ] Separate design proposal if beneficial

---

### Task 7.3: Evaluate weak Package for Cache-like Implementations (Go 1.24+)
**Priority:** LOW
**Effort:** HIGH
**Files:** New files or extensions

**Description:**
Go 1.24 introduced the `weak` package for weak pointers. Evaluate whether this could enable new cache-like data structure implementations.

**Potential Use Cases:**
- Weak-reference cache implementations
- Memory-efficient canonicalization maps (complementing `unique` package)
- Auto-expiring data structures

**Considerations:**
- Would likely be new types, not modifications to existing ones
- Use with `runtime.AddCleanup` (also new in Go 1.24) for cleanup actions
- Performance characteristics vs traditional caching

**Acceptance Criteria:**
- [ ] Spike/POC completed
- [ ] Performance benchmarks run
- [ ] Memory benchmarks run
- [ ] Decision documented with data
- [ ] If beneficial: design proposal created for new types

---

## Phase 8: Documentation and Examples

### Task 8.1: Update README with Go 1.23 Iterator Examples
**Priority:** MEDIUM
**Effort:** LOW
**Files:** `README.md`

**Description:**
Add examples showing modern Go 1.23 iterator usage.

**Examples to Add:**
```go
// Iterate over map
for k, v := range myMap.All() {
    fmt.Println(k, v)
}

// Collect set to sorted slice
items := slices.Sorted(mySet.All())

// Use with slices.Collect
allValues := slices.Collect(myQueue.All())
```

**Acceptance Criteria:**
- [ ] README updated with iterator examples
- [ ] Examples are runnable and correct
- [ ] Clear explanation of Go 1.23 requirement

---

### Task 8.2: Add Iterator Usage to GoDoc Examples
**Priority:** MEDIUM
**Effort:** MEDIUM
**Files:** All implementation files

**Description:**
Add `Example*` functions demonstrating iterator usage for each data structure.

**Acceptance Criteria:**
- [ ] Example functions added for Map iterators
- [ ] Example functions added for Set iterators
- [ ] Example functions added for Queue, Heap, PriorityQueue iterators
- [ ] All examples include expected output
- [ ] `go test` runs examples successfully

---

### Task 8.3: Update Interface Documentation
**Priority:** MEDIUM
**Effort:** LOW
**Files:** `map.go`, `set.go`, `queue.go`, `heap.go`, `priority_queue.go`, `slice.go`

**Description:**
Ensure all interface documentation clearly explains iterator behavior, thread-safety guarantees during iteration, and Go version requirements.

**Acceptance Criteria:**
- [ ] Each iterator method has comprehensive documentation
- [ ] Thread-safety behavior during iteration is clear
- [ ] Iteration order guarantees (or lack thereof) documented
- [ ] Go 1.23 requirement noted in package documentation

---

## Phase 9: Testing and Benchmarking

### Task 9.1: Add Iterator Unit Tests
**Priority:** HIGH
**Effort:** MEDIUM
**Files:** All `*_test.go` files

**Description:**
Add comprehensive unit tests for all new iterator methods.

**Test Cases to Cover:**
- [ ] Basic iteration over populated data structure
- [ ] Iteration over empty data structure
- [ ] Early termination (return false from range)
- [ ] Concurrent iteration with modifications (verify thread-safety)
- [ ] Collect to slices/maps using standard library functions

**Acceptance Criteria:**
- [ ] Unit tests added for all iterator implementations
- [ ] Coverage maintained or improved
- [ ] All tests pass
- [ ] Thread-safety verified

---

### Task 9.2: Add Iterator Benchmarks
**Priority:** MEDIUM
**Effort:** MEDIUM
**Files:** All `*_test.go` files

**Description:**
Add benchmarks comparing iterator-based approaches vs existing Range methods.

**Go 1.24+ Enhancement:** Consider using `testing.B.Loop()` for new benchmarks instead of the traditional `for i := 0; i < b.N; i++` pattern. The new method is faster, less error-prone, and prevents compiler optimizations from eliminating benchmark code.

**Example using B.Loop (Go 1.24+):**
```go
func BenchmarkMapIterator(b *testing.B) {
    m := NewSyncMap[string, int](nil)
    m.Set("key", 42)

    for b.Loop() {
        for k, v := range m.All() {
            _ = k
            _ = v
        }
    }
}
```

**Benchmarks to Add:**
- [ ] Iteration using `All()` + for-range
- [ ] Iteration using `Range()` callback
- [ ] `slices.Collect()` vs manual slice building
- [ ] Memory allocation comparisons
- [ ] (Optional) Refactor existing benchmarks to use `testing.B.Loop()` where appropriate

**Acceptance Criteria:**
- [ ] Benchmarks added for iterators
- [ ] Consider using `testing.B.Loop()` for cleaner benchmark code
- [ ] Performance comparison documented
- [ ] No significant performance regression

---

### Task 9.3: Benchmark sync.Map.Clear() Improvements
**Priority:** MEDIUM
**Effort:** LOW
**Files:** `map_test.go`, `set_test.go`

**Description:**
Add benchmarks to verify performance improvement from using native `sync.Map.Clear()`.

**Acceptance Criteria:**
- [ ] Benchmark for SyncMap.Clear() added
- [ ] Benchmark for SyncMapSet.Clear() added
- [ ] Performance improvement documented

---

## Phase 10: Optional Enhancements

### Task 10.1: Add Filtered Iterators
**Priority:** LOW
**Effort:** MEDIUM
**Files:** New utility file or add to existing interfaces

**Description:**
Consider adding filtered iterator helpers that work with predicates.

**Example:**
```go
// Helper function (not in interface)
func FilterSeq[T any](seq iter.Seq[T], predicate func(T) bool) iter.Seq[T] {
    return func(yield func(T) bool) {
        for v := range seq {
            if predicate(v) {
                if !yield(v) {
                    return
                }
            }
        }
    }
}
```

**Acceptance Criteria:**
- [ ] Design decision documented
- [ ] If implemented: helper functions added
- [ ] Tests and examples added
- [ ] Documentation updated

---

### Task 10.2: Add Transform Iterators
**Priority:** LOW
**Effort:** MEDIUM
**Files:** New utility file

**Description:**
Consider adding iterator transformation helpers.

**Example:**
```go
func MapSeq[T, U any](seq iter.Seq[T], transform func(T) U) iter.Seq[U] {
    return func(yield func(U) bool) {
        for v := range seq {
            if !yield(transform(v)) {
                return
            }
        }
    }
}
```

**Acceptance Criteria:**
- [ ] Design decision documented
- [ ] If implemented: helper functions added
- [ ] Tests and examples added

---

## Success Metrics

1. **Code Quality**
   - All new code passes linting
   - Test coverage maintained or improved (target: >90%)
   - No breaking changes to existing APIs

2. **Performance**
   - sync.Map.Clear() shows measurable improvement (Go 1.23)
   - SyncMap/SyncMapSet automatically benefit from Go 1.24's improved sync.Map implementation
   - SyncMap/SyncMapSet automatically benefit from Go 1.25's further sync.Map improvements
   - Iterator performance comparable to Range methods
   - No regression in existing benchmarks

3. **Usability**
   - Modern, idiomatic Go 1.23 code
   - Clear documentation and examples
   - Positive community feedback

4. **Compatibility**
   - go.mod specifies `go 1.23` minimum
   - All existing functionality preserved
   - New methods are additive only

---

## Dependencies

- Go 1.23 or later required
- No new external dependencies
- Leverages standard library packages: `iter`, `slices`, `maps`

---

## Timeline Estimate

- **Phase 1**: 1-2 hours (quick wins)
- **Phase 2**: 2-3 hours (interface updates)
- **Phase 3**: 4-6 hours (map iterator implementations)
- **Phase 4**: 4-6 hours (other iterator implementations)
- **Phase 5**: 2-3 hours (slices package adoption)
- **Phase 6**: 2-3 hours (maps package exploration)
- **Phase 7**: 5-10 hours (unique/weak package evaluation, optional, includes new Task 7.3)
- **Phase 8**: 2-3 hours (documentation)
- **Phase 9**: 4-6 hours (testing and benchmarking)
- **Phase 10**: 2-4 hours (optional enhancements)

**Total Estimated Effort**: 28-46 hours (excluding optional tasks in Phases 7 and 10)

---

## Notes

### Implementation Order
- Each task can be worked on independently after its dependencies are complete
- Phases 1-2 should be completed in order
- Phases 3-4 can be parallelized
- Phases 5-7 can be worked on after Phase 4
- Testing (Phase 9) should be done alongside implementation
- Documentation (Phase 8) should be ongoing throughout

### Go Version Considerations
- **Current version (go.mod)**: Go 1.25.3
- **Minimum required**: Go 1.23 (for iterators and sync.Map.Clear())
- **Automatic benefits**: Using Go 1.24+ provides automatic performance improvements to sync.Map-based implementations (SyncMap, SyncMapSet)
- **Optional Go 1.24+ features**: testing.B.Loop(), weak package (Task 7.3)

### Performance Expectations
- **Immediate wins**: sync.Map.Clear() optimization (Task 1.1, 1.2)
- **Automatic gains**: Go 1.24 and 1.25 improved sync.Map implementation
- **New capabilities**: Iterator support enables modern idiomatic Go code
