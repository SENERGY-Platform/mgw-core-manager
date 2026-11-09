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

// Package state_hdl reads the state files the installer writes: the version file and the key=value records of
// the last update and version check. Records may be written by a newer installer than this core-manager, so
// reading is tolerant: unknown keys are ignored, missing keys stay empty.
package state_hdl

import (
	"errors"
	"fmt"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"io"
	"io/fs"
	"os"
	"strings"
)

// maxFileLen is the size above which a state file is refused instead of truncated.
const maxFileLen = 64 * 1024

const warningKey = "warning"

// Paths are the locations of the installer's state files. An empty path is not read.
type Paths struct {
	Version string
	Update  string
	Check   string
}

// ReadVersion returns the trimmed content of the version file, empty if it can't be read.
func ReadVersion(p string) string {
	if p == "" {
		return ""
	}
	b, err := readLimited(p)
	if err != nil {
		logReadErr(p, err)
		return ""
	}
	return strings.TrimSpace(string(b))
}

// ReadUpdate returns the update record, nil if there is none or it can't be read.
func ReadUpdate(p string) *lib_model.UpdateStatus {
	rec, ok := readRecord(p)
	if !ok {
		return nil
	}
	return &lib_model.UpdateStatus{
		FromVersion: rec.values["from_version"],
		ToVersion:   rec.values["to_version"],
		Started:     rec.values["started"],
		Finished:    rec.values["finished"],
		Result:      rec.values["result"],
		Phase:       rec.values["phase"],
		Kind:        rec.values["kind"],
		RolledBack:  rec.values["rolled_back"] == "true",
		Interrupted: rec.values["interrupted"] == "true",
		Step:        rec.values["step"],
		Warnings:    rec.warnings,
	}
}

// ReadCheck returns the version check record, nil if there is none or it can't be read.
func ReadCheck(p string) *lib_model.CheckStatus {
	rec, ok := readRecord(p)
	if !ok {
		return nil
	}
	return &lib_model.CheckStatus{
		Time:     rec.values["check_time"],
		Result:   rec.values["check_result"],
		Version:  rec.values["check_version"],
		Step:     rec.values["check_step"],
		Warnings: rec.warnings,
	}
}

type record struct {
	values   map[string]string
	warnings []string
}

func readRecord(p string) (record, bool) {
	if p == "" {
		return record{}, false
	}
	b, err := readLimited(p)
	if err != nil {
		logReadErr(p, err)
		return record{}, false
	}
	return parse(b), true
}

// parse reads one key=value pair per line: the key is everything before the first '=', the value everything
// after it. Lines without '=', empty lines and lines starting with '#' are ignored. A repeated key keeps its
// last value, except for warnings, which are collected in order.
func parse(b []byte) record {
	rec := record{values: make(map[string]string)}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		if key == warningKey {
			rec.warnings = append(rec.warnings, value)
			continue
		}
		rec.values[key] = value
	}
	return rec
}

// readLimited fails instead of truncating if the file exceeds maxFileLen.
func readLimited(p string) ([]byte, error) {
	file, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	b, err := io.ReadAll(io.LimitReader(file, maxFileLen+1))
	if err != nil {
		return nil, err
	}
	if len(b) > maxFileLen {
		return nil, fmt.Errorf("file exceeds %d bytes", maxFileLen)
	}
	return b, nil
}

func logReadErr(p string, err error) {
	if errors.Is(err, fs.ErrNotExist) {
		util.Logger.Debugf("no %s", p)
		return
	}
	util.Logger.Warningf("reading %s failed: %s", p, err)
}
