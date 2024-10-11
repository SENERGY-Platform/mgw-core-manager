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

package api

import (
	"context"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"io"
)

func (a *Api) ListLogs(ctx context.Context) (map[string]lib_model.Log, error) {
	return a.logHandler.List(ctx)
}

func (a *Api) GetLog(ctx context.Context, id string, numOfLines int) (io.ReadCloser, error) {
	return a.logHandler.GetReader(ctx, id, numOfLines)
}
