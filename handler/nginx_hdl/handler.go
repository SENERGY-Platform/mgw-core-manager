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
	"encoding/json"
	"errors"
	"fmt"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"github.com/tufanbarisyildirim/gonginx"
	"github.com/tufanbarisyildirim/gonginx/parser"
	"io"
	"os"
	"strings"
	"sync"
)

type Handler struct {
	confPath  string
	templates map[int]string
	endpoints map[string]endpoint
	m         sync.RWMutex
}

func New(confPath string, templates map[int]string) *Handler {
	return &Handler{
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
	return e.Endpoint, nil
}

func (h *Handler) Add(ctx context.Context, ept lib_model.Endpoint) error {
	h.m.Lock()
	defer h.m.Unlock()
	endpointsCopy := make(map[string]endpoint)
	for id, e := range h.endpoints {
		endpointsCopy[id] = e
	}
	e := newEndpoint(ept, h.templates)
	if ept2, ok := endpointsCopy[e.ID]; ok {
		return lib_model.NewInvalidInputError(fmt.Errorf("duplicate endpoint '%s' & '%s' -> '%s'", ept.Ref, ept2.Ref, ept2.GetLocationValue()))
	}
	endpointsCopy[e.ID] = e
	return h.update(ctx, endpointsCopy)
}

func (h *Handler) AddList(ctx context.Context, endpoints []lib_model.Endpoint) error {
	if len(endpoints) > 0 {
		h.m.Lock()
		defer h.m.Unlock()
		endpointsCopy := make(map[string]endpoint)
		for id, e := range h.endpoints {
			endpointsCopy[id] = e
		}
		for _, e := range endpoints {
			ept := newEndpoint(e, h.templates)
			if ept2, ok := endpointsCopy[ept.ID]; ok {
				return lib_model.NewInvalidInputError(fmt.Errorf("duplicate endpoint '%s' & '%s' -> '%s'", ept.Ref, ept2.Ref, ept2.GetLocationValue()))
			}
			endpointsCopy[ept.ID] = ept
		}
		return h.update(ctx, endpointsCopy)
	}
	return nil
}

func (h *Handler) Remove(ctx context.Context, id string) error {
	h.m.Lock()
	defer h.m.Unlock()
	if _, ok := h.endpoints[id]; !ok {
		return lib_model.NewNotFoundError(fmt.Errorf("endpoint '%s' not found", id))
	}
	endpointsCopy := make(map[string]endpoint)
	for id2, e := range h.endpoints {
		endpointsCopy[id2] = e
	}
	delete(endpointsCopy, id)
	return h.update(ctx, endpointsCopy)
}

func (h *Handler) RemoveAll(ctx context.Context, filter lib_model.EndpointFilter) error {
	h.m.Lock()
	defer h.m.Unlock()
	filtered := filterEndpoints(h.endpoints, filter)
	endpoints := make(map[string]endpoint)
	for id, e := range h.endpoints {
		if _, ok := filtered[id]; !ok {
			endpoints[id] = e
		}
	}
	return h.update(ctx, endpoints)
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
	h.endpoints = endpoints
	return nil
}

func getDirectives(endpoints map[string]endpoint) ([]gonginx.IDirective, error) {
	var directives []gonginx.IDirective
	for _, e := range endpoints {
		cmt, err := e.GenComment()
		if err != nil {
			return nil, err
		}
		directives = append(directives, newDirective(setDirective, []string{e.GetSetValue()}, []string{cmt}, nil))
		locDirectives := []gonginx.IDirective{
			newDirective(proxyPassDirective, []string{e.GetProxyPassValue()}, nil, nil),
		}
		directives = append(directives, newDirective(locationDirective, []string{e.GetLocationValue()}, nil, newBlock(locDirectives)))
	}
	return directives, nil
}

func getEndpoints(directives []gonginx.IDirective, templates map[int]string) (map[string]endpoint, error) {
	endpoints := make(map[string]endpoint)
	for _, directive := range directives {
		if directive.GetName() == setDirective {
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
	var e lib_model.Endpoint
	err := json.Unmarshal([]byte(s), &e)
	if err != nil {
		return endpoint{}, err
	}
	return newEndpoint(e, templates), nil
}

func newDirective(name string, parameters, comment []string, block gonginx.IBlock) *gonginx.Directive {
	return &gonginx.Directive{
		Block:      block,
		Name:       name,
		Parameters: parameters,
		Comment:    comment,
	}
}

func newBlock(directives []gonginx.IDirective) *gonginx.Block {
	return &gonginx.Block{
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

func writeConfig(directives []gonginx.IDirective, path string) error {
	err := copy(path, path+".bk")
	if err != nil {
		return err
	}
	err = gonginx.WriteConfig(&gonginx.Config{
		Block:    newBlock(directives),
		FilePath: path,
	}, gonginx.IndentedStyle, false)
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
		if filter.Type != nil && e.Type != *filter.Type {
			continue
		}
		if filter.Ref != "" && e.Ref != filter.Ref {
			continue
		}
		filtered[id] = e
	}
	return filtered
}
