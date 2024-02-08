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
	standardGrp := e.Group("")
	restrictedGrp := e.Group(lib_model.RestrictedPath)
	setCoreServiceRoutes(a, standardGrp.Group(lib_model.CoreServicesPath), restrictedGrp.Group(lib_model.CoreServicesPath))
	setJobsRoutes(a, standardGrp.Group(lib_model.JobsPath), restrictedGrp.Group(lib_model.JobsPath))
	standardEndpointsGrp := standardGrp.Group(lib_model.EndpointsPath)
	restrictedEndpointsGrp := restrictedGrp.Group(lib_model.EndpointsPath)
	setEndpointsSharedRoutes(a, standardEndpointsGrp, restrictedEndpointsGrp)
	setEndpointsRoutes(a, standardEndpointsGrp)
	setEndpointsRestrictedRoutes(a, restrictedEndpointsGrp)
	setEndpointsBatchRoutes(a, standardGrp.Group(lib_model.EndpointsBatchPath))
	setEndpointsBatchRestrictedRoutes(a, restrictedGrp.Group(lib_model.EndpointsBatchPath))
}

func setCoreServiceRoutes(a lib.Api, rGroups ...*gin.RouterGroup) {
	for _, rg := range rGroups {
		rg.GET("", getCoreServicesH(a))
		rg.GET(":"+coreSrvNameParam, getCoreServiceH(a))
		rg.PATCH(":"+coreSrvNameParam+"/"+lib_model.RestartPath, patchRestartCoreServiceH(a))
	}
}

func setJobsRoutes(a lib.Api, rGroups ...*gin.RouterGroup) {
	for _, rg := range rGroups {
		rg.GET("", getJobsH(a))
		rg.GET(":"+jobIdParam, getJobH(a))
		rg.PATCH(":"+jobIdParam+"/"+lib_model.JobsCancelPath, patchJobCancelH(a))
	}
}

func setEndpointsSharedRoutes(a lib.Api, rGroups ...*gin.RouterGroup) {
	for _, rg := range rGroups {
		rg.GET("", getEndpointsH(a))
		rg.GET(":"+endpointIdParam, getEndpointH(a))
	}
}

func setEndpointsRoutes(a lib.Api, rg *gin.RouterGroup) {
	rg.POST("", postEndpointH(a))
	rg.DELETE(":"+endpointIdParam, deleteEndpointH(a))
}

func setEndpointsRestrictedRoutes(a lib.Api, rg *gin.RouterGroup) {
	rg.POST("", postEndpointRestrictedH(a))
	rg.DELETE(":"+endpointIdParam, deleteEndpointRestrictedH(a))
}

func setEndpointsBatchRoutes(a lib.Api, rg *gin.RouterGroup) {
	rg.POST("", postEndpointBatchH(a))
	rg.DELETE("", deleteEndpointBatchH(a))
}

func setEndpointsBatchRestrictedRoutes(a lib.Api, rg *gin.RouterGroup) {
	rg.DELETE("", deleteEndpointBatchRestrictedH(a))
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
