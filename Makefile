.PHONY: all test build benchmark

OS := $(shell uname | tr '[:upper:]' '[:lower:]')
GIT_COMMIT := $(shell git rev-parse --short=7 HEAD)

all: test build

generate:
	go generate ./...
	go build -buildmode=plugin -o ./transport/http/client/plugin/tests/lura-client-example.so ./transport/http/client/plugin/tests
	go build -buildmode=plugin -o ./transport/http/server/plugin/tests/lura-server-example.so ./transport/http/server/plugin/tests
	go build -buildmode=plugin -o ./proxy/plugin/tests/lura-request-modifier-example.so ./proxy/plugin/tests/logger
	go build -buildmode=plugin -o ./proxy/plugin/tests/lura-error-example.so ./proxy/plugin/tests/error

test: generate
	go test -cover -race ./...
	#go test -tags integration --coverpkg=./... ./test/...
	go test -tags integration ./transport/...
	go test -tags integration ./proxy/...

benchmark:
	@mkdir -p bench_res
	@touch bench_res/${GIT_COMMIT}.out
	@go test -run none -bench . -benchmem ./... >> bench_res/${GIT_COMMIT}.out

build:
	go build ./...
