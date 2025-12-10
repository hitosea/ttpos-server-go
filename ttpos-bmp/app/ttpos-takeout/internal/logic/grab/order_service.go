package grab

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/google/uuid"
	grabfood "github.com/grab/grabfood-api-sdk-go"

	"ttpos-bmp/app/ttpos-takeout/internal/dao"
	"ttpos-bmp/app/ttpos-takeout/internal/model/do"
	grabDto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/model/entity"
	"ttpos-bmp/internal/pkg/queue"
)

const (
	// TopicGrabOrder Grab 订单 MQ Topic
	TopicGrabOrder = "takeout_grab_order"
)

// OrderService 订单服务
// 内部使用，通过 sGrab 统一管理
type OrderService struct {
	verifier *SignatureVerifier
}

// OrderEvent 订单事件
type OrderEvent struct {
	Action       string `json:"action"`       // create, status_update, cancel
	ProviderName string `json:"providerName"` // grab
	OrderUUID    string `json:"orderUuid"`    // 订单 UUID
	OrderID      string `json:"orderId"`      // 平台订单 ID
	MerchantID   string `json:"merchantId"`   // 商户 ID
	Status       string `json:"status"`       // 当前状态
	Timestamp    int64  `json:"timestamp"`    // 事件时间戳
}

// HandleSubmitOrder 处理 Grab 提交订单 Webhook
// 使用 SDK grabfood.SubmitOrderRequest 替换自定义 DTO
func (s *OrderService) HandleSubmitOrder(ctx context.Context, signature, timestamp string, body []byte) error {
	// 1. 验证签名
	if err := s.verifier.VerifySignature(signature, timestamp, body); err != nil {
		g.Log().Errorf(ctx, "Grab signature verification failed: %v", err)
		return fmt.Errorf("signature verification failed: %w", err)
	}

	// 2. 解析请求 - 使用 SDK Model
	var req grabfood.SubmitOrderRequest
	if err := json.Unmarshal(body, &req); err != nil {
		g.Log().Errorf(ctx, "Failed to parse submit order request: %v", err)
		return fmt.Errorf("failed to parse request: %w", err)
	}

	// 3. 保存订单
	orderUUID, err := s.saveOrderFromSDK(ctx, &req, body)
	if err != nil {
		g.Log().Errorf(ctx, "Failed to save order: %v", err)
		return fmt.Errorf("failed to save order: %w", err)
	}

	// 4. 发送 MQ 消息
	event := &OrderEvent{
		Action:       "create",
		ProviderName: "grab",
		OrderUUID:    orderUUID,
		OrderID:      req.GetOrderID(),
		MerchantID:   req.GetPartnerMerchantID(), // 保持 MQ 事件中的字段名不变
		Status:       req.GetOrderState(),
		Timestamp:    gtime.Now().Unix(),
	}
	if err := queue.PushWithContext(ctx, TopicGrabOrder, event); err != nil {
		// MQ 发送失败只记录日志，不影响主流程（订单已入库）
		g.Log().Warningf(ctx, "Failed to send MQ event for order %s: %v", orderUUID, err)
	}

	g.Log().Infof(ctx, "Successfully processed Grab order: %s (UUID: %s)", req.GetOrderID(), orderUUID)
	return nil
}

// saveOrderFromSDK 保存订单到数据库 (使用 SDK Model)
func (s *OrderService) saveOrderFromSDK(ctx context.Context, req *grabfood.SubmitOrderRequest, rawData []byte) (string, error) {
	orderUUID := uuid.New().String()

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
			if addrBytes, err := json.Marshal(receiver.GetAddress()); err == nil {
				deliveryAddressJSON = string(addrBytes)
			}
		}
	}

	// 获取价格信息
	price := req.GetPrice()

	// 开启事务
	err := dao.Order.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 1. 插入订单主表
		orderDo := &do.Order{
			Uuid:               orderUUID,
			ShopUuid:           "", // TODO: 从配置或上下文获取 shop_uuid
			ProviderMerchantId: req.GetPartnerMerchantID(),
			PartnerOrderId:     req.GetOrderID(),
			ShortOrderNumber:   req.GetShortOrderNumber(),
			ProviderName:       "grab",
			OrderType:          getOrderTypeFromSDK(req),
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
			RawData:            string(rawData),
			CreatedAt:          gtime.Now(),
			UpdatedAt:          gtime.Now(),
		}

		_, err := dao.Order.Ctx(ctx).Data(orderDo).Insert()
		if err != nil {
			return fmt.Errorf("insert order failed: %w", err)
		}

		// 2. 插入订单明细
		for _, item := range req.GetItems() {
			var modifiersJSON string
			if len(item.GetModifiers()) > 0 {
				if mBytes, err := json.Marshal(item.GetModifiers()); err == nil {
					modifiersJSON = string(mBytes)
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
				PartnerItemId:         item.GetId(),
				ItemName:              item.GetSpecifications(), // Grab 用 specifications 字段表示商品名
				Quantity:              int(item.GetQuantity()),
				Price:                 float64(item.GetPrice()) / divisor,
				TotalPrice:            float64(item.GetPrice()*int64(item.GetQuantity())) / divisor,
				Specifications:        item.GetSpecifications(),
				Modifiers:             modifiersJSON,
				OutOfStockInstruction: outOfStockInstr,
				CreatedAt:             gtime.Now(),
			}

			_, err := dao.OrderItem.Ctx(ctx).Data(itemDo).Insert()
			if err != nil {
				return fmt.Errorf("insert order item failed: %w", err)
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
// 使用 SDK grabfood.OrderStateRequest 替换自定义 DTO
func (s *OrderService) HandlePushOrderState(ctx context.Context, signature, timestamp string, body []byte) error {
	// 1. 验证签名
	if err := s.verifier.VerifySignature(signature, timestamp, body); err != nil {
		g.Log().Errorf(ctx, "Grab signature verification failed: %v", err)
		return fmt.Errorf("signature verification failed: %w", err)
	}

	// 2. 解析请求 - 使用 SDK Model
	var req grabfood.OrderStateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		g.Log().Errorf(ctx, "Failed to parse order state request: %v", err)
		return fmt.Errorf("failed to parse request: %w", err)
	}

	// 3. 查询订单
	var order entity.Order
	err := dao.Order.Ctx(ctx).
		Where(dao.Order.Columns().ProviderName, "grab").
		Where(dao.Order.Columns().PartnerOrderId, req.GetOrderID()).
		Scan(&order)
	if err != nil {
		g.Log().Errorf(ctx, "Order not found: %s", req.GetOrderID())
		return fmt.Errorf("order not found: %s", req.GetOrderID())
	}

	// 4. 记录状态变更日志
	logUUID := uuid.New().String()
	var driverEta int
	if req.HasDriverETA() {
		driverEta = int(req.GetDriverETA())
	}

	logDo := &do.OrderStatusLog{
		Uuid:         logUUID,
		OrderUuid:    order.Uuid,
		ProviderName: "grab",
		StatusBefore: order.OrderStatus,
		StatusAfter:  req.GetState(),
		ChangeSource: "WEBHOOK",
		DriverEta:    driverEta,
		Remark:       req.GetMessage(),
		RawData:      string(body),
		CreatedAt:    gtime.Now(),
	}

	_, err = dao.OrderStatusLog.Ctx(ctx).Data(logDo).Insert()
	if err != nil {
		g.Log().Errorf(ctx, "Failed to insert status log: %v", err)
		return fmt.Errorf("failed to insert status log: %w", err)
	}

	// 5. 更新订单状态
	_, err = dao.Order.Ctx(ctx).
		Where(dao.Order.Columns().Uuid, order.Uuid).
		Data(g.Map{
			dao.Order.Columns().OrderStatus: req.GetState(),
			dao.Order.Columns().UpdatedAt:   gtime.Now(),
		}).Update()
	if err != nil {
		g.Log().Errorf(ctx, "Failed to update order status: %v", err)
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// 6. 发送 MQ 消息
	event := &OrderEvent{
		Action:       "status_update",
		ProviderName: "grab",
		OrderUUID:    order.Uuid,
		OrderID:      req.GetOrderID(),
		MerchantID:   req.GetPartnerMerchantID(), // 保持 MQ 事件中的字段名不变
		Status:       req.GetState(),
		Timestamp:    gtime.Now().Unix(),
	}
	if err := queue.PushWithContext(ctx, TopicGrabOrder, event); err != nil {
		g.Log().Warningf(ctx, "Failed to send MQ event for order status update %s: %v", order.Uuid, err)
	}

	g.Log().Infof(ctx, "Order status updated: %s -> %s (Order: %s)", order.OrderStatus, req.GetState(), req.GetOrderID())
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

// ============================================================================
// 旧 DTO 辅助函数 (保留以兼容测试)
// Deprecated: Phase 3 迁移后删除
// ============================================================================

// getOrderType 从旧 DTO 获取订单类型
// Deprecated: 使用 getOrderTypeFromSDK
func getOrderType(req *grabDto.SubmitOrderRequest) string {
	if req.DineIn != nil {
		return "DineIn"
	}
	return "DeliveryByProvider"
}

// getEaterCount 从旧 DTO 获取用餐人数
// Deprecated: 使用 getEaterCountFromSDK
func getEaterCount(dineIn *grabDto.DineInInfo) int {
	if dineIn != nil {
		return dineIn.EaterCount
	}
	return 0
}

// getCustomerName 从旧 DTO 获取客户姓名
// Deprecated: 使用 getCustomerNameFromSDK
func getCustomerName(receiver *grabDto.Receiver) string {
	if receiver != nil {
		return receiver.Name
	}
	return ""
}

// getCustomerPhone 从旧 DTO 获取客户电话
// Deprecated: 使用 getCustomerPhoneFromSDK
func getCustomerPhone(receiver *grabDto.Receiver) string {
	if receiver == nil {
		return ""
	}
	if receiver.VirtualContact != nil && receiver.VirtualContact.Phone != "" {
		return receiver.VirtualContact.Phone
	}
	return receiver.Phone
}
