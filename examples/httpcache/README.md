Krakend httpcache example
====

Simple example of how to extend the basic `proxy.BackendFactory` in order to implement all kinds of features, like client-side http cache so the KrakenD is able to cache the responses from the backends and re-use them in other compositions.

## Build

Go 1.8 is a requirement

	$ make

## Run

Running it as a common executable, logs are send to the stdOut and some options are available at the CLI

	$ ./krakend_httpcache_example
	Usage of ./krakend_httpcache_example:
	  -c string
	    	Path to the configuration filename (default "/etc/krakend/configuration.json")
	  -d	Enable the debug
	  -p int
	    	Port of the service
