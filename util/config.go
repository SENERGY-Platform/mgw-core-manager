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
	"github.com/SENERGY-Platform/go-service-base/config-hdl"
	sb_logger "github.com/SENERGY-Platform/go-service-base/logger"
	envldr "github.com/y-du/go-env-loader"
	"github.com/y-du/go-log-level/level"
	"io/fs"
	"os"
	"reflect"
	"time"
)

type JobsConfig struct {
	BufferSize  int   `json:"buffer_size" env_var:"JOBS_BUFFER_SIZE"`
	MaxNumber   int   `json:"max_number" env_var:"JOBS_MAX_NUMBER"`
	CCHInterval int   `json:"cch_interval" env_var:"JOBS_CCH_INTERVAL"`
	JHInterval  int   `json:"jh_interval" env_var:"JOBS_JH_INTERVAL"`
	PJHInterval int64 `json:"pjh_interval" env_var:"JOBS_PJH_INTERVAL"`
	MaxAge      int64 `json:"max_age" env_var:"JOBS_MAX_AGE"`
}

type HttpClientConfig struct {
	CewSocketPath string `json:"cew_socket_path" env_var:"CEW_SOCKET_PATH"`
	Timeout       int64  `json:"timeout" env_var:"HTTP_TIMEOUT"`
}

type CoreServiceConfig struct {
	GatewaySrvName string `json:"gateway_srv_name" env_var:"CORE_GATEWAY_SRV_NAME"`
}

type SocketConfig struct {
	Path     string      `json:"path" env_var:"SOCKET_PATH"`
	GroupID  int         `json:"group_id" env_var:"SOCKET_GROUP_ID"`
	FileMode fs.FileMode `json:"file_mode" env_var:"SOCKET_FILE_MODE"`
}

type LoggerConfig struct {
	Level        level.Level `json:"level" env_var:"LOGGER_LEVEL"`
	Utc          bool        `json:"utc" env_var:"LOGGER_UTC"`
	Path         string      `json:"path" env_var:"LOGGER_PATH"`
	FileName     string      `json:"file_name" env_var:"LOGGER_FILE_NAME"`
	Terminal     bool        `json:"terminal" env_var:"LOGGER_TERMINAL"`
	Microseconds bool        `json:"microseconds" env_var:"LOGGER_MICROSECONDS"`
	Prefix       string      `json:"prefix" env_var:"LOGGER_PREFIX"`
}

type KratosConfig struct {
	Version      string `json:"version" env_var:"KRATOS_VERSION"`
	ConfigPath   string `json:"config_path" env_var:"KRATOS_CONFIG_PATH"`
	SecretLength int    `json:"secret_length" env_var:"KRATOS_SECRET_LENGTH"`
	SecretMaxAge int64  `json:"secret_max_age" env_var:"KRATOS_SECRET_MAX_AGE"`
	Interval     int64  `json:"interval" env_var:"KRATOS_INTERVAL"`
}

type LogHandlerConfig struct {
	Path       string `json:"path" env_var:"LOG_HANDLER_PATH"`
	BufferSize int    `json:"buffer_size" env_var:"LOG_HANDLER_BUFFER_SIZE"`
}

type Config struct {
	Logger            LoggerConfig      `json:"logger" env_var:"LOGGER_CONFIG"`
	Socket            SocketConfig      `json:"socket" env_var:"SOCKET_CONFIG"`
	Jobs              JobsConfig        `json:"jobs" env_var:"JOBS_CONFIG"`
	CoreService       CoreServiceConfig `json:"core_service" env_var:"CORE_SERVICE_CONFIG"`
	HttpClient        HttpClientConfig  `json:"http_client" env_var:"HTTP_CLIENT_CONFIG"`
	Kratos            KratosConfig      `json:"kratos" env_var:"KRATOS_CONFIG"`
	EndpointsConfPath string            `json:"endpoints_conf_path" env_var:"ENDPOINTS_CONF_PATH"`
	ComposeFilePath   string            `json:"compose_file_path" env_var:"COMPOSE_FILE_PATH"`
	CoreID            string            `json:"core_id" env_var:"CORE_ID"`
	ImgPurgeDelay     int64             `json:"img_purge_delay" env_var:"IMG_PURGE_DELAY"`
	LogHandler        LogHandlerConfig  `json:"log_handler" env_var:"LOG_HANDLER_CONFIG"`
}

func NewConfig(path string) (*Config, error) {
	cfg := Config{
		Logger: LoggerConfig{
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
			BufferSize:  200,
			MaxNumber:   20,
			CCHInterval: 500000,
			JHInterval:  500000,
			PJHInterval: 300000000000,
			MaxAge:      172800000000000,
		},
		HttpClient: HttpClientConfig{
			CewSocketPath: "./ce_wrapper.sock",
			Timeout:       10000000000,
		},
		Kratos: KratosConfig{
			SecretLength: 32,
			SecretMaxAge: int64(time.Hour * 168),
			Interval:     int64(time.Hour),
		},
		ImgPurgeDelay: int64(time.Minute),
		LogHandler: LogHandlerConfig{
			BufferSize: 32768,
		},
	}
	err := config_hdl.Load(&cfg, nil, map[reflect.Type]envldr.Parser{reflect.TypeOf(level.Off): sb_logger.LevelParser}, nil, path)
	return &cfg, err
}
