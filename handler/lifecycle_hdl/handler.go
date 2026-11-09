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

// Package lifecycle_hdl starts and stops the core containers. The containers have restart policy "no", so docker
// never starts them on its own: the core-manager starts them when it starts and stops them when it stops.
//
// Starting and stopping go through "docker compose start/stop" on the core's compose file, because compose
// honours the services' dependencies (healthy databases, completed migrations), which the container engine
// wrapper doesn't know about. The containers have to exist; the installer creates them.
package lifecycle_hdl

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/SENERGY-Platform/mgw-core-manager/util"
)

// retryDelays are the waits after the first, second, ... failed start; the last one repeats.
var retryDelays = []time.Duration{10 * time.Second, 30 * time.Second, time.Minute, 5 * time.Minute}

// maxOutputLen limits how much of a failed command's output is logged.
const maxOutputLen = 1000

type Handler struct {
	composePath  string
	dockerPath   string
	sockets      []string
	startTimeout time.Duration
	stopTimeout  time.Duration
	maxAttempts  int

	// replaceable in tests
	run   func(ctx context.Context, name string, args ...string) error
	dial  func(ctx context.Context, path string) error
	sleep func(ctx context.Context, d time.Duration) error

	mu      sync.Mutex
	cancel  context.CancelFunc
	done    chan struct{}
	stopped bool
}

// New returns a handler for the containers of the compose file at composePath. Before each start attempt it waits
// for the unix sockets the containers use (the other host binaries' sockets). An attempt that takes longer than
// startTimeout fails; after maxAttempts failed attempts it gives up.
func New(composePath, dockerPath string, sockets []string, startTimeout, stopTimeout time.Duration, maxAttempts int) *Handler {
	return &Handler{
		composePath:  composePath,
		dockerPath:   dockerPath,
		sockets:      sockets,
		startTimeout: startTimeout,
		stopTimeout:  stopTimeout,
		maxAttempts:  maxAttempts,
		run:          runCmd,
		dial:         dialUnix,
		sleep:        sleepCtx,
	}
}

// Start starts the containers in the background, retrying failed attempts. It returns at once.
func (h *Handler) Start() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.cancel != nil || h.stopped {
		return
	}
	ctx, cf := context.WithCancel(context.Background())
	h.cancel = cf
	h.done = make(chan struct{})
	go h.startLoop(ctx, h.done)
}

// Stop ends a start still in progress and stops the containers. It is meant to be called once, when the
// core-manager stops.
func (h *Handler) Stop() error {
	h.mu.Lock()
	h.stopped = true
	cancel, done := h.cancel, h.done
	h.mu.Unlock()
	if cancel != nil {
		cancel()
		<-done
	}
	util.Logger.Info("stopping core containers ...")
	ctx, cf := context.WithTimeout(context.Background(), h.stopTimeout)
	defer cf()
	if err := h.compose(ctx, "stop"); err != nil {
		return fmt.Errorf("stopping core containers failed: %w", err)
	}
	util.Logger.Info("core containers stopped")
	return nil
}

func (h *Handler) startLoop(ctx context.Context, done chan struct{}) {
	defer close(done)
	for attempt := 1; ; attempt++ {
		err := h.startOnce(ctx)
		if err == nil {
			util.Logger.Info("core containers started")
			return
		}
		if ctx.Err() != nil {
			return
		}
		if attempt >= h.maxAttempts {
			util.Logger.Errorf("starting core containers failed %d times, giving up: %s", attempt, err)
			return
		}
		delay := retryDelays[min(attempt, len(retryDelays))-1]
		util.Logger.Warningf("starting core containers failed (attempt %d of %d), retrying in %s: %s", attempt, h.maxAttempts, delay, err)
		if h.sleep(ctx, delay) != nil {
			return
		}
	}
}

func (h *Handler) startOnce(ctx context.Context) error {
	ctx, cf := context.WithTimeout(ctx, h.startTimeout)
	defer cf()
	if err := h.waitForSockets(ctx); err != nil {
		return err
	}
	util.Logger.Info("starting core containers ...")
	return h.compose(ctx, "start")
}

func (h *Handler) waitForSockets(ctx context.Context) error {
	for _, p := range h.sockets {
		for {
			err := h.dial(ctx, p)
			if err == nil {
				break
			}
			if ctx.Err() != nil {
				return fmt.Errorf("socket %s not reachable: %w", p, err)
			}
			if h.sleep(ctx, time.Second) != nil {
				return fmt.Errorf("socket %s not reachable: %w", p, err)
			}
		}
	}
	return nil
}

func (h *Handler) compose(ctx context.Context, cmd string) error {
	return h.run(ctx, h.dockerPath, "compose", "-f", h.composePath, cmd)
}

// runCmd runs a command and returns its output in the error if it fails. On cancellation the command gets SIGTERM
// and a few seconds to end before it is killed.
func runCmd(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Cancel = func() error { return cmd.Process.Signal(syscall.SIGTERM) }
	cmd.WaitDelay = 5 * time.Second
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err == nil {
		return nil
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		err = fmt.Errorf("%w (%w)", ctxErr, err)
	}
	if o := lastBytes(strings.TrimSpace(out.String()), maxOutputLen); o != "" {
		return fmt.Errorf("%w: %s", err, o)
	}
	return err
}

func lastBytes(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return "..." + s[len(s)-n:]
}

func dialUnix(ctx context.Context, path string) error {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "unix", path)
	if err != nil {
		return err
	}
	return conn.Close()
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
