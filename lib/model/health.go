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

package model

const (
	HealthUndetermined = 0
	HealthHealthy      = 1
	HealthUnhealthy    = 2
)

// CoreHealth describes the health of the core services and the state of the last update.
type CoreHealth struct {
	State    int                 `json:"state"`    // 1 healthy, 2 unhealthy, 0 undetermined (container engine not reachable)
	Version  string              `json:"version"`  // content of .version, empty if not readable
	Services []CoreServiceHealth `json:"services"` // all services of the compose file
	Update   *UpdateStatus       `json:"update"`   // null if no .update_status exists or it cannot be parsed
}

// CoreServiceHealth describes the health of a single core service.
type CoreServiceHealth struct {
	Name    string  `json:"name"`
	State   *string `json:"state"`    // container state from the container engine wrapper, null if the container is missing or the wrapper is unreachable
	Health  *string `json:"health"`   // container health from the container engine wrapper, null if the image defines no healthcheck
	OneShot bool    `json:"one_shot"` // service runs once and is expected to stop (compose restart policy on-failure)
	Ok      bool    `json:"ok"`
}

// UpdateStatus is the content of .update_status as written by the update script. Any field may be empty.
type UpdateStatus struct {
	FromVersion string `json:"from_version"`
	ToVersion   string `json:"to_version"`
	Started     string `json:"started"`
	Finished    string `json:"finished"`
	Result      string `json:"result"` // success or failed
	Step        string `json:"step"`   // last step the update reached
}
