.PHONY: build test test-race vet lint cover explain

.DEFAULT_GOAL := explain

explain:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Options for test targets:"
	@echo "  [N=...] - Number of times to run burst tests (default 1)"
	@echo "  [V=1]   - Add V=1 for verbose output"
	@echo ""
	@echo "Targets:"
	@echo "  test             - Run tests."
	@echo "  test-race        - Run tests for race conditions."
	@echo "  bench            - Run benchmarks."
	@echo "  lint             - Run golangci-lint, including multiple linters (see .golangci.yml)."
	@echo "  explain          - Display this help message."

# Flag V=1 for verbose mode
TEST_FLAGS :=
ifdef V
	TEST_FLAGS += -v
endif

# Number of times to run burst tests, default 1
N ?= 1

test:
	@echo "==> Running tests..."
	@go test -count=$(N) $(TEST_FLAGS) ./...

test-race:
	@echo "==> Running race tests..."
	@go test -count=$(N) $(TEST_FLAGS) -race ./...

bench:
	@echo "==> Running benchmarks..."
	@go test -count=$(N) $(TEST_FLAGS) -bench=. -benchmem -benchtime=5s ./...

lint:
	@echo "==> Running golangci-lint..."
	@golangci-lint run
