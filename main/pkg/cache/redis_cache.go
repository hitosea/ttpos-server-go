package cache

import (
	"context"
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
			log.Fatal("initClusterRedis client.Ping err: ", errPing)
		}
	}
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

func (c *redisCache) Del(key ...string) {
	if c.clusterClient != nil {
		c.clusterClient.Del(context.Background(), key...)
		return
	}
	c.client.Del(context.Background(), key...)
}
