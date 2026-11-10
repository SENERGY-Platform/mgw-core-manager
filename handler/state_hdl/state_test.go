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
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	log_level "github.com/y-du/go-log-level"
	"github.com/y-du/go-log-level/level"
)

func TestMain(m *testing.M) {
	util.Logger, _ = log_level.New(log.New(io.Discard, "", 0), level.Off)
	os.Exit(m.Run())
}

func writeFile(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "state")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestReadUpdate(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    *lib_model.UpdateStatus
	}{
		{"all fields",
			"format=1\nwriter=v1_3_0\nfrom_version=v1.2.3\nto_version=v1.3.0\nstarted=s\nfinished=f\nresult=failed\nphase=switching\nkind=update\nrolled_back=true\ninterrupted=true\nrecovery_failed=true\nstep=reconciling /etc links\nwarning=avahi reload failed\n",
			&lib_model.UpdateStatus{FromVersion: "v1.2.3", ToVersion: "v1.3.0", Started: "s", Finished: "f", Result: "failed", Phase: "switching", Kind: "update", RolledBack: true, Interrupted: true, RecoveryFailed: true, Step: "reconciling /etc links", Warnings: []string{"avahi reload failed"}}},
		{"missing fields stay empty", "result=running\n", &lib_model.UpdateStatus{Result: "running"}},
		{"empty file", "", &lib_model.UpdateStatus{}},
		{"unknown keys are ignored", "result=success\nadded_by_a_newer_installer=x\n", &lib_model.UpdateStatus{Result: "success"}},
		{"unknown values are kept as they are", "phase=a_phase_added_later\n", &lib_model.UpdateStatus{Phase: "a_phase_added_later"}},
		{"booleans other than true are false", "rolled_back=yes\ninterrupted=1\nrecovery_failed=TRUE\n", &lib_model.UpdateStatus{}},
		{"a repeated key keeps its last value", "result=running\nresult=success\n", &lib_model.UpdateStatus{Result: "success"}},
		{"warnings are collected in order", "warning=first\nwarning=second\n", &lib_model.UpdateStatus{Warnings: []string{"first", "second"}}},
		{"a value may contain '='", "step=set a=b\n", &lib_model.UpdateStatus{Step: "set a=b"}},
		{"comments, empty lines and lines without '=' are ignored", "# comment\n\nno separator\nresult=failed\n", &lib_model.UpdateStatus{Result: "failed"}},
		{"windows line endings", "result=success\r\nphase=done\r\n", &lib_model.UpdateStatus{Result: "success", Phase: "done"}},
		{"a missing trailing newline", "result=success", &lib_model.UpdateStatus{Result: "success"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ReadUpdate(writeFile(t, c.content)); !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %+v, want %+v", got, c.want)
			}
		})
	}
}

func TestReadUpdateReturnsNilWithoutARecord(t *testing.T) {
	cases := map[string]string{
		"empty path":   "",
		"missing file": filepath.Join(t.TempDir(), "missing"),
	}
	for name, p := range cases {
		t.Run(name, func(t *testing.T) {
			if got := ReadUpdate(p); got != nil {
				t.Errorf("got %+v, want nil", got)
			}
		})
	}
}

func TestReadUpdateRefusesAnOversizedFile(t *testing.T) {
	p := writeFile(t, "step="+strings.Repeat("a", maxFileLen))
	if got := ReadUpdate(p); got != nil {
		t.Errorf("got %+v, want nil", got)
	}
}

func TestReadUpdateReadsAFileOfExactlyTheMaximumSize(t *testing.T) {
	p := writeFile(t, "step="+strings.Repeat("a", maxFileLen-len("step=")))
	got := ReadUpdate(p)
	if got == nil || len(got.Step) != maxFileLen-len("step=") {
		t.Errorf("got %v, want a step of %d bytes", got, maxFileLen-len("step="))
	}
}

func TestReadCheck(t *testing.T) {
	p := writeFile(t, "check_time=2026-11-09T03:00:00Z\ncheck_result=failed\ncheck_version=v1.3.0\ncheck_step=downloading release v1.3.0\nwarning=could not delete /tmp/mgw\nunknown=x\n")
	want := &lib_model.CheckStatus{Time: "2026-11-09T03:00:00Z", Result: "failed", Version: "v1.3.0", Step: "downloading release v1.3.0", Warnings: []string{"could not delete /tmp/mgw"}}
	if got := ReadCheck(p); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestReadCheckReturnsNilForAMissingFile(t *testing.T) {
	if got := ReadCheck(filepath.Join(t.TempDir(), "missing")); got != nil {
		t.Errorf("got %+v, want nil", got)
	}
}

func TestReadVersion(t *testing.T) {
	cases := []struct {
		name    string
		content *string
		want    string
	}{
		{"trims whitespace", strPtr("v2.13.2\n"), "v2.13.2"},
		{"missing file", nil, ""},
		{"oversized", strPtr(strings.Repeat("v", maxFileLen+1)), ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "version")
			if c.content != nil {
				if err := os.WriteFile(p, []byte(*c.content), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if got := ReadVersion(p); got != c.want {
				t.Errorf("got %d bytes %q, want %q", len(got), truncate(got), c.want)
			}
		})
	}
}

func TestReadVersionReturnsEmptyForAnEmptyPath(t *testing.T) {
	if got := ReadVersion(""); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func strPtr(s string) *string { return &s }

func truncate(s string) string {
	if len(s) > 20 {
		return s[:20] + "..."
	}
	return s
}
