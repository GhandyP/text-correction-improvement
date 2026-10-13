GO ?= go

.PHONY: check build test vet fmt tidy run clean

# Everything CI runs, in the same order.
check: fmt vet build test

build:
	$(GO) build ./...

test:
	$(GO) test ./... -count=1

vet:
	$(GO) vet ./...

fmt:
	@out="$$(gofmt -l .)"; \
	if [ -n "$$out" ]; then \
		echo "these files need gofmt:"; \
		echo "$$out"; \
		exit 1; \
	fi

tidy:
	$(GO) mod tidy

# make run ARGS="-file draft.txt"
run:
	$(GO) run . $(ARGS)

clean:
	$(GO) clean
