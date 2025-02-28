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

package client

import (
	"context"
	"github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"io"
)

func (c *Client) ListLogs(_ context.Context) ([]model.Log, error) {
	panic("not implemented")
}

func (c *Client) GetLog(_ context.Context, id string, numOfLines int) (io.ReadCloser, error) {
	panic("not implemented")
}
