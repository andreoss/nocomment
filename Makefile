GO ?= go
ANTLR_VERSION ?= 4.13.2

.PHONY: all generate fetch build test corpus clean

all: generate build

generate: fetch
	./scripts/generate-lexers.sh

fetch:
	./scripts/fetch-grammars.sh

build: generate
	$(GO) build ./...

test: generate
	$(GO) test ./...

corpus:
	./scripts/fetch-corpus.sh

clean:
	rm -rf internal/lexer/generated internal/conformance/testdata/files tools grammars/golang bin scratch
