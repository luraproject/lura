.PHONY: all deps build

# This Makefile is a simple example that demonstrates usual steps to build a binary that can be run in the same
# architecture that was compiled in. The "ldflags" in the build assure that any needed dependency is included in the
# binary and no external dependencies are needed to run the service.

KRAKEND_VERSION=$(shell git describe --always --long --dirty --tags)
BIN_NAME=krakend_rss_example_${KRAKEND_VERSION}

all: deps build

deps:
	go get "github.com/devopsfaith/krakend/config/viper"
	go get "github.com/devopsfaith/krakend/proxy"
	go get "github.com/devopsfaith/krakend/router/gin"
	go get "github.com/devopsfaith/krakend/logging/gologging"

build:
	go build -ldflags="-X github.com/devopsfaith/krakend/core.KrakendVersion=${KRAKEND_VERSION}" -o ${BIN_NAME}
	@echo "You can now use ./${BIN_NAME}"
