package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"jjjshop-server-go/pkg/utils"
	"strings"
)

// GenerateRSAKeyPairPEM 生成RSA密钥对并返回PEM格式的字符串
func GenerateRSAKeyPairPEM(bits int) (KeyPair, error) {
	var keyPair KeyPair
	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return keyPair, err
	}

	// 将私钥转换为PEM格式
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	privateKeyPEM := string(pem.EncodeToMemory(privateKeyBlock))

	// 从私钥中提取公钥
	publicKey := &privateKey.PublicKey

	// 将公钥转换为PKIX, ASN.1 DER格式
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return keyPair, err
	}

	// 将公钥转换为PEM格式
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	publicKeyPEM := string(pem.EncodeToMemory(publicKeyBlock))

	return KeyPair{
		PublicKey:  publicKeyPEM,
		PrivateKey: privateKeyPEM,
		Passphrase: "",
	}, nil
}

// StringToPrivateKey 将 PEM 格式的私钥字符串转换为 rsa.PrivateKey
func StringToPrivateKey(privateKeyPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// DecryptWithPrivateKey RSA私钥解密
func DecryptWithPrivateKey(ciphertext []byte, priv *rsa.PrivateKey) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
}

// DecryptAES 使用AES解密数据
func DecryptAES(encryptedData, key, iv []byte) ([]byte, error) {
	// 创建cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 检查加密数据的长度
	if len(encryptedData) < aes.BlockSize {
		return nil, errors.New("密文太短")
	}

	// 确保加密数据长度是块大小的倍数
	if len(encryptedData)%aes.BlockSize != 0 {
		return nil, errors.New("密文长度必须是块大小的倍数")
	}

	// 检查IV的长度
	if len(iv) != aes.BlockSize {
		return nil, errors.New("IV长度必须等于块大小")
	}

	// 解密
	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(encryptedData))
	mode.CryptBlocks(decrypted, encryptedData)

	// 去除填充
	paddingLen := int(decrypted[len(decrypted)-1])
	if paddingLen > aes.BlockSize || paddingLen == 0 {
		return nil, errors.New("无效的填充")
	}
	return decrypted[:len(decrypted)-paddingLen], nil
}

func JsDecryptMessage(privateKeyPem, encryptedBody string) ([]byte, error) {
	// pem格式的私钥字符串转*rsa.PrivateKey
	privateKey, err := StringToPrivateKey(privateKeyPem)
	if err != nil {
		return nil, err
	}
	// base64 解密encrypted字段值
	decodedBody, err := base64.StdEncoding.DecodeString(encryptedBody)
	if err != nil {
		return nil, err
	}
	// 分割字符串
	parts := strings.Split(string(decodedBody), "||")
	// 最后一部分就是rsa加密后的密钥
	decodedKey, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, err
	}
	// 使用rsa私钥解密，获取aes密钥
	decryptedKey, err := DecryptWithPrivateKey(decodedKey, privateKey) // 解密密钥
	if err != nil {
		return nil, err
	}

	// base64 解密iv
	decodedIv, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}

	// base64 解密加密内容
	decodedContent, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	// 使用AES解密
	raw, err := DecryptAES(decodedContent, decryptedKey, decodedIv)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func JsEncryptMessage(plaintext string, publicKeyPEM string) (string, error) {
	// 解析公钥
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return "", errors.New("failed to parse PEM block containing the public key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}
	rsaPublicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return "", errors.New("not an RSA public key")
	}

	symmetricKey := utils.RandomLetter(32)
	// 使用对称密钥加密数据
	aesBlock, err := aes.NewCipher([]byte(symmetricKey)) // 使用新的变量名 aesBlock
	if err != nil {
		return "", err
	}
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}
	paddedPlaintext := pkcs7Pad([]byte(plaintext), aes.BlockSize)
	ciphertext := make([]byte, len(paddedPlaintext))
	mode := cipher.NewCBCEncrypter(aesBlock, iv) // 使用 aesBlock
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	// 使用 RSA 加密对称密钥
	encryptedKey, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPublicKey, []byte(symmetricKey))
	if err != nil {
		return "", err
	}

	// 编码并拼接结果
	encryptedData := base64.StdEncoding.EncodeToString(append(iv, ciphertext...))
	encryptedKeyStr := base64.StdEncoding.EncodeToString(encryptedKey)
	return encryptedData + "||" + encryptedKeyStr, nil
}

// PKCS7 填充
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}
