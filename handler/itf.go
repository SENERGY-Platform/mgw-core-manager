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

package handler

import (
	"context"
	"github.com/SENERGY-Platform/mgw-core-manager/lib/model"
)

type GatewayEndpointHandler interface {
	List(ctx context.Context) ([]model.Endpoint, error)
	Add(ctx context.Context, endpoints []model.Endpoint) error
	Remove(ctx context.Context, id string) error
	RemoveAll(ctx context.Context, ref string) error
}

type JobHandler interface {
	List(filter model.JobFilter) []model.Job
	Get(id string) (model.Job, error)
	Create(desc string, tFunc func(context.Context, context.CancelFunc) error) (string, error)
	Cancel(id string) error
}
