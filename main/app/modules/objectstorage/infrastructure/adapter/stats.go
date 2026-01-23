package adapter

import (
	"context"
	"encoding/json"
	"sort"
	"sync"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ==================== 缓存命中率统计 ====================

// KeyStats 单个key的统计信息
type KeyStats struct {
	Key        string    // key名称
	Hits       int64     // 命中次数
	Misses     int64     // 未命中次数
	HitRate    float64   // 命中率
	Total      int64     // 总请求数
	LastAccess time.Time // 最后访问时间
}

// CacheStats 缓存统计信息（不包含锁，避免复制时的问题）
type CacheStats struct {
	Hits       int64               // 命中次数
	Misses     int64               // 未命中次数
	HitRate    float64             // 命中率
	Total      int64               // 总请求数
	LastUpdate time.Time           // 最后更新时间
	KeyStats   map[string]KeyStats // 每个key的统计信息
}

// CacheStatsManager 缓存统计管理器（包含锁）
type CacheStatsManager struct {
	stats    CacheStats
	keyStats map[string]*KeyStats // key级别的统计
	mu       sync.RWMutex
}

// NewCacheStatsManager 创建缓存统计管理器
func NewCacheStatsManager() *CacheStatsManager {
	return &CacheStatsManager{
		stats: CacheStats{
			KeyStats: make(map[string]KeyStats),
		},
		keyStats: make(map[string]*KeyStats),
	}
}

// GetHitRate 获取命中率
func (m *CacheStatsManager) GetHitRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats.HitRate
}

// RecordHit 记录命中（带key信息）
func (m *CacheStatsManager) RecordHit(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stats.Hits++
	m.stats.Total++
	m.updateHitRate()
	m.stats.LastUpdate = time.Now()

	// 记录key级别的统计
	m.recordKeyHit(key)
}

// RecordMiss 记录未命中（带key信息）
func (m *CacheStatsManager) RecordMiss(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stats.Misses++
	m.stats.Total++
	m.updateHitRate()
	m.stats.LastUpdate = time.Now()

	// 记录key级别的统计
	m.recordKeyMiss(key)
}

// recordKeyHit 记录key命中
func (m *CacheStatsManager) recordKeyHit(key string) {
	if m.keyStats[key] == nil {
		m.keyStats[key] = &KeyStats{
			Key: key,
		}
	}
	ks := m.keyStats[key]
	ks.Hits++
	ks.Total++
	ks.LastAccess = time.Now()
	if ks.Total > 0 {
		ks.HitRate = float64(ks.Hits) / float64(ks.Total) * 100
	}
}

// recordKeyMiss 记录key未命中
func (m *CacheStatsManager) recordKeyMiss(key string) {
	if m.keyStats[key] == nil {
		m.keyStats[key] = &KeyStats{
			Key: key,
		}
	}
	ks := m.keyStats[key]
	ks.Misses++
	ks.Total++
	ks.LastAccess = time.Now()
	if ks.Total > 0 {
		ks.HitRate = float64(ks.Hits) / float64(ks.Total) * 100
	}
}

// updateHitRate 更新命中率
func (m *CacheStatsManager) updateHitRate() {
	if m.stats.Total > 0 {
		m.stats.HitRate = float64(m.stats.Hits) / float64(m.stats.Total) * 100
	}
}

// Reset 重置统计
func (m *CacheStatsManager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stats.Hits = 0
	m.stats.Misses = 0
	m.stats.Total = 0
	m.stats.HitRate = 0
	m.stats.LastUpdate = time.Now()
	m.stats.KeyStats = make(map[string]KeyStats)
	m.keyStats = make(map[string]*KeyStats)
}

// Snapshot 获取统计快照
func (m *CacheStatsManager) Snapshot() CacheStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 复制key级别的统计
	keyStatsCopy := make(map[string]KeyStats)
	for k, v := range m.keyStats {
		keyStatsCopy[k] = KeyStats{
			Key:        v.Key,
			Hits:       v.Hits,
			Misses:     v.Misses,
			HitRate:    v.HitRate,
			Total:      v.Total,
			LastAccess: v.LastAccess,
		}
	}

	return CacheStats{
		Hits:       m.stats.Hits,
		Misses:     m.stats.Misses,
		HitRate:    m.stats.HitRate,
		Total:      m.stats.Total,
		LastUpdate: m.stats.LastUpdate,
		KeyStats:   keyStatsCopy,
	}
}

// GetKeyStats 获取指定key的统计信息
func (m *CacheStatsManager) GetKeyStats(key string) (KeyStats, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if ks, exists := m.keyStats[key]; exists {
		return KeyStats{
			Key:        ks.Key,
			Hits:       ks.Hits,
			Misses:     ks.Misses,
			HitRate:    ks.HitRate,
			Total:      ks.Total,
			LastAccess: ks.LastAccess,
		}, true
	}
	return KeyStats{}, false
}

// GetTopKeys 获取Top N的key（按总访问次数排序）
func (m *CacheStatsManager) GetTopKeys(n int, sortBy string) []KeyStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]KeyStats, 0, len(m.keyStats))
	for _, ks := range m.keyStats {
		keys = append(keys, KeyStats{
			Key:        ks.Key,
			Hits:       ks.Hits,
			Misses:     ks.Misses,
			HitRate:    ks.HitRate,
			Total:      ks.Total,
			LastAccess: ks.LastAccess,
		})
	}

	// 排序
	switch sortBy {
	case "total":
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].Total > keys[j].Total
		})
	case "hits":
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].Hits > keys[j].Hits
		})
	case "hitrate":
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].HitRate > keys[j].HitRate
		})
	default:
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].Total > keys[j].Total
		})
	}

	// 返回Top N
	if n > len(keys) {
		n = len(keys)
	}
	return keys[:n]
}

// ==================== 观察者模式：缓存命中率监听 ====================

// CacheHitRateObserver 缓存命中率观察者接口
type CacheHitRateObserver interface {
	OnHitRateChanged(stats CacheStats)
	OnHitRateAlert(stats CacheStats, threshold float64)
	OnKeyStatsChanged(key string, keyStats KeyStats) // 新增：key级别统计变化通知
}

// CacheHitRateSubject 缓存命中率被观察者
type CacheHitRateSubject struct {
	observers []CacheHitRateObserver
	stats     *CacheStatsManager
	threshold float64 // 告警阈值（命中率低于此值时告警）
	mu        sync.RWMutex
}

// NewCacheHitRateSubject 创建缓存命中率被观察者
func NewCacheHitRateSubject(threshold float64) *CacheHitRateSubject {
	return &CacheHitRateSubject{
		observers: make([]CacheHitRateObserver, 0),
		stats:     NewCacheStatsManager(),
		threshold: threshold,
	}
}

// Attach 添加观察者
func (s *CacheHitRateSubject) Attach(observer CacheHitRateObserver) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

// Detach 移除观察者
func (s *CacheHitRateSubject) Detach(observer CacheHitRateObserver) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, obs := range s.observers {
		if obs == observer {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			break
		}
	}
}

// notifyHitRateChanged 通知命中率变化
func (s *CacheHitRateSubject) notifyHitRateChanged() {
	s.mu.RLock()
	observers := make([]CacheHitRateObserver, len(s.observers))
	copy(observers, s.observers)
	stats := s.stats.Snapshot()
	s.mu.RUnlock()

	for _, observer := range observers {
		observer.OnHitRateChanged(stats)
	}

	// 检查是否需要告警
	// 只有在总计 >= 1000 且命中率低于阈值时才告警（避免样本量太小时的误报）
	if stats.Total >= 1000 && stats.HitRate < s.threshold {
		for _, observer := range observers {
			observer.OnHitRateAlert(stats, s.threshold)
		}
	}
}

// notifyKeyStatsChanged 通知key级别统计变化
func (s *CacheHitRateSubject) notifyKeyStatsChanged(key string) {
	s.mu.RLock()
	observers := make([]CacheHitRateObserver, len(s.observers))
	copy(observers, s.observers)
	keyStats, exists := s.stats.GetKeyStats(key)
	s.mu.RUnlock()

	if exists {
		for _, observer := range observers {
			observer.OnKeyStatsChanged(key, keyStats)
		}
	}
}

// RecordHit 记录命中并通知观察者（带key信息）
func (s *CacheHitRateSubject) RecordHit(key string) {
	s.stats.RecordHit(key)
	s.notifyHitRateChanged()
	s.notifyKeyStatsChanged(key)
}

// RecordMiss 记录未命中并通知观察者（带key信息）
func (s *CacheHitRateSubject) RecordMiss(key string) {
	s.stats.RecordMiss(key)
	s.notifyHitRateChanged()
	s.notifyKeyStatsChanged(key)
}

// GetStats 获取统计信息
func (s *CacheHitRateSubject) GetStats() CacheStats {
	return s.stats.Snapshot()
}

// SetThreshold 设置告警阈值
func (s *CacheHitRateSubject) SetThreshold(threshold float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.threshold = threshold
}

// Reset 重置统计
func (s *CacheHitRateSubject) Reset() {
	s.stats.Reset()
}

// GetKeyStats 获取指定key的统计信息
func (s *CacheHitRateSubject) GetKeyStats(key string) (KeyStats, bool) {
	return s.stats.GetKeyStats(key)
}

// GetTopKeys 获取Top N的key
func (s *CacheHitRateSubject) GetTopKeys(n int, sortBy string) []KeyStats {
	return s.stats.GetTopKeys(n, sortBy)
}

// ==================== 具体观察者实现 ====================

// LogObserver 日志观察者
type LogObserver struct {
	name string
}

// NewLogObserver 创建日志观察者
func NewLogObserver(name string) *LogObserver {
	return &LogObserver{name: name}
}

// OnHitRateChanged 命中率变化通知
func (o *LogObserver) OnHitRateChanged(stats CacheStats) {
	// 仅在开发模式下输出
	if config.Server.Mode == gin.DebugMode {
		logger.Logger.Debug("缓存命中率变化",
			zap.String("observer", o.name),
			zap.Float64("hit_rate", stats.HitRate),
			zap.Int64("hits", stats.Hits),
			zap.Int64("misses", stats.Misses),
			zap.Int64("total", stats.Total),
		)
	}
}

// OnHitRateAlert 命中率告警通知
func (o *LogObserver) OnHitRateAlert(stats CacheStats, threshold float64) {
	logger.Logger.Warn("缓存命中率告警",
		zap.String("observer", o.name),
		zap.Float64("hit_rate", stats.HitRate),
		zap.Float64("threshold", threshold),
		zap.Int64("hits", stats.Hits),
		zap.Int64("misses", stats.Misses),
		zap.Int64("total", stats.Total),
	)
}

// OnKeyStatsChanged key级别统计变化通知
func (o *LogObserver) OnKeyStatsChanged(key string, keyStats KeyStats) {
	// 可选：记录key级别的统计变化（避免输出过多，这里注释掉）
	// logger.Logger.Debug("Key统计变化",
	// 	zap.String("observer", o.name),
	// 	zap.String("key", key),
	// 	zap.Float64("hit_rate", keyStats.HitRate),
	// 	zap.Int64("hits", keyStats.Hits),
	// 	zap.Int64("misses", keyStats.Misses),
	// 	zap.Int64("total", keyStats.Total),
	// )
}

// ==================== 快照保存器 ====================

// SnapshotReporter 快照保存器（定期保存缓存命中率快照到数据库）
type SnapshotReporter struct {
	monitor     *CacheHitRateSubject
	interval    time.Duration
	saveFunc    func(stats CacheStats, instanceID string, isRestart bool) error
	instanceID  string
	latestTotal int64 // 记录上一次的 total 值，用于检测重启
}

// NewSnapshotReporter 创建快照保存器
// dbGetter 是一个函数，用于获取 saas 库的数据库连接（index=0）
func NewSnapshotReporter(monitor *CacheHitRateSubject, interval time.Duration, dbGetter func(uint64) *gorm.DB, instanceID string) *SnapshotReporter {
	return &SnapshotReporter{
		monitor:     monitor,
		interval:    interval,
		saveFunc:    saveCacheHitRateSnapshot(dbGetter),
		instanceID:  instanceID,
		latestTotal: 0,
	}
}

// saveCacheHitRateSnapshot 保存缓存命中率快照的函数
// 供 SnapshotReporter 调用
func saveCacheHitRateSnapshot(dbGetter func(uint64) *gorm.DB) func(stats CacheStats, instanceID string, isRestart bool) error {
	return func(stats CacheStats, instanceID string, isRestart bool) error {
		// 将 KeyStats map 序列化为 JSON
		keyStatsJSON, err := json.Marshal(stats.KeyStats)
		if err != nil {
			logger.Logger.Error("序列化 KeyStats 失败", zap.Error(err))
			return err
		}

		// 构建 snapshot 记录
		snapshot := &model.CacheHitRateSnapshot{
			InstanceID:   instanceID,
			SnapshotTime: time.Now(),
			Hits:         stats.Hits,
			Misses:       stats.Misses,
			Total:        stats.Total,
			HitRate:      stats.HitRate,
			KeyCount:     len(stats.KeyStats),
			KeyStats:     string(keyStatsJSON),
			IsRestart:    boolToInt(isRestart),
			CreateTime:   time.Now().Unix(),
		}

		// 直接使用数据库连接保存（避免导入 database 造成循环依赖）
		db := dbGetter(0) // saas 库的 index 是 0
		if db == nil {
			return gorm.ErrRecordNotFound
		}
		err = db.Create(snapshot).Error
		if err == nil {
			// 保存成功后，启动协程执行清理任务（带防抖）
			utils.Go(func() {
				cleanupSnapshots(dbGetter)
			})
		}
		return err
	}
}

// boolToInt 将 bool 转换为 int（1 或 0）
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// cleanupSnapshots 清理快照表中的冗余数据（带防抖机制）
// 清理逻辑：
// 仅处理上一周的数据（7天前到14天前），对这些数据进行清洗：
// 对于同一个 instance_id，且 is_restart = 0 的连续记录，只保留最新的一条，删除旧的
func cleanupSnapshots(dbGetter func(uint64) *gorm.DB) {
	cleanupMutex.Lock()
	// 检查是否需要执行清理（防抖：每小时最多执行一次）
	now := time.Now()
	if !lastCleanupTime.IsZero() && now.Sub(lastCleanupTime) < cleanupInterval {
		cleanupMutex.Unlock()
		return
	}
	lastCleanupTime = now
	cleanupMutex.Unlock()

	// 获取数据库连接
	db := dbGetter(0) // saas 库的 index 是 0
	if db == nil {
		logger.Logger.Warn("清理快照失败：无法获取数据库连接")
		return
	}

	// 计算上一周的时间范围（7天前到14天前）
	oneWeekAgo := now.AddDate(0, 0, -7).Unix()   // 7天前
	twoWeeksAgo := now.AddDate(0, 0, -14).Unix() // 14天前
	// oneWeekAgo := now.Unix()                    // 今天
	// twoWeeksAgo := now.AddDate(0, 0, -1).Unix() // 昨天

	// 查询上一周的数据中所有不同的 instance_id
	var instanceIDs []string
	if err := db.Model(&model.CacheHitRateSnapshot{}).
		Where("create_time >= ? AND create_time < ?", twoWeeksAgo, oneWeekAgo).
		Distinct("instance_id").
		Pluck("instance_id", &instanceIDs).Error; err != nil {
		logger.Logger.Warn("查询上一周实例ID列表失败", zap.Error(err))
		return
	}

	if len(instanceIDs) == 0 {
		logger.Logger.Debug("清理快照：上一周没有数据需要清理")
		return
	}

	totalDeleted := int64(0)
	for _, instanceID := range instanceIDs {
		// 获取该实例在上一周的所有快照（按时间排序）
		var snapshots []model.CacheHitRateSnapshot
		if err := db.Where("instance_id = ? AND create_time >= ? AND create_time < ?", instanceID, twoWeeksAgo, oneWeekAgo).
			Order("snapshot_time ASC").
			Find(&snapshots).Error; err != nil {
			logger.Logger.Warn("查询实例快照失败",
				zap.String("instance_id", instanceID),
				zap.Error(err),
			)
			continue
		}

		if len(snapshots) == 0 {
			continue
		}

		// 找出连续的 is_restart = 0 的记录组
		var toDelete []uint64
		var consecutiveGroup []*model.CacheHitRateSnapshot

		// 处理连续组的函数
		processConsecutiveGroup := func() {
			if len(consecutiveGroup) > 1 {
				// 保留最后一条（最新的），删除前面的
				for j := 0; j < len(consecutiveGroup)-1; j++ {
					toDelete = append(toDelete, consecutiveGroup[j].ID)
				}
			}
			consecutiveGroup = make([]*model.CacheHitRateSnapshot, 0)

		}

		for i := range snapshots {
			snapshot := &snapshots[i]
			if snapshot.IsRestart == 1 {
				// 遇到重启标记，结束当前连续组
				processConsecutiveGroup()
				continue
			}

			// is_restart = 0，加入连续组
			consecutiveGroup = append(consecutiveGroup, snapshot)

			// 如果是最后一条记录，也需要处理连续组
			if i == len(snapshots)-1 {
				processConsecutiveGroup()
			}
		}

		if len(toDelete) > 0 {
			if err := db.Where("id IN ?", toDelete).
				Delete(&model.CacheHitRateSnapshot{}).Error; err != nil {
				logger.Logger.Warn("删除连续记录失败",
					zap.String("instance_id", instanceID),
					zap.Error(err),
				)
				continue
			}
			totalDeleted += int64(len(toDelete))
			logger.Logger.Debug("清理快照：已删除实例连续记录",
				zap.String("instance_id", instanceID),
				zap.Int("count", len(toDelete)),
			)
		}
	}

	if totalDeleted > 0 {
		logger.Logger.Info("清理快照：已删除上一周的连续记录",
			zap.Int64("count", totalDeleted),
			zap.Int64("time_range_start", twoWeeksAgo),
			zap.Int64("time_range_end", oneWeekAgo),
		)
	} else {
		logger.Logger.Debug("清理快照：上一周没有需要删除的连续记录")
	}
}

// Start 启动快照保存器
func (r *SnapshotReporter) Start(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 动态检查缓存配置：只有在开启缓存且只有一个白名单商家时才保存快照
			config := GetObjectStorageCacheConfig()

			// 检查是否开启缓存
			if !config.Enabled {
				logger.Logger.Debug("缓存未开启，跳过保存缓存命中率快照",
					zap.String("instance_id", r.instanceID),
				)
				continue
			}

			stats := r.monitor.GetStats()

			// 检测是否是重启后的第一条快照
			isRestart := false
			if r.latestTotal > 0 && stats.Total < r.latestTotal {
				// 如果当前 total 小于上一次的 total，说明实例重启了
				isRestart = true
			} else if r.latestTotal == 0 {
				// 第一次保存，认为是重启后的第一条
				isRestart = true
			}

			// 保存快照
			if err := r.saveFunc(stats, r.instanceID, isRestart); err != nil {
				logger.Logger.Error("保存缓存命中率快照失败",
					zap.String("instance_id", r.instanceID),
					zap.Error(err),
				)
			} else {
				logger.Logger.Info("缓存命中率快照保存成功",
					zap.String("instance_id", r.instanceID),
					zap.Int64("hits", stats.Hits),
					zap.Int64("misses", stats.Misses),
					zap.Int64("total", stats.Total),
					zap.Float64("hit_rate", stats.HitRate),
					zap.Bool("is_restart", isRestart),
				)
			}

			// 更新 latestTotal
			r.latestTotal = stats.Total
		}
	}
}

// ==================== 全局统计管理器 ====================

var (
	globalStatsManager   *CacheHitRateSubject
	globalStatsOnce      sync.Once
	snapshotReporter     *SnapshotReporter
	snapshotReporterOnce sync.Once
	// 清理任务防抖：记录最后一次清理时间，避免频繁执行
	lastCleanupTime time.Time
	cleanupMutex    sync.Mutex
	cleanupInterval = 1 * time.Hour // 清理任务执行间隔：1小时
)

// GetGlobalStatsManager 获取全局统计管理器（单例模式）
func GetGlobalStatsManager() *CacheHitRateSubject {
	globalStatsOnce.Do(func() {
		// 默认告警阈值：70%
		globalStatsManager = NewCacheHitRateSubject(70.0)

		// 默认注册日志观察者
		logObserver := NewLogObserver("对象存储缓存统计")
		globalStatsManager.Attach(logObserver)
	})
	return globalStatsManager
}

// InitSnapshotReporter 初始化并启动快照保存器
// 快照保存器会动态检查缓存配置，只有在开启缓存且只有一个白名单商家时才保存快照
// 参数：
//   - dbGetter: 数据库获取函数（用于获取 saas 库的连接，index=0）
//   - instanceID: 实例ID（用于标识不同的应用实例）
//   - interval: 快照保存间隔（默认30分钟）
func InitSnapshotReporter(dbGetter func(uint64) *gorm.DB, instanceID string, interval time.Duration) {
	snapshotReporterOnce.Do(func() {
		monitor := GetGlobalStatsManager()
		snapshotReporter = NewSnapshotReporter(
			monitor,
			interval,
			dbGetter,
			instanceID,
		)

		// 在后台 goroutine 中安全启动快照保存器（使用 utils.Go 自动捕获 panic）
		ctx := context.Background()
		utils.Go(func() {
			snapshotReporter.Start(ctx)
		})

		logger.Logger.Info("缓存命中率快照保存器已启动",
			zap.String("instance_id", instanceID),
			zap.Duration("interval", interval),
		)
	})
}
