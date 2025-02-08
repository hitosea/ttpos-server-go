package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"websocket/config"
)

func NewRedisClient(conf config.RedisConf) *redis.Client {
	cli := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", conf.Host, conf.Port),
		Password: conf.Password,
		DB:       conf.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := cli.Ping(ctx).Result()
	if err != nil {
		log.Fatal("initRedis client.Ping err: ", err)
	}
	return cli
}
