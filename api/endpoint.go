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
	"context"
	"fmt"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
)

func (a *Api) GetEndpoints(ctx context.Context, filter lib_model.EndpointFilter) (map[string]lib_model.Endpoint, error) {
	return a.gwEndpointHdl.List(ctx, filter)
}

func (a *Api) GetEndpoint(ctx context.Context, id string) (lib_model.Endpoint, error) {
	return a.gwEndpointHdl.Get(ctx, id)
}

func (a *Api) AddEndpoint(ctx context.Context, endpoint lib_model.EndpointBase) (string, error) {
	return a.jobHandler.Create(ctx, fmt.Sprintf("add endpoint '%+v'", endpoint), func(ctx context.Context, cf context.CancelFunc) error {
		defer cf()
		err := a.gwEndpointHdl.Add(ctx, endpoint)
		if err == nil {
			err = ctx.Err()
		}
		return err
	})
}

func (a *Api) AddEndpoints(ctx context.Context, endpoints []lib_model.EndpointBase) (string, error) {
	return a.jobHandler.Create(ctx, fmt.Sprintf("add endpoints '%+v'", endpoints), func(ctx context.Context, cf context.CancelFunc) error {
		defer cf()
		err := a.gwEndpointHdl.AddList(ctx, endpoints)
		if err == nil {
			err = ctx.Err()
		}
		return err
	})
}

func (a *Api) AddEndpointAlias(ctx context.Context, id, path string) (string, error) {
	return a.jobHandler.Create(ctx, fmt.Sprintf("add alias for endpoint '%s'", id), func(ctx context.Context, cf context.CancelFunc) error {
		defer cf()
		err := a.gwEndpointHdl.AddAlias(ctx, id, path)
		if err == nil {
			err = ctx.Err()
		}
		return err
	})
}

func (a *Api) AddDefaultGuiEndpoint(ctx context.Context, id string) (string, error) {
	return a.jobHandler.Create(ctx, fmt.Sprintf("add endpoint '%s' as default gui", id), func(ctx context.Context, cf context.CancelFunc) error {
		defer cf()
		err := a.gwEndpointHdl.AddDefaultGui(ctx, id)
		if err == nil {
			err = ctx.Err()
		}
		return err
	})
}

func (a *Api) RemoveEndpoint(ctx context.Context, id string, restrictStd bool) (string, error) {
	return a.jobHandler.Create(ctx, fmt.Sprintf("remove endpoint '%s'", id), func(ctx context.Context, cf context.CancelFunc) error {
		defer cf()
		err := a.gwEndpointHdl.Remove(ctx, id, restrictStd)
		if err == nil {
			err = ctx.Err()
		}
		return err
	})
}

func (a *Api) RemoveEndpointsByRef(ctx context.Context, ref string) (string, error) {
	return a.jobHandler.Create(ctx, fmt.Sprintf("remove endpoints by ref '%s'", ref), func(ctx context.Context, cf context.CancelFunc) error {
		defer cf()
		err := a.gwEndpointHdl.RemoveByRef(ctx, ref)
		if err == nil {
			err = ctx.Err()
		}
		return err
	})
}
