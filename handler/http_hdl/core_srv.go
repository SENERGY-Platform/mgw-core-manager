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
	"github.com/gin-gonic/gin"
	"net/http"
)

const coreSrvNameParam = "n"

func getCoreServicesH(a lib.Api) gin.HandlerFunc {
	return func(gc *gin.Context) {
		endpoint, err := a.GetCoreServices(gc.Request.Context())
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.JSON(http.StatusOK, endpoint)
	}
}

func getCoreServiceH(a lib.Api) gin.HandlerFunc {
	return func(gc *gin.Context) {
		endpoint, err := a.GetCoreService(gc.Request.Context(), gc.Param(coreSrvNameParam))
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.JSON(http.StatusOK, endpoint)
	}
}

func patchRestartCoreServiceH(a lib.Api) gin.HandlerFunc {
	return func(gc *gin.Context) {
		jID, err := a.RestartCoreService(gc.Request.Context(), gc.Param(coreSrvNameParam))
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.String(http.StatusOK, jID)
	}
}
