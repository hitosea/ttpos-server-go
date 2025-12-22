package order

import (
	"context"

	api "ttpos-bmp/app/ttpos-takeout/api/order"
	"ttpos-bmp/app/ttpos-takeout/internal/dao"
	"ttpos-bmp/app/ttpos-takeout/internal/model/entity"
	"ttpos-bmp/app/ttpos-takeout/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
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

	res = &api.GetOrderInfoResp{
		ShopUuid:     orderEntity.ShopUuid,
		OrderStatus:  orderEntity.OrderStatus,
		OrderType:    orderEntity.OrderType,
		RawData:      orderEntity.RawData,
		ProviderName: orderEntity.ProviderName,
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
	case "grab":
		// 调用 Grab 订单处理逻辑
		err = service.GrabOrder().PrepareOrder(ctx, orderEntity, req.ToState)
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
