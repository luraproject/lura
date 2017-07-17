![Krakend logo](docs/images/krakend.png)

# KrakenD

[![Travis-CI](https://travis-ci.org/devopsfaith/krakend.svg?branch=master)](https://travis-ci.org/devopsfaith/krakend) [![Go Report Card](https://goreportcard.com/badge/github.com/devopsfaith/krakend)](https://goreportcard.com/report/github.com/devopsfaith/krakend) [![Coverage Status](https://coveralls.io/repos/github/devopsfaith/krakend/badge.svg?branch=master)](https://coveralls.io/github/devopsfaith/krakend?branch=master) [![GoDoc](https://godoc.org/github.com/devopsfaith/krakend?status.svg)](https://godoc.org/github.com/devopsfaith/krakend)

Ultra performant API Gateway with middlewares

## Motivation

Consumers of REST API content (specially in microservices) often query backend services that weren't coded for the UI implementation. This is of course a good practice, but the UI consumers need to do implementations that suffer a lot of complexity and burden with the sizes of their microservices responses.

KrakenD is an **API Gateway** builder and proxy generator that sits between the client and all the source servers, adding a new layer that removes all the complexity to the clients, providing them only the information that the UI needs. KrakenD acts as an **aggregator** of many sources into single endpoints and allows you to group, wrap, transform and shrink responses. Additionally it supports a myriad of middelwares and plugins that allow you to extend the functionality, such as adding Oauth authorization or security layers.

KrakenD not only supports HTTP(S), but because it is a set of generic libraries you can build all type of API Gateways and proxies, including for instance, a RPC gateway.

### Practical Example

Fred Calamari is a mobile developer that needs to construct a single front page that requires data from several calls to their backend services, e.g:

    1) api.store.server/products
    2) api.store.server/marketing-promos
    3) api.users.server/users/{id_user}
    4) api.users.server/shopping-cart/{id_user}

The screen is very simple and _only_ needs to retrieve data from 4 different sources, wait for the round trip and then pick only a few fields of the response. Instead of thing these calls, the mobile could call a single endpoint to KrakenD:

    1) krakend.server/frontpage/{id_user}

And this is how it would look like:

![Gateway](docs/images/krakend-gateway.png)

The difference in size in this example would be because KrakenD server would have removed unneeded attributes from the responses.

## What's in this repository?
The source code on which the [KrakenD](http://www.krakend.io) service core is built on. It is designed to work with your own middleware and extend the functionality by using small, independent, reusable components following the Unix philosophy.

**This repository is only for those who want to build from source a Krakend service** or for those who will reuse any of the components in another application.

If you just want to use the server, please [download the binary for your architecture](http://www.krakend.io/download).


## Library Usage
Krakend is presented as a **go library** that you can include in your own go application to build a powerful proxy or API gateway. In order to get you started several examples of implementations are included in the `examples` folder.

Of course you will need [go installed](https://golang.org/doc/install) in your system to compile the code.
There is a `Makefile` in every example that will download library dependencies and compile a binary for you to test. Just run:

    $ cd examples/gin
    $ make

Or, if you want to build all the examples, from the root of the project

    $ make

For the lazy, a ready to use example:

    package main

    import (
        "flag"
        "log"
        "os"

        "github.com/devopsfaith/krakend/config/viper"
        "github.com/devopsfaith/krakend/logging/gologging"
        "github.com/devopsfaith/krakend/proxy"
        "github.com/devopsfaith/krakend/router/gin"
    )

    func main() {
        port := flag.Int("p", 0, "Port of the service")
        logLevel := flag.String("l", "ERROR", "Logging level")
        debug := flag.Bool("d", false, "Enable the debug")
        configFile := flag.String("c", "/etc/krakend/configuration.json", "Path to the configuration filename")
        flag.Parse()

        parser := viper.New()
        serviceConfig, err := parser.Parse(*configFile)
        if err != nil {
            log.Fatal("ERROR:", err.Error())
        }
        serviceConfig.Debug = serviceConfig.Debug || *debug
        if *port != 0 {
            serviceConfig.Port = *port
        }

        logger := gologging.NewLogger(*logLevel, os.Stdout, "[KRAKEND]")

        routerFactory := gin.DefaultFactory(proxy.DefaultFactory(logger), logger)

        routerFactory.New().Run(serviceConfig)
    }

Visit the [framework overview](/docs/OVERVIEW.md) for more details about the components of the KrakenD.

### Examples

1. [gin router](/examples/gin/README.md)
2. [mux router](/examples/mux/README.md)
3. [gorilla router](/examples/gorilla/README.md)
4. [negroni middlewares](/examples/negroni/README.md)
5. [dns srv service discovery](/examples/dns/README.md)
6. [rss backends](/examples/rss/README.md)
7. [jwt middlewares](/examples/jwt/README.md)
8. [httpcache based proxies](/examples/httpcache/README.md)
9. [etcd service discovery](/examples/httpcache/README.md)

## Configuration file

[KrakenD config file](/docs/CONFIG.md)

## Benchmarks

Check out the [benchmark results](/docs/BENCHMARKS.md) of several KrakenD components

## Contributing
We are always happy to receive contributions. If you have questions, suggestions, bugs please open an issue.
If you want to submit the code, create the issue and send us a pull request for review.

Enjoy the KrakenD!
