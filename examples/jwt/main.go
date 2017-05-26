package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aviddiviner/gin-limit"
	jwt_lib "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/contrib/jwt"
	"github.com/gin-gonic/contrib/secure"
	"github.com/gin-gonic/gin"
	"gopkg.in/gin-contrib/cors.v1"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/config/viper"
	"github.com/devopsfaith/krakend/logging"
	"github.com/devopsfaith/krakend/logging/gologging"
	"github.com/devopsfaith/krakend/proxy"
	krakendgin "github.com/devopsfaith/krakend/router/gin"
)

func main() {
	port := flag.Int("p", 0, "Port of the service")
	logLevel := flag.String("l", "ERROR", "Logging level")
	allowedHosts := flag.String("hosts", "127.0.0.1:8080,example.com,ssl.example.com", "Comma-separated list of allowed hosts")
	allowedOrigins := flag.String("cors-origins", "http://example.com", "Comma-separated list of CORS allowed origins")
	allowedMethods := flag.String("cors-methods", "HEAD,GET,POST,PUT,PATCH,DELETE", "Comma-separated list of CORS allowed methods")
	allowedHeaders := flag.String("cors-headers", "Origin,Authorization,Content-Type", "Comma-separated list of CORS allowed headers")
	exposedHeaders := flag.String("cors-headers-exposed", "Content-Length", "Comma-separated list of CORS exposed headers")
	corsTTL := flag.Duration("cors-ttl", 12*time.Hour, "Max age for the CORS layer")
	jwtSecret := flag.String("jwt-secret", "KrakenDrulez123.4567890!", "Secret for signing jwt")
	jwtIssuer := flag.String("jwt-issuer", "http://example.com/", "Issuer for the jwt")
	jwtPort := flag.Int("jwt-port", 8090, "Port for the jwt generator api")
	jwsTTL := flag.Duration("jwt-ttl", 1*time.Hour, "Expiration for the JWT")
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
		log.Println("ERROR:", err.Error())
		return
	}

	// run the dummy jwt generator http service in a dedicated goroutine
	go runJWTGeneratorHTTPService("/token", *jwtSecret, *jwtIssuer, *jwsTTL, *jwtPort)

	routerFactory := krakendgin.NewFactory(krakendgin.Config{
		Engine:         gin.Default(),
		ProxyFactory:   customProxyFactory{logger, proxy.DefaultFactory(logger)},
		Logger:         logger,
		HandlerFactory: krakendgin.EndpointHandler,
		Middlewares: []gin.HandlerFunc{
			secure.Secure(secure.Options{
				AllowedHosts:          strings.Split(*allowedHosts, ","),
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
			cors.New(cors.Config{
				AllowOrigins:     strings.Split(*allowedOrigins, ","),
				AllowMethods:     strings.Split(*allowedMethods, ","),
				AllowHeaders:     strings.Split(*allowedHeaders, ","),
				ExposeHeaders:    strings.Split(*exposedHeaders, ","),
				AllowCredentials: true,
				MaxAge:           *corsTTL,
			}),
			jwt.Auth(*jwtSecret),
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

// runJWTGeneratorHTTPService sets up and runs a dummy http service with a single endpoint ready to create signed JWT
// issued for the received resource id
func runJWTGeneratorHTTPService(resource, jwtSecret, jwtIssuer string, jwsTTL time.Duration, jwtPort int) {
	engine := gin.Default()
	engine.GET(fmt.Sprintf("%s/:id", resource), func(c *gin.Context) {
		token := jwt_lib.New(jwt_lib.GetSigningMethod("HS256"))
		token.Claims = jwt_lib.MapClaims{
			"Id":  c.Param("id"),
			"iss": jwtIssuer,
			"exp": time.Now().Add(jwsTTL).Unix(),
		}
		tokenString, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			c.JSON(500, gin.H{"message": "Could not generate token"})
		}
		c.JSON(200, gin.H{"token": tokenString})
	})
	log.Fatal(engine.Run(fmt.Sprintf(":%d", jwtPort)))
}
