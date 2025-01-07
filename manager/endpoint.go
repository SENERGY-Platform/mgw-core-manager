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

package manager

import (
	"context"
	"fmt"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
)

func (m *Manager) GetEndpoints(ctx context.Context, filter lib_model.EndpointFilter) (map[string]lib_model.Endpoint, error) {
	return m.gwEndpointHdl.List(ctx, filter)
}

func (m *Manager) GetEndpoint(ctx context.Context, id string) (lib_model.Endpoint, error) {
	return m.gwEndpointHdl.Get(ctx, id)
}

func (m *Manager) SetEndpoint(ctx context.Context, endpoint lib_model.EndpointBase) (string, error) {
	return m.jobHandler.Create(ctx, fmt.Sprintf("set endpoint '%+v'", endpoint), func(ctx context.Context, cf context.CancelFunc) (any, error) {
		defer cf()
		err := m.gwEndpointHdl.Set(ctx, endpoint)
		if err == nil {
			err = ctx.Err()
		}
		return nil, err
	})
}

func (m *Manager) SetEndpoints(ctx context.Context, endpoints []lib_model.EndpointBase) (string, error) {
	return m.jobHandler.Create(ctx, fmt.Sprintf("set endpoints '%+v'", endpoints), func(ctx context.Context, cf context.CancelFunc) (any, error) {
		defer cf()
		err := m.gwEndpointHdl.SetList(ctx, endpoints)
		if err == nil {
			err = ctx.Err()
		}
		return nil, err
	})
}

func (m *Manager) AddEndpointAlias(ctx context.Context, id, path string) (string, error) {
	return m.jobHandler.Create(ctx, fmt.Sprintf("add alias for endpoint '%s'", id), func(ctx context.Context, cf context.CancelFunc) (any, error) {
		defer cf()
		err := m.gwEndpointHdl.AddAlias(ctx, id, path)
		if err == nil {
			err = ctx.Err()
		}
		return nil, err
	})
}

func (m *Manager) AddDefaultGuiEndpoint(ctx context.Context, id string) (string, error) {
	return m.jobHandler.Create(ctx, fmt.Sprintf("add endpoint '%s' as default gui", id), func(ctx context.Context, cf context.CancelFunc) (any, error) {
		defer cf()
		err := m.gwEndpointHdl.AddDefaultGui(ctx, id)
		if err == nil {
			err = ctx.Err()
		}
		return nil, err
	})
}

func (m *Manager) RemoveEndpoint(ctx context.Context, id string, restrictStd bool) (string, error) {
	return m.jobHandler.Create(ctx, fmt.Sprintf("remove endpoint '%s'", id), func(ctx context.Context, cf context.CancelFunc) (any, error) {
		defer cf()
		err := m.gwEndpointHdl.Remove(ctx, id, restrictStd)
		if err == nil {
			err = ctx.Err()
		}
		return nil, err
	})
}

func (m *Manager) RemoveEndpoints(ctx context.Context, filter lib_model.EndpointFilter, restrictStd bool) (string, error) {
	return m.jobHandler.Create(ctx, fmt.Sprintf("remove endpoints '%+v'", filter), func(ctx context.Context, cf context.CancelFunc) (any, error) {
		defer cf()
		err := m.gwEndpointHdl.RemoveAll(ctx, filter, restrictStd)
		if err == nil {
			err = ctx.Err()
		}
		return nil, err
	})
}
