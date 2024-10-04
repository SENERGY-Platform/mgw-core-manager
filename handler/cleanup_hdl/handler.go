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

package cleanup_hdl

import (
	"context"
	"github.com/SENERGY-Platform/go-service-base/context-hdl"
	cew_lib "github.com/SENERGY-Platform/mgw-container-engine-wrapper/lib"
	cew_model "github.com/SENERGY-Platform/mgw-container-engine-wrapper/lib/model"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"net/url"
	"strings"
	"time"
)

type Handler struct {
	cewClient   cew_lib.Api
	httpTimeout time.Duration
}

func New(cewClient cew_lib.Api, httpTimeout time.Duration) *Handler {
	return &Handler{
		cewClient:   cewClient,
		httpTimeout: httpTimeout,
	}
}

func (h *Handler) PurgeImages(ctx context.Context, repository, excludeTag string) error {
	ch := context_hdl.New()
	defer ch.CancelAll()
	images, err := h.cewClient.GetImages(ch.Add(context.WithTimeout(ctx, h.httpTimeout)), cew_model.ImageFilter{Name: repository})
	if err != nil {
		return lib_model.NewInternalError(err)
	}
	for _, image := range images {
		if excludeTag != "" && inTags(image.Tags, excludeTag) {
			continue
		}
		err = h.cewClient.RemoveImage(ch.Add(context.WithTimeout(ctx, h.httpTimeout)), url.QueryEscape(image.ID))
		if err != nil {
			util.Logger.Error(err)
		}
	}
	return nil
}

func inTags(tags []string, s string) bool {
	for _, entry := range tags {
		if strings.HasSuffix(entry, s) {
			return true
		}
	}
	return false
}
