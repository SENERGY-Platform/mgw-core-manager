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

package log_hdl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"io"
	"os"
	"path"
)

type Log struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type Handler struct {
	logs       map[string]Log
	bufferSize int64
}

func New(logs []Log, bufferSize int) (*Handler, error) {
	lMap := make(map[string]Log)
	for _, log := range logs {
		if !path.IsAbs(log.Path) {
			return nil, fmt.Errorf("path not absolute: %s", log.Path)
		}
		lMap[util.GenHash(log.Path)] = log
	}
	return &Handler{
		logs:       lMap,
		bufferSize: int64(bufferSize),
	}, nil
}

func (h *Handler) List(_ context.Context) ([]lib_model.Log, error) {
	var logs []lib_model.Log
	for id, log := range h.logs {
		logs = append(logs, lib_model.Log{
			ID:          id,
			ServiceName: log.Name,
		})
	}
	return logs, nil
}

func (h *Handler) GetReader(ctx context.Context, id string, numOfLines int) (io.ReadCloser, error) {
	log, ok := h.logs[id]
	if !ok {
		return nil, lib_model.NewNotFoundError(errors.New("not found"))
	}
	file, err := os.Open(log.Path)
	if err != nil {
		return nil, lib_model.NewInternalError(err)
	}
	defer func() {
		if err != nil {
			file.Close()
		}
	}()
	_, err = seek(ctx, file, numOfLines, h.bufferSize)
	if err != nil {
		return nil, lib_model.NewInternalError(err)
	}
	return file, nil
}

func ReadConfig(p string) ([]Log, error) {
	file, err := os.Open(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var config []Log
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
