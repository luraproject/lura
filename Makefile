.PHONY: all deps test build benchmark coveralls build_gin_example build_dns_example build_mux_example build_gorilla_example build_negroni_example build_httpcache_example build_rss_example build_jwt_example build_etcd_example

PACKAGES = $(shell go list ./... | grep -v /examples/)

all: deps test build

deps:
	go get -u github.com/gin-gonic/gin
	go get -u github.com/spf13/viper
	go get -u github.com/op/go-logging
	go get -u github.com/gorilla/mux
	go get -u github.com/urfave/negroni
	go get -u github.com/clbanning/mxj/x2j
	go get -u github.com/mmcdole/gofeed
	go get -u github.com/coreos/etcd/client

test:
	go fmt ./...
	go test -cover $(PACKAGES)
	go vet ./...

benchmark:
	go test -bench=. -benchtime=3s $(PACKAGES)

build: build_gin_example build_dns_example build_mux_example build_gorilla_example build_negroni_example build_httpcache_example build_rss_example build_jwt_example build_etcd_example

build_gin_example:
	cd examples/gin/ && make && cd ../.. && cp examples/gin/krakend_gin_example* .

build_dns_example:
	cd examples/dns/ && make && cd ../.. && cp examples/dns/krakend_dns_example* .

build_mux_example:
	cd examples/mux/ && make && cd ../.. && cp examples/mux/krakend_mux_example* .

build_gorilla_example:
	cd examples/gorilla/ && make && cd ../.. && cp examples/gorilla/krakend_gorilla_example* .

build_negroni_example:
	cd examples/negroni/ && make && cd ../.. && cp examples/negroni/krakend_negroni_example* .

build_httpcache_example:
	cd examples/httpcache/ && make && cd ../.. && cp examples/httpcache/krakend_httpcache_example* .

build_rss_example:
	cd examples/rss/ && make && cd ../.. && cp examples/rss/krakend_rss_example* .

build_jwt_example:
	cd examples/jwt/ && make && cd ../.. && cp examples/jwt/krakend_jwt_example* .

build_etcd_example:
	cd examples/etcd/ && make && cd ../.. && cp examples/etcd/krakend_etcd_example* .

coveralls: all
	go get github.com/mattn/goveralls
	sh coverage.sh --coveralls
