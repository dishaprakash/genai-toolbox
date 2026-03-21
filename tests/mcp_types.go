// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tests

// McpResponse represents a response from the MCP JSON-RPC server payload
type McpResponse struct {
	JSONRPC string     `json:"jsonrpc"`
	ID      any        `json:"id"`
	Result  *McpResult `json:"result,omitempty"`
	Error   *McpError  `json:"error,omitempty"`
}

type McpResult struct {
	Content []McpContent   `json:"content,omitempty"`
	Tools   []McpTool      `json:"tools,omitempty"`
	IsError bool           `json:"isError,omitempty"`
}

type McpContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
	Data     string `json:"data,omitempty"`
}

type McpRequest struct {
	JSONRPC string     `json:"jsonrpc"`
	ID      any        `json:"id"`
	Method  string     `json:"method"`
	Params  *McpParams `json:"params,omitempty"`
}

type McpParams struct {
	Name      string `json:"name"`
	Arguments any    `json:"arguments,omitempty"`
}

type McpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type McpTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"inputSchema,omitempty"`
}
