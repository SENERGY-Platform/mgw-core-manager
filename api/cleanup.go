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
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"strings"
)

func (a *Api) PurgeImages(ctx context.Context, repository, excludeTag string) (string, error) {
	return a.jobHandler.Create(ctx, fmt.Sprintf("purge images (repository=%s exclude_tag=%s)", repository, excludeTag), func(ctx context.Context, cf context.CancelFunc) (any, error) {
		defer cf()
		err := a.cleanupHdl.PurgeImages(ctx, repository, excludeTag)
		if err == nil {
			err = ctx.Err()
		}
		return nil, err
	})
}

func (a *Api) PurgeCoreImages() error {
	_, err := a.jobHandler.Create(context.Background(), fmt.Sprintf("purge core images"), func(ctx context.Context, cf context.CancelFunc) (any, error) {
		defer cf()
		err := a.purgeCoreImages(ctx)
		if err == nil {
			err = ctx.Err()
		}
		return nil, err
	})
	return err
}

func (a *Api) purgeCoreImages(ctx context.Context) error {
	services, err := a.coreSrvHdl.List(ctx)
	if err != nil {
		return err
	}
	for _, service := range services {
		parts := strings.Split(service.Image, ":")
		if len(parts) != 2 {
			util.Logger.Errorf("purge core images: malformed image string '%s'", service.Image)
			continue
		}
		if err = a.cleanupHdl.PurgeImages(ctx, parts[0], parts[1]); err != nil {
			util.Logger.Error("purge core images:", err)
		}
	}
	return nil
}
