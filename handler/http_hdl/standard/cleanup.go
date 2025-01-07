/*
 * Copyright 2025 InfAI (CC SES)
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

package standard

import (
	"github.com/SENERGY-Platform/mgw-core-manager/lib"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"path"
)

type purgeImagesQuery struct {
	Repository string `form:"repository"`
	ExcludeTag string `form:"exclude_tag"`
}

// PatchPurgeImagesH
// @Summary Purge images
// @Description	Purge unused images of a repository.
// @Tags Docker
// @Produce	plain
// @Param repository query string true "docker repository name"
// @Param exclude_tag query string false "image tag name to exclude"
// @Success	200 {string} string "job ID"
// @Failure	400 {string} string "error message"
// @Failure	500 {string} string "error message"
// @Router /cleanup/images [patch]
func PatchPurgeImagesH(a lib.Api) (string, string, gin.HandlerFunc) {
	return http.MethodPatch, path.Join(lib_model.CleanupPath, lib_model.ImagesPath), func(gc *gin.Context) {
		query := purgeImagesQuery{}
		if err := gc.ShouldBindQuery(&query); err != nil {
			_ = gc.Error(lib_model.NewInvalidInputError(err))
			return
		}
		jID, err := a.PurgeImages(gc.Request.Context(), query.Repository, query.ExcludeTag)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.String(http.StatusOK, jID)
	}
}
