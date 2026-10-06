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
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	log_level "github.com/y-du/go-log-level"
	"github.com/y-du/go-log-level/level"
	"io"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	util.Logger, _ = log_level.New(log.New(io.Discard, "", 0), level.Off)
	os.Exit(m.Run())
}
