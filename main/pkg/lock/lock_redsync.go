package lock

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
	"ttpos-server-go/pkg/cache"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	goredislib "github.com/redis/go-redis/v9"
)

type Config struct {
	conf *cache.Config
}

var config *Config

func InitRedisLock(conf cache.Config) {
	config = &Config{
		conf: &conf,
	}
}

func newRedisCache(conf cache.Config) *goredislib.Client {
	client := goredislib.NewClient(&goredislib.Options{
		Addr:     fmt.Sprintf("%s:%s", conf.Host, conf.Port),
		Password: conf.Password,
		DB:       conf.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatal("initRedis client.Ping err: ", err)
	}

	return client
}

func newRedisCluster(conf cache.Config) *goredislib.ClusterClient {
	clusterConfig := cache.ParseRedisConfig(conf)
	var addressList []string
	for _, conf := range clusterConfig {
		address := fmt.Sprintf("%s:%s", conf.Host, conf.Port)
		addressList = append(addressList, address)
	}
	opt := &goredislib.ClusterOptions{
		Addrs:    addressList,
		Password: clusterConfig[0].Password,
	}
	clusterClient := goredislib.NewClusterClient(opt)
	return clusterClient
}

func NewRedSync() *redsync.Redsync {
	var client goredislib.UniversalClient
	if strings.Contains(config.conf.Host, ",") {
		client = newRedisCluster(*config.conf)
	} else {
		client = newRedisCache(*config.conf)
	}
	pool := goredis.NewPool(client)
	rs := redsync.New(pool)
	return rs
}

type RedSyncLock struct {
	uuidLock sync.Map
	rs       *redsync.Redsync
}

// NewRedSyncLock 创建系统锁
func NewRedSyncLock(rs *redsync.Redsync) *RedSyncLock {
	return &RedSyncLock{
		uuidLock: sync.Map{},
		rs:       rs,
	}
}

// 获取uuid锁
func (d *RedSyncLock) getUuidLock(uuid uint64) *redsync.Mutex {
	mutex := d.rs.NewMutex(fmt.Sprintf("%d", uuid))
	actual, _ := d.uuidLock.LoadOrStore(uuid, mutex)
	return actual.(*redsync.Mutex)
}

// LockUuid 锁定uuid
func (d *RedSyncLock) LockUuid(uuid uint64) {
	err := d.getUuidLock(uuid).Lock()
	if err != nil {
		//logger.Logger.Warn("获取分布式并发锁失败", zap.Uint64("uuid", uuid), zap.Error(err))
		fmt.Println(err)
	}
}

// TryLockUuid 非阻塞尝试获取uuid锁
func (d *RedSyncLock) TryLockUuid(uuid uint64) bool {
	err := d.getUuidLock(uuid).Lock()
	return err == nil
}

// UnlockUuid 解锁uuid
func (d *RedSyncLock) UnlockUuid(uuid uint64) {
	unlock, err := d.getUuidLock(uuid).Unlock()
	if err != nil {
		//logger.Logger.Warn("分布式并发锁释放失败", zap.Uint64("uuid", uuid), zap.Error(err))
		//fmt.Println(err)

	}
	if !unlock {
		//logger.Logger.Warn("分布式并发锁解锁失败", zap.Uint64("uuid", uuid), zap.Error(err))
		//fmt.Println(err)

	}
}

// ClearUuidLock 在uuid对应的资源完成或删除后，清除uuid锁
func (d *RedSyncLock) ClearUuidLock(uuid uint64) {
	unlock, err := d.getUuidLock(uuid).Unlock()
	if err != nil {
		panic(err)
	}
	if !unlock {
		panic("unlock failed")
	}
}
