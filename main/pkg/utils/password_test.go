package utils

import (
	"strings"
	"testing"
)

// TestHashPasswordBcrypt 测试 bcrypt 加密
func TestHashPasswordBcrypt(t *testing.T) {
	password := "123456"

	// 测试加密
	hash, err := HashPasswordBcrypt(password)
	if err != nil {
		t.Fatalf("HashPasswordBcrypt 失败: %v", err)
	}

	// 验证哈希格式
	if !strings.HasPrefix(hash, "$2") {
		t.Errorf("bcrypt 哈希格式错误: %s", hash)
	}

	// 验证哈希长度（bcrypt 哈希通常是 60 个字符）
	if len(hash) != 60 {
		t.Errorf("bcrypt 哈希长度错误: 期望 60，实际 %d", len(hash))
	}

	// 相同密码加密两次，哈希应该不同（因为盐值随机）
	hash2, err := HashPasswordBcrypt(password)
	if err != nil {
		t.Fatalf("HashPasswordBcrypt 第二次失败: %v", err)
	}
	if hash == hash2 {
		t.Error("相同密码的两次 bcrypt 哈希应该不同")
	}
}

// TestVerifyPasswordBcrypt 测试 bcrypt 验证
func TestVerifyPasswordBcrypt(t *testing.T) {
	password := "test123456"
	hash, _ := HashPasswordBcrypt(password)

	// 测试正确密码
	if !VerifyPasswordBcrypt(password, hash) {
		t.Error("正确密码验证失败")
	}

	// 测试错误密码
	if VerifyPasswordBcrypt("wrongpassword", hash) {
		t.Error("错误密码验证应该失败")
	}

	// 测试空密码
	if VerifyPasswordBcrypt("", hash) {
		t.Error("空密码验证应该失败")
	}
}

// TestIsBcryptHash 测试 bcrypt 格式判断
func TestIsBcryptHash(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"bcrypt $2a$", "$2a$10$abcdefghijklmnopqrstuvwxyz", true},
		{"bcrypt $2b$", "$2b$10$abcdefghijklmnopqrstuvwxyz", true},
		{"bcrypt $2y$", "$2y$10$abcdefghijklmnopqrstuvwxyz", true},
		{"MD5 32位", "5f4dcc3b5aa765d61d8327deb882cf99", false},
		{"普通字符串", "password123", false},
		{"空字符串", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBcryptHash(tt.password); got != tt.want {
				t.Errorf("IsBcryptHash(%q) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}

// TestVerifyPassword 测试通用密码验证
func TestVerifyPassword(t *testing.T) {
	plainPassword := "123456"

	// 准备测试数据
	md5Password := EncryptPassword(plainPassword)
	bcryptPassword, _ := HashPasswordBcrypt(plainPassword)

	// 测试 1: MD5 密码验证成功，需要升级
	t.Run("MD5密码验证成功", func(t *testing.T) {
		isValid, needUpgrade := VerifyPassword(plainPassword, md5Password)
		if !isValid {
			t.Error("MD5 密码验证应该成功")
		}
		if !needUpgrade {
			t.Error("MD5 密码应该需要升级")
		}
	})

	// 测试 2: bcrypt 密码验证成功，不需要升级
	t.Run("bcrypt密码验证成功", func(t *testing.T) {
		isValid, needUpgrade := VerifyPassword(plainPassword, bcryptPassword)
		if !isValid {
			t.Error("bcrypt 密码验证应该成功")
		}
		if needUpgrade {
			t.Error("bcrypt 密码不应该需要升级")
		}
	})

	// 测试 3: MD5 密码验证失败
	t.Run("MD5密码验证失败", func(t *testing.T) {
		isValid, needUpgrade := VerifyPassword("wrongpassword", md5Password)
		if isValid {
			t.Error("错误密码验证应该失败")
		}
		if needUpgrade {
			t.Error("验证失败时不应该需要升级")
		}
	})

	// 测试 4: bcrypt 密码验证失败
	t.Run("bcrypt密码验证失败", func(t *testing.T) {
		isValid, needUpgrade := VerifyPassword("wrongpassword", bcryptPassword)
		if isValid {
			t.Error("错误密码验证应该失败")
		}
		if needUpgrade {
			t.Error("验证失败时不应该需要升级")
		}
	})

	// 测试 5: 空密码
	t.Run("空密码", func(t *testing.T) {
		isValid, needUpgrade := VerifyPassword("", md5Password)
		if isValid {
			t.Error("空密码验证应该失败")
		}
		if needUpgrade {
			t.Error("验证失败时不应该需要升级")
		}
	})
}

// TestVerifyPassword_EdgeCases 测试边界情况
func TestVerifyPassword_EdgeCases(t *testing.T) {
	// 测试空存储密码
	t.Run("空存储密码", func(t *testing.T) {
		isValid, needUpgrade := VerifyPassword("123456", "")
		if isValid {
			t.Error("空存储密码验证应该失败")
		}
		if needUpgrade {
			t.Error("空存储密码不应该需要升级")
		}
	})

	// 测试无效的 bcrypt 格式
	t.Run("无效bcrypt格式", func(t *testing.T) {
		isValid, needUpgrade := VerifyPassword("123456", "$2a$10$invalid")
		if isValid {
			t.Error("无效 bcrypt 格式验证应该失败")
		}
		if needUpgrade {
			t.Error("无效格式不应该需要升级")
		}
	})
}

// BenchmarkHashPasswordBcrypt 性能测试：bcrypt 加密
func BenchmarkHashPasswordBcrypt(b *testing.B) {
	password := "test123456"
	for i := 0; i < b.N; i++ {
		HashPasswordBcrypt(password)
	}
}

// BenchmarkVerifyPasswordBcrypt 性能测试：bcrypt 验证
func BenchmarkVerifyPasswordBcrypt(b *testing.B) {
	password := "test123456"
	hash, _ := HashPasswordBcrypt(password)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifyPasswordBcrypt(password, hash)
	}
}

// BenchmarkVerifyPassword_MD5 性能测试：MD5 验证
func BenchmarkVerifyPassword_MD5(b *testing.B) {
	password := "test123456"
	md5Password := EncryptPassword(password)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifyPassword(password, md5Password)
	}
}

// BenchmarkVerifyPassword_Bcrypt 性能测试：bcrypt 验证
func BenchmarkVerifyPassword_Bcrypt(b *testing.B) {
	password := "test123456"
	bcryptPassword, _ := HashPasswordBcrypt(password)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifyPassword(password, bcryptPassword)
	}
}
