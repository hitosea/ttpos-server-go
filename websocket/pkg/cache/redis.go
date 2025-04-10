package cache

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	Client        *redis.Client
	ClusterClient *redis.ClusterClient
}

type Config struct {
	Host     string
	Port     string
	Password string
	DB       int
}

var GlobalRedis *redisCache
var once sync.Once

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

func Init(config Config) {
	once.Do(func() {
		GlobalRedis = newRedisCache(config)
	})
}

func newRedisCache(conf Config) *redisCache {
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
			log.Fatal("initRedis client.Ping err: ", err)
		}
	} else {
		var addressList []string
		for _, conf := range clusterConfig {
			address := fmt.Sprintf("%s:%s", conf.Host, conf.Port)
			addressList = append(addressList, address)

		}
		opt := &redis.ClusterOptions{
			Addrs:    addressList,
			Password: conf.Password,
		}
		clusterClient = redis.NewClusterClient(opt)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_, errPing := clusterClient.Ping(ctx).Result()
		if errPing != nil {
			log.Fatal("initClusterRedis client.Ping err: ", errPing)
		}
	}
	return &redisCache{
		Client:        client,
		ClusterClient: clusterClient,
	}
}

func (c *redisCache) Set(key string, value interface{}, expiration time.Duration) error {
	ctx := context.Background()
	if c.ClusterClient != nil {
		return c.ClusterClient.Set(ctx, key, value, expiration).Err()
	}
	return c.Client.Set(ctx, key, value, expiration).Err()
}

func (c *redisCache) Get(key string) (interface{}, bool) {
	ctx := context.Background()
	if c.ClusterClient != nil {
		val, err := c.ClusterClient.Get(ctx, key).Result()
		if err != nil {
			return nil, false
		}
		return val, true
	}
	val, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}
	return val, true
}

// Subscribe 订阅消息
func (c *redisCache) Subscribe(ctx context.Context, key string) *redis.PubSub {
	if c.ClusterClient != nil {
		return c.ClusterClient.Subscribe(ctx, key)
	}
	return c.Client.Subscribe(ctx, key)
}

// Publish 发布消息
func (c *redisCache) Publish(key string, data string) error {
	ctx := context.Background()
	if c.ClusterClient != nil {
		err := c.ClusterClient.Publish(ctx, key, data).Err()
		if err != nil {
			return err
		}
		return nil
	}
	err := c.Client.Publish(ctx, key, data).Err()
	if err != nil {
		return err
	}
	return nil
}

func (c *redisCache) Del(key ...string) {
	if c.ClusterClient != nil {
		c.ClusterClient.Del(context.Background(), key...)
		return
	}
	c.Client.Del(context.Background(), key...)
}
