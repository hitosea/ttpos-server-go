# 并发锁实现设计文档

### 1. 简介

本文档旨在详细描述 TTPOS-Server-Go 项目中并发锁的实现设计。在分布式系统中，并发控制是确保数据一致性和系统稳定性的关键。本项目通过抽象的 `Lock` 接口，结合 `redsync` 库实现了高性能、可靠的分布式锁，同时也提供了本地锁作为备选。

### 2. 设计目标

*   **数据一致性**: 确保在多个并发请求下，对共享资源的访问是互斥的，防止数据损坏或不一致。
*   **高可用性**: 分布式锁应具备一定的容错能力，避免单点故障导致系统不可用。
*   **性能**: 锁机制应尽量减少对系统性能的影响。
*   **易用性**: 提供简洁明了的 API 接口，方便业务层调用。
*   **可扩展性**: 能够根据业务需求，灵活切换或扩展不同的锁实现。

### 3. 架构概览

并发锁模块的核心设计思想是提供一个统一的 `Lock` 接口，并在初始化时根据配置选择具体的实现。目前支持两种锁实现：

*   **分布式锁**: 基于 Redis 和 `redsync` 库实现。适用于多实例部署的分布式系统。
*   **本地锁**: 基于 Go 标准库 `sync.Mutex` 实现。适用于单实例部署或无需跨进程同步的场景。

`main/pkg/lock` 目录结构：

```
main/pkg/lock/
├── local.go         # 本地锁实现
├── lock_redsync.go  # 基于 redsync 的分布式锁实现
├── system_lock.go   # 统一的 Lock 接口及系统锁入口
└── system_lock_test.go # 测试文件
```

### 4. 核心接口定义 (`system_lock.go`)

`Lock` 接口定义了并发锁的基本操作，支持 `uint64` 和 `string` 两种类型的 UUID 作为锁的标识符。

```go
// ... existing code ...
type Lock interface {
	LockUuid(uuid uint64)
	UnlockUuid(uuid uint64)
	ClearUuidLock(uuid uint64)
	// 字符串锁
	LockUuidString(uuid string)
	TryLockUuidString(uuid string) bool // 非阻塞尝试获取锁，返回是否成功获取
	UnlockUuidString(uuid string)
	ClearUuidLockString(uuid string)
}

// ... existing code ...
```

**方法说明**:

*   `LockUuid(uuid uint64)`: 获取 `uint64` 类型的 UUID 锁。如果锁已被占用，则阻塞直到获取成功。
*   `UnlockUuid(uuid uint64)`: 释放 `uint64` 类型的 UUID 锁。
*   `ClearUuidLock(uuid uint64)`: 在资源完成或删除后，清除 `uint64` 类型的 UUID 锁，本质上也是释放锁。
*   `LockUuidString(uuid string)`: 获取 `string` 类型的 UUID 锁。阻塞。
*   `TryLockUuidString(uuid string) bool`: 非阻塞尝试获取 `string` 类型的 UUID 锁，返回是否成功获取。
*   `UnlockUuidString(uuid string)`: 释放 `string` 类型的 UUID 锁。
*   `ClearUuidLockString(uuid string)`: 清除 `string` 类型的 UUID 锁。

### 5. 系统锁初始化 (`system_lock.go`)

`NewSystemLock()` 函数是获取并发锁实例的统一入口。它使用 `sync.Once` 确保锁实例只被初始化一次。默认情况下，系统使用 `NewRedSyncLock(NewRedSync())` 初始化分布式锁。

```go
// ... existing code ...
var systemLock Lock
var once sync.Once

// NewSystemLock 创建系统锁
func NewSystemLock() Lock {
	once.Do(func() {
		//systemLock = InitLocalLock() // 本地锁，仅适用于单体应用
		systemLock = NewRedSyncLock(NewRedSync()) // 分布式锁
	})
	return systemLock
}
// ... existing code ...
```

通过修改 `NewSystemLock()` 中的注释，可以方便地切换到本地锁实现，这提供了良好的可配置性。

### 6. 分布式锁实现 (`lock_redsync.go`)

#### 6.1 `RedSyncLock` 结构体

`RedSyncLock` 结构体封装了 `redsync.Redsync` 实例和 `sync.Map` 用于缓存已创建的 `redsync.Mutex`。

```go
// ... existing code ...
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
// ... existing code ...
```

#### 6.2 Redis 连接管理

`NewRedSync()` 函数负责初始化 `redsync.Redsync` 实例。它根据配置 (`cache.Config`) 判断是连接单 Redis 实例还是 Redis Cluster。

*   **单 Redis 实例**: 通过 `newRedisCache()` 创建 `goredislib.Client`。
*   **Redis Cluster**: 通过 `newRedisCluster()` 创建 `goredislib.ClusterClient`。

无论哪种情况，最终都会通过 `goredis.NewPool()` 将 Redis 客户端适配到 `redsync` 的 `Pool` 接口。

```go
// ... existing code ...
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
// ... existing code ...
```

#### 6.3 锁的获取与释放

`RedSyncLock` 的核心方法 `getUuidLock()` 和 `getUuidLockString()` 负责创建或获取 `redsync.Mutex` 实例。

*   **`redsync.WithExpiry(60*3*time.Second)`**: 设置锁的过期时间为 3 分钟。这是防止死锁的关键机制。
*   **`redsync.WithTries(60*4)`**: 在获取锁失败时，最多重试 240 次 (60 * 4)，每次重试间隔 1 秒。
*   **`redsync.WithRetryDelay(1*time.Second)`**: 重试间隔为 1 秒。

`LockUuid()` 和 `LockUuidString()` 调用 `redsync.Mutex.Lock()` 获取锁，如果失败会打印错误 (日志被注释掉)。
`UnlockUuid()` 和 `UnlockUuidString()` 调用 `redsync.Mutex.Unlock()` 释放锁。
`ClearUuidLock()` 和 `ClearUuidLockString()` 本质上也是释放锁，但如果解锁失败会 `panic`，这表明在资源清理时对锁的释放有更强的保证要求。

```go
// ... existing code ...
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
// ... existing code ...
```

#### 6.4 关键考量

*   **Redlock 算法**: `redsync` 库实现了 Redlock 算法，确保在 Redis 分布式环境下的锁的正确性和可靠性。
*   **锁续期**: `redsync` 默认支持锁续期机制，防止因业务逻辑执行时间过长导致锁提前过期。
*   **缓存 `redsync.Mutex`**: 使用 `sync.Map` 缓存 `redsync.Mutex` 实例，避免重复创建，提高性能。
*   **错误处理**: 锁获取和释放失败时，目前是打印错误或 `panic`。在生产环境中，应根据业务场景选择更健壮的错误处理机制，例如重试、告警等。

### 7. 本地锁实现 (`local.go`)

#### 7.1 `LocalLock` 结构体

`LocalLock` 结构体使用 `sync.Map` 存储不同 UUID 对应的 `sync.Mutex` 实例。

```go
// ... existing code ...
type LocalLock struct {
	uuidLock sync.Map
}

func InitLocalLock() *LocalLock {
	return &LocalLock{
		uuidLock: sync.Map{},
	}
}
// ... existing code ...
```

#### 7.2 锁的获取与释放

`LocalLock` 实现了 `Lock` 接口，通过 `sync.Mutex` 的 `Lock()` 和 `Unlock()` 方法实现互斥访问。`ClearUuidLock()` 和 `ClearUuidLockString()` 通过 `sync.Map.Delete()` 移除对应的 `sync.Mutex` 实例。

```go
// ... existing code ...
// 获取uuid锁
func (d *LocalLock) getUuidLock(uuid uint64) *sync.Mutex {
	actual, _ := d.uuidLock.LoadOrStore(uuid, &sync.Mutex{})
	return actual.(*sync.Mutex)
}

// LockUuid 锁定uuid
func (d *LocalLock) LockUuid(uuid uint64) {
	d.getUuidLock(uuid).Lock()
}
// ... existing code ...
```

#### 7.3 关键考量

*   **单体应用**: 本地锁仅适用于单体应用，无法在多实例部署时实现跨进程的并发控制。
*   **性能**: 相较于分布式锁，本地锁的性能开销更小。
*   **适用场景**: 适用于对性能要求较高且无需分布式同步的局部资源锁定。

### 8. 使用场景

并发锁在 TTPOS-Server-Go 项目中主要应用于以下场景：

*   **订单处理**: 确保订单状态变更、商品库存扣减、支付流程等操作的原子性。
*   **桌台管理**: 防止多个收银员同时操作同一桌台导致数据混乱 (例如合并桌台、切换桌台、修改订单商品)。
*   **会员余额/积分**: 确保会员余额或积分变动的准确性。
*   **交班操作**: 保证交班过程中现金操作和报备信息的准确记录。
*   **库存管理**: 保证商品库存更新的原子性，避免超卖。
*   **报表统计**: 确保在生成实时报表数据时，数据源的一致性。

### 9. 总结

本项目通过 `pkg/lock` 模块提供了灵活、可配置的并发锁机制。默认采用基于 `redsync` 库的分布式锁，以满足分布式部署下高并发场景的数据一致性需求。同时，本地锁的实现也为特定场景或单体应用提供了备选方案。通过统一的 `Lock` 接口，业务层可以无感知地切换底层锁实现，极大地提高了代码的灵活性和可维护性。在实际应用中，需要根据具体业务场景选择合适的锁粒度和超时策略，并结合日志和监控，确保锁机制的正确运行。
