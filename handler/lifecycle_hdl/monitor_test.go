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
	"errors"
	"reflect"
	"testing"
	"time"
)

// fakeContainers answers with the states set by the test; a started container becomes running unless startErr is
// set.
type fakeContainers struct {
	names    []string
	states   map[string]string
	err      error
	startErr error
	started  []string
}

func (f *fakeContainers) LongRunningContainers() []string { return f.names }

func (f *fakeContainers) ContainerStates(context.Context) (map[string]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	states := make(map[string]string, len(f.states))
	for k, v := range f.states {
		states[k] = v
	}
	return states, nil
}

func (f *fakeContainers) StartContainer(_ context.Context, name string) error {
	f.started = append(f.started, name)
	if f.startErr != nil {
		return f.startErr
	}
	f.states[name] = "running"
	return nil
}

// monitorFixture runs single checks with a clock the test moves.
type monitorFixture struct {
	h     *Handler
	ctrs  *fakeContainers
	st    monitorState
	clock time.Time
}

func newMonitorFixture(states map[string]string) *monitorFixture {
	ctrs := &fakeContainers{names: []string{"a", "b"}, states: states}
	m := &monitorFixture{ctrs: ctrs, clock: time.Date(2026, 11, 9, 3, 0, 0, 0, time.UTC)}
	m.h = New(Config{CheckInterval: checkInterval, RestartLimit: 3, RestartWindow: 10 * time.Minute}, ctrs)
	m.h.now = func() time.Time { return m.clock }
	m.st = monitorState{down: map[string]bool{}, restarts: map[string][]time.Time{}, gaveUp: map[string]bool{}, missing: map[string]bool{}}
	return m
}

func (m *monitorFixture) check(n int) {
	for i := 0; i < n; i++ {
		m.h.check(context.Background(), m.ctrs.names, &m.st)
		m.clock = m.clock.Add(checkInterval)
	}
}

func TestMonitorLeavesRunningContainersAlone(t *testing.T) {
	m := newMonitorFixture(map[string]string{"a": "running", "b": "running"})
	m.check(5)
	if len(m.ctrs.started) != 0 {
		t.Errorf("started %v", m.ctrs.started)
	}
}

func TestMonitorStartsAContainerThatStaysDown(t *testing.T) {
	for _, state := range []string{"stopped", "dead", "initialized"} {
		t.Run(state, func(t *testing.T) {
			m := newMonitorFixture(map[string]string{"a": "running", "b": state})
			m.check(1)
			if len(m.ctrs.started) != 0 {
				t.Fatalf("started %v after one check, want to wait for a second", m.ctrs.started)
			}
			m.check(1)
			if !reflect.DeepEqual(m.ctrs.started, []string{"b"}) {
				t.Errorf("started %v, want [b]", m.ctrs.started)
			}
		})
	}
}

func TestMonitorIgnoresAShortStop(t *testing.T) {
	m := newMonitorFixture(map[string]string{"a": "running", "b": "stopped"})
	m.check(1)
	m.ctrs.states["b"] = "running" // e.g. a restart through the API
	m.check(1)
	m.ctrs.states["b"] = "stopped"
	m.check(1)
	if len(m.ctrs.started) != 0 {
		t.Errorf("started %v, want nothing", m.ctrs.started)
	}
}

func TestMonitorLeavesOtherStatesAlone(t *testing.T) {
	for _, state := range []string{"paused", "restarting", "removing"} {
		t.Run(state, func(t *testing.T) {
			m := newMonitorFixture(map[string]string{"a": "running", "b": state})
			m.check(5)
			if len(m.ctrs.started) != 0 {
				t.Errorf("started %v", m.ctrs.started)
			}
		})
	}
}

func TestMonitorOnlyWatchesLongRunningContainers(t *testing.T) {
	m := newMonitorFixture(map[string]string{"a": "running", "b": "running", "migrate": "stopped"})
	m.check(5)
	if len(m.ctrs.started) != 0 {
		t.Errorf("started %v, want the one-shot service left alone", m.ctrs.started)
	}
}

func TestMonitorCantStartAMissingContainer(t *testing.T) {
	m := newMonitorFixture(map[string]string{"a": "running"})
	m.check(5)
	if len(m.ctrs.started) != 0 || !m.st.missing["b"] {
		t.Errorf("started %v, missing %v", m.ctrs.started, m.st.missing)
	}
	m.ctrs.states["b"] = "running"
	m.check(1)
	if m.st.missing["b"] {
		t.Error("still reported missing after it reappeared")
	}
}

func TestMonitorDoesNothingWhileTheWrapperIsUnreachable(t *testing.T) {
	m := newMonitorFixture(map[string]string{"a": "stopped", "b": "stopped"})
	m.ctrs.err = errors.New("unreachable")
	m.check(5)
	if len(m.ctrs.started) != 0 || !m.st.unreachable {
		t.Errorf("started %v, unreachable %v", m.ctrs.started, m.st.unreachable)
	}
	m.ctrs.err = nil
	m.check(2)
	if !reflect.DeepEqual(m.ctrs.started, []string{"a", "b"}) || m.st.unreachable {
		t.Errorf("started %v, unreachable %v", m.ctrs.started, m.st.unreachable)
	}
}

// crash sets the container down and runs the two checks that start it again.
func (m *monitorFixture) crash(name string) {
	m.ctrs.states[name] = "stopped"
	m.check(2)
}

func TestMonitorGivesUpAfterTheRestartLimit(t *testing.T) {
	m := newMonitorFixture(map[string]string{"a": "running", "b": "running"})
	for i := 0; i < 3; i++ {
		m.crash("b")
	}
	if len(m.ctrs.started) != 3 {
		t.Fatalf("started %v, want 3 restarts", m.ctrs.started)
	}
	m.crash("b")
	m.clock = m.clock.Add(time.Hour)
	m.check(5)
	if len(m.ctrs.started) != 3 || !m.st.gaveUp["b"] {
		t.Errorf("started %v, gave up %v, want no more restarts", m.ctrs.started, m.st.gaveUp)
	}
	m.crash("a")
	if len(m.ctrs.started) != 4 {
		t.Errorf("started %v, want the limit per container", m.ctrs.started)
	}
}

func TestMonitorCountsOnlyRestartsWithinTheWindow(t *testing.T) {
	m := newMonitorFixture(map[string]string{"a": "running", "b": "running"})
	for i := 0; i < 6; i++ {
		m.crash("b")
		m.clock = m.clock.Add(5 * time.Minute)
	}
	if len(m.ctrs.started) != 6 || m.st.gaveUp["b"] {
		t.Errorf("started %d times, gave up %v, want restarts spread over time to go on", len(m.ctrs.started), m.st.gaveUp)
	}
}

func TestMonitorRetriesAFailedStart(t *testing.T) {
	m := newMonitorFixture(map[string]string{"a": "running", "b": "stopped"})
	m.ctrs.startErr = errors.New("port in use")
	m.check(4)
	if !reflect.DeepEqual(m.ctrs.started, []string{"b", "b"}) {
		t.Errorf("started %v, want a retry two checks later", m.ctrs.started)
	}
}
