// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package generic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/googleapis/genai-toolbox/internal/auth"
)

const AuthServiceType string = "generic"

// validate interface
var _ auth.AuthServiceConfig = Config{}

// Auth service configuration
type Config struct {
	Name                   string   `yaml:"name" validate:"required"`
	Type                   string   `yaml:"type" validate:"required"`
	Audience               string   `yaml:"audience" validate:"required"`
	McpEnabled             bool     `yaml:"mcpEnabled"`
	AuthorizationServerURL string   `yaml:"authorizationServerUrl" validate:"required"`
	ScopesRequired         []string `yaml:"scopesRequired"`
}

// Returns the auth service type
func (cfg Config) AuthServiceConfigType() string {
	return AuthServiceType
}

// Initialize a generic auth service
func (cfg Config) Initialize() (auth.AuthService, error) {
	// Discover the JWKS URL from the OIDC configuration endpoint
	jwksURL, err := discoverJWKSURL(cfg.AuthorizationServerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to discover JWKS URL: %w", err)
	}

	// Create the keyfunc to fetch and cache the JWKS in the background
	kf, err := keyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("failed to create keyfunc from JWKS URL %s: %w", jwksURL, err)
	}

	a := &AuthService{
		Config: cfg,
		kf:     kf,
	}
	return a, nil
}

func discoverJWKSURL(authorizationServerURL string) (string, error) {
	authorizationServerURL = strings.TrimSuffix(authorizationServerURL, "/")
	oidcConfigURL := authorizationServerURL + "/.well-known/openid-configuration"
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(oidcConfigURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch OIDC config, status code %d", resp.StatusCode)
	}

	resp.Body = http.MaxBytesReader(nil, resp.Body, 1<<20) // 1MB limit

	var config map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return "", err
	}

	jwksURI, ok := config["jwks_uri"].(string)
	if !ok || jwksURI == "" {
		return "", fmt.Errorf("jwks_uri not found in OIDC configuration at %s", oidcConfigURL)
	}

	return jwksURI, nil
}

var _ auth.AuthService = AuthService{}

// struct used to store auth service info
type AuthService struct {
	Config
	kf keyfunc.Keyfunc
}

// Returns the auth service type
func (a AuthService) AuthServiceType() string {
	return AuthServiceType
}

func (a AuthService) ToConfig() auth.AuthServiceConfig {
	return a.Config
}

// Returns the name of the auth service
func (a AuthService) GetName() string {
	return a.Name
}

// Verifies generic JWT access token inside the Authorization header
func (a AuthService) GetClaimsFromHeader(ctx context.Context, h http.Header) (map[string]any, error) {
	authHeader := h.Get("Authorization")
	if authHeader == "" {
		return nil, nil // Return nil, nil if no authorization header is found
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, fmt.Errorf("authorization header format must be Bearer {token}")
	}
	tokenString := parts[1]

	// Parse and verify the token signature
	token, err := jwt.Parse(tokenString, a.kf.Keyfunc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse and verify JWT token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid JWT claims format")
	}

	// Validate 'aud' (audience) claim
	aud, err := claims.GetAudience()
	if err != nil {
		return nil, fmt.Errorf("could not parse audience from token: %w", err)
	}

	isAudValid := false
	for _, audItem := range aud {
		if audItem == a.Audience {
			isAudValid = true
			break
		}
	}

	// Some IDPs use 'client_id' instead of 'aud' or put it as a single string, checking that if aud not found or not matched
	if !isAudValid {
		if clientIDClaim, ok := claims["client_id"].(string); ok && clientIDClaim == a.Audience {
			isAudValid = true
		} else if audStr, ok := claims["aud"].(string); ok && audStr == a.Audience {
			isAudValid = true
		}
	}

	if !isAudValid {
		return nil, fmt.Errorf("audience validation failed: expected %s, got %v", a.Audience, aud)
	}

	// Validate 'scope' claim against ScopesRequired
	if len(a.ScopesRequired) > 0 {
		var tokenScopes []string

		switch s := claims["scope"].(type) {
		case string:
			tokenScopes = strings.Split(s, " ") // space-separated string is common
		case []interface{}:
			for _, v := range s {
				if str, ok := v.(string); ok {
					tokenScopes = append(tokenScopes, str)
				}
			}
		}

		scopeMap := make(map[string]bool)
		for _, s := range tokenScopes {
			scopeMap[s] = true
		}

		for _, requiredScope := range a.ScopesRequired {
			if !scopeMap[requiredScope] {
				return nil, fmt.Errorf("missing required scope: %s", requiredScope)
			}
		}
	}

	// Return claims dynamically
	claimsMap := make(map[string]any)
	for k, v := range claims {
		claimsMap[k] = v
	}

	return claimsMap, nil
}
