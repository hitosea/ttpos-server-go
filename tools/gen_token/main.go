package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
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

// deskToken 用于 H5 扫码点餐的特殊 token 格式
type deskToken struct {
	CompanyUuid float64 `json:"a"`
	DeskUuid    string  `json:"t"`
	QrcodeToken string  `json:"q"`
}

func main() {
	source := flag.String("source", "shop", "Token source: shop/cashier/assistant/h5/kiosk/tablet/kitchen")
	companyUuid := flag.Uint64("company", 8267304538112000, "Company UUID")
	staffUuid := flag.Uint64("staff", 1724054105, "Staff UUID")
	deviceId := flag.String("device", "", "Device ID (auto-detect from DB if empty)")
	deviceUuid := flag.Uint64("device-uuid", 0, "Device UUID (auto-detect from DB if empty)")
	secret := flag.String("secret", "", "JWT secret (default: read from JWT_SECRET env or use dev secret)")
	expire := flag.Int("expire", 86400*90, "Token expire seconds (default: 90 days)")
	dbDSN := flag.String("db", "", "MySQL DSN for auto device lookup (default: dev env)")
	deskUuid := flag.Uint64("desk-uuid", 0, "Desk UUID (for h5 source, auto-detect if empty)")
	assistantCashierDevice := flag.String("assistant-device", "", "Assistant's bound cashier device_id (auto-detect from DB)")
	flag.Parse()

	jwtSecret := *secret
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
	}
	if jwtSecret == "" {
		jwtSecret = "dkjhd00a08" // dev env default
	}

	// 设置 DB DSN 默认值
	dsn := *dbDSN
	if dsn == "" {
		dsn = os.Getenv("TTPOS_DB_DSN")
	}
	if dsn == "" {
		dsn = "saas:d4f3c7d055516a3b@tcp(192.168.100.1:13306)/"
	}

	// H5 走完全不同的 DeskToken 格式
	if *source == "h5" {
		generateH5Token(dsn, *companyUuid, *deskUuid)
		return
	}

	// 非 shop 端：自动查询设备
	if *source != "shop" && *deviceId == "" {
		autoDetectDevice(dsn, *companyUuid, *source, deviceId, deviceUuid, assistantCashierDevice)
	}
	if *deviceId == "" {
		*deviceId = "claude-test-device" // fallback
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

	// Assistant 端：设置关联的收银机 device_id
	if *source == "assistant" && *assistantCashierDevice != "" {
		claims.Assistant = Assistant{
			DeviceId:  *assistantCashierDevice,
			StaffUuid: *staffUuid,
		}
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
	if *source == "assistant" && *assistantCashierDevice != "" {
		fmt.Fprintf(os.Stderr, "Assistant:   %s (bound cashier)\n", *assistantCashierDevice)
	}
	fmt.Fprintf(os.Stderr, "Expire:      %s\n", time.Now().Add(time.Second*time.Duration(*expire)).Format("2006-01-02 15:04:05"))
	fmt.Fprintf(os.Stderr, "---\n")
	fmt.Print(tokenString)
}

// autoDetectDevice 从 DB 自动查询该 source 的第一个绑定设备
//
// Assistant 端特殊处理：
//   - JWT device_id = 绑定的收银机 device_id（auth.go:1229 校验）
//   - JWT Assistant.DeviceId = 助手自己的 device_id（auth.go:1138 IsDeviceBind 校验）
func autoDetectDevice(dsn string, companyUuid uint64, source string, deviceId *string, deviceUuid *uint64, assistantCashierDevice *string) {
	dbName := fmt.Sprintf("shop%d", companyUuid)
	db, err := sql.Open("mysql", dsn+dbName+"?parseTime=true")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[warn] DB connect failed: %v\n", err)
		return
	}
	defer db.Close()
	db.SetConnMaxLifetime(5 * time.Second)

	if source == "assistant" {
		// Assistant: device_id 要填收银机的，Assistant.DeviceId 填助手自己的
		var assistantDevId string
		var assistantDevUuid uint64
		err = db.QueryRow("SELECT uuid, device_id FROM ttpos_device WHERE source = 'assistant' AND delete_time = 0 LIMIT 1").Scan(&assistantDevUuid, &assistantDevId)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[warn] No assistant device found in DB: %v\n", err)
			return
		}
		fmt.Fprintf(os.Stderr, "[auto] Found assistant device: id=%s uuid=%d\n", assistantDevId, assistantDevUuid)
		*assistantCashierDevice = assistantDevId // Assistant.DeviceId = 助手自己的 ID
		*deviceUuid = assistantDevUuid

		var cashierDevId string
		err = db.QueryRow("SELECT device_id FROM ttpos_device WHERE source = 'cashier' AND delete_time = 0 LIMIT 1").Scan(&cashierDevId)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[warn] No cashier device found for assistant binding: %v\n", err)
			return
		}
		*deviceId = cashierDevId // JWT device_id = 收银机的 ID
		fmt.Fprintf(os.Stderr, "[auto] Bound to cashier device: %s\n", cashierDevId)
		return
	}

	var devId string
	var devUuid uint64
	err = db.QueryRow("SELECT uuid, device_id FROM ttpos_device WHERE source = ? AND delete_time = 0 LIMIT 1", source).Scan(&devUuid, &devId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[warn] No %s device found in DB: %v\n", source, err)
		return
	}

	*deviceId = devId
	*deviceUuid = devUuid
	fmt.Fprintf(os.Stderr, "[auto] Found %s device: id=%s uuid=%d\n", source, devId, devUuid)
}

// generateH5Token 生成 H5 扫码点餐的 DeskToken（Base64 + MD5 格式）
func generateH5Token(dsn string, companyUuid uint64, deskUuidParam uint64) {
	dbName := fmt.Sprintf("shop%d", companyUuid)
	db, err := sql.Open("mysql", dsn+dbName+"?parseTime=true")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: DB connect failed: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	db.SetConnMaxLifetime(5 * time.Second)

	var dUuid uint64
	var qrcodeToken string

	if deskUuidParam > 0 {
		err = db.QueryRow("SELECT uuid, qrcode_token FROM ttpos_desk WHERE uuid = ? AND delete_time = 0", deskUuidParam).Scan(&dUuid, &qrcodeToken)
	} else {
		// 自动选第一个有 qrcode_token 的桌台
		err = db.QueryRow("SELECT uuid, qrcode_token FROM ttpos_desk WHERE delete_time = 0 AND qrcode_token != '' LIMIT 1").Scan(&dUuid, &qrcodeToken)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: No desk found: %v\n", err)
		os.Exit(1)
	}

	// 构造 DeskToken: base64(md5(json).base64(json))
	dt := deskToken{
		CompanyUuid: float64(companyUuid),
		DeskUuid:    fmt.Sprintf("%d", dUuid),
		QrcodeToken: qrcodeToken,
	}
	jsonData, _ := json.Marshal(dt)
	hash := fmt.Sprintf("%x", md5.Sum(jsonData))
	dataB64 := base64.StdEncoding.EncodeToString(jsonData)
	tokenStr := base64.StdEncoding.EncodeToString([]byte(hash + "." + dataB64))

	fmt.Fprintf(os.Stderr, "Source:      h5\n")
	fmt.Fprintf(os.Stderr, "Company:     %d\n", companyUuid)
	fmt.Fprintf(os.Stderr, "DeskUuid:    %d\n", dUuid)
	fmt.Fprintf(os.Stderr, "QrcodeToken: %s\n", qrcodeToken)
	fmt.Fprintf(os.Stderr, "---\n")
	fmt.Print(tokenStr)
}
