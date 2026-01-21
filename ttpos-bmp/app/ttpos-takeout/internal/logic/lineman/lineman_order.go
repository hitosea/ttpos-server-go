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
	"ttpos-bmp/internal/pkg/queue"
)

const (
	// TopicLinemanOrder LINE MAN 订单 MQ Topic（复用 Grab 的 topic）
	TopicLinemanOrder = "takeout_grab_order"
	// ProviderNameLineman LINE MAN 供应商名称常量
	ProviderNameLineman = string(consts.ProviderLineman) // "lineman"
)

// HandlePlaceOrder 处理 LINE MAN 提交订单 Webhook
// 参数验证已由 GoFrame 自动完成，此处只处理业务逻辑
func (s *sLineman) HandlePlaceOrder(ctx context.Context, req *v1.PlaceOrderReq) error {
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
func (s *sLineman) saveOrder(ctx context.Context, req *v1.PlaceOrderReq) (string, error) {
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
			PaymentType:        "LINEMAN",             // LINE MAN 平台支付，固定为 LINEMAN
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

// HandleOrderUpdate 处理 LINE MAN 订单更新 Webhook
// 参数验证已由 GoFrame 自动完成，此处只处理业务逻辑
func (s *sLineman) HandleOrderUpdate(ctx context.Context, req *v1.OrderUpdateReq) error {
	// 1. 查询现有订单
	existingOrder, err := dao.Order.Ctx(ctx).
		Where(dao.Order.Columns().ProviderName, ProviderNameLineman).
		Where(dao.Order.Columns().ProviderOrderId, req.OrderId).
		One()
	if err != nil {
		g.Log().Errorf(ctx, "查询订单失败: orderId=%s, error=%v", req.OrderId, err)
		return gerror.Wrap(err, "查询订单失败")
	}
	if existingOrder.IsEmpty() {
		g.Log().Warningf(ctx, "订单不存在: orderId=%s", req.OrderId)
		return gerror.New("订单不存在")
	}

	// 2. 幂等性检查（比较 LINE MAN 的 orderUpdatedTime 与数据库的 updated_at）
	dbUpdatedAt := existingOrder[dao.Order.Columns().UpdatedAt].Int64()
	if dbUpdatedAt > 0 {
		// 解析 LINE MAN 的 orderUpdatedTime
		newUpdatedTime, err := gtime.StrToTime(req.OrderUpdatedTime)
		if err == nil && newUpdatedTime != nil {
			// 如果 LINE MAN 的更新时间 <= 数据库记录时间，则跳过
			if newUpdatedTime.Unix() <= dbUpdatedAt {
				g.Log().Infof(ctx, "订单更新时间未变化，跳过: orderId=%s, db_updated_at=%d, lineman_updated_time=%s",
					req.OrderId, dbUpdatedAt, req.OrderUpdatedTime)
				return nil
			}
		}
	}

	// 3. 更新订单（事务）
	orderUUID := existingOrder[dao.Order.Columns().Uuid].String()
	err = s.updateOrder(ctx, orderUUID, req)
	if err != nil {
		g.Log().Errorf(ctx, "更新订单失败: orderId=%s, error=%v", req.OrderId, err)
		return gerror.Wrap(err, "更新订单失败")
	}

	// 4. 发送 RocketMQ 事件
	event := &grab.OrderEvent{
		Action:       string(consts.OrderActionUpdate),
		ProviderName: ProviderNameLineman,
		ShopUUID:     req.StoreId, // 使用 storeId 作为 shopUuid
		OrderUUID:    orderUUID,
		OrderID:      req.OrderId,
		// MerchantID:   req.StoreId,  //lineman 的商户id 未知
		Status:    string(consts.OrderStatusAccepted),
		Timestamp: gtime.Now().Unix(),
	}
	if err := queue.PushWithContext(ctx, TopicLinemanOrder, event); err != nil {
		// RocketMQ 发送失败只记录日志，不影响主流程（订单已入库）
		g.Log().Warningf(ctx, "发送订单更新 RocketMQ 事件失败 %s: %v", orderUUID, err)
	}

	g.Log().Infof(ctx, "成功更新 LINE MAN 订单: %s (UUID: %s)", req.OrderId, orderUUID)
	return nil
}

// updateOrder 更新订单到数据库（事务）
func (s *sLineman) updateOrder(ctx context.Context, orderUUID string, req *v1.OrderUpdateReq) error {
	return dao.Order.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 1. 更新订单主表（包括 raw_data）
		orderTime, err := gtime.StrToTime(req.OrderAcceptedTime)
		if err != nil {
			return gerror.Wrap(err, "解析订单时间失败")
		}
		rawDataJSON, err := gjson.EncodeString(req)
		if err != nil {
			return gerror.Wrap(err, "序列化 raw_data 失败")
		}

		_, err = dao.Order.Ctx(ctx).Where(dao.Order.Columns().Uuid, orderUUID).Update(&do.Order{
			TotalAmount: req.RestaurantRevenue,
			Subtotal:    req.RestaurantRevenue,
			OrderTime:   orderTime,
			RawData:     rawDataJSON,
			// UpdatedAt:   gtime.Now().Unix(),
		})
		if err != nil {
			return gerror.Wrap(err, "更新订单主表失败")
		}

		// 2. 删除旧的订单明细
		_, err = dao.OrderItem.Ctx(ctx).Where(dao.OrderItem.Columns().OrderUuid, orderUUID).Delete()
		if err != nil {
			return gerror.Wrap(err, "删除旧订单明细失败")
		}

		// 3. 插入新的订单明细
		for _, item := range req.Items {
			propertiesJSON, err := gjson.EncodeString(item.Properties)
			if err != nil {
				g.Log().Warningf(ctx, "序列化商品属性失败: itemId=%s, error=%v", item.Id, err)
				propertiesJSON = "[]"
			}

			_, err = dao.OrderItem.Ctx(ctx).Data(&do.OrderItem{
				OrderUuid:      orderUUID,
				ProviderItemId: item.Id,
				ItemName:       item.Id, // LINE MAN 只提供 ID，商品名称可能需要从菜单数据查询
				Quantity:       item.Quantity,
				Price:          item.UnitPrice,
				TotalPrice:     item.UnitPrice * float64(item.Quantity),
				Modifiers:      propertiesJSON,
				Note:           item.Memo,
				// CreatedAt:      gtime.Now().Unix(),
			}).Insert()
			if err != nil {
				return gerror.Wrapf(err, "插入订单明细失败: itemId=%s", item.Id)
			}
		}

		return nil
	})
}

// LINE MAN 订单状态常量
const (
	LinemanStatusFinish   = "FINISH"
	LinemanStatusCanceled = "CANCELED" // LINE MAN 使用一个 L
)

// mapLinemanStatusToTTPOS 将 LINE MAN 状态映射为 TTPOS 内部状态
func mapLinemanStatusToTTPOS(linemanStatus string) string {
	switch linemanStatus {
	case LinemanStatusFinish:
		return string(consts.OrderStatusCompleted) // "COMPLETED"
	case LinemanStatusCanceled:
		return string(consts.OrderStatusCancelled) // "CANCELLED" (TTPOS 使用两个 L)
	default:
		g.Log().Warningf(context.Background(), "未知的 LINE MAN 状态: %s", linemanStatus)
		return linemanStatus // 未知状态保持原样
	}
}

// HandleOrderStatusUpdate 处理 LINE MAN 订单状态更新 Webhook
// 参数验证已由 GoFrame 自动完成，此处只处理业务逻辑
func (s *sLineman) HandleOrderStatusUpdate(ctx context.Context, req *v1.OrderStatusUpdateReq) error {
	// 1. 状态映射
	ttposStatus := mapLinemanStatusToTTPOS(req.OrderStatus)

	// 2. 序列化原始请求（用于状态日志）
	rawData, _ := gjson.EncodeString(req)

	// 3. 查询现有订单
	existingOrder, err := dao.Order.Ctx(ctx).
		Where(dao.Order.Columns().ProviderName, ProviderNameLineman).
		Where(dao.Order.Columns().ProviderOrderId, req.OrderId).
		Where(dao.Order.Columns().DeletedAt, 0). // 使用正确的软删除字段
		One()
	if err != nil || existingOrder.IsEmpty() {
		// 订单不存在时也记录状态日志（便于排查问题）
		logDo := &do.OrderStatusLog{
			Uuid:         guid.S(),
			OrderUuid:    "",
			ProviderName: ProviderNameLineman,
			StatusBefore: "",
			StatusAfter:  ttposStatus,
			ChangeSource: "WEBHOOK",
			Remark:       "[订单不存在] orderId=" + req.OrderId,
			RawData:      rawData,
		}
		if _, logErr := dao.OrderStatusLog.Ctx(ctx).Data(logDo).Insert(); logErr != nil {
			g.Log().Errorf(ctx, "插入状态日志失败: %v", logErr)
		}
		return gerror.New("订单不存在")
	}

	// 4. 幂等性检查
	orderUUID := existingOrder[dao.Order.Columns().Uuid].String()
	currentStatus := existingOrder[dao.Order.Columns().OrderStatus].String()
	if currentStatus == ttposStatus {
		g.Log().Infof(ctx, "订单状态未变化，跳过: orderId=%s, status=%s", req.OrderId, ttposStatus)
		return nil
	}

	// 5. 记录状态变更日志
	logDo := &do.OrderStatusLog{
		Uuid:         guid.S(),
		OrderUuid:    orderUUID,
		ProviderName: ProviderNameLineman,
		StatusBefore: currentStatus,
		StatusAfter:  ttposStatus,
		ChangeSource: "WEBHOOK",
		RawData:      rawData,
	}
	if _, logErr := dao.OrderStatusLog.Ctx(ctx).Data(logDo).Insert(); logErr != nil {
		g.Log().Errorf(ctx, "插入状态日志失败: %v", logErr)
	}

	// 6. 更新订单状态
	_, err = dao.Order.Ctx(ctx).Where("uuid", orderUUID).Update(&do.Order{
		OrderStatus: ttposStatus,
		UpdatedAt:   gtime.Now().Unix(),
	})
	if err != nil {
		return gerror.Wrap(err, "更新订单状态失败")
	}

	// 5. 发送 RocketMQ 事件
	event := &grab.OrderEvent{
		Action:       string(consts.OrderActionStatusUpdate),
		ProviderName: ProviderNameLineman,
		ShopUUID:     existingOrder[dao.Order.Columns().ShopUuid].String(),
		OrderUUID:    orderUUID,
		OrderID:      req.OrderId,
		Status:       ttposStatus,
		Timestamp:    gtime.Now().Unix(),
	}
	if err := queue.PushWithContext(ctx, TopicLinemanOrder, event); err != nil {
		// RocketMQ 发送失败只记录日志，不影响主流程（订单状态已更新）
		g.Log().Warningf(ctx, "发送订单状态更新 MQ 事件失败 %s: %v", orderUUID, err)
	}

	g.Log().Infof(ctx, "成功更新 LINE MAN 订单状态: %s (UUID: %s) %s -> %s",
		req.OrderId, orderUUID, currentStatus, ttposStatus)
	return nil
}
