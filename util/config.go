/*
 * Copyright 2023 InfAI (CC SES)
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

package util

import (
	sb_util "github.com/SENERGY-Platform/go-service-base/util"
	"github.com/y-du/go-log-level/level"
	"io/fs"
	"os"
)

type JobsConfig struct {
	BufferSize  int   `json:"buffer_size" env_var:"JOBS_BUFFER_SIZE"`
	MaxNumber   int   `json:"max_number" env_var:"JOBS_MAX_NUMBER"`
	CCHInterval int   `json:"cch_interval" env_var:"JOBS_CCH_INTERVAL"`
	JHInterval  int   `json:"jh_interval" env_var:"JOBS_JH_INTERVAL"`
	MaxAge      int64 `json:"max_age" env_var:"JOBS_MAX_AGE"`
}

type SocketConfig struct {
	Path     string      `json:"path" env_var:"SOCKET_PATH"`
	GroupID  int         `json:"group_id" env_var:"SOCKET_GROUP_ID"`
	FileMode fs.FileMode `json:"file_mode" env_var:"SOCKET_FILE_MODE"`
}

type Config struct {
	Logger            sb_util.LoggerConfig `json:"logger" env_var:"LOGGER_CONFIG"`
	Socket            SocketConfig         `json:"socket" env_var:"SOCKET_CONFIG"`
	Jobs              JobsConfig           `json:"jobs" env_var:"JOBS_CONFIG"`
	EndpointsConfPath string               `json:"endpoints_conf_path" env_var:"ENDPOINTS_CONF_PATH"`
}

func NewConfig(path string) (*Config, error) {
	cfg := Config{
		Logger: sb_util.LoggerConfig{
			Level:        level.Warning,
			Utc:          true,
			Path:         "./",
			FileName:     "mgw_core_manager",
			Microseconds: true,
		},
		Socket: SocketConfig{
			Path:     "./c_manager.sock",
			GroupID:  os.Getgid(),
			FileMode: 0660,
		},
		Jobs: JobsConfig{
			BufferSize:  50,
			MaxNumber:   10,
			CCHInterval: 500000,
			JHInterval:  500000,
			MaxAge:      3600000000,
		},
	}
	err := sb_util.LoadConfig(path, &cfg, nil, nil, nil)
	return &cfg, err
}
