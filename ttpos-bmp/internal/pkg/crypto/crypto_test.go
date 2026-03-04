package crypto

import (
	"fmt"
	"testing"

	"github.com/gogf/gf/v2/crypto/gaes"
	"github.com/gogf/gf/v2/encoding/gbase64"
)

func TestDec(t *testing.T) {
	enc := "PZOwN76KL37qnW/4gwuxhA=="
	// 使用 gbase64.DecodeString 解码
	encryptedData, err := gbase64.DecodeString(enc)
	if err != nil {
		t.Logf("解密失败: %v", err)
	}

	encKey := "IesahquufojahCaiceet7Pha"
	// 使用 gaes.Decrypt 解密
	decryptedData, err := gaes.Decrypt(encryptedData, []byte(encKey))
	if err != nil {
		t.Logf("解密失败: %v", err)
	}

	fmt.Println(string(decryptedData))
}
