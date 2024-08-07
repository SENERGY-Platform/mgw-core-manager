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

import lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"

const (
	locationDirective         = "location"
	rewriteDirective          = "rewrite"
	proxyPassDirective        = "proxy_pass"
	setDirective              = "set"
	proxyHttpVerDirective     = "proxy_http_version"
	proxySetHeaderDirective   = "proxy_set_header"
	proxyReadTimeoutDirective = "proxy_read_timeout"
	subFilterDirective        = "sub_filter"
	subFilterOnceDirective    = "sub_filter_once"
	subFilterTypesDirective   = "sub_filter_types"
)

const (
	varPlaceholder  = "{var}"
	portPlaceholder = "{port}"
	pathPlaceholder = "{path}"
	refPlaceholder  = "{ref}"
	locPlaceholder  = "{loc}"
)

const (
	locationTmpl = iota
	rewriteTmpl
	proxyPassTmpl
)

const (
	StandardLocationTmpl = iota
	StandardRewriteTmpl
	StandardProxyPassTmpl
	DefaultGuiLocationTmpl
	DefaultGuiProxyPassTmpl
	AliasLocationTmpl
	AliasRewriteTmpl
	AliasProxyPassTmpl
)

var endpointTypeMap = map[int]map[int]int{
	lib_model.StandardEndpoint: {
		locationTmpl:  StandardLocationTmpl,
		rewriteTmpl:   StandardRewriteTmpl,
		proxyPassTmpl: StandardProxyPassTmpl,
	},
	lib_model.DefaultGuiEndpoint: {
		locationTmpl:  DefaultGuiLocationTmpl,
		proxyPassTmpl: DefaultGuiProxyPassTmpl,
	},
	lib_model.AliasEndpoint: {
		locationTmpl:  AliasLocationTmpl,
		rewriteTmpl:   AliasRewriteTmpl,
		proxyPassTmpl: AliasProxyPassTmpl,
	},
}
