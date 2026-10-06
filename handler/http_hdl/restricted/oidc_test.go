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

package restricted

import (
	"context"
	"encoding/json"
	"errors"
	gin_mw "github.com/SENERGY-Platform/gin-middleware"
	"github.com/SENERGY-Platform/mgw-core-manager/lib"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type fakeApi struct {
	lib.Api
	settings lib_model.OIDCSettings
	setErr   error
	reqs     []lib_model.OIDCSettingsReq
}

func (f *fakeApi) GetOIDCSettings(_ context.Context) (lib_model.OIDCSettings, error) {
	return f.settings, nil
}

func (f *fakeApi) SetOIDCSettings(_ context.Context, settings lib_model.OIDCSettingsReq) (string, error) {
	f.reqs = append(f.reqs, settings)
	if f.setErr != nil {
		return "", f.setErr
	}
	return "job-1", nil
}

func newTestEngine(a lib.Api) *gin.Engine {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	e.Use(gin_mw.ErrorHandler(util.GetStatusCode, ", "))
	for _, h := range []func(lib.Api) (string, string, gin.HandlerFunc){GetOIDCSettingsH, PutOIDCSettingsH} {
		method, p, hf := h(a)
		e.Handle(method, "/"+lib_model.RestrictedPath+"/"+p, hf)
	}
	return e
}

func do(e *gin.Engine, method, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/restricted/oidc", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestGetOIDCSettingsResponseShape(t *testing.T) {
	a := &fakeApi{settings: lib_model.OIDCSettings{Enabled: true, IssuerURL: "https://idp.example", ClientID: "cid", Scope: []string{"openid"}, ExternalURL: "https://gw.example", HasSecret: true, ProviderID: "sso-0123456789ab"}}
	rec := do(newTestEngine(a), http.MethodGet, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body)
	}
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{"enabled": true, "issuer_url": "https://idp.example", "client_id": "cid", "scope": []any{"openid"}, "external_url": "https://gw.example", "has_secret": true, "provider_id": "sso-0123456789ab"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("body = %v, want %v", got, want)
	}
}

func TestPutOIDCSettings(t *testing.T) {
	a := &fakeApi{}
	rec := do(newTestEngine(a), http.MethodPut, `{"enabled":true,"issuer_url":"https://idp.example","client_id":"cid","client_secret":"s","scope":["openid","email"],"external_url":"https://gw.example"}`)
	if rec.Code != http.StatusOK || rec.Body.String() != "job-1" {
		t.Fatalf("status = %d, body = %q", rec.Code, rec.Body)
	}
	want := lib_model.OIDCSettingsReq{Enabled: true, IssuerURL: "https://idp.example", ClientID: "cid", ClientSecret: "s", Scope: []string{"openid", "email"}, ExternalURL: "https://gw.example"}
	if len(a.reqs) != 1 || !reflect.DeepEqual(a.reqs[0], want) {
		t.Errorf("request = %+v, want %+v", a.reqs, want)
	}
}

func TestPutOIDCSettingsErrors(t *testing.T) {
	cases := []struct {
		name   string
		body   string
		setErr error
		status int
	}{
		{"malformed body", `{"enabled":"yes"}`, nil, http.StatusBadRequest},
		{"invalid settings", `{"enabled":true}`, lib_model.NewInvalidInputError(errors.New("client_id is required")), http.StatusBadRequest},
		{"internal", `{"enabled":false}`, lib_model.NewInternalError(errors.New("disk full")), http.StatusInternalServerError},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := do(newTestEngine(&fakeApi{setErr: c.setErr}), http.MethodPut, c.body)
			if rec.Code != c.status {
				t.Errorf("status = %d, want %d (%s)", rec.Code, c.status, rec.Body)
			}
		})
	}
}
