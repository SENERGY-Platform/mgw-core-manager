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
// never starts them on its own: the core-manager starts them when it starts, starts a crashed one again and stops
// them when it stops.
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

// Config controls starting, monitoring and stopping the containers.
type Config struct {
	ComposePath   string        // the core's compose file
	DockerPath    string        // the docker CLI
	WaitSockets   []string      // unix sockets the containers use (the other host binaries'), waited for before each start
	StartTimeout  time.Duration // per start attempt, including the wait for the sockets
	StopTimeout   time.Duration
	StartAttempts int           // failed start attempts before giving up
	CheckInterval time.Duration // how often the monitor checks the containers
	RestartLimit  int           // restarts of a container within RestartWindow before the monitor gives up on it
	RestartWindow time.Duration
}

// Containers are the core containers as seen through the container engine wrapper.
type Containers interface {
	LongRunningContainers() []string
	ContainerStates(ctx context.Context) (map[string]string, error)
	StartContainer(ctx context.Context, name string) error
}

type Handler struct {
	cfg  Config
	ctrs Containers

	// replaceable in tests
	run   func(ctx context.Context, name string, args ...string) error
	dial  func(ctx context.Context, path string) error
	sleep func(ctx context.Context, d time.Duration) error
	now   func() time.Time

	mu      sync.Mutex
	cancel  context.CancelFunc
	done    chan struct{}
	stopped bool
}

// New returns a handler for the containers of the compose file in cfg. ctrs must describe the same containers.
func New(cfg Config, ctrs Containers) *Handler {
	return &Handler{
		cfg:   cfg,
		ctrs:  ctrs,
		run:   runCmd,
		dial:  dialUnix,
		sleep: sleepCtx,
		now:   time.Now,
	}
}

// Start starts the containers in the background, retrying failed attempts, and then keeps them running (see
// monitor). It returns at once.
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

// Stop ends a start still in progress or the monitor and stops the containers. It is meant to be called once, when the
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
	ctx, cf := context.WithTimeout(context.Background(), h.cfg.StopTimeout)
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
			h.monitor(ctx)
			return
		}
		if ctx.Err() != nil {
			return
		}
		if attempt >= h.cfg.StartAttempts {
			util.Logger.Errorf("starting core containers failed %d times, giving up: %s", attempt, err)
			return
		}
		delay := retryDelays[min(attempt, len(retryDelays))-1]
		util.Logger.Warningf("starting core containers failed (attempt %d of %d), retrying in %s: %s", attempt, h.cfg.StartAttempts, delay, err)
		if h.sleep(ctx, delay) != nil {
			return
		}
	}
}

func (h *Handler) startOnce(ctx context.Context) error {
	ctx, cf := context.WithTimeout(ctx, h.cfg.StartTimeout)
	defer cf()
	if err := h.waitForSockets(ctx); err != nil {
		return err
	}
	util.Logger.Info("starting core containers ...")
	return h.compose(ctx, "start")
}

func (h *Handler) waitForSockets(ctx context.Context) error {
	for _, p := range h.cfg.WaitSockets {
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
	return h.run(ctx, h.cfg.DockerPath, "compose", "-f", h.cfg.ComposePath, cmd)
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
