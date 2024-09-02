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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"os"
	"path"
	"sync"
	"time"
)

const logPrefix = "[kratos-hdl]"

type Handler struct {
	kratosVer string
	path      string
	secretLen int
	maxAge    time.Duration
	interval  time.Duration
	running   bool
	loopMu    sync.RWMutex
	dChan     chan struct{}
	ctx       context.Context
}

func New(ctx context.Context, kratosVer, configPath string, secretLen int, maxAge, interval time.Duration) (*Handler, error) {
	if !path.IsAbs(configPath) {
		return nil, errors.New(configPath + " is not an absolute path")
	}
	return &Handler{
		kratosVer: kratosVer,
		path:      configPath,
		secretLen: secretLen,
		maxAge:    maxAge,
		interval:  interval,
		dChan:     make(chan struct{}),
		ctx:       ctx,
	}, nil
}

func (h *Handler) Init() error {
	if _, err := os.Stat(h.path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return writeConfig(h.path, newConf(h.kratosVer, h.secretLen))
		}
		return err
	}
	return nil
}

func (h *Handler) Start() {
	go h.run()
}

func (h *Handler) Running() bool {
	h.loopMu.RLock()
	defer h.loopMu.RUnlock()
	return h.running
}

func (h *Handler) Wait() {
	<-h.dChan
}

func (h *Handler) refreshSecrets() error {
	fileInfo, err := os.Stat(h.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return writeConfig(h.path, newConf(h.kratosVer, h.secretLen))
		}
		return err
	}
	if time.Since(fileInfo.ModTime()) > h.maxAge {
		oldConfig, err := readConfig(h.path)
		if err != nil {
			return err
		}
		newConfig := newConf(h.kratosVer, h.secretLen)
		if len(oldConfig.Secrets.Default) > 0 {
			newConfig.Secrets.Default = append(newConfig.Secrets.Default, oldConfig.Secrets.Default[0])
		}
		if len(oldConfig.Secrets.Cookie) > 0 {
			newConfig.Secrets.Cookie = append(newConfig.Secrets.Cookie, oldConfig.Secrets.Cookie[0])
		}
		if len(oldConfig.Secrets.Cipher) > 0 {
			newConfig.Secrets.Cipher = append(newConfig.Secrets.Cipher, oldConfig.Secrets.Cipher[0])
		}
		return writeConfig(h.path, newConfig)
	}
	return nil
}

func (h *Handler) run() {
	h.loopMu.Lock()
	h.running = true
	h.loopMu.Unlock()
	timer := time.NewTimer(h.interval)
	loop := true
	var err error
	for loop {
		select {
		case <-timer.C:
			if err = h.refreshSecrets(); err != nil {
				util.Logger.Errorf("%s %s", logPrefix, err)
			}
			timer.Reset(h.interval)
		case <-h.ctx.Done():
			loop = false
			break
		}
	}
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	h.loopMu.Lock()
	h.running = false
	h.loopMu.Unlock()
	h.dChan <- struct{}{}
}

func writeConfig(p string, c conf) error {
	file, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(c)
}

func readConfig(p string) (conf, error) {
	file, err := os.Open(p)
	if err != nil {
		return conf{}, err
	}
	defer file.Close()
	var c conf
	if err = json.NewDecoder(file).Decode(&c); err != nil {
		return conf{}, err
	}
	return c, nil
}
