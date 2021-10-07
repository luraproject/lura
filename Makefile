.PHONY: all test build benchmark

OS := $(shell uname | tr '[:upper:]' '[:lower:]')
GIT_COMMIT := $(shell git rev-parse --short=7 HEAD)

all: test build

test:
	go generate ./...
	go build -buildmode=plugin -o ./transport/http/client/plugin/tests/lura-client-example.so ./transport/http/client/plugin/tests
	go build -buildmode=plugin -o ./transport/http/server/plugin/tests/lura-server-example.so ./transport/http/server/plugin/tests
	go build -buildmode=plugin -o ./proxy/plugin/tests/lura-client-example.so ./proxy/plugin/tests
	go test -cover -race ./...
	go test -tags integration ./test/...
	go test -tags integration ./transport/...
	go test -tags integration ./proxy/...

benchmark:
	@mkdir -p bench_res
	@touch bench_res/${GIT_COMMIT}.out
	@go test -run none -bench . -benchmem ./... >> bench_res/${GIT_COMMIT}.out

build:
	go build ./...
