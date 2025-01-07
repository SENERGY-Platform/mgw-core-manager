/*
 * Copyright 2023 InfAI (CC SES)
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
	"sort"
)

type routes []func(a lib.Api, rg *gin.RouterGroup)

func (r routes) Set(a lib.Api, rg *gin.RouterGroup) {
	for _, f := range r {
		f(a, rg)
	}
}

var shared = routes{
	setGetEndpointsH,
	setGetEndpointH,
	setGetCoreServicesH,
	setGetCoreServiceH,
	setPatchRestartCoreServiceH,
	setGetJobsH,
	setGetJobH,
	setPatchJobCancelH,
	setGetLogsH,
	setGetLogH,
	setGetSrvInfo,
}

var standard = routes{
	setPostEndpointH,
	setDeleteEndpointH,
	setPostEndpointBatchH,
	setDeleteEndpointBatchH,
	setPatchPurgeImagesH,
}

var restricted = routes{
	setPostEndpointRestrictedH,
	setDeleteEndpointRestrictedH,
	setDeleteEndpointBatchRestrictedH,
}

func setStdRoutes(e *gin.Engine, a lib.Api) {
	rg := e.Group("")
	shared.Set(a, rg)
	standard.Set(a, rg)
}

func setRstRoutes(e *gin.Engine, a lib.Api) {
	rg := e.Group(lib_model.RestrictedPath)
	shared.Set(a, rg)
	restricted.Set(a, rg)
}

func GetRoutes(e *gin.Engine) [][2]string {
	routesInfo := e.Routes()
	sort.Slice(routesInfo, func(i, j int) bool {
		return routesInfo[i].Path < routesInfo[j].Path
	})
	var rInfo [][2]string
	for _, info := range routesInfo {
		rInfo = append(rInfo, [2]string{info.Method, info.Path})
	}
	return rInfo
}
