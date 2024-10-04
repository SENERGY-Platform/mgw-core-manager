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

package client

import (
	"context"
	"github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) PurgeImages(ctx context.Context, repository, excludeTag string) (string, error) {
	u, err := url.JoinPath(c.baseUrl, model.CleanupPath, model.ImagesPath)
	if err != nil {
		return "", err
	}
	u += genPurgeImagesQuery(repository, excludeTag)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, u, nil)
	if err != nil {
		return "", err
	}
	return c.baseClient.ExecRequestString(req)
}

func genPurgeImagesQuery(repository, excludeTag string) string {
	var q []string
	if repository != "" {
		q = append(q, "repository="+repository)
	}
	if excludeTag != "" {
		q = append(q, "exclude_tag="+excludeTag)
	}
	if len(q) > 0 {
		return "?" + strings.Join(q, "&")
	}
	return ""
}
