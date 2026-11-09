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

package state_hdl

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"

	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
)

const lockContent = "holder=update\npid=1234\nsince=2026-11-04T03:00:00Z\n"

// holdLock takes an exclusive flock on p through its own open file description, like the installer's flock does,
// and releases it when the test ends.
func holdLock(t *testing.T, p string) {
	t.Helper()
	f, err := os.OpenFile(p, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatal(err)
	}
}

func TestReadLockReportsAHeldLock(t *testing.T) {
	p := writeFile(t, lockContent)
	holdLock(t, p)
	want := &lib_model.LockStatus{Holder: "update", Pid: "1234", Since: "2026-11-04T03:00:00Z"}
	if got := ReadLock(p); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestReadLockReportsAHeldLockWithoutContent(t *testing.T) {
	p := writeFile(t, "")
	holdLock(t, p)
	if got := ReadLock(p); !reflect.DeepEqual(got, &lib_model.LockStatus{}) {
		t.Errorf("got %+v, want an empty lock status", got)
	}
}

func TestReadLockIgnoresAStaleEntry(t *testing.T) {
	if got := ReadLock(writeFile(t, lockContent)); got != nil {
		t.Errorf("got %+v, want nil", got)
	}
}

func TestReadLockDoesNotTakeTheLock(t *testing.T) {
	p := writeFile(t, lockContent)
	ReadLock(p)
	f, err := os.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Errorf("lock not free after reading: %s", err)
	}
}

func TestReadLockReturnsNilWithoutAFile(t *testing.T) {
	for name, p := range map[string]string{"empty path": "", "missing file": filepath.Join(t.TempDir(), "missing")} {
		t.Run(name, func(t *testing.T) {
			if got := ReadLock(p); got != nil {
				t.Errorf("got %+v, want nil", got)
			}
		})
	}
}

// The installer takes the lock like this: the shell opens the file on descriptor 9, the flock command locks it and
// exits, the shell keeps holding the lock.
func TestReadLockReportsALockHeldByTheShell(t *testing.T) {
	if _, err := exec.LookPath("flock"); err != nil {
		t.Skip("no flock command")
	}
	p := writeFile(t, lockContent)
	cmd := exec.Command("sh", "-c", `exec 9<>"$1" && flock -n 9 && echo locked && read _`, "sh", p)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		stdin.Close()
		cmd.Wait()
	})
	if line, _ := bufio.NewReader(stdout).ReadString('\n'); line != "locked\n" {
		t.Fatalf("the shell did not take the lock: %q", line)
	}
	if got := ReadLock(p); got == nil || got.Holder != "update" {
		t.Errorf("got %+v, want the holder update", got)
	}
}
