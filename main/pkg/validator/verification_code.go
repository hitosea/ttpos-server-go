package validator

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
)

const (
	CodeCacheKey = "member:login:code:%d:%s"
	CodeCacheTTL = 5 * time.Minute
)

// 获取验证码
func GetCode(cache cache.Cache, companyUuid uint64, phone string) (string, error) {
	// 生成验证码
	code := fmt.Sprintf("%06d", rand.Intn(1000000)) // 生成6位随机数字验证码，范围：000000-999999
	// 如果是否debug模式，则打印验证码
	if config.Server.Mode == "debug" {
		fmt.Println("code", code)
	}
	// 设置缓存key
	cacheKey := fmt.Sprintf(CodeCacheKey, companyUuid, phone)
	// 将验证码存储到缓存中，设置5分钟过期
	if err := cache.Set(cacheKey, code, CodeCacheTTL); err != nil {
		return "", fmt.Errorf("存储验证码失败: %v", err)
	}
	return code, nil
}

// 验证验证码
func VerifyCode(cache cache.Cache, companyUuid uint64, phone string, code string) error {
	cacheKey := fmt.Sprintf(CodeCacheKey, companyUuid, phone)
	cacheCode, ok := cache.Get(cacheKey)
	if !ok || cacheCode == "" || cacheCode.(string) != code {
		return errors.New("验证码不正确")
	}
	cache.Del(cacheKey)
	return nil
}
