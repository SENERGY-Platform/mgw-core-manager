/*
 * Copyright 2026 InfAI (CC SES)
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
	"bytes"
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"net/http"
	"net/url"
)

func (c *Client) GetOIDCSettings(ctx context.Context) (model.OIDCSettings, error) {
	u, err := url.JoinPath(c.baseUrl, model.RestrictedPath, model.OIDCPath)
	if err != nil {
		return model.OIDCSettings{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return model.OIDCSettings{}, err
	}
	var settings model.OIDCSettings
	err = c.baseClient.ExecRequestJSON(req, &settings)
	if err != nil {
		return model.OIDCSettings{}, err
	}
	return settings, nil
}

func (c *Client) SetOIDCSettings(ctx context.Context, settings model.OIDCSettingsReq) (string, error) {
	u, err := url.JoinPath(c.baseUrl, model.RestrictedPath, model.OIDCPath)
	if err != nil {
		return "", err
	}
	body, err := json.Marshal(settings)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	return c.baseClient.ExecRequestString(req)
}
