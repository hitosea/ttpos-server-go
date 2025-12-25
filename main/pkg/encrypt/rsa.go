package encrypt

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strings"
)

const (
	// MaxEncryptBlockSize RSA加密最大块大小（PKCS1v15）
	MaxEncryptBlockSize = 117
	// MaxDecryptBlockSize RSA解密最大块大小（PKCS1v15）
	MaxDecryptBlockSize = 128
	// DefaultCharset 默认字符集
	DefaultCharset = "UTF-8"
)

// GetPrivateKey 从文件或字符串创建私钥资源标识符
// 如果 privateKey 是文件路径，则读取文件内容；否则作为PEM字符串处理
// 如果字符串不包含PEM头尾，则自动添加
func GetPrivateKey(privateKey string) (*rsa.PrivateKey, error) {
	var keyData []byte
	var err error

	// 检查是否是文件路径
	if keyData, err = os.ReadFile(privateKey); err != nil {
		// 不是文件路径，作为字符串处理
		keyData = []byte(privateKey)
	}

	keyStr := string(keyData)

	// 如果没有PEM头尾，自动添加
	if !strings.Contains(keyStr, "-----BEGIN PRIVATE KEY-----") {
		keyStr = formatPrivateKey(keyStr)
	}

	// 解析PEM块
	block, _ := pem.Decode([]byte(keyStr))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the private key")
	}

	// 尝试PKCS1格式
	privateKeyObj, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return privateKeyObj, nil
	}

	// 尝试PKCS8格式
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}

	return rsaKey, nil
}

// GetPublicKey 从文件或字符串创建公钥资源标识符
// 如果 publicKey 是文件路径，则读取文件内容；否则作为PEM字符串处理
// 如果字符串不包含PEM头尾，则自动添加
func GetPublicKey(publicKey string) (*rsa.PublicKey, error) {
	var keyData []byte
	var err error

	// 检查是否是文件路径
	if keyData, err = os.ReadFile(publicKey); err != nil {
		// 不是文件路径，作为字符串处理
		keyData = []byte(publicKey)
	}

	keyStr := string(keyData)

	// 如果没有PEM头尾，自动添加
	if !strings.Contains(keyStr, "-----BEGIN PUBLIC KEY-----") {
		keyStr = formatPublicKey(keyStr)
	}

	// 解析PEM块
	block, _ := pem.Decode([]byte(keyStr))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
	}

	// 解析公钥
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPublicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return rsaPublicKey, nil
}

// PubEncrypt RSA公钥分段加密数据后返回base64编码
func PubEncrypt(data string, rsaPublicKey *rsa.PublicKey) (string, error) {
	dataBytes := []byte(data)
	var encrypted []byte

	// 分段加密
	for i := 0; i < len(dataBytes); i += MaxEncryptBlockSize {
		end := i + MaxEncryptBlockSize
		if end > len(dataBytes) {
			end = len(dataBytes)
		}

		chunk := dataBytes[i:end]
		encryptedChunk, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPublicKey, chunk)
		if err != nil {
			return "", fmt.Errorf("encryption failed: %w", err)
		}

		encrypted = append(encrypted, encryptedChunk...)
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// PubDecrypt RSA公钥分段解密（用于解密用私钥加密的数据）
// 注意：这实际上是用公钥解密用私钥加密的数据，需要使用底层RSA操作
func PubDecrypt(data string, rsaPublicKey *rsa.PublicKey) (string, error) {
	encryptBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	var decrypted []byte

	// 分段解密
	for i := 0; i < len(encryptBytes); i += MaxDecryptBlockSize {
		end := i + MaxDecryptBlockSize
		if end > len(encryptBytes) {
			end = len(encryptBytes)
		}

		chunk := encryptBytes[i:end]
		// 使用公钥解密（实际上是验证私钥签名，但不使用哈希）
		decryptedChunk, err := decryptWithPublicKey(chunk, rsaPublicKey)
		if err != nil {
			return "", fmt.Errorf("decryption failed: %w", err)
		}

		decrypted = append(decrypted, decryptedChunk...)
	}

	return string(decrypted), nil
}

// PrivEncrypt RSA私钥分段加密数据后返回base64编码
// 注意：这是用私钥加密数据（不使用哈希），与签名不同
func PrivEncrypt(data string, rsaPrivateKey *rsa.PrivateKey) (string, error) {
	dataBytes := []byte(data)
	var encrypted []byte

	// 分段加密
	for i := 0; i < len(dataBytes); i += MaxEncryptBlockSize {
		end := i + MaxEncryptBlockSize
		if end > len(dataBytes) {
			end = len(dataBytes)
		}

		chunk := dataBytes[i:end]
		encryptedChunk, err := encryptWithPrivateKey(chunk, rsaPrivateKey)
		if err != nil {
			return "", fmt.Errorf("encryption failed: %w", err)
		}

		encrypted = append(encrypted, encryptedChunk...)
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// PrivDecrypt RSA私钥分段解密
func PrivDecrypt(data string, rsaPrivateKey *rsa.PrivateKey) (string, error) {
	encryptBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	var decrypted []byte

	// 分段解密
	for i := 0; i < len(encryptBytes); i += MaxDecryptBlockSize {
		end := i + MaxDecryptBlockSize
		if end > len(encryptBytes) {
			end = len(encryptBytes)
		}

		chunk := encryptBytes[i:end]
		decryptedChunk, err := rsa.DecryptPKCS1v15(rand.Reader, rsaPrivateKey, chunk)
		if err != nil {
			return "", fmt.Errorf("decryption failed: %w", err)
		}

		decrypted = append(decrypted, decryptedChunk...)
	}

	return string(decrypted), nil
}

// Sign RSA私钥生成签名base64编码
func Sign(content string, privateKey *rsa.PrivateKey) (string, error) {
	hashed := sha256.Sum256([]byte(content))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", fmt.Errorf("signing failed: %w", err)
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// Verify RSA公钥验签
func Verify(data, sign string, rsaPublicKey *rsa.PublicKey) (bool, error) {
	signature, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %w", err)
	}

	hashed := sha256.Sum256([]byte(data))
	err = rsa.VerifyPKCS1v15(rsaPublicKey, crypto.SHA256, hashed[:], signature)
	if err != nil {
		return false, nil
	}

	return true, nil
}

// GetSignContent 由map拆分拼接成待签名字符串
// 按照PHP的实现方式：键值拼接，排除sign字段，按key排序
func GetSignContent(params map[string]any) string {
	// 删除sign字段
	delete(params, "sign")

	// 获取所有key并排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接字符串
	var parts []string
	for i, k := range keys {
		v := params[k]
		if CheckEmpty(v) {
			continue
		}

		valueStr := fmt.Sprintf("%v", v)
		// 排除以@开头的值（PHP中的特殊处理）
		if strings.HasPrefix(valueStr, "@") {
			continue
		}

		if i == 0 {
			parts = append(parts, fmt.Sprintf("%s=%s", k, valueStr))
		} else {
			parts = append(parts, fmt.Sprintf("&%s=%s", k, valueStr))
		}
	}

	return strings.Join(parts, "")
}

// CheckEmpty 检查参数是否为空
func CheckEmpty(value any) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []byte:
		return len(v) == 0
	case []string:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	default:
		return false
	}
}

// DecryptStringToArr 带签名字符串转map
// 格式：key1=value1&key2=value2
func DecryptStringToArr(decryptString string) map[string]string {
	dataArr := make(map[string]string)

	if decryptString == "" {
		return dataArr
	}

	arr := strings.Split(decryptString, "&")
	for _, keyVal := range arr {
		keyValArr := strings.SplitN(keyVal, "=", 2)
		if len(keyValArr) == 2 {
			dataArr[keyValArr[0]] = keyValArr[1]
		} else if len(keyValArr) == 1 {
			dataArr[keyValArr[0]] = ""
		}
	}

	return dataArr
}

// ToPemPublicKey 公钥转成PEM格式
func ToPemPublicKey(pubKey string) string {
	if strings.Contains(pubKey, "-----BEGIN PUBLIC KEY-----") {
		return pubKey
	}

	return formatPublicKey(pubKey)
}

// ToPemPrivateKey 私钥转成PEM格式
func ToPemPrivateKey(privateKey string) string {
	if strings.Contains(privateKey, "-----BEGIN PRIVATE KEY-----") {
		return privateKey
	}

	return formatPrivateKey(privateKey)
}

// formatPublicKey 格式化公钥为PEM格式
func formatPublicKey(pubKey string) string {
	// 移除所有空白字符
	pubKey = strings.ReplaceAll(pubKey, " ", "")
	pubKey = strings.ReplaceAll(pubKey, "\n", "")
	pubKey = strings.ReplaceAll(pubKey, "\r", "")
	pubKey = strings.ReplaceAll(pubKey, "\t", "")

	// 每64个字符换行
	formatted := ""
	for i := 0; i < len(pubKey); i += 64 {
		end := i + 64
		if end > len(pubKey) {
			end = len(pubKey)
		}
		if i > 0 {
			formatted += "\n"
		}
		formatted += pubKey[i:end]
	}

	return "-----BEGIN PUBLIC KEY-----\n" + formatted + "\n-----END PUBLIC KEY-----"
}

// formatPrivateKey 格式化私钥为PEM格式
func formatPrivateKey(privateKey string) string {
	// 移除所有空白字符
	privateKey = strings.ReplaceAll(privateKey, " ", "")
	privateKey = strings.ReplaceAll(privateKey, "\n", "")
	privateKey = strings.ReplaceAll(privateKey, "\r", "")
	privateKey = strings.ReplaceAll(privateKey, "\t", "")

	// 每64个字符换行
	formatted := ""
	for i := 0; i < len(privateKey); i += 64 {
		end := i + 64
		if end > len(privateKey) {
			end = len(privateKey)
		}
		if i > 0 {
			formatted += "\n"
		}
		formatted += privateKey[i:end]
	}

	return "-----BEGIN PRIVATE KEY-----\n" + formatted + "\n-----END PRIVATE KEY-----"
}

// encryptWithPrivateKey 使用私钥加密数据（PKCS1v15填充，type 2加密填充）
// 这相当于PHP的openssl_private_encrypt
func encryptWithPrivateKey(data []byte, priv *rsa.PrivateKey) ([]byte, error) {
	// RSA私钥加密实际上是使用私钥进行RSA操作
	// 在PKCS1v15中，这相当于 m^d mod n
	k := (priv.N.BitLen() + 7) / 8
	if len(data) > k-11 {
		return nil, errors.New("message too long")
	}

	// PKCS1v15 type 2 填充（加密填充）
	em := make([]byte, k)
	em[0] = 0
	em[1] = 2
	psLen := k - len(data) - 3
	for i := 0; i < psLen; i++ {
		// 生成非零随机字节
		for {
			b := make([]byte, 1)
			rand.Read(b)
			if b[0] != 0 {
				em[2+i] = b[0]
				break
			}
		}
	}
	em[2+psLen] = 0
	copy(em[3+psLen:], data)

	// RSA加密: c = m^d mod n
	m := new(big.Int).SetBytes(em)
	c := new(big.Int).Exp(m, priv.D, priv.N)
	ciphertext := make([]byte, k)
	copy(ciphertext[k-len(c.Bytes()):], c.Bytes())

	return ciphertext, nil
}

// decryptWithPublicKey 使用公钥解密数据（PKCS1v15填充，type 2加密填充）
// 这相当于PHP的openssl_public_decrypt
func decryptWithPublicKey(ciphertext []byte, pub *rsa.PublicKey) ([]byte, error) {
	k := (pub.N.BitLen() + 7) / 8
	if len(ciphertext) != k {
		return nil, errors.New("ciphertext length mismatch")
	}

	// RSA解密: m = c^e mod n
	c := new(big.Int).SetBytes(ciphertext)
	m := new(big.Int).Exp(c, big.NewInt(int64(pub.E)), pub.N)

	em := make([]byte, k)
	copy(em[k-len(m.Bytes()):], m.Bytes())

	// 去除PKCS1v15 type 2 填充（加密填充）
	if len(em) < 11 {
		return nil, errors.New("decryption error")
	}
	if em[0] != 0 || em[1] != 2 {
		return nil, errors.New("decryption error")
	}

	// 查找分隔符0
	var psEnd int
	for psEnd = 2; psEnd < len(em); psEnd++ {
		if em[psEnd] == 0 {
			break
		}
	}
	psEnd++

	if psEnd == len(em) {
		return nil, errors.New("decryption error")
	}

	return em[psEnd:], nil
}
