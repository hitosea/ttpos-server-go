// Package http 实现 HTTP 控制器
package http

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/google/uuid"

	"ttpos-bmp/app/ttpos-websocket/internal/consts"
	"ttpos-bmp/app/ttpos-websocket/internal/model/dto"
	"ttpos-bmp/app/ttpos-websocket/internal/service"
)

// 全局限流器，控制同时运行的goroutine数量
var (
	// 最大并发处理的消息数
	maxConcurrentMessages = 500
	// 信号量通道，用于限制并发
	messageSemaphore = make(chan struct{}, maxConcurrentMessages)
	// 基于MessageKey的细粒度锁映射
	httpMessageKeyMutexes = sync.Map{} // map[string]*sync.Mutex
	// 保护messageKeyMutexes的锁
	httpMutexMapLock sync.Mutex
)

// 消息任务结构体
type messageTask struct {
	Params     *dto.PushMessageInput
	UUID       string
	MessageKey string
	CountKey   string
	Count      int64
}

// 初始化函数，在包初始化时调用
func init() {
	// 初始化信号量通道
	for i := 0; i < maxConcurrentMessages; i++ {
		messageSemaphore <- struct{}{}
	}
}

// getHttpMessageKeyMutex 获取指定MessageKey的专用锁
func getHttpMessageKeyMutex(messageKey string) *sync.Mutex {
	// 先尝试获取已存在的锁
	if mutex, exists := httpMessageKeyMutexes.Load(messageKey); exists {
		return mutex.(*sync.Mutex)
	}

	// 如果不存在，需要创建新锁
	httpMutexMapLock.Lock()
	defer httpMutexMapLock.Unlock()

	// 双重检查，防止并发创建
	if mutex, exists := httpMessageKeyMutexes.Load(messageKey); exists {
		return mutex.(*sync.Mutex)
	}

	// 创建新锁并存储
	newMutex := &sync.Mutex{}
	httpMessageKeyMutexes.Store(messageKey, newMutex)

	return newMutex
}

// PushClient HTTP推送接口
// 兼容旧的HTTP API，方便过渡
func PushMessage(r *ghttp.Request) {
	ctx := r.Context()

	// 解析请求参数
	var params struct {
		CompanyUuid  uint64      `json:"company_uuid"`
		SourceClient string      `json:"source_client"`
		DeviceId     string      `json:"device_id"`
		NotDeviceId  string      `json:"not_device_id"`
		StaffUuid    uint64      `json:"staff_uuid"`
		NotStaffUuid uint64      `json:"not_staff_uuid"`
		MessageType  string      `json:"message_type"`
		MessageKey   string      `json:"message_key"`
		Data         interface{} `json:"data"`
	}

	if err := r.Parse(&params); err != nil {
		g.Log().Error(ctx, "解析请求参数失败", err)
		r.Response.WriteJson(g.Map{
			"code":    consts.CodeFail,
			"message": "参数解析失败",
		})
		return
	}

	// 设置默认值
	if params.DeviceId == "" {
		params.DeviceId = "*"
	}
	if params.SourceClient == "" {
		params.SourceClient = "*"
	}

	// 将 Data 转换为 JSON 字符串
	var dataStr string
	if params.Data != nil {
		dataBytes, err := json.Marshal(params.Data)
		if err != nil {
			g.Log().Error(ctx, "序列化数据失败", err)
			r.Response.WriteJson(g.Map{
				"code":    consts.CodeFail,
				"message": "数据格式错误",
			})
			return
		}
		dataStr = string(dataBytes)
	}

	// 生成UUID
	uuidStr := uuid.New().String()

	// 准备消息输入
	in := &dto.PushMessageInput{
		CompanyUuid:  params.CompanyUuid,
		StaffUuid:    params.StaffUuid,
		NotStaffUuid: params.NotStaffUuid,
		SourceClient: params.SourceClient,
		DeviceId:     params.DeviceId,
		NotDeviceId:  params.NotDeviceId,
		MessageType:  params.MessageType,
		MessageKey:   params.MessageKey,
		Data:         dataStr,
	}

	// 准备消息任务
	task := messageTask{
		Params:     in,
		UUID:       uuidStr,
		MessageKey: params.MessageKey,
	}

	// 设置缓存和计数
	if params.MessageKey != "" {
		// 使用MessageKey专用的细粒度锁
		keyMutex := getHttpMessageKeyMutex(params.MessageKey)
		keyMutex.Lock()

		_, err := g.Redis().Set(ctx, params.MessageKey, uuidStr)
		if err != nil {
			g.Log().Error(ctx, "设置Redis键失败", err)
		}
		_, err = g.Redis().Expire(ctx, params.MessageKey, 2)
		if err != nil {
			g.Log().Error(ctx, "设置过期时间失败", err)
		}

		// 检查消息次数
		task.CountKey = fmt.Sprintf("%s_count", params.MessageKey)
		countVal, _ := g.Redis().Get(ctx, task.CountKey)

		// 安全地将缓存值转换为int64
		if !countVal.IsNil() {
			task.Count = countVal.Int64()
		}

		_, err = g.Redis().Set(ctx, task.CountKey, task.Count+1)
		if err != nil {
			g.Log().Error(ctx, "设置计数器失败", err)
		}
		_, err = g.Redis().Expire(ctx, task.CountKey, 2)
		if err != nil {
			g.Log().Error(ctx, "设置计数器过期时间失败", err)
		}

		keyMutex.Unlock()
	}

	// 非阻塞地尝试获取信号量
	select {
	case <-messageSemaphore:
		// 成功获取信号量，启动goroutine处理
		go processHttpMessage(ctx, task)
	default:
		// 信号量已满，直接处理
		go processHttpMessage(ctx, task)
	}

	// 立即返回成功响应，不等待处理完成
	r.Response.WriteJson(g.Map{
		"code":    consts.CodeSuccess,
		"message": "success",
	})
}

// processHttpMessage 处理HTTP消息的函数，在goroutine中运行
func processHttpMessage(ctx context.Context, task messageTask) {
	defer func() {
		// 处理完成后释放信号量
		select {
		case messageSemaphore <- struct{}{}:
		default:
		}
	}()

	// 标记是否应该发送消息
	shouldSend := true

	// 检查Redis缓存是否被更新或消息次数超过10
	if task.MessageKey != "" {
		// 等待300毫秒
		time.Sleep(300 * time.Millisecond)

		// 检查消息次数
		if task.Count <= 10 {
			// 使用MessageKey专用的细粒度锁
			keyMutex := getHttpMessageKeyMutex(task.MessageKey)
			keyMutex.Lock()

			// 检查UUID
			cachedUUID, _ := g.Redis().Get(ctx, task.MessageKey)
			if !cachedUUID.IsNil() {
				uuidStr := cachedUUID.String()
				if task.UUID != uuidStr {
					shouldSend = false
				}
			}

			keyMutex.Unlock()
		}
	}

	// 如果不应该发送消息，直接返回
	if !shouldSend {
		g.Log().Info(ctx, "HTTP防抖取消推送", "message_key", task.MessageKey)
		return
	}

	// 使用MessageKey专用的细粒度锁保护Redis操作
	if task.MessageKey != "" {
		keyMutex := getHttpMessageKeyMutex(task.MessageKey)
		keyMutex.Lock()
		if task.CountKey != "" {
			// 使用独立的context避免因原始context取消导致删除失败
			cleanupCtx := context.Background()
			_, err := g.Redis().Del(cleanupCtx, task.CountKey)
			if err != nil {
				g.Log().Warning(cleanupCtx, "删除计数器失败", err, "count_key", task.CountKey)
			}
		}
		keyMutex.Unlock()
	}

	// 调用业务逻辑层推送消息
	g.Log().Info(ctx, "HTTP推送消息", "message_key", task.MessageKey)
	_, err := service.Websocket().PushMessage(ctx, task.Params)
	if err != nil {
		g.Log().Error(ctx, "HTTP推送消息失败", err, "message_key", task.MessageKey)
	}
}
