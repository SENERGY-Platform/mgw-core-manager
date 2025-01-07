/*
 * Copyright 2025 InfAI (CC SES)
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

package util

import (
	"github.com/SENERGY-Platform/mgw-core-manager/lib"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"github.com/gin-gonic/gin"
	"path"
	"strings"
)

type Route func(a lib.Api) (m, p string, hf gin.HandlerFunc)

func SetRoutes(a lib.Api, rg *gin.RouterGroup, routes []Route) {
	set := make(map[string]struct{})
	for _, route := range routes {
		m, p, hf := route(a)
		if _, ok := set[m+p]; ok {
			panic("duplicate route: " + m + " " + path.Join(rg.BasePath(), p))
		}
		set[m+p] = struct{}{}
		rg.Handle(m, p, hf)
		util.Logger.Debug("set route: " + m + " " + path.Join(rg.BasePath(), p))
	}
}

func GenLabels(sl []string) (l map[string]string) {
	if len(sl) > 0 {
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

func ParseStringSlice(s, sep string) []string {
	if s != "" {
		return strings.Split(s, sep)
	}
	return nil
}
