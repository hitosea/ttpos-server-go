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
