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
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"
)

const (
	// oidcLegacyProviderID marks a managed entry written before the id was derived from the issuer.
	oidcLegacyProviderID = "sso"
	oidcProviderIDPrefix = "sso-"
	oidcProviderIDHexLen = 12
	oidcProviderType     = "generic"
	oidcRedirectPath     = "/core/auth/"
	oidcMapper           = "local claims = std.extVar('claims'); { identity: { traits: {} } }"
	oidcScopeOpenID      = "openid"
)

var oidcMapperURL = "base64://" + base64.StdEncoding.EncodeToString([]byte(oidcMapper))

var oidcMethodKeys = []string{"selfservice", "methods", "oidc"}

type oidcData struct {
	enabled      bool
	providerID   string
	issuerURL    string
	clientID     string
	clientSecret string
	scope        []string
	externalURL  string
}

func (o oidcData) hasProviderData() bool {
	return o.issuerURL != "" || o.clientID != "" || o.clientSecret != "" || len(o.scope) > 0
}

// complete reports whether a provider entry can be written; Kratos rejects the whole config without a client, and
// the provider id is derived from the issuer.
func (o oidcData) complete() bool {
	return o.issuerURL != "" && o.clientID != "" && o.clientSecret != ""
}

// oidcProviderID derives the provider id from the issuer, because Kratos keys OIDC credentials by provider id and
// subject only. Links made at another issuer therefore no longer match and become inert.
func oidcProviderID(issuerURL string) string {
	sum := sha256.Sum256([]byte(issuerURL))
	return oidcProviderIDPrefix + hex.EncodeToString(sum[:])[:oidcProviderIDHexLen]
}

// isManagedProviderID matches the ids this handler writes, so providers configured by other means are left alone.
func isManagedProviderID(id string) bool {
	if id == oidcLegacyProviderID {
		return true
	}
	h, ok := strings.CutPrefix(id, oidcProviderIDPrefix)
	if !ok || len(h) != oidcProviderIDHexLen {
		return false
	}
	for i := 0; i < len(h); i++ {
		if c := h[i]; (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

func (h *Handler) GetOIDC(_ context.Context) (lib_model.OIDCSettings, error) {
	h.fileMu.Lock()
	defer h.fileMu.Unlock()
	d, err := readConfig(h.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return lib_model.OIDCSettings{Scope: []string{}}, nil
		}
		return lib_model.OIDCSettings{}, lib_model.NewInternalError(err)
	}
	o := readOIDC(d)
	settings := lib_model.OIDCSettings{
		Enabled:     o.enabled,
		IssuerURL:   o.issuerURL,
		ClientID:    o.clientID,
		Scope:       o.scope,
		ExternalURL: o.externalURL,
		HasSecret:   o.clientSecret != "",
		ProviderID:  o.providerID,
	}
	if settings.Scope == nil {
		settings.Scope = []string{}
	}
	return settings, nil
}

func (h *Handler) SetOIDC(_ context.Context, req lib_model.OIDCSettingsReq) error {
	h.fileMu.Lock()
	defer h.fileMu.Unlock()
	var d document
	var modTime time.Time
	fileInfo, err := os.Stat(h.path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return lib_model.NewInternalError(err)
		}
		d = newDocument(h.kratosVer, newRand(), h.secretLen)
	} else {
		if d, err = readConfig(h.path); err != nil {
			return lib_model.NewInternalError(err)
		}
		modTime = fileInfo.ModTime()
	}
	o, err := mergeOIDC(readOIDC(d), req)
	if err != nil {
		return err
	}
	writeOIDC(d, o)
	if err = h.write(d, modTime); err != nil {
		return lib_model.NewInternalError(err)
	}
	return nil
}

func readOIDC(d document) oidcData {
	method := lookup(d, oidcMethodKeys...)
	var o oidcData
	o.enabled, _ = method["enabled"].(bool)
	config := lookup(method, "config")
	if base, ok := config["base_redirect_uri"].(string); ok && strings.HasSuffix(base, oidcRedirectPath) {
		o.externalURL = strings.TrimSuffix(base, oidcRedirectPath)
	}
	if provider := findProvider(config); provider != nil {
		o.providerID, _ = provider["id"].(string)
		o.issuerURL, _ = provider["issuer_url"].(string)
		o.clientID, _ = provider["client_id"].(string)
		o.clientSecret, _ = provider["client_secret"].(string)
		if l, ok := provider["scope"].([]any); ok {
			for _, v := range l {
				if s, ok := v.(string); ok {
					o.scope = append(o.scope, s)
				}
			}
		}
	}
	return o
}

// writeOIDC only touches the enabled flag, base_redirect_uri and the managed provider entry. The entry for the current
// id replaces every other managed entry in place; other providers, and other keys of an entry with the same id, are kept.
func writeOIDC(d document, o oidcData) {
	method := ensure(d, oidcMethodKeys...)
	method["enabled"] = o.enabled
	if o.externalURL == "" && !o.complete() {
		return
	}
	config := ensure(method, "config")
	if o.externalURL != "" {
		config["base_redirect_uri"] = o.externalURL + oidcRedirectPath
	}
	if !o.complete() {
		return
	}
	id := oidcProviderID(o.issuerURL)
	providers, _ := config["providers"].([]any)
	kept := make([]any, 0, len(providers)+1)
	var provider map[string]any
	pos := -1
	for _, p := range providers {
		if pm, ok := p.(map[string]any); ok {
			if pid, _ := pm["id"].(string); isManagedProviderID(pid) {
				if pos < 0 {
					pos = len(kept)
				}
				if pid == id && provider == nil {
					provider = pm
				}
				continue
			}
		}
		kept = append(kept, p)
	}
	if provider == nil {
		provider = map[string]any{}
	}
	if pos < 0 {
		pos = len(kept)
	}
	config["providers"] = slices.Insert(kept, pos, any(provider))
	provider["id"] = id
	provider["provider"] = oidcProviderType
	provider["client_id"] = o.clientID
	provider["client_secret"] = o.clientSecret
	provider["mapper_url"] = oidcMapperURL
	scope := o.scope
	if len(scope) == 0 {
		scope = []string{oidcScopeOpenID}
	}
	provider["scope"] = scope
	provider["issuer_url"] = o.issuerURL
}

func findProvider(config map[string]any) map[string]any {
	providers, _ := config["providers"].([]any)
	for _, p := range providers {
		if provider, ok := p.(map[string]any); ok {
			if id, _ := provider["id"].(string); isManagedProviderID(id) {
				return provider
			}
		}
	}
	return nil
}

// mergeOIDC validates req against the stored settings; enabling requires a complete provider and an empty field keeps
// the stored value. The stored secret is reused only for the same issuer and client, so it never reaches another IdP.
func mergeOIDC(cur oidcData, req lib_model.OIDCSettingsReq) (oidcData, error) {
	issuerURL := strings.TrimSpace(req.IssuerURL)
	clientID := strings.TrimSpace(req.ClientID)
	externalURL := strings.TrimSpace(req.ExternalURL)
	idpChanged := (issuerURL != "" && issuerURL != cur.issuerURL) || (clientID != "" && clientID != cur.clientID)
	var errs []string
	out := cur
	out.enabled = req.Enabled
	if req.Enabled {
		if issuerURL == "" {
			errs = append(errs, "issuer_url is required")
		}
		if externalURL == "" {
			errs = append(errs, "external_url is required")
		}
		if clientID == "" {
			errs = append(errs, "client_id is required")
		}
		if req.ClientSecret == "" && cur.clientSecret == "" {
			errs = append(errs, "client_secret is required")
		}
	}
	if issuerURL != "" {
		if err := checkURL(issuerURL); err != nil {
			errs = append(errs, "issuer_url "+err.Error())
		}
		out.issuerURL = issuerURL
	}
	if externalURL != "" {
		if err := checkURL(externalURL); err != nil {
			errs = append(errs, "external_url "+err.Error())
		}
		out.externalURL = strings.TrimRight(externalURL, "/")
	}
	if clientID != "" {
		out.clientID = clientID
	}
	if req.ClientSecret != "" {
		out.clientSecret = req.ClientSecret
	} else if idpChanged {
		if cur.clientSecret != "" {
			errs = append(errs, "client_secret is required when issuer_url or client_id changes")
		}
		out.clientSecret = ""
	}
	if len(req.Scope) > 0 {
		if err := checkScope(req.Scope); err != nil {
			errs = append(errs, err.Error())
		}
		out.scope = slices.Clone(req.Scope)
	} else if req.Enabled && len(out.scope) == 0 {
		out.scope = []string{oidcScopeOpenID}
	}
	if len(errs) == 0 && out.hasProviderData() && !out.complete() {
		errs = append(errs, "issuer_url, client_id and client_secret are required to store provider data")
	}
	if len(errs) > 0 {
		return cur, lib_model.NewInvalidInputError(errors.New(strings.Join(errs, "; ")))
	}
	return out, nil
}

// checkURL only accepts absolute http(s) URLs without query, fragment or user info, since the value becomes an
// issuer or a redirect base that Kratos compares and appends paths to.
func checkURL(s string) error {
	u, err := url.Parse(s)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return errors.New("must be an absolute http(s) URL")
	}
	if strings.ContainsAny(s, "?#") || u.User != nil {
		return errors.New("must not contain a query, fragment or user info")
	}
	return nil
}

// checkScope enforces the scope-token grammar of RFC 6749 section 3.3 and the openid scope the generic provider needs.
func checkScope(scope []string) error {
	for _, s := range scope {
		if s == "" {
			return errors.New("scope must not contain empty values")
		}
		for i := 0; i < len(s); i++ {
			c := s[i]
			if c < 0x21 || c == 0x22 || c == 0x5c || c > 0x7e {
				return fmt.Errorf("scope %q contains an invalid character", s)
			}
		}
	}
	if !slices.Contains(scope, oidcScopeOpenID) {
		return errors.New("scope must contain " + oidcScopeOpenID)
	}
	return nil
}
