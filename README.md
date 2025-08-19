# threadsafe [![Go Documentation](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)][godocs]  [![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]

[godocs]: http://godoc.org/github.com/jkbrsn/threadsafe
[license]: /LICENSE


The threadsafe package provides thread-safe operations for various data structures common in concurrent Go applications.

The interfaces provided by the package are generic, and attempt to be quite exhaustive feature wise. If a more minimal interface would be better aligned for your application, create it as needed.

All interface implementations in this package are thread-safe and can be used concurrently.

## Tests and benchmarks

The package provides a Makefile with targets for running tests and benchmarks:

```bash
make test
make bench
```

## Contribute

For contributions, please open a GitHub issue with your questions and suggestions. Before submitting an issue, have a look at the existing [TODO list](TODO.md) to see if your idea is already in the works.
