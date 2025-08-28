package queue

import (
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

type MqProducer interface {
	SendMsg(topic string, body string) (mqMsg MqMsg, err error)
	SendByteMsg(topic string, body []byte) (mqMsg MqMsg, err error)
	SendDelayMsg(topic string, body string, delay time.Duration) (mqMsg MqMsg, err error)
}

type MqConsumer interface {
	ListenReceiveMsgDo(topic string, receiveDo func(mqMsg MqMsg) error) (err error)
}

const (
	_ = iota
	SendMsg
	ReceiveMsg
)

type Config struct {
	Switch    bool   `json:"switch"`
	Driver    string `json:"driver"`
	GroupName string `json:"groupName"`
	PoolSize  int    `json:"poolSize"`
	Redis     RedisConf
	Rocketmq  RocketmqConf
	Kafka     KafkaConf
}

type RedisConf struct {
	Timeout int64 `json:"timeout"`
}

type RocketmqConf struct {
	NameSrvAdders []string `json:"nameSrvAdders"`
	Endpoint      string   `json:"endpoint"`
	AccessKey     string   `json:"accessKey"`
	SecretKey     string   `json:"secretKey"`
	BrokerAddr    string   `json:"brokerAddr"`
	Retry         int      `json:"retry"`
	LogLevel      string   `json:"logLevel"`
}

type KafkaConf struct {
	Address       []string `json:"address"`
	Version       string   `json:"version"`
	RandClient    bool     `json:"randClient"`
	MultiConsumer bool     `json:"multiConsumer"`
}

type MqMsg struct {
	RunType   int       `json:"run_type"`
	Topic     string    `json:"topic"`
	MsgId     string    `json:"msg_id"`
	Offset    int64     `json:"offset"`
	Partition int32     `json:"partition"`
	Timestamp time.Time `json:"timestamp"`
	Body      []byte    `json:"body"`
}

var (
	ctx                   = gctx.GetInitCtx()
	mqProducerInstanceMap map[string]MqProducer
	mqConsumerInstanceMap map[string]MqConsumer
	mutex                 sync.Mutex
	config                Config
)

func init() {
	mqProducerInstanceMap = make(map[string]MqProducer)
	mqConsumerInstanceMap = make(map[string]MqConsumer)
	if err := g.Cfg().MustGet(ctx, "queue").Scan(&config); err != nil {
		Logger().Warningf(ctx, "queue init err:%+v", err)
	}
}

// InstanceConsumer 实例化消费者
func InstanceConsumer() (mqClient MqConsumer, err error) {
	return NewConsumer(config.GroupName)
}

// InstanceProducer 实例化生产者
func InstanceProducer() (mqClient MqProducer, err error) {
	return NewProducer(config.GroupName)
}

// NewProducer 初始化生产者实例
func NewProducer(groupName string) (mqClient MqProducer, err error) {
	if item, ok := mqProducerInstanceMap[groupName]; ok {
		return item, nil
	}

	if groupName == "" {
		err = gerror.New("mq groupName is empty.")
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	switch config.Driver {
	case "rocketmq":
		//if len(config.Rocketmq.Endpoint) == 0 {
		//	err = gerror.New("queue.rocketmq.endpoint is empty.")
		//	return
		//}
		mqClient, err = RegisterRocketProducer()
	//case "kafka":
	//	if len(config.Kafka.Address) == 0 {
	//		err = gerror.New("queue kafka address is not support")
	//		return
	//	}
	//	mqClient, err = RegisterKafkaProducer(KafkaConfig{
	//		Brokers: config.Kafka.Address,
	//		GroupID: groupName,
	//		Version: config.Kafka.Version,
	//	})
	case "redis":
		if _, err = g.Redis().Do(ctx, "ping"); err == nil {
			mqClient = RegisterRedisMqProducer(RedisOption{
				Timeout: config.Redis.Timeout,
			}, groupName)
		}

	//case "disk":
	//	config.Disk.GroupName = groupName
	//	mqClient, err = RegisterDiskMqProducer(config.Disk)
	default:
		err = gerror.New("queue driver is not support")
	}

	if err != nil {
		Logger().Error(ctx, err)
		return
	}
	mqProducerInstanceMap[groupName] = mqClient
	return
}

// NewConsumer 初始化消费者实例
func NewConsumer(groupName string) (mqClient MqConsumer, err error) {
	if groupName == "" {
		err = gerror.New("mq groupName is empty.")
		return
	}

	switch config.Driver {
	case "rocketmq":
		if len(config.Rocketmq.NameSrvAdders) == 0 {
			err = gerror.New("queue.rocketmq.nameSrvAdders is empty.")
			return
		}
		mqClient, err = RegisterRocketConsumer()

	case "redis":
		if _, err = g.Redis().Do(ctx, "ping"); err == nil {
			mqClient = RegisterRedisMqConsumer(RedisOption{
				Timeout: config.Redis.Timeout,
			}, groupName)
		}
	default:
		err = gerror.New("queue driver is not support")
	}

	if err != nil {
		return
	}

	mutex.Lock()
	defer mutex.Unlock()
	mqConsumerInstanceMap[groupName] = mqClient
	return
}

// BodyString 返回消息体
func (m *MqMsg) BodyString() string {
	return string(m.Body)
}
