/*
 * Copyright 2024 InfAI (CC SES)
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

package http_hdl

import (
	"github.com/SENERGY-Platform/mgw-container-engine-wrapper/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/lib"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"path"
)

type logQuery struct {
	MaxLines int `form:"max_lines"`
}

func setGetLogsH(a lib.Api, rg *gin.RouterGroup) {
	rg.GET(lib_model.LogsPath, func(gc *gin.Context) {
		logs, err := a.ListLogs(gc.Request.Context())
		if err != nil {
			_ = gc.Error(err)
			return
		}
		gc.JSON(http.StatusOK, logs)
	})
}

func setGetLogH(a lib.Api, rg *gin.RouterGroup) {
	rg.GET(path.Join(lib_model.LogsPath, ":id"), func(gc *gin.Context) {
		query := logQuery{}
		if err := gc.ShouldBindQuery(&query); err != nil {
			_ = gc.Error(model.NewInvalidInputError(err))
			return
		}
		rc, err := a.GetLog(gc.Request.Context(), gc.Param("id"), query.MaxLines)
		if err != nil {
			_ = gc.Error(err)
			return
		}
		defer rc.Close()
		gc.Status(http.StatusOK)
		gc.Header("Transfer-Encoding", "chunked")
		gc.Header("Content-Type", gin.MIMEPlain)
		for {
			var b = make([]byte, 204800)
			n, rErr := rc.Read(b)
			if rErr != nil {
				if rErr == io.EOF {
					if n > 0 {
						_, wErr := gc.Writer.Write(b[:n])
						if wErr != nil {
							_ = gc.Error(model.NewInternalError(wErr))
							return
						}
						gc.Writer.Flush()
					}
					break
				}
				_ = gc.Error(model.NewInternalError(rErr))
				return
			}
			_, wErr := gc.Writer.Write(b[:n])
			if wErr != nil {
				_ = gc.Error(model.NewInternalError(wErr))
				return
			}
			gc.Writer.Flush()
		}
	})
}
