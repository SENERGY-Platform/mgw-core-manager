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

package shared

import (
	"context"
	"encoding/json"
	gin_mw "github.com/SENERGY-Platform/gin-middleware"
	"github.com/SENERGY-Platform/mgw-core-manager/lib"
	lib_model "github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type fakeApi struct {
	lib.Api
	health lib_model.CoreHealth
}

func (f *fakeApi) GetCoreHealth(_ context.Context) (lib_model.CoreHealth, error) {
	return f.health, nil
}

func TestGetHealthResponseShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	e.Use(gin_mw.ErrorHandler(util.GetStatusCode, ", "))
	method, p, hf := GetHealthH(&fakeApi{health: lib_model.CoreHealth{State: 2, Version: "v1", Services: []lib_model.CoreServiceHealth{{Name: "gw"}}}})
	e.Handle(method, "/"+p, hf)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body)
	}
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"state":    float64(2),
		"version":  "v1",
		"services": []any{map[string]any{"name": "gw", "state": nil, "health": nil, "one_shot": false, "ok": false}},
		"update":   nil,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("body = %v, want %v", got, want)
	}
}
