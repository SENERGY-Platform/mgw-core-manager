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
	job_hdl_lib "github.com/SENERGY-Platform/go-service-base/job-hdl/lib"
	"github.com/SENERGY-Platform/mgw-core-manager/lib/model"
)

type Api interface {
	ListEndpoints(ctx context.Context) ([]model.Endpoint, error)
	AddEndpoints(ctx context.Context, endpoints []model.Endpoint) error
	RemoveEndpoint(ctx context.Context, id string) error
	RemoveEndpoints(ctx context.Context, ref string) error
	job_hdl_lib.Api
}
