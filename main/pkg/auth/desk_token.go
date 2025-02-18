package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

type DeskToken struct {
	CompanyUuid    float64 `json:"a"`
	DeskUuid       string  `json:"t"`
	DeskTokenValue string  `json:"q"`
}

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

	// Recalculate the hash of the data for verification
	newHash := sha256.Sum256(dataBytes)
	newHashString := fmt.Sprintf("%x", newHash)

	// Compare the hash values to verify the token's integrity
	if newHashString == hash {
		// Decode the JSON string into a map
		//var result map[string]interface{}
		var result DeskToken

		if err := json.Unmarshal(dataBytes, &result); err != nil {
			return nil, fmt.Errorf("failed to decode JSON data: %v", err)
		}
		return &result, nil
	}

	return nil, fmt.Errorf("hash does not match, token may have been tampered with")
}
