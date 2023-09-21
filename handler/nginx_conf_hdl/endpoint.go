package nginx_conf_hdl

import (
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"strconv"
)

type endpoint struct {
	Host    string
	Port    *int
	Path    string // internal path
	VarName string // hash(DeploymentID, Host, Path)
}

func newEndpoint(dID, host string, port *int, path string) *endpoint {
	var hash string
	if port != nil {
		hash = util.GenHash(dID, host, strconv.FormatInt(int64(*port), 10), path)
	} else {
		util.GenHash(dID, host, path)
	}
	return &endpoint{
		Host:    host,
		Port:    port,
		Path:    path,
		VarName: hash,
	}
}

func (e *endpoint) FullPath() string {
	s := e.VarName
	if e.Port != nil {
		s += ":" + strconv.FormatInt(int64(*e.Port), 10)
	}
	return s + e.Path
}
