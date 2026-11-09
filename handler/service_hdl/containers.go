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
	"sort"

	cew_model "github.com/SENERGY-Platform/mgw-container-engine-wrapper/lib/model"
)

// LongRunningContainers returns the container names of the services that are expected to keep running, i.e. all
// but the one-shot services.
func (h *Handler) LongRunningContainers() []string {
	var names []string
	for _, srv := range h.services {
		if !srv.OneShot {
			names = append(names, srv.ContainerName)
		}
	}
	sort.Strings(names)
	return names
}

// ContainerStates returns the state of each existing core container by container name.
func (h *Handler) ContainerStates(ctx context.Context) (map[string]string, error) {
	ctrMap, err := h.coreContainers(ctx)
	if err != nil {
		return nil, err
	}
	states := make(map[string]string, len(ctrMap))
	for name, ctr := range ctrMap {
		states[name] = ctr.State
	}
	return states, nil
}

// StartContainer starts a stopped core container.
func (h *Handler) StartContainer(ctx context.Context, name string) error {
	ctxWt, cf := context.WithTimeout(ctx, h.httpTimeout)
	defer cf()
	return h.cewClient.StartContainer(ctxWt, name)
}

// coreContainers returns the containers labelled as core services of this core by container name.
func (h *Handler) coreContainers(ctx context.Context) (map[string]cew_model.Container, error) {
	ctxWt, cf := context.WithTimeout(ctx, h.httpTimeout)
	defer cf()
	ctrList, err := h.cewClient.GetContainers(ctxWt, cew_model.ContainerFilter{
		Labels: map[string]string{
			CoreIDLabel:  h.coreID,
			CoreSrvLabel: "true",
		},
	})
	if err != nil {
		return nil, err
	}
	ctrMap := make(map[string]cew_model.Container, len(ctrList))
	for _, ctr := range ctrList {
		ctrMap[ctr.Name] = ctr
	}
	return ctrMap, nil
}
