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
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"github.com/tufanbarisyildirim/gonginx/config"
	"github.com/tufanbarisyildirim/gonginx/dumper"
	"github.com/tufanbarisyildirim/gonginx/parser"
	"io"
	"os"
	"strings"
	"sync"
)

type Handler struct {
	ctrHdl    ContainerHandler
	confPath  string
	templates map[int]string
	endpoints map[string]endpoint
	m         sync.RWMutex
}

func New(containerHandler ContainerHandler, confPath string, templates map[int]string) *Handler {
	return &Handler{
		ctrHdl:    containerHandler,
		confPath:  confPath,
		templates: templates,
	}
}

func (h *Handler) Init() error {
	_, err := os.Stat(h.confPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		_, err = os.Create(h.confPath)
		if err != nil {
			return err
		}
	}
	p, err := parser.NewParser(h.confPath)
	if err != nil {
		return err
	}
	conf, err := p.Parse()
	if err != nil {
		return err
	}
	h.endpoints, err = getEndpoints(conf.GetDirectives(), h.templates)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) List(ctx context.Context, filter lib_model.EndpointFilter) (map[string]lib_model.Endpoint, error) {
	h.m.RLock()
	defer h.m.RUnlock()
	filtered := filterEndpoints(h.endpoints, filter)
	endpoints := make(map[string]lib_model.Endpoint)
	for id, e := range filtered {
		if ctx.Err() != nil {
			return nil, lib_model.NewInternalError(ctx.Err())
		}
		e.Location = e.GetLocationValue()
		endpoints[id] = e.Endpoint
	}
	return endpoints, nil
}

func (h *Handler) Get(_ context.Context, id string) (lib_model.Endpoint, error) {
	h.m.RLock()
	defer h.m.RUnlock()
	e, ok := h.endpoints[id]
	if !ok {
		return lib_model.Endpoint{}, lib_model.NewNotFoundError(errors.New("endpoint not found"))
	}
	e.Location = e.GetLocationValue()
	return e.Endpoint, nil
}

func (h *Handler) Set(ctx context.Context, eBase lib_model.EndpointBase) error {
	h.m.Lock()
	defer h.m.Unlock()
	if err := checkIntPath(eBase.IntPath); err != nil {
		return err
	}
	if err := checkExtPath(eBase.ExtPath); err != nil {
		return err
	}
	endpointsCopy := make(map[string]endpoint)
	for id, e := range h.endpoints {
		endpointsCopy[id] = e
	}
	ept := newEndpoint(lib_model.Endpoint{Type: lib_model.StandardEndpoint, EndpointBase: eBase}, h.templates)
	if ept2, ok := endpointsCopy[ept.ID]; ok {
		util.Logger.Warningf("endpoint '%+v' replaced by '%+v'", ept2.EndpointBase, ept.EndpointBase)
	}
	endpointsCopy[ept.ID] = ept
	return h.update(ctx, endpointsCopy)
}

func (h *Handler) SetList(ctx context.Context, eBaseSl []lib_model.EndpointBase) error {
	if len(eBaseSl) > 0 {
		h.m.Lock()
		defer h.m.Unlock()
		endpointsCopy := make(map[string]endpoint)
		for id, e := range h.endpoints {
			endpointsCopy[id] = e
		}
		for _, eBase := range eBaseSl {
			if err := checkIntPath(eBase.IntPath); err != nil {
				return err
			}
			if err := checkExtPath(eBase.ExtPath); err != nil {
				return err
			}
			ept := newEndpoint(lib_model.Endpoint{Type: lib_model.StandardEndpoint, EndpointBase: eBase}, h.templates)
			if ept2, ok := endpointsCopy[ept.ID]; ok {
				util.Logger.Warningf("endpoint '%+v' replaced by '%+v'", ept2.EndpointBase, ept.EndpointBase)
			}
			endpointsCopy[ept.ID] = ept
		}
		return h.update(ctx, endpointsCopy)
	}
	return nil
}

func (h *Handler) AddAlias(ctx context.Context, id, path string) error {
	return h.addAlias(ctx, id, path, lib_model.AliasEndpoint)
}

func (h *Handler) AddDefaultGui(ctx context.Context, id string) error {
	return h.addAlias(ctx, id, "", lib_model.DefaultGuiEndpoint)
}

func (h *Handler) Remove(ctx context.Context, id string, restrictStd bool) error {
	h.m.Lock()
	defer h.m.Unlock()
	e, ok := h.endpoints[id]
	if !ok {
		return lib_model.NewNotFoundError(fmt.Errorf("endpoint '%s' not found", id))
	}
	if restrictStd && e.Type == lib_model.StandardEndpoint {
		return lib_model.NewNotAllowedError(fmt.Errorf("remove endpoint '%s' not allowed", id))
	}
	endpointsCopy := make(map[string]endpoint)
	for id2, e2 := range h.endpoints {
		endpointsCopy[id2] = e2
	}
	delete(endpointsCopy, id)
	aliases := h.getAliases(id)
	for _, id := range aliases {
		delete(endpointsCopy, id)
	}
	return h.update(ctx, endpointsCopy)
}

func (h *Handler) RemoveAll(ctx context.Context, filter lib_model.EndpointFilter, restrictStd bool) error {
	if restrictStd && filterEmpty(filter) {
		return nil
	}
	h.m.Lock()
	defer h.m.Unlock()
	filtered := filterEndpoints(h.endpoints, filter)
	if len(filtered) == 0 {
		return nil
	}
	endpointsCopy := make(map[string]endpoint)
	for id, e := range h.endpoints {
		endpointsCopy[id] = e
	}
	for id, e := range filtered {
		if restrictStd && e.Type == lib_model.StandardEndpoint {
			return lib_model.NewNotAllowedError(fmt.Errorf("remove endpoint '%s' not allowed", id))
		}
		delete(endpointsCopy, id)
		aliases := h.getAliases(id)
		for _, id2 := range aliases {
			delete(endpointsCopy, id2)
		}
	}
	return h.update(ctx, endpointsCopy)
}

func (h *Handler) update(ctx context.Context, endpoints map[string]endpoint) error {
	directives, err := getDirectives(endpoints)
	if err != nil {
		return lib_model.NewInternalError(err)
	}
	if ctx.Err() != nil {
		return lib_model.NewInternalError(ctx.Err())
	}
	if err = writeConfig(directives, h.confPath); err != nil {
		return lib_model.NewInternalError(err)
	}
	if err = h.ctrHdl.ExecCmd(ctx, []string{"nginx", "-s", "reload"}, true, nil, ""); err != nil {
		e := copy(h.confPath+".bk", h.confPath)
		if e != nil {
			util.Logger.Error(e)
		}
		return lib_model.NewInternalError(err)
	}
	h.endpoints = endpoints
	return nil
}

func (h *Handler) addAlias(ctx context.Context, pID, path string, eType lib_model.EndpointType) error {
	h.m.Lock()
	defer h.m.Unlock()
	if err := checkExtPath(path); err != nil {
		return err
	}
	e, ok := h.endpoints[pID]
	if !ok {
		return lib_model.NewNotFoundError(errors.New("endpoint not found"))
	}
	if e.Type != lib_model.StandardEndpoint {
		return lib_model.NewInvalidInputError(errors.New("invalid parent type"))
	}
	endpointsCopy := make(map[string]endpoint)
	for eID, e2 := range h.endpoints {
		endpointsCopy[eID] = e2
	}
	e.ExtPath = path
	ept := newEndpoint(lib_model.Endpoint{
		ParentID:     e.ID,
		Type:         eType,
		EndpointBase: e.EndpointBase,
	}, h.templates)
	if ept2, ok := endpointsCopy[ept.ID]; ok {
		return lib_model.NewInvalidInputError(fmt.Errorf("duplicate endpoint '%s' & '%s' -> '%s'", ept.Ref, ept2.Ref, ept2.GetLocationValue()))
	}
	endpointsCopy[ept.ID] = ept
	return h.update(ctx, endpointsCopy)
}

func (h *Handler) getAliases(pID string) []string {
	var aIDs []string
	for id, e := range h.endpoints {
		if e.ParentID == pID {
			aIDs = append(aIDs, id)
		}
	}
	return aIDs
}

func getDirectives(endpoints map[string]endpoint) ([]config.IDirective, error) {
	var directives []config.IDirective
	for _, e := range endpoints {
		cmt, err := e.GenComment()
		if err != nil {
			return nil, err
		}
		var locDirectives []config.IDirective
		locDirectives = append(locDirectives, newDirective(setDirective, []string{e.GetSetValue()}, nil, nil))
		if e.Type != lib_model.DefaultGuiEndpoint {
			locDirectives = append(locDirectives, newDirective(rewriteDirective, []string{e.GetRewriteValue()}, nil, nil))
			locDirectives = append(locDirectives, getProxyDirectives(e)...)
			locDirectives = append(locDirectives, getSubFilterDirectives(e)...)
		}
		locDirectives = append(locDirectives, newDirective(proxyPassDirective, []string{e.GetProxyPassValue()}, nil, nil))
		directives = append(directives, newDirective(locationDirective, []string{e.GetLocationValue()}, []string{cmt}, newBlock(locDirectives)))
	}
	return directives, nil
}

func getProxyDirectives(e endpoint) []config.IDirective {
	var directives []config.IDirective
	headers := make(map[string]string)
	for key, val := range e.ProxyConf.Headers {
		headers[key] = strings.Replace(val, locPlaceholder, e.GetLocationValue(), -1)
	}
	if e.ProxyConf.WebSocket {
		directives = append(directives, newDirective(proxyHttpVerDirective, []string{"1.1"}, nil, nil))
		headers["Upgrade"] = "$http_upgrade"
		headers["Connection"] = "$connection_upgrade"
	}
	if e.ProxyConf.ReadTimeout > 0 {
		directives = append(directives, newDirective(proxyReadTimeoutDirective, []string{e.ProxyConf.ReadTimeout.String()}, nil, nil))
	}
	for key, val := range headers {
		directives = append(directives, newDirective(proxySetHeaderDirective, []string{key, val}, nil, nil))
	}
	return directives
}

func getSubFilterDirectives(e endpoint) []config.IDirective {
	var directives []config.IDirective
	if len(e.StringSub.Filters) > 0 {
		for orgStr, newStr := range e.StringSub.Filters {
			directives = append(directives, newDirective(subFilterDirective, []string{"'" + orgStr + "'", "'" + strings.Replace(newStr, locPlaceholder, e.GetLocationValue(), -1) + "'"}, nil, nil))
		}
		subFilterTypes := []string{"*"}
		if len(e.StringSub.MimeTypes) > 0 {
			subFilterTypes = e.StringSub.MimeTypes
		}
		directives = append(directives, newDirective(subFilterTypesDirective, subFilterTypes, nil, nil))
		subFilterOnce := "off"
		if e.StringSub.ReplaceOnce {
			subFilterOnce = "on"
		}
		directives = append(directives, newDirective(subFilterOnceDirective, []string{subFilterOnce}, nil, nil))
	}
	return directives
}

func getEndpoints(directives []config.IDirective, templates map[int]string) (map[string]endpoint, error) {
	endpoints := make(map[string]endpoint)
	for _, directive := range directives {
		if directive.GetName() == locationDirective {
			comment := directive.GetComment()
			if len(comment) > 0 {
				e, err := getEndpoint(comment[0], templates)
				if err != nil {
					return nil, err
				}
				endpoints[e.ID] = e
			}
		}
	}
	return endpoints, nil
}

func getEndpoint(s string, templates map[int]string) (endpoint, error) {
	s, _ = strings.CutPrefix(s, "#")
	var b []byte
	if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
		b = []byte(s)
	} else {
		var err error
		b, err = base64.StdEncoding.DecodeString(s)
		if err != nil {
			return endpoint{}, err
		}
	}
	var e lib_model.Endpoint
	err := json.Unmarshal(b, &e)
	if err != nil {
		return endpoint{}, err
	}
	return newEndpoint(e, templates), nil
}

func newDirective(name string, parameters, comment []string, block config.IBlock) *config.Directive {
	return &config.Directive{
		Block:      block,
		Name:       name,
		Parameters: parameters,
		Comment:    comment,
	}
}

func newBlock(directives []config.IDirective) *config.Block {
	return &config.Block{
		Directives: directives,
	}
}

func copy(src, dst string) error {
	sFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sFile.Close()
	dFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dFile.Close()
	_, err = io.Copy(dFile, sFile)
	return err
}

func writeConfig(directives []config.IDirective, path string) error {
	err := copy(path, path+".bk")
	if err != nil {
		return err
	}
	err = dumper.WriteConfig(&config.Config{
		Block:    newBlock(directives),
		FilePath: path,
	}, dumper.IndentedStyle, false)
	if err != nil {
		e := copy(path+".bk", path)
		if e != nil {
			util.Logger.Error(e)
		}
		return err
	}
	return nil
}

func filterEndpoints(endpoints map[string]endpoint, filter lib_model.EndpointFilter) map[string]endpoint {
	filtered := make(map[string]endpoint)
	var ids map[string]struct{}
	if len(filter.IDs) > 0 {
		ids = make(map[string]struct{})
		for _, id := range filter.IDs {
			ids[id] = struct{}{}
		}
	}
	for id, e := range endpoints {
		if len(ids) > 0 {
			if _, ok := ids[id]; !ok {
				continue
			}
		}
		if filter.Type > 0 && e.Type != filter.Type {
			continue
		}
		if filter.Ref != "" && e.Ref != filter.Ref {
			continue
		}
		if len(filter.Labels) > 0 {
			if !mapInMap(filter.Labels, e.Labels) {
				continue
			}
		}
		filtered[id] = e
	}
	return filtered
}

func mapInMap(a, b map[string]string) bool {
	for key, v1 := range a {
		v2, ok := b[key]
		if !ok {
			return false
		}
		if v1 != v2 {
			return false
		}
	}
	return true
}

func checkIntPath(p string) error {
	if p != "" && !strings.HasPrefix(p, "/") {
		return lib_model.NewInvalidInputError(fmt.Errorf("path '%s' not absolute", p))
	}
	return nil
}

func checkExtPath(p string) error {
	if strings.HasPrefix(p, "/") {
		return lib_model.NewInvalidInputError(fmt.Errorf("path '%s' not relative", p))
	}
	return nil
}

func filterEmpty(f lib_model.EndpointFilter) bool {
	return !(len(f.IDs) > 0 || f.Type > 0 || f.Ref != "" || len(f.Labels) > 0)
}
