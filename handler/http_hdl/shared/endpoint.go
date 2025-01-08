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

// GetEndpointsH
// @Summary List endpoints
// @Description	List HTTP endpoints.
// @Tags HTTP Endpoints
// @Produce	json
// @Param ids query string false "comma seperated list of endpoint ids (e.g.: id1,id2,...)"
// @Param ref query string false "reference value (e.g.: a foreign id)"
// @Param labels query string false "comma seperated list of labels (e.g.: key1=val1,key2=val2,...)"
// @Success	200 {object} map[string]lib_model.Endpoint "endpoints"
// @Failure	400 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /endpoints [get]
func GetEndpointsH(a lib.Api) (string, string, gin.HandlerFunc) {
	return http.MethodGet, lib_model.EndpointsPath, func(gc *gin.Context) {
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
	}
}

// GetEndpointH
// @Summary Get endpoint
// @Description	Get HTTP endpoint.
// @Tags HTTP Endpoints
// @Produce	json
// @Param id path string true "endpoint id"
// @Success	200 {object} map[string]lib_model.Endpoint "endpoints"
// @Failure	404 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /endpoints [get]
func GetEndpointH(a lib.Api) (string, string, gin.HandlerFunc) {
	return http.MethodGet, path.Join(lib_model.EndpointsPath, ":id"), func(gc *gin.Context) {
		endpoint, err := a.GetEndpoint(gc.Request.Context(), gc.Param("id"))
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.JSON(http.StatusOK, endpoint)
	}
}

// PostEndpointAliasH
// @Summary Create endpoint alias
// @Description	Create an endpoint alias.
// @Tags HTTP Endpoints
// @Accept json
// @Produce	plain
// @Param id path string true "endpoint id"
// @Param alias body lib_model.EndpointAliasReq false "endpoint alias information"
// @Success	200 {string} string "job ID"
// @Failure	400 {string} string "error message"
// @Failure	404 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /endpoints/{id}/alias [post]
func PostEndpointAliasH(a lib.Api) (string, string, gin.HandlerFunc) {
	return http.MethodPost, path.Join(lib_model.EndpointsPath, ":id", lib_model.AliasPath), func(gc *gin.Context) {
		var jID string
		var err error
		var aliasReq lib_model.EndpointAliasReq
		if err = gc.ShouldBindJSON(&aliasReq); err != nil {
			_ = gc.Error(lib_model.NewInvalidInputError(err))
			return
		}
		jID, err = a.AddEndpointAlias(gc.Request.Context(), gc.Param("id"), aliasReq.Path)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.String(http.StatusOK, jID)
	}
}
