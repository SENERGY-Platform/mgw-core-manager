package nginx_conf_hdl

import (
	"encoding/json"
	"fmt"
	"github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"strconv"
	"strings"
)

type endpoint struct {
	ID string `json:"id"`
	model.Endpoint
}

func newEndpoint(e model.Endpoint) endpoint {
	return endpoint{
		Endpoint: e,
		ID:       genID(e),
	}
}

func parseEndpoint(s string) (endpoint, error) {
	s, _ = strings.CutPrefix(s, "#")
	var e endpoint
	err := json.Unmarshal([]byte(s), &e)
	if err != nil {
		return endpoint{}, err
	}
	e.ID = genID(e.Endpoint)
	return e, nil
}

func (e endpoint) GenComment() (string, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return "#" + string(b), err
}

func (e endpoint) GenProxyPassValue(templates map[int]string) string {
	template := templates[endpointTypeMap[e.Type][proxyPassTmpl]]
	template = strings.Replace(template, varPlaceholder, "$"+e.ID, -1)
	var port string
	if e.Port != nil && *e.Port != 80 {
		port = ":" + strconv.FormatInt(int64(*e.Port), 10)
	}
	template = strings.Replace(template, portPlaceholder, port, -1)
	return strings.Replace(template, pathPlaceholder, e.IntPath, -1)
}

func (e endpoint) GenLocationValue(templates map[int]string) string {
	template := templates[endpointTypeMap[e.Type][locationTmpl]]
	template = strings.Replace(template, depIDPlaceholder, e.DeploymentID, -1)
	return strings.Replace(template, pathPlaceholder, e.ExtPath, -1)
}

func (e endpoint) GenSetValue() string {
	return fmt.Sprintf("$%s %s", e.ID, e.Host)
}

func genID(e model.Endpoint) string {
	return util.GenHash(strconv.FormatInt(int64(e.Type), 10), e.DeploymentID, e.ExtPath)
}
