VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build lint test vet ci clean install

build:
	go build -ldflags "$(LDFLAGS)" -o attn .

lint:
	golangci-lint run ./...

test:
	go test -race ./...

vet:
	go vet ./...

ci: vet test lint

clean:
	rm -f attn

install: build
	cp attn ~/bin/attn
