// Copyright 2024 Google LLC
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

package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/goccy/go-yaml"
)

// AuthConfigFactory defines the function signature for creating a AuthServiceConfig.
type AuthConfigFactory func(ctx context.Context, name string, decoder *yaml.Decoder) (AuthServiceConfig, error)

var authRegistry = make(map[string]AuthConfigFactory)

// Register registers a new Auth type with its factory.
// It returns false if the type is already registered.
func Register(authType string, factory AuthConfigFactory) bool {
	if _, exists := authRegistry[authType]; exists {
		// Auth with this type already exists, do not overwrite.
		return false
	}
	authRegistry[authType] = factory
	return true
}

// DecodeConfig decodes a Auth configuration using the registered factory for the given type.
func DecodeConfig(ctx context.Context, authType string, name string, decoder *yaml.Decoder) (AuthServiceConfig, error) {
	factory, found := authRegistry[authType]
	if !found {
		return nil, fmt.Errorf("unknown Auth type: %q", authType)
	}
	authConfig, err := factory(ctx, name, decoder)
	if err != nil {
		return nil, fmt.Errorf("unable to parse Auth %q as %q: %w", name, authType, err)
	}
	return authConfig, err
}

// AuthServiceConfig is the interface for configuring authentication services.
type AuthServiceConfig interface {
	AuthServiceConfigType() string
	Initialize() (AuthService, error)
}

// AuthService is the interface for authentication services.
type AuthService interface {
	AuthServiceType() string
	GetName() string
	GetClaimsFromHeader(context.Context, http.Header) (map[string]any, error)
	ToConfig() AuthServiceConfig
}
