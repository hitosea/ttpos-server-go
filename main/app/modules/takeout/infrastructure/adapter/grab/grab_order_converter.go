package grab

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"ttpos-server-go/app/errors"
	takeoutModel "ttpos-server-go/app/modules/takeout/domain/model"
	valueobject "ttpos-server-go/app/modules/takeout/domain/value_object"

	grabfood "github.com/grab/grabfood-api-sdk-go"
	"github.com/shopspring/decimal"
)

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
		if ttposOrderState == valueobject.TakeoutOrderStateAccepted {
			return ttposOrderState // 不合法状态，不能从已接单配餐中到店
		}
		return valueobject.TakeoutOrderStateRiderProcessing // 3 - 骑手配送中（已到店/已取餐/配送中）

	case "DELIVERED":
		if ttposOrderState == valueobject.TakeoutOrderStateAccepted {
			return ttposOrderState // 不合法状态，不能从已接单配餐中到店
		}
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
func (c *GrabConverter) ParseOrderWebhook(rawData []byte) (interface{}, error) {
	// 先尝试解析为 SubmitOrderRequest（Grab SDK 类型）
	var submitOrderReq grabfood.SubmitOrderRequest
	if err := json.Unmarshal(rawData, &submitOrderReq); err != nil {
		return nil, fmt.Errorf("解析 Grab 订单数据失败: %w", err)
	}

	// 检查是否解析成功（通过必填字段判断）
	if submitOrderReq.OrderID != "" && submitOrderReq.MerchantID != "" {
		return &submitOrderReq, nil
	}

	// 如果不是 SubmitOrderRequest 格式，尝试解析为 OrderStateRequest（状态更新）
	var orderStateReq grabfood.OrderStateRequest
	if err := json.Unmarshal(rawData, &orderStateReq); err != nil {
		return nil, fmt.Errorf("解析 Grab 订单状态数据失败: %w", err)
	}

	// 验证 OrderStateRequest
	if orderStateReq.OrderID != "" && orderStateReq.MerchantID != "" {
		return &orderStateReq, nil
	}

	return nil, fmt.Errorf("Grab 订单数据格式错误：无法识别的订单格式")
}

// ConvertOrderToTakeoutOrder 将 Grab 订单转换为通用外卖订单格式
func (c *GrabConverter) ConvertOrderToTakeoutOrder(
	orderUuid uint64,
	platform string,
	platformOrderId string,
	platformOrder interface{},
	rawDataJSON []byte,
	currentTime int64,
) (*takeoutModel.TakeoutOrder, error) {
	// 类型断言 - 支持 SubmitOrderRequest
	submitOrderReq, ok := platformOrder.(*grabfood.SubmitOrderRequest)
	if !ok {
		return nil, errors.New("平台订单数据类型错误，期望 *grabfood.SubmitOrderRequest")
	}

	order := &takeoutModel.TakeoutOrder{
		BaseModel: takeoutModel.BaseModel{
			Uuid:       orderUuid,
			CreateTime: currentTime,
			UpdateTime: currentTime,
		},
		Platform:           platform,
		PlatformOrderId:    platformOrderId,
		PlatformOrderState: submitOrderReq.GetOrderState(),
		OrderState:         ConvertPlatformStateToOrderState(submitOrderReq.GetOrderState(), valueobject.TakeoutOrderStatePending), // 使用状态转换函数
		IsAbnormal:         0,
		StockStatus:        0, // 库存充足
		RawData:            string(rawDataJSON),
	}

	// 基础字段映射
	order.ShortOrderNumber = submitOrderReq.GetShortOrderNumber()
	order.MerchantId = submitOrderReq.GetMerchantID()

	if partnerMerchantID, ok := submitOrderReq.GetPartnerMerchantIDOk(); ok {
		order.PartnerMerchantId = *partnerMerchantID
	}

	order.PaymentType = submitOrderReq.GetPaymentType()

	featureFlags := submitOrderReq.GetFeatureFlags()
	// 转换 Grab 订单类型为 TTPOS 通用订单类型
	order.OrderType = valueobject.ConvertGrabOrderTypeToTakeoutOrderType(featureFlags.OrderType)
	order.OrderAcceptedType = featureFlags.OrderAcceptedType

	if membershipID, ok := submitOrderReq.GetMembershipIDOk(); ok {
		order.MembershipId = *membershipID
	}

	// 布尔字段转整数
	if submitOrderReq.GetCutlery() {
		order.Cutlery = 1
	}
	if featureFlags.HasIsMexEditOrder() && featureFlags.GetIsMexEditOrder() {
		order.IsMexEditOrder = 1
	}

	// 价格信息映射
	price := submitOrderReq.GetPrice()
	order.Subtotal = decimal.NewFromInt(int64(price.GetSubtotal())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
	order.DeliveryFee = decimal.NewFromInt(int64(price.GetDeliveryFee())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
	order.SmallOrderFee = decimal.NewFromInt(int64(price.GetSmallOrderFee())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
	order.PlatformDiscount = decimal.NewFromInt(int64(price.GetGrabFundPromo())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
	order.MerchantDiscount = decimal.NewFromInt(int64(price.GetMerchantFundPromo())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
	order.BasketPromo = decimal.NewFromInt(int64(price.GetBasketPromo())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
	order.Tax = decimal.NewFromInt(int64(price.GetTax())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
	order.MerchantChargeFee = decimal.NewFromInt(int64(price.GetMerchantChargeFee())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
	order.EaterPayment = func() float64 {
		if order.PaymentType == "CASH" || featureFlags.OrderType == "DeliveredByRestaurant" {
			return decimal.NewFromInt(int64(price.GetEaterPayment())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
		}
		return order.Subtotal
	}()

	// 货币信息映射
	currency := submitOrderReq.GetCurrency()
	order.CurrencyCode = currency.GetCode()
	order.CurrencySymbol = currency.GetSymbol()
	order.CurrencyExponent = int(currency.GetExponent())

	// 时间字段映射（RFC3339 格式转 Unix 时间戳）
	order.OrderTime = c.parseRFC3339Time(submitOrderReq.GetOrderTime())

	if submitOrderReq.HasSubmitTime() {
		order.SubmitTime = submitOrderReq.GetSubmitTime().Unix()
	}

	if scheduledTime, ok := submitOrderReq.GetScheduledTimeOk(); ok {
		order.ScheduledTime = c.parseRFC3339Time(*scheduledTime)
	}

	// 预计准备时间
	if submitOrderReq.HasOrderReadyEstimation() {
		estimation := submitOrderReq.GetOrderReadyEstimation()
		order.EstimatedReadyTime = estimation.GetEstimatedOrderReadyTime().Unix()
		order.MaxReadyTime = estimation.GetMaxOrderReadyTime().Unix()
	}

	// 解析商品数据并填充到 order.TakeoutOrderItems
	if len(submitOrderReq.Items) > 0 {
		order.TakeoutOrderItems = make([]takeoutModel.TakeoutOrderItem, 0, len(submitOrderReq.Items))
		for _, item := range submitOrderReq.Items {
			// 根据 item.Id 前缀判断商品类型
			// TTPOS-ITEM- 开头为普通商品(0), TTPOS-PACKAGE- 开头为套餐(1)
			ttposProductType := 0
			if strings.HasPrefix(item.Id, "TTPOS-PACKAGE-") {
				ttposProductType = 1
			}

			// 注意：UUID、CreateTime、UpdateTime 会在 Service 层设置
			// Grab API 不提供商品名称，这里暂时使用 item.Id 作为名称
			orderItem := takeoutModel.TakeoutOrderItem{
				Platform:         platform,
				PlatformItemId:   item.Id,
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
	receiver, err := c.ConvertReceiverInfo(orderUuid, platform, submitOrderReq, currentTime)
	if err != nil {
		return nil, errors.WithMessage(err, "转换收货人信息失败")
	}
	order.SetTakeoutOrderReceiver(receiver)

	// 转换活动信息
	campaigns, err := c.ConvertCampaigns(orderUuid, platform, submitOrderReq, currentTime)
	if err != nil {
		return nil, errors.WithMessage(err, "转换活动信息失败")
	}
	order.SetTakeoutOrderCampaigns(campaigns)

	// 转换促销信息
	promos, err := c.ConvertOrderPromos(orderUuid, platform, submitOrderReq, currentTime)
	if err != nil {
		return nil, errors.WithMessage(err, "转换促销信息失败")
	}
	order.SetTakeoutOrderPromos(promos)

	return order, nil
}

// ConvertReceiverInfo 转换收货人信息
func (c *GrabConverter) ConvertReceiverInfo(
	orderUuid uint64,
	platform string,
	submitOrderReq *grabfood.SubmitOrderRequest,
	currentTime int64,
) (*takeoutModel.TakeoutOrderReceiver, error) {
	// 检查是否有收货人信息
	if !submitOrderReq.HasReceiver() {
		return nil, nil // 没有收货人信息不是错误，返回 nil
	}

	receiver := submitOrderReq.GetReceiver()

	receiverInfo := &takeoutModel.TakeoutOrderReceiver{
		BaseModel: takeoutModel.BaseModel{
			CreateTime: currentTime,
			UpdateTime: currentTime,
		},
		TakeoutOrderUuid: orderUuid,
		Platform:         platform,
	}

	// 收货人姓名
	if receiver.HasName() {
		receiverInfo.ReceiverName = receiver.GetName()
	}

	// 收货人电话
	if receiver.HasPhones() {
		receiverInfo.ReceiverPhones = receiver.GetPhones()
	}

	// 地址信息
	if receiver.HasAddress() {
		address := receiver.GetAddress()

		if address.HasUnitNumber() {
			receiverInfo.UnitNumber = address.GetUnitNumber()
		}

		if address.HasDeliveryInstruction() {
			receiverInfo.DeliveryInstruction = address.GetDeliveryInstruction()
		}

		if address.HasPoiSource() {
			receiverInfo.PoiSource = address.GetPoiSource()
		}

		if address.HasPoiID() {
			receiverInfo.PoiID = address.GetPoiID()
		}

		if address.HasAddress() {
			receiverInfo.Address = address.GetAddress()
		}

		if address.HasPostcode() {
			receiverInfo.Postcode = address.GetPostcode()
		}

		// 坐标信息
		if address.HasCoordinates() {
			coordinates := address.GetCoordinates()

			if coordinates.HasLatitude() {
				receiverInfo.Latitude = coordinates.GetLatitude()
			}

			if coordinates.HasLongitude() {
				receiverInfo.Longitude = coordinates.GetLongitude()
			}
		}
	}

	return receiverInfo, nil
}

// parseRFC3339Time 解析 RFC3339 格式时间为 Unix 时间戳
func (c *GrabConverter) parseRFC3339Time(timeStr string) int64 {
	if timeStr == "" {
		return 0
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
func (c *GrabConverter) ConvertCampaigns(
	orderUuid uint64,
	platform string,
	submitOrderReq *grabfood.SubmitOrderRequest,
	currentTime int64,
) ([]*takeoutModel.TakeoutOrderCampaign, error) {
	if !submitOrderReq.HasCampaigns() {
		return nil, nil // 没有活动信息
	}

	campaigns := submitOrderReq.GetCampaigns()
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
		if campaign.HasId() {
			orderCampaign.CampaignID = campaign.GetId()
		}
		if campaign.HasName() {
			orderCampaign.CampaignName = campaign.GetName()
		}
		if campaign.HasCampaignNameForMex() {
			orderCampaign.CampaignNameForMex = campaign.GetCampaignNameForMex()
		}
		if campaign.HasLevel() {
			orderCampaign.CampaignLevel = campaign.GetLevel()
		}
		if campaign.HasType() {
			orderCampaign.CampaignType = campaign.GetType()
		}

		// 活动使用信息
		if campaign.HasUsageCount() {
			orderCampaign.UsageCount = int32(campaign.GetUsageCount())
		}
		if campaign.HasMexFundedRatio() {
			orderCampaign.MexFundedRatio = campaign.GetMexFundedRatio()
		}
		if campaign.HasDeductedAmount() {
			orderCampaign.DeductedAmount = decimal.NewFromInt(int64(campaign.GetDeductedAmount())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
		}
		if campaign.HasDeductedPart() {
			orderCampaign.DeductedPart = campaign.GetDeductedPart()
		}

		// 应用的商品ID列表 (转换为JSON字符串)
		if campaign.HasAppliedItemIDs() {
			itemIDs := campaign.GetAppliedItemIDs()
			if len(itemIDs) > 0 {
				if jsonData, err := json.Marshal(itemIDs); err == nil {
					orderCampaign.AppliedItemIDs = string(jsonData)
				}
			}
		}

		// 赠品信息
		if campaign.HasFreeItem() {
			freeItem := campaign.GetFreeItem()
			if freeItem.HasId() {
				orderCampaign.FreeItemID = freeItem.GetId()
			}
			if freeItem.HasName() {
				orderCampaign.FreeItemName = freeItem.GetName()
			}
			if freeItem.HasQuantity() {
				orderCampaign.FreeItemQuantity = freeItem.GetQuantity()
			}
			if freeItem.HasPrice() {
				orderCampaign.FreeItemPrice = freeItem.GetPrice()
			}
		}

		orderCampaigns = append(orderCampaigns, orderCampaign)
	}

	return orderCampaigns, nil
}

// ConvertOrderPromos 将 Grab 订单促销转换为 TTPOS 订单促销格式
func (c *GrabConverter) ConvertOrderPromos(
	orderUuid uint64,
	platform string,
	submitOrderReq *grabfood.SubmitOrderRequest,
	currentTime int64,
) ([]*takeoutModel.TakeoutOrderPromo, error) {
	if !submitOrderReq.HasPromos() {
		return nil, nil // 没有促销信息
	}

	promos := submitOrderReq.GetPromos()
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
		if promo.HasCode() {
			orderPromo.PromoCode = promo.GetCode()
		}
		if promo.HasName() {
			orderPromo.PromoName = promo.GetName()
		}
		if promo.HasDescription() {
			orderPromo.PromoDescription = promo.GetDescription()
		}

		// 金额信息
		if promo.HasPromoAmountInMin() {
			orderPromo.PromoAmount = decimal.NewFromInt(int64(promo.GetPromoAmountInMin())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
			orderPromo.PromoAmountInMin = decimal.NewFromInt(int64(promo.GetPromoAmountInMin())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
		} else if promo.HasPromoAmount() {
			// 如果没有 PromoAmountInMin，使用 PromoAmount（需要转换为最小单位）
			orderPromo.PromoAmount = decimal.NewFromInt(int64(promo.GetPromoAmount())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
			orderPromo.PromoAmountInMin = decimal.NewFromInt(int64(promo.GetPromoAmount())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
		}

		if promo.HasMexFundedRatio() {
			orderPromo.MexFundedRatio = int32(promo.GetMexFundedRatio())
		}
		if promo.HasMexFundedAmount() {
			orderPromo.MexFundedAmount = decimal.NewFromInt(int64(promo.GetMexFundedAmount())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
		}
		if promo.HasTargetedPrice() {
			orderPromo.TargetedPrice = decimal.NewFromInt(int64(promo.GetTargetedPrice())).Div(decimal.NewFromInt(c.amountConversionFactor)).InexactFloat64()
		}

		orderPromos = append(orderPromos, orderPromo)
	}

	return orderPromos, nil
}
