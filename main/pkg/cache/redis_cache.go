package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client        *redis.Client
	clusterClient *redis.ClusterClient
}

type ClusterConfig []Config

func ParseRedisConfig(config Config) ClusterConfig {
	var c ClusterConfig
	hostList := strings.Split(config.Host, ",")
	portList := strings.Split(config.Port, ",")
	if len(hostList) > 1 {
		for index, host := range hostList {
			conf := Config{
				Host:     host,
				Port:     portList[index],
				Password: config.Password,
				DB:       0,
			}
			c = append(c, conf)
		}
	} else {
		conf := Config{
			Host:     config.Host,
			Port:     config.Port,
			Password: config.Password,
			DB:       config.DB,
		}
		c = append(c, conf)
	}
	return c
}

func newRedisCache(conf Config) Cache {
	var client *redis.Client
	var clusterClient *redis.ClusterClient
	clusterConfig := ParseRedisConfig(conf)
	if len(clusterConfig) == 1 {
		address := fmt.Sprintf("%s:%s", conf.Host, conf.Port)
		client = redis.NewClient(&redis.Options{
			Addr:     address,
			Password: conf.Password,
			DB:       conf.DB,
		})
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_, err := client.Ping(ctx).Result()
		if err != nil {
			fmt.Printf("[FATAL] Redis connection failed: %v\n", err)
			fmt.Printf("[FATAL] Redis address: %s:%s\n", conf.Host, conf.Port)
			log.Fatal("initRedis client.Ping err: ", err)
		}
	} else {
		var addressList []string
		for _, conf := range clusterConfig {
			address := fmt.Sprintf("%s:%s", conf.Host, conf.Port)
			addressList = append(addressList, address)
		}
		clusterClient = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:         addressList,
			Password:      conf.Password,
			RouteRandomly: true,
			MaxRetries:    5,
			DialTimeout:   3 * time.Second,
			ReadTimeout:   2 * time.Second,
			WriteTimeout:  2 * time.Second,
			ReadOnly:      true, // 允许从副本读取
		})
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_, errPing := clusterClient.Ping(ctx).Result()
		if errPing != nil {
			fmt.Printf("[FATAL] Redis cluster connection failed: %v\n", errPing)
			fmt.Printf("[FATAL] Redis cluster addresses: %v\n", addressList)
			log.Fatal("initClusterRedis client.Ping err: ", errPing)
		}
	}
	// 启用 OpenTelemetry 追踪
	enableRedisTracing(client, clusterClient)

	return &redisCache{
		client:        client,
		clusterClient: clusterClient,
	}
}

func (c *redisCache) Set(key string, value interface{}, expiration time.Duration) error {
	ctx := context.Background()
	if c.clusterClient != nil {
		return c.clusterClient.Set(ctx, key, value, expiration).Err()
	}
	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *redisCache) Get(key string) (interface{}, bool) {
	ctx := context.Background()
	if c.clusterClient != nil {
		val, err := c.clusterClient.Get(ctx, key).Result()
		if err != nil {
			return nil, false

		}
		return val, true
	}
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}
	return val, true
}

func (c *redisCache) GetBytes(key string) ([]byte, bool) {
	ctx := context.Background()
	if c.clusterClient != nil {
		val, err := c.clusterClient.Get(ctx, key).Bytes()
		if err != nil {
			return nil, false

		}
		return val, true
	}
	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}
	return val, true
}

func (c *redisCache) GetBatchBytes(keys []string) (map[string][]byte, []string) {
	ctx := context.Background()
	result := make(map[string][]byte)
	var missedKeys []string

	if c.clusterClient != nil {
		vals, err := c.clusterClient.MGet(ctx, keys...).Result()
		if err != nil {
			return result, keys
		}

		// 处理结果
		for i, key := range keys {
			if i < len(vals) && vals[i] != nil {
				// 将值转换为字节数组
				switch v := vals[i].(type) {
				case string:
					result[key] = []byte(v)
				case []byte:
					result[key] = v
				default:
					// 如果不是字符串或字节数组，则尝试JSON序列化
					bytes, err := json.Marshal(v)
					if err != nil {
						missedKeys = append(missedKeys, key)
						continue
					}
					result[key] = bytes
				}
			} else {
				missedKeys = append(missedKeys, key)
			}
		}
	} else {
		// 对于单机模式，我们可以使用MGet
		vals, err := c.client.MGet(ctx, keys...).Result()
		if err != nil {
			return result, keys
		}

		// 处理结果
		for i, key := range keys {
			if i < len(vals) && vals[i] != nil {
				// 将值转换为字节数组
				switch v := vals[i].(type) {
				case string:
					result[key] = []byte(v)
				case []byte:
					result[key] = v
				default:
					// 如果不是字符串或字节数组，则尝试JSON序列化
					bytes, err := json.Marshal(v)
					if err != nil {
						missedKeys = append(missedKeys, key)
						continue
					}
					result[key] = bytes
				}
			} else {
				missedKeys = append(missedKeys, key)
			}
		}
	}

	return result, missedKeys
}

func (c *redisCache) Del(key ...string) {
	if c.clusterClient != nil {
		c.clusterClient.Del(context.Background(), key...)
		return
	}
	c.client.Del(context.Background(), key...)
}

// ==================== WithContext 方法 - 支持链路追踪 ====================

func (c *redisCache) SetWithContext(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if c.clusterClient != nil {
		return c.clusterClient.Set(ctx, key, value, expiration).Err()
	}
	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *redisCache) GetWithContext(ctx context.Context, key string) (interface{}, bool) {
	if c.clusterClient != nil {
		val, err := c.clusterClient.Get(ctx, key).Result()
		if err != nil {
			return nil, false
		}
		return val, true
	}
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}
	return val, true
}

func (c *redisCache) GetBytesWithContext(ctx context.Context, key string) ([]byte, bool) {
	if c.clusterClient != nil {
		val, err := c.clusterClient.Get(ctx, key).Bytes()
		if err != nil {
			return nil, false
		}
		return val, true
	}
	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}
	return val, true
}

func (c *redisCache) GetBatchBytesWithContext(ctx context.Context, keys []string) (map[string][]byte, []string) {
	result := make(map[string][]byte)
	var missedKeys []string

	if c.clusterClient != nil {
		vals, err := c.clusterClient.MGet(ctx, keys...).Result()
		if err != nil {
			return result, keys
		}

		for i, key := range keys {
			if i < len(vals) && vals[i] != nil {
				switch v := vals[i].(type) {
				case string:
					result[key] = []byte(v)
				case []byte:
					result[key] = v
				default:
					bytes, err := json.Marshal(v)
					if err != nil {
						missedKeys = append(missedKeys, key)
						continue
					}
					result[key] = bytes
				}
			} else {
				missedKeys = append(missedKeys, key)
			}
		}
	} else {
		vals, err := c.client.MGet(ctx, keys...).Result()
		if err != nil {
			return result, keys
		}

		for i, key := range keys {
			if i < len(vals) && vals[i] != nil {
				switch v := vals[i].(type) {
				case string:
					result[key] = []byte(v)
				case []byte:
					result[key] = v
				default:
					bytes, err := json.Marshal(v)
					if err != nil {
						missedKeys = append(missedKeys, key)
						continue
					}
					result[key] = bytes
				}
			} else {
				missedKeys = append(missedKeys, key)
			}
		}
	}

	return result, missedKeys
}

func (c *redisCache) DelWithContext(ctx context.Context, keys ...string) {
	if c.clusterClient != nil {
		c.clusterClient.Del(ctx, keys...)
		return
	}
	c.client.Del(ctx, keys...)
}
