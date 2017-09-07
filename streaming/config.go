package streaming

import "github.com/devopsfaith/krakend/config"

// StreamNamespace is the default namespace where to find the config parser,StreamConfigGetter, for streaming
// operations
const StreamNamespace = "github.com/devopsfaith/krakend/streaming"

// StreamConfigGetter is the ConfigGetter implementation for streaming Endpoints and Proxies
// it expects something like that in the endpoint definition
// "extra_config": {
//	"Forward": true
//	},
func StreamConfigGetter(extra config.ExtraConfig) interface{} {
	ok := extra["Forward"]
	return StreamExtraConfig{ok != nil}
}

// StreamExtraConfig is the expected type to be returned by StreamConfigGetter
type StreamExtraConfig struct {
	Forward bool
}

func init() {
	config.ConfigGetters[StreamNamespace] = StreamConfigGetter
}
