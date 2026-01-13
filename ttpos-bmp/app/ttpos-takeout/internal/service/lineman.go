// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	v1 "ttpos-bmp/app/ttpos-takeout/api/lineman/v1"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
)

type (
	ILineman interface {
		// UpdateMenuItemStatus 更新菜单商品状态
		//
		// 参数:
		//   - ctx: 上下文
		//   - merchantId: 商户 ID (Grab MerchantID，对应 Lineman storeId)
		//   - itemId: 商品 ID (partner item id)
		//   - menuStatus: Lineman 状态 (AVAILABLE, SUSPENDED, SOLD_OUT_TODAY)
		//
		// 返回:
		//   - error: 错误信息
		UpdateMenuItemStatus(ctx context.Context, merchantId string, itemId string, menuStatus string) error
		// SyncMenu 同步菜单到 Lineman
		// 参数:
		//   - ctx: 上下文
		//   - shopUUID: 门店UUID
		//
		// 返回:
		//   - error: 错误信息
		SyncMenu(ctx context.Context, shopUUID uint64) error
		// BuildMenuPayload 构建菜单数据
		// 参数:
		//   - ctx: 上下文
		//   - ttposMenuJSON: TTPOS 菜单 JSON 字符串
		//
		// 返回:
		//   - *lineman.MenuSyncRequest: 菜单数据
		//   - error: 错误信息
		BuildMenuPayload(ctx context.Context, ttposMenuJSON string) (*lineman.MenuSyncRequest, error)
	}
	ILinemanOrder interface {
		// HandlePlaceOrder 处理 LINE MAN 提交订单 Webhook
		// 参数验证已由 GoFrame 自动完成，此处只处理业务逻辑
		HandlePlaceOrder(ctx context.Context, req *v1.PlaceOrderReq) error
		// HandleOrderUpdate 处理 LINE MAN 订单更新 Webhook
		// 参数验证已由 GoFrame 自动完成，此处只处理业务逻辑
		HandleOrderUpdate(ctx context.Context, req *v1.OrderUpdateReq) error
		// HandleOrderStatusUpdate 处理 LINE MAN 订单状态更新 Webhook
		// 参数验证已由 GoFrame 自动完成，此处只处理业务逻辑
		HandleOrderStatusUpdate(ctx context.Context, req *v1.OrderStatusUpdateReq) error
	}
)

var (
	localLineman      ILineman
	localLinemanOrder ILinemanOrder
)

func Lineman() ILineman {
	if localLineman == nil {
		panic("implement not found for interface ILineman, forgot register?")
	}
	return localLineman
}

func RegisterLineman(i ILineman) {
	localLineman = i
}

func LinemanOrder() ILinemanOrder {
	if localLinemanOrder == nil {
		panic("implement not found for interface ILinemanOrder, forgot register?")
	}
	return localLinemanOrder
}

func RegisterLinemanOrder(i ILinemanOrder) {
	localLinemanOrder = i
}
