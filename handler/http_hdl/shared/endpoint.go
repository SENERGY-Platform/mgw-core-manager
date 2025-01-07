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

package shared

import (
	"github.com/SENERGY-Platform/mgw-core-manager/handler/http_hdl/util"
	"github.com/SENERGY-Platform/mgw-core-manager/lib"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"path"
)

type endpointFilterQuery struct {
	IDs    string `form:"ids"`
	Type   int    `form:"type"`
	Ref    string `form:"ref"`
	Labels string `form:"labels"`
}

func GetEndpointsH(a lib.Api, rg *gin.RouterGroup) {
	rg.GET(lib_model.EndpointsPath, func(gc *gin.Context) {
		query := endpointFilterQuery{}
		if err := gc.ShouldBindQuery(&query); err != nil {
			_ = gc.Error(lib_model.NewInvalidInputError(err))
			return
		}
		filter := lib_model.EndpointFilter{
			IDs:    util.ParseStringSlice(query.IDs, ","),
			Type:   query.Type,
			Ref:    query.Ref,
			Labels: util.GenLabels(util.ParseStringSlice(query.Labels, ",")),
		}
		endpoints, err := a.GetEndpoints(gc.Request.Context(), filter)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.JSON(http.StatusOK, endpoints)
	})
}

func GetEndpointH(a lib.Api, rg *gin.RouterGroup) {
	rg.GET(path.Join(lib_model.EndpointsPath, ":id"), func(gc *gin.Context) {
		endpoint, err := a.GetEndpoint(gc.Request.Context(), gc.Param("id"))
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.JSON(http.StatusOK, endpoint)
	})
}
