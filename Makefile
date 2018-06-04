.PHONY: all deps test build benchmark coveralls

DEP_VERSION=0.4.1
OS := $(shell uname | tr '[:upper:]' '[:lower:]')

all: deps test build

prepare:
	@echo "Installing dep..."
	@curl -L -s https://github.com/golang/dep/releases/download/v${DEP_VERSION}/dep-${OS}-amd64 -o ${GOPATH}/bin/dep
	@chmod a+x ${GOPATH}/bin/dep

deps:
	@echo "Setting up the vendors folder..."
	@dep ensure -v
	@echo ""
	@echo "Resolved dependencies:"
	@dep status
	@echo ""

test:
	go test -cover -race ./...

benchmark:
	@echo "Proxy middleware stack"
	@go test -bench=BenchmarkProxyStack -benchtime=3s ./proxy
	@echo "Proxy middlewares"
	@go test -bench="BenchmarkNewLoadBalanced|BenchmarkNewConcurrent|BenchmarkNewRequestBuilder|BenchmarkNewMergeData" -benchtime=3s ./proxy
	@echo "Response manipulation"
	@echo "Response property whitelisting"
	@go test -bench=BenchmarkEntityFormatter_whitelistingFilter -benchtime=3s ./proxy
	@echo "Response property blacklisting"
	@go test -bench=BenchmarkEntityFormatter_blacklistingFilter -benchtime=3s ./proxy
	@echo "Response property groupping"
	@go test -bench=BenchmarkEntityFormatter_grouping -benchtime=3s ./proxy
	@echo "Response property mapping"
	@go test -bench=BenchmarkEntityFormatter_mapping -benchtime=3s ./proxy
	@echo "Request generator"
	@go test -bench=BenchmarkRequestGeneratePath -benchtime=3s ./proxy

build:
	go build ./...

coveralls: all
	go get github.com/mattn/goveralls
	sh coverage.sh --coveralls
