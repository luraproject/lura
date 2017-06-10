Krakend ETCD example
====

## Build

Go 1.8 is a requirement

	$ make

## Run

Running it as a common executable, logs are send to the stdOut and some options are available at the CLI

	$ ./krakend_etcd_example
	Usage of ./krakend_etcd_example:
	  -c string
	    	Path to the configuration filename (default "/etc/krakend/configuration.json")
	  -d	Enable the debug
	  -etcd string
	    	Comma-separated list of etcd servers (with port and schema) (default "http://192.168.99.100:4001")
	  -l string
	    	Logging level (default "ERROR")
	  -p int
	    	Port of the service
