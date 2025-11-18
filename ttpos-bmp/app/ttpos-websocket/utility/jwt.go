package utility

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Source      string `json:"source"`       // 终端
	CompanyUuid uint64 `json:"company_uuid"` // 集团ID
	StaffUuid   uint64 `json:"staff_uuid"`   // 员工ID
	DeviceId    string `json:"device_id"`    // 设备ID
	jwt.RegisteredClaims
}

func GenerateToken(source string, companyUuid, staffUuid uint64, deviceId string, secret string, expire int) (string, error) {
	claims := Claims{
		Source:      source,
		CompanyUuid: companyUuid,
		StaffUuid:   staffUuid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(expire))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
