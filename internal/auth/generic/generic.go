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
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
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
	AuthorizationServerUrl string   `yaml:"AuthorizationServerUrl" validate:"required"`
	ScopesRequired         []string `yaml:"scopesRequired"`
}

// Returns the auth service type
func (cfg Config) AuthServiceConfigType() string {
	return AuthServiceType
}

// Initialize a generic auth service
func (cfg Config) Initialize() (auth.AuthService, error) {
	// Discover the JWKS URL from the OIDC configuration endpoint
	jwksURL, err := discoverJWKSURL(cfg.AuthorizationServerUrl)
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

func safeDialer() *net.Dialer {
	return &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
		Control: func(network, address string, c syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return err
			}
			ip := net.ParseIP(host)
			if ip == nil {
				return fmt.Errorf("invalid IP address")
			}
			// Block private, loopback, and link-local
			if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
				return fmt.Errorf("connection to internal/private IP blocked: %s", ip)
			}
			return nil
		},
	}
}

var AllowInsecureForTest = false

func discoverJWKSURL(AuthorizationServerUrl string) (string, error) {
	u, err := url.Parse(AuthorizationServerUrl)
	if err != nil || (u.Scheme != "https" && !AllowInsecureForTest) {
		return "", fmt.Errorf("invalid or insecure auth URL: must be HTTPS")
	}

	oidcConfigURL, err := url.JoinPath(AuthorizationServerUrl, ".well-known/openid-configuration")
	if err != nil {
		return "", err
	}

	// HTTP Client
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DialContext:           safeDialer().DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          10,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		// Prevent redirect loops or redirects to internal sites
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	if AllowInsecureForTest {
		client.Transport.(*http.Transport).DialContext = nil
	}

	resp, err := client.Get(oidcConfigURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch OIDC config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	// Limit read size to 1MB to prevent memory exhaustion
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}

	var config struct {
		JWKSURI string `json:"jwks_uri"`
	}
	if err := json.Unmarshal(body, &config); err != nil {
		return "", err
	}

	if config.JWKSURI == "" {
		return "", fmt.Errorf("jwks_uri not found in config")
	}

	// Sanitize the resulting JWKS URI before returning it
	parsedJWKS, err := url.Parse(config.JWKSURI)
	if err != nil || (parsedJWKS.Scheme != "https" && !AllowInsecureForTest) {
		return "", fmt.Errorf("malicious jwks_uri detected")
	}

	return config.JWKSURI, nil
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
	if a.McpEnabled {
		return nil, nil
	}

	authHeader := h.Get("Authorization")
	if authHeader == "" {
		return nil, nil // Return nil, nil if no authorization header is found
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, fmt.Errorf("Authorization header format must be Bearer {token}")
	}
	tokenString := parts[1]

	// Parse and verify the token signature
	var token *jwt.Token
	var err error

	if a.kf != nil {
		token, err = jwt.Parse(tokenString, a.kf.Keyfunc)
	} else {
		// If no keyfunc is configured (AuthURL was empty), we parse without verifying signature
		token, _, err = new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT token: %w", err)
	}

	if a.kf != nil && !token.Valid {
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

	return claims, nil
}
