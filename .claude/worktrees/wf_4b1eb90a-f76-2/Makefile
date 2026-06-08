BINARY   ?= baize-wiki
OUTPUT   ?= bin/$(BINARY)
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE     ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS  := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: build test lint bench clean cross

build:
	go build $(LDFLAGS) -o $(OUTPUT) ./cmd/$(BINARY)

test:
	go test -v -count=1 ./...

lint:
	golangci-lint run ./...

bench:
	go test -bench=. -benchmem ./...

clean:
	rm -rf bin/ dist/

cross:
	$(foreach platform,$(PLATFORMS), \
		GOOS=$(word 1,$(subst /, ,$(platform))) \
		GOARCH=$(word 2,$(subst /, ,$(platform))) \
		go build $(LDFLAGS) -o bin/$(BINARY)-$(subst /,-,$(platform)) ./cmd/$(BINARY); \
	)
