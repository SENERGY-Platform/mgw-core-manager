package nginx_conf_hdl

import "github.com/SENERGY-Platform/mgw-core-manager/lib/model"

const (
	locationDirective  = "location"
	proxyPassDirective = "proxy_pass"
	allowDirective     = "allow"
	denyDirective      = "deny"
	setDirective       = "set"
	serverDirective    = "server"
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
}
