package config

// Parser reads a configuration file, parses it and returns the content as an init ServiceConfig struct
type Parser interface {
	Parse(configFile string) (ServiceConfig, error)
}

// ParserFunc type is an adapter to allow the use of ordinary functions as subscribers.
// If f is a function with the appropriate signature, ParserFunc(f) is a Parser that calls f.
type ParserFunc func(string) (ServiceConfig, error)

// Parse implements the Parser interface
func (f ParserFunc) Parse(configFile string) (ServiceConfig, error) { return f(configFile) }
