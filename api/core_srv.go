/*
 * Copyright 2024 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/mgw-core-manager/lib/model"
)

func (a *Api) GetCoreServices(ctx context.Context) (map[string]model.CoreService, error) {
	return a.coreSrvHdl.List(ctx)
}

func (a *Api) GetCoreService(ctx context.Context, name string) (model.CoreService, error) {
	return a.coreSrvHdl.Get(ctx, name)
}

func (a *Api) RestartCoreService(ctx context.Context, name string) (string, error) {
	return a.jobHandler.Create(ctx, fmt.Sprintf("restart core service '%s'", name), func(ctx context.Context, cf context.CancelFunc) (any, error) {
		defer cf()
		err := a.coreSrvHdl.Restart(ctx, name)
		if err == nil {
			err = ctx.Err()
		}
		return nil, err
	})
}
