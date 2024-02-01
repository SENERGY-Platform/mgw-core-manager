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
	IDs  string `form:"name"`
	Type *int   `form:"type"`
	Ref  string `form:"ref"`
}

type postEndpointQuery struct {
	Action string `form:"action"`
}

func getEndpointsH(a lib.Api) gin.HandlerFunc {
	return func(gc *gin.Context) {
		query := endpointFilterQuery{}
		if err := gc.ShouldBindQuery(&query); err != nil {
			_ = gc.Error(lib_model.NewInvalidInputError(err))
			return
		}
		filter := lib_model.EndpointFilter{
			IDs:  parseStringSlice(query.IDs, ","),
			Type: query.Type,
			Ref:  query.Ref,
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
		switch query.Action {
		case "list":
			var endpointBaseSl []lib_model.EndpointBase
			if err = gc.ShouldBindJSON(&endpointBaseSl); err != nil {
				_ = gc.Error(lib_model.NewInvalidInputError(err))
				return
			}
			jID, err = a.AddEndpoints(gc.Request.Context(), endpointBaseSl)
			if err != nil {
				_ = gc.Error(err)
				return
			}
		case "alias":
			var aliasReq lib_model.EndpointAliasReq
			if err = gc.ShouldBindJSON(&aliasReq); err != nil {
				_ = gc.Error(lib_model.NewInvalidInputError(err))
				return
			}
			jID, err = a.AddEndpointAlias(gc.Request.Context(), aliasReq.ID, aliasReq.Path)
			if err != nil {
				_ = gc.Error(err)
				return
			}
		default:
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
		jID, err := a.RemoveEndpoint(gc.Request.Context(), gc.Param(endpointIdParam))
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.String(http.StatusOK, jID)
	}
}

func deleteEndpointsH(a lib.Api) gin.HandlerFunc {
	return func(gc *gin.Context) {
		query := endpointFilterQuery{}
		if err := gc.ShouldBindQuery(&query); err != nil {
			_ = gc.Error(lib_model.NewInvalidInputError(err))
			return
		}
		filter := lib_model.EndpointFilter{
			IDs:  parseStringSlice(query.IDs, ","),
			Type: query.Type,
			Ref:  query.Ref,
		}
		jID, err := a.RemoveEndpoints(gc.Request.Context(), filter)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.String(http.StatusOK, jID)
	}
}
