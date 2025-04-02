package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/duke-git/lancet/cryptor"
	"strconv"
	"strings"
)

type deskToken struct {
	CompanyUuid    float64 `json:"a"`
	DeskUuid       string  `json:"t"`
	DeskTokenValue string  `json:"q"`
}

type DeskToken struct {
	CompanyUuid    uint64 `json:"company_uuid"`
	DeskUuid       uint64 `json:"desk_uuid"`
	DeskTokenValue string `json:"desk_token_value"`
}

// DecodeDeskToken 解码桌台token
func DecodeDeskToken(token string) (*DeskToken, error) {
	// Base64 decode the token
	decodedToken, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode token: %v", err)
	}

	// Split the token into hash and data parts
	parts := strings.SplitN(string(decodedToken), ".", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid token format")
	}

	hash, data := parts[0], parts[1]

	// Base64 decode the data part
	dataBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode data: %v", err)
	}

	// Compare the hash values to verify the token's integrity
	if cryptor.Md5Byte(dataBytes) == hash {
		// Decode the JSON string into a map
		//var result map[string]interface{}
		var result deskToken

		if err := json.Unmarshal(dataBytes, &result); err != nil {
			return nil, fmt.Errorf("failed to decode JSON data: %v", err)
		}
		companyUuid, err := strconv.ParseUint(fmt.Sprintf("%d", uint64(result.CompanyUuid)), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse company uuid: %v", err)
		}
		deskUuid, err := strconv.ParseUint(result.DeskUuid, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse desk uuid: %v", err)
		}
		return &DeskToken{
			CompanyUuid:    companyUuid,
			DeskUuid:       deskUuid,
			DeskTokenValue: result.DeskTokenValue,
		}, nil
	}

	return nil, fmt.Errorf("hash does not match, token may have been tampered with")
}
