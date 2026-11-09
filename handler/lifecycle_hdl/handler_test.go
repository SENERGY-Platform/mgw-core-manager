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

package lifecycle_hdl

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/mgw-core-manager/util"
	log_level "github.com/y-du/go-log-level"
	"github.com/y-du/go-log-level/level"
)

func TestMain(m *testing.M) {
	util.Logger, _ = log_level.New(log.New(io.Discard, "", 0), level.Off)
	os.Exit(m.Run())
}

// fake records the commands and waits and answers each "compose start" with the next result from startResults
// (nil once they run out).
type fake struct {
	mu           sync.Mutex
	commands     []string
	sleeps       []time.Duration
	startResults []error
	blockStart   bool // "compose start" blocks until its context ends
	unreachable  map[string]int
}

func (f *fake) run(ctx context.Context, name string, args ...string) error {
	f.mu.Lock()
	cmd := strings.Join(append([]string{name}, args...), " ")
	f.commands = append(f.commands, cmd)
	var err error
	isStart := strings.HasSuffix(cmd, " start")
	if isStart && len(f.startResults) > 0 {
		err, f.startResults = f.startResults[0], f.startResults[1:]
	}
	block := isStart && f.blockStart
	f.mu.Unlock()
	if block {
		<-ctx.Done()
		return ctx.Err()
	}
	return err
}

func (f *fake) sleep(ctx context.Context, d time.Duration) error {
	f.mu.Lock()
	f.sleeps = append(f.sleeps, d)
	f.mu.Unlock()
	return ctx.Err()
}

// dial fails the given number of times per socket, then succeeds.
func (f *fake) dial(_ context.Context, p string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.unreachable[p] > 0 {
		f.unreachable[p]--
		return errors.New("connection refused")
	}
	return nil
}

func newHandler(f *fake, sockets []string, maxAttempts int) *Handler {
	h := New("/core/docker-compose.yml", "docker", sockets, time.Minute, time.Minute, maxAttempts)
	h.run, h.sleep, h.dial = f.run, f.sleep, f.dial
	return h
}

// startAndWait runs the start loop to its end.
func startAndWait(h *Handler) {
	h.Start()
	<-h.done
}

const (
	composeStart = "docker compose -f /core/docker-compose.yml start"
	composeStop  = "docker compose -f /core/docker-compose.yml stop"
)

func TestStartStartsTheContainers(t *testing.T) {
	f := &fake{}
	startAndWait(newHandler(f, nil, 3))
	if !reflect.DeepEqual(f.commands, []string{composeStart}) || len(f.sleeps) != 0 {
		t.Errorf("commands = %q, sleeps = %v", f.commands, f.sleeps)
	}
}

func TestStartRetriesWithGrowingDelays(t *testing.T) {
	fail := errors.New("dependency failed to start")
	f := &fake{startResults: []error{fail, fail, fail, fail, fail}}
	startAndWait(newHandler(f, nil, 10))
	if len(f.commands) != 6 {
		t.Errorf("got %d attempts, want 6: %q", len(f.commands), f.commands)
	}
	want := []time.Duration{10 * time.Second, 30 * time.Second, time.Minute, 5 * time.Minute, 5 * time.Minute}
	if !reflect.DeepEqual(f.sleeps, want) {
		t.Errorf("delays = %v, want %v", f.sleeps, want)
	}
}

func TestStartGivesUpAfterMaxAttempts(t *testing.T) {
	fail := errors.New("dependency failed to start")
	f := &fake{startResults: []error{fail, fail, fail, fail, fail}}
	startAndWait(newHandler(f, nil, 3))
	if len(f.commands) != 3 || len(f.sleeps) != 2 {
		t.Errorf("commands = %q, sleeps = %v, want 3 attempts and 2 delays", f.commands, f.sleeps)
	}
}

func TestStartWaitsForTheSockets(t *testing.T) {
	f := &fake{unreachable: map[string]int{"/s/a.sock": 2, "/s/b.sock": 1}}
	startAndWait(newHandler(f, []string{"/s/a.sock", "/s/b.sock"}, 3))
	if !reflect.DeepEqual(f.commands, []string{composeStart}) {
		t.Errorf("commands = %q", f.commands)
	}
	if want := []time.Duration{time.Second, time.Second, time.Second}; !reflect.DeepEqual(f.sleeps, want) {
		t.Errorf("sleeps = %v, want %v", f.sleeps, want)
	}
}

func TestStartRetriesIfASocketStaysUnreachable(t *testing.T) {
	f := &fake{unreachable: map[string]int{"/s/a.sock": 1 << 30}}
	h := newHandler(f, []string{"/s/a.sock"}, 2)
	h.startTimeout = 20 * time.Millisecond
	h.sleep = func(ctx context.Context, d time.Duration) error {
		if d == time.Second {
			<-ctx.Done() // the attempt's timeout ends the wait for the socket
			return ctx.Err()
		}
		return f.sleep(ctx, d)
	}
	startAndWait(h)
	if len(f.commands) != 0 {
		t.Errorf("compose ran without the socket: %q", f.commands)
	}
	if want := []time.Duration{10 * time.Second}; !reflect.DeepEqual(f.sleeps, want) {
		t.Errorf("retry delays = %v, want %v", f.sleeps, want)
	}
}

func TestStopEndsARunningStartAndStopsTheContainers(t *testing.T) {
	f := &fake{blockStart: true}
	h := newHandler(f, nil, 3)
	h.Start()
	waitFor(t, func() bool { f.mu.Lock(); defer f.mu.Unlock(); return len(f.commands) == 1 })
	if err := h.Stop(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(f.commands, []string{composeStart, composeStop}) || len(f.sleeps) != 0 {
		t.Errorf("commands = %q, sleeps = %v, want no retry after stop", f.commands, f.sleeps)
	}
}

func TestStopWithoutStart(t *testing.T) {
	f := &fake{}
	h := newHandler(f, nil, 3)
	if err := h.Stop(); err != nil {
		t.Fatal(err)
	}
	h.Start()
	if !reflect.DeepEqual(f.commands, []string{composeStop}) {
		t.Errorf("commands = %q, want only stop, and no start after stop", f.commands)
	}
}

func TestStopReportsAFailure(t *testing.T) {
	h := newHandler(&fake{}, nil, 3)
	h.run = func(context.Context, string, ...string) error { return errors.New("no such service") }
	if err := h.Stop(); err == nil || !strings.Contains(err.Error(), "no such service") {
		t.Errorf("got %v, want the failure", err)
	}
}

func TestRunCmdReturnsTheOutputOfAFailure(t *testing.T) {
	err := runCmd(context.Background(), "sh", "-c", "echo some output; echo the reason >&2; exit 3")
	if err == nil || !strings.Contains(err.Error(), "exit status 3") || !strings.Contains(err.Error(), "the reason") {
		t.Errorf("got %v", err)
	}
	if err := runCmd(context.Background(), "true"); err != nil {
		t.Errorf("got %v, want nil", err)
	}
}

func TestRunCmdEndsOnCancellation(t *testing.T) {
	ctx, cf := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cf()
	start := time.Now()
	err := runCmd(ctx, "sleep", "10")
	if err == nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("got %v, want a deadline error", err)
	}
	if time.Since(start) > 3*time.Second {
		t.Errorf("took %s, want it to end at the deadline", time.Since(start))
	}
}

func TestLastBytes(t *testing.T) {
	if got := lastBytes("abcdef", 3); got != "...def" {
		t.Errorf("got %q", got)
	}
	if got := lastBytes("abc", 3); got != "abc" {
		t.Errorf("got %q", got)
	}
}

func TestDialUnix(t *testing.T) {
	p := filepath.Join(t.TempDir(), "s.sock")
	if err := dialUnix(context.Background(), p); err == nil {
		t.Error("got nil for a missing socket")
	}
	l, err := net.Listen("unix", p)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	go func() {
		if c, err := l.Accept(); err == nil {
			c.Close()
		}
	}()
	if err := dialUnix(context.Background(), p); err != nil {
		t.Errorf("got %v for a listening socket", err)
	}
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("timed out")
		}
		time.Sleep(time.Millisecond)
	}
}
