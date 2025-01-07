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

package http_hdl

import (
	gin_mw "github.com/SENERGY-Platform/gin-middleware"
	"github.com/SENERGY-Platform/mgw-core-manager/lib"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

func New(a lib.Api, staticHeader map[string]string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	httpHandler := gin.New()
	httpHandler.Use(gin_mw.StaticHeaderHandler(staticHeader), requestid.New(requestid.WithCustomHeaderStrKey(lib_model.HeaderRequestID)), gin_mw.LoggerHandler(util.Logger, nil, func(gc *gin.Context) string {
		return requestid.Get(gc)
	}), gin_mw.ErrorHandler(util.GetStatusCode, ", "), gin.Recovery())
	httpHandler.UseRawPath = true
	setStdRoutes(httpHandler, a)
	setRstRoutes(httpHandler, a)
	return httpHandler
}
