package nginx_conf_hdl

import "github.com/SENERGY-Platform/mgw-core-manager/lib/model"

const (
	locationDirective  = "location"
	proxyPassDirective = "proxy_pass"
	allowDirective     = "allow"
	denyDirective      = "deny"
	setDirective       = "set"
)

const (
	varPlaceholder   = "{var}"
	portPlaceholder  = "{port}"
	pathPlaceholder  = "{path}"
	depIDPlaceholder = "{did}"
)

const (
	locationTmpl = iota
	proxyPassTmpl
)

const (
	StandardLocationTmpl = iota
	StandardProxyPassTmpl
	DefaultGuiLocationTmpl
	DefaultGuiProxyPassTmpl
	NamedLocationTmpl
	NamedProxyPassTmpl
)

var endpointTypeMap = map[int]map[int]int{
	model.StandardEndpoint: {
		locationTmpl:  StandardLocationTmpl,
		proxyPassTmpl: StandardProxyPassTmpl,
	},
	model.DefaultGuiEndpoint: {
		locationTmpl:  DefaultGuiLocationTmpl,
		proxyPassTmpl: DefaultGuiProxyPassTmpl,
	},
	model.NamedEndpoint: {
		locationTmpl:  NamedLocationTmpl,
		proxyPassTmpl: NamedProxyPassTmpl,
	},
}
