package crypto

import (
	"context"

	"github.com/gogf/gf/v2/crypto/gaes"
	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

// getAesKey 从配置中获取 AES 密钥
func getAesKey(ctx context.Context) ([]byte, error) {
	aesKeyStr := g.Cfg().MustGet(ctx, "app.aesKey").String()
	if aesKeyStr == "" {
		return nil, gerror.New("AES 密钥未配置")
	}
	return []byte(aesKeyStr), nil
}

// Encrypt 加密字符串
// 使用 gaes.Encrypt 加密，然后使用 gbase64.EncodeToString 编码成字符串
// 参数：
//   - ctx: 上下文
//   - plaintext: 明文
//
// 返回：
//   - string: 加密并编码后的字符串
//   - error: 错误信息
func Encrypt(ctx context.Context, plaintext string) (string, error) {
	if plaintext == "" {
		return "", gerror.New("明文不能为空")
	}

	// 获取 AES 密钥
	aesKey, err := getAesKey(ctx)
	if err != nil {
		return "", err
	}

	// 使用 gaes.Encrypt 加密
	encryptedData, err := gaes.Encrypt([]byte(plaintext), aesKey)
	if err != nil {
		return "", gerror.Wrapf(err, "加密失败")
	}

	// 使用 gbase64.EncodeToString 编码
	encodedStr := gbase64.EncodeToString(encryptedData)

	return encodedStr, nil
}

// EncryptWithCtx 使用全局上下文加密字符串
// 参数：
//   - plaintext: 明文
//
// 返回：
//   - string: 加密并编码后的字符串
//   - error: 错误信息
func EncryptWithCtx(plaintext string) (string, error) {
	return Encrypt(gctx.GetInitCtx(), plaintext)
}

// Decrypt 解密字符串
// 先使用 gbase64.DecodeString 解码，然后使用 gaes.Decrypt 解密
// 参数：
//   - ctx: 上下文
//   - ciphertext: 密文（base64 编码的加密字符串）
//
// 返回：
//   - string: 解密后的明文
//   - error: 错误信息
func Decrypt(ctx context.Context, ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", gerror.New("密文不能为空")
	}

	// 获取 AES 密钥
	aesKey, err := getAesKey(ctx)
	if err != nil {
		return "", err
	}

	// 使用 gbase64.DecodeString 解码
	encryptedData, err := gbase64.DecodeString(ciphertext)
	if err != nil {
		return "", gerror.Wrapf(err, "base64 解码失败")
	}

	// 使用 gaes.Decrypt 解密
	decryptedData, err := gaes.Decrypt(encryptedData, aesKey)
	if err != nil {
		return "", gerror.Wrapf(err, "解密失败")
	}

	return string(decryptedData), nil
}

// DecryptWithCtx 使用全局上下文解密字符串
// 参数：
//   - ciphertext: 密文（base64 编码的加密字符串）
//
// 返回：
//   - string: 解密后的明文
//   - error: 错误信息
func DecryptWithCtx(ciphertext string) (string, error) {
	return Decrypt(gctx.GetInitCtx(), ciphertext)
}
