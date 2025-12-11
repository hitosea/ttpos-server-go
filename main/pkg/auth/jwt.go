package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Assistant struct {
	DeviceId  string `json:"device_id"`  // 收银设备ID
	StaffUuid uint64 `json:"staff_uuid"` // 收银员工ID
}

type Claims struct {
	Source         string    `json:"source"`           // 终端
	CompanyUuid    uint64    `json:"company_uuid"`     // 集团ID
	StaffUuid      uint64    `json:"staff_uuid"`       // 员工ID
	MemberUuid     uint64    `json:"member_uuid"`      // 会员ID
	DeviceUuid     uint64    `json:"device_uuid"`      // 设备Uuid
	DeviceId       string    `json:"device_id"`        // 设备ID
	Assistant      Assistant `json:"assistant"`        // 点餐助手绑定的收银机信息
	IsRefreshToken bool      `json:"is_refresh_token"` // 是否refresh_token
	Brand          string    `json:"brand"`            // 品牌名称
	jwt.RegisteredClaims
}

func GenerateToken(claims Claims, secret string, expire int, isRefreshToken bool) (string, error) {
	claims.IsRefreshToken = isRefreshToken
	claims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(expire))),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
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
