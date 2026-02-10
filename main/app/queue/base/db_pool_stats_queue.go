package queue

import (
	"time"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

const (
	// 队列配置
	dbPoolStatsQueueSize  = 1000             // 队列容量
	dbPoolStatsBatchSize  = 50               // 批量写入阈值
	dbPoolStatsFlushTime  = 10               // 刷新间隔（秒）
	dbPoolStatsSampleTime = 30 * time.Minute // 采样间隔
)

// DBPoolStatsQueue 数据库连接池统计记录队列
// 使用内存 channel 作为缓冲，消费者协程批量写入数据库
type DBPoolStatsQueue struct {
	ch         chan *req.DBPoolStatsRecord
	dbm        *database.DBManager // 存储 DBManager 而非 *gorm.DB，确保每次获取最新连接
	batchSize  int
	flushTime  time.Duration
	sampleTime time.Duration
	instanceId string
	stopCh     chan struct{}
}

// NewDBPoolStatsQueue 创建数据库连接池统计记录队列
// 如果 dbm 为 nil 或无法获取 SaaS 主库连接，返回 nil
func NewDBPoolStatsQueue(dbm *database.DBManager) *DBPoolStatsQueue {
	if dbm == nil {
		logger.Logger.Error("DBPoolStatsQueue 创建失败: DBManager 为 nil")
		return nil
	}

	// 验证能够获取 SaaS 主库连接
	if db := dbm.GetDB(constant.DefaultDB); db == nil {
		logger.Logger.Error("DBPoolStatsQueue 创建失败: 无法获取 SaaS 主库连接")
		return nil
	}

	return &DBPoolStatsQueue{
		ch:         make(chan *req.DBPoolStatsRecord, dbPoolStatsQueueSize),
		dbm:        dbm, // 存储 DBManager，每次写入时获取最新连接
		batchSize:  dbPoolStatsBatchSize,
		flushTime:  time.Duration(dbPoolStatsFlushTime) * time.Second,
		sampleTime: dbPoolStatsSampleTime,
		instanceId: utils.GetInstanceID(),
		stopCh:     make(chan struct{}),
	}
}

// Start 启动消费者协程和采集协程
func (q *DBPoolStatsQueue) Start() {
	utils.Go(q.consume)
	utils.Go(q.collect)
	logger.Logger.Info("DBPoolStatsQueue 已启动",
		zap.String("instance_id", q.instanceId),
		zap.Duration("sample_interval", q.sampleTime),
		zap.Int("queue_size", dbPoolStatsQueueSize),
		zap.Int("batch_size", q.batchSize))
}

// Stop 优雅关闭
func (q *DBPoolStatsQueue) Stop() {
	close(q.stopCh)
	logger.Logger.Info("DBPoolStatsQueue 正在关闭")
}

// collect 定时采集数据库连接池统计
func (q *DBPoolStatsQueue) collect() {
	ticker := time.NewTicker(q.sampleTime)
	defer ticker.Stop()

	// 启动时立即采集一次
	q.sample()

	for {
		select {
		case <-q.stopCh:
			logger.Logger.Info("DBPoolStatsQueue 采集器已关闭")
			return
		case <-ticker.C:
			q.sample()
		}
	}
}

// sample 采集当前所有数据库连接池统计
func (q *DBPoolStatsQueue) sample() {
	stats := database.GetAllStats()
	if len(stats) == 0 {
		return
	}

	sampleTime := time.Now().UnixMilli()

	for _, stat := range stats {
		record := &req.DBPoolStatsRecord{
			InstanceId:        q.instanceId,
			DBName:            stat.DBName,
			MaxOpenConns:      stat.MaxOpenConns,
			OpenConns:         stat.OpenConns,
			InUse:             stat.InUse,
			Idle:              stat.Idle,
			WaitCount:         stat.WaitCount,
			WaitDurationMs:    stat.WaitDurationMs,
			MaxIdleClosed:     stat.MaxIdleClosed,
			MaxIdleTimeClosed: stat.MaxIdleTimeClosed,
			MaxLifetimeClosed: stat.MaxLifetimeClosed,
			SampleTime:        sampleTime,
		}
		q.Push(record)
	}
}

// Push 非阻塞推送记录到队列
func (q *DBPoolStatsQueue) Push(record *req.DBPoolStatsRecord) {
	select {
	case q.ch <- record:
		// 推送成功
	default:
		// 队列满，丢弃并告警
		logger.Logger.Warn("DBPoolStatsQueue 队列已满，记录被丢弃",
			zap.String("db_name", record.DBName),
			zap.Int("queue_len", len(q.ch)))
	}
}

// consume 消费者：批量读取并写入数据库
func (q *DBPoolStatsQueue) consume() {
	batch := make([]*req.DBPoolStatsRecord, 0, q.batchSize)
	ticker := time.NewTicker(q.flushTime)
	defer ticker.Stop()

	for {
		select {
		case <-q.stopCh:
			// 关闭前刷新剩余数据
			if len(batch) > 0 {
				q.flush(batch)
			}
			logger.Logger.Info("DBPoolStatsQueue 消费者已关闭")
			return

		case record := <-q.ch:
			batch = append(batch, record)
			if len(batch) >= q.batchSize {
				q.flush(batch)
				batch = batch[:0] // 重置切片但保留容量
			}

		case <-ticker.C:
			if len(batch) > 0 {
				q.flush(batch)
				batch = batch[:0]
			}
		}
	}
}

// flush 批量写入数据库
func (q *DBPoolStatsQueue) flush(records []*req.DBPoolStatsRecord) {
	if len(records) == 0 {
		return
	}

	// 每次写入时从 DBManager 获取最新连接，避免使用已关闭的连接
	db := q.dbm.GetDB(constant.DefaultDB)
	if db == nil {
		logger.Logger.Error("DBPoolStatsQueue flush 失败: 无法获取 SaaS 主库连接",
			zap.Int("dropped_count", len(records)))
		return
	}

	models := make([]model.DBPoolStats, 0, len(records))
	now := time.Now().Unix()

	for _, r := range records {
		uuid, err := utils.GetID()
		if err != nil {
			logger.Logger.Error("生成 UUID 失败", zap.Error(err))
			continue
		}

		models = append(models, model.DBPoolStats{
			Uuid:              uuid,
			InstanceId:        r.InstanceId,
			DBName:            r.DBName,
			MaxOpenConns:      r.MaxOpenConns,
			OpenConns:         r.OpenConns,
			InUse:             r.InUse,
			Idle:              r.Idle,
			WaitCount:         r.WaitCount,
			WaitDurationMs:    r.WaitDurationMs,
			MaxIdleClosed:     r.MaxIdleClosed,
			MaxIdleTimeClosed: r.MaxIdleTimeClosed,
			MaxLifetimeClosed: r.MaxLifetimeClosed,
			SampleTime:        r.SampleTime,
			CreateTime:        now,
			UpdateTime:        now,
			DeleteTime:        0,
		})
	}

	if len(models) == 0 {
		return
	}

	// 批量写入
	if err := db.CreateInBatches(&models, q.batchSize).Error; err != nil {
		logger.Logger.Error("批量写入数据库连接池统计记录失败",
			zap.Error(err),
			zap.Int("count", len(models)))
	} else {
		logger.Logger.Debug("批量写入数据库连接池统计记录成功",
			zap.Int("count", len(models)))
	}
}

// GetQueueLen 获取当前队列长度（用于监控）
func (q *DBPoolStatsQueue) GetQueueLen() int {
	return len(q.ch)
}

// InitDBPoolStatsQueue 初始化并启动数据库连接池统计记录队列
func InitDBPoolStatsQueue() {
	dbm := database.GetDBManager(config.Database)
	dbPoolStatsQueue := NewDBPoolStatsQueue(dbm)
	if dbPoolStatsQueue == nil {
		logger.Logger.Warn("DBPoolStatsQueue 初始化失败，连接池监控功能已禁用")
		return
	}
	dbPoolStatsQueue.Start()
}
