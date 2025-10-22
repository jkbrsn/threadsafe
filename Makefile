.PHONY: test bench lint explain

.DEFAULT_GOAL := explain

explain:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Options for test targets:"
	@echo "  [N=...] - Number of times to run burst tests (default 1)"
	@echo "  [RACE=1] - Add RACE=1 for race conditions."
	@echo "  [V=1]   - Add V=1 for verbose output"
	@echo ""
	@echo "Targets:"
	@echo "  test             - Run tests."
	@echo "  bench            - Run benchmarks."
	@echo "  lint             - Run golangci-lint, including multiple linters (see .golangci.yml)."
	@echo "  explain          - Display this help message."

# Flag V=1 for verbose mode
TEST_FLAGS :=
ifdef V
	TEST_FLAGS += -v
endif
ifdef RACE
	TEST_FLAGS += -race
endif

# Number of times to run burst tests, default 1
N ?= 1

test:
	@echo "==> Running tests..."
	@go test -count=$(N) $(TEST_FLAGS) ./...

bench:
	@echo "==> Running benchmarks..."
	@go test -count=$(N) $(TEST_FLAGS) -bench=. -benchmem -benchtime=2s -cpu=8 -run=^$ -v ./...

lint:
	@echo "==> Running golangci-lint..."
	@golangci-lint run
