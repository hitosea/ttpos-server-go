package utils

import (
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	// BcryptCost bcrypt 成本因子
	// 成本因子为 10 时，验证时间约 60-100ms，平衡安全性和用户体验
	BcryptCost = 10

	// bcrypt 密码前缀
	bcryptPrefixA = "$2a$"
	bcryptPrefixB = "$2b$"
	bcryptPrefixY = "$2y$"
)

// HashPasswordBcrypt 使用 bcrypt 加密密码
// 参数:
//   - password: 明文密码
//
// 返回:
//   - string: bcrypt 哈希值（60位）
//   - error: 加密失败时返回错误
func HashPasswordBcrypt(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPasswordBcrypt 验证 bcrypt 密码
// 参数:
//   - password: 明文密码
//   - hash: bcrypt 哈希值
//
// 返回:
//   - bool: 验证是否成功
func VerifyPasswordBcrypt(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// IsBcryptHash 判断是否为 bcrypt 格式密码
// bcrypt 密码以 $2a$、$2b$ 或 $2y$ 开头
func IsBcryptHash(password string) bool {
	return strings.HasPrefix(password, bcryptPrefixA) ||
		strings.HasPrefix(password, bcryptPrefixB) ||
		strings.HasPrefix(password, bcryptPrefixY)
}

// VerifyPassword 通用密码验证（支持 MD5 和 bcrypt）
// 参数:
//   - password: 明文密码
//   - storedPassword: 存储的密码哈希
//
// 返回:
//   - bool: 验证是否成功
//   - bool: 是否需要升级为 bcrypt（true 表示当前是 MD5 格式且验证成功）
func VerifyPassword(password, storedPassword string) (bool, bool) {
	// 判断是否为 bcrypt 格式
	if IsBcryptHash(storedPassword) {
		// 使用 bcrypt 验证
		isValid := VerifyPasswordBcrypt(password, storedPassword)
		return isValid, false // 已经是 bcrypt，不需要升级
	}

	// 使用旧的 MD5 验证
	md5Password := EncryptPassword(password)
	if md5Password == storedPassword {
		return true, true // 验证成功，需要升级为 bcrypt
	}

	return false, false // 验证失败
}

// UpgradePasswordAsync 异步升级密码为 bcrypt
// 用于登录成功后自动升级 MD5 密码
// 参数:
//   - db: 数据库实例
//   - table: 表名
//   - fieldName: 密码字段名（如 "password" 或 "permission_password"）
//   - idField: ID 字段名（如 "uuid" 或 "id"）
//   - id: 用户 ID
//   - plainPassword: 明文密码
func UpgradePasswordAsync(db *gorm.DB, table, fieldName string, idField string, id uint64, plainPassword string) {
	go func() {
		// 生成 bcrypt 哈希
		newHash, err := HashPasswordBcrypt(plainPassword)
		if err != nil {
			// 记录日志但不影响主流程
			// 下次登录时会重新尝试升级
			return
		}

		// 更新数据库
		db.Table(table).
			Where(idField+" = ?", id).
			Update(fieldName, newHash)
	}()
}
