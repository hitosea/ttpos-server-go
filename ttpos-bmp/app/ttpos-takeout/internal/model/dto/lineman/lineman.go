package lineman

import "github.com/golang-jwt/jwt/v5"

// LinemanTokenClaims LINE MAN Token 的 JWT Claims
type LinemanTokenClaims struct {
	ClientID string `json:"client_id"`
	Scope    string `json:"scope,omitempty"`
	jwt.RegisteredClaims
}

