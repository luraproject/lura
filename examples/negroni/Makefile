.PHONY: all deps build

# This Makefile is a simple example that demonstrates usual steps to build a binary that can be run in the same
# architecture that was compiled in. The "ldflags" in the build assure that any needed dependency is included in the
# binary and no external dependencies are needed to run the service.

KRAKEND_VERSION=$(shell git describe --always --long --dirty --tags)
BIN_NAME=krakend_negroni_example_${KRAKEND_VERSION}

all: deps build

deps:
	go get "github.com/devopsfaith/krakend/config/viper"
	go get "github.com/devopsfaith/krakend/proxy"
	go get "github.com/devopsfaith/krakend/router/mux"
	go get "github.com/devopsfaith/krakend/router/gorilla"
	go get "github.com/devopsfaith/krakend/router/negroni"
	go get "github.com/devopsfaith/krakend/logging/gologging"
	go get "github.com/urfave/negroni"
	go get "github.com/zbindenren/negroni-prometheus"
	go get "github.com/prometheus/client_golang/prometheus"
	go get "gopkg.in/unrolled/secure.v1"

build:
	go build -a -ldflags="-X github.com/devopsfaith/krakend/core.KrakendVersion=${KRAKEND_VERSION}" -o ${BIN_NAME}
	@echo "You can now use ./${BIN_NAME}"
