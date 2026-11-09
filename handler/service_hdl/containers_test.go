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
	"reflect"
	"testing"

	cew_model "github.com/SENERGY-Platform/mgw-container-engine-wrapper/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/handler/state_hdl"
)

type startRecorder struct {
	fakeCew
	started []string
}

func (s *startRecorder) StartContainer(_ context.Context, id string) error {
	s.started = append(s.started, id)
	return nil
}

func TestLongRunningContainers(t *testing.T) {
	h := newHandler(&fakeCew{}, state_hdl.Paths{},
		service{Name: "b", ContainerName: "b-ctr"},
		service{Name: "mig", ContainerName: "mig-ctr", OneShot: true},
		service{Name: "a", ContainerName: "a-ctr"},
	)
	if got, want := h.LongRunningContainers(), []string{"a-ctr", "b-ctr"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestContainerStates(t *testing.T) {
	h := newHandler(&fakeCew{containers: []cew_model.Container{
		{Name: "a-ctr", State: cew_model.RunningState},
		{Name: "b-ctr", State: cew_model.StoppedState},
	}}, state_hdl.Paths{})
	got, err := h.ContainerStates(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if want := map[string]string{"a-ctr": "running", "b-ctr": "stopped"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestContainerStatesReturnsTheWrappersError(t *testing.T) {
	h := newHandler(&fakeCew{err: errors.New("unreachable")}, state_hdl.Paths{})
	if got, err := h.ContainerStates(context.Background()); err == nil || got != nil {
		t.Errorf("got %v, %v, want an error", got, err)
	}
}

func TestStartContainer(t *testing.T) {
	rec := &startRecorder{}
	h := newHandler(rec, state_hdl.Paths{})
	if err := h.StartContainer(context.Background(), "a-ctr"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rec.started, []string{"a-ctr"}) {
		t.Errorf("started %v", rec.started)
	}
}
