package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/devopsfaith/krakend/config/viper"
	"github.com/devopsfaith/krakend/logging/gologging"
	"github.com/devopsfaith/krakend/proxy"
	"github.com/devopsfaith/krakend/proxy/martian"
	krakendgin "github.com/devopsfaith/krakend/router/gin"
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

	logger, err := gologging.NewLogger(*logLevel, os.Stdout, "[KRAKEND]")
	if err != nil {
		log.Fatal("ERROR:", err.Error())
	}

	logger.Debug("config:", serviceConfig)

	ctx, cancel := context.WithCancel(context.Background())

	routerFactory := krakendgin.NewFactory(krakendgin.Config{
		Engine:         gin.Default(),
		Middlewares:    []gin.HandlerFunc{},
		Logger:         logger,
		HandlerFactory: krakendgin.EndpointHandler,
		ProxyFactory:   proxy.NewDefaultFactory(martian.NewBackendFactory(logger, proxy.DefaultHTTPRequestExecutor(proxy.NewHTTPClient)), logger),
	})

	routerFactory.NewWithContext(ctx).Run(serviceConfig)

	cancel()
}
