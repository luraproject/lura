Krakend Martian example
====

## Build

Go 1.8 is a requirement

	$ make

## Run

Running it as a common executable, logs are send to the stdOut and some options are available at the CLI

	$ ./krakend_martian_example
	Usage of ./krakend_martian_example:
	  -c string
	    	Path to the configuration filename (default "/etc/krakend/configuration.json")
	  -d	Enable the debug
	  -l string
	    	Logging level (default "ERROR")
	  -p int
	    	Port of the service
