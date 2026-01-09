// Package grab_order 提供 GrabFood 订单服务的业务逻辑
package grab

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/guid"
	grabfood "github.com/grab/grabfood-api-sdk-go"

	api "ttpos-bmp/app/ttpos-takeout/api/order"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/dao"
	"ttpos-bmp/app/ttpos-takeout/internal/model/do"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/model/entity"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
	"ttpos-bmp/internal/pkg/queue"
)

const (
	// TopicGrabOrder Grab 订单 MQ Topic
	TopicGrabOrder = "takeout_grab_order"
	// ProviderNameGrab Grab 供应商名称常量
	ProviderNameGrab = string(consts.ProviderGrab) // "grab"
)

// sGrabOrder 订单服务
type sGrabOrder struct{}



// HandleSubmitOrder 处理 Grab 提交订单 Webhook
// 签名验证已由中间件完成，此处只处理业务逻辑
// 使用 SDK grabfood.SubmitOrderRequest 替换自定义 DTO
func (s *sGrab) HandleSubmitOrder(ctx context.Context, req *grabfood.SubmitOrderRequest) error {
	// 保存订单
	orderUUID, err := s.saveOrderFromSDK(ctx, req)
	if err != nil {
		g.Log().Errorf(ctx, "保存订单失败: %v", err)
		return gerror.Wrap(err, "保存订单失败")
	}

	// 3. 发送 MQ 消息
	event := &grab.OrderEvent{
		Action:       "create",
		ProviderName: string(consts.ProviderGrab),
		ShopUUID:     req.GetPartnerMerchantID(), // partnerMerchantID 即为 shopUuid
		OrderUUID:    orderUUID,
		OrderID:      req.GetOrderID(),
		MerchantID:   req.GetPartnerMerchantID(),
		Status:       req.GetOrderState(),
		Timestamp:    gtime.Now().Unix(),
	}
	if err := queue.PushWithContext(ctx, TopicGrabOrder, event); err != nil {
		// MQ 发送失败只记录日志，不影响主流程（订单已入库）
		g.Log().Warningf(ctx, "发送订单 MQ 事件失败 %s: %v", orderUUID, err)
	}

	g.Log().Infof(ctx, "成功处理 Grab 订单: %s (UUID: %s)", req.GetOrderID(), orderUUID)
	return nil
}

// saveOrderFromSDK 保存订单到数据库 (使用 SDK Model)
func (s *sGrab) saveOrderFromSDK(ctx context.Context, req *grabfood.SubmitOrderRequest) (string, error) {
	orderUUID := guid.S()

	// 查询 ShopUuid - 优先使用 partnerMerchantID
	shopUuid := req.GetPartnerMerchantID()
	if shopUuid == "" {
		// fallback 到 merchantID
		cfg, err := service.ShopProviderCfg().GetShopProviderCfgByMerchantID(ctx, req.GetMerchantID(), string(consts.ProviderGrab))
		if err != nil {
			g.Log().Warningf(ctx, "查询门店配置失败: merchantID=%s, error=%v", req.GetMerchantID(), err)
		} else if cfg != nil {
			shopUuid = strconv.FormatUint(cfg.ShopUuid, 10)
			g.Log().Infof(ctx, "找到门店配置: merchantID=%s, shopUuid=%s", req.GetMerchantID(), shopUuid)
		} else {
			g.Log().Warningf(ctx, "未找到门店配置: merchantID=%s, provider=%s", req.GetMerchantID(), string(consts.ProviderGrab))
		}
	}

	// 转换价格 (最小单位 -> 元)
	currency := req.GetCurrency()
	exponent := int(currency.GetExponent())
	if exponent == 0 {
		exponent = 2 // 默认 2 位小数
	}
	divisor := float64(1)
	for i := 0; i < exponent; i++ {
		divisor *= 10
	}

	// 解析配送地址
	var deliveryAddressJSON string
	if req.HasReceiver() {
		receiver := req.GetReceiver()
		if receiver.HasAddress() {
			if addrJSON, err := gjson.EncodeString(receiver.GetAddress()); err == nil {
				deliveryAddressJSON = addrJSON
			}
		}
	}

	// 获取价格信息
	price := req.GetPrice()

	// 序列化请求体用于保存原始数据
	rawData, err := gjson.EncodeString(req)
	if err != nil {
		g.Log().Errorf(ctx, "序列化请求失败: %v", err)
		return "", gerror.Wrap(err, "序列化请求失败")
	}

	// 开启事务
	err = dao.Order.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 1. 插入订单主表
		orderDo := &do.Order{
			Uuid:               orderUUID,
			ShopUuid:           shopUuid,
			ProviderMerchantId: req.GetPartnerMerchantID(),
			ProviderOrderId:    req.GetOrderID(),
			ShortOrderNumber:   req.GetShortOrderNumber(),
			ProviderName:       string(consts.ProviderGrab),
			OrderType:          req.GetFeatureFlags().OrderType,
			OrderTime:          parseTime(req.GetOrderTime()),
			OrderStatus:        req.GetOrderState(),
			ScheduledTime:      parseTime(req.GetScheduledTime()),
			Currency:           currency.GetCode(),
			Subtotal:           float64(price.GetSubtotal()) / divisor,
			TotalAmount:        float64(price.GetEaterPayment()) / divisor,
			MerchantCharge:     float64(price.GetMerchantChargeFee()) / divisor,
			TaxAmount:          float64(price.GetTax()) / divisor,
			DiscountAmount:     float64(price.GetGrabFundPromo()+price.GetMerchantFundPromo()) / divisor,
			MerchantFundPromo:  float64(price.GetMerchantFundPromo()) / divisor,
			PaymentType:        req.GetPaymentType(),
			IsMexEditOrder:     boolToIntFromFeatureFlags(req.GetFeatureFlags()),
			Cutlery:            boolToInt(req.GetCutlery()),
			EaterCount:         getEaterCountFromSDK(req),
			CustomerName:       getCustomerNameFromSDK(req),
			CustomerPhone:      getCustomerPhoneFromSDK(req),
			DeliveryAddress:    deliveryAddressJSON,
			Note:               "", // 从 items 中提取
			RawData:            rawData,
		}

		_, err := dao.Order.Ctx(ctx).Data(orderDo).Insert()
		if err != nil {
			return gerror.Wrap(err, "插入订单失败")
		}

		// 2. 插入订单明细
		for _, item := range req.GetItems() {
			// modifiers 字段已改为 TEXT 类型，可以直接使用字符串
			var modifiersJSON string
			if len(item.GetModifiers()) > 0 {
				if mJSON, err := gjson.EncodeString(item.GetModifiers()); err == nil {
					modifiersJSON = mJSON
				}
			}

			var outOfStockInstr string
			if item.HasOutOfStockInstruction() {
				instr := item.GetOutOfStockInstruction()
				if instr.HasInstructionType() {
					outOfStockInstr = instr.GetInstructionType()
				}
			}

			itemDo := &do.OrderItem{
				OrderUuid:             orderUUID,
				ProviderItemId:        item.GetId(),
				ItemName:              item.GetSpecifications(), // Grab 用 specifications 字段表示商品名
				Quantity:              int(item.GetQuantity()),
				Price:                 float64(item.GetPrice()) / divisor,
				TotalPrice:            float64(item.GetPrice()*int64(item.GetQuantity())) / divisor,
				Specifications:        item.GetSpecifications(),
				Modifiers:             modifiersJSON,
				OutOfStockInstruction: outOfStockInstr,
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

// HandlePushOrderState 处理订单状态变更 Webhook
// 签名验证已由中间件完成，此处只处理业务逻辑
// 使用 SDK grabfood.OrderStateRequest
func (s *sGrab) HandlePushOrderState(ctx context.Context, req *grabfood.OrderStateRequest) error {
	// 1. 序列化用于保存原始数据
	rawData, _ := gjson.EncodeString(req)

	// 2. 查询订单
	var order entity.Order
	var orderExists bool
	err := dao.Order.Ctx(ctx).
		Where(dao.Order.Columns().ProviderName, string(consts.ProviderGrab)).
		Where(dao.Order.Columns().ProviderOrderId, req.GetOrderID()).
		Scan(&order)
	if err == nil {
		orderExists = true
	}

	// 3. 记录状态变更日志（无论订单是否存在都记录）
	logUUID := guid.S()
	var driverEta int
	if req.HasDriverETA() {
		driverEta = int(req.GetDriverETA())
	}

	remark := req.GetMessage()
	if !orderExists {
		remark = fmt.Sprintf("[订单不存在] %s", remark)
	}

	logDo := &do.OrderStatusLog{
		Uuid:         logUUID,
		OrderUuid:    order.Uuid, // 订单不存在时为空
		ProviderName: string(consts.ProviderGrab),
		StatusBefore: order.OrderStatus, // 订单不存在时为空
		StatusAfter:  req.GetState(),
		ChangeSource: "WEBHOOK",
		DriverEta:    driverEta,
		Remark:       remark,
		RawData:      rawData,
	}

	_, logErr := dao.OrderStatusLog.Ctx(ctx).Data(logDo).Insert()
	if logErr != nil {
		g.Log().Errorf(ctx, "插入状态日志失败: %v", logErr)
	}

	// 4. 如果订单不存在，返回错误
	if !orderExists {
		g.Log().Errorf(ctx, "订单不存在: %s", req.GetOrderID())
		return gerror.Newf("订单不存在: %s", req.GetOrderID())
	}

	// 5. 更新订单状态
	_, err = dao.Order.Ctx(ctx).
		Where(dao.Order.Columns().Uuid, order.Uuid).
		Data(g.Map{
			dao.Order.Columns().OrderStatus: req.GetState(),
		}).Update()
	if err != nil {
		g.Log().Errorf(ctx, "更新订单状态失败: %v", err)
		return gerror.Wrap(err, "更新订单状态失败")
	}

	// 6. 发送 MQ 消息
	event := &grab.OrderEvent{
		Action:       "status_update",
		ProviderName: string(consts.ProviderGrab),
		ShopUUID:     order.ShopUuid,
		OrderUUID:    order.Uuid,
		OrderID:      req.GetOrderID(),
		MerchantID:   req.GetMerchantID(),
		Status:       req.GetState(),
		Timestamp:    gtime.Now().Unix(),
		Message:      req.GetMessage(),
	}
	if err := queue.PushWithContext(ctx, TopicGrabOrder, event); err != nil {
		g.Log().Warningf(ctx, "发送订单状态更新 MQ 事件失败 %s: %v", order.Uuid, err)
	}

	g.Log().Infof(ctx, "订单状态已更新: %s -> %s (订单ID: %s)", order.OrderStatus, req.GetState(), req.GetOrderID())
	return nil
}

// 辅助函数
func parseTime(timeStr string) *gtime.Time {
	if timeStr == "" {
		return nil
	}
	t, err := gtime.StrToTime(timeStr)
	if err != nil {
		return nil
	}
	return t
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// boolToIntFromFeatureFlags 从 OrderFeatureFlags 获取 IsMexEditOrder 并转为 int
func boolToIntFromFeatureFlags(flags grabfood.OrderFeatureFlags) int {
	if flags.HasIsMexEditOrder() && flags.GetIsMexEditOrder() {
		return 1
	}
	return 0
}

// ============================================================================
// SDK Model 辅助函数
// ============================================================================

// getOrderTypeFromSDK 从 SDK SubmitOrderRequest 获取订单类型
func getOrderTypeFromSDK(req *grabfood.SubmitOrderRequest) string {
	if req.HasDineIn() {
		return "DineIn"
	}
	// 根据 featureFlags 或其他字段判断
	return "DeliveryByProvider"
}

// getEaterCountFromSDK 从 SDK SubmitOrderRequest 获取用餐人数
func getEaterCountFromSDK(req *grabfood.SubmitOrderRequest) int {
	if req.HasDineIn() {
		dineIn := req.GetDineIn()
		return int(dineIn.GetEaterCount())
	}
	return 0
}

// getCustomerNameFromSDK 从 SDK SubmitOrderRequest 获取客户姓名
func getCustomerNameFromSDK(req *grabfood.SubmitOrderRequest) string {
	if req.HasReceiver() {
		receiver := req.GetReceiver()
		if receiver.HasName() {
			return receiver.GetName()
		}
	}
	return ""
}

// getCustomerPhoneFromSDK 从 SDK SubmitOrderRequest 获取客户电话
func getCustomerPhoneFromSDK(req *grabfood.SubmitOrderRequest) string {
	if !req.HasReceiver() {
		return ""
	}
	receiver := req.GetReceiver()
	// SDK Receiver 使用 Phones 字段存储电话
	if receiver.HasPhones() {
		return receiver.GetPhones()
	}
	return ""
}

// PrepareOrder 准备订单（接受/拒绝）
// 参数：
//   - ctx: 上下文对象
//   - orderEntity: 订单实体
//   - toState: 目标状态 (Accepted/Rejected)
//
// 返回：
//   - err: 错误信息
func (s *sGrab) PrepareOrder(ctx context.Context, orderEntityInterface interface{}, toState string) error {
	// 类型断言获取订单实体
	orderEntity, ok := orderEntityInterface.(*entity.Order)
	if !ok {
		return gerror.New("订单实体类型错误")
	}

	// 调用 GrabFood SDK 接受/拒绝订单
	err := s.callGrabAcceptRejectAPI(ctx, orderEntity, toState)
	if err != nil {
		g.Log().Errorf(ctx, "调用 GrabFood API 失败: %v", err)
		return gerror.Wrap(err, "调用 GrabFood API 失败")
	}

	g.Log().Infof(ctx, "订单 %s 已成功调用 Grab API: %s", orderEntity.Uuid, toState)
	return nil
}

// callGrabAcceptRejectAPI 调用 GrabFood SDK 的 accept-reject-order API
// 参数：
//   - ctx: 上下文对象
//   - orderEntity: 订单实体，包含 ProviderOrderId
//   - toState: 目标状态，"Accepted" 或 "Rejected"
//
// 返回：
//   - err: 错误信息
func (s *sGrab) callGrabAcceptRejectAPI(ctx context.Context, orderEntity *entity.Order, toState string) error {
	g.Log().Infof(ctx, "准备调用 GrabFood API: orderID=%s, toState=%s", orderEntity.ProviderOrderId, toState)

	// 根据 toState 调用相应的 Grab 服务方法
	switch toState {
	case string(consts.OrderPrepareStateAccepted):
		// 调用接受订单 API
		err := service.Grab().AcceptOrder(ctx, orderEntity.ProviderOrderId)
		if err != nil {
			g.Log().Errorf(ctx, "接受订单失败: orderID=%s, error=%v", orderEntity.ProviderOrderId, err)
			return gerror.Wrap(err, "接受订单失败")
		}
		g.Log().Infof(ctx, "成功接受订单: orderID=%s", orderEntity.ProviderOrderId)
		return nil

	case string(consts.OrderPrepareStateRejected):
		// 调用拒绝订单 API
		err := service.Grab().RejectOrder(ctx, orderEntity.ProviderOrderId, 0)
		if err != nil {
			g.Log().Errorf(ctx, "拒绝订单失败: orderID=%s, error=%v", orderEntity.ProviderOrderId, err)
			return gerror.Wrap(err, "拒绝订单失败")
		}
		g.Log().Infof(ctx, "成功拒绝订单: orderID=%s", orderEntity.ProviderOrderId)
		return nil

	default:
		return gerror.Newf("不支持的订单状态: %s，必须为 %s 或 %s", toState, consts.OrderPrepareStateAccepted, consts.OrderPrepareStateRejected)
	}
}

// MarkOrderReady 标记订单准备完成
// 参数：
//   - ctx: 上下文对象
//   - orderEntity: 订单实体，包含 ProviderOrderId 等信息
//
// 返回：
//   - err: 错误信息
func (s *sGrab) MarkOrderReadyEntity(ctx context.Context, orderEntity *entity.Order) error {
	// 1. 参数验证
	if orderEntity == nil {
		return gerror.New("订单实体不能为空")
	}
	if orderEntity.ProviderName != ProviderNameGrab {
		return gerror.Newf("订单渠道错误，期望 grab，实际 %s", orderEntity.ProviderName)
	}
	if orderEntity.ProviderOrderId == "" {
		return gerror.New("provider_order_id 不能为空")
	}

	// 2. 记录开始日志
	g.Log().Infof(ctx, "开始调用 GrabFood MarkOrderReady API: provider_order_id=%s, order_uuid=%s",
		orderEntity.ProviderOrderId, orderEntity.Uuid)

	// 3. 根据订单类型确定 markStatus
	markStatus := "1" // 默认值
	if orderEntity.OrderType == "DineIn" {
		markStatus = "2"
	}
	err := service.Grab().MarkOrderReady(ctx, orderEntity.ProviderOrderId, markStatus)
	if err != nil {
		// 记录详细错误日志
		g.Log().Errorf(ctx, "调用 GrabFood MarkOrderReady API 失败: order_id=%s, order_uuid=%s, error=%v",
			orderEntity.ProviderOrderId, orderEntity.Uuid, err)
		return gerror.Wrapf(err, "调用 GrabFood API 失败")
	}

	// 4. 记录成功日志
	g.Log().Infof(ctx, "调用 GrabFood MarkOrderReady API 成功: provider_order_id=%s, order_uuid=%s",
		orderEntity.ProviderOrderId, orderEntity.Uuid)

	return nil
}

// CheckOrderCancelable 检查订单是否可取消
// 参数：
//   - ctx: 上下文对象
//   - orderEntity: 订单实体
//
// 返回：
//   - res: 检查订单可取消性响应
//   - err: 错误信息
func (s *sGrab) CheckOrderCancelableEntity(ctx context.Context, orderEntity *entity.Order) (*api.CheckOrderCancelableResp, error) {
	// 1. 参数验证
	if orderEntity == nil {
		return nil, gerror.New("订单实体不能为空")
	}
	if orderEntity.ProviderName != ProviderNameGrab {
		return nil, gerror.Newf("订单渠道错误，期望 grab，实际 %s", orderEntity.ProviderName)
	}

	// 2. 调用 Grab API 检查订单是否可取消
	sdkResp, err := service.Grab().CheckOrderCancelable(ctx, orderEntity.ProviderMerchantId, orderEntity.ProviderOrderId)
	if err != nil {
		g.Log().Errorf(ctx, "检查订单可取消性失败: order_id=%s, merchant_id=%s, error=%v",
			orderEntity.ProviderOrderId, orderEntity.ProviderMerchantId, err)
		return nil, gerror.Wrap(err, "检查订单可取消性失败")
	}

	// 4. 提取核心字段
	canCancel := sdkResp.CancelAble != nil && *sdkResp.CancelAble
	nonCancelReason := ""
	if sdkResp.NonCancellationReason != nil {
		nonCancelReason = *sdkResp.NonCancellationReason
	}

	// 5. 序列化 SDK 完整响应为 JSON（包含所有字段）
	rawDataJSON, err := gjson.EncodeString(sdkResp)
	if err != nil {
		g.Log().Warningf(ctx, "序列化SDK响应失败: %v", err)
		rawDataJSON = "{}"
	}

	// 6. 返回精简响应
	g.Log().Infof(ctx, "订单可取消性检查完成: order_uuid=%s, can_cancel=%v, reason=%s",
		orderEntity.Uuid, canCancel, nonCancelReason)
	return &api.CheckOrderCancelableResp{
		OrderUuid:             orderEntity.Uuid,
		CanCancel:             canCancel,
		NonCancellationReason: nonCancelReason,
		RawData:               rawDataJSON,
	}, nil
}

// CancelOrder 取消订单
// 参数：
//   - ctx: 上下文对象
//   - orderEntity: 订单实体
//   - cancelCode: 取消原因码（字符串格式，可根据不同平台传入不同的编码）
//
// 返回：
//   - res: 取消订单响应
//   - err: 错误信息
func (s *sGrab) CancelOrderEntity(ctx context.Context, orderEntity *entity.Order, cancelCode string) (res *api.CancelOrderResp, err error) {
	// 1. 参数验证
	if orderEntity == nil {
		return nil, gerror.New("订单实体不能为空")
	}
	if orderEntity.ProviderName != ProviderNameGrab {
		return nil, gerror.Newf("订单渠道错误，期望 grab，实际 %s", orderEntity.ProviderName)
	}
	if orderEntity.ProviderOrderId == "" {
		return nil, gerror.New("provider_order_id 不能为空")
	}
	if orderEntity.ProviderMerchantId == "" {
		return nil, gerror.New("provider_merchant_id 不能为空")
	}

	// 2. 记录开始日志
	g.Log().Infof(ctx, "开始取消订单: order_uuid=%s, order_id=%s, merchant_id=%s, cancel_code=%s",
		orderEntity.Uuid, orderEntity.ProviderOrderId, orderEntity.ProviderMerchantId, cancelCode)

	// 3. 执行取消操作（不再包含预检查逻辑）
	err = service.Grab().CancelOrder(ctx, orderEntity.ProviderMerchantId, orderEntity.ProviderOrderId, cancelCode)
	if err != nil {
		g.Log().Errorf(ctx, "取消订单失败: order_id=%s, cancel_code=%s, error=%v",
			orderEntity.ProviderOrderId, cancelCode, err)
		return nil, gerror.Wrap(err, "取消订单失败")
	}

	// 4. 返回成功响应
	g.Log().Infof(ctx, "订单取消成功: order_uuid=%s, order_id=%s", orderEntity.Uuid, orderEntity.ProviderOrderId)
	return &api.CancelOrderResp{
		OrderUuid: orderEntity.Uuid,
	}, nil
}
