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

package service_hdl

import (
	"context"
	cew_model "github.com/SENERGY-Platform/mgw-container-engine-wrapper/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/handler/state_hdl"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"sort"
)

func (h *Handler) Health(ctx context.Context) (lib_model.CoreHealth, error) {
	ctrMap, err := h.coreContainers(ctx)
	if err != nil {
		util.Logger.Error(err)
	}
	health := lib_model.CoreHealth{
		Services: make([]lib_model.CoreServiceHealth, 0, len(h.services)),
	}
	for _, srv := range h.services {
		var ctr *cew_model.Container
		if c, ok := ctrMap[srv.ContainerName]; ok {
			ctr = &c
		} else if ctrMap != nil {
			util.Logger.Errorf("service '%s' missing container '%s'", srv.Name, srv.ContainerName)
		}
		health.Services = append(health.Services, serviceHealth(srv, ctr))
	}
	sort.Slice(health.Services, func(i, j int) bool {
		return health.Services[i].Name < health.Services[j].Name
	})
	health.State = deriveState(ctrMap != nil, health.Services)
	health.Version = state_hdl.ReadVersion(h.statePaths.Version)
	health.Update = state_hdl.ReadUpdate(h.statePaths.Update)
	health.Check = state_hdl.ReadCheck(h.statePaths.Check)
	health.Lock = state_hdl.ReadLock(h.statePaths.Lock)
	return health, nil
}

// serviceHealth derives the health of a service from its container, nil meaning the container is missing.
func serviceHealth(srv service, ctr *cew_model.Container) lib_model.CoreServiceHealth {
	sh := lib_model.CoreServiceHealth{Name: srv.Name, OneShot: srv.OneShot}
	if ctr == nil {
		return sh
	}
	state := ctr.State
	sh.State = &state
	if ctr.Health != nil {
		health := *ctr.Health
		sh.Health = &health
	}
	if srv.OneShot {
		// Stopped is the normal end state of a one-shot service and the container engine wrapper exposes no exit code.
		sh.Ok = state != cew_model.RestartingState && state != cew_model.DeadState
	} else {
		sh.Ok = state == cew_model.RunningState && (sh.Health == nil || *sh.Health != cew_model.UnhealthyState)
	}
	return sh
}

// deriveState returns undetermined if the container engine was not reachable, healthy if all services are ok and unhealthy otherwise.
func deriveState(reachable bool, services []lib_model.CoreServiceHealth) int {
	if !reachable {
		return lib_model.HealthUndetermined
	}
	for _, s := range services {
		if !s.Ok {
			return lib_model.HealthUnhealthy
		}
	}
	return lib_model.HealthHealthy
}
