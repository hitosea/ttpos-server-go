package utils

import (
	"strings"
)

// ParseEncrypt 解析字符串
func ParseEncrypt(encrypt, field string) map[string]string {
	parsedMap := make(map[string]string)
	for _, s := range strings.Split(encrypt, ";") {
		kv := strings.SplitN(s, "=", 2)
		if len(kv) == 2 {
			parsedMap[kv[0]] = kv[1]
		}
	}

	if publicKey, ok := parsedMap[field]; ok {
		publicKey = strings.ReplaceAll(publicKey, "-", "+")
		publicKey = strings.ReplaceAll(publicKey, "_", "/")
		publicKey = strings.ReplaceAll(publicKey, "$", "\n")
		if encryptType, ok2 := parsedMap["encrypt_type"]; ok2 {
			if encryptType == "jsencrypt" {
				publicKey = "-----BEGIN PUBLIC KEY-----\n" + publicKey + "\n-----END PUBLIC KEY-----"
			}
		}
		parsedMap[field] = publicKey
	}
	return parsedMap
}
