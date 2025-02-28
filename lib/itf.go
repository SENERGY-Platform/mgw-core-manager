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

package lib

import (
	"context"
	"github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	job_hdl_lib "github.com/SENERGY-Platform/mgw-go-service-base/job-hdl/lib"
	srv_info_lib "github.com/SENERGY-Platform/mgw-go-service-base/srv-info-hdl/lib"
	"io"
)

type Api interface {
	GetEndpoints(ctx context.Context, filter model.EndpointFilter) (map[string]model.Endpoint, error)
	GetEndpoint(ctx context.Context, id string) (model.Endpoint, error)
	SetEndpoint(ctx context.Context, endpoint model.EndpointBase) (string, error)
	SetEndpoints(ctx context.Context, endpoints []model.EndpointBase) (string, error)
	AddEndpointAlias(ctx context.Context, id, path string) (string, error)
	AddDefaultGuiEndpoint(ctx context.Context, id string) (string, error)
	RemoveEndpoint(ctx context.Context, id string, restrictStd bool) (string, error)
	RemoveEndpoints(ctx context.Context, filter model.EndpointFilter, restrictStd bool) (string, error)
	GetCoreServices(ctx context.Context) (map[string]model.CoreService, error)
	GetCoreService(ctx context.Context, name string) (model.CoreService, error)
	RestartCoreService(ctx context.Context, name string) (string, error)
	PurgeImages(ctx context.Context, repository, excludeTag string) (string, error)
	ListLogs(ctx context.Context) ([]model.Log, error)
	GetLog(ctx context.Context, id string, numOfLines int) (io.ReadCloser, error)
	job_hdl_lib.Api
	srv_info_lib.Api
}
