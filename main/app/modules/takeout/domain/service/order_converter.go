package service

import (
	"ttpos-server-go/app/modules/takeout/domain/model"
)

// IOrderConverter 订单转换器接口（用于解耦平台特定的订单解析逻辑）
type IOrderConverter interface {
	// ParseOrderWebhook 解析订单 Webhook 数据
	ParseOrderWebhook(rawData []byte) (interface{}, error)

	// ConvertOrderToTakeoutOrder 将平台订单转换为通用外卖订单格式
	ConvertOrderToTakeoutOrder(
		orderUuid uint64,
		platform string,
		platformOrderId string,
		platformOrder interface{},
		rawDataJSON []byte,
		currentTime int64,
	) (*model.TakeoutOrder, error)
}
