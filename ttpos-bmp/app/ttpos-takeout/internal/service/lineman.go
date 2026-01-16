// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	v1 "ttpos-bmp/app/ttpos-takeout/api/lineman/v1"
	api "ttpos-bmp/app/ttpos-takeout/api/menu"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
)

type (
	ILineman interface {
		// HandlePlaceOrder 处理 LINE MAN 提交订单 Webhook
		// 参数验证已由 GoFrame 自动完成，此处只处理业务逻辑
		HandlePlaceOrder(ctx context.Context, req *v1.PlaceOrderReq) error
		// HandleOrderUpdate 处理 LINE MAN 订单更新 Webhook
		// 参数验证已由 GoFrame 自动完成，此处只处理业务逻辑
		HandleOrderUpdate(ctx context.Context, req *v1.OrderUpdateReq) error
		// HandleOrderStatusUpdate 处理 LINE MAN 订单状态更新 Webhook
		// 参数验证已由 GoFrame 自动完成，此处只处理业务逻辑
		HandleOrderStatusUpdate(ctx context.Context, req *v1.OrderStatusUpdateReq) error
		// UpdateMenuItemStatus 更新菜单商品状态（单个商品）
		//
		// 参数:
		//   - ctx: 上下文
		//   - shopUuid: 店铺 UUID（内部会查询对应的 Lineman storeId）
		//   - itemId: 商品 ID (partner item id)
		//   - menuStatus: Lineman 状态 (AVAILABLE, SUSPENDED, SOLD_OUT_TODAY)
		//
		// 返回:
		//   - error: 错误信息
		UpdateMenuItemStatus(ctx context.Context, shopUuid string, itemId string, menuStatus string) error
		// BatchUpdateMenuItems 批量更新 Lineman 菜单项
		//
		// 参数:
		//   - ctx: 上下文
		//   - req: 批量更新请求（Proto 类型）
		//
		// 返回:
		//   - *api.BatchUpdateMenuResp: 更新结果
		//   - error: 错误信息
		BatchUpdateMenuItems(ctx context.Context, req *api.BatchUpdateMenuReq) (*api.BatchUpdateMenuResp, error)
		// SyncMenu 同步菜单到 Lineman
		// 参数:
		//   - ctx: 上下文
		//   - shopUUID: 门店UUID
		//
		// 返回:
		//   - error: 错误信息
		SyncMenu(ctx context.Context, shopUUID uint64) error
		// BuildMenuPayload 从 TTPOS 菜单 JSON 构建 Lineman 菜单数据
		// 参数:
		//   - ctx: 上下文
		//   - ttposMenuJSON: TTPOS 菜单 JSON 字符串（Grab 格式）
		//
		// 返回:
		//   - *lineman.MenuSyncRequest: 菜单同步请求数据
		//   - error: 错误信息
		BuildMenuPayload(ctx context.Context, ttposMenuJSON string) (*lineman.MenuSyncRequest, error)
		// HandleMenuSyncNotification 处理 LINE MAN 菜单同步通知并写入日志
		// 参数:
		//   - ctx: 上下文
		//   - req: 菜单同步通知请求
		//
		// 返回:
		//   - error: 错误信息
		HandleMenuSyncNotification(ctx context.Context, req *v1.MenuSyncNotificationReq) error
		// UpdateModifierStatus 更新修饰符状态
		//
		// 参数:
		//   - ctx: 上下文
		//   - storeId: 店铺 ID（Lineman storeId）
		//   - modifierId: 修饰符 ID（Partner Modifier ID）
		//   - status: Lineman 状态（1=AVAILABLE, 2=SOLD_OUT_TODAY, 3=SUSPENDED）
		//
		// 返回:
		//   - error: 错误信息
		UpdateModifierStatus(ctx context.Context, storeId string, modifierId string, status int) error
		// HandleTriggerSyncMenu 处理 LINE MAN 触发菜单同步请求
		// 参数:
		//   - ctx: 上下文
		//   - req: 触发同步请求
		//
		// 返回:
		//   - error: 错误信息
		HandleTriggerSyncMenu(ctx context.Context, req *v1.TriggerSyncMenuReq) error
	}
)

var (
	localLineman ILineman
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
