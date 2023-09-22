package nginx_conf_hdl

import (
	"github.com/tufanbarisyildirim/gonginx"
	"github.com/tufanbarisyildirim/gonginx/parser"
)

type Handler struct {
	endPntPath   string
	tgtConfPath  string
	allowSubnets []string
	denySubnets  []string
	endpoints    map[string]map[string]endpoint // {dID:{extPath:endpoint}}
	srcConf      gonginx.IBlock
}

func New(tgtConfPath, endPntPath string, allowSubnets, denySubnets []string) *Handler {
	return &Handler{
		endPntPath:   endPntPath,
		tgtConfPath:  tgtConfPath,
		allowSubnets: allowSubnets,
		denySubnets:  denySubnets,
	}
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
				h.endpoints, err = getEndpoints(block)
				if err != nil {
					return err
				}
			}
			break
		}
	}
	return nil
}

func getEndpoints(block gonginx.IBlock) (map[string]map[string]endpoint, error) {
	endpoints := make(map[string]map[string]endpoint)
	for _, directive := range block.GetDirectives() {
		if directive.GetName() == setDirective {
			comment := directive.GetComment()
			if len(comment) > 0 {
				e, err := parseEndpoint(comment[0])
				if err != nil {
					return nil, err
				}
				dMap, ok := endpoints[e.DeploymentID]
				if !ok {
					dMap = make(map[string]endpoint)
					endpoints[e.DeploymentID] = dMap
				}
				dMap[e.ExtPath] = e
			}
		}
	}
	return endpoints, nil
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

func writeConf(conf *gonginx.Config, path string) error {
	conf.FilePath = path
	return gonginx.WriteConfig(conf, gonginx.IndentedStyle, false)
}

func setEndpoint(directives []gonginx.IDirective, ept endpoint, locationTemplate, proxyPassTemplate string, allowSubnets, denySubnets []string) error {
	cmt, err := ept.GenComment()
	if err != nil {
		return err
	}
	directives = append(directives, newDirective(setDirective, []string{ept.GenSetValue()}, []string{cmt}, nil))
	locDirectives := []gonginx.IDirective{
		newDirective(proxyPassDirective, []string{ept.GenProxyPassValue(proxyPassTemplate)}, nil, nil),
	}
	for _, subnet := range allowSubnets {
		locDirectives = append(locDirectives, newDirective(allowDirective, []string{subnet}, nil, nil))
	}
	for _, subnet := range denySubnets {
		locDirectives = append(locDirectives, newDirective(denyDirective, []string{subnet}, nil, nil))
	}
	directives = append(directives, newDirective(locationDirective, []string{ept.GenLocationValue(locationTemplate)}, nil, newBlock(directives)))
	return nil
}
