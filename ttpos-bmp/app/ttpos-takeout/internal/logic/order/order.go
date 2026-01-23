package order

import (
	"context"
	"time"

	linemanv1 "ttpos-bmp/app/ttpos-takeout/api/lineman/v1"
	api "ttpos-bmp/app/ttpos-takeout/api/order"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/dao"
	"ttpos-bmp/app/ttpos-takeout/internal/model/entity"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
	"ttpos-bmp/app/ttpos-takeout/utility"

	"ttpos-api/ttpos-takeout/message"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	grabsdk "github.com/grab/grabfood-api-sdk-go"
)

type sOrder struct{}

func New() *sOrder {
	return &sOrder{}
}

func init() {
	service.RegisterOrder(New())
}

// GetOrderInfo 获取订单信息
// 参数：
//   - ctx: 上下文对象
//   - req: 获取订单信息请求，包含 shop_uuid 和 order_uuid
//
// 返回：
//   - res: 订单信息响应
//   - err: 错误信息
func (s *sOrder) GetOrderInfo(ctx context.Context, req *api.GetOrderInfoReq) (res *api.GetOrderInfoResp, err error) {
	if req.RequestId != "" {
		g.Log().Infof(ctx, "GetOrderInfo start, requestId: %s, shopUuid: %s, orderUuid: %s", req.RequestId, req.ShopUuid, req.OrderUuid)
	}

	// 直接使用 DAO 链式调用查询订单
	var orderEntity *entity.Order
	err = dao.Order.Ctx(ctx).
		Where(dao.Order.Columns().ShopUuid, req.ShopUuid).
		Where(dao.Order.Columns().Uuid, req.OrderUuid).
		Scan(&orderEntity)
	if err != nil {
		return nil, gerror.Wrap(err, "查询订单失败")
	}

	if orderEntity == nil {
		return nil, gerror.New("订单不存在")
	}

	// 转换订单数据为统一格式（TakeoutOrder）
	var orderData string
	var takeoutOrder *message.TakeoutOrder
	startTime := time.Now()

	switch orderEntity.ProviderName {
	case string(consts.ProviderGrab):
		// Grab 订单：反序列化 raw_data 并转换为统一模型
		var grabReq grabsdk.SubmitOrderRequest
		if unmarshalErr := gjson.DecodeTo(orderEntity.RawData, &grabReq); unmarshalErr != nil {
			g.Log().Errorf(ctx, "Grab订单数据反序列化失败, orderId=%s, error=%v",
				req.OrderUuid, unmarshalErr)
		} else {
			//特殊处理 price.total
			takeoutOrder, err = utility.FromGrabSDK(&grabReq)
		}
	case string(consts.ProviderLineman):
		// Lineman 订单：反序列化 raw_data 并转换为统一模型
		var linemanReq linemanv1.PlaceOrderReq
		if unmarshalErr := gjson.DecodeTo(orderEntity.RawData, &linemanReq); unmarshalErr != nil {
			g.Log().Errorf(ctx, "Lineman订单数据反序列化失败, orderId=%s, error=%v",
				req.OrderUuid, unmarshalErr)
		} else {
			takeoutOrder, err = utility.FromLinemanPlaceOrder(&linemanReq)
		}
	default:
		g.Log().Warningf(ctx, "不支持的外卖平台: %s, orderId=%s", orderEntity.ProviderName, req.OrderUuid)
	}

	if err != nil {
		// 转换失败，记录错误日志但不影响接口返回（优雅降级）
		g.Log().Errorf(ctx, "订单数据转换失败, orderId=%s, provider=%s, error=%v",
			req.OrderUuid, orderEntity.ProviderName, err)
		orderData = "" // 返回空字符串
	} else if takeoutOrder != nil {
		// 序列化为 JSON 字符串
		orderData, err = gjson.New(takeoutOrder).ToJsonString()
		if err != nil {
			g.Log().Errorf(ctx, "订单数据序列化失败, orderId=%s, error=%v",
				req.OrderUuid, err)
			orderData = ""
		} else {
			// 记录转换耗时（用于性能监控）
			duration := time.Since(startTime).Milliseconds()
			g.Log().Debugf(ctx, "订单数据转换完成, orderId=%s, provider=%s, duration=%dms",
				req.OrderUuid, orderEntity.ProviderName, duration)
		}
	}

	res = &api.GetOrderInfoResp{
		ShopUuid:     orderEntity.ShopUuid,
		OrderStatus:  orderEntity.OrderStatus,
		OrderType:    orderEntity.OrderType,
		RawData:      orderEntity.RawData,
		ProviderName: orderEntity.ProviderName,
		OrderData:    orderData, // 转换后的统一订单数据（TakeoutOrder JSON）
	}

	return res, nil
}

// PrepareOrder 准备订单（接受/拒绝）
// 参数：
//   - ctx: 上下文对象
//   - req: 准备订单请求，包含 order_uuid 和 to_state
//
// 返回：
//   - res: 准备订单响应
//   - err: 错误信息
func (s *sOrder) PrepareOrder(ctx context.Context, req *api.PrepareOrderReq) (res *api.PrepareOrderResp, err error) {
	if req.RequestId != "" {
		g.Log().Infof(ctx, "开始准备订单, requestId: %s, orderUuid: %s, toState: %s", req.RequestId, req.TakeoutOrderUuid, req.ToState)
	}

	// 查询订单信息
	var orderEntity *entity.Order
	err = dao.Order.Ctx(ctx).
		Where(dao.Order.Columns().Uuid, req.TakeoutOrderUuid).
		Scan(&orderEntity)
	if err != nil {
		return nil, gerror.Wrap(err, "查询订单失败")
	}

	if orderEntity == nil {
		return nil, gerror.New("订单不存在")
	}

	// 根据 provider_name 路由到不同平台的处理逻辑
	switch orderEntity.ProviderName {
	case string(consts.ProviderGrab):
		// 调用 Grab 订单处理逻辑
		err = service.Grab().PrepareOrder(ctx, orderEntity, req.ToState)
		if err != nil {
			return nil, gerror.Wrap(err, "Grab订单准备失败")
		}
	default:
		return nil, gerror.Newf("不支持的平台: %s", orderEntity.ProviderName)
	}

	res = &api.PrepareOrderResp{
		OrderUuid: orderEntity.Uuid,
	}

	g.Log().Infof(ctx, "订单准备成功, orderUuid: %s, toState: %s", req.TakeoutOrderUuid, req.ToState)
	return res, nil
}

// MarkOrderReady 标记订单准备完成
// 参数：
//   - ctx: 上下文对象
//   - takeoutOrderUuid: 外卖订单 UUID
//   - requestId: 请求追踪 ID（可选）
//
// 返回：
//   - orderUuid: 订单 UUID
//   - err: 错误信息
func (s *sOrder) MarkOrderReady(ctx context.Context, takeoutOrderUuid string, requestId string) (orderUuid string, err error) {
	if requestId != "" {
		g.Log().Infof(ctx, "开始标记订单准备完成, requestId: %s, orderUuid: %s", requestId, takeoutOrderUuid)
	}

	// 查询订单信息
	var orderEntity *entity.Order
	err = dao.Order.Ctx(ctx).
		Where(dao.Order.Columns().Uuid, takeoutOrderUuid).
		Scan(&orderEntity)
	if err != nil {
		return "", gerror.Wrap(err, "查询订单失败")
	}

	if orderEntity == nil {
		return "", gerror.New("订单不存在")
	}

	// 根据 provider_name 路由到不同平台的处理逻辑
	switch orderEntity.ProviderName {
	case string(consts.ProviderGrab):
		// 调用 Grab 订单处理逻辑
		err = service.Grab().MarkOrderReadyEntity(ctx, orderEntity)
		if err != nil {
			return "", gerror.Wrap(err, "Grab订单标记准备完成失败")
		}
	default:
		return "", gerror.Newf("不支持的平台: %s", orderEntity.ProviderName)
	}

	g.Log().Infof(ctx, "订单标记准备完成成功, orderUuid: %s", takeoutOrderUuid)
	return orderEntity.Uuid, nil
}

// CancelOrder 取消订单
// 参数：
//   - ctx: 上下文对象
//   - req: 取消订单请求，包含 order_uuid 和 cancel_code
//
// 返回：
//   - res: 取消订单响应
//   - err: 错误信息
func (s *sOrder) CancelOrder(ctx context.Context, req *api.CancelOrderReq) (res *api.CancelOrderResp, err error) {
	if req.RequestId != "" {
		g.Log().Infof(ctx, "开始取消订单, requestId: %s, orderUuid: %s, cancelCode: %s", req.RequestId, req.TakeoutOrderUuid, req.CancelCode)
	}

	// 参数验证
	if req.TakeoutOrderUuid == "" {
		return nil, gerror.New("订单UUID不能为空")
	}

	// 查询订单信息
	var orderEntity *entity.Order
	err = dao.Order.Ctx(ctx).
		Where(dao.Order.Columns().Uuid, req.TakeoutOrderUuid).
		Scan(&orderEntity)
	if err != nil {
		return nil, gerror.Wrap(err, "查询订单失败")
	}

	if orderEntity == nil {
		return nil, gerror.New("订单不存在")
	}

	// 根据 provider_name 路由到不同平台的处理逻辑
	switch orderEntity.ProviderName {
	case string(consts.ProviderGrab):
		// 调用 Grab 订单处理逻辑
		res, err = service.Grab().CancelOrderEntity(ctx, orderEntity, req.CancelCode)
		if err != nil {
			return nil, gerror.Wrap(err, "Grab订单取消失败")
		}
	default:
		return nil, gerror.Newf("不支持的平台: %s", orderEntity.ProviderName)
	}

	g.Log().Infof(ctx, "订单取消处理完成, orderUuid: %s", req.TakeoutOrderUuid)
	return res, nil
}

// CheckOrderCancelable 检查订单是否可取消
// 参数：
//   - ctx: 上下文对象
//   - req: 检查订单可取消性请求，包含 order_uuid
//
// 返回：
//   - res: 检查订单可取消性响应
//   - err: 错误信息
func (s *sOrder) CheckOrderCancelable(ctx context.Context, req *api.CheckOrderCancelableReq) (res *api.CheckOrderCancelableResp, err error) {
	if req.RequestId != "" {
		g.Log().Infof(ctx, "开始检查订单可取消性, requestId: %s, orderUuid: %s", req.RequestId, req.TakeoutOrderUuid)
	}

	// 参数验证
	if req.TakeoutOrderUuid == "" {
		return nil, gerror.New("订单UUID不能为空")
	}

	// 查询订单信息
	var orderEntity *entity.Order
	err = dao.Order.Ctx(ctx).
		Where(dao.Order.Columns().Uuid, req.TakeoutOrderUuid).
		Scan(&orderEntity)
	if err != nil {
		return nil, gerror.Wrap(err, "查询订单失败")
	}

	if orderEntity == nil {
		return nil, gerror.New("订单不存在")
	}

	// 根据 provider_name 路由到不同平台的处理逻辑
	switch orderEntity.ProviderName {
	case string(consts.ProviderGrab):
		// 调用 Grab 订单处理逻辑
		res, err = service.Grab().CheckOrderCancelableEntity(ctx, orderEntity)
		if err != nil {
			return nil, gerror.Wrap(err, "Grab订单检查可取消性失败")
		}
	default:
		return nil, gerror.Newf("不支持的平台: %s", orderEntity.ProviderName)
	}

	g.Log().Infof(ctx, "订单可取消性检查完成, orderUuid: %s, canCancel: %t", req.TakeoutOrderUuid, res.CanCancel)
	return res, nil
}
