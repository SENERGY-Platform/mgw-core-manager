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
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testVer = "v1.3.1"

// foreignJSON carries keys the handler does not own, including an integer that float64 cannot represent.
const foreignJSON = `{
  "version": "v1.3.1",
  "secrets": {"default": ["d-current-0123456789"], "cookie": ["c-current-0123456789"], "cipher": ["x-current-0123456789012345678901"], "pagination": ["keep-me"]},
  "courier": {"smtp": {"from_name": "gateway"}},
  "big": 9007199254740993,
  "selfservice": {
    "flows": {"login": {"lifespan": "5m"}},
    "methods": {
      "password": {"enabled": true},
      "oidc": {"enabled": true, "config": {"base_redirect_uri": "https://gw.example/core/auth/", "providers": [
        {"id": "other", "provider": "google", "client_id": "o", "client_secret": "o", "mapper_url": "base64://e30="},
        {"id": "sso-92da88febb50", "provider": "generic", "client_id": "cid", "client_secret": "stored-secret", "issuer_url": "https://idp.example/realms/a", "scope": ["openid", "email"], "mapper_url": "base64://e30=", "label": "Company SSO"}
      ]}}
    }
  }
}`

func newTestHandler(t *testing.T, content string) *Handler {
	t.Helper()
	p := filepath.Join(t.TempDir(), "config.json")
	if content != "" {
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	h, err := New(context.Background(), testVer, p, 32, time.Hour, time.Hour, -1, -1)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

// readStrict fails unless the whole file is exactly one JSON object.
func readStrict(t *testing.T, p string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	var m map[string]any
	if err = dec.Decode(&m); err != nil {
		t.Fatalf("file is not valid JSON: %v\n%s", err, b)
	}
	if dec.More() {
		t.Fatalf("file has data after the JSON object:\n%s", b)
	}
	return m
}

func setModTime(t *testing.T, p string, mt time.Time) {
	t.Helper()
	if err := os.Chtimes(p, mt, mt); err != nil {
		t.Fatal(err)
	}
}

func modTime(t *testing.T, p string) time.Time {
	t.Helper()
	fi, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	return fi.ModTime()
}

func get(t *testing.T, m map[string]any, keys ...string) any {
	t.Helper()
	var v any = m
	for _, k := range keys {
		obj, ok := v.(map[string]any)
		if !ok {
			t.Fatalf("%v: not an object at %q", keys, k)
		}
		v = obj[k]
	}
	return v
}

// ssoProvider returns the only provider entry with an id this handler manages.
func ssoProvider(t *testing.T, m map[string]any) map[string]any {
	t.Helper()
	providers, _ := get(t, m, "selfservice", "methods", "oidc", "config", "providers").([]any)
	var found []map[string]any
	for _, p := range providers {
		if pm, ok := p.(map[string]any); ok {
			if id, _ := pm["id"].(string); id == "sso" || (strings.HasPrefix(id, "sso-") && len(id) == 16) {
				found = append(found, pm)
			}
		}
	}
	if len(found) != 1 {
		t.Fatalf("managed providers = %v, want exactly one", found)
	}
	return found[0]
}

// assertForeignKept checks every key of foreignJSON that neither rotation nor a version bump owns.
func assertForeignKept(t *testing.T, m map[string]any) {
	t.Helper()
	if n, ok := m["big"].(json.Number); !ok || n.String() != "9007199254740993" {
		t.Errorf("big = %v, want 9007199254740993 unchanged", m["big"])
	}
	if v := get(t, m, "courier", "smtp", "from_name"); v != "gateway" {
		t.Errorf("courier lost: %v", v)
	}
	if v := get(t, m, "selfservice", "flows", "login", "lifespan"); v != "5m" {
		t.Errorf("flows lost: %v", v)
	}
	if v := get(t, m, "selfservice", "methods", "password", "enabled"); v != true {
		t.Errorf("password method lost: %v", v)
	}
	if v, _ := get(t, m, "secrets", "pagination").([]any); len(v) != 1 || v[0] != "keep-me" {
		t.Errorf("unknown secrets key lost: %v", v)
	}
	if v := get(t, m, "selfservice", "methods", "oidc", "enabled"); v != true {
		t.Errorf("oidc enabled = %v", v)
	}
	if v := get(t, m, "selfservice", "methods", "oidc", "config", "base_redirect_uri"); v != "https://gw.example/core/auth/" {
		t.Errorf("base_redirect_uri = %v", v)
	}
	providers, _ := get(t, m, "selfservice", "methods", "oidc", "config", "providers").([]any)
	if len(providers) != 2 {
		t.Fatalf("providers = %v", providers)
	}
	sso := ssoProvider(t, m)
	for k, want := range map[string]string{"id": "sso-92da88febb50", "client_id": "cid", "client_secret": "stored-secret", "issuer_url": "https://idp.example/realms/a", "label": "Company SSO"} {
		if sso[k] != want {
			t.Errorf("sso %s = %v, want %s", k, sso[k], want)
		}
	}
}

func TestRefreshSecretsKeepsForeignKeysAndOIDC(t *testing.T) {
	h := newTestHandler(t, foreignJSON)
	setModTime(t, h.path, time.Now().Add(-2*time.Hour))
	if err := h.refreshSecrets(); err != nil {
		t.Fatal(err)
	}
	m := readStrict(t, h.path)
	assertForeignKept(t, m)
	for key, old := range map[string]string{"default": "d-current-0123456789", "cookie": "c-current-0123456789", "cipher": "x-current-0123456789012345678901"} {
		l, _ := get(t, m, "secrets", key).([]any)
		if len(l) != 2 || l[1] != old || l[0] == old {
			t.Errorf("secrets.%s = %v, want [new, %s]", key, l, old)
		}
		if s, _ := l[0].(string); len(s) != 32 {
			t.Errorf("secrets.%s new secret has length %d", key, len(s))
		}
	}
	if time.Since(modTime(t, h.path)) > time.Minute {
		t.Error("rotation must advance the modification time")
	}
}

func TestRefreshSecretsSkipsFreshFile(t *testing.T) {
	h := newTestHandler(t, foreignJSON)
	if err := h.refreshSecrets(); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(h.path)
	if string(b) != foreignJSON {
		t.Error("file younger than maxAge was rewritten")
	}
}

func TestInitVersionBumpKeepsForeignKeysAndOIDC(t *testing.T) {
	h := newTestHandler(t, strings.Replace(foreignJSON, `"version": "v1.3.1"`, `"version": "v1.2.0"`, 1))
	old := time.Now().Add(-30 * time.Minute).Truncate(time.Second)
	setModTime(t, h.path, old)
	if err := h.Init(); err != nil {
		t.Fatal(err)
	}
	m := readStrict(t, h.path)
	if m["version"] != testVer {
		t.Errorf("version = %v", m["version"])
	}
	assertForeignKept(t, m)
	if l, _ := get(t, m, "secrets", "default").([]any); len(l) != 1 || l[0] != "d-current-0123456789" {
		t.Errorf("version bump must not rotate secrets: %v", l)
	}
	if !modTime(t, h.path).Equal(old) {
		t.Errorf("version bump moved the rotation clock: %v, want %v", modTime(t, h.path), old)
	}
}

func TestInitCreatesMissingFile(t *testing.T) {
	h := newTestHandler(t, "")
	if err := h.Init(); err != nil {
		t.Fatal(err)
	}
	m := readStrict(t, h.path)
	if m["version"] != testVer {
		t.Errorf("version = %v", m["version"])
	}
	for _, key := range secretKeys {
		if l, _ := get(t, m, "secrets", key).([]any); len(l) != 1 {
			t.Errorf("secrets.%s = %v", key, l)
		}
	}
	fi, _ := os.Stat(h.path)
	if fi.Mode().Perm() != 0644 {
		t.Errorf("mode = %v, want 0644 without configured owner", fi.Mode().Perm())
	}
}

func TestShorterWriteLeavesValidJSON(t *testing.T) {
	long := strings.Replace(foreignJSON, `"version": "v1.3.1"`, `"version": "`+strings.Repeat("x", 4096)+`"`, 1)
	h := newTestHandler(t, long)
	if err := h.Init(); err != nil {
		t.Fatal(err)
	}
	assertForeignKept(t, readStrict(t, h.path))
	if b, _ := os.ReadFile(h.path); len(b) >= len(long) {
		t.Fatalf("test needs a shorter write: %d >= %d", len(b), len(long))
	}
}

func TestReadToleratesTrailingBytes(t *testing.T) {
	h := newTestHandler(t, strings.Replace(foreignJSON, `"version": "v1.3.1"`, `"version": "v1.2.0"`, 1)+`fset": ["abc"]}}`)
	if err := h.Init(); err != nil {
		t.Fatal(err)
	}
	assertForeignKept(t, readStrict(t, h.path))
}

func TestFileProtection(t *testing.T) {
	uid, gid := os.Getuid(), os.Getgid()
	cases := []struct {
		name     string
		uid, gid int
		mode     os.FileMode
	}{
		{"no owner", -1, -1, 0644},
		{"owner", uid, -1, 0600},
		{"owner and group", uid, gid, 0600},
		{"group only", -1, gid, 0640},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newTestHandler(t, "")
			h.fileUID, h.fileGID = c.uid, c.gid
			if err := h.Init(); err != nil {
				t.Fatal(err)
			}
			if err := h.SetOIDC(context.Background(), validReq()); err != nil {
				t.Fatal(err)
			}
			fi, _ := os.Stat(h.path)
			if fi.Mode().Perm() != c.mode {
				t.Errorf("mode = %v, want %v", fi.Mode().Perm(), c.mode)
			}
			entries, _ := os.ReadDir(filepath.Dir(h.path))
			if len(entries) != 1 {
				t.Errorf("temp files left behind: %v", entries)
			}
		})
	}
}

func TestInitTightensExistingFile(t *testing.T) {
	h := newTestHandler(t, foreignJSON)
	h.fileUID = os.Getuid()
	before, _ := os.ReadFile(h.path)
	if err := h.Init(); err != nil {
		t.Fatal(err)
	}
	fi, _ := os.Stat(h.path)
	if fi.Mode().Perm() != 0600 {
		t.Errorf("mode = %v, want 0600 once an owner is configured", fi.Mode().Perm())
	}
	after, _ := os.ReadFile(h.path)
	if !bytes.Equal(before, after) {
		t.Error("content changed")
	}
}

func TestWriteFailureKeepsOldFile(t *testing.T) {
	h := newTestHandler(t, foreignJSON)
	h.fileUID = os.Getuid() + 1 // chown to another user fails without privileges
	if os.Getuid() == 0 {
		t.Skip("chown succeeds as root")
	}
	err := h.SetOIDC(context.Background(), validReq())
	if err == nil {
		t.Fatal("expected chown error")
	}
	b, _ := os.ReadFile(h.path)
	if string(b) != foreignJSON {
		t.Error("failed write changed the file")
	}
	entries, _ := os.ReadDir(filepath.Dir(h.path))
	if len(entries) != 1 {
		t.Errorf("temp files left behind: %v", entries)
	}
}
