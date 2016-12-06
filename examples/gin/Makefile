.PHONY: all deps build

# This Makefile is a simple example that demonstrates usual steps to build a binary that can be run in the same
# architecture that was compiled in. The "ldflags" in the build assure that any needed dependency is included in the
# binary and no external dependencies are needed to run the service.

KRAKEND_VERSION=$(shell git describe --always --long --dirty --tags)
BIN_NAME=krakend_gin_example_${KRAKEND_VERSION}

all: deps build

deps:
	go get "github.com/devopsfaith/krakend/config/viper"
	go get "github.com/devopsfaith/krakend/proxy"
	go get "github.com/devopsfaith/krakend/router/gin"
	go get "github.com/devopsfaith/krakend/logging/gologging"
	go get "github.com/gin-gonic/contrib/cache"
	go get "github.com/gin-gonic/contrib/secure"
	go get "github.com/aviddiviner/gin-limit"

build:
	go build -a -ldflags="-X github.com/devopsfaith/krakend/core.KrakendVersion=${KRAKEND_VERSION}" -o ${BIN_NAME}
	@echo "You can now use ./${BIN_NAME}"

docker_build:
	docker run --name krakend_gin -it -e "GOPATH=/go" -p 8080:8080 \
	-v "${PWD}/main.go:/go/src/github.com/devopsfaith/krakend/examples/gin/main.go" \
	-v "${PWD}/Makefile:/go/src/github.com/devopsfaith/krakend/examples/gin/Makefile" \
	-w /go/src/github.com/devopsfaith/krakend/examples/gin \
	golang:1.7.4 make all
