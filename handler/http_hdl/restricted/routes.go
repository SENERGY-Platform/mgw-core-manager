/*
 * Copyright 2025 InfAI (CC SES)
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

package restricted

import (
	"github.com/SENERGY-Platform/mgw-core-manager/handler/http_hdl/shared"
	"github.com/SENERGY-Platform/mgw-core-manager/handler/http_hdl/util"
	"github.com/SENERGY-Platform/mgw-core-manager/lib"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/gin-gonic/gin"
)

var routes = util.Routes{
	DeleteEndpointH,
	PostEndpointH,
	DeleteEndpointH,
	DeleteEndpointBatchH,
}

// SetRoutes
// @title Core Manager Public API
// @version 0.8.0
// @description Provides access to public management options for the multi-gateway core.
// @license.name Apache-2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /restricted
func SetRoutes(e *gin.Engine, a lib.Api) {
	rg := e.Group(lib_model.RestrictedPath)
	shared.Routes.Set(a, rg)
	routes.Set(a, rg)
}
