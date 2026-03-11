//go:build integration

package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestCache(t *testing.T) {
	// 创建Redis缓存
	redisCache := NewCache(Redis, Config{
		Host:     "192.168.100.117",
		Port:     "36379",
		Password: "123123",
		DB:       1,
	})

	// 设置缓存
	err := redisCache.Set("test_key", "test_value", 1*time.Hour)
	if err != nil {
		t.Errorf("Failed to set Redis cache: %v", err)
	}

	// 获取缓存
	value, exists := redisCache.Get("test_key")
	if !exists {
		t.Error("Failed to get value from Redis cache")
	}
	fmt.Printf("Redis cache value: %v\n", value)

	// 创建Go-Cache缓存
	goCache := NewCache(GoCache, Config{})

	// 设置缓存
	err = goCache.Set("test_key", "test_value", 1*time.Hour)
	if err != nil {
		t.Errorf("Failed to set Go-Cache: %v", err)
	}

	// 获取缓存
	value, exists = goCache.Get("test_key")
	if !exists {
		t.Error("Failed to get value from Go-Cache")
	}
	fmt.Printf("Go-Cache value: %v\n", value)
}

func TestCache_Set(t *testing.T) {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    []string{"192.168.100.117:36379"},
		Password: "123123",
	})
	ctx := context.Background()
	client.Ping(ctx)

}
