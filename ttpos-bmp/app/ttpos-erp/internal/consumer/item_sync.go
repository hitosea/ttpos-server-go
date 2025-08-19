package consumer

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/internal/pkg/queue"
)

type ItemSyncConsumer struct {
}

// GetTopic 获取消费主题，返回商品同步的主题名称
func (c *ItemSyncConsumer) GetTopic() string {
	// 这里返回商品同步的队列主题
	return string(consts.TopicItemSync)
}

// Handle 处理消息的方法
// ctx: 上下文
// mqMsg: 队列消息
// 返回 error，如处理失败可返回错误
func (c *ItemSyncConsumer) Handle(ctx context.Context, mqMsg queue.MqMsg) (err error) {
	// 这里编写商品同步的具体处理逻辑
	// 可以根据 mqMsg.Body 解析消息内容并进行相应的业务处理
	// 示例：打印消息内容
	// 实际业务中请替换为同步商品的具体实现
	// 中文注释：处理商品同步队列消息
	g.Log().Info(ctx, "收到商品同步消息：", string(mqMsg.Body))
	return nil
}
