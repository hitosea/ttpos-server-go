package boot

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/internal/consumer"
	"ttpos-bmp/app/ttpos-erp/internal/consumer/selling"
	"ttpos-bmp/internal/pkg/queue"
)

func InitConsumer(ctx context.Context) {
	// 初始化消费者
	queue.RegisterConsumer(&consumer.ItemSyncConsumer{})
	queue.RegisterConsumer(&selling.SavePosInvoiceConsumer{})
	//注册其他消费者

	queue.StartConsumersListener(ctx)
}
