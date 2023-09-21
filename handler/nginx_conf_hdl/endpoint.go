package nginx_conf_hdl

import (
	"encoding/json"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"strconv"
	"strings"
)

type endpoint struct {
	DeploymentID string `json:"d_id"`
	Host         string `json:"host"`
	Port         *int   `json:"port"`
	IntPath      string `json:"int_path"`
	ExtPath      string `json:"ext_path"`
	VarName      string `json:"-"`
}

func newEndpoint(dID, host string, port *int, intPath, extPath string) endpoint {
	return endpoint{
		DeploymentID: dID,
		Host:         host,
		Port:         port,
		IntPath:      intPath,
		ExtPath:      extPath,
		VarName:      util.GenHash(dID, extPath),
	}
}

func parseEndpoint(s string) (endpoint, error) {
	s, _ = strings.CutPrefix(s, "#")
	var e endpoint
	err := json.Unmarshal([]byte(s), &e)
	if err != nil {
		return endpoint{}, err
	}
	e.VarName = util.GenHash(e.DeploymentID, e.ExtPath)
	return e, nil
}

func (e endpoint) ToString() (string, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return "#" + string(b), err
}

func (e endpoint) FullIntPath() string {
	s := "$" + e.VarName
	if e.Port != nil {
		s += ":" + strconv.FormatInt(int64(*e.Port), 10)
	}
	return s + e.IntPath
}
