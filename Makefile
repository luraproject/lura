.PHONY: all deps test build benchmark coveralls

DEP_VERSION=0.5.0
OS := $(shell uname | tr '[:upper:]' '[:lower:]')
GIT_COMMIT := $(shell git rev-parse --short=7 HEAD)

all: test build

test:
	go generate ./...
	go test -cover -race ./...
	go test -tags integration ./test

benchmark:
	@mkdir -p bench_res
	@touch bench_res/${GIT_COMMIT}.out
	@go test -run none -bench . -benchmem ./... >> bench_res/${GIT_COMMIT}.out

build:
	go build ./...

coveralls: all
	go get github.com/mattn/goveralls
	go install github.com/mattn/goveralls
	sh coverage.sh --coveralls
