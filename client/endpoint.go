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

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (c *Client) GetEndpoints(ctx context.Context, filter model.EndpointFilter) (map[string]model.Endpoint, error) {
	u, err := url.JoinPath(c.baseUrl, model.EndpointsPath)
	if err != nil {
		return nil, err
	}
	u += genGetEndpointsQuery(filter)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	endpoints := make(map[string]model.Endpoint)
	err = c.baseClient.ExecRequestJSON(req, &endpoints)
	if err != nil {
		return nil, err
	}
	return endpoints, nil
}

func (c *Client) GetEndpoint(ctx context.Context, id string) (model.Endpoint, error) {
	u, err := url.JoinPath(c.baseUrl, model.EndpointsPath, id)
	if err != nil {
		return model.Endpoint{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return model.Endpoint{}, err
	}
	var endpoint model.Endpoint
	err = c.baseClient.ExecRequestJSON(req, &endpoint)
	if err != nil {
		return model.Endpoint{}, err
	}
	return endpoint, nil
}

func (c *Client) AddEndpoint(ctx context.Context, endpoint model.EndpointBase) (string, error) {
	u, err := url.JoinPath(c.baseUrl, model.EndpointsPath)
	if err != nil {
		return "", err
	}
	body, err := json.Marshal(endpoint)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	return c.baseClient.ExecRequestString(req)
}

func (c *Client) AddEndpoints(ctx context.Context, endpoints []model.EndpointBase) (string, error) {
	u, err := url.JoinPath(c.baseUrl, model.EndpointsBatchPath)
	if err != nil {
		return "", err
	}
	body, err := json.Marshal(endpoints)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	return c.baseClient.ExecRequestString(req)
}

func (c *Client) AddEndpointAlias(ctx context.Context, id, path string) (string, error) {
	u, err := url.JoinPath(c.baseUrl, model.EndpointsPath)
	if err != nil {
		return "", err
	}
	u += "?action=alias"
	body, err := json.Marshal(model.EndpointAliasReq{
		ParentID: id,
		Path:     path,
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	return c.baseClient.ExecRequestString(req)
}

func (c *Client) AddDefaultGuiEndpoint(ctx context.Context, id string) (string, error) {
	u, err := url.JoinPath(c.baseUrl, model.EndpointsPath)
	if err != nil {
		return "", err
	}
	u += "?action=alias"
	body, err := json.Marshal(model.EndpointAliasReq{
		ParentID: id,
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	return c.baseClient.ExecRequestString(req)
}

func (c *Client) RemoveEndpoint(ctx context.Context, id string, _ bool) (string, error) {
	u, err := url.JoinPath(c.baseUrl, model.EndpointsPath, id)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return "", err
	}
	return c.baseClient.ExecRequestString(req)
}

func (c *Client) RemoveEndpoints(ctx context.Context, filter model.EndpointFilter, _ bool) (string, error) {
	u, err := url.JoinPath(c.baseUrl, model.EndpointsBatchPath)
	if err != nil {
		return "", err
	}
	u += genGetEndpointsQuery(filter)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return "", err
	}
	return c.baseClient.ExecRequestString(req)
}

func genGetEndpointsQuery(filter model.EndpointFilter) string {
	var q []string
	if filter.Type != nil {
		q = append(q, "type="+strconv.FormatInt(int64(*filter.Type), 10))
	}
	if filter.Ref != "" {
		q = append(q, "ref="+filter.Ref)
	}
	if len(filter.Labels) > 0 {
		q = append(q, "labels="+genLabels(filter.Labels, "=", ","))
	}
	if len(filter.IDs) > 0 {
		q = append(q, "ids"+strings.Join(filter.IDs, ","))
	}
	if len(q) > 0 {
		return "?" + strings.Join(q, "&")
	}
	return ""
}

func genLabels(m map[string]string, eqs, sep string) string {
	var sl []string
	for k, v := range m {
		sl = append(sl, k+eqs+v)
	}
	return strings.Join(sl, sep)
}
