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
	"fmt"
	cew_lib "github.com/SENERGY-Platform/mgw-container-engine-wrapper/lib"
	cew_model "github.com/SENERGY-Platform/mgw-container-engine-wrapper/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/handler"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

const (
	CoreIDLabel  = "mgw_cid"
	CoreSrvLabel = "core_srv"
)

type Handler struct {
	cewClient   cew_lib.Api
	services    map[string]service
	coreID      string
	httpTimeout time.Duration
}

type service struct {
	Name          string
	ContainerName string
	CtrHandler    handler.ContainerHandler
}

type composeFile struct {
	Services map[string]struct {
		ContainerName string `yaml:"container_name"`
	} `yaml:"services"`
}

func New(cewClient cew_lib.Api, coreID string, httpTimeout time.Duration) *Handler {
	return &Handler{
		cewClient:   cewClient,
		coreID:      coreID,
		httpTimeout: httpTimeout,
	}
}

func (h *Handler) Init(composePath string) error {
	file, err := os.Open(composePath)
	if err != nil {
		return err
	}
	decoder := yaml.NewDecoder(file)
	var cFile composeFile
	err = decoder.Decode(&cFile)
	if err != nil {
		return err
	}
	h.services = make(map[string]service)
	for name, srv := range cFile.Services {
		h.services[name] = service{
			ContainerName: srv.ContainerName,
			Name:          name,
			CtrHandler: &ctrHandler{
				cewClient:     h.cewClient,
				srvName:       name,
				containerName: srv.ContainerName,
				httpTimeout:   h.httpTimeout,
			},
		}
	}
	return nil
}

func (h *Handler) List(ctx context.Context) (map[string]lib_model.CoreService, error) {
	var ctrMap map[string]cew_model.Container
	ctxWt, cf := context.WithTimeout(ctx, h.httpTimeout)
	defer cf()
	ctrList, err := h.cewClient.GetContainers(ctxWt, cew_model.ContainerFilter{
		Labels: map[string]string{
			CoreIDLabel:  h.coreID,
			CoreSrvLabel: "true",
		},
	})
	if err != nil {
		util.Logger.Error(err)
	} else {
		ctrMap = make(map[string]cew_model.Container)
		for _, ctr := range ctrList {
			ctrMap[ctr.Name] = ctr
		}
	}
	services := make(map[string]lib_model.CoreService)
	for name, srv := range h.services {
		cs := lib_model.CoreService{
			Name:      name,
			Container: lib_model.SrvContainer{Name: srv.ContainerName},
		}
		ctr, ok := ctrMap[srv.ContainerName]
		if !ok {
			util.Logger.Errorf("service '%s' missing container", name)
		} else {
			cs.Container.ID = &ctr.ID
			cs.Container.State = &ctr.State
		}
		services[name] = cs
	}
	return services, nil
}

func (h *Handler) Get(ctx context.Context, name string) (lib_model.CoreService, error) {
	srv, ok := h.services[name]
	if !ok {
		return lib_model.CoreService{}, lib_model.NewNotFoundError(fmt.Errorf("service '%s' not found", name))
	}
	cs, err := srv.CtrHandler.Info(ctx)
	if err != nil {
		return lib_model.CoreService{}, lib_model.NewInternalError(err)
	}
	return cs, nil
}

func (h *Handler) Restart(ctx context.Context, name string) error {
	srv, ok := h.services[name]
	if !ok {
		return lib_model.NewNotFoundError(fmt.Errorf("service '%s' not found", name))
	}
	err := srv.CtrHandler.Restart(ctx)
	if err != nil {
		return lib_model.NewInternalError(err)
	}
	return nil
}

func (h *Handler) GetCtrHandler(name string) (handler.ContainerHandler, error) {
	srv, ok := h.services[name]
	if !ok {
		return nil, fmt.Errorf("service '%s' not defined", name)
	}
	return srv.CtrHandler, nil
}
