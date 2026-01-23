// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"time"

	grabfood "github.com/grab/grabfood-api-sdk-go"

	api "ttpos-bmp/app/ttpos-takeout/api/menu"
	orderApi "ttpos-bmp/app/ttpos-takeout/api/order"
	grabDto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/model/entity"
)

type (
	IGrab interface {
		// AcceptOrder 接受订单
		AcceptOrder(ctx context.Context, orderID string) error
		// RejectOrder 拒绝订单
		RejectOrder(ctx context.Context, orderID string, rejectCode int) error
		// CancelOrder 取消订单
		CancelOrder(ctx context.Context, merchantID string, orderID string, cancelCode string) error
		// MarkOrderReady 标记订单准备完成
		MarkOrderReady(ctx context.Context, orderID string, markStatus string) error
		// UpdateDeliveryState 更新配送状态 (自配送)
		UpdateDeliveryState(ctx context.Context, orderID string, fromState string, toState string) error
		// UpdateOrderReadyTime 更新订单准备时间
		UpdateOrderReadyTime(ctx context.Context, orderID string, newReadyTime time.Time) error
		// CheckOrderCancelable 检查订单是否可取消
		// 返回 Grab SDK 的完整响应对象
		CheckOrderCancelable(ctx context.Context, merchantID string, orderID string) (*grabfood.CheckOrderCancelableResponse, error)
		// ListOrders 查询 Grab 订单列表
		// 封装 GrabFood SDK ListOrdersAPI，支持按商户、日期、订单ID等维度查询
		// 注意：此接口在 Grab 测试环境下不可用，仅生产环境支持
		// 参数：
		//   - merchantID: Grab 商户 ID（必填）
		//   - date: 日期过滤，格式 YYYY-MM-DD（可选）
		//   - orderIDs: 订单 ID 列表过滤（可选）
		//   - page: 分页页码（可选）
		// 返回：
		//   - resp: ListOrdersResponse（使用 TakeoutOrder 以获取完整 Price 含 Total）
		//   - err: 错误信息
		ListOrders(ctx context.Context, merchantID string, date string, orderIDs []string, page int32) (*grabDto.ListOrdersResponse, error)
		// PauseStore 暂停门店
		PauseStore(ctx context.Context, merchantID string, duration int) error
		// ResumeStore 恢复门店营业
		ResumeStore(ctx context.Context, merchantID string) error
		// GetStoreStatus 获取门店状态
		GetStoreStatus(ctx context.Context, merchantID string) (*grabfood.StoreStatusResponse, error)
		// GetStoreHours 获取门店营业时间
		GetStoreHours(ctx context.Context, merchantID string) (*grabfood.StoreHourResponse, error)
		// NotifyMenuUpdate 通知 Grab 菜单已更新
		// 返回 requestID 用于追踪同步状态
		// 此方法实现 MenuNotifier 接口
		NotifyMenuUpdate(ctx context.Context, merchantID string) (string, error)
		// TraceMenuSync 追踪菜单同步状态
		TraceMenuSync(ctx context.Context, merchantID string) (*grabfood.MenuSyncResponse, error)
		// UpdateMenuRecord 更新单个菜单记录 (商品或修饰符)
		// 调用 GrabFood API PUT /partner/v1/merchants/menu/record
		// req 可以是 UpdateMenuItem 或 UpdateMenuModifier
		UpdateMenuRecord(ctx context.Context, merchantID string, req grabfood.UpdateMenuRequest) error
		// BatchUpdateMenu 批量更新菜单记录 (商品或修饰符)
		// 调用 GrabFood API POST /partner/v1/batch/menu
		// 参数：
		//   - ctx: 上下文对象
		//   - merchantID: Grab Merchant ID
		//   - req: 批量更新请求（GrabFood SDK 结构体）
		//
		// 返回：
		//   - resp: 批量更新响应，包含状态和错误列表
		//   - err: 错误信息
		BatchUpdateMenu(ctx context.Context, merchantID string, req *grabfood.BatchUpdateMenuItem) (*grabfood.BatchUpdateMenuResponse, error)
		// CreateSelfServeJourney 创建自助激活链接
		// merchantID: Grab Merchant ID (shop_uuid)
		// 返回: activation_url, request_id
		CreateSelfServeJourney(ctx context.Context, merchantID string) (string, string, error)
		// SaveMenuSnapshot 保存菜单快照到数据库
		// 使用 shop_uuid + provider_name 作为唯一键，存在则更新，不存在则插入
		SaveMenuSnapshot(ctx context.Context, dto *grabDto.PushGrabMenuDTO) (uint64, error)
		// NotifyMenuUpdate 发送菜单更新通知 (RocketMQ)
		NotifyMenuUpdateEvent(ctx context.Context, event *grabDto.ProviderMenuUpdateEvent) error
		// UpdateMenuItem 更新单个菜单项 (商品)
		// 调用 GrabFood API PUT /partner/v1/merchants/menu/record 更新商品信息
		// 支持更新：价格、可用状态、库存、高级定价配置、购买能力配置
		UpdateMenuItem(ctx context.Context, req *grabDto.UpdateMenuItemReq) error
		// UpdateMenuModifier 更新单个修饰符
		// 调用 GrabFood API PUT /partner/v1/merchants/menu/record 更新修饰符信息
		// 支持更新：价格、可用状态、是否免费、高级定价配置
		UpdateMenuModifier(ctx context.Context, req *grabDto.UpdateMenuModifierReq) error
		// BatchUpdateMenu 批量更新菜单记录 (商品或修饰符)
		// 调用 GrabFood API POST /partner/v1/batch/menu 批量更新菜单信息
		// 支持批量更新：价格、可用状态、库存、高级定价配置、购买能力配置
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 批量更新请求
		//
		// 返回：
		//   - resp: 批量更新响应，包含状态和错误列表
		//   - err: 错误信息
		BatchUpdateMenuItems(ctx context.Context, req *api.BatchUpdateMenuReq) (*api.BatchUpdateMenuResp, error)
		// HandleSubmitOrder 处理 Grab 提交订单 Webhook
		// 签名验证已由中间件完成，此处只处理业务逻辑
		// 使用 SDK grabfood.SubmitOrderRequest 替换自定义 DTO
		HandleSubmitOrder(ctx context.Context, req *grabfood.SubmitOrderRequest) error
		// HandlePushOrderState 处理订单状态变更 Webhook
		// 签名验证已由中间件完成，此处只处理业务逻辑
		// 使用 SDK grabfood.OrderStateRequest
		HandlePushOrderState(ctx context.Context, req *grabfood.OrderStateRequest) error
		// PrepareOrder 准备订单（接受/拒绝）
		// 参数：
		//   - ctx: 上下文对象
		//   - orderEntity: 订单实体
		//   - toState: 目标状态 (Accepted/Rejected)
		//
		// 返回：
		//   - err: 错误信息
		PrepareOrder(ctx context.Context, orderEntityInterface interface{}, toState string) error
		// MarkOrderReady 标记订单准备完成
		// 参数：
		//   - ctx: 上下文对象
		//   - orderEntity: 订单实体，包含 ProviderOrderId 等信息
		//
		// 返回：
		//   - err: 错误信息
		MarkOrderReadyEntity(ctx context.Context, orderEntity *entity.Order) error
		// CheckOrderCancelable 检查订单是否可取消
		// 参数：
		//   - ctx: 上下文对象
		//   - orderEntity: 订单实体
		//
		// 返回：
		//   - res: 检查订单可取消性响应
		//   - err: 错误信息
		CheckOrderCancelableEntity(ctx context.Context, orderEntity *entity.Order) (*orderApi.CheckOrderCancelableResp, error)
		// CancelOrder 取消订单
		// 参数：
		//   - ctx: 上下文对象
		//   - orderEntity: 订单实体
		//   - cancelCode: 取消原因码（字符串格式，可根据不同平台传入不同的编码）
		//
		// 返回：
		//   - res: 取消订单响应
		//   - err: 错误信息
		CancelOrderEntity(ctx context.Context, orderEntity *entity.Order, cancelCode string) (res *orderApi.CancelOrderResp, err error)
		// GeneratePartnerToken 根据 client_id / client_secret 生成访问 Token
		// 采用 JWT（HS256）实现
		// 参数：
		//   - ctx: 上下文
		//   - clientID: Grab 分配的 client_id
		//   - clientSecret: Grab 分配的 client_secret，用于校验身份
		//   - scope: 请求的权限范围
		//
		// 返回：
		//   - token: 生成的 JWT Token
		//   - expiresIn: Token 有效期（秒）
		//   - err: 错误信息
		GetPartnerToken(ctx context.Context, clientID string, clientSecret string, scope string) (token string, expiresIn int, err error)
		// HandleGetMenu 处理 Grab 获取菜单请求 (Partner Endpoint)
		// 签名验证已由中间件完成
		HandleGetMenu(ctx context.Context, merchantID string) (*grabfood.GetMenuNewResponse, error)
		// HandlePushGrabMenu 处理 Grab 菜单推送 Webhook
		// 此方法组合了 SaveMenuSnapshot 和 NotifyMenuUpdate 两个操作
		HandlePushGrabMenu(ctx context.Context, dto *grabDto.PushGrabMenuDTO) error
		// HandleMenuSyncState 处理菜单同步状态回调
		// 使用 SDK grabfood.MenuSyncWebhookRequest
		HandleMenuSyncState(ctx context.Context, req *grabfood.MenuSyncWebhookRequest) error
	}
)

var (
	localGrab IGrab
)

func Grab() IGrab {
	if localGrab == nil {
		panic("implement not found for interface IGrab, forgot register?")
	}
	return localGrab
}

func RegisterGrab(i IGrab) {
	localGrab = i
}
