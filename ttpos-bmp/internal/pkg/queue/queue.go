package queue

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

// MqProducer 消息生产者接口
type MqProducer interface {
	// SendMsg 发送字符串消息
	SendMsg(ctx context.Context, topic string, body string) (mqMsg MqMsg, err error)
	// SendByteMsg 发送字节消息
	SendByteMsg(ctx context.Context, topic string, body []byte) (mqMsg MqMsg, err error)
	// SendDelayMsg 发送延时消息
	SendDelayMsg(ctx context.Context, topic string, body string, delay time.Duration) (mqMsg MqMsg, err error)
	// Close 关闭生产者连接
	Close() error
}

// MqConsumer 消息消费者接口
type MqConsumer interface {
	// Subscribe 订阅主题消息
	Subscribe(ctx context.Context, topic string, handler func(ctx context.Context, mqMsg MqMsg) error) error
	// Close 关闭消费者连接
	Close() error
}

// 消息运行类型常量
const (
	_ = iota
	// MsgTypeSend 发送消息类型
	MsgTypeSend
	// MsgTypeReceive 接收消息类型
	MsgTypeReceive
)

// Config 队列配置结构
type Config struct {
	Switch    bool         `json:"switch"`    // 队列开关
	Driver    string       `json:"driver"`    // 驱动类型：redis、rocketmq
	GroupName string       `json:"groupName"` // 组名称
	PoolSize  int          `json:"poolSize"`  // 连接池大小
	Redis     RedisConf    `json:"redis"`     // Redis配置
	Rocketmq  RocketmqConf `json:"rocketmq"`  // RocketMQ配置
	Kafka     KafkaConf    `json:"kafka"`     // Kafka配置
}

// RedisConf Redis消息队列配置
type RedisConf struct {
	Timeout int64 `json:"timeout"` // 超时时间（秒）
}

// RocketmqConf RocketMQ消息队列配置
type RocketmqConf struct {
	NameSrvAdders []string `json:"nameSrvAdders"` // Name Server地址列表
	Endpoint      string   `json:"endpoint"`      // 服务端点
	AccessKey     string   `json:"accessKey"`     // 访问密钥
	SecretKey     string   `json:"secretKey"`     // 密钥
	BrokerAddr    string   `json:"brokerAddr"`    // Broker地址
	Retry         int      `json:"retry"`         // 重试次数
	LogLevel      string   `json:"logLevel"`      // 日志级别
}

// KafkaConf Kafka消息队列配置
type KafkaConf struct {
	Address       []string `json:"address"`       // Kafka地址列表
	Version       string   `json:"version"`       // Kafka版本
	RandClient    bool     `json:"randClient"`    // 随机客户端
	MultiConsumer bool     `json:"multiConsumer"` // 多消费者模式
}

// MqMsg 消息队列消息结构
type MqMsg struct {
	RunType   int       `json:"run_type"`  // 运行类型：发送/接收
	Topic     string    `json:"topic"`     // 主题
	MsgId     string    `json:"msg_id"`    // 消息ID
	Offset    int64     `json:"offset"`    // 偏移量
	Partition int32     `json:"partition"` // 分区
	Timestamp time.Time `json:"timestamp"` // 时间戳
	Body      []byte    `json:"body"`      // 消息体
}

var (
	// 全局Context
	globalCtx = gctx.GetInitCtx()
	// 生产者实例映射
	mqProducerInstanceMap = make(map[string]MqProducer)
	// 消费者实例映射
	mqConsumerInstanceMap = make(map[string]MqConsumer)
	// 互斥锁保护实例映射
	mutex sync.RWMutex
	// 队列配置
	config Config
)

func init() {
	// 加载队列配置
	if err := g.Cfg().MustGet(globalCtx, "queue").Scan(&config); err != nil {
		Logger().Errorf(globalCtx, "队列初始化失败: %+v", err)
	} else {
		Logger().Infof(globalCtx, "队列初始化成功，驱动类型: %s, 组名: %s", config.Driver, config.GroupName)
	}
}

// InstanceConsumer 实例化消费者
func InstanceConsumer() (mqClient MqConsumer, err error) {
	if !config.Switch {
		return nil, gerror.New("队列服务未启用")
	}
	return NewConsumer(config.GroupName)
}

// InstanceProducer 实例化生产者
func InstanceProducer() (mqClient MqProducer, err error) {
	if !config.Switch {
		return nil, gerror.New("队列服务未启用")
	}
	return NewProducer(config.GroupName)
}

// NewProducer 初始化生产者实例
func NewProducer(groupName string) (mqClient MqProducer, err error) {
	// 读锁检查是否已存在
	mutex.RLock()
	if item, ok := mqProducerInstanceMap[groupName]; ok {
		mutex.RUnlock()
		return item, nil
	}
	mutex.RUnlock()

	// 参数验证
	if groupName == "" {
		return nil, gerror.New("队列组名不能为空")
	}

	// 写锁创建新实例
	mutex.Lock()
	defer mutex.Unlock()

	// 双重检查，避免并发创建
	if item, ok := mqProducerInstanceMap[groupName]; ok {
		return item, nil
	}

	// 根据驱动类型创建生产者
	switch config.Driver {
	case "rocketmq":
		mqClient, err = RegisterRocketProducer()
	case "redis":
		// 检查Redis连接
		if _, err = g.Redis().Do(globalCtx, "ping"); err != nil {
			return nil, gerror.Wrapf(err, "Redis连接检查失败")
		}
		mqClient = RegisterRedisMqProducer(RedisOption{
			Timeout: config.Redis.Timeout,
		}, groupName)
	default:
		return nil, gerror.Newf("不支持的队列驱动类型: %s", config.Driver)
	}

	if err != nil {
		Logger().Errorf(globalCtx, "创建生产者失败 [%s]: %+v", groupName, err)
		return nil, err
	}

	mqProducerInstanceMap[groupName] = mqClient
	Logger().Infof(globalCtx, "生产者创建成功 [%s], 驱动类型: %s", groupName, config.Driver)
	return mqClient, nil
}

// NewConsumer 初始化消费者实例
func NewConsumer(groupName string) (mqClient MqConsumer, err error) {
	// 参数验证
	if groupName == "" {
		return nil, gerror.New("队列组名不能为空")
	}

	// 根据驱动类型创建消费者
	switch config.Driver {
	case "rocketmq":
		if len(config.Rocketmq.NameSrvAdders) == 0 {
			return nil, gerror.New("RocketMQ NameServer地址配置为空")
		}
		mqClient, err = RegisterRocketConsumer()
	case "redis":
		// 检查Redis连接
		if _, err = g.Redis().Do(globalCtx, "ping"); err != nil {
			return nil, gerror.Wrapf(err, "Redis连接检查失败")
		}
		mqClient = RegisterRedisMqConsumer(RedisOption{
			Timeout: config.Redis.Timeout,
		}, groupName)
	default:
		return nil, gerror.Newf("不支持的队列驱动类型: %s", config.Driver)
	}

	if err != nil {
		Logger().Errorf(globalCtx, "创建消费者失败 [%s]: %+v", groupName, err)
		return nil, err
	}

	// 保存消费者实例
	mutex.Lock()
	mqConsumerInstanceMap[groupName] = mqClient
	mutex.Unlock()

	Logger().Infof(globalCtx, "消费者创建成功 [%s], 驱动类型: %s", groupName, config.Driver)
	return mqClient, nil
}

// BodyString 返回消息体字符串
func (m *MqMsg) BodyString() string {
	return string(m.Body)
}

// IsValid 检查消息是否有效
func (m *MqMsg) IsValid() bool {
	return m.Topic != "" && m.MsgId != "" && len(m.Body) > 0
}

// CloseAllProducers 关闭所有生产者连接
func CloseAllProducers() error {
	mutex.Lock()
	defer mutex.Unlock()

	for groupName, producer := range mqProducerInstanceMap {
		if err := producer.Close(); err != nil {
			Logger().Errorf(globalCtx, "关闭生产者失败 [%s]: %+v", groupName, err)
			return err
		}
		Logger().Infof(globalCtx, "生产者已关闭 [%s]", groupName)
	}
	// 清空映射
	mqProducerInstanceMap = make(map[string]MqProducer)
	return nil
}

// CloseAllConsumers 关闭所有消费者连接
func CloseAllConsumers() error {
	mutex.Lock()
	defer mutex.Unlock()

	for groupName, consumer := range mqConsumerInstanceMap {
		if err := consumer.Close(); err != nil {
			Logger().Errorf(globalCtx, "关闭消费者失败 [%s]: %+v", groupName, err)
			return err
		}
		Logger().Infof(globalCtx, "消费者已关闭 [%s]", groupName)
	}
	// 清空映射
	mqConsumerInstanceMap = make(map[string]MqConsumer)
	return nil
}
