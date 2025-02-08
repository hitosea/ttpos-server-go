package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Source    string `json:"source"`     // 终端
	CompanyId uint   `json:"company_id"` // 集团ID
	StaffId   uint   `json:"staff_id"`   // 员工ID
	DeviceId  string `json:"device_id"`  // 设备ID
	jwt.RegisteredClaims
}

func GenerateToken(source string, companyId, staffId uint, deviceId string, secret string, expire int) (string, error) {
	claims := Claims{
		Source:    source,
		CompanyId: companyId,
		StaffId:   staffId,
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
