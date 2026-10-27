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
	"strings"
	"testing"
	"time"

	cew_lib "github.com/SENERGY-Platform/mgw-container-engine-wrapper/lib"
	cew_model "github.com/SENERGY-Platform/mgw-container-engine-wrapper/lib/model"
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

func newHandler(c cew_lib.Api, installPath string, services ...service) *Handler {
	h := New(c, "core", installPath, time.Second)
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
			h := newHandler(&fakeCew{containers: c.containers, err: c.cewErr}, "", c.services...)
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
			if got.Version != "" || got.Update != nil {
				t.Errorf("version = %q, update = %+v, want empty without install path", got.Version, got.Update)
			}
		})
	}
}

func TestHealthReadsInstallPath(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".version"), []byte("v2.13.2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".update_status"), []byte(`{"from_version":"v2.13.1","to_version":"v2.13.2","started":"2026-10-27T10:00:00Z","finished":"2026-10-27T10:03:12Z","result":"success","step":"starting components"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	h := newHandler(&fakeCew{}, dir)
	got, err := h.Health(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := &lib_model.UpdateStatus{FromVersion: "v2.13.1", ToVersion: "v2.13.2", Started: "2026-10-27T10:00:00Z", Finished: "2026-10-27T10:03:12Z", Result: "success", Step: "starting components"}
	if got.Version != "v2.13.2" || !reflect.DeepEqual(got.Update, want) {
		t.Errorf("version = %q, update = %+v", got.Version, got.Update)
	}
	if got.State != lib_model.HealthHealthy || got.Services == nil || len(got.Services) != 0 {
		t.Errorf("state = %d, services = %#v", got.State, got.Services)
	}
}

func TestHealthMissingVersionFile(t *testing.T) {
	got, err := newHandler(&fakeCew{}, t.TempDir()).Health(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != "" || got.Update != nil {
		t.Errorf("version = %q, update = %+v", got.Version, got.Update)
	}
}

func TestReadUpdateStatus(t *testing.T) {
	cases := []struct {
		name    string
		content *string
		want    *lib_model.UpdateStatus
	}{
		{"missing file", nil, nil},
		{"valid", strPtr(`{"from_version":"v1","to_version":"v2","started":"s","finished":"f","result":"failed","step":"pull"}`),
			&lib_model.UpdateStatus{FromVersion: "v1", ToVersion: "v2", Started: "s", Finished: "f", Result: "failed", Step: "pull"}},
		{"missing fields", strPtr(`{"result":"failed"}`), &lib_model.UpdateStatus{Result: "failed"}},
		{"invalid json", strPtr(`{"result":`), nil},
		{"wrong field type", strPtr(`{"result":1}`), nil},
		{"null", strPtr("null"), nil},
		{"oversized", strPtr(`{"step":"` + strings.Repeat("a", maxStatusFileLen) + `"}`), nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			if c.content != nil {
				if err := os.WriteFile(filepath.Join(dir, ".update_status"), []byte(*c.content), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if got := readUpdateStatus(dir); !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %+v, want %+v", got, c.want)
			}
		})
	}
}

func TestReadVersionOversized(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".version"), []byte(strings.Repeat("v", maxStatusFileLen+1)), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := readVersion(dir); got != "" {
		t.Errorf("got %d bytes, want empty", len(got))
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
	h := New(&fakeCew{}, "core", "", time.Second)
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
