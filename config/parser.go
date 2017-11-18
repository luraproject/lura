package config

// Parser reads a configuration file, parses it and returns the content as an init ServiceConfig struct
type Parser interface {
	Parse(configFile string) (ServiceConfig, error)
}

type ParserFunc func(string) (ServiceConfig, error)

func (f ParserFunc) Parse(configFile string) (ServiceConfig, error) { return f(configFile) }
