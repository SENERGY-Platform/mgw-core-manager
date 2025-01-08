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
	"github.com/SENERGY-Platform/mgw-core-manager/lib"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"path"
)

// GetCoreServicesH
// @Summary List services
// @Description	List core services including image and container information.
// @Tags Core Services
// @Produce	json
// @Success	200 {object} map[string]lib_model.CoreService "services"
// @Failure	500 {string} string "error message"
// @Router /core-services [get]
func GetCoreServicesH(a lib.Api) (string, string, gin.HandlerFunc) {
	return http.MethodGet, lib_model.CoreServicesPath, func(gc *gin.Context) {
		services, err := a.GetCoreServices(gc.Request.Context())
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.JSON(http.StatusOK, services)
	}
}

// GetCoreServiceH
// @Summary Get service
// @Description	Get core service including image and container information.
// @Tags Core Services
// @Produce	json
// @Param name path string true "service name"
// @Success	200 {object} lib_model.CoreService "service"
// @Failure	500 {string} string "error message"
// @Router /core-services/{name} [get]
func GetCoreServiceH(a lib.Api) (string, string, gin.HandlerFunc) {
	return http.MethodGet, path.Join(lib_model.CoreServicesPath, ":name"), func(gc *gin.Context) {
		service, err := a.GetCoreService(gc.Request.Context(), gc.Param("name"))
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.JSON(http.StatusOK, service)
	}
}

// PatchRestartCoreServiceH
// @Summary Restart service
// @Description	Restart core service container.
// @Tags Core Services
// @Produce	plain
// @Param name path string true "service name"
// @Success	200 {string} string "job ID"
// @Failure	500 {string} string "error message"
// @Router /core-services/{name}/restart [patch]
func PatchRestartCoreServiceH(a lib.Api) (string, string, gin.HandlerFunc) {
	return http.MethodPatch, path.Join(lib_model.CoreServicesPath, ":name", lib_model.RestartPath), func(gc *gin.Context) {
		jID, err := a.RestartCoreService(gc.Request.Context(), gc.Param("name"))
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.String(http.StatusOK, jID)
	}
}
