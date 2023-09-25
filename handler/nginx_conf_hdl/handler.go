package nginx_conf_hdl

import (
	"fmt"
	"github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/tufanbarisyildirim/gonginx"
	"github.com/tufanbarisyildirim/gonginx/parser"
	"os"
)

type Handler struct {
	endPntPath   string
	tgtConfPath  string
	allowSubnets []string
	denySubnets  []string
	templates    map[int]string
	endpoints    map[string]map[string]endpoint // {dID:{extPath:endpoint}}
	locations    map[string]struct{}
	srcConfBlock gonginx.IBlock
}

func New(tgtConfPath, endPntPath string, allowSubnets, denySubnets []string, templates map[int]string) *Handler {
	return &Handler{
		endPntPath:   endPntPath,
		tgtConfPath:  tgtConfPath,
		allowSubnets: allowSubnets,
		denySubnets:  denySubnets,
		templates:    templates,
	}
}

func (h *Handler) Init(baseConfPath string) error {
	conf, err := readConf(baseConfPath)
	if err != nil {
		return err
	}
	h.srcConfBlock = conf.Block
	_, err = os.Stat(h.tgtConfPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		h.endpoints = make(map[string]map[string]endpoint)
		h.locations = map[string]struct{}{}
		conf.FilePath = h.tgtConfPath
		return writeConf(conf)
	} else {
		if err = h.readEndpoints(); err != nil {
			return err
		}
		return h.writeEndpoints()
	}
}

func (h *Handler) Add(e model.Endpoint, t model.EndpointType) error {
	dMap, ok := h.endpoints[e.DeploymentID]
	if !ok {
		dMap = make(map[string]endpoint)
		h.endpoints[e.DeploymentID] = dMap
	}
	if e2, ok := dMap[e.ExtPath]; ok {
	ept := newEndpoint(e, t)
	loc := ept.GenLocationValue(h.templates)
	}
	dMap[e.ExtPath] = newEndpoint(e, endpointTypeMap[t])
	return nil
}

func (h *Handler) readEndpoints() error {
	conf, err := readConf(h.tgtConfPath)
	if err != nil {
		return err
	}
	for _, directive := range conf.GetDirectives() {
		if directive.GetName() == serverDirective {
			block := directive.GetBlock()
			if block != nil {
				h.endpoints, h.locations, err = getEndpoints(block, h.templates)
				if err != nil {
					return err
				}
			}
			break
		}
	}
	return nil
}

func (h *Handler) writeEndpoints() error {
	var directives []gonginx.IDirective
	var err error
	for _, directive := range h.srcConfBlock.GetDirectives() {
		if directive.GetName() == serverDirective {
			var srvDirectives []gonginx.IDirective
			block := directive.GetBlock()
			if block != nil {
				srvDirectives = block.GetDirectives()
			}
			for _, dMap := range h.endpoints {
				for _, e := range dMap {
					srvDirectives, err = setEndpoint(srvDirectives, e, h.templates, h.allowSubnets, h.denySubnets)
					if err != nil {
						return err
					}
				}
			}
			directives = append(directives, newDirective(directive.GetName(), directive.GetParameters(), directive.GetComment(), newBlock(srvDirectives)))
		} else {
			directives = append(directives, directive)
		}
	}
	return writeConf(&gonginx.Config{
		Block:    newBlock(directives),
		FilePath: h.tgtConfPath,
	})
}

func getEndpoints(block gonginx.IBlock, templates map[int]string) (map[string]map[string]endpoint, map[string]struct{}, error) {
	endpoints := make(map[string]map[string]endpoint)
	locations := make(map[string]struct{})
	for _, directive := range block.GetDirectives() {
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

func readConf(path string) (*gonginx.Config, error) {
	p, err := parser.NewParser(path)
	if err != nil {
		return nil, err
	}
	return p.Parse(), err
}

func writeConf(conf *gonginx.Config) error {
	return gonginx.WriteConfig(conf, gonginx.IndentedStyle, false)
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
