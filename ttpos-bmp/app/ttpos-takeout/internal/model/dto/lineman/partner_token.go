package lineman

import "github.com/golang-jwt/jwt/v5"

// PartnerTokenClaims LINE MAN Partner Token JWT Claims
type PartnerTokenClaims struct {
	ClientID    string `json:"clientId"`
	Scope       string `json:"scope"`
	PartnerCode string `json:"partnerCode"`
	jwt.RegisteredClaims
}
