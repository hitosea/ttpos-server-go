package lineman

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	apiMessage "ttpos-api/ttpos-takeout/message"
	"ttpos-server-go/app/errors"
	takeoutModel "ttpos-server-go/app/modules/takeout/domain/model"
	valueobject "ttpos-server-go/app/modules/takeout/domain/value_object"

	"github.com/shopspring/decimal"
)

// LinemanConverter LINE MAN 平台转换器实现
type LineManConverter struct {
	amountConversionFactor int64 // 金额转换因子(分转元)
}

// NewLinemanConverter 创建 LINE MAN 转换器
func NewLineManConverter() *LineManConverter {
	return &LineManConverter{
		amountConversionFactor: 100,
	}
}

// ConvertPlatformStateToOrderState 将 Grab 平台订单状态转换为内部订单状态
//
// Grab 平台官方状态列表及映射：
// - NEW/PENDING: 新订单，待接单 → TakeoutOrderStatePending (0)
// - ACCEPTED/PREPARING/READY: 已接单，配餐中 → TakeoutOrderStateAccepted (1)
// - ALLOCATING/DRIVER_ALLOCATED: 待骑手接单/骑手已分配 → TakeoutOrderStateRiderPending (2)
// - DRIVER_ARRIVED/COLLECTED/DELIVERING: 骑手已到店/已取餐/配送中 → TakeoutOrderStateRiderProcessing (3)
// - DELIVERED/COMPLETED/BILL_PAID: 已送达/已完成/已支付 → TakeoutOrderStateCompleted (4)
// - CANCELLED/REJECTED/FAILED/REFUNDED: 已取消/拒单/失败/退款 → TakeoutOrderStateRejected (5)
//
// Grab 官方状态说明：
// DRIVER_ALLOCATED - Driver has been allocated
// DRIVER_ARRIVED - Driver has reached your store to collect the order
// COLLECTED - Driver has collected the order from your store
// DELIVERED - Driver has delivered the order to the consumer location
// COMPLETED - Order has been completed successfully
// BILL_PAID - When the order is paid by diner via Grab, the value will be BILL_PAID
// REFUNDED - Order has been refunded to the consumer
// CANCELLED - Order has been cancelled by the consumer, merchant, or driver for some reason
// FAILED - The order might fail because of unallocation, reallocation, system issues, etc
func ConvertPlatformStateToOrderState(platformState string, ttposOrderState int) int {
	// 转换为大写进行匹配（容错）
	state := strings.ToUpper(strings.TrimSpace(platformState))

	switch state {
	case "NEW", "PENDING":
		return ttposOrderState // 不合法状态，不能从已接单配餐中到店

	case "ACCEPTED", "PREPARING", "READY":
		return valueobject.TakeoutOrderStateAccepted // 1 - 已接单配餐中

	// case "DRIVER_ARRIVED", "DRIVER_ALLOCATED":
	// 	return valueobject.TakeoutOrderStateRiderPending // 2 - 待骑手接单/骑手已分配

	case "COLLECTED":
		return valueobject.TakeoutOrderStateRiderProcessing // 3 - 骑手配送中（已到店/已取餐/配送中）

	case "DELIVERED":
		return valueobject.TakeoutOrderStateCompleted // 4 - 已完成（已送达/已完成/已支付）

	case "REJECTED":
		return valueobject.TakeoutOrderStateRejected // 5 - 已拒单

	case "CANCELLED", "CANCELED", "FAILED", "REFUNDED":
		return valueobject.TakeoutOrderStateCanceled // 6 - 已取消

	default:
		// 未知状态，默认为待接单
		return ttposOrderState
	}
}

// ParseOrderWebhook 解析 Grab 订单数据
//
// Grab 的订单数据有两种可能的格式：
// 1. 直接的订单对象（最常见，包括新订单推送、订单查询等场景）
// 2. Webhook 包装格式（包含 order 字段，用于某些特定的 Webhook 通知）
//
// 本方法会自动识别并处理这两种格式，统一返回 GrabOrderWebhook 结构
func (c *LineManConverter) ParseOrderWebhook(rawData []byte) (interface{}, error) {
	// 预处理：处理 additionalProperties 字段类型不匹配的情况
	// Grab SDK 期望 map[string]interface{}，但实际可能是 string 或 array
	var tempMap map[string]interface{}
	var additionalPropsValue interface{} // 保存原始值

	if err := json.Unmarshal(rawData, &tempMap); err == nil {
		if addProps, exists := tempMap["additionalItems"]; exists {
			additionalPropsValue = addProps // 保存原始值供后续使用

			// 检查类型，如果不是 map，则删除该字段避免解析错误
			switch addProps.(type) {
			case map[string]interface{}:
				// 正确的类型，不需要处理
			default:
				delete(tempMap, "additionalItems")
				// 重新序列化
				var err error
				rawData, err = json.Marshal(tempMap)
				if err != nil {
					return nil, fmt.Errorf("预处理 Grab 订单数据失败: %w", err)
				}
			}
		}
	}

	// 先尝试解析为 SubmitOrderRequest（Grab SDK 类型）
	var takeoutOrder apiMessage.TakeoutOrder
	if err := json.Unmarshal(rawData, &takeoutOrder); err != nil {
		return nil, fmt.Errorf("解析 Grab 订单数据失败: %w", err)
	}

	// 如果原始 additionalProperties 是数组，手动恢复并转换为 map 格式
	if additionalPropsValue != nil {
		if arrayValue, ok := additionalPropsValue.([]interface{}); ok && len(arrayValue) > 0 {
			// 初始化 map（避免 nil map 赋值 panic）
			if takeoutOrder.AdditionalProperties == nil {
				takeoutOrder.AdditionalProperties = make(map[string]interface{})
			}
			for index, value := range arrayValue {
				if valueMap, ok := value.(map[string]interface{}); ok {
					if name, ok := valueMap["name"].(string); ok {
						takeoutOrder.AdditionalProperties[name+"_"+strconv.Itoa(index)] = valueMap["name"]
					}
				}
			}
		}
	}

	// 检查是否解析成功（通过必填字段判断）
	if takeoutOrder.OrderID != "" {
		return &takeoutOrder, nil
	}

	return nil, fmt.Errorf("LINE MAN 订单数据格式错误：缺少必填字段 orderID 或 merchantID")
}

// ConvertOrderToTakeoutOrder 将 LINE MAN 订单转换为通用外卖订单格式
func (c *LineManConverter) ConvertOrderToTakeoutOrder(
	orderUuid uint64,
	platform string,
	platformOrderId string,
	platformOrder interface{},
	rawDataJSON []byte,
	currentTime int64,
) (*takeoutModel.TakeoutOrder, error) {
	// 类型断言 - 支持 SubmitOrderRequest
	takeoutOrder, ok := platformOrder.(*apiMessage.TakeoutOrder)
	if !ok {
		return nil, errors.New("平台订单数据类型错误，期望 *apiMessage.TakeoutOrder")
	}

	order := &takeoutModel.TakeoutOrder{
		BaseModel: takeoutModel.BaseModel{
			Uuid:       orderUuid,
			CreateTime: currentTime,
			UpdateTime: currentTime,
		},
		Platform:        platform,
		PlatformOrderId: platformOrderId,
		PlatformOrderState: func() string {
			if takeoutOrder.GetOrderState() == "" {
				return "NEW"
			}
			return takeoutOrder.GetOrderState()
		}(),
		OrderState:  ConvertPlatformStateToOrderState(takeoutOrder.GetOrderState(), valueobject.TakeoutOrderStatePending), // 使用状态转换函数
		IsAbnormal:  0,
		StockStatus: 0, // 库存充足
		RawData:     string(rawDataJSON),
	}

	// 基础字段映射
	order.ShortOrderNumber = takeoutOrder.GetShortOrderNumber()
	order.MerchantId = takeoutOrder.GetMerchantID()

	if partnerMerchantID, ok := takeoutOrder.GetPartnerMerchantIDOk(); ok {
		order.PartnerMerchantId = *partnerMerchantID
	}

	order.PaymentType = takeoutOrder.GetPaymentType()

	featureFlags := takeoutOrder.GetFeatureFlags()
	// 转换 Grab 订单类型为 TTPOS 通用订单类型
	order.OrderType = valueobject.ConvertGrabOrderTypeToTakeoutOrderType(featureFlags.OrderType)
	order.OrderAcceptedType = featureFlags.OrderAcceptedType

	if membershipID, ok := takeoutOrder.GetMembershipIDOk(); ok {
		order.MembershipId = *membershipID
	}

	// 布尔字段转整数
	if takeoutOrder.GetCutlery() {
		order.Cutlery = 1
	}
	if featureFlags.HasIsMexEditOrder() && featureFlags.GetIsMexEditOrder() {
		order.IsMexEditOrder = 1
	}
	// 价格信息映射
	price := takeoutOrder.GetPrice()
	order.Subtotal = decimal.NewFromInt(int64(price.GetSubtotal())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
	order.DeliveryFee = decimal.NewFromInt(int64(price.GetDeliveryFee())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
	order.SmallOrderFee = decimal.NewFromInt(int64(price.GetSmallOrderFee())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
	order.PlatformDiscount = decimal.NewFromInt(int64(price.GetGrabFundPromo())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
	order.MerchantDiscount = decimal.NewFromInt(int64(price.GetMerchantFundPromo())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
	order.BasketPromo = decimal.NewFromInt(int64(price.GetBasketPromo())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
	order.Tax = decimal.NewFromInt(int64(price.GetTax())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
	order.MerchantChargeFee = decimal.NewFromInt(int64(price.GetMerchantChargeFee())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
	order.EaterPayment = decimal.NewFromInt(int64(price.GetEaterPayment())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
	// 计算平台结算总额: subtotal + merchant_charge_fee - merchant_discount
	order.PlatformTotal = decimal.NewFromFloat(order.Subtotal).
		Add(decimal.NewFromFloat(order.MerchantChargeFee)).
		Sub(decimal.NewFromFloat(order.MerchantDiscount)).
		InexactFloat64()

	// 货币信息映射
	currency := takeoutOrder.GetCurrency()
	order.CurrencyCode = currency.Code
	order.CurrencySymbol = currency.Symbol
	order.CurrencyExponent = int(currency.Exponent)

	// 时间字段映射（RFC3339 格式转 Unix 时间戳）
	order.OrderTime = c.parseRFC3339Time(takeoutOrder.GetOrderTime())

	// LINE MAN 平台特殊处理：订单接受时间使用订单时间
	if order.Platform == valueobject.TakeoutPlatformLineman {
		order.AcceptedTime = order.OrderTime
	}

	if takeoutOrder.HasSubmitTime() {
		order.SubmitTime = takeoutOrder.GetSubmitTime().Unix()
	}

	if scheduledTime, ok := takeoutOrder.GetScheduledTimeOk(); ok {
		order.ScheduledTime = c.parseRFC3339Time(*scheduledTime)
	}

	// 预计准备时间
	if takeoutOrder.HasOrderReadyEstimation() {
		estimation := takeoutOrder.GetOrderReadyEstimation()
		order.EstimatedReadyTime = estimation.EstimatedOrderReadyTime.Unix()
		order.MaxReadyTime = estimation.MaxOrderReadyTime.Unix()
	}

	// 附加项目映射 (additionalProperties -> additional_properties)
	// Grab SDK 的 AdditionalProperties 是 map[string]interface{}
	if len(takeoutOrder.AdditionalProperties) > 0 {
		additionalItemNames := make([]string, 0, len(takeoutOrder.AdditionalProperties))
		for _, value := range takeoutOrder.AdditionalProperties {
			if value != nil {
				// 将值转换为字符串
				valueStr := fmt.Sprintf("%v", value)
				if valueStr != "" {
					additionalItemNames = append(additionalItemNames, valueStr)
				}
			}
		}
		if len(additionalItemNames) > 0 {
			order.AdditionalProperties = strings.Join(additionalItemNames, ", ")
		}
	}

	// 解析商品数据并填充到 order.TakeoutOrderItems
	if len(takeoutOrder.Items) > 0 {
		order.TakeoutOrderItems = make([]takeoutModel.TakeoutOrderItem, 0, len(takeoutOrder.Items))
		for _, item := range takeoutOrder.Items {
			// 根据 item.Id 前缀判断商品类型
			// TTPOS-ITEM- 开头为普通商品(0), TTPOS-PACKAGE- 开头为套餐(1)
			ttposProductType := 0
			if strings.HasPrefix(item.ID, "TTPOS-PACKAGE-") {
				ttposProductType = 1
			}

			// 注意：UUID、CreateTime、UpdateTime 会在 Service 层设置
			// Grab API 不提供商品名称，这里暂时使用 item.Id 作为名称
			orderItem := takeoutModel.TakeoutOrderItem{
				Platform:         platform,
				PlatformItemId:   item.ID,
				TtposProductType: ttposProductType,
				Quantity:         int(item.GetQuantity()),
				Price:            decimal.NewFromInt(int64(item.GetPrice())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64(), // Grab API 返回的是分（exponent=2），转换为元
				Tax:              decimal.NewFromInt(int64(item.GetTax())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64(),   // 同上
				Specifications:   item.GetSpecifications(),
			}

			// 解析修饰符
			if len(item.Modifiers) > 0 {
				orderItem.TakeoutOrderItemModifiers = make([]takeoutModel.TakeoutOrderItemModifier, 0, len(item.Modifiers))
				for _, modifier := range item.Modifiers {
					// Grab API 不提供修饰符名称，这里暂时使用 modifier.GetId() 作为名称
					orderItemModifier := takeoutModel.TakeoutOrderItemModifier{
						Platform:           platform,
						PlatformModifierId: modifier.GetId(),
						Quantity:           int(modifier.GetQuantity()),
						Price:              decimal.NewFromInt(int64(modifier.GetPrice())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64(), // Grab API 返回的是分（exponent=2），转换为元
						Tax:                decimal.NewFromInt(int64(modifier.GetTax())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64(),   // 同上
					}
					orderItem.TakeoutOrderItemModifiers = append(orderItem.TakeoutOrderItemModifiers, orderItemModifier)
				}
			}

			order.TakeoutOrderItems = append(order.TakeoutOrderItems, orderItem)
		}
	}

	// 转换收货人信息
	receiver, err := c.ConvertReceiverInfo(orderUuid, platform, takeoutOrder, currentTime)
	if err != nil {
		return nil, errors.WithMessage(err, "转换收货人信息失败")
	}
	order.SetTakeoutOrderReceiver(receiver)

	// 转换活动信息
	campaigns, err := c.ConvertCampaigns(orderUuid, platform, takeoutOrder, currentTime)
	if err != nil {
		return nil, errors.WithMessage(err, "转换活动信息失败")
	}
	order.SetTakeoutOrderCampaigns(campaigns)

	// 转换促销信息
	promos, err := c.ConvertOrderPromos(orderUuid, platform, takeoutOrder, currentTime)
	if err != nil {
		return nil, errors.WithMessage(err, "转换促销信息失败")
	}
	order.SetTakeoutOrderPromos(promos)

	// 计算订单小计金额
	order.Subtotal = 0.0
	for _, item := range order.TakeoutOrderItems {
		order.Subtotal = decimal.NewFromFloat(order.Subtotal).Add(decimal.NewFromFloat(item.GetTotalPrice())).InexactFloat64()
	}

	return order, nil
}

// ConvertReceiverInfo 转换收货人信息
func (c *LineManConverter) ConvertReceiverInfo(
	orderUuid uint64,
	platform string,
	takeoutOrder *apiMessage.TakeoutOrder,
	currentTime int64,
) (*takeoutModel.TakeoutOrderReceiver, error) {
	// 检查是否有收货人信息
	if !takeoutOrder.HasReceiver() {
		return nil, nil // 没有收货人信息不是错误，返回 nil
	}

	receiver := takeoutOrder.GetReceiver()

	receiverInfo := &takeoutModel.TakeoutOrderReceiver{
		BaseModel: takeoutModel.BaseModel{
			CreateTime: currentTime,
			UpdateTime: currentTime,
		},
		TakeoutOrderUuid: orderUuid,
		Platform:         platform,
	}

	// 收货人姓名
	if receiver.Name != nil {
		receiverInfo.ReceiverName = *receiver.Name
	}

	// 收货人电话
	if receiver.Phones != nil {
		receiverInfo.ReceiverPhones = *receiver.Phones
	}

	// 地址信息
	if receiver.Address != nil {
		address := receiver.Address

		if address.UnitNumber != nil {
			receiverInfo.UnitNumber = *address.UnitNumber
		}

		if address.DeliveryInstruction != nil {
			receiverInfo.DeliveryInstruction = *address.DeliveryInstruction
		}

		if address.PoiSource != nil {
			receiverInfo.PoiSource = *address.PoiSource
		}

		if address.PoiID != nil {
			receiverInfo.PoiID = *address.PoiID
		}

		if address.Address != nil {
			receiverInfo.Address = *address.Address
		}

		if address.Postcode != nil {
			receiverInfo.Postcode = *address.Postcode
		}

		// 坐标信息
		if address.Coordinates != nil {
			coordinates := address.Coordinates

			if coordinates.Latitude != nil {
				receiverInfo.Latitude = *coordinates.Latitude
			}

			if coordinates.Longitude != nil {
				receiverInfo.Longitude = *coordinates.Longitude
			}
		}
	}

	return receiverInfo, nil
}

// parseRFC3339Time 解析 RFC3339 格式时间为 Unix 时间戳
func (c *LineManConverter) parseRFC3339Time(timeStr string) int64 {
	if timeStr == "" {
		return 0
	}

	// 判断如果是 2026-01-23 09:26:59 格式，直接解析
	if strings.Contains(timeStr, " ") && !strings.Contains(timeStr, "T") {
		// 尝试解析标准日期时间格式
		t, err := time.Parse("2006-01-02 15:04:05", timeStr)
		if err == nil {
			return t.Unix()
		}
		// 尝试带毫秒的格式
		t, err = time.Parse("2006-01-02 15:04:05.000", timeStr)
		if err == nil {
			return t.Unix()
		}
	}

	// 尝试多种时间格式
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05.000-07:00",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t.Unix()
		}
	}

	return 0
}

// ConvertCampaigns 将 Grab 订单活动转换为 TTPOS 订单活动格式
func (c *LineManConverter) ConvertCampaigns(
	orderUuid uint64,
	platform string,
	takeoutOrder *apiMessage.TakeoutOrder,
	currentTime int64,
) ([]*takeoutModel.TakeoutOrderCampaign, error) {
	if !takeoutOrder.HasCampaigns() {
		return nil, nil // 没有活动信息
	}

	campaigns := takeoutOrder.GetCampaigns()
	if len(campaigns) == 0 {
		return nil, nil
	}

	orderCampaigns := make([]*takeoutModel.TakeoutOrderCampaign, 0, len(campaigns))

	for _, campaign := range campaigns {
		orderCampaign := &takeoutModel.TakeoutOrderCampaign{
			BaseModel: takeoutModel.BaseModel{
				CreateTime: currentTime,
				UpdateTime: currentTime,
			},
			TakeoutOrderUuid: orderUuid,
			Platform:         platform,
		}

		// 活动基本信息
		if campaign.ID != nil {
			orderCampaign.CampaignID = *campaign.ID
		}
		if campaign.Name != nil {
			orderCampaign.CampaignName = *campaign.Name
		}
		if campaign.CampaignNameForMex != nil {
			orderCampaign.CampaignNameForMex = *campaign.CampaignNameForMex
		}
		if campaign.Level != nil {
			orderCampaign.CampaignLevel = *campaign.Level
		}
		if campaign.Type != nil {
			orderCampaign.CampaignType = *campaign.Type
		}

		// 活动使用信息
		if campaign.UsageCount != nil {
			orderCampaign.UsageCount = int32(*campaign.UsageCount)
		}
		if campaign.MexFundedRatio != nil {
			orderCampaign.MexFundedRatio = *campaign.MexFundedRatio
		}
		if campaign.DeductedAmount != nil {
			orderCampaign.DeductedAmount = decimal.NewFromInt(*campaign.DeductedAmount).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
		}
		if campaign.DeductedPart != nil {
			orderCampaign.DeductedPart = *campaign.DeductedPart
		}

		// 应用的商品ID列表 (转换为JSON字符串)
		if campaign.AppliedItemIDs != nil {
			itemIDs := campaign.AppliedItemIDs
			if len(itemIDs) > 0 {
				if jsonData, err := json.Marshal(itemIDs); err == nil {
					orderCampaign.AppliedItemIDs = string(jsonData)
				}
			}
		}

		// 赠品信息
		if campaign.FreeItem != nil {
			freeItem := campaign.FreeItem
			if freeItem.ID != nil {
				orderCampaign.FreeItemID = *freeItem.ID
			}
			if freeItem.Name != nil {
				orderCampaign.FreeItemName = *freeItem.Name
			}
			if freeItem.Quantity != nil {
				orderCampaign.FreeItemQuantity = *freeItem.Quantity
			}
			if freeItem.Price != nil {
				orderCampaign.FreeItemPrice = *freeItem.Price
			}
		}

		orderCampaigns = append(orderCampaigns, orderCampaign)
	}

	return orderCampaigns, nil
}

// ConvertOrderPromos 将 Grab 订单促销转换为 TTPOS 订单促销格式
func (c *LineManConverter) ConvertOrderPromos(
	orderUuid uint64,
	platform string,
	takeoutOrder *apiMessage.TakeoutOrder,
	currentTime int64,
) ([]*takeoutModel.TakeoutOrderPromo, error) {
	if !takeoutOrder.HasPromos() {
		return nil, nil // 没有促销信息
	}

	promos := takeoutOrder.GetPromos()
	if len(promos) == 0 {
		return nil, nil
	}

	orderPromos := make([]*takeoutModel.TakeoutOrderPromo, 0, len(promos))

	for _, promo := range promos {
		orderPromo := &takeoutModel.TakeoutOrderPromo{
			BaseModel: takeoutModel.BaseModel{
				CreateTime: currentTime,
				UpdateTime: currentTime,
			},
			TakeoutOrderUuid: fmt.Sprintf("%d", orderUuid),
			Platform:         platform,
		}

		// 促销基本信息
		if promo.Code != nil {
			orderPromo.PromoCode = *promo.Code
		}
		if promo.Name != nil {
			orderPromo.PromoName = *promo.Name
		}
		if promo.Description != nil {
			orderPromo.PromoDescription = *promo.Description
		}

		// 金额信息
		if promo.PromoAmountInMin != nil {
			orderPromo.PromoAmount = decimal.NewFromInt(*promo.PromoAmountInMin).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
			orderPromo.PromoAmountInMin = decimal.NewFromInt(*promo.PromoAmountInMin).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
		} else if promo.PromoAmount != nil {
			// 如果没有 PromoAmountInMin，使用 PromoAmount（需要转换为最小单位）
			orderPromo.PromoAmount = decimal.NewFromInt(*promo.PromoAmount).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
			orderPromo.PromoAmountInMin = decimal.NewFromInt(*promo.PromoAmount).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
		}

		if promo.MexFundedRatio != nil {
			orderPromo.MexFundedRatio = int32(*promo.MexFundedRatio)
		}
		if promo.MexFundedAmount != nil {
			orderPromo.MexFundedAmount = decimal.NewFromInt(*promo.MexFundedAmount).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
		}
		if promo.TargetedPrice != nil {
			orderPromo.TargetedPrice = decimal.NewFromInt(*promo.TargetedPrice).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
		}

		orderPromos = append(orderPromos, orderPromo)
	}

	return orderPromos, nil
}
