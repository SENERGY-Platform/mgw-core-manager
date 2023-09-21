package nginx_conf_hdl

import (
	"encoding/json"
	"strings"
)

type comment struct {
	DeploymentID string `json:"d_id"`
	Host         string `json:"host"`
	Port         *int   `json:"port"`
	IntPath      string `json:"int_path"`
	ExtPath      string `json:"ext_path"`
}

func newComment(dID, host string, port *int, intPath, extPath string) comment {
	return comment{
		DeploymentID: dID,
		Host:         host,
		Port:         port,
		IntPath:      intPath,
		ExtPath:      extPath,
	}
}

func parseComment(s string) (comment, error) {
	s, _ = strings.CutPrefix(s, "#")
	var c comment
	err := json.Unmarshal([]byte(s), &c)
	if err != nil {
		return comment{}, err
	}
	return c, nil
}

func (c comment) ToString() (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return "#" + string(b), err
}
