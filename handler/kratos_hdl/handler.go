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
	crand "crypto/rand"
	"encoding/binary"
	"errors"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"math/rand"
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
	fileUID   int
	fileGID   int
	fileMu    sync.Mutex
	running   bool
	loopMu    sync.RWMutex
	dChan     chan struct{}
	ctx       context.Context
}

// New creates a handler for the dynamic Kratos config. A fileUID or fileGID below zero leaves that owner unchanged.
func New(ctx context.Context, kratosVer, configPath string, secretLen int, maxAge, interval time.Duration, fileUID, fileGID int) (*Handler, error) {
	if !path.IsAbs(configPath) {
		return nil, errors.New(configPath + " is not an absolute path")
	}
	return &Handler{
		kratosVer: kratosVer,
		path:      configPath,
		secretLen: secretLen,
		maxAge:    maxAge,
		interval:  interval,
		fileUID:   fileUID,
		fileGID:   fileGID,
		dChan:     make(chan struct{}),
		ctx:       ctx,
	}, nil
}

func (h *Handler) Init() error {
	h.fileMu.Lock()
	defer h.fileMu.Unlock()
	fileInfo, err := os.Stat(h.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return h.write(newDocument(h.kratosVer, newRand(), h.secretLen), time.Time{})
		}
		return err
	}
	d, err := readConfig(h.path)
	if err != nil {
		return err
	}
	if d[versionKey] != h.kratosVer {
		d[versionKey] = h.kratosVer
		return h.write(d, fileInfo.ModTime())
	}
	return h.protectFile()
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
	h.fileMu.Lock()
	defer h.fileMu.Unlock()
	fileInfo, err := os.Stat(h.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return h.write(newDocument(h.kratosVer, newRand(), h.secretLen), time.Time{})
		}
		return err
	}
	if time.Since(fileInfo.ModTime()) > h.maxAge {
		d, err := readConfig(h.path)
		if err != nil {
			return err
		}
		d[versionKey] = h.kratosVer
		rotateSecrets(d, newRand(), h.secretLen)
		util.Logger.Debug(logPrefix, " rotating secrets and keys ...")
		return h.write(d, time.Time{})
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

// write must be called with fileMu held. The file's mtime marks the last secret rotation, so writes that do not
// rotate pass the previous mtime to restore it; a zero modTime leaves the new mtime in place.
func (h *Handler) write(d document, modTime time.Time) error {
	if err := writeConfig(h.path, d, fileMode(h.fileUID, h.fileGID), h.fileUID, h.fileGID); err != nil {
		return err
	}
	if !modTime.IsZero() {
		if err := os.Chtimes(h.path, time.Time{}, modTime); err != nil {
			util.Logger.Errorf("%s restoring modification time: %s", logPrefix, err)
		}
	}
	return nil
}

// protectFile applies a configured owner and mode to an existing file, so a newly set owner takes effect at
// startup and not only at the next write.
func (h *Handler) protectFile() error {
	if h.fileUID < 0 && h.fileGID < 0 {
		return nil
	}
	if err := os.Chown(h.path, h.fileUID, h.fileGID); err != nil {
		return err
	}
	return os.Chmod(h.path, fileMode(h.fileUID, h.fileGID))
}

// newRand draws from crypto/rand: the generated values become Kratos' cookie
// and cipher secrets, so a time-seeded source would make them guessable.
func newRand() *rand.Rand {
	return rand.New(cryptoSource{})
}

type cryptoSource struct{}

func (cryptoSource) Int63() int64 {
	return int64(cryptoSource{}.Uint64() & (1<<63 - 1))
}

func (cryptoSource) Uint64() uint64 {
	var b [8]byte
	if _, err := crand.Read(b[:]); err != nil {
		panic(err)
	}
	return binary.LittleEndian.Uint64(b[:])
}

func (cryptoSource) Seed(int64) {}
