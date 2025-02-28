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
	job_hdl_lib "github.com/SENERGY-Platform/mgw-go-service-base/job-hdl/lib"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (c *Client) GetJobs(ctx context.Context, filter job_hdl_lib.JobFilter) ([]job_hdl_lib.Job, error) {
	u, err := url.JoinPath(c.baseUrl, model.JobsPath)
	if err != nil {
		return nil, err
	}
	u += genJobsFilter(filter)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	var jobs []job_hdl_lib.Job
	err = c.baseClient.ExecRequestJSON(req, &jobs)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

func (c *Client) GetJob(ctx context.Context, id string) (job_hdl_lib.Job, error) {
	u, err := url.JoinPath(c.baseUrl, model.JobsPath, id)
	if err != nil {
		return job_hdl_lib.Job{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return job_hdl_lib.Job{}, err
	}
	var job job_hdl_lib.Job
	err = c.baseClient.ExecRequestJSON(req, &job)
	if err != nil {
		return job_hdl_lib.Job{}, err
	}
	return job, nil
}

func (c *Client) CancelJob(ctx context.Context, id string) error {
	u, err := url.JoinPath(c.baseUrl, model.JobsPath, id, model.JobsCancelPath)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, u, nil)
	if err != nil {
		return err
	}
	return c.baseClient.ExecRequestVoid(req)
}

func genJobsFilter(filter job_hdl_lib.JobFilter) string {
	var q []string
	if filter.SortDesc {
		q = append(q, "sort_desc=true")
	}
	if filter.Status != "" {
		q = append(q, "status="+filter.Status)
	}
	if !filter.Since.IsZero() {
		q = append(q, "since="+filter.Since.Format(time.RFC3339Nano))
	}
	if !filter.Until.IsZero() {
		q = append(q, "until="+filter.Until.Format(time.RFC3339Nano))
	}
	if len(q) > 0 {
		return "?" + strings.Join(q, "&")
	}
	return ""
}
