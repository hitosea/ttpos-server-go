package queue

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"
)

// Consumer 消费者接口，实现该接口即可加入到消费队列中
type Consumer interface {
	// GetTopic 获取消费主题
	GetTopic() string
	// Handle 处理消息的方法
	Handle(ctx context.Context, mqMsg MqMsg) error
	// GetConcurrency 获取并发处理数（可选，默认为1）
	GetConcurrency() int
}

// consumerManager 消费者管理器
type consumerManager struct {
	sync.RWMutex
	list    map[string]Consumer // 维护的消费者列表
	running bool                // 消费者是否正在运行
	cancel  context.CancelFunc  // 用于停止所有消费者的cancel函数
}

var consumers = &consumerManager{
	list:    make(map[string]Consumer),
	running: false,
}

// RegisterConsumer 注册任务到消费者队列
func RegisterConsumer(cs Consumer) error {
	// 参数验证
	if cs == nil {
		return gerror.New("消费者不能为空")
	}
	topic := cs.GetTopic()
	if topic == "" {
		return gerror.New("消费者主题不能为空")
	}

	consumers.Lock()
	defer consumers.Unlock()

	// 检查是否已经注册
	if _, ok := consumers.list[topic]; ok {
		Logger().Warningf(globalCtx, "消费者主题 [%s] 重复注册", topic)
		return gerror.Newf("消费者主题 [%s] 已存在", topic)
	}

	consumers.list[topic] = cs
	Logger().Infof(globalCtx, "消费者注册成功 [%s]", topic)
	return nil
}

// StartConsumersListener 启动所有已注册的消费者监听
func StartConsumersListener(ctx context.Context) error {
	consumers.Lock()
	defer consumers.Unlock()

	if consumers.running {
		return gerror.New("消费者监听器已经在运行")
	}

	if len(consumers.list) == 0 {
		Logger().Warning(globalCtx, "没有注册的消费者")
		return nil
	}

	// 创建可取消的Context
	ctxWithCancel, cancel := context.WithCancel(ctx)
	consumers.cancel = cancel
	consumers.running = true

	// 启动所有消费者
	for topic, consumer := range consumers.list {
		go func(topic string, consumer Consumer) {
			if err := consumerListen(ctxWithCancel, consumer); err != nil {
				Logger().Errorf(ctxWithCancel, "消费者 [%s] 监听失败: %+v", topic, err)
			}
		}(topic, consumer)
	}

	Logger().Infof(globalCtx, "所有消费者监听器已启动，总数: %d", len(consumers.list))
	return nil
}

// consumerListen 消费者监听
func consumerListen(ctx context.Context, job Consumer) error {
	topic := job.GetTopic()

	// 创建消费者实例
	consumer, err := InstanceConsumer()
	if err != nil {
		return gerror.Wrapf(err, "创建消费者实例失败 [%s]", topic)
	}

	// 启动消息监听
	Logger().Infof(ctx, "启动消费者监听 [%s]", topic)

	// 使用新的接口方法
	err = consumer.Subscribe(ctx, topic, func(msgCtx context.Context, mqMsg MqMsg) error {
		// 处理消息
		handleErr := job.Handle(msgCtx, mqMsg)
		// 记录消费日志
		ConsumerLog(msgCtx, topic, mqMsg, handleErr)
		return handleErr
	})

	if err != nil {
		return gerror.Wrapf(err, "消费者监听失败 [%s]", topic)
	}

	return nil
}

// StopConsumersListener 停止所有消费者监听
func StopConsumersListener() error {
	consumers.Lock()
	defer consumers.Unlock()

	if !consumers.running {
		return gerror.New("消费者监听器未运行")
	}

	// 取消所有消费者
	if consumers.cancel != nil {
		consumers.cancel()
		consumers.cancel = nil
	}

	consumers.running = false
	Logger().Info(globalCtx, "所有消费者监听器已停止")
	return nil
}

// UnregisterConsumer 取消注册消费者
func UnregisterConsumer(topic string) error {
	if topic == "" {
		return gerror.New("主题名称不能为空")
	}

	consumers.Lock()
	defer consumers.Unlock()

	if _, ok := consumers.list[topic]; !ok {
		return gerror.Newf("消费者主题 [%s] 不存在", topic)
	}

	delete(consumers.list, topic)
	Logger().Infof(globalCtx, "消费者取消注册成功 [%s]", topic)
	return nil
}

// GetRegisteredConsumers 获取所有已注册的消费者主题
func GetRegisteredConsumers() []string {
	consumers.RLock()
	defer consumers.RUnlock()

	topics := make([]string, 0, len(consumers.list))
	for topic := range consumers.list {
		topics = append(topics, topic)
	}
	return topics
}

// IsConsumerRunning 检查消费者监听器是否正在运行
func IsConsumerRunning() bool {
	consumers.RLock()
	defer consumers.RUnlock()
	return consumers.running
}
