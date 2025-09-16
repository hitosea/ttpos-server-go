package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
	"websocket/constant"
	"websocket/pkg/cache"
	"websocket/utils"

	"github.com/google/uuid"
)

// 全局限流器，控制同时运行的goroutine数量
var (
	// 最大并发处理的消息数
	maxConcurrentMessages = 500
	// 信号量通道，用于限制并发
	messageSemaphore = make(chan struct{}, maxConcurrentMessages)
	// 基于MessageKey的细粒度锁映射
	messageKeyMutexes = sync.Map{} // map[string]*sync.Mutex
	// 保护messageKeyMutexes的锁
	mutexMapLock sync.Mutex
)

// 消息任务结构体
type messageTask struct {
	Params     interface{}
	UUID       string
	MessageKey string
	CountKey   string
	Count      int64
	Writer     http.ResponseWriter
}

// 初始化函数，在包初始化时调用
func init() {
	// 初始化信号量通道
	for i := 0; i < maxConcurrentMessages; i++ {
		messageSemaphore <- struct{}{}
	}
}

// getMessageKeyMutex 获取指定MessageKey的专用锁
func getMessageKeyMutex(messageKey string) *sync.Mutex {
	// 先尝试获取已存在的锁
	if mutex, exists := messageKeyMutexes.Load(messageKey); exists {
		return mutex.(*sync.Mutex)
	}

	// 如果不存在，需要创建新锁
	mutexMapLock.Lock()
	defer mutexMapLock.Unlock()

	// 双重检查，防止并发创建
	if mutex, exists := messageKeyMutexes.Load(messageKey); exists {
		return mutex.(*sync.Mutex)
	}

	// 创建新锁并存储
	newMutex := &sync.Mutex{}
	messageKeyMutexes.Store(messageKey, newMutex)

	return newMutex
}

func PushClient(w http.ResponseWriter, r *http.Request) {
	// 解析请求参数
	var params struct {
		CompanyUuid  uint        `json:"company_uuid"`
		SourceClient string      `json:"source_client"`
		DeviceId     string      `json:"device_id"`
		NotDeviceId  string      `json:"not_device_id"`
		StaffUuid    uint64      `json:"staff_uuid"`
		NotStaffUuid uint64      `json:"not_staff_uuid"`
		MessageType  string      `json:"message_type"`
		MessageKey   string      `json:"message_key"`
		Data         interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		fmt.Println("Failed to decode JSON", err)
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}

	// 设置默认值
	if params.DeviceId == "" {
		params.DeviceId = "*"
	}
	if params.SourceClient == "" {
		params.SourceClient = "*"
	}

	// 生成UUID
	uuidStr := uuid.New().String()

	// 准备消息任务
	task := messageTask{
		Params:     params,
		UUID:       uuidStr,
		MessageKey: params.MessageKey,
		Writer:     w,
	}

	// 设置缓存和计数
	if params.MessageKey != "" {
		// 使用MessageKey专用的细粒度锁
		keyMutex := getMessageKeyMutex(params.MessageKey)
		keyMutex.Lock()
		cache.GlobalRedis.Set(params.MessageKey, uuidStr, 2*time.Second)
		// 检查消息次数
		task.CountKey = fmt.Sprintf("%s_count", params.MessageKey)
		if countVal, countExists := cache.GlobalRedis.Get(task.CountKey); countExists {
			// 安全地将缓存值转换为int64
			switch v := countVal.(type) {
			case int64:
				task.Count = v
			case string:
				// 尝试将字符串转换为int64
				if intVal, err := strconv.ParseInt(v, 10, 64); err == nil {
					task.Count = intVal
				} else {
					// 转换失败，使用默认值0
					fmt.Printf("Failed to parse count string '%s': %v\n", v, err)
					task.Count = 11
				}
			default:
				// 如果是其他类型，使用默认值0
				fmt.Printf("Unexpected type for count: %T\n", v)
				task.Count = 11
			}
		}
		cache.GlobalRedis.Set(task.CountKey, fmt.Sprintf("%d", task.Count+1), 2*time.Second)
		keyMutex.Unlock()
	}

	// 非阻塞地尝试获取信号量
	select {
	case <-messageSemaphore:
		// 成功获取信号量，启动goroutine处理
		go processMessage(task)
	default:
	}

	// 立即返回成功响应，不等待处理完成
	fmt.Fprintf(w, "%s", utils.StructToJson(map[string]interface{}{
		"code":    constant.CodeSuccess,
		"message": "success",
	}))
}

// 处理消息的函数，在goroutine中运行
func processMessage(task messageTask) {
	defer func() {
		// 处理完成后释放信号量
		messageSemaphore <- struct{}{}
	}()

	// 标记是否应该发送消息
	shouldSend := true

	// 检查Redis缓存是否被更新或消息次数超过50
	if task.MessageKey != "" {
		// 等待900毫秒
		time.Sleep(900 * time.Millisecond)

		// 检查消息次数
		if task.Count <= 10 {
			// 使用MessageKey专用的细粒度锁
			keyMutex := getMessageKeyMutex(task.MessageKey)
			keyMutex.Lock()
			// 检查UUID
			if cachedUUID, exists := cache.GlobalRedis.Get(task.MessageKey); exists {
				// 安全地转换类型
				var uuidStr string
				switch v := cachedUUID.(type) {
				case string:
					uuidStr = v
				default:
					// 如果不是字符串，尝试转换
					uuidStr = fmt.Sprintf("%v", cachedUUID)
				}
				if task.UUID != uuidStr {
					shouldSend = false
				}
			}
			keyMutex.Unlock()
		}
	}

	// 如果不应该发送消息，直接返回
	if !shouldSend {
		return
	}

	// 推送消息
	params := task.Params

	// 使用MessageKey专用的细粒度锁保护Redis操作
	if task.MessageKey != "" {
		keyMutex := getMessageKeyMutex(task.MessageKey)
		keyMutex.Lock()
		if task.CountKey != "" {
			cache.GlobalRedis.Del(task.CountKey)
		}
		keyMutex.Unlock()
	}

	// Redis发布操作不需要锁，因为Redis本身是线程安全的
	err := cache.GlobalRedis.Publish("websocket_msg_push", utils.StructToJson(params))

	if err != nil {
		// 注意：这里不能再写入HTTP响应，因为响应可能已经发送
		fmt.Println("Error publishing message:", err)
	}
}
