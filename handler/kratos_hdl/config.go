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

package kratos_hdl

import (
	"math/rand"
	"time"
)

type conf struct {
	Version string  `json:"version"`
	Secrets secrets `json:"secrets"`
}

type secrets struct {
	Default []string `json:"default"`
	Cookie  []string `json:"cookie"`
	Cipher  []string `json:"cipher"`
}

func newConf(ver string, secretLen int) conf {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	return conf{
		Version: ver,
		Secrets: secrets{
			Default: []string{getRandomStr(seededRand, secretLen)},
			Cookie:  []string{getRandomStr(seededRand, secretLen)},
			Cipher:  []string{getRandomStr(seededRand, secretLen)},
		},
	}
}
