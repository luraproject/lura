package core

import "fmt"

const KrakendHeaderName = "X-KRAKEND"

var (
	KrakendVersion     = "undefined"
	KrakendHeaderValue = fmt.Sprintf("Version %s", KrakendVersion)
	KrakendUserAgent   = fmt.Sprintf("KrakenD Version %s", KrakendVersion)
)
