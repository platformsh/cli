package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// unsafeGetJWTExpiry parses a JWT without verifying its signature and returns its expiry time.
// WARNING: This is intentionally unsafe and must not be used for trust decisions.
func unsafeGetJWTExpiry(token string) (time.Time, error) {
	if token == "" {
		return time.Time{}, errors.New("jwt: empty token")
	}
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return time.Time{}, fmt.Errorf("jwt: malformed token, expected 3 parts, got %d", len(parts))
	}
	payloadSeg := parts[1]

	// Base64 URL decode without padding as per RFC 7515.
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadSeg)
	if err != nil {
		return time.Time{}, fmt.Errorf("jwt: decode payload: %w", err)
	}

	var claims struct {
		ExpiresAt *int64 `json:"exp,omitempty"`
	}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return time.Time{}, fmt.Errorf("jwt: unmarshal claims: %w", err)
	}

	if claims.ExpiresAt == nil {
		return time.Time{}, errors.New("jwt: no expiry time found")
	}

	return time.Unix(*claims.ExpiresAt, 0), nil
}
