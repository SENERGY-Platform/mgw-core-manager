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

package http_hdl

import "strings"

func genLabels(sl []string) (l map[string]string) {
	if sl != nil && len(sl) > 0 {
		l = make(map[string]string)
		for _, s := range sl {
			p := strings.Split(s, "=")
			if len(p) > 1 {
				l[p[0]] = p[1]
			} else {
				l[p[0]] = ""
			}
		}
	}
	return
}

func parseStringSlice(s, sep string) []string {
	if s != "" {
		return strings.Split(s, sep)
	}
	return nil
}
