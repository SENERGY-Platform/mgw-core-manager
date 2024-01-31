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

package service_hdl

import (
	"context"
	"errors"
	job_hdl_lib "github.com/SENERGY-Platform/go-service-base/job-hdl/lib"
	cew_lib "github.com/SENERGY-Platform/mgw-container-engine-wrapper/lib"
	cew_model "github.com/SENERGY-Platform/mgw-container-engine-wrapper/lib/model"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"net/http"
	"sync"
	"time"
)

type ctrHandler struct {
	cewClient     cew_lib.Api
	srvName       string
	containerName string
	httpTimeout   time.Duration
	mu            sync.Mutex
}

func (h *ctrHandler) Info(ctx context.Context) (lib_model.CoreService, error) {
	cs := lib_model.CoreService{
		Name:      h.srvName,
		Container: lib_model.SrvContainer{Name: h.containerName},
	}
	ctxWt, cf := context.WithTimeout(ctx, h.httpTimeout)
	defer cf()
	ctr, err := h.cewClient.GetContainer(ctxWt, h.containerName)
	if err != nil {
		util.Logger.Error(err)
	} else {
		cs.Container.ID = &ctr.ID
		cs.Container.State = &ctr.State
	}
	return cs, nil
}

func (h *ctrHandler) Restart(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	ctxWt, cf := context.WithTimeout(ctx, h.httpTimeout)
	defer cf()
	jID, err := h.cewClient.RestartContainer(ctxWt, h.containerName)
	if err != nil {
		return lib_model.NewInternalError(err)
	}
	return h.awaitJob(ctx, jID)
}

func (h *ctrHandler) ExecCmd(ctx context.Context, cmd []string, tty bool, envVars map[string]string, workDir string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	ctxWt, cf := context.WithTimeout(ctx, h.httpTimeout)
	defer cf()
	jID, err := h.cewClient.ContainerExec(ctxWt, h.containerName, cew_model.ExecConfig{
		Tty:     tty,
		EnvVars: envVars,
		WorkDir: workDir,
		Cmd:     cmd,
	})
	if err != nil {
		return lib_model.NewInternalError(err)
	}
	return h.awaitJob(ctx, jID)
}

func (h *ctrHandler) awaitJob(ctx context.Context, jID string) error {
	job, err := job_hdl_lib.Await(ctx, h.cewClient, jID, time.Second, h.httpTimeout, util.Logger)
	if err != nil {
		return lib_model.NewInternalError(err)
	}
	if job.Error != nil {
		if job.Error.Code != nil && *job.Error.Code == http.StatusNotFound {
			return lib_model.NewNotFoundError(errors.New(job.Error.Message))
		}
		return lib_model.NewInternalError(errors.New(job.Error.Message))
	}
	return nil
}
