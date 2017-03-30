package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/aviddiviner/gin-limit"
	"github.com/gin-gonic/contrib/cache"
	"github.com/gin-gonic/contrib/secure"
	"github.com/gin-gonic/gin"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/config/viper"
	"github.com/devopsfaith/krakend/logging"
	"github.com/devopsfaith/krakend/logging/gologging"
	"github.com/devopsfaith/krakend/proxy"
	krakendgin "github.com/devopsfaith/krakend/router/gin"
	"github.com/devopsfaith/krakend/sd/dnssrv"
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

	store := cache.NewInMemoryStore(time.Minute)

	mws := []gin.HandlerFunc{
		secure.Secure(secure.Options{
			AllowedHosts:          []string{"127.0.0.1:8080", "example.com", "ssl.example.com"},
			SSLRedirect:           false,
			SSLHost:               "ssl.example.com",
			SSLProxyHeaders:       map[string]string{"X-Forwarded-Proto": "https"},
			STSSeconds:            315360000,
			STSIncludeSubdomains:  true,
			FrameDeny:             true,
			ContentTypeNosniff:    true,
			BrowserXssFilter:      true,
			ContentSecurityPolicy: "default-src 'self'",
		}),
		limit.MaxAllowed(20),
	}

	// routerFactory := krakendgin.DefaultFactory(proxy.DefaultFactory(logger), logger)

	routerFactory := krakendgin.NewFactory(krakendgin.Config{
		Engine:       gin.Default(),
		ProxyFactory: customProxyFactory{logger, proxy.DefaultFactoryWithSubscriber(logger, dnssrv.SubscriberFactory)},
		Middlewares:  mws,
		Logger:       logger,
		HandlerFactory: func(configuration *config.EndpointConfig, proxy proxy.Proxy) gin.HandlerFunc {
			return cache.CachePage(store, configuration.CacheTTL, krakendgin.EndpointHandler(configuration, proxy))
		},
	})

	routerFactory.New().Run(serviceConfig)
}

// customProxyFactory adds a logging middleware wrapping the internal factory
type customProxyFactory struct {
	logger  logging.Logger
	factory proxy.Factory
}

// New implements the Factory interface
func (cf customProxyFactory) New(cfg *config.EndpointConfig) (p proxy.Proxy, err error) {
	p, err = cf.factory.New(cfg)
	if err == nil {
		p = proxy.NewLoggingMiddleware(cf.logger, cfg.Endpoint)(p)
	}
	return
}
