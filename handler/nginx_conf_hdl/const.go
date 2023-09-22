package nginx_conf_hdl

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
