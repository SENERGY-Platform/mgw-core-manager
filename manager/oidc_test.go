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

package manager

import (
	"context"
	"errors"
	"fmt"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"github.com/SENERGY-Platform/mgw-go-service-base/job-hdl"
	"net/http"
	"testing"
)

type fakeCoreSrvHdl struct {
	CoreServiceHandler
	known     string
	gets      []string
	restarted []string
}

func (f *fakeCoreSrvHdl) Get(_ context.Context, name string) (lib_model.CoreService, error) {
	f.gets = append(f.gets, name)
	if name != f.known {
		return lib_model.CoreService{}, lib_model.NewNotFoundError(fmt.Errorf("service '%s' not found", name))
	}
	return lib_model.CoreService{Name: name}, nil
}

func (f *fakeCoreSrvHdl) Restart(_ context.Context, name string) error {
	f.restarted = append(f.restarted, name)
	return nil
}

type fakeKratosHdl struct {
	KratosHandler
	err  error
	reqs []lib_model.OIDCSettingsReq
}

func (f *fakeKratosHdl) SetOIDC(_ context.Context, req lib_model.OIDCSettingsReq) error {
	f.reqs = append(f.reqs, req)
	return f.err
}

// fakeJobHdl runs every job synchronously.
type fakeJobHdl struct {
	job_hdl.JobHandler
}

func (f *fakeJobHdl) Create(ctx context.Context, _ string, tFunc job_hdl.TargetFunc) (string, error) {
	ctx, cf := context.WithCancel(ctx)
	if _, err := tFunc(ctx, cf); err != nil {
		return "", err
	}
	return "job-1", nil
}

func newTestManager(known string, kratosErr error) (*Manager, *fakeCoreSrvHdl, *fakeKratosHdl) {
	srv := &fakeCoreSrvHdl{known: known}
	kratos := &fakeKratosHdl{err: kratosErr}
	return New(srv, nil, nil, nil, kratos, "kratos", &fakeJobHdl{}, nil), srv, kratos
}

func TestSetOIDCSettingsRestartsKratos(t *testing.T) {
	m, srv, kratos := newTestManager("kratos", nil)
	jID, err := m.SetOIDCSettings(context.Background(), lib_model.OIDCSettingsReq{Enabled: true, ClientID: "cid"})
	if err != nil {
		t.Fatal(err)
	}
	if jID != "job-1" {
		t.Errorf("job id = %q", jID)
	}
	if len(kratos.reqs) != 1 || kratos.reqs[0].ClientID != "cid" {
		t.Errorf("SetOIDC calls = %+v", kratos.reqs)
	}
	if len(srv.restarted) != 1 || srv.restarted[0] != "kratos" {
		t.Errorf("restarted = %v, want [kratos]", srv.restarted)
	}
}

func TestSetOIDCSettingsInvalidInputSkipsRestart(t *testing.T) {
	m, srv, _ := newTestManager("kratos", lib_model.NewInvalidInputError(errors.New("client_id is required")))
	_, err := m.SetOIDCSettings(context.Background(), lib_model.OIDCSettingsReq{Enabled: true})
	if util.GetStatusCode(err) != http.StatusBadRequest {
		t.Errorf("err = %v, want a 400 error", err)
	}
	if len(srv.restarted) != 0 {
		t.Errorf("restarted = %v after rejected settings", srv.restarted)
	}
}

func TestSetOIDCSettingsUnknownServiceWritesNothing(t *testing.T) {
	m, srv, kratos := newTestManager("identity-server", nil)
	_, err := m.SetOIDCSettings(context.Background(), lib_model.OIDCSettingsReq{})
	if util.GetStatusCode(err) != http.StatusInternalServerError {
		t.Errorf("err = %v (status %d), want a 500 error", err, util.GetStatusCode(err))
	}
	if len(kratos.reqs) != 0 || len(srv.restarted) != 0 {
		t.Errorf("SetOIDC calls = %v, restarted = %v", kratos.reqs, srv.restarted)
	}
}
