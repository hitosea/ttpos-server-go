package cache

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	Client *redis.Client
}

type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
}

var GlobalRedis *redisCache
var once sync.Once

func Init(config Config) {
	once.Do(func() {

		client := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
			Password: config.Password,
			DB:       config.DB,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_, err := client.Ping(ctx).Result()
		if err != nil {
			log.Fatal("initRedis client.Ping err: ", err)
		}

		GlobalRedis = &redisCache{
			Client: client,
		}
	})
}

func (c *redisCache) Set(key string, value interface{}, expiration time.Duration) error {
	ctx := context.Background()
	return c.Client.Set(ctx, key, value, expiration).Err()
}

func (c *redisCache) Get(key string) (interface{}, bool) {
	ctx := context.Background()
	val, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}
	return val, true
}

func (c *redisCache) Del(key ...string) {
	c.Client.Del(context.Background(), key...)
}
