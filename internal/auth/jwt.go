package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// jwtClaims represents the expected claims contained in the JWT token.
type jwtClaims struct {
	AccessID           string         `json:"access_id"`
	Actor              map[string]any `json:"act,omitempty"`
	AuthMethods        []string       `json:"amr,omitempty"`
	AuthenticationTime int64          `json:"auth_time,omitempty"`
	ClientID           string         `json:"cid,omitempty"`
	ExpiresAt          int64          `json:"exp,omitempty"`
	GrantType          string         `json:"grant,omitempty"`
	IssuedAt           int64          `json:"iat,omitempty"`
	Issuer             string         `json:"iss,omitempty"`
	JWTID              string         `json:"jti,omitempty"`
	NotBefore          int64          `json:"nbf,omitempty"`
	Namespace          string         `json:"ns,omitempty"`
	Scopes             []string       `json:"scp,omitempty"`
	Subject            string         `json:"sub,omitempty"`
}

// unsafeParseJWT parses a JWT without verifying its signature and returns its claims.
// WARNING: This is intentionally unsafe and must not be used for trust decisions.
func unsafeParseJWT(token string) (*jwtClaims, error) {
	if token == "" {
		return nil, errors.New("jwt: empty token")
	}
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("jwt: malformed token, expected 3 parts, got %d", len(parts))
	}
	payloadSeg := parts[1]

	// Base64 URL decode without padding as per RFC 7515.
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadSeg)
	if err != nil {
		return nil, fmt.Errorf("jwt: decode payload: %w", err)
	}

	var claims jwtClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("jwt: unmarshal claims: %w", err)
	}
	return &claims, nil
}
