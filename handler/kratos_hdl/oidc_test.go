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

package kratos_hdl

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

const baseJSON = `{"version": "v1.3.1", "secrets": {"default": ["d-current-0123456789"], "cookie": ["c-current-0123456789"], "cipher": ["x-current-0123456789012345678901"]}}`

func validReq() lib_model.OIDCSettingsReq {
	return lib_model.OIDCSettingsReq{
		Enabled:      true,
		IssuerURL:    "https://idp.example/realms/a",
		ClientID:     "cid",
		ClientSecret: "first-secret",
		ExternalURL:  "https://gw.example/",
	}
}

func TestSetOIDCWritesKratosBlock(t *testing.T) {
	h := newTestHandler(t, baseJSON)
	if err := h.SetOIDC(context.Background(), validReq()); err != nil {
		t.Fatal(err)
	}
	m := readStrict(t, h.path)
	if v := get(t, m, "selfservice", "methods", "oidc", "enabled"); v != true {
		t.Errorf("enabled = %v", v)
	}
	if v := get(t, m, "selfservice", "methods", "oidc", "config", "base_redirect_uri"); v != "https://gw.example/core/auth/" {
		t.Errorf("base_redirect_uri = %v", v)
	}
	sso := ssoProvider(t, m)
	want := map[string]any{
		"id":            "sso-92da88febb50",
		"provider":      "generic",
		"client_id":     "cid",
		"client_secret": "first-secret",
		"issuer_url":    "https://idp.example/realms/a",
		"scope":         []any{"openid"},
		"mapper_url":    sso["mapper_url"],
	}
	if !reflect.DeepEqual(sso, want) {
		t.Errorf("provider = %v, want %v", sso, want)
	}
	mapper, _ := sso["mapper_url"].(string)
	src, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(mapper, "base64://"))
	if !strings.HasPrefix(mapper, "base64://") || err != nil || string(src) != "local claims = std.extVar('claims'); { identity: { traits: {} } }" {
		t.Errorf("mapper_url = %q (%v)", mapper, err)
	}
	if l, _ := get(t, m, "secrets", "default").([]any); len(l) != 1 || l[0] != "d-current-0123456789" {
		t.Errorf("secrets changed: %v", l)
	}
}

func TestSetOIDCKeepsForeignKeysAndOtherProviders(t *testing.T) {
	h := newTestHandler(t, foreignJSON)
	req := validReq()
	req.ClientID = "cid"
	req.ClientSecret = ""
	req.IssuerURL = "https://idp.example/realms/a"
	req.Scope = []string{"openid", "email"}
	req.ExternalURL = "https://gw.example"
	if err := h.SetOIDC(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	assertForeignKept(t, readStrict(t, h.path))
}

func TestSetOIDCKeepsRotationClock(t *testing.T) {
	h := newTestHandler(t, baseJSON)
	old := time.Now().Add(-100 * time.Hour).Truncate(time.Second)
	setModTime(t, h.path, old)
	if err := h.SetOIDC(context.Background(), validReq()); err != nil {
		t.Fatal(err)
	}
	if !modTime(t, h.path).Equal(old) {
		t.Errorf("mtime = %v, want %v: saving OIDC settings must not postpone secret rotation", modTime(t, h.path), old)
	}
}

func TestGetOIDCNeverReturnsSecret(t *testing.T) {
	h := newTestHandler(t, baseJSON)
	if err := h.SetOIDC(context.Background(), validReq()); err != nil {
		t.Fatal(err)
	}
	s, err := h.GetOIDC(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := lib_model.OIDCSettings{Enabled: true, IssuerURL: "https://idp.example/realms/a", ClientID: "cid", Scope: []string{"openid"}, ExternalURL: "https://gw.example", HasSecret: true, ProviderID: "sso-92da88febb50"}
	if !reflect.DeepEqual(s, want) {
		t.Errorf("settings = %+v, want %+v", s, want)
	}
	b, _ := json.Marshal(s)
	if strings.Contains(string(b), "first-secret") || strings.Contains(string(b), "client_secret") {
		t.Errorf("secret leaked: %s", b)
	}
}

func TestGetOIDCWithoutSettings(t *testing.T) {
	h := newTestHandler(t, baseJSON)
	s, err := h.GetOIDC(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(s)
	if string(b) != `{"enabled":false,"issuer_url":"","client_id":"","scope":[],"external_url":"","has_secret":false,"provider_id":""}` {
		t.Errorf("got %s", b)
	}
}

func TestSetOIDCWithoutSecretKeepsStored(t *testing.T) {
	h := newTestHandler(t, baseJSON)
	if err := h.SetOIDC(context.Background(), validReq()); err != nil {
		t.Fatal(err)
	}
	req := validReq()
	req.ClientSecret = ""
	req.IssuerURL = " " + req.IssuerURL + " "
	if err := h.SetOIDC(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	sso := ssoProvider(t, readStrict(t, h.path))
	if sso["client_secret"] != "first-secret" || sso["id"] != "sso-92da88febb50" {
		t.Errorf("provider = %v", sso)
	}
	req.ClientSecret = "second-secret"
	if err := h.SetOIDC(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if sso = ssoProvider(t, readStrict(t, h.path)); sso["client_secret"] != "second-secret" {
		t.Errorf("sent secret not stored: %v", sso["client_secret"])
	}
}

func TestDisableKeepsProviderData(t *testing.T) {
	h := newTestHandler(t, baseJSON)
	if err := h.SetOIDC(context.Background(), validReq()); err != nil {
		t.Fatal(err)
	}
	before := ssoProvider(t, readStrict(t, h.path))
	if err := h.SetOIDC(context.Background(), lib_model.OIDCSettingsReq{Enabled: false}); err != nil {
		t.Fatal(err)
	}
	m := readStrict(t, h.path)
	if v := get(t, m, "selfservice", "methods", "oidc", "enabled"); v != false {
		t.Errorf("enabled = %v", v)
	}
	if after := ssoProvider(t, m); !reflect.DeepEqual(before, after) {
		t.Errorf("provider changed on disable: %v -> %v", before, after)
	}
	s, _ := h.GetOIDC(context.Background())
	if s.Enabled || s.ClientID != "cid" || !s.HasSecret || s.ExternalURL != "https://gw.example" {
		t.Errorf("settings after disable = %+v", s)
	}
	req := validReq()
	req.ClientSecret = ""
	if err := h.SetOIDC(context.Background(), req); err != nil {
		t.Fatalf("re-enabling with the stored secret: %v", err)
	}
}

func TestDisableWithoutStoredProvider(t *testing.T) {
	h := newTestHandler(t, baseJSON)
	if err := h.SetOIDC(context.Background(), lib_model.OIDCSettingsReq{Enabled: false}); err != nil {
		t.Fatal(err)
	}
	m := readStrict(t, h.path)
	oidc := get(t, m, "selfservice", "methods", "oidc")
	if !reflect.DeepEqual(oidc, map[string]any{"enabled": false}) {
		t.Errorf("oidc = %v, want only the enabled flag (an incomplete provider breaks Kratos)", oidc)
	}
}

func TestSetOIDCValidation(t *testing.T) {
	mod := func(f func(*lib_model.OIDCSettingsReq)) lib_model.OIDCSettingsReq {
		r := validReq()
		f(&r)
		return r
	}
	cases := []struct {
		name string
		req  lib_model.OIDCSettingsReq
		msg  string
	}{
		{"missing issuer", mod(func(r *lib_model.OIDCSettingsReq) { r.IssuerURL = " " }), "issuer_url is required"},
		{"missing external", mod(func(r *lib_model.OIDCSettingsReq) { r.ExternalURL = "" }), "external_url is required"},
		{"missing client id", mod(func(r *lib_model.OIDCSettingsReq) { r.ClientID = "" }), "client_id is required"},
		{"missing secret", mod(func(r *lib_model.OIDCSettingsReq) { r.ClientSecret = "" }), "client_secret is required"},
		{"relative issuer", mod(func(r *lib_model.OIDCSettingsReq) { r.IssuerURL = "/realms/a" }), "issuer_url must be an absolute http(s) URL"},
		{"ftp issuer", mod(func(r *lib_model.OIDCSettingsReq) { r.IssuerURL = "ftp://idp.example" }), "issuer_url must be an absolute http(s) URL"},
		{"no host", mod(func(r *lib_model.OIDCSettingsReq) { r.ExternalURL = "https://" }), "external_url must be an absolute http(s) URL"},
		{"opaque", mod(func(r *lib_model.OIDCSettingsReq) { r.ExternalURL = "https:gw.example" }), "external_url must be an absolute http(s) URL"},
		{"query", mod(func(r *lib_model.OIDCSettingsReq) { r.ExternalURL = "https://gw.example/?a=b" }), "external_url must not contain"},
		{"empty fragment", mod(func(r *lib_model.OIDCSettingsReq) { r.IssuerURL = "https://idp.example/#" }), "issuer_url must not contain"},
		{"user info", mod(func(r *lib_model.OIDCSettingsReq) { r.IssuerURL = "https://u:p@idp.example" }), "issuer_url must not contain"},
		{"scope without openid", mod(func(r *lib_model.OIDCSettingsReq) { r.Scope = []string{"email"} }), "scope must contain openid"},
		{"scope with space", mod(func(r *lib_model.OIDCSettingsReq) { r.Scope = []string{"openid email"} }), "invalid character"},
		{"empty scope value", mod(func(r *lib_model.OIDCSettingsReq) { r.Scope = []string{"openid", ""} }), "empty values"},
		{"disabled incomplete", lib_model.OIDCSettingsReq{IssuerURL: "https://idp.example"}, "issuer_url, client_id and client_secret are required"},
		{"disabled without issuer", lib_model.OIDCSettingsReq{ClientID: "cid", ClientSecret: "s"}, "issuer_url, client_id and client_secret are required"},
		{"disabled bad url", lib_model.OIDCSettingsReq{ExternalURL: "gw.example"}, "external_url must be an absolute http(s) URL"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newTestHandler(t, baseJSON)
			err := h.SetOIDC(context.Background(), c.req)
			var iie *lib_model.InvalidInputError
			if !errors.As(err, &iie) {
				t.Fatalf("err = %v, want InvalidInputError", err)
			}
			if !strings.Contains(err.Error(), c.msg) {
				t.Errorf("err = %q, want it to contain %q", err, c.msg)
			}
			if strings.Contains(err.Error(), "first-secret") {
				t.Errorf("secret in error message: %q", err)
			}
			if b, _ := os.ReadFile(h.path); string(b) != baseJSON {
				t.Error("rejected request changed the file")
			}
		})
	}
}

func TestSetOIDCReportsAllErrors(t *testing.T) {
	h := newTestHandler(t, baseJSON)
	err := h.SetOIDC(context.Background(), lib_model.OIDCSettingsReq{Enabled: true})
	for _, msg := range []string{"issuer_url is required", "external_url is required", "client_id is required", "client_secret is required"} {
		if err == nil || !strings.Contains(err.Error(), msg) {
			t.Errorf("err = %v, want it to contain %q", err, msg)
		}
	}
}

// Without the file mutex a rotation can read the file before a concurrent SetOIDC writes it and then write the
// stale copy back, so a check right after SetOIDC returns would see an older client_id.
func TestConcurrentRotationKeepsOIDC(t *testing.T) {
	h := newTestHandler(t, baseJSON)
	h.maxAge = -time.Hour
	stop := make(chan struct{})
	rotErr := make(chan error, 1)
	go func() {
		defer close(rotErr)
		for {
			select {
			case <-stop:
				return
			default:
			}
			if err := h.refreshSecrets(); err != nil {
				rotErr <- err
				return
			}
		}
	}()
	defer func() {
		close(stop)
		if err := <-rotErr; err != nil {
			t.Error(err)
		}
	}()
	for i := 0; i < 300; i++ {
		req := validReq()
		req.ClientID = fmt.Sprintf("cid-%d", i)
		if err := h.SetOIDC(context.Background(), req); err != nil {
			t.Fatal(err)
		}
		for j := 0; j < 5; j++ {
			d, err := readConfig(h.path)
			if err != nil {
				t.Fatal(err)
			}
			if got := readOIDC(d).clientID; got != req.ClientID {
				t.Fatalf("client_id = %q after writing %q: a rotation wrote back a stale copy", got, req.ClientID)
			}
		}
	}
}

func TestOIDCProviderID(t *testing.T) {
	// Expected values are sha256 prefixes computed outside the code, so a changed derivation fails here.
	for issuer, want := range map[string]string{
		"https://idp.example/realms/a":  "sso-92da88febb50",
		"https://idp.example/realms/a/": "sso-0b0bf25afd96",
		"https://idp.example/realms/b":  "sso-8d004c901988",
	} {
		if got := oidcProviderID(issuer); got != want {
			t.Errorf("oidcProviderID(%q) = %q, want %q", issuer, got, want)
		}
		if !isManagedProviderID(want) {
			t.Errorf("%q not recognised as managed", want)
		}
	}
	for id, want := range map[string]bool{"sso": true, "sso-": false, "sso-custom": false, "sso-92DA88FEBB50": false, "sso-92da88febb5": false, "sso-92da88febb500": false, "other": false, "": false} {
		if got := isManagedProviderID(id); got != want {
			t.Errorf("isManagedProviderID(%q) = %v, want %v", id, got, want)
		}
	}
}

const legacyJSON = `{"version": "v1.3.1", "secrets": {"default": ["d-current-0123456789"], "cookie": ["c-current-0123456789"], "cipher": ["x-current-0123456789012345678901"]},
  "selfservice": {"methods": {"oidc": {"enabled": true, "config": {"base_redirect_uri": "https://gw.example/core/auth/", "providers": [
    {"id": "other", "provider": "google", "client_id": "o", "client_secret": "o", "mapper_url": "base64://e30="},
    {"id": "sso", "provider": "generic", "client_id": "cid", "client_secret": "stored-secret", "issuer_url": "https://idp.example/realms/a", "scope": ["openid"], "mapper_url": "base64://e30=", "label": "Old"},
    {"id": "sso-custom", "provider": "generic", "client_id": "c", "client_secret": "c", "issuer_url": "https://custom.example", "mapper_url": "base64://e30="},
    {"id": "sso-8d004c901988", "provider": "generic", "client_id": "cid", "client_secret": "older-secret", "issuer_url": "https://idp.example/realms/b", "mapper_url": "base64://e30="}
  ]}}}}}`

func TestSetOIDCReplacesEarlierManagedEntries(t *testing.T) {
	h := newTestHandler(t, legacyJSON)
	if s, err := h.GetOIDC(context.Background()); err != nil || s.ProviderID != "sso" || s.ClientID != "cid" || !s.HasSecret {
		t.Fatalf("legacy settings = %+v, %v", s, err)
	}
	foreign := map[string]map[string]any{}
	providers, _ := get(t, readStrict(t, h.path), "selfservice", "methods", "oidc", "config", "providers").([]any)
	for _, p := range providers {
		pm := p.(map[string]any)
		if id := pm["id"].(string); id == "other" || id == "sso-custom" {
			foreign[id] = pm
		}
	}
	req := validReq()
	req.ClientSecret = ""
	if err := h.SetOIDC(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	m := readStrict(t, h.path)
	providers, _ = get(t, m, "selfservice", "methods", "oidc", "config", "providers").([]any)
	var ids []string
	for _, p := range providers {
		ids = append(ids, p.(map[string]any)["id"].(string))
	}
	if !reflect.DeepEqual(ids, []string{"other", "sso-92da88febb50", "sso-custom"}) {
		t.Fatalf("provider ids = %v", ids)
	}
	if !reflect.DeepEqual(providers[0], foreign["other"]) || !reflect.DeepEqual(providers[2], foreign["sso-custom"]) {
		t.Errorf("foreign providers changed: %v", providers)
	}
	sso := ssoProvider(t, m)
	if sso["client_secret"] != "stored-secret" || sso["issuer_url"] != "https://idp.example/realms/a" || sso["label"] != nil {
		t.Errorf("provider = %v, want a fresh entry reusing the secret of the same issuer and client", sso)
	}
	if s, _ := h.GetOIDC(context.Background()); s.ProviderID != "sso-92da88febb50" {
		t.Errorf("provider_id = %q", s.ProviderID)
	}
}

func TestSetOIDCIssuerChangeReplacesEntry(t *testing.T) {
	h := newTestHandler(t, baseJSON)
	if err := h.SetOIDC(context.Background(), validReq()); err != nil {
		t.Fatal(err)
	}
	req := validReq()
	req.IssuerURL = "https://idp.example/realms/b"
	req.ClientSecret = "b-secret"
	if err := h.SetOIDC(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	sso := ssoProvider(t, readStrict(t, h.path))
	if sso["id"] != "sso-8d004c901988" || sso["client_secret"] != "b-secret" || sso["issuer_url"] != "https://idp.example/realms/b" {
		t.Errorf("provider = %v", sso)
	}
	if s, _ := h.GetOIDC(context.Background()); s.ProviderID != "sso-8d004c901988" {
		t.Errorf("provider_id = %q", s.ProviderID)
	}
}

func TestSetOIDCSecretRequiredOnIdPChange(t *testing.T) {
	cases := []struct {
		name string
		f    func(*lib_model.OIDCSettingsReq)
	}{
		{"issuer", func(r *lib_model.OIDCSettingsReq) { r.IssuerURL = "https://idp.example/realms/b" }},
		{"issuer trailing slash", func(r *lib_model.OIDCSettingsReq) { r.IssuerURL = "https://idp.example/realms/a/" }},
		{"client id", func(r *lib_model.OIDCSettingsReq) { r.ClientID = "cid-2" }},
		{"client id while disabled", func(r *lib_model.OIDCSettingsReq) {
			*r = lib_model.OIDCSettingsReq{ClientID: "cid-2"}
		}},
		{"issuer while disabled", func(r *lib_model.OIDCSettingsReq) {
			*r = lib_model.OIDCSettingsReq{IssuerURL: "https://idp.example/realms/b"}
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newTestHandler(t, baseJSON)
			if err := h.SetOIDC(context.Background(), validReq()); err != nil {
				t.Fatal(err)
			}
			before, _ := os.ReadFile(h.path)
			req := validReq()
			req.ClientSecret = ""
			c.f(&req)
			err := h.SetOIDC(context.Background(), req)
			var iie *lib_model.InvalidInputError
			if !errors.As(err, &iie) || !strings.Contains(err.Error(), "client_secret is required when issuer_url or client_id changes") {
				t.Fatalf("err = %v, want the secret to be required", err)
			}
			if after, _ := os.ReadFile(h.path); string(after) != string(before) {
				t.Error("rejected request changed the file")
			}
		})
	}
}

func TestSetOIDCSecretNotRequiredWithoutIdPChange(t *testing.T) {
	cases := []struct {
		name string
		f    func(*lib_model.OIDCSettingsReq)
	}{
		{"external url", func(r *lib_model.OIDCSettingsReq) { r.ExternalURL = "https://gw2.example" }},
		{"scope", func(r *lib_model.OIDCSettingsReq) { r.Scope = []string{"openid", "email"} }},
		{"disable only", func(r *lib_model.OIDCSettingsReq) { *r = lib_model.OIDCSettingsReq{} }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newTestHandler(t, baseJSON)
			if err := h.SetOIDC(context.Background(), validReq()); err != nil {
				t.Fatal(err)
			}
			req := validReq()
			req.ClientSecret = ""
			c.f(&req)
			if err := h.SetOIDC(context.Background(), req); err != nil {
				t.Fatal(err)
			}
			if sso := ssoProvider(t, readStrict(t, h.path)); sso["client_secret"] != "first-secret" {
				t.Errorf("client_secret = %v, want the stored one", sso["client_secret"])
			}
		})
	}
}

func TestSetOIDCEmptyScopeKeepsStored(t *testing.T) {
	h := newTestHandler(t, baseJSON)
	req := validReq()
	req.Scope = []string{"openid", "email"}
	if err := h.SetOIDC(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	for _, scope := range [][]string{nil, {}} {
		req = validReq()
		req.ClientSecret = ""
		req.Scope = scope
		if err := h.SetOIDC(context.Background(), req); err != nil {
			t.Fatal(err)
		}
		if s, _ := h.GetOIDC(context.Background()); !reflect.DeepEqual(s.Scope, []string{"openid", "email"}) {
			t.Errorf("scope after empty %v = %v, want the stored one", scope, s.Scope)
		}
	}
}
