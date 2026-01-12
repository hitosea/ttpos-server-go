// Package lineman 提供 LINE MAN 订单服务的业务逻辑
package lineman

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/guid"

	v1 "ttpos-bmp/app/ttpos-takeout/api/lineman/v1"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/dao"
	"ttpos-bmp/app/ttpos-takeout/internal/model/do"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
	"ttpos-bmp/internal/pkg/queue"
)

const (
	// TopicLinemanOrder LINE MAN 订单 MQ Topic（复用 Grab 的 topic）
	TopicLinemanOrder = "takeout_grab_order"
	// ProviderNameLineman LINE MAN 供应商名称常量
	ProviderNameLineman = string(consts.ProviderLineman) // "lineman"
)

// sLinemanOrder 订单服务
type sLinemanOrder struct{}

func init() {
	service.RegisterLinemanOrder(&sLinemanOrder{})
}

// HandlePlaceOrder 处理 LINE MAN 提交订单 Webhook
// 参数验证已由 GoFrame 自动完成，此处只处理业务逻辑
func (s *sLinemanOrder) HandlePlaceOrder(ctx context.Context, req *v1.PlaceOrderReq) error {
	// 1. 检查订单是否已存在（防止重复推送 - HTTP 409）
	existingOrder, _ := dao.Order.Ctx(ctx).
		Where(dao.Order.Columns().ProviderName, ProviderNameLineman).
		Where(dao.Order.Columns().ProviderOrderId, req.OrderId).
		One()
	if !existingOrder.IsEmpty() {
		g.Log().Warningf(ctx, "订单已存在: orderId=%s", req.OrderId)
		return gerror.New("Order ID already exists")
	}

	// 2. 保存订单
	orderUUID, err := s.saveOrder(ctx, req)
	if err != nil {
		g.Log().Errorf(ctx, "保存订单失败: %v", err)
		return gerror.Wrap(err, "保存订单失败")
	}

	// 3. 发送 MQ 消息
	event := &grab.OrderEvent{
		Action:       string(consts.OrderActionCreate),
		ProviderName: ProviderNameLineman,
		ShopUUID:     req.StoreId, // 使用 storeId 作为 shopUuid
		OrderUUID:    orderUUID,
		OrderID:      req.OrderId,
		MerchantID:   req.StoreId,
		Status:       string(consts.OrderStatusAccepted),
		Timestamp:    gtime.Now().Unix(),
	}
	if err := queue.PushWithContext(ctx, TopicLinemanOrder, event); err != nil {
		// MQ 发送失败只记录日志，不影响主流程（订单已入库）
		g.Log().Warningf(ctx, "发送订单 MQ 事件失败 %s: %v", orderUUID, err)
	}

	g.Log().Infof(ctx, "成功处理 LINE MAN 订单: %s (UUID: %s)", req.OrderId, orderUUID)
	return nil
}

// saveOrder 保存订单到数据库
func (s *sLinemanOrder) saveOrder(ctx context.Context, req *v1.PlaceOrderReq) (string, error) {
	orderUUID := guid.S()

	// 查询 shop_uuid（通过 storeId 查询门店配置）
	shopUuid := req.StoreId // 默认使用 storeId
	// cfg, err := service.ShopProviderCfg().GetShopProviderCfgByMerchantID(ctx, req.StoreId, ProviderNameLineman)
	// if err != nil {
	// 	g.Log().Warningf(ctx, "查询门店配置失败: storeId=%s, error=%v", req.StoreId, err)
	// } else if cfg != nil {
	// 	shopUuid = strconv.FormatUint(cfg.ShopUuid, 10)
	// 	g.Log().Infof(ctx, "找到门店配置: storeId=%s, shopUuid=%s", req.StoreId, shopUuid)
	// } else {
	// 	g.Log().Warningf(ctx, "未找到门店配置: storeId=%s, provider=%s", req.StoreId, ProviderNameLineman)
	// }

	// 转换订单类型
	orderType := string(consts.OrderTypeDeliveryByLineman) // 默认外送
	if req.CustomerType == string(consts.LinemanCustomerTypePickup) {
		orderType = string(consts.OrderTypeTakeAway)
	}

	// 解析订单时间（ISO 8601 → Unix 时间戳）
	orderTime, err := gtime.StrToTime(req.OrderAcceptedTime)
	if err != nil {
		g.Log().Warningf(ctx, "解析订单时间失败: %v, 使用当前时间", err)
		orderTime = gtime.Now()
	}

	// 序列化原始数据（完整保存用于问题排查）
	rawData, err := gjson.EncodeString(req)
	if err != nil {
		g.Log().Errorf(ctx, "序列化请求失败: %v", err)
		return "", gerror.Wrap(err, "序列化请求失败")
	}

	// 序列化 additionalItems（附加项列表）
	// additionalItemsJSON := ""
	// if len(req.AdditionalItems) > 0 {
	// 	if aJSON, err := gjson.EncodeString(req.AdditionalItems); err == nil {
	// 		additionalItemsJSON = aJSON
	// 	}
	// }

	// 开启事务保存订单主表和明细表
	err = dao.Order.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 1. 插入订单主表
		orderDo := &do.Order{
			Uuid:               orderUUID,
			ShopUuid:           shopUuid,
			ProviderName:       ProviderNameLineman,
			ProviderOrderId:    req.OrderId,
			ProviderMerchantId: req.StoreId,
			ShortOrderNumber:   req.OrderShortCode,
			OrderType:          orderType,
			OrderTime:          orderTime,
			OrderStatus:        string(consts.OrderStatusAccepted), // LINE MAN 订单固定为 ACCEPTED
			TotalAmount:        req.RestaurantRevenue,
			Subtotal:           req.RestaurantRevenue, // LINE MAN 只提供 restaurantRevenue
			// Note:               additionalItemsJSON,   // 附加项序列化
			RawData: rawData,
		}

		_, err := dao.Order.Ctx(ctx).Data(orderDo).Insert()
		if err != nil {
			return gerror.Wrap(err, "插入订单失败")
		}

		// 2. 插入订单明细
		for _, item := range req.Items {
			// 序列化 properties（商品选项）
			propertiesJSON := ""
			if len(item.Properties) > 0 {
				if pJSON, err := gjson.EncodeString(item.Properties); err == nil {
					propertiesJSON = pJSON
				}
			}

			itemDo := &do.OrderItem{
				OrderUuid:      orderUUID,
				ProviderItemId: item.Id,
				ItemName:       item.Id, // LINE MAN 没有单独的商品名称字段，使用 ID
				Quantity:       item.Quantity,
				Price:          item.UnitPrice,
				TotalPrice:     item.UnitPrice * float64(item.Quantity),
				Modifiers:      propertiesJSON, // properties 序列化为 modifiers
				Note:           item.Memo,
			}

			_, err := dao.OrderItem.Ctx(ctx).Data(itemDo).Insert()
			if err != nil {
				return gerror.Wrap(err, "插入订单明细失败")
			}
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return orderUUID, nil
}
