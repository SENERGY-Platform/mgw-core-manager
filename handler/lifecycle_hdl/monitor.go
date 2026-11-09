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

package lifecycle_hdl

import (
	"context"
	"time"

	cew_model "github.com/SENERGY-Platform/mgw-container-engine-wrapper/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
)

// downStates are the states in which a long-running container is started again. A paused container was paused on
// purpose; restarting and removing are transitions someone else is driving.
var downStates = map[string]struct{}{
	cew_model.StoppedState: {},
	cew_model.DeadState:    {},
	cew_model.InitState:    {},
}

// monitorState is what the monitor remembers between checks.
type monitorState struct {
	down        map[string]bool        // seen down in the previous check
	restarts    map[string][]time.Time // restarts within the window
	gaveUp      map[string]bool        // restart limit reached, left alone until the core-manager restarts
	missing     map[string]bool        // reported missing
	unreachable bool                   // the wrapper was unreachable in the previous check
}

// monitor checks the long-running core containers every CheckInterval until ctx ends and starts those that are
// down. One-shot services are left alone: compose start ran them. A container is only started if it was down in two
// checks in a row, so a restart through the API, which passes through stopped, isn't interfered with.
func (h *Handler) monitor(ctx context.Context) {
	st := monitorState{
		down:     make(map[string]bool),
		restarts: make(map[string][]time.Time),
		gaveUp:   make(map[string]bool),
		missing:  make(map[string]bool),
	}
	names := h.ctrs.LongRunningContainers()
	for {
		if h.sleep(ctx, h.cfg.CheckInterval) != nil {
			return
		}
		h.check(ctx, names, &st)
	}
}

func (h *Handler) check(ctx context.Context, names []string, st *monitorState) {
	states, err := h.ctrs.ContainerStates(ctx)
	if err != nil {
		if ctx.Err() == nil && !st.unreachable {
			util.Logger.Warningf("checking core containers failed: %s", err)
		}
		st.unreachable = true
		return
	}
	if st.unreachable {
		util.Logger.Info("checking core containers works again")
		st.unreachable = false
	}
	for _, name := range names {
		state, ok := states[name]
		if !ok {
			if !st.missing[name] {
				util.Logger.Errorf("core container '%s' is missing", name)
				st.missing[name] = true
			}
			st.down[name] = false
			continue
		}
		st.missing[name] = false
		if _, isDown := downStates[state]; !isDown {
			st.down[name] = false
			continue
		}
		if !st.down[name] {
			st.down[name] = true // act on the next check if it's still down
			continue
		}
		h.restart(ctx, name, state, st)
	}
}

func (h *Handler) restart(ctx context.Context, name, state string, st *monitorState) {
	if st.gaveUp[name] {
		return
	}
	now := h.now()
	var recent []time.Time
	for _, t := range st.restarts[name] {
		if now.Sub(t) < h.cfg.RestartWindow {
			recent = append(recent, t)
		}
	}
	if len(recent) >= h.cfg.RestartLimit {
		util.Logger.Errorf("core container '%s' is %s after %d restarts within %s, giving up on it", name, state, len(recent), h.cfg.RestartWindow)
		st.gaveUp[name] = true
		return
	}
	st.restarts[name] = append(recent, now)
	st.down[name] = false
	util.Logger.Warningf("core container '%s' is %s, starting it", name, state)
	if err := h.ctrs.StartContainer(ctx, name); err != nil && ctx.Err() == nil {
		util.Logger.Warningf("starting core container '%s' failed: %s", name, err)
	}
}
