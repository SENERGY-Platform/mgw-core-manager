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

package api

import (
	"github.com/SENERGY-Platform/go-service-base/job-hdl"
	"github.com/SENERGY-Platform/go-service-base/srv-info-hdl"
	"github.com/SENERGY-Platform/mgw-core-manager/handler"
)

type Api struct {
	coreSrvHdl    handler.CoreServiceHandler
	gwEndpointHdl handler.GatewayEndpointHandler
	cleanupHdl    handler.CleanupHandler
	jobHandler    job_hdl.JobHandler
	srvInfoHdl    srv_info_hdl.SrvInfoHandler
}

func New(coreServiceHandler handler.CoreServiceHandler, gwEndpointHdl handler.GatewayEndpointHandler, cleanupHdl handler.CleanupHandler, jobHandler job_hdl.JobHandler, srvInfoHandler srv_info_hdl.SrvInfoHandler) *Api {
	return &Api{
		coreSrvHdl:    coreServiceHandler,
		gwEndpointHdl: gwEndpointHdl,
		cleanupHdl:    cleanupHdl,
		jobHandler:    jobHandler,
		srvInfoHdl:    srvInfoHandler,
	}
}
