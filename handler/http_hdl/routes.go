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

func SetRoutes(e *gin.Engine, a lib.Api) {
	e.GET(lib_model.EndpointsPath, getEndpointsH(a))
	e.POST(lib_model.EndpointsPath, postEndpointH(a))
	e.GET(lib_model.EndpointsPath+"/:"+endpointIdParam, getEndpointH(a))
	e.DELETE(lib_model.EndpointsPath+"/:"+endpointIdParam, deleteEndpointH(a))
	e.POST(lib_model.EndpointsBatchPath, postEndpointBatchH(a))
	e.DELETE(lib_model.EndpointsBatchPath, deleteEndpointBatchH(a))
	e.GET(lib_model.RestrictedPath+"/"+lib_model.EndpointsPath, getEndpointsH(a))
	e.POST(lib_model.RestrictedPath+"/"+lib_model.EndpointsPath, postEndpointRestrictedH(a))
	e.GET(lib_model.RestrictedPath+"/"+lib_model.EndpointsPath+"/:"+endpointIdParam, getEndpointH(a))
	e.DELETE(lib_model.RestrictedPath+"/"+lib_model.EndpointsPath+"/:"+endpointIdParam, deleteEndpointRestrictedH(a))
	e.GET("/"+lib_model.JobsPath, getJobsH(a))
	e.GET("/"+lib_model.JobsPath+"/:"+jobIdParam, getJobH(a))
	e.PATCH("/"+lib_model.JobsPath+"/:"+jobIdParam+"/"+lib_model.JobsCancelPath, patchJobCancelH(a))
}

func GetRoutes(e *gin.Engine) [][2]string {
	routes := e.Routes()
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Path < routes[j].Path
	})
	var rInfo [][2]string
	for _, info := range routes {
		rInfo = append(rInfo, [2]string{info.Method, info.Path})
	}
	return rInfo
}
