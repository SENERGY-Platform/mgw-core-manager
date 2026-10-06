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

package model

type OIDCSettings struct {
	Enabled     bool     `json:"enabled"`
	IssuerURL   string   `json:"issuer_url"`
	ClientID    string   `json:"client_id"`
	Scope       []string `json:"scope"`
	ExternalURL string   `json:"external_url"`
	HasSecret   bool     `json:"has_secret"`
	ProviderID  string   `json:"provider_id"` // read-only, derived from issuer_url; part of the callback URL
}

type OIDCSettingsReq struct {
	Enabled      bool     `json:"enabled"`
	IssuerURL    string   `json:"issuer_url"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"` // empty -> keep stored secret; required when issuer_url or client_id changes
	Scope        []string `json:"scope"`         // empty -> keep stored scope, ["openid"] if none is stored
	ExternalURL  string   `json:"external_url"`
}
