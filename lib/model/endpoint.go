/*
 * Copyright 2023 InfAI (CC SES)
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

package model

type EndpointType = int

type Endpoint struct {
	ID      string       `json:"id"`
	Type    EndpointType `json:"type"`
	Ref     string       `json:"ref"`
	Host    string       `json:"host"`
	Port    *int         `json:"port"`
	IntPath string       `json:"int_path"`
	ExtPath string       `json:"ext_path"`
}

type EndpointFilter struct {
	IDs  []string
	Type *EndpointType
	Ref  string
}
