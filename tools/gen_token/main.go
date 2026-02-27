package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Assistant struct {
	DeviceId  string `json:"device_id"`
	StaffUuid uint64 `json:"staff_uuid"`
}

type Claims struct {
	Source         string    `json:"source"`
	CompanyUuid    uint64    `json:"company_uuid"`
	StaffUuid      uint64    `json:"staff_uuid"`
	MemberUuid     uint64    `json:"member_uuid"`
	DeviceUuid     uint64    `json:"device_uuid"`
	DeviceId       string    `json:"device_id"`
	Assistant      Assistant `json:"assistant"`
	IsRefreshToken bool      `json:"is_refresh_token"`
	Brand          string    `json:"brand"`
	jwt.RegisteredClaims
}

func main() {
	source := flag.String("source", "shop", "Token source: shop/pos/kds/qds/assistant/tablet/mobile/menu/member")
	companyUuid := flag.Uint64("company", 8267304538112000, "Company UUID")
	staffUuid := flag.Uint64("staff", 1724054105, "Staff UUID")
	deviceId := flag.String("device", "claude-test-device", "Device ID")
	deviceUuid := flag.Uint64("device-uuid", 0, "Device UUID (from ttpos_device.uuid)")
	secret := flag.String("secret", "", "JWT secret (default: read from JWT_SECRET env or use dev secret)")
	expire := flag.Int("expire", 86400*90, "Token expire seconds (default: 90 days)")
	flag.Parse()

	jwtSecret := *secret
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
	}
	if jwtSecret == "" {
		jwtSecret = "dkjhd00a08" // dev env default
	}

	claims := Claims{
		Source:      *source,
		CompanyUuid: *companyUuid,
		StaffUuid:   *staffUuid,
		DeviceId:    *deviceId,
		DeviceUuid:  *deviceUuid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(*expire))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating token: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Source:      %s\n", *source)
	fmt.Fprintf(os.Stderr, "Company:     %d\n", *companyUuid)
	fmt.Fprintf(os.Stderr, "Staff:       %d\n", *staffUuid)
	fmt.Fprintf(os.Stderr, "Device:      %s\n", *deviceId)
	fmt.Fprintf(os.Stderr, "DeviceUuid:  %d\n", *deviceUuid)
	fmt.Fprintf(os.Stderr, "Expire:      %s\n", time.Now().Add(time.Second*time.Duration(*expire)).Format("2006-01-02 15:04:05"))
	fmt.Fprintf(os.Stderr, "---\n")
	fmt.Print(tokenString)
}
