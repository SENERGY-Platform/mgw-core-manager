/*
 * Copyright 2026 InfAI (CC SES)
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
)

// GetHealthH
// @Summary Get core health
// @Description	Get the health of the core services, the installed version and the status of the last update.
// @Tags Health
// @Produce	json
// @Success	200 {object} lib_model.CoreHealth "core health"
// @Failure	500 {string} string "error message"
// @Router /health [get]
func GetHealthH(a lib.Api) (string, string, gin.HandlerFunc) {
	return http.MethodGet, lib_model.HealthPath, func(gc *gin.Context) {
		health, err := a.GetCoreHealth(gc.Request.Context())
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.JSON(http.StatusOK, health)
	}
}
