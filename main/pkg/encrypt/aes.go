package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

var aesSecretKey = []byte("TTPOS-HITOSEA-SECRET-KEY-HERE!!!")

// EncryptAesString 加密字符串
func EncryptAesString(text string) (string, error) {
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
	// 返回base64编码的密文
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAesString 解密字符串
func DecryptAesString(cryptoText string) (string, error) {
	// 解码base64
	ciphertext, err := base64.StdEncoding.DecodeString(cryptoText)
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
