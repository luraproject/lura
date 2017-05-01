package core

import "fmt"

// KrakendHeaderName is the name of the custom KrakenD header
const KrakendHeaderName = "X-KRAKEND"

// KrakendVersion is the version of the build
var KrakendVersion = "undefined"

// KrakendHeaderValue is the value of the custom KrakenD header
var KrakendHeaderValue = fmt.Sprintf("Version %s", KrakendVersion)

// KrakendUserAgent is the value of the user agent header sent to the backends
var KrakendUserAgent = fmt.Sprintf("KrakenD Version %s", KrakendVersion)
