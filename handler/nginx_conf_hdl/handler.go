package nginx_conf_hdl

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/tufanbarisyildirim/gonginx"
	"github.com/tufanbarisyildirim/gonginx/parser"
	"strconv"
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

func genEndpointDirectives(dID string, ept endpoint, extPath string, allowSubnets, denySubnets []string) ([]gonginx.IDirective, error) {
	uID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	varName := uID.String()
	commentItems := map[string]string{
		CommentDeploymentIDKey: ept.VarName,
		CommentHostKey:         ept.Host,
		CommentIntPathKey:      ept.Path,
		CommentExtPathKey:      extPath,
	}
	if ept.Port != nil {
		commentItems[CommentPortKey] = strconv.FormatInt(int64(*ept.Port), 10)
	}
	directives := []gonginx.IDirective{
		newDirective(proxyPassDirective, []string{fmt.Sprintf("http://%s%s$1$is_args$args", varName, internalPath)}, nil, nil),
	}
	for _, subnet := range allowSubnets {
		directives = append(directives, newDirective(allowDirective, []string{subnet}, nil, nil))
	}
	for _, subnet := range denySubnets {
		directives = append(directives, newDirective(denyDirective, []string{subnet}, nil, nil))
	}
	return []gonginx.IDirective{
		newDirective(setDirective, []string{fmt.Sprintf("$%s %s", varName, host)}, comment, nil),
		newDirective(locationDirective, []string{fmt.Sprintf("~ ^/%s/%s/(.*)$", basePath, dID)}, comment, newBlock(directives)),
	}, nil
}
