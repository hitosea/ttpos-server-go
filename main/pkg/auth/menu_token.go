package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/duke-git/lancet/cryptor"
)

type menuToken struct {
	CompanyUuid float64 `json:"a"`
	QrCode      string  `json:"q"`
}

type MenuToken struct {
	CompanyUuid uint64 `json:"company_uuid"`
	QrCode      string `json:"qr_code"`
}

// DecodeMenuToken 解码电子菜单token
func DecodeMenuToken(token string) (*MenuToken, error) {
	decodedToken, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode token: %v", err)
	}

	parts := strings.SplitN(string(decodedToken), ".", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid token format")
	}

	hash, data := parts[0], parts[1]

	dataBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode data: %v", err)
	}

	if cryptor.Md5Byte(dataBytes) == hash {
		var result menuToken

		if err := json.Unmarshal(dataBytes, &result); err != nil {
			return nil, fmt.Errorf("failed to decode JSON data: %v", err)
		}
		companyUuid, err := strconv.ParseUint(fmt.Sprintf("%d", uint64(result.CompanyUuid)), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse company uuid: %v", err)
		}

		return &MenuToken{
			CompanyUuid: companyUuid,
			QrCode:      result.QrCode,
		}, nil
	}

	return nil, fmt.Errorf("hash does not match, token may have been tampered with")
}
