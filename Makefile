BIN_DIR := bin
BIN := $(BIN_DIR)/wf

.PHONY: build clean fmt install test

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN) ./cmd/wf

fmt:
	gofmt -w cmd internal

test:
	go test ./...

install:
	GOBIN=$(abspath $(BIN_DIR)) go install ./cmd/wf

clean:
	rm -f $(BIN)
