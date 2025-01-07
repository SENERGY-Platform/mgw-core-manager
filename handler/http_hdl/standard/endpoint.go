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

package standard

import (
	"github.com/SENERGY-Platform/mgw-core-manager/handler/http_hdl/util"
	"github.com/SENERGY-Platform/mgw-core-manager/lib"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"path"
)

type deleteEndpointBatchQuery struct {
	IDs    string `form:"ids"`
	Ref    string `form:"ref"`
	Labels string `form:"labels"`
}

// PostEndpointH
// @Summary Create HTTP endpoint
// @Description	Create an HTTP endpoint accessible via the core reverse proxy.
// @Tags Endpoints
// @Accept json
// @Produce	plain
// @Param endpoint body lib_model.EndpointBase true "endpoint information"
// @Success	200 {string} string "job ID"
// @Failure	400 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /endpoints [post]
func PostEndpointH(a lib.Api, rg *gin.RouterGroup) {
	rg.POST(lib_model.EndpointsPath, func(gc *gin.Context) {
		var jID string
		var err error
		var endpointBase lib_model.EndpointBase
		if err = gc.ShouldBindJSON(&endpointBase); err != nil {
			_ = gc.Error(lib_model.NewInvalidInputError(err))
			return
		}
		jID, err = a.SetEndpoint(gc.Request.Context(), endpointBase)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.String(http.StatusOK, jID)
	})
}

func DeleteEndpointH(a lib.Api, rg *gin.RouterGroup) {
	rg.DELETE(path.Join(lib_model.EndpointsPath, ":id"), func(gc *gin.Context) {
		jID, err := a.RemoveEndpoint(gc.Request.Context(), gc.Param("id"), false)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.String(http.StatusOK, jID)
	})
}

func PostEndpointBatchH(a lib.Api, rg *gin.RouterGroup) {
	rg.POST(lib_model.EndpointsBatchPath, func(gc *gin.Context) {
		var endpointBaseSl []lib_model.EndpointBase
		if err := gc.ShouldBindJSON(&endpointBaseSl); err != nil {
			_ = gc.Error(lib_model.NewInvalidInputError(err))
			return
		}
		jID, err := a.SetEndpoints(gc.Request.Context(), endpointBaseSl)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.String(http.StatusOK, jID)
	})
}

func DeleteEndpointBatchH(a lib.Api, rg *gin.RouterGroup) {
	rg.DELETE(lib_model.EndpointsBatchPath, func(gc *gin.Context) {
		query := deleteEndpointBatchQuery{}
		if err := gc.ShouldBindQuery(&query); err != nil {
			_ = gc.Error(lib_model.NewInvalidInputError(err))
			return
		}
		jID, err := a.RemoveEndpoints(gc.Request.Context(), lib_model.EndpointFilter{
			IDs:    util.ParseStringSlice(query.IDs, ","),
			Ref:    query.Ref,
			Labels: util.GenLabels(util.ParseStringSlice(query.Labels, ",")),
		}, false)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.String(http.StatusOK, jID)
	})
}
