# threadsafe

[![Go Documentation](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)][godocs]
[![Go Report Card](https://goreportcard.com/badge/github.com/jkbrsn/threadsafe)](https://goreportcard.com/report/github.com/jkbrsn/threadsafe)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]

[godocs]: http://godoc.org/github.com/jkbrsn/threadsafe
[license]: /LICENSE

The threadsafe package provides thread-safe operations for various data structures common in concurrent Go applications.

The interfaces provided by the package are generic, and attempt to be quite exhaustive feature wise. If a more minimal interface would be better aligned for your application, create it as needed.

All interface implementations in this package are thread-safe and can be used concurrently.

## Key Features

- Generic, thread-safe maps, sets, queues, heaps, and priority queues.
- Iterator-first APIs for idiomatic `range` loops.
  - Note: due to the snapshotting used to keep the iterators thread-safe, some iterators may be less performant than a standard Range iteration.
- Multiple concurrency strategies (mutex, RWMutex, sync.Map) so you can pick the right trade-offs.

## Tests and benchmarks

The package provides a Makefile with targets for running tests and benchmarks:

```bash
make test
make bench
```

## Contribute

For contributions, please open a GitHub issue with your questions and suggestions. Before submitting an issue, have a look at the existing [TODO list](TODO.md) to see if your idea is already in the works.
