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
	CommentDeploymentIDKey = "did"
	CommentExtPathKey      = "ext_path"
	CommentIntPathKey      = "int_path"
	CommentHostKey         = "host"
	CommentPortKey         = "port"
	CommentDelimiter       = " "
	CommentItemDelimiter   = "="
)
