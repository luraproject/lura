package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	gorilla "github.com/gorilla/mux"
	"gopkg.in/unrolled/secure.v1"

	"github.com/devopsfaith/krakend/config/viper"
	"github.com/devopsfaith/krakend/logging/gologging"
	"github.com/devopsfaith/krakend/proxy"
	"github.com/devopsfaith/krakend/router/mux"
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

	secureMiddleware := secure.New(secure.Options{
		AllowedHosts:          []string{"127.0.0.1:8080", "example.com", "ssl.example.com"},
		SSLRedirect:           false,
		SSLHost:               "ssl.example.com",
		SSLProxyHeaders:       map[string]string{"X-Forwarded-Proto": "https"},
		STSSeconds:            315360000,
		STSIncludeSubdomains:  true,
		STSPreload:            true,
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self'",
	})

	routerFactory := mux.NewFactory(mux.Config{
		Engine:         gorillaEngine{gorilla.NewRouter()},
		ProxyFactory:   proxy.DefaultFactory(logger),
		Middlewares:    []mux.HandlerMiddleware{secureMiddleware},
		Logger:         logger,
		HandlerFactory: mux.EndpointHandler,
		DebugPattern:   "/__debug/{params}",
	})

	routerFactory.New().Run(serviceConfig)
}

type gorillaEngine struct {
	r *gorilla.Router
}

// Handle implements the mux.Engine interface from the krakend router package
func (g gorillaEngine) Handle(pattern string, handler http.Handler) {
	g.r.Handle(pattern, handler)
}

// ServeHTTP implements the http:Handler interface from the stdlib
func (g gorillaEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.r.ServeHTTP(w, r)
}
