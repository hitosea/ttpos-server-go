package queue

import "github.com/redis/go-redis/v9"

const TAKEOUT = "takeout"

func Init(redisCli *redis.Client) {

	InitTakeoutCancel(redisCli)

}
