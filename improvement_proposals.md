Thread-safe Set and Queue Fixes (PRD)

## 1) RWMutexSet zero-value panic

- **Problem**: `RWMutexSet` embeds an uninitialized `map[T]struct{}`. Calling `Add` on a zero-value set (e.g., `var s threadsafe.RWMutexSet[int]`) writes to a nil map and panics.
- **Impact**: Violates the repository pattern where most types are safe at zero value (e.g., `RWMutexQueue`), surprises callers, and can crash services if a zero value escapes DI/init code.
- **Repro**:
  ```go
  var s threadsafe.RWMutexSet[int]
  s.Add(1) // panic: assignment to entry in nil map
  ```
- **Proposed fix**: lazily initialize `items` on first mutation so the zero value is usable. No API change.
  ```go
  // Add stores an item in the set.
  func (s *RWMutexSet[T]) Add(item T) (added bool) {
      s.mu.Lock()
      if s.items == nil { // allow zero-value usage
          s.items = make(map[T]struct{})
      }
      if _, exists := s.items[item]; !exists {
          s.items[item] = struct{}{}
          s.size++
          s.mu.Unlock()
          return true
      }
      s.mu.Unlock()
      return false
  }
  ```
- **Notes**: No behavioral change for existing constructor users. Optional follow-up: add a doc comment stating zero-value is ready to use.

## 2) RWMutexQueue Range concurrency race

- **Problem**: `Range` takes a slice view under `RLock`, unlocks, then iterates. Concurrent `Push`/`Pop` under `Lock` mutate the same backing array, so `Range` reads without synchronization.
- **Impact**: Data race in concurrent scenarios; possible stale or corrupted reads. Violates the “All operations must be safe for concurrent use” contract of `Queue`.
- **Repro (conceptual)**:
  ```go
  q := &threadsafe.RWMutexQueue[int]{}
  var wg sync.WaitGroup
  wg.Add(2)
  go func() { defer wg.Done(); q.Range(func(int) bool { time.Sleep(time.Microsecond); return true }) }()
  go func() { defer wg.Done(); q.Push(1); q.Pop() }()
  wg.Wait() // race detector flags unsynchronized access
  ```
- **Proposed fix**: Snapshot under lock (matching `All`) or keep the lock during iteration. Snapshot keeps callers’ callback isolated from locks.
  ```go
  // Range calls f sequentially for each item from front to back.
  func (q *RWMutexQueue[T]) Range(f func(item T) bool) {
      q.mu.RLock()
      snapshot := make([]T, len(q.items)-q.head)
      copy(snapshot, q.items[q.head:])
      q.mu.RUnlock()

      for _, it := range snapshot {
          if !f(it) {
              break
          }
      }
  }
  ```
- **Notes**: This aligns `Range` with `All`, preserves lock-free callback execution, and removes the race. Performance impact is minimal because a copy already exists in `All` and the queue’s size is bounded by workload.
