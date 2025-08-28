package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

type RedisMq struct {
	poolName  string
	groupName string
	timeout   int64
}

type RedisOption struct {
	Timeout int64
}

// SendMsg 按字符串类型生产数据
func (r *RedisMq) SendMsg(ctx context.Context, topic string, body string) (mqMsg MqMsg, err error) {
	return r.SendByteMsg(ctx, topic, []byte(body))
}

// SendByteMsg 生产数据
func (r *RedisMq) SendByteMsg(ctx context.Context, topic string, body []byte) (mqMsg MqMsg, err error) {
	// 参数验证
	if r.poolName == "" {
		return mqMsg, gerror.New("Redis消息队列未初始化")
	}
	if topic == "" {
		return mqMsg, gerror.New("主题名称不能为空")
	}
	if len(body) == 0 {
		return mqMsg, gerror.New("消息内容不能为空")
	}

	// 构建消息对象
	mqMsg = MqMsg{
		RunType:   MsgTypeSend,
		Topic:     topic,
		MsgId:     getRandMsgId(),
		Body:      body,
		Timestamp: time.Now(),
	}

	// 序列化消息
	data, err := json.Marshal(mqMsg)
	if err != nil {
		return mqMsg, gerror.Wrapf(err, "Redis消息序列化失败")
	}

	// 发送消息到Redis
	startTime := time.Now()
	key := r.genKey(r.groupName, topic)
	if _, err = g.Redis().Do(ctx, "LPUSH", key, data); err != nil {
		return mqMsg, gerror.Wrapf(err, "Redis发送消息失赅 [%s]", topic)
	}

	// 设置过期时间
	if r.timeout > 0 {
		if _, err = g.Redis().Do(ctx, "EXPIRE", key, r.timeout); err != nil {
			Logger().Warningf(ctx, "Redis设置过期时间失败 [%s]: %v", topic, err)
			// 设置过期失败不影响消息发送
			err = nil
		}
	}

	// 性能监控
	duration := time.Since(startTime)
	if duration > 100*time.Millisecond {
		Logger().Warningf(ctx, "Redis发送消息耗时 [%s] - 消息ID: %s, 耗时: %v",
			topic, mqMsg.MsgId, duration)
	}

	return mqMsg, nil
}

// SendDelayMsg 发送延时消息
func (r *RedisMq) SendDelayMsg(ctx context.Context, topic string, body string, delay time.Duration) (mqMsg MqMsg, err error) {
	// 参数验证
	delaySecond := int64(delay.Seconds())
	if delaySecond < 1 {
		return r.SendMsg(ctx, topic, body)
	}

	if r.poolName == "" {
		return mqMsg, gerror.New("Redis消息队列未初始化")
	}
	if topic == "" {
		return mqMsg, gerror.New("主题名称不能为空")
	}
	if body == "" {
		return mqMsg, gerror.New("消息内容不能为空")
	}

	// 构建消息对象
	mqMsg = MqMsg{
		RunType:   MsgTypeSend,
		Topic:     topic,
		MsgId:     getRandMsgId(),
		Body:      []byte(body),
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(mqMsg)
	if err != nil {
		return
	}

	var (
		conn         = g.Redis()
		key          = r.genKey(r.groupName, "delay:"+topic)
		expireSecond = time.Now().Unix() + delaySecond
		timePiece    = fmt.Sprintf("%s:%d", key, expireSecond)
		z            = gredis.ZAddMember{Score: float64(expireSecond), Member: timePiece}
	)

	if _, err = conn.ZAdd(ctx, key, &gredis.ZAddOption{}, z); err != nil {
		return
	}

	if _, err = conn.RPush(ctx, timePiece, data); err != nil {
		return
	}

	// consumer will also delete the item
	if r.timeout > 0 {
		_, _ = conn.Expire(ctx, timePiece, r.timeout+delaySecond)
		_, _ = conn.Expire(ctx, key, r.timeout)
	}
	return
}

// Subscribe 订阅主题消息（新接口方法）
func (r *RedisMq) Subscribe(ctx context.Context, topic string, handler func(ctx context.Context, mqMsg MqMsg) error) error {
	// 参数验证
	if r.poolName == "" {
		return gerror.New("Redis消息队列未初始化")
	}
	if topic == "" {
		return gerror.New("主题名称不能为空")
	}
	if handler == nil {
		return gerror.New("消息处理函数不能为空")
	}

	var (
		key      = r.genKey(r.groupName, topic)
		delayKey = r.genKey(r.groupName, "delay:"+topic)
	)

	Logger().Infof(ctx, "Redis消费者开始监听 [%s]", topic)

	// 启动普通消息消费协程
	go func() {
		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				Logger().Infof(ctx, "Redis普通消息消费已停止 [%s]", topic)
				return
			case <-ticker.C:
				mqMsgList := r.loopReadQueue(ctx, key)
				for _, mqMsg := range mqMsgList {
					r.handleMessage(ctx, topic, mqMsg, handler)
				}
			}
		}
	}()

	// 启动延时消息消费协程
	go func() {
		defer func() {
			Logger().Infof(ctx, "Redis延时消息消费已停止 [%s]", topic)
		}()

		mqMsgCh, errCh := r.loopReadDelayQueue(ctx, delayKey)
		for {
			select {
			case <-ctx.Done():
				return
			case mqMsg, ok := <-mqMsgCh:
				if !ok {
					return
				}
				r.handleMessage(ctx, topic, mqMsg, handler)
			case err, ok := <-errCh:
				if !ok {
					return
				}
				if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
					Logger().Warningf(ctx, "Redis延时消息处理错误 [%s]: %+v", topic, err)
				}
			}
		}
	}()

	return nil
}

// handleMessage 处理单个消息
func (r *RedisMq) handleMessage(ctx context.Context, topic string, mqMsg MqMsg, handler func(ctx context.Context, mqMsg MqMsg) error) {
	startTime := time.Now()

	// 处理消息
	handleErr := handler(ctx, mqMsg)
	duration := time.Since(startTime)

	// 记录日志
	ConsumerLogWithDuration(ctx, topic, mqMsg, handleErr, duration)
}

// 生成队列key
func (r *RedisMq) genKey(groupName string, topic string) string {
	return fmt.Sprintf("queue:%s_%s", groupName, topic)
}

// loopReadQueue 循环读取普通队列消息
func (r *RedisMq) loopReadQueue(ctx context.Context, key string) (mqMsgList []MqMsg) {
	conn := g.Redis()
	for {
		data, err := conn.Do(ctx, "RPOP", key)
		if err != nil {
			Logger().Debugf(ctx, "Redis读取队列消息失败 [%s]: %+v", key, err)
			break
		}

		if data.IsEmpty() {
			break
		}

		var mqMsg MqMsg
		if err = data.Scan(&mqMsg); err != nil {
			Logger().Warningf(ctx, "Redis消息反序列化失败 [%s]: %+v", key, err)
			break
		}

		// 验证消息有效性
		if mqMsg.IsValid() {
			mqMsg.RunType = MsgTypeReceive // 设置为接收类型
			mqMsgList = append(mqMsgList, mqMsg)
		}
	}
	return mqMsgList
}

func RegisterRedisMqProducer(connOpt RedisOption, groupName string) (client MqProducer) {
	return RegisterRedisMq(connOpt, groupName)
}

// RegisterRedisMqConsumer 注册消费者
func RegisterRedisMqConsumer(connOpt RedisOption, groupName string) (client MqConsumer) {
	return RegisterRedisMq(connOpt, groupName)
}

// RegisterRedisMq 注册redis实例
func RegisterRedisMq(connOpt RedisOption, groupName string) *RedisMq {
	return &RedisMq{
		poolName:  gmd5.MustEncrypt(fmt.Sprintf("%s-%d", groupName, time.Now().UnixNano())),
		groupName: groupName,
		timeout:   connOpt.Timeout,
	}
}

func getRandMsgId() string {
	rand.NewSource(time.Now().UnixNano())
	radium := rand.Intn(999) + 1
	timeCode := time.Now().UnixNano()
	return fmt.Sprintf("%d%.4d", timeCode, radium)
}

// loopReadDelayQueue 循环读取延时队列消息
func (r *RedisMq) loopReadDelayQueue(ctx context.Context, key string) (resCh chan MqMsg, errCh chan error) {
	resCh = make(chan MqMsg)
	errCh = make(chan error, 1)

	go func() {
		defer close(resCh)
		defer close(errCh)

		conn := g.Redis()
		for {
			now := time.Now().Unix()
			do, err := conn.Do(ctx, "zrangebyscore", key, "0", strconv.FormatInt(now, 10), "limit", 0, 1)
			if err != nil {
				return
			}

			val := do.Strings()
			if len(val) == 0 {
				select {
				case <-ctx.Done():
					errCh <- ctx.Err()
					return
				case <-time.After(time.Second):
					continue
				}
			}
			for _, listK := range val {
				for {
					pop, err := conn.LPop(ctx, listK)
					if err != nil {
						errCh <- err
						return
					} else if pop.IsEmpty() {
						_, _ = conn.ZRem(ctx, key, listK)
						_, _ = conn.Del(ctx, listK)
						break
					} else {
						var mqMsg MqMsg
						if err = pop.Scan(&mqMsg); err != nil {
							g.Log().Warningf(ctx, "loopReadDelayQueue Scan err:%+v", err)
							break
						}

						if mqMsg.MsgId == "" {
							continue
						}

						// 设置消息类型为接收
						mqMsg.RunType = MsgTypeReceive

						select {
						case resCh <- mqMsg:
						case <-ctx.Done():
							errCh <- ctx.Err()
							return
						}
					}
				}
			}
		}
	}()
	return resCh, errCh
}

// Close 关闭Redis连接（Redis连接由gf框架管理，这里只是记录日志）
func (r *RedisMq) Close() error {
	Logger().Info(globalCtx, "Redis消息队列连接已关闭")
	return nil
}
