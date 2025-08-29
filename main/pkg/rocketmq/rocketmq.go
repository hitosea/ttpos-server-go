package rocketmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap"
)

// Config RocketMQ 配置
type Config struct {
	NameServers []string `json:"name_servers"` // NameServer 地址列表
	AccessKey   string   `json:"access_key"`   // 访问密钥
	SecretKey   string   `json:"secret_key"`   // 密钥
	GroupName   string   `json:"group_name"`   // 消费者组名
	Retry       int      `json:"retry"`        // 重试次数
	LogLevel    string   `json:"log_level"`    // 日志级别
	Enabled     bool     `json:"enabled"`      //是否启用
}

// MessageHandler 消息处理函数类型
type MessageHandler func(ctx context.Context, msg *primitive.MessageExt) error

// Consumer RocketMQ 消费者
type Consumer struct {
	config    *Config
	client    rocketmq.PushConsumer
	logger    *zap.Logger
	handlers  map[string]MessageHandler
	mu        sync.RWMutex
	isStarted bool
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewConsumer 创建新的 RocketMQ 消费者
func NewConsumer(config *Config, logger *zap.Logger) *Consumer {
	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		config:   config,
		logger:   logger,
		handlers: make(map[string]MessageHandler),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// validateConfig 验证配置
func (c *Consumer) validateConfig() error {
	if c.config == nil {
		return fmt.Errorf("RocketMQ 配置不能为空")
	}
	if len(c.config.NameServers) == 0 {
		return fmt.Errorf("NameServer 地址不能为空")
	}
	if c.config.GroupName == "" {
		return fmt.Errorf("消费者组名不能为空")
	}
	return nil
}

// Start 启动消费者
func (c *Consumer) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isStarted {
		return fmt.Errorf("消费者已经启动")
	}

	// 验证配置
	if err := c.validateConfig(); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	// 创建消费者选项
	options := []consumer.Option{
		consumer.WithGroupName(c.config.GroupName),
		consumer.WithNsResolver(primitive.NewPassthroughResolver(c.config.NameServers)),
		consumer.WithConsumerModel(consumer.Clustering),               // 集群消费模式
		consumer.WithConsumeFromWhere(consumer.ConsumeFromLastOffset), // 从最新位置开始消费
	}

	// 添加认证信息（如果配置了）
	if c.config.AccessKey != "" && c.config.SecretKey != "" {
		options = append(options, consumer.WithCredentials(primitive.Credentials{
			AccessKey: c.config.AccessKey,
			SecretKey: c.config.SecretKey,
		}))
	}

	// 创建消费者客户端
	client, err := rocketmq.NewPushConsumer(options...)
	if err != nil {
		return fmt.Errorf("创建 RocketMQ 消费者失败: %w", err)
	}

	c.client = client
	c.isStarted = true

	c.logger.Info("RocketMQ 消费者启动成功",
		zap.String("group_name", c.config.GroupName),
		zap.Strings("name_servers", c.config.NameServers))

	return nil
}

// Subscribe 订阅主题
func (c *Consumer) Subscribe(topic string, handler MessageHandler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isStarted {
		return fmt.Errorf("消费者未启动，请先调用 Start() 方法")
	}

	if topic == "" {
		return fmt.Errorf("主题名称不能为空")
	}

	if handler == nil {
		return fmt.Errorf("消息处理函数不能为空")
	}

	// 保存处理函数
	c.handlers[topic] = handler

	// 订阅主题
	err := c.client.Subscribe(topic, consumer.MessageSelector{}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range msgs {
			// 处理消息
			if err := c.handleMessage(ctx, topic, msg); err != nil {
				c.logger.Error("处理消息失败",
					zap.String("topic", topic),
					zap.String("msg_id", msg.MsgId),
					zap.Error(err))
				// 返回重试，让消息重新投递
				return consumer.ConsumeRetryLater, err
			}
		}
		return consumer.ConsumeSuccess, nil
	})

	if err != nil {
		return fmt.Errorf("订阅主题失败 [%s]: %w", topic, err)
	}

	// 启动消费者客户端（如果还没启动）
	if err := c.client.Start(); err != nil {
		return fmt.Errorf("启动消费者客户端失败: %w", err)
	}

	c.logger.Info("成功订阅主题",
		zap.String("topic", topic),
		zap.String("group_name", c.config.GroupName))

	return nil
}

// handleMessage 处理单个消息
func (c *Consumer) handleMessage(ctx context.Context, topic string, msg *primitive.MessageExt) error {
	startTime := time.Now()

	// 获取处理函数
	c.mu.RLock()
	handler, exists := c.handlers[topic]
	c.mu.RUnlock()

	if !exists {
		return fmt.Errorf("未找到主题 [%s] 的处理函数", topic)
	}

	// 添加超时控制
	msgCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 执行处理函数
	err := handler(msgCtx, msg)
	duration := time.Since(startTime)

	// 记录处理结果
	if err != nil {
		c.logger.Error("消息处理失败",
			zap.String("topic", topic),
			zap.String("msg_id", msg.MsgId),
			zap.Int("queue_id", int(msg.Queue.QueueId)),
			zap.Duration("duration", duration),
			zap.Error(err))
	} else {
		// 如果处理时间过长，记录警告
		if duration > 5*time.Second {
			c.logger.Warn("消息处理耗时过长",
				zap.String("topic", topic),
				zap.String("msg_id", msg.MsgId),
				zap.Duration("duration", duration))
		} else {
			c.logger.Debug("消息处理成功",
				zap.String("topic", topic),
				zap.String("msg_id", msg.MsgId),
				zap.Duration("duration", duration))
		}
	}

	return err
}

// Unsubscribe 取消订阅主题
func (c *Consumer) Unsubscribe(topic string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isStarted {
		return fmt.Errorf("消费者未启动")
	}

	// 从客户端取消订阅
	if err := c.client.Unsubscribe(topic); err != nil {
		return fmt.Errorf("取消订阅主题失败 [%s]: %w", topic, err)
	}

	// 删除处理函数
	delete(c.handlers, topic)

	c.logger.Info("成功取消订阅主题",
		zap.String("topic", topic),
		zap.String("group_name", c.config.GroupName))

	return nil
}

// Stop 停止消费者
func (c *Consumer) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isStarted {
		return nil
	}

	// 取消上下文
	c.cancel()

	// 关闭客户端
	if c.client != nil {
		if err := c.client.Shutdown(); err != nil {
			c.logger.Error("关闭 RocketMQ 消费者失败", zap.Error(err))
			return fmt.Errorf("关闭消费者失败: %w", err)
		}
	}

	c.isStarted = false
	c.client = nil

	c.logger.Info("RocketMQ 消费者已停止",
		zap.String("group_name", c.config.GroupName))

	return nil
}

// IsStarted 检查消费者是否已启动
func (c *Consumer) IsStarted() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isStarted
}

// GetSubscribedTopics 获取已订阅的主题列表
func (c *Consumer) GetSubscribedTopics() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	topics := make([]string, 0, len(c.handlers))
	for topic := range c.handlers {
		topics = append(topics, topic)
	}
	return topics
}
