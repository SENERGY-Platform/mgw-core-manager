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
