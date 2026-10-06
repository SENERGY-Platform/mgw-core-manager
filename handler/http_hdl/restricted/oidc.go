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
	"github.com/SENERGY-Platform/mgw-core-manager/lib"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetOIDCSettingsH
// @Summary Get OIDC settings
// @Description	Get the OpenID Connect login settings of the identity server. The client secret is never returned.
// @Tags Authentication
// @Produce	json
// @Success	200 {object} lib_model.OIDCSettings "settings"
// @Failure	500 {string} string "error message"
// @Router /oidc [get]
func GetOIDCSettingsH(a lib.Api) (string, string, gin.HandlerFunc) {
	return http.MethodGet, lib_model.OIDCPath, func(gc *gin.Context) {
		settings, err := a.GetOIDCSettings(gc.Request.Context())
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.JSON(http.StatusOK, settings)
	}
}

// PutOIDCSettingsH
// @Summary Set OIDC settings
// @Description	Store the OpenID Connect login settings and restart the identity server. An empty client secret keeps the stored one unless the issuer URL or client ID changes, an empty scope keeps the stored one or defaults to openid.
// @Tags Authentication
// @Accept json
// @Produce	plain
// @Param settings body lib_model.OIDCSettingsReq true "settings"
// @Success	200 {string} string "job ID"
// @Failure	400 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /oidc [put]
func PutOIDCSettingsH(a lib.Api) (string, string, gin.HandlerFunc) {
	return http.MethodPut, lib_model.OIDCPath, func(gc *gin.Context) {
		var settings lib_model.OIDCSettingsReq
		if err := gc.ShouldBindJSON(&settings); err != nil {
			_ = gc.Error(lib_model.NewInvalidInputError(err))
			return
		}
		jID, err := a.SetOIDCSettings(gc.Request.Context(), settings)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.String(http.StatusOK, jID)
	}
}
