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

package manager

import (
	"context"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"io"
)

type GatewayEndpointHandler interface {
	List(ctx context.Context, filter lib_model.EndpointFilter) (map[string]lib_model.Endpoint, error)
	Get(ctx context.Context, id string) (lib_model.Endpoint, error)
	Set(ctx context.Context, endpoint lib_model.EndpointBase) error
	SetList(ctx context.Context, endpoints []lib_model.EndpointBase) error
	AddAlias(ctx context.Context, id, path string) error
	AddDefaultGui(ctx context.Context, id string) error
	Remove(ctx context.Context, id string, restrictStd bool) error
	RemoveAll(ctx context.Context, filter lib_model.EndpointFilter, restrictStd bool) error
}

type CoreServiceHandler interface {
	List(ctx context.Context) (map[string]lib_model.CoreService, error)
	Get(ctx context.Context, name string) (lib_model.CoreService, error)
	Restart(ctx context.Context, name string) error
}

type CleanupHandler interface {
	PurgeImages(ctx context.Context, repository, excludeTag string) error
}

type LogHandler interface {
	List(ctx context.Context) ([]lib_model.Log, error)
	GetReader(ctx context.Context, id string, numOfLines int) (io.ReadCloser, error)
}
