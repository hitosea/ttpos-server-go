package utils

import (
	"strings"
)

// PGPParse 解析字符串
func PGPParse(encrypt, field string) map[string]string {
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
		publicKey = "-----BEGIN PGP PUBLIC KEY BLOCK-----\n\n" + publicKey + "\n-----END PGP PUBLIC KEY BLOCK-----"
		parsedMap[field] = publicKey
	}
	return parsedMap
}
