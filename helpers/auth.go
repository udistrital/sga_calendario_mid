package helpers

import (
	"encoding/base64"
	"encoding/json"
	"strings"
)

func DecodeJWTPayload(segment string) (map[string]interface{}, error) {
	payload, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return nil, err
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}
	return claims, nil
}

func ClaimsDesdeBearer(authHeader string) (map[string]interface{}, bool) {
	if authHeader == "" {
		return nil, false
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, false
	}
	segments := strings.Split(parts[1], ".")
	if len(segments) < 2 {
		return nil, false
	}
	claims, err := DecodeJWTPayload(segments[1])
	return claims, err == nil
}
