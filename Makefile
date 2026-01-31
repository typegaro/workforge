BIN_DIR := bin
BIN := $(BIN_DIR)/wf
PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
VERBOSE_FLAG := -ldflags "-X workforge/internal/infra/log.verboseFlag=true"

.PHONY: build build-verbose clean fmt install test uninstall

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN) ./cmd/wf

build-verbose:
	mkdir -p $(BIN_DIR)
	go build $(VERBOSE_FLAG) -o $(BIN) ./cmd/wf

fmt:
	gofmt -w cmd internal

test:
	go test ./...

install: build
	mkdir -p $(BINDIR)
	install -m 755 $(BIN) $(BINDIR)/wf

uninstall:
	rm -f $(BINDIR)/wf $(BIN)

clean:
	rm -f $(BIN)
