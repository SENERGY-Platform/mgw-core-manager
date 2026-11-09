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
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	cew_lib "github.com/SENERGY-Platform/mgw-container-engine-wrapper/lib"
	cew_model "github.com/SENERGY-Platform/mgw-container-engine-wrapper/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/handler/state_hdl"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	log_level "github.com/y-du/go-log-level"
	"github.com/y-du/go-log-level/level"
	"io"
	"log"
)

func TestMain(m *testing.M) {
	util.Logger, _ = log_level.New(log.New(io.Discard, "", 0), level.Off)
	os.Exit(m.Run())
}

type fakeCew struct {
	cew_lib.Api
	containers []cew_model.Container
	err        error
}

func (f *fakeCew) GetContainers(_ context.Context, _ cew_model.ContainerFilter) ([]cew_model.Container, error) {
	return f.containers, f.err
}

func strPtr(s string) *string { return &s }

func newHandler(c cew_lib.Api, statePaths state_hdl.Paths, services ...service) *Handler {
	h := New(c, "core", statePaths, time.Second)
	h.services = make(map[string]service)
	for _, s := range services {
		h.services[s.Name] = s
	}
	return h
}

func TestHealth(t *testing.T) {
	reg := service{Name: "reg", ContainerName: "reg-ctr"}
	one := service{Name: "one", ContainerName: "one-ctr", OneShot: true}
	ctr := func(name, state string, health *string) cew_model.Container {
		return cew_model.Container{Name: name, State: state, Health: health}
	}
	cases := []struct {
		name       string
		services   []service
		containers []cew_model.Container
		cewErr     error
		wantState  int
		want       []lib_model.CoreServiceHealth
	}{
		{"regular running", []service{reg}, []cew_model.Container{ctr("reg-ctr", "running", nil)}, nil, 1,
			[]lib_model.CoreServiceHealth{{Name: "reg", State: strPtr("running"), Ok: true}}},
		{"regular running healthy", []service{reg}, []cew_model.Container{ctr("reg-ctr", "running", strPtr("healthy"))}, nil, 1,
			[]lib_model.CoreServiceHealth{{Name: "reg", State: strPtr("running"), Health: strPtr("healthy"), Ok: true}}},
		{"regular stopped", []service{reg}, []cew_model.Container{ctr("reg-ctr", "stopped", nil)}, nil, 2,
			[]lib_model.CoreServiceHealth{{Name: "reg", State: strPtr("stopped")}}},
		{"regular unhealthy", []service{reg}, []cew_model.Container{ctr("reg-ctr", "running", strPtr("unhealthy"))}, nil, 2,
			[]lib_model.CoreServiceHealth{{Name: "reg", State: strPtr("running"), Health: strPtr("unhealthy")}}},
		{"one-shot stopped", []service{one}, []cew_model.Container{ctr("one-ctr", "stopped", nil)}, nil, 1,
			[]lib_model.CoreServiceHealth{{Name: "one", State: strPtr("stopped"), OneShot: true, Ok: true}}},
		{"one-shot restarting", []service{one}, []cew_model.Container{ctr("one-ctr", "restarting", nil)}, nil, 2,
			[]lib_model.CoreServiceHealth{{Name: "one", State: strPtr("restarting"), OneShot: true}}},
		{"one-shot dead", []service{one}, []cew_model.Container{ctr("one-ctr", "dead", nil)}, nil, 2,
			[]lib_model.CoreServiceHealth{{Name: "one", State: strPtr("dead"), OneShot: true}}},
		{"missing container", []service{reg, one}, []cew_model.Container{ctr("one-ctr", "stopped", nil)}, nil, 2,
			[]lib_model.CoreServiceHealth{{Name: "one", State: strPtr("stopped"), OneShot: true, Ok: true}, {Name: "reg"}}},
		{"cew error", []service{reg, one}, nil, errors.New("socket closed"), 0,
			[]lib_model.CoreServiceHealth{{Name: "one", OneShot: true}, {Name: "reg"}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newHandler(&fakeCew{containers: c.containers, err: c.cewErr}, state_hdl.Paths{}, c.services...)
			got, err := h.Health(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if got.State != c.wantState {
				t.Errorf("state = %d, want %d", got.State, c.wantState)
			}
			if !reflect.DeepEqual(got.Services, c.want) {
				t.Errorf("services = %+v, want %+v", got.Services, c.want)
			}
			if got.Version != "" || got.Update != nil || got.Check != nil {
				t.Errorf("version = %q, update = %+v, check = %+v, want empty without state file paths", got.Version, got.Update, got.Check)
			}
		})
	}
}

func TestHealthReadsStateFiles(t *testing.T) {
	dir := t.TempDir()
	paths := state_hdl.Paths{
		Version: filepath.Join(dir, "version"),
		Update:  filepath.Join(dir, "update"),
		Check:   filepath.Join(dir, "check"),
	}
	files := map[string]string{
		paths.Version: "v2.13.2\n",
		paths.Update:  "from_version=v2.13.1\nto_version=v2.13.2\nstarted=2026-10-27T10:00:00Z\nfinished=2026-10-27T10:03:12Z\nresult=success\nphase=done\nkind=update\nstep=cleaning up\n",
		paths.Check:   "check_time=2026-10-27T09:59:58Z\ncheck_result=update_available\ncheck_version=v2.13.2\n",
	}
	for p, content := range files {
		if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	got, err := newHandler(&fakeCew{}, paths).Health(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	wantUpdate := &lib_model.UpdateStatus{FromVersion: "v2.13.1", ToVersion: "v2.13.2", Started: "2026-10-27T10:00:00Z", Finished: "2026-10-27T10:03:12Z", Result: "success", Phase: "done", Kind: "update", Step: "cleaning up"}
	wantCheck := &lib_model.CheckStatus{Time: "2026-10-27T09:59:58Z", Result: "update_available", Version: "v2.13.2"}
	if got.Version != "v2.13.2" || !reflect.DeepEqual(got.Update, wantUpdate) || !reflect.DeepEqual(got.Check, wantCheck) {
		t.Errorf("version = %q, update = %+v, check = %+v", got.Version, got.Update, got.Check)
	}
	if got.State != lib_model.HealthHealthy || got.Services == nil || len(got.Services) != 0 {
		t.Errorf("state = %d, services = %#v", got.State, got.Services)
	}
}

func TestHealthMissingStateFiles(t *testing.T) {
	dir := t.TempDir()
	paths := state_hdl.Paths{Version: filepath.Join(dir, "version"), Update: filepath.Join(dir, "update"), Check: filepath.Join(dir, "check")}
	got, err := newHandler(&fakeCew{}, paths).Health(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != "" || got.Update != nil || got.Check != nil {
		t.Errorf("version = %q, update = %+v, check = %+v", got.Version, got.Update, got.Check)
	}
}

func TestInitMarksOneShotServices(t *testing.T) {
	compose := `services:
  gateway:
    image: example/gateway:1.0
    container_name: gw
    restart: always
  migrate:
    image: example/migrate:1.0
    container_name: mig
    restart: on-failure
  limited:
    image: example/limited:1.0
    container_name: lim
    restart: "on-failure:3"
  plain:
    image: example/plain:1.0
    container_name: plain
`
	p := filepath.Join(t.TempDir(), "docker-compose.yml")
	if err := os.WriteFile(p, []byte(compose), 0o600); err != nil {
		t.Fatal(err)
	}
	h := New(&fakeCew{}, "core", state_hdl.Paths{}, time.Second)
	if err := h.Init(p); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"gateway": false, "migrate": true, "limited": true, "plain": false}
	got := make(map[string]bool)
	for name, s := range h.services {
		got[name] = s.OneShot
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("one-shot = %v, want %v", got, want)
	}
}
