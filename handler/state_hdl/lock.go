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
	"errors"
	"os"
	"syscall"

	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
)

// ReadLock returns the holder of the installer's lock, nil if the lock is not held or that can't be determined.
//
// The holder writes itself into the lock file, and the content outlives the lock, so the content alone doesn't
// tell whether the lock is held. To find out, ReadLock tries to take a shared lock without waiting on its own
// descriptor and releases it at once: if that fails, someone holds the lock.
//
// /proc/locks can't be used instead: the installer takes the lock with the flock command, which exits right
// away while the shell keeps the descriptor, and the kernel hides a lock from /proc/locks once the process that
// took it is gone.
func ReadLock(p string) *lib_model.LockStatus {
	if p == "" {
		return nil
	}
	file, err := os.Open(p)
	if err != nil {
		logReadErr(p, err)
		return nil
	}
	defer file.Close()
	held, err := lockHeld(file)
	if err != nil {
		logReadErr(p, err)
		return nil
	}
	if !held {
		return nil
	}
	b, err := readAllLimited(file)
	if err != nil {
		logReadErr(p, err)
		// The lock is held even if its content can't be read.
		return &lib_model.LockStatus{}
	}
	rec := parse(b)
	return &lib_model.LockStatus{
		Holder: rec.values["holder"],
		Pid:    rec.values["pid"],
		Since:  rec.values["since"],
	}
}

// lockHeld reports whether another open file description holds an exclusive flock(2) lock on the file.
func lockHeld(file *os.File) (bool, error) {
	fd := int(file.Fd())
	err := syscall.Flock(fd, syscall.LOCK_SH|syscall.LOCK_NB)
	if errors.Is(err, syscall.EWOULDBLOCK) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, syscall.Flock(fd, syscall.LOCK_UN)
}
