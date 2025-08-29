package rocketmq

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// Manager RocketMQ 消费者管理器
type Manager struct {
	consumers map[string]*Consumer
	logger    *zap.Logger
	mu        sync.RWMutex
}

// NewManager 创建新的消费者管理器
func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		consumers: make(map[string]*Consumer),
		logger:    logger,
	}
}

// RegisterConsumer 注册消费者
func (m *Manager) RegisterConsumer(name string, config *Config) (*Consumer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已存在同名消费者
	if _, exists := m.consumers[name]; exists {
		return nil, fmt.Errorf("消费者 [%s] 已存在", name)
	}

	// 创建消费者
	consumer := NewConsumer(config, m.logger.With(zap.String("consumer", name)))
	m.consumers[name] = consumer

	m.logger.Info("注册消费者",
		zap.String("name", name),
		zap.String("group_name", config.GroupName))

	return consumer, nil
}

// GetConsumer 获取消费者
func (m *Manager) GetConsumer(name string) (*Consumer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	consumer, exists := m.consumers[name]
	if !exists {
		return nil, fmt.Errorf("消费者 [%s] 不存在", name)
	}

	return consumer, nil
}

// StartConsumer 启动指定消费者
func (m *Manager) StartConsumer(name string) error {
	consumer, err := m.GetConsumer(name)
	if err != nil {
		return err
	}

	return consumer.Start()
}

// StartAll 启动所有消费者
func (m *Manager) StartAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errors []error
	for name, consumer := range m.consumers {
		if err := consumer.Start(); err != nil {
			errors = append(errors, fmt.Errorf("启动消费者 [%s] 失败: %w", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("启动消费者失败: %v", errors)
	}

	m.logger.Info("所有消费者启动成功", zap.Int("count", len(m.consumers)))
	return nil
}

// StopConsumer 停止指定消费者
func (m *Manager) StopConsumer(name string) error {
	consumer, err := m.GetConsumer(name)
	if err != nil {
		return err
	}

	return consumer.Stop()
}

// StopAll 停止所有消费者
func (m *Manager) StopAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errors []error
	for name, consumer := range m.consumers {
		if err := consumer.Stop(); err != nil {
			errors = append(errors, fmt.Errorf("停止消费者 [%s] 失败: %w", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("停止消费者失败: %v", errors)
	}

	m.logger.Info("所有消费者已停止", zap.Int("count", len(m.consumers)))
	return nil
}

// RemoveConsumer 移除消费者
func (m *Manager) RemoveConsumer(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	consumer, exists := m.consumers[name]
	if !exists {
		return fmt.Errorf("消费者 [%s] 不存在", name)
	}

	// 确保消费者已停止
	if consumer.IsStarted() {
		if err := consumer.Stop(); err != nil {
			return fmt.Errorf("停止消费者 [%s] 失败: %w", name, err)
		}
	}

	delete(m.consumers, name)
	m.logger.Info("移除消费者", zap.String("name", name))

	return nil
}

// Subscribe 为指定消费者订阅主题
func (m *Manager) Subscribe(consumerName, topic string, handler MessageHandler) error {
	consumer, err := m.GetConsumer(consumerName)
	if err != nil {
		return err
	}

	return consumer.Subscribe(topic, handler)
}

// Unsubscribe 为指定消费者取消订阅主题
func (m *Manager) Unsubscribe(consumerName, topic string) error {
	consumer, err := m.GetConsumer(consumerName)
	if err != nil {
		return err
	}

	return consumer.Unsubscribe(topic)
}

// GetConsumerNames 获取所有消费者名称
func (m *Manager) GetConsumerNames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.consumers))
	for name := range m.consumers {
		names = append(names, name)
	}
	return names
}

// GetConsumerInfo 获取消费者信息
func (m *Manager) GetConsumerInfo(name string) (map[string]interface{}, error) {
	consumer, err := m.GetConsumer(name)
	if err != nil {
		return nil, err
	}

	info := map[string]interface{}{
		"name":              name,
		"group_name":        consumer.config.GroupName,
		"name_servers":      consumer.config.NameServers,
		"is_started":        consumer.IsStarted(),
		"subscribed_topics": consumer.GetSubscribedTopics(),
	}

	return info, nil
}

// GetAllConsumerInfo 获取所有消费者信息
func (m *Manager) GetAllConsumerInfo() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info := make(map[string]interface{})
	for name := range m.consumers {
		if consumerInfo, err := m.GetConsumerInfo(name); err == nil {
			info[name] = consumerInfo
		}
	}

	return info
}

// HealthCheck 健康检查
func (m *Manager) HealthCheck(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errors []error
	for name, consumer := range m.consumers {
		if !consumer.IsStarted() {
			errors = append(errors, fmt.Errorf("消费者 [%s] 未启动", name))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("健康检查失败: %v", errors)
	}

	return nil
}

// Shutdown 优雅关闭所有消费者
func (m *Manager) Shutdown(ctx context.Context) error {
	m.logger.Info("开始关闭 RocketMQ 消费者管理器")

	// 停止所有消费者
	if err := m.StopAll(); err != nil {
		m.logger.Error("停止消费者失败", zap.Error(err))
		return err
	}

	// 清空消费者映射
	m.mu.Lock()
	m.consumers = make(map[string]*Consumer)
	m.mu.Unlock()

	m.logger.Info("RocketMQ 消费者管理器已关闭")
	return nil
}
