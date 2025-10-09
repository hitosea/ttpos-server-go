package queue

import (
	"context"
	"sync"
	"time"
	"ttpos-bmp/internal/consts"
	"ttpos-bmp/utility/simple"
	"ttpos-bmp/utility/validate"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/apache/rocketmq-client-go/v2/rlog"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/grpool"
)

type RocketMq struct {
	producerIns rocketmq.Producer
	consumerIns rocketmq.PushConsumer
}

type RocketManager struct {
	Producer *RocketMq
	Consumer *RocketMq
	pMutex   sync.Mutex
	cMutex   sync.Mutex
	started  bool
	goPool   *grpool.Pool
}

var rocketManager = &RocketManager{}

func init() {
	setRocketCloseEvent()
}

func setRocketCloseEvent() {
	simple.Event().Register(consts.EventServerClose, func(ctx context.Context, args ...interface{}) {
		if rocketManager == nil {
			return
		}

		if rocketManager.Producer != nil {
			err := rocketManager.Producer.producerIns.Shutdown()
			if err != nil {
				Logger().Warningf(ctx, "rocketmq producer close err:%v", err)
				return
			}
			Logger().Debug(ctx, "rocketmq producer close...")
		}

		if rocketManager.Consumer != nil {
			err := rocketManager.Consumer.consumerIns.Shutdown()
			if err != nil {
				Logger().Warningf(ctx, "rocketmq consumer close err:%v", err)
				return
			}
			Logger().Debug(ctx, "rocketmq consumer close...")
		}
		rocketManager.started = false

		for rocketManager.goPool != nil && rocketManager.goPool.Size() != 0 {
			Logger().Debugf(ctx, "waiting for rocketmq consumer to complete execution[%v][%v]...", rocketManager.goPool.Size(), rocketManager.goPool.Jobs())
			time.Sleep(time.Second)
		}
	})
}

func GetRocketManager() *RocketManager {
	return rocketManager
}

// RegisterRocketProducer 注册并启动生产者接口实现
func RegisterRocketProducer() (client MqProducer, err error) {
	return RegisterRocketMqProducer()
}

// RegisterRocketConsumer 注册消费者
func RegisterRocketConsumer() (client MqConsumer, err error) {
	return RegisterRocketMqConsumer()
}

// createTopicIfNotExists 主题不存在就自动创建
func (r *RocketMq) createTopicIfNotExists(topic string) (err error) {
	if len(config.Rocketmq.BrokerAddr) == 0 {
		return
	}

	client, err := admin.NewAdmin(
		admin.WithResolver(primitive.NewPassthroughResolver(config.Rocketmq.NameSrvAdders)),
		admin.WithCredentials(primitive.Credentials{
			AccessKey: config.Rocketmq.AccessKey,
			SecretKey: config.Rocketmq.SecretKey,
		}),
	)
	if err != nil {
		return err
	}
	defer client.Close()

	result, err := client.FetchAllTopicList(globalCtx)
	if err != nil {
		return err
	}

	if validate.InSlice(result.TopicList, topic) {
		return
	}

	Logger().Debugf(globalCtx, "create topic:%v", topic)
	err = client.CreateTopic(globalCtx, admin.WithTopicCreate(topic), admin.WithBrokerAddrCreate(config.Rocketmq.BrokerAddr))
	return
}

// SendMsg 按字符串类型生产数据
func (r *RocketMq) SendMsg(ctx context.Context, topic string, body string) (mqMsg MqMsg, err error) {
	return r.SendByteMsg(ctx, topic, []byte(body))
}

// SendByteMsg 生产数据
func (r *RocketMq) SendByteMsg(ctx context.Context, topic string, body []byte) (mqMsg MqMsg, err error) {
	// 参数验证
	if r.producerIns == nil {
		return mqMsg, gerror.New("RocketMQ生产者未初始化")
	}
	if topic == "" {
		return mqMsg, gerror.New("主题名称不能为空")
	}
	if len(body) == 0 {
		return mqMsg, gerror.New("消息内容不能为空")
	}

	// 自动创建主题
	if err = r.createTopicIfNotExists(topic); err != nil {
		return mqMsg, gerror.Wrapf(err, "创建主题失败 [%s]", topic)
	}

	// 发送消息
	startTime := time.Now()
	result, err := r.producerIns.SendSync(ctx, &primitive.Message{
		Topic: topic,
		Body:  body,
	})
	duration := time.Since(startTime)

	if err != nil {
		return mqMsg, gerror.Wrapf(err, "RocketMQ发送消息失败 [%s]", topic)
	}
	if result.Status != primitive.SendOK {
		return mqMsg, gerror.Newf("RocketMQ发送消息状态异常 [%s]: %v", topic, result.Status)
	}

	// 构建返回消息
	mqMsg = MqMsg{
		RunType:   MsgTypeSend,
		Topic:     topic,
		MsgId:     result.MsgID,
		Body:      body,
		Timestamp: time.Now(),
	}

	// 记录性能监控
	if duration > 500*time.Millisecond {
		Logger().Warningf(ctx, "RocketMQ发送消息耗时 [%s] - 消息ID: %s, 耗时: %v",
			topic, result.MsgID, duration)
	}

	return mqMsg, nil
}

// SendDelayMsg 发送延时消息
func (r *RocketMq) SendDelayMsg(ctx context.Context, topic string, body string, delay time.Duration) (mqMsg MqMsg, err error) {
	// 参数验证
	if r.producerIns == nil {
		return mqMsg, gerror.New("RocketMQ生产者未初始化")
	}
	if topic == "" {
		return mqMsg, gerror.New("主题名称不能为空")
	}
	if body == "" {
		return mqMsg, gerror.New("消息内容不能为空")
	}
	if delay < 0 {
		return mqMsg, gerror.New("延时时间不能为负数")
	}

	// 创建延时消息
	msg := primitive.NewMessage(topic, []byte(body))
	msg.WithDelayTimestamp(time.Now().Add(delay))

	// 发送消息
	startTime := time.Now()
	result, err := r.producerIns.SendSync(ctx, msg)
	duration := time.Since(startTime)

	if err != nil {
		return mqMsg, gerror.Wrapf(err, "RocketMQ发送延时消息失败 [%s]", topic)
	}
	if result.Status != primitive.SendOK {
		return mqMsg, gerror.Newf("RocketMQ发送延时消息状态异常 [%s]: %v", topic, result.Status)
	}

	// 构建返回消息
	mqMsg = MqMsg{
		RunType:   MsgTypeSend,
		Topic:     topic,
		MsgId:     result.MsgID,
		Body:      []byte(body),
		Timestamp: time.Now(),
	}

	// 记录性能监控
	if duration > 500*time.Millisecond {
		Logger().Warningf(ctx, "RocketMQ发送延时消息耗时 [%s] - 消息ID: %s, 延时: %v, 耗时: %v",
			topic, result.MsgID, delay, duration)
	}

	return mqMsg, nil
}

// Subscribe 订阅主题消息（新接口方法）
func (r *RocketMq) Subscribe(ctx context.Context, topic string, handler func(ctx context.Context, mqMsg MqMsg) error) error {
	// 参数验证
	if r.consumerIns == nil {
		return gerror.New("RocketMQ消费者未初始化")
	}
	if topic == "" {
		return gerror.New("主题名称不能为空")
	}
	if handler == nil {
		return gerror.New("消息处理函数不能为空")
	}

	rocketManager.cMutex.Lock()
	defer rocketManager.cMutex.Unlock()

	// 订阅主题
	err := r.consumerIns.Subscribe(topic, consumer.MessageSelector{}, func(consumerCtx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range msgs {
			// 使用协程池处理消息
			_ = rocketManager.goPool.Add(consumerCtx, func(poolCtx context.Context) {
				startTime := time.Now()

				// 构建消息对象
				mqMsg := MqMsg{
					RunType:   MsgTypeReceive,
					Topic:     msg.Topic,
					MsgId:     msg.MsgId,
					Body:      msg.Body,
					Timestamp: time.Now(),
				}

				// 处理消息
				handleErr := handler(poolCtx, mqMsg)
				duration := time.Since(startTime)

				// 记录日志
				ConsumerLogWithDuration(poolCtx, topic, mqMsg, handleErr, duration)

				// 错误处理（未来可以添加重试机制）
				if handleErr != nil {
					Logger().Errorf(poolCtx, "RocketMQ消息处理失败 [%s] - 消息ID: %s, 错误: %+v",
						topic, msg.MsgId, handleErr)
				}
			})
		}
		return consumer.ConsumeSuccess, nil
	})

	if err != nil {
		return gerror.Wrapf(err, "RocketMQ订阅主题失败 [%s]", topic)
	}

	// 启动消费者，只在首次订阅时启动
	if !rocketManager.started {
		if err = r.consumerIns.Start(); err != nil {
			// 失败时取消订阅
			_ = r.consumerIns.Unsubscribe(topic)
			return gerror.Wrapf(err, "RocketMQ启动消费者失败 [%s]", topic)
		}
		rocketManager.started = true
	}

	Logger().Infof(ctx, "RocketMQ消费者订阅成功 [%s]", topic)
	return nil
}

// Close 关闭RocketMQ连接
func (r *RocketMq) Close() error {
	var errors []error

	// 关闭生产者
	if r.producerIns != nil {
		if err := r.producerIns.Shutdown(); err != nil {
			errors = append(errors, gerror.Wrapf(err, "关闭RocketMQ生产者失败"))
		} else {
			Logger().Info(globalCtx, "RocketMQ生产者已关闭")
		}
	}

	// 关闭消费者
	if r.consumerIns != nil {
		if err := r.consumerIns.Shutdown(); err != nil {
			errors = append(errors, gerror.Wrapf(err, "关闭RocketMQ消费者失败"))
		} else {
			Logger().Info(globalCtx, "RocketMQ消费者已关闭")
		}
	}

	// 如果有错误，返回第一个错误
	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

// RegisterRocketMqProducer 注册rocketmq生产者
func RegisterRocketMqProducer() (mqIns *RocketMq, err error) {
	if rocketManager.Producer != nil {
		return rocketManager.Producer, nil
	}

	rocketManager.pMutex.Lock()
	defer rocketManager.pMutex.Unlock()

	if rocketManager.Producer != nil {
		return rocketManager.Producer, nil
	}

	mqIns = new(RocketMq)
	retry := config.Rocketmq.Retry
	if retry <= 0 {
		retry = 0
	}

	mqIns.producerIns, err = rocketmq.NewProducer(
		producer.WithNsResolver(primitive.NewPassthroughResolver(config.Rocketmq.NameSrvAdders)),
		producer.WithRetry(retry),
		producer.WithGroupName(config.GroupName),
		producer.WithCredentials(primitive.Credentials{
			AccessKey: config.Rocketmq.AccessKey,
			SecretKey: config.Rocketmq.SecretKey,
		}),
	)

	if err != nil {
		return nil, err
	}

	if err = mqIns.producerIns.Start(); err != nil {
		return nil, err
	}

	_, err = mqIns.producerIns.SendSync(globalCtx, primitive.NewMessage("ttpos-ping", []byte("1")))
	if err != nil {
		err = gerror.Newf("连通性测试不通过，请检查`queue.rocketmq.nameSrvAdders`或权限配置是否有误。err:%+v", err.Error())
		return nil, err
	}

	SetRLogLevel()
	rocketManager.Producer = mqIns
	return rocketManager.Producer, nil
}

// RegisterRocketMqConsumer 注册rocketmq消费者
func RegisterRocketMqConsumer() (mqIns *RocketMq, err error) {
	rocketManager.cMutex.Lock()
	defer rocketManager.cMutex.Unlock()
	SetRLogLevel()
	if rocketManager.Consumer != nil {
		return rocketManager.Consumer, nil
	}
	// 利用生产者检查一下连通性
	if _, err = RegisterRocketMqProducer(); err != nil {
		return nil, err
	}

	mqIns = new(RocketMq)
	mqIns.consumerIns, err = rocketmq.NewPushConsumer(
		consumer.WithConsumerModel(consumer.Clustering),
		consumer.WithNsResolver(primitive.NewPassthroughResolver(config.Rocketmq.NameSrvAdders)),
		consumer.WithGroupName(config.GroupName),
		consumer.WithCredentials(primitive.Credentials{
			AccessKey: config.Rocketmq.AccessKey,
			SecretKey: config.Rocketmq.SecretKey,
		}),
	)

	if err != nil {
		return nil, err
	}

	// 开多携程处理消费任务，可以根据业务实际情况调整该配置
	rocketManager.goPool = grpool.New(5)

	rocketManager.Consumer = mqIns
	rocketManager.started = false
	return rocketManager.Consumer, nil
}

// SetRLogLevel 设置rocketmq日志输出等级
func SetRLogLevel() {
	level := g.Cfg().MustGet(globalCtx, "queue.rocketmq.logLevel", "warn").String()
	rlog.SetLogLevel(level)
	//rlog.SetOutputPath(g.Cfg().MustGet(ctx, "log.queue.path", "./log/queue.log").String())
}
