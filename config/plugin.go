package config

// Plugin contains the config required by the plugin module
type Plugin struct {
	Folder  string `mapstructure:"folder"`
	Pattern string `mapstructure:"pattern"`
}

// GetFolder returns the path of the plugins folder
func (p *Plugin) GetFolder() string {
	return p.Folder
}

// GetPattern returns the defined pattern to filter the contents of the plugin folder
func (p *Plugin) GetPattern() string {
	return p.Pattern
}
