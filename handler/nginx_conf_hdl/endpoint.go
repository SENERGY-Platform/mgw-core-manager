package nginx_conf_hdl

import (
	"encoding/json"
	"fmt"
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
	varName      string
}

func newEndpoint(dID, host string, port *int, intPath, extPath string) endpoint {
	return endpoint{
		DeploymentID: dID,
		Host:         host,
		Port:         port,
		IntPath:      intPath,
		ExtPath:      extPath,
		varName:      util.GenHash(dID, extPath),
	}
}

func parseEndpoint(s string) (endpoint, error) {
	s, _ = strings.CutPrefix(s, "#")
	var e endpoint
	err := json.Unmarshal([]byte(s), &e)
	if err != nil {
		return endpoint{}, err
	}
	e.varName = util.GenHash(e.DeploymentID, e.ExtPath)
	return e, nil
}

func (e endpoint) GenComment() (string, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return "#" + string(b), err
}

func (e endpoint) GenProxyPassValue(template string) string {
	template = strings.Replace(template, varPlaceholder, "$"+e.varName, -1)
	var port string
	if e.Port != nil && *e.Port != 80 {
		port = ":" + strconv.FormatInt(int64(*e.Port), 10)
	}
	template = strings.Replace(template, portPlaceholder, port, -1)
	return strings.Replace(template, pathPlaceholder, e.IntPath, -1)
}

func (e endpoint) GenLocationValue(template string) string {
	template = strings.Replace(template, depIDPlaceholder, e.DeploymentID, -1)
	return strings.Replace(template, pathPlaceholder, e.ExtPath, -1)
}

func (e endpoint) GenSetValue() string {
	return fmt.Sprintf("$%s %s", e.varName, e.Host)
}
