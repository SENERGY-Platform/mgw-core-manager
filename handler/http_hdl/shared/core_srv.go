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

func SetGetCoreServicesH(a lib.Api, rg *gin.RouterGroup) {
	rg.GET(lib_model.CoreServicesPath, func(gc *gin.Context) {
		services, err := a.GetCoreServices(gc.Request.Context())
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.JSON(http.StatusOK, services)
	})
}

func SetGetCoreServiceH(a lib.Api, rg *gin.RouterGroup) {
	rg.GET(path.Join(lib_model.CoreServicesPath, ":name"), func(gc *gin.Context) {
		service, err := a.GetCoreService(gc.Request.Context(), gc.Param("name"))
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.JSON(http.StatusOK, service)
	})
}

func SetPatchRestartCoreServiceH(a lib.Api, rg *gin.RouterGroup) {
	rg.PATCH(path.Join(lib_model.CoreServicesPath, ":name", lib_model.RestartPath), func(gc *gin.Context) {
		jID, err := a.RestartCoreService(gc.Request.Context(), gc.Param("name"))
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.String(http.StatusOK, jID)
	})
}
