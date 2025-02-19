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

package manager

import (
	"context"
	"errors"
	"fmt"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"time"
)

func (m *Manager) PurgeImages(ctx context.Context, repository, excludeTag string) (string, error) {
	if repository == "" {
		return "", lib_model.NewInvalidInputError(errors.New("missing repository"))
	}
	return m.jobHandler.Create(ctx, fmt.Sprintf("purge images (repository=%s exclude_tag=%s)", repository, excludeTag), func(ctx context.Context, cf context.CancelFunc) (any, error) {
		defer cf()
		err := m.cleanupHdl.PurgeImages(ctx, repository, excludeTag)
		if err == nil {
			err = ctx.Err()
		}
		return nil, err
	})
}

func (m *Manager) PurgeCoreImages(delay time.Duration) error {
	_, err := m.jobHandler.Create(context.Background(), fmt.Sprintf("purge old core images (delay=%d)", delay), func(ctx context.Context, cf context.CancelFunc) (any, error) {
		defer cf()
		err := m.purgeCoreImages(ctx, delay)
		if err == nil {
			err = ctx.Err()
		}
		return nil, err
	})
	return err
}

func (m *Manager) purgeCoreImages(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	select {
	case <-timer.C:
		break
	case <-ctx.Done():
		return ctx.Err()
	}
	defer func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}()
	services, err := m.coreSrvHdl.List(ctx)
	if err != nil {
		return err
	}
	for _, service := range services {
		if err = m.cleanupHdl.PurgeImages(ctx, service.Image.Repository, service.Image.Tag); err != nil {
			util.Logger.Error("purge core images:", err)
		}
	}
	return nil
}
