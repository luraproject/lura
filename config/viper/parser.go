// Package viper defines a config parser implementation based on the viper pkg
package viper

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/devopsfaith/krakend/config"
)

// New creates a new parser using the viper library
func New() config.Parser {
	return parser{viper.New()}
}

type parser struct {
	viper *viper.Viper
}

// Parser implements the Parse interface
func (p parser) Parse(configFile string) (config.ServiceConfig, error) {
	p.viper.SetConfigFile(configFile)
	p.viper.AutomaticEnv()
	var cfg config.ServiceConfig
	if err := p.viper.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("Fatal error config file: %s \n", err.Error())
	}
	if err := p.viper.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("Fatal error unmarshalling config file: %s \n", err.Error())
	}
	err := cfg.Init()

	return cfg, err
}
