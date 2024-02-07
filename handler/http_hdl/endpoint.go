/*
 * Copyright 2024 InfAI (CC SES)
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

package http_hdl

import (
	"github.com/SENERGY-Platform/mgw-core-manager/lib"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/gin-gonic/gin"
	"net/http"
)

const endpointIdParam = "e"

type endpointFilterQuery struct {
	IDs    string `form:"ids"`
	Type   *int   `form:"type"`
	Ref    string `form:"ref"`
	Labels string `form:"labels"`
}

type postEndpointQuery struct {
	Action string `form:"action"`
}

type deleteEndpointBatchQuery struct {
	IDs    string `form:"ids"`
	Ref    string `form:"ref"`
	Labels string `form:"labels"`
}

func getEndpointsH(a lib.Api) gin.HandlerFunc {
	return func(gc *gin.Context) {
		query := endpointFilterQuery{}
		if err := gc.ShouldBindQuery(&query); err != nil {
			_ = gc.Error(lib_model.NewInvalidInputError(err))
			return
		}
		filter := lib_model.EndpointFilter{
			IDs:    parseStringSlice(query.IDs, ","),
			Type:   query.Type,
			Ref:    query.Ref,
			Labels: genLabels(parseStringSlice(query.Labels, ",")),
		}
		endpoints, err := a.GetEndpoints(gc.Request.Context(), filter)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.JSON(http.StatusOK, endpoints)
	}
}

func getEndpointH(a lib.Api) gin.HandlerFunc {
	return func(gc *gin.Context) {
		endpoint, err := a.GetEndpoint(gc.Request.Context(), gc.Param(endpointIdParam))
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.JSON(http.StatusOK, endpoint)
	}
}

func postEndpointH(a lib.Api) gin.HandlerFunc {
	return func(gc *gin.Context) {
		query := postEndpointQuery{}
		if err := gc.ShouldBindQuery(&query); err != nil {
			_ = gc.Error(lib_model.NewInvalidInputError(err))
			return
		}
		var jID string
		var err error
		if query.Action != "" {
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
			case "alias":
				jID, err = a.AddEndpointAlias(gc.Request.Context(), aliasReq.ParentID, aliasReq.Path)
				if err != nil {
					_ = gc.Error(err)
					return
				}
			}
		} else {
			var endpointBase lib_model.EndpointBase
			if err = gc.ShouldBindJSON(&endpointBase); err != nil {
				_ = gc.Error(lib_model.NewInvalidInputError(err))
				return
			}
			jID, err = a.AddEndpoint(gc.Request.Context(), endpointBase)
			if err != nil {
				_ = gc.Error(err)
				return
			}
		}
		gc.String(http.StatusOK, jID)
	}
}

func deleteEndpointH(a lib.Api) gin.HandlerFunc {
	return func(gc *gin.Context) {
		jID, err := a.RemoveEndpoint(gc.Request.Context(), gc.Param(endpointIdParam), false)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.String(http.StatusOK, jID)
	}
}

func postEndpointBatchH(a lib.Api) gin.HandlerFunc {
	return func(gc *gin.Context) {
		var endpointBaseSl []lib_model.EndpointBase
		if err := gc.ShouldBindJSON(&endpointBaseSl); err != nil {
			_ = gc.Error(lib_model.NewInvalidInputError(err))
			return
		}
		jID, err := a.AddEndpoints(gc.Request.Context(), endpointBaseSl)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.String(http.StatusOK, jID)
	}
}

func deleteEndpointBatchH(a lib.Api) gin.HandlerFunc {
	return func(gc *gin.Context) {
		query := deleteEndpointBatchQuery{}
		if err := gc.ShouldBindQuery(&query); err != nil {
			_ = gc.Error(lib_model.NewInvalidInputError(err))
			return
		}
		jID, err := a.RemoveEndpoints(gc.Request.Context(), lib_model.EndpointFilter{
			IDs:    parseStringSlice(query.IDs, ","),
			Ref:    query.Ref,
			Labels: genLabels(parseStringSlice(query.Labels, ",")),
		}, false)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.String(http.StatusOK, jID)
	}
}

func postEndpointRestrictedH(a lib.Api) gin.HandlerFunc {
	return func(gc *gin.Context) {
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
	}
}

func deleteEndpointRestrictedH(a lib.Api) gin.HandlerFunc {
	return func(gc *gin.Context) {
		jID, err := a.RemoveEndpoint(gc.Request.Context(), gc.Param(endpointIdParam), true)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.String(http.StatusOK, jID)
	}
}
