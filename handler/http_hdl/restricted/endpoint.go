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
	"github.com/SENERGY-Platform/mgw-core-manager/handler/http_hdl/util"
	"github.com/SENERGY-Platform/mgw-core-manager/lib"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"path"
)

type postEndpointQuery struct {
	Action string `form:"action"`
}

type deleteEndpointBatchQuery struct {
	IDs    string `form:"ids"`
	Ref    string `form:"ref"`
	Labels string `form:"labels"`
}

func SetPostEndpointH(a lib.Api, rg *gin.RouterGroup) {
	rg.POST(lib_model.EndpointsPath, func(gc *gin.Context) {
		query := postEndpointQuery{}
		if err := gc.ShouldBindQuery(&query); err != nil {
			_ = gc.Error(lib_model.NewInvalidInputError(err))
			return
		}
		var jID string
		var err error
		var aliasReq lib_model.EndpointAliasReq
		if err = gc.ShouldBindJSON(&aliasReq); err != nil {
			_ = gc.Error(lib_model.NewInvalidInputError(err))
			return
		}
		switch query.Action {
		case "gui":
			jID, err = a.AddDefaultGuiEndpoint(gc.Request.Context(), aliasReq.ParentID)
			if err != nil {
				_ = gc.Error(err)
				return
			}
		default:
			jID, err = a.AddEndpointAlias(gc.Request.Context(), aliasReq.ParentID, aliasReq.Path)
			if err != nil {
				_ = gc.Error(err)
				return
			}
		}
		gc.String(http.StatusOK, jID)
	})
}

func SetDeleteEndpointH(a lib.Api, rg *gin.RouterGroup) {
	rg.DELETE(path.Join(lib_model.EndpointsPath, ":id"), func(gc *gin.Context) {
		jID, err := a.RemoveEndpoint(gc.Request.Context(), gc.Param("id"), true)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.String(http.StatusOK, jID)
	})
}

func SetDeleteEndpointBatchH(a lib.Api, rg *gin.RouterGroup) {
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
		}, true)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.String(http.StatusOK, jID)
	})
}
