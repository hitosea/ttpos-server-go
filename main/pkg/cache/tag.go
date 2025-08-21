package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type TaggedCache struct {
	cache Cache
	ctx   context.Context
}

func NewTaggedCache(cache Cache) *TaggedCache {
	return &TaggedCache{
		cache: cache,
		ctx:   context.Background(),
	}
}

// 设置带标签的缓存
func (tc *TaggedCache) TagPut(tag, key, value string, expiration time.Duration) error {
	var pipe redis.Pipeliner
	if tc.cache.GetClusterClient() != nil {
		pipe = tc.cache.GetClusterClient().Pipeline()
	} else {
		pipe = tc.cache.GetClient().Pipeline()
	}

	// 1. 设置实际的缓存值
	pipe.Set(tc.ctx, key, value, expiration)

	// 2. 将key添加到tag集合中
	tagKey := fmt.Sprintf("tag:%s", tag)
	pipe.SAdd(tc.ctx, tagKey, key)

	// 3. 为tag集合设置过期时间（可选）
	pipe.Expire(tc.ctx, tagKey, expiration)

	_, err := pipe.Exec(tc.ctx)
	return err
}

// 清理标签下的所有缓存
func (tc *TaggedCache) TagClear(tag string) error {
	tagKey := fmt.Sprintf("tag:%s", tag)

	var client redis.Cmdable
	if tc.cache.GetClusterClient() != nil {
		client = tc.cache.GetClusterClient()
	} else {
		client = tc.cache.GetClient()
	}

	// 1. 获取该标签下的所有键
	keys, err := client.SMembers(tc.ctx, tagKey).Result()
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	// 2. 移除该标签下的所有键
	if err := client.SRem(tc.ctx, tagKey, keys).Err(); err != nil {
		return err
	}

	// 3. 删除所有相关的键
	if err := client.Del(tc.ctx, keys...).Err(); err != nil {
		return err
	}

	// 4. 删除标签
	return client.Del(tc.ctx, tagKey).Err()
}
