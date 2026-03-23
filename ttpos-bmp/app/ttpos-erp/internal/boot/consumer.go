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
	// ERPNext 文档变更消费者（处理 Webhook 回调，如 SI 外部取消）
	queue.RegisterConsumer(&consumer.DocChangeConsumer{})
	// 原生 DLQ：重试耗尽后 MQ 自动投入 %DLQ%{groupName}
	// 不注册 consumer，死信积压在 broker 中，通过 Dashboard Message 页查看
	// 手动重试走 retry API（POST /callback/retry/sales_invoice）

	queue.StartConsumersListener(ctx)
}
