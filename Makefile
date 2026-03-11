DEFAULT := help

GO ?= go
CMD_PKG := ./cmd/tutor

.PHONY: help fmt vet test check build run once clean

help:
	@printf "%s\n" \
	"Targets:" \
	"  make fmt    - format Go code (modifies files)" \
	"  make vet    - run go vet" \
	"  make test   - run go test" \
	"  make check  - run vet + test" \
	"  make build  - build ./cmd/tutor into $(BIN)" \
	"  make run    - run the turor" \
	"  make clean  - remove built binary"

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

check: vet test

build:
	$(GO) build -o "$(BIN)" $(CMD_PKG)

run:
	$(GO) run $(CMD_PKG)

clean:
	rm -f "$(BIN)"

binlink:
	rm ../tutor && ln -s $(pwd)/cmd/py/tutor.py ~/.local/bin/tutor

