package nginx_conf_hdl

import (
	"fmt"
	"github.com/tufanbarisyildirim/gonginx"
	"github.com/tufanbarisyildirim/gonginx/parser"
)

type Handler struct {
	endPntPath   string
	srcConfPath  string
	tgtConfPath  string
	allowSubnets []string
	denySubnets  []string
	endpoints    map[string]map[string]endpoint // {dID:{extPath:endpoint}}
}

func New(srcPath, tgtPath, endPntPath string, allowSubnets, denySubnets []string) *Handler {
	return &Handler{
		endPntPath:   endPntPath,
		srcConfPath:  srcPath,
		tgtConfPath:  tgtPath,
		allowSubnets: allowSubnets,
		denySubnets:  denySubnets,
	}
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

func (h *Handler) setEndpoint(directives []gonginx.IDirective, ept endpoint, locationTemplate, proxyPassTemplate string) error {
	cmt, err := ept.GenComment()
	if err != nil {
		return err
	}
	directives = append(directives, newDirective(setDirective, []string{fmt.Sprintf("$%s %s", ept.VarName, ept.Host)}, []string{cmt}, nil))
	locDirectives := []gonginx.IDirective{
		newDirective(proxyPassDirective, []string{ept.GenProxyPassValue(proxyPassTemplate)}, nil, nil),
	}
	for _, subnet := range h.allowSubnets {
		locDirectives = append(locDirectives, newDirective(allowDirective, []string{subnet}, nil, nil))
	}
	for _, subnet := range h.denySubnets {
		locDirectives = append(locDirectives, newDirective(denyDirective, []string{subnet}, nil, nil))
	}
	directives = append(directives, newDirective(locationDirective, []string{ept.GenLocationValue(locationTemplate)}, nil, newBlock(directives)))
	return nil
}
