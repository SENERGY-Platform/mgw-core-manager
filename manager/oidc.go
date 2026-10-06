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
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
)

func (m *Manager) GetOIDCSettings(ctx context.Context) (lib_model.OIDCSettings, error) {
	return m.kratosHdl.GetOIDC(ctx)
}

// SetOIDCSettings writes the settings and returns the ID of the job restarting Kratos, which applies them.
func (m *Manager) SetOIDCSettings(ctx context.Context, settings lib_model.OIDCSettingsReq) (string, error) {
	// An unknown service name is a server misconfiguration, so it must not surface as the caller's 404.
	if _, err := m.coreSrvHdl.Get(ctx, m.kratosSrvName); err != nil {
		return "", lib_model.NewInternalError(errors.New("identity server service '" + m.kratosSrvName + "': " + err.Error()))
	}
	if err := m.kratosHdl.SetOIDC(ctx, settings); err != nil {
		return "", err
	}
	return m.RestartCoreService(ctx, m.kratosSrvName)
}
