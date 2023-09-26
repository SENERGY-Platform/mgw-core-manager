package nginx_conf_hdl

import (
	"fmt"
	"github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"github.com/tufanbarisyildirim/gonginx"
	"github.com/tufanbarisyildirim/gonginx/parser"
	"io"
	"os"
)

type Handler struct {
	confPath     string
	endPntPath   string
	allowSubnets []string
	denySubnets  []string
	templates    map[int]string
	endpoints    map[string]map[string]endpoint // {dID:{extPath:endpoint}}
	locations    map[string]struct{}
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
	h.endpoints, h.locations, err = getEndpoints(conf.GetDirectives(), h.templates)
	return nil
}

func (h *Handler) Add(e model.Endpoint, t model.EndpointType) error {
	dMap, ok := h.endpoints[e.DeploymentID]
	if !ok {
		dMap = make(map[string]endpoint)
		h.endpoints[e.DeploymentID] = dMap
	}
	if e2, ok := dMap[e.ExtPath]; ok {
		return model.NewInvalidInputError(fmt.Errorf("duplicate endpoint for '%s': '%s' -> '%s' & '%s'", e.DeploymentID, e.ExtPath, e.IntPath, e2.IntPath))
	}
	ept := newEndpoint(e, t)
	loc := ept.GenLocationValue(h.templates)
	if _, ok = h.locations[loc]; ok {
		return model.NewInvalidInputError(fmt.Errorf("duplicate location '%s'", loc))
	}
	h.locations[loc] = struct{}{}
	dMap[e.ExtPath] = ept
	return nil
}

func (h *Handler) Remove(dID, extPath string) error {
	if dMap, ok := h.endpoints[dID]; ok {
		if _, ok := dMap[extPath]; ok {
			delete(dMap, extPath)
			return nil
		}
	}
	return model.NewNotFoundError(fmt.Errorf("endpoint '%s' not found for '%s'", extPath, dID))
}

func (h *Handler) RemoveAll(dID string) error {
	if _, ok := h.endpoints[dID]; ok {
		delete(h.endpoints, dID)
		return nil
	}
	return model.NewNotFoundError(fmt.Errorf("no endpoints found for '%s'", dID))
}

func (h *Handler) writeEndpoints() error {
	var directives []gonginx.IDirective
	var err error
	for _, dMap := range h.endpoints {
		for _, e := range dMap {
			directives, err = setEndpoint(directives, e, h.templates, h.allowSubnets, h.denySubnets)
			if err != nil {
				return err
			}
		}
	}
	err = copy(h.confPath, h.confPath+".bk")
	if err != nil {
		return err
	}
	err = gonginx.WriteConfig(&gonginx.Config{
		Block:    newBlock(directives),
		FilePath: h.confPath,
	}, gonginx.IndentedStyle, false)
	if err != nil {
		e := copy(h.confPath+".bk", h.confPath)
		if e != nil {
			util.Logger.Error(e)
		}
		return err
	}
	return nil
}

func getEndpoints(directives []gonginx.IDirective, templates map[int]string) (map[string]map[string]endpoint, map[string]struct{}, error) {
	endpoints := make(map[string]map[string]endpoint)
	locations := make(map[string]struct{})
	for _, directive := range directives {
		if directive.GetName() == setDirective {
			comment := directive.GetComment()
			if len(comment) > 0 {
				e, err := parseEndpoint(comment[0])
				if err != nil {
					return nil, nil, err
				}
				dMap, ok := endpoints[e.DeploymentID]
				if !ok {
					dMap = make(map[string]endpoint)
					endpoints[e.DeploymentID] = dMap
				}
				dMap[e.ExtPath] = e
				locations[e.GenLocationValue(templates)] = struct{}{}
			}
		}
	}
	return endpoints, locations, nil
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

func setEndpoint(directives []gonginx.IDirective, ept endpoint, templates map[int]string, allowSubnets, denySubnets []string) ([]gonginx.IDirective, error) {
	cmt, err := ept.GenComment()
	if err != nil {
		return nil, err
	}
	directives = append(directives, newDirective(setDirective, []string{ept.GenSetValue()}, []string{cmt}, nil))
	locDirectives := []gonginx.IDirective{
		newDirective(proxyPassDirective, []string{ept.GenProxyPassValue(templates)}, nil, nil),
	}
	for _, subnet := range allowSubnets {
		locDirectives = append(locDirectives, newDirective(allowDirective, []string{subnet}, nil, nil))
	}
	for _, subnet := range denySubnets {
		locDirectives = append(locDirectives, newDirective(denyDirective, []string{subnet}, nil, nil))
	}
	directives = append(directives, newDirective(locationDirective, []string{ept.GenLocationValue(templates)}, nil, newBlock(locDirectives)))
	return directives, nil
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
