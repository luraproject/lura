<img src="https://luraproject.org/images/lura-logo-header.svg" width="300" />

# The Lura Project framework

[![Go Report Card](https://goreportcard.com/badge/github.com/luraproject/lura)](https://goreportcard.com/report/github.com/luraproject/lura)
[![GoDoc](https://godoc.org/github.com/luraproject/lura?status.svg)](https://godoc.org/github.com/luraproject/lura)
![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/3151/badge)
![Docker Pulls](https://img.shields.io/docker/pulls/devopsfaith/krakend.svg)
[![Slack Widget](https://img.shields.io/badge/join-us%20on%20slack-gray.svg?longCache=true&logo=slack&colorB=red)](https://gophers.slack.com/messages/lura)


An open framework to assemble ultra performance API Gateways with middlewares; formerly known as _KrakenD framework_, and core service of the [KrakenD API Gateway](http://www.krakend.io).

## Motivation

Consumers of REST API content (specially in microservices) often query backend services that weren't coded for the UI implementation. This is of course a good practice, but the UI consumers need to do implementations that suffer a lot of complexity and burden with the sizes of their microservices responses.

Lura is an **API Gateway** builder and proxy generator that sits between the client and all the source servers, adding a new layer that removes all the complexity to the clients, providing them only the information that the UI needs. Lura acts as an **aggregator** of many sources into single endpoints and allows you to group, wrap, transform and shrink responses. Additionally it supports a myriad of middlewares and plugins that allow you to extend the functionality, such as adding Oauth authorization or security layers.

Lura not only supports HTTP(S), but because it is a set of generic libraries you can build all type of API Gateways and proxies, including for instance, an RPC gateway.

### Practical Example

A mobile developer needs to construct a single front page that requires data from 4 different calls to their backend services, e.g:

    1) api.store.server/products
    2) api.store.server/marketing-promos
    3) api.users.server/users/{id_user}
    4) api.users.server/shopping-cart/{id_user}

The screen is very simple, and the mobile client _only_ needs to retrieve data from 4 different sources, wait for the round trip and then hand pick only a few fields from the response.

What if the mobile could call a single endpoint?

    1) lura.server/frontpage/{id_user}

That's something Lura can do for you. And this is how it would look like:

![Gateway](https://luraproject.org/images/docs/lura-gateway.png)

Lura would merge all the data and return only the fields you need (the difference in size in the graph).

Visit the [Lura Project website](https://luraproject.org) for more information.

## What's in this repository?

The source code for the [Lura project](https://luraproject.org) framework. It is designed to work with your own middleware and extend the functionality by using small, independent, reusable components following the Unix philosophy.

Use this repository if you want to **build from source your API Gateway** or if you want to **reuse the components in another application**.

If you need a fully functional API Gateway you can [download the KrakenD binary for your architecture](http://www.krakend.io/download) or [build it yourself](https://github.com/krakendio/krakend-ce).


## Library Usage
The Lura project is presented as a **Go library** that you can include in your own Go application to build a powerful proxy or API gateway. In order to get you started several examples of implementations are included in the `examples` folder.

Of course, you will need [Go installed](https://golang.org/doc/install) in your system to compile the code.

A ready to use example:

```go
    package main

    import (
        "flag"
        "log"
        "os"

        "github.com/luraproject/lura/config"
        "github.com/luraproject/lura/logging"
        "github.com/luraproject/lura/proxy"
        "github.com/luraproject/lura/router/gin"
    )

    func main() {
        port := flag.Int("p", 0, "Port of the service")
        logLevel := flag.String("l", "ERROR", "Logging level")
        debug := flag.Bool("d", false, "Enable the debug")
        configFile := flag.String("c", "/etc/lura/configuration.json", "Path to the configuration filename")
        flag.Parse()

        parser := config.NewParser()
        serviceConfig, err := parser.Parse(*configFile)
        if err != nil {
            log.Fatal("ERROR:", err.Error())
        }
        serviceConfig.Debug = serviceConfig.Debug || *debug
        if *port != 0 {
            serviceConfig.Port = *port
        }

        logger, _ := logging.NewLogger(*logLevel, os.Stdout, "[LURA]")

        routerFactory := gin.DefaultFactory(proxy.DefaultFactory(logger), logger)

        routerFactory.New().Run(serviceConfig)
    }
```

Visit the [framework overview](/docs/OVERVIEW.md) for more details about the components of the Lura project.

## Configuration file

[Lura config file](/docs/CONFIG.md)

## Benchmarks

Check out the [benchmark results](/docs/BENCHMARKS.md) of several Lura components

## Contributing
We are always happy to receive contributions. If you have questions, suggestions, bugs please open an issue.
If you want to submit the code, create the issue and send us a pull request for review.

Read [CONTRIBUTING.md](/CONTRIBUTING.md) for more information.


## Want more?
- Follow us on Twitter: [@luraproject](https://twitter.com/luraproject)
- Visit our [Slack channel](https://gophers.slack.com/messages/lura)
- **Read the [documentation](/docs/OVERVIEW.md)**

Enjoy Lura!
