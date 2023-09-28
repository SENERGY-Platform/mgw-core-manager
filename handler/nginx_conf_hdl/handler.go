package nginx_conf_hdl

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"github.com/tufanbarisyildirim/gonginx"
	"github.com/tufanbarisyildirim/gonginx/parser"
	"io"
	"os"
	"strings"
	"sync"
)

type Handler struct {
	confPath     string
	endPntPath   string
	allowSubnets []string
	denySubnets  []string
	templates    map[int]string
	endpoints    map[string]endpoint
	m            sync.RWMutex
}

func New(confPath, endPntPath string, allowSubnets, denySubnets []string, templates map[int]string) *Handler {
	return &Handler{
		confPath:     confPath,
		endPntPath:   endPntPath,
		allowSubnets: allowSubnets,
		denySubnets:  denySubnets,
		templates:    templates,
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
	conf := p.Parse()
	h.endpoints, err = getEndpoints(conf.GetDirectives(), h.templates)
	return nil
}

func (h *Handler) List(ctx context.Context) ([]model.Endpoint, error) {
	h.m.RLock()
	defer h.m.RUnlock()
	var endpoints []model.Endpoint
	for _, e := range h.endpoints {
		if ctx.Err() != nil {
			return nil, model.NewInternalError(ctx.Err())
		}
		endpoints = append(endpoints, e.Endpoint)
	}
	return endpoints, nil
}

func (h *Handler) Add(ctx context.Context, endpoints []model.Endpoint) error {
	h.m.Lock()
	defer h.m.Unlock()
	endpointsCopy := make(map[string]endpoint)
	for id, e := range h.endpoints {
		endpointsCopy[id] = e
	}
	for _, e := range endpoints {
		ept := newEndpoint(e, h.templates)
		if ept2, ok := endpointsCopy[ept.ID]; ok {
			return model.NewInvalidInputError(fmt.Errorf("duplicate endpoint '%s' & '%s' -> '%s'", ept.Ref, ept2.Ref, ept2.GetLocationValue()))
		}
		endpointsCopy[ept.ID] = ept
	}
	return h.update(ctx, endpointsCopy)
}

func (h *Handler) Remove(ctx context.Context, id string) error {
	h.m.Lock()
	defer h.m.Unlock()
	if _, ok := h.endpoints[id]; !ok {
		return model.NewNotFoundError(fmt.Errorf("endpoint '%s' not found", id))
	}
	endpointsCopy := make(map[string]endpoint)
	for id2, e := range h.endpoints {
		endpointsCopy[id2] = e
	}
	delete(endpointsCopy, id)
	return h.update(ctx, endpointsCopy)
}

func (h *Handler) RemoveAll(ctx context.Context, ref string) error {
	h.m.Lock()
	defer h.m.Unlock()
	endpointsCopy := make(map[string]endpoint)
	for id, e := range h.endpoints {
		if e.Ref != ref {
			endpointsCopy[id] = e
		}
	}
	if len(h.endpoints) == len(endpointsCopy) {
		return model.NewNotFoundError(fmt.Errorf("no endpoints found for '%s'", ref))
	}
	return h.update(ctx, endpointsCopy)
}

func (h *Handler) update(ctx context.Context, endpoints map[string]endpoint) error {
	directives, err := getDirectives(endpoints, h.allowSubnets, h.denySubnets)
	if err != nil {
		return model.NewInternalError(err)
	}
	if ctx.Err() != nil {
		return model.NewInternalError(ctx.Err())
	}
	if err = writeConfig(directives, h.confPath); err != nil {
		return model.NewInternalError(err)
	}
	h.endpoints = endpoints
	return nil
}

func getDirectives(endpoints map[string]endpoint, allowSubnets, denySubnets []string) ([]gonginx.IDirective, error) {
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
		for _, subnet := range allowSubnets {
			locDirectives = append(locDirectives, newDirective(allowDirective, []string{subnet}, nil, nil))
		}
		for _, subnet := range denySubnets {
			locDirectives = append(locDirectives, newDirective(denyDirective, []string{subnet}, nil, nil))
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
	var e model.Endpoint
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
