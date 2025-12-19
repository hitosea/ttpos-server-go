// Package grab_order 提供 GrabFood 订单服务的业务逻辑
package grab_order

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/guid"
	grabfood "github.com/grab/grabfood-api-sdk-go"

	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/dao"
	"ttpos-bmp/app/ttpos-takeout/internal/model/do"
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

// sGrabOrder 订单服务
type sGrabOrder struct{}

func init() {
	service.RegisterGrabOrder(New())
}

// New 创建订单服务实例
func New() *sGrabOrder {
	return &sGrabOrder{}
}

// HandleSubmitOrder 处理 Grab 提交订单 Webhook
// 签名验证已由中间件完成，此处只处理业务逻辑
// 使用 SDK grabfood.SubmitOrderRequest 替换自定义 DTO
func (s *sGrabOrder) HandleSubmitOrder(ctx context.Context, req *grabfood.SubmitOrderRequest) error {
	// 保存订单
	orderUUID, err := s.saveOrderFromSDK(ctx, req)
	if err != nil {
		g.Log().Errorf(ctx, "保存订单失败: %v", err)
		return fmt.Errorf("保存订单失败: %w", err)
	}

	// 3. 发送 MQ 消息
	event := &OrderEvent{
		Action:       "create",
		ProviderName: string(consts.ProviderGrab),
		OrderUUID:    orderUUID,
		OrderID:      req.GetOrderID(),
		MerchantID:   req.GetPartnerMerchantID(), // 保持 MQ 事件中的字段名不变
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
func (s *sGrabOrder) saveOrderFromSDK(ctx context.Context, req *grabfood.SubmitOrderRequest) (string, error) {
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
		return "", fmt.Errorf("序列化请求失败: %w", err)
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
			RawData:            rawData,
		}

		_, err := dao.Order.Ctx(ctx).Data(orderDo).Insert()
		if err != nil {
			return fmt.Errorf("插入订单失败: %w", err)
		}

		// 2. 插入订单明细
		for _, item := range req.GetItems() {
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
				return fmt.Errorf("插入订单明细失败: %w", err)
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
// 使用 SDK grabfood.OrderStateRequest 替换自定义 DTO
func (s *sGrabOrder) HandlePushOrderState(ctx context.Context, body []byte) error {
	// 1. 解析请求 - 使用 SDK Model
	var req grabfood.OrderStateRequest
	if err := gjson.DecodeTo(body, &req); err != nil {
		g.Log().Errorf(ctx, "解析订单状态请求失败: %v", err)
		return fmt.Errorf("解析请求失败: %w", err)
	}

	// 2. 查询订单
	var order entity.Order
	err := dao.Order.Ctx(ctx).
		Where(dao.Order.Columns().ProviderName, string(consts.ProviderGrab)).
		Where(dao.Order.Columns().ProviderOrderId, req.GetOrderID()).
		Scan(&order)
	if err != nil {
		g.Log().Errorf(ctx, "订单不存在: %s", req.GetOrderID())
		return fmt.Errorf("订单不存在: %s", req.GetOrderID())
	}

	// 4. 记录状态变更日志
	logUUID := guid.S()
	var driverEta int
	if req.HasDriverETA() {
		driverEta = int(req.GetDriverETA())
	}

	logDo := &do.OrderStatusLog{
		Uuid:         logUUID,
		OrderUuid:    order.Uuid,
		ProviderName: string(consts.ProviderGrab),
		StatusBefore: order.OrderStatus,
		StatusAfter:  req.GetState(),
		ChangeSource: "WEBHOOK",
		DriverEta:    driverEta,
		Remark:       req.GetMessage(),
		RawData:      string(body),
	}

	_, err = dao.OrderStatusLog.Ctx(ctx).Data(logDo).Insert()
	if err != nil {
		g.Log().Errorf(ctx, "插入状态日志失败: %v", err)
		return fmt.Errorf("插入状态日志失败: %w", err)
	}

	// 5. 更新订单状态
	_, err = dao.Order.Ctx(ctx).
		Where(dao.Order.Columns().Uuid, order.Uuid).
		Data(g.Map{
			dao.Order.Columns().OrderStatus: req.GetState(),
			dao.Order.Columns().UpdatedAt:   gtime.Now(),
		}).Update()
	if err != nil {
		g.Log().Errorf(ctx, "更新订单状态失败: %v", err)
		return fmt.Errorf("更新订单状态失败: %w", err)
	}

	// 6. 发送 MQ 消息
	event := &OrderEvent{
		Action:       "status_update",
		ProviderName: string(consts.ProviderGrab),
		OrderUUID:    order.Uuid,
		OrderID:      req.GetOrderID(),
		MerchantID:   req.GetPartnerMerchantID(), // 保持 MQ 事件中的字段名不变
		Status:       req.GetState(),
		Timestamp:    gtime.Now().Unix(),
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
