package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"io"
	"strings"
	"ttpos-server-go/config"
)

var aesSecretKey = []byte("")

// EncryptAesString 加密字符串
func EncryptAesString(text string) (string, error) {
	if config.Encrypt.AesSecretKey != "" {
		aesSecretKey = []byte(config.Encrypt.AesSecretKey)
	} else {
		return "", errors.New("aes secret key is not set")
	}
	// 创建cipher
	block, err := aes.NewCipher(aesSecretKey)
	if err != nil {
		return "", err
	}
	// 创建GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	// 创建nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	// 加密
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(text), nil)
	// 使用Base32编码并转为小写
	encoded := base32.StdEncoding.EncodeToString(ciphertext)
	// 移除填充字符
	encoded = strings.TrimRight(encoded, "=")
	return strings.ToLower(encoded), nil
}

// DecryptAesString 解密字符串
func DecryptAesString(cryptoText string) (string, error) {
	// 将小写转回大写
	cryptoText = strings.ToUpper(cryptoText)

	// 添加必要的填充
	padding := len(cryptoText) % 8
	if padding > 0 {
		cryptoText += strings.Repeat("=", 8-padding)
	}

	// 解码base32
	ciphertext, err := base32.StdEncoding.DecodeString(cryptoText)
	if err != nil {
		return "", err
	}
	// 创建cipher
	block, err := aes.NewCipher(aesSecretKey)
	if err != nil {
		return "", err
	}
	// 创建GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	// 获取nonce
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("密文太短")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	// 解密
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
