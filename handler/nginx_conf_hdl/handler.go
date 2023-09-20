package nginx_conf_hdl

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/tufanbarisyildirim/gonginx"
	"github.com/tufanbarisyildirim/gonginx/parser"
	"strings"
)

type Handler struct {
	endPntPath   string
	srcConfPath  string
	tgtConfPath  string
	allowSubnets []string
	denySubnets  []string
	endpoints    map[string]map[string]endpoint // {dID:{extPath:endpoint}}
}

type endpoint struct {
	Host    string
	Port    *int
	Path    string // intPath
	VarName string // hash(DeploymentID, Host, Path)
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

func getEndpointDirectives(mID, dID, basePath, internalPath, host string, allowSubnets, denySubnets []string) ([]gonginx.IDirective, error) {
	uID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	varName := uID.String()
	comment := []string{fmt.Sprintf("#mid=%s,did=%s", mID, dID)}
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

func genComment(items map[string]string) string {
	c := "#"
	n := len(items) - 1
	i := 0
	for key, val := range items {
		item := key + CommentItemDelimiter + val
		if i < n {
			c += item + CommentDelimiter
		} else {
			c += item
		}
		i++
	}
	return c
}

func parseComment(c string) (map[string]string, error) {
	c, _ = strings.CutPrefix(c, "#")
	if c == "" {
		return nil, fmt.Errorf("parsing nginx config comment failed: empty")
	}
	items := strings.Split(c, CommentDelimiter)
	m := make(map[string]string)
	for _, item := range items {
		parts := strings.Split(item, CommentItemDelimiter)
		if len(parts) != 2 {
			return nil, fmt.Errorf("parsing nginx config comment failed: '%s' -> '%s'", c, item)
		}
		key := parts[0]
		if key == "" {
			return nil, fmt.Errorf("parsing nginx config comment failed: '%s' -> '%s'", c, item)
		}
		m[key] = parts[1]
	}
	return m, nil
}
