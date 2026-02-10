package queue

import (
	"encoding/json"
	"time"

	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/service"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

const (
	// 队列配置
	operationDurationQueueSize = 10000 // 队列容量
	operationDurationBatchSize = 100   // 批量写入阈值
	operationDurationFlushTime = 5     // 刷新间隔（秒）
)

// OperationDurationQueue 操作耗时记录队列
// 使用内存 channel 作为缓冲，消费者协程批量写入数据库
type OperationDurationQueue struct {
	ch         chan *req.OperationDurationRecord
	dbm        *database.DBManager // 存储 DBManager 而非 *gorm.DB，确保每次获取最新连接
	batchSize  int
	flushTime  time.Duration
	instanceId string
	stopCh     chan struct{}
}

// NewOperationDurationQueue 创建操作耗时记录队列
// 如果 dbm 为 nil 或无法获取 SaaS 主库连接，返回 nil
func NewOperationDurationQueue(dbm *database.DBManager) *OperationDurationQueue {
	if dbm == nil {
		logger.Logger.Error("OperationDurationQueue 创建失败: DBManager 为 nil")
		return nil
	}

	// 验证能够获取 SaaS 主库连接
	if db := dbm.GetDB(constant.DefaultDB); db == nil {
		logger.Logger.Error("OperationDurationQueue 创建失败: 无法获取 SaaS 主库连接")
		return nil
	}

	return &OperationDurationQueue{
		ch:         make(chan *req.OperationDurationRecord, operationDurationQueueSize),
		dbm:        dbm, // 存储 DBManager，每次写入时获取最新连接
		batchSize:  operationDurationBatchSize,
		flushTime:  time.Duration(operationDurationFlushTime) * time.Second,
		instanceId: utils.GetInstanceID(),
		stopCh:     make(chan struct{}),
	}
}

// Start 启动消费者协程
func (q *OperationDurationQueue) Start() {
	utils.Go(q.consume)
	logger.Logger.Info("OperationDurationQueue 消费者已启动",
		zap.String("instance_id", q.instanceId),
		zap.Int("queue_size", operationDurationQueueSize),
		zap.Int("batch_size", q.batchSize),
		zap.Duration("flush_time", q.flushTime))
}

// Stop 优雅关闭
func (q *OperationDurationQueue) Stop() {
	close(q.stopCh)
	logger.Logger.Info("OperationDurationQueue 消费者正在关闭")
}

// Push 非阻塞推送记录到队列
// 如果队列满，丢弃记录并输出告警日志
func (q *OperationDurationQueue) Push(record *req.OperationDurationRecord) {
	if record == nil {
		return
	}

	// 自动填充服务实例标识
	record.InstanceId = q.instanceId

	select {
	case q.ch <- record:
		// 推送成功
	default:
		// 队列满，丢弃并告警
		logger.Logger.Warn("OperationDurationQueue 队列已满，记录被丢弃",
			zap.Uint64("company_uuid", record.CompanyUuid),
			zap.String("action", record.Action),
			zap.Int("queue_len", len(q.ch)))
	}
}

// consume 消费者：批量读取并写入数据库
func (q *OperationDurationQueue) consume() {
	batch := make([]*req.OperationDurationRecord, 0, q.batchSize)
	ticker := time.NewTicker(q.flushTime)
	defer ticker.Stop()

	for {
		select {
		case <-q.stopCh:
			// 关闭前刷新剩余数据
			if len(batch) > 0 {
				q.flush(batch)
			}
			logger.Logger.Info("OperationDurationQueue 消费者已关闭")
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
func (q *OperationDurationQueue) flush(records []*req.OperationDurationRecord) {
	if len(records) == 0 {
		return
	}

	// 每次写入时从 DBManager 获取最新连接，避免使用已关闭的连接
	db := q.dbm.GetDB(constant.DefaultDB)
	if db == nil {
		logger.Logger.Error("OperationDurationQueue flush 失败: 无法获取 SaaS 主库连接",
			zap.Int("dropped_count", len(records)))
		return
	}

	models := make([]model.OrderOperationDuration, 0, len(records))
	now := time.Now().Unix()

	for _, r := range records {
		uuid, err := utils.GetID()
		if err != nil {
			logger.Logger.Error("生成 UUID 失败", zap.Error(err))
			continue
		}

		// 序列化时间节点为 JSON
		timeNodesJSON := ""
		if len(r.TimeNodes) > 0 {
			if jsonBytes, err := json.Marshal(r.TimeNodes); err == nil {
				timeNodesJSON = string(jsonBytes)
			}
		}

		models = append(models, model.OrderOperationDuration{
			Uuid:          uuid,
			CompanyUuid:   r.CompanyUuid,
			SaleBillUuid:  r.SaleBillUuid,
			SaleOrderUuid: r.SaleOrderUuid,
			Action:        r.Action,
			Source:        r.Source,
			StaffUuid:     r.StaffUuid,
			DeviceSn:      r.DeviceSn,
			InstanceId:    r.InstanceId,
			StartTime:     r.StartTime,
			EndTime:       r.EndTime,
			DurationMs:    int(r.DurationMs),
			RequestPath:   r.RequestPath,
			Status:        r.Status,
			ErrorMsg:      r.ErrorMsg,
			TimeNodes:     timeNodesJSON,
			CreateTime:    now,
			UpdateTime:    now,
			DeleteTime:    0,
		})
	}

	if len(models) == 0 {
		return
	}

	// 批量写入
	if err := db.CreateInBatches(&models, q.batchSize).Error; err != nil {
		logger.Logger.Error("批量写入操作耗时记录失败",
			zap.Error(err),
			zap.Int("count", len(models)))
	} else {
		logger.Logger.Debug("批量写入操作耗时记录成功",
			zap.Int("count", len(models)))
	}
}

// GetQueueLen 获取当前队列长度（用于监控）
func (q *OperationDurationQueue) GetQueueLen() int {
	return len(q.ch)
}

// GetInstanceID 获取服务实例标识
func (q *OperationDurationQueue) GetInstanceID() string {
	return q.instanceId
}

// InitOperationDurationQueue 初始化并启动操作耗时记录队列
func InitOperationDurationQueue() {
	dbm := database.GetDBManager(config.Database)
	operationDurationQueue := NewOperationDurationQueue(dbm)
	if operationDurationQueue == nil {
		logger.Logger.Warn("OperationDurationQueue 初始化失败，操作耗时监控功能已禁用")
		return
	}
	operationDurationQueue.Start()
	service.Queue.RegisterOperationDurationQueue(operationDurationQueue)

	// 设置 helper 包的推送函数，避免循环依赖
	helper.SetDurationRecordPusher(operationDurationQueue.Push)
}
