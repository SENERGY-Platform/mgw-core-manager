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
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
)

const (
	versionKey = "version"
	secretsKey = "secrets"
)

var secretKeys = []string{"default", "cookie", "cipher"}

// document is the generic view of the dynamic config file, so keys the handler does not own survive every rewrite.
type document map[string]any

func newDocument(ver string, random *rand.Rand, secretLen int) document {
	d := document{versionKey: ver}
	rotateSecrets(d, random, secretLen)
	return d
}

// rotateSecrets puts a new secret in front of each list and keeps only the previous first secret, so data signed
// before the rotation stays verifiable. Other keys of the secrets object are left untouched.
func rotateSecrets(d document, random *rand.Rand, secretLen int) {
	s, ok := d[secretsKey].(map[string]any)
	if !ok {
		s = make(map[string]any)
		d[secretsKey] = s
	}
	for _, key := range secretKeys {
		list := []any{getRandomStr(random, secretLen)}
		if old := firstString(s[key]); old != "" {
			list = append(list, old)
		}
		s[key] = list
	}
}

func firstString(v any) string {
	if l, ok := v.([]any); ok && len(l) > 0 {
		if s, ok := l[0].(string); ok {
			return s
		}
	}
	return ""
}

// lookup returns the object found by following keys, or nil if any step is missing or not an object.
func lookup(m map[string]any, keys ...string) map[string]any {
	for _, key := range keys {
		next, ok := m[key].(map[string]any)
		if !ok {
			return nil
		}
		m = next
	}
	return m
}

// ensure returns the object found by following keys and creates every missing or non-object step.
func ensure(m map[string]any, keys ...string) map[string]any {
	for _, key := range keys {
		next, ok := m[key].(map[string]any)
		if !ok {
			next = make(map[string]any)
			m[key] = next
		}
		m = next
	}
	return m
}

// readConfig decodes only the first JSON value, because files written before writes were atomic can carry
// trailing bytes after it. Numbers are kept as json.Number so foreign integers are not rounded through float64.
func readConfig(p string) (document, error) {
	file, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.UseNumber()
	var d document
	if err = decoder.Decode(&d); err != nil {
		return nil, err
	}
	if d == nil {
		d = make(document)
	}
	return d, nil
}

// writeConfig replaces the file via rename, so Kratos never reads a partial file and the secrets are never
// readable with a wider mode than requested (the temp file is created with 0600).
func writeConfig(p string, d document, mode os.FileMode, uid, gid int) (err error) {
	tmp, err := os.CreateTemp(filepath.Dir(p), "."+filepath.Base(p)+".tmp-*")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmp.Name())
		}
	}()
	if err = json.NewEncoder(tmp).Encode(d); err != nil {
		return err
	}
	if err = tmp.Chmod(mode); err != nil {
		return err
	}
	if uid >= 0 || gid >= 0 {
		if err = tmp.Chown(uid, gid); err != nil {
			return err
		}
	}
	if err = tmp.Sync(); err != nil {
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), p)
}

// fileMode keeps the file world-readable only when no owner is configured, which is what existing installs rely on.
func fileMode(uid, gid int) os.FileMode {
	switch {
	case uid >= 0:
		return 0600
	case gid >= 0:
		return 0640
	default:
		return 0644
	}
}
