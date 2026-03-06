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
	queue.RegisterConsumer(&selling.ReturnPosInvoiceConsumer{})
	queue.RegisterConsumer(&selling.CancelPosInvoice{})
	queue.RegisterConsumer(&selling.ClosePosEntryConsumer{})
	queue.RegisterConsumer(&selling.RedoPosConsumer{})
	// Sales Invoice 消费者
	queue.RegisterConsumer(&selling.SaveSalesInvoiceConsumer{})
	queue.RegisterConsumer(&selling.CancelSalesInvoiceConsumer{})
	queue.RegisterConsumer(&selling.ReturnSalesInvoiceConsumer{})
	queue.RegisterConsumer(&selling.SalesInvoiceCallbackConsumer{})

	queue.StartConsumersListener(ctx)
}
