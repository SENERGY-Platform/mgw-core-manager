/*
 * Copyright 2023 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package nginx_hdl

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"strconv"
	"strings"
)

type endpoint struct {
	lib_model.Endpoint
	proxyPassVal string
	rewriteVal   string
	locationVal  string
	setVal       string
}

func newEndpoint(e lib_model.Endpoint, templates map[int]string) endpoint {
	locVal := genLocationValue(e.EndpointBase, e.Type, templates)
	if e.ID == "" {
		e.ID = util.GenHash(locVal)
	}
	return endpoint{
		Endpoint:     e,
		proxyPassVal: genProxyPassValue(e, templates),
		rewriteVal:   genRewriteValue(e, e.Type, templates),
		locationVal:  locVal,
		setVal:       genSetValue(e),
	}
}

func (e endpoint) GenComment() (string, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return "#" + base64.StdEncoding.EncodeToString(b), err
}

func (e endpoint) GetLocationValue() string {
	return e.locationVal
}

func (e endpoint) GetRewriteValue() string {
	return e.rewriteVal
}

func (e endpoint) GetProxyPassValue() string {
	return e.proxyPassVal
}

func (e endpoint) GetSetValue() string {
	return e.setVal
}

func genProxyPassValue(e lib_model.Endpoint, templates map[int]string) string {
	template := templates[endpointTypeMap[e.Type][proxyPassTmpl]]
	template = strings.Replace(template, varPlaceholder, "$v"+e.ID, -1)
	var port string
	if e.Port != nil && *e.Port != 80 {
		port = ":" + strconv.FormatInt(int64(*e.Port), 10)
	}
	template = strings.Replace(template, portPlaceholder, port, -1)
	return strings.Replace(template, pathPlaceholder, e.IntPath, -1)
}

func genLocationValue(eBase lib_model.EndpointBase, eType lib_model.EndpointType, templates map[int]string) string {
	template := templates[endpointTypeMap[eType][locationTmpl]]
	template = strings.Replace(template, refPlaceholder, eBase.Ref, -1)
	return strings.Replace(template, pathPlaceholder, eBase.ExtPath, -1)
}

func genSetValue(e lib_model.Endpoint) string {
	return fmt.Sprintf("$v%s %s", e.ID, e.Host)
}

func genRewriteValue(e lib_model.Endpoint, eType lib_model.EndpointType, templates map[int]string) string {
	template := templates[endpointTypeMap[eType][rewriteTmpl]]
	template = strings.Replace(template, refPlaceholder, e.Ref, -1)
	return strings.Replace(template, pathPlaceholder, e.ExtPath, -1)
}
