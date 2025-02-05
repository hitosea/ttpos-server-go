package utils

import (
	"crypto/md5"
	"encoding/hex"
)

const (
	// DefaultSalt 默认盐值，与PHP代码保持一致
	DefaultSalt = "jjjshop_salt_2020"
)

// EncryptPassword 加密密码，模拟PHP的 md5(md5($password) . 'jjjshop_salt_2020')
func EncryptPassword(password string) string {
	// 第一次 MD5
	h1 := md5.New()
	h1.Write([]byte(password))
	firstMd5 := hex.EncodeToString(h1.Sum(nil))

	// 拼接盐值后第二次 MD5
	h2 := md5.New()
	h2.Write([]byte(firstMd5 + DefaultSalt))

	return hex.EncodeToString(h2.Sum(nil))
}
