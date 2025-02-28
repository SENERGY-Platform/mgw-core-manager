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
// @Summary Create endpoint
// @Description	Create an HTTP endpoint accessible via the core reverse proxy.
// @Tags HTTP Endpoints
// @Accept json
// @Produce	plain
// @Param endpoint body lib_model.EndpointBase true "endpoint information"
// @Success	200 {string} string "job ID"
// @Failure	400 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /endpoints [post]
func PostEndpointH(a lib.Api) (string, string, gin.HandlerFunc) {
	return http.MethodPost, lib_model.EndpointsPath, func(gc *gin.Context) {
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
	}
}

// DeleteEndpointH
// @Summary Delete endpoint
// @Description	Remove an HTTP endpoint.
// @Tags HTTP Endpoints
// @Produce	plain
// @Param id path string true "endpoint id"
// @Success	200 {string} string "job ID"
// @Failure	500 {string} string "error message"
// @Router /endpoints/{id} [delete]
func DeleteEndpointH(a lib.Api) (string, string, gin.HandlerFunc) {
	return http.MethodDelete, path.Join(lib_model.EndpointsPath, ":id"), func(gc *gin.Context) {
		jID, err := a.RemoveEndpoint(gc.Request.Context(), gc.Param("id"), false)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.String(http.StatusOK, jID)
	}
}

// PostEndpointBatchH
// @Summary Create endpoints
// @Description	Create multiple HTTP endpoints accessible via the core reverse proxy.
// @Tags HTTP Endpoints
// @Accept json
// @Produce	plain
// @Param endpoints body []lib_model.EndpointBase true "list of endpoint information items"
// @Success	200 {string} string "job ID"
// @Failure	400 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /endpoints-batch [post]
func PostEndpointBatchH(a lib.Api) (string, string, gin.HandlerFunc) {
	return http.MethodPost, lib_model.EndpointsBatchPath, func(gc *gin.Context) {
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
	}
}

// DeleteEndpointBatchH
// @Summary Delete endpoints
// @Description	Remove multiple HTTP endpoints.
// @Tags HTTP Endpoints
// @Produce	plain
// @Param ids query string false "comma seperated list of endpoint ids (e.g.: id1,id2,...)"
// @Param ref query string false "reference value (e.g.: a foreign id)"
// @Param labels query string false "comma seperated list of labels (e.g.: key1=val1,key2=val2,...)"
// @Success	200 {string} string "job ID"
// @Failure	400 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /endpoints-batch [delete]
func DeleteEndpointBatchH(a lib.Api) (string, string, gin.HandlerFunc) {
	return http.MethodDelete, lib_model.EndpointsBatchPath, func(gc *gin.Context) {
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
	}
}
