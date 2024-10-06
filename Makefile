GO ?= go
ANTLR_VERSION ?= 4.13.2

.PHONY: all generate fetch build test corpus bench cover clean

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

bench: generate
	$(GO) test -run='^$$' -bench=BenchmarkStrip -benchtime=100ms ./internal/bench

cover: generate
	$(GO) test -cover ./internal/cli ./internal/filter ./internal/lang ./internal/langs ./internal/lexer ./internal/process ./internal/scan

clean:
	rm -rf internal/lexer/generated internal/conformance/testdata/files tools grammars/golang bin scratch
