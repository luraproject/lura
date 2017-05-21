Krakend JWT example
====

## Build

Go 1.8 is a requirement

	$ make

## Run

Running it as a common executable, logs are send to the stdOut and some options are available at the CLI.

Notice this example starts a dedicated service just for issuing signed JWT.

	$ ./krakend_jwt_example
	Usage of ./krakend_jwt_example_92cc18c:
	  -c string
	    	Path to the configuration filename (default "/etc/krakend/configuration.json")
	  -cors-headers string
	    	Comma-separated list of CORS allowed headers (default "Origin,Authorization,Content-Type")
	  -cors-headers-exposed string
	    	Comma-separated list of CORS exposed headers (default "Content-Length")
	  -cors-methods string
	    	Comma-separated list of CORS allowed methods (default "HEAD,GET,POST,PUT,PATCH,DELETE")
	  -cors-origins string
	    	Comma-separated list of CORS allowed origins (default "http://example.com")
	  -cors-ttl duration
	    	Max age for the CORS layer (default 12h0m0s)
	  -d	Enable the debug
	  -hosts string
	    	Comma-separated list of allowed hosts (default "127.0.0.1:8080,example.com,ssl.example.com")
	  -jwt-issuer string
	    	Issuer for the jwt (default "http://example.com/")
	  -jwt-port int
	    	Port for the jwt generator api (default 8090)
	  -jwt-secret string
	    	Secret for signing jwt (default "KrakenDrulez123.4567890!")
	  -jwt-ttl duration
	    	Expiration for the JWT (default 1h0m0s)
	  -l string
	    	Logging level (default "ERROR")
	  -p int
	    	Port of the service
