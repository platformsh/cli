package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/go-jose/go-jose/v4/jwt"
)

// jwtClaims represents the expected claims contained in the JWT token.
type jwtClaims struct {
	ExpiresAt *jwt.NumericDate `json:"exp,omitempty"`
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
