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
	// 三分钟锁. min(过期时间, 重试次数 * 重试间隔) 取最小值, 在该参数下过期时间较小(3分钟),故以过期时间为准
	mutex := d.rs.NewMutex(fmt.Sprintf("%d", uuid), redsync.WithExpiry(60*3*time.Second), redsync.WithTries(60*4), redsync.WithRetryDelay(1*time.Second))
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
func (d *RedSyncLock) TryLockUuid(uuid string) bool {
	return d.getUuidLockString(uuid).TryLock() == nil
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

// 获取字符串锁
func (d *RedSyncLock) getUuidLockString(uuid string) *redsync.Mutex {
	if cached, ok := d.uuidLock.Load(uuid); ok {
		return cached.(*redsync.Mutex)
	}
	// 设置锁过期时间为15分钟，确保长时间操作不会因锁过期而失效
	mutex := d.rs.NewMutex(uuid, redsync.WithExpiry(15*time.Minute))
	actual, _ := d.uuidLock.LoadOrStore(uuid, mutex)
	return actual.(*redsync.Mutex)
}

// LockUuidString 锁定uuid
func (d *RedSyncLock) LockUuidString(uuid string) {
	err := d.getUuidLockString(uuid).Lock()
	if err != nil {
		fmt.Println(err)
	}
}

// TryLockUuidString 非阻塞尝试获取uuid锁
func (d *RedSyncLock) TryLockUuidString(uuid string) bool {
	return d.getUuidLockString(uuid).TryLock() == nil
}

// UnlockUuidString 解锁uuid
func (d *RedSyncLock) UnlockUuidString(uuid string) {
	unlock, err := d.getUuidLockString(uuid).Unlock()
	if err != nil {
		fmt.Println(err)
	}
	if !unlock {
		fmt.Println(err)
	}
}

// ClearUuidLockString 在uuid对应的资源完成或删除后，清除uuid锁
func (d *RedSyncLock) ClearUuidLockString(uuid string) {
	unlock, err := d.getUuidLockString(uuid).Unlock()
	if err != nil {
		panic(err)
	}
	if !unlock {
		panic("unlock failed")
	}
}
