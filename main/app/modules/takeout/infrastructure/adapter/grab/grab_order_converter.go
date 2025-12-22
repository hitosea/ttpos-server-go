package grab

import (
	"encoding/json"
	"fmt"
	"time"
	"ttpos-server-go/app/errors"
	takeoutModel "ttpos-server-go/app/modules/takeout/domain/model"

	grabfood "github.com/grab/grabfood-api-sdk-go"
)

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
		Platform:        platform,
		PlatformOrderId: platformOrderId,
		OrderState:      0, // 待接单
		IsAbnormal:      0,
		StockStatus:     0, // 库存充足
		RawData:         string(rawDataJSON),
	}

	// 基础字段映射
	order.ShortOrderNumber = submitOrderReq.GetShortOrderNumber()
	order.MerchantId = submitOrderReq.GetMerchantID()

	if partnerMerchantID, ok := submitOrderReq.GetPartnerMerchantIDOk(); ok {
		order.PartnerMerchantId = *partnerMerchantID
	}

	order.PaymentType = submitOrderReq.GetPaymentType()

	featureFlags := submitOrderReq.GetFeatureFlags()
	order.OrderType = featureFlags.OrderType
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
	order.Subtotal = price.GetSubtotal()
	order.DeliveryFee = price.GetDeliveryFee()
	order.SmallOrderFee = price.GetSmallOrderFee()
	order.TotalAmount = price.GetEaterPayment() // 使用 EaterPayment 作为总金额
	order.EaterPayment = price.GetEaterPayment()
	order.PlatformDiscount = price.GetGrabFundPromo()
	order.MerchantDiscount = price.GetMerchantFundPromo()
	order.BasketPromo = price.GetBasketPromo()
	order.Tax = price.GetTax()
	order.MerchantChargeFee = price.GetMerchantChargeFee()

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
			// 将 Grab item 转换为 JSON 保存在 PlatformData
			itemBytes, _ := json.Marshal(item)

			// 注意：UUID、CreateTime、UpdateTime 会在 Service 层设置
			orderItem := takeoutModel.TakeoutOrderItem{
				Platform:       platform,
				PlatformItemId: item.Id,
				Quantity:       int(item.Quantity),
				Price:          item.Price,
				Tax:            item.GetTax(),
				Specifications: item.GetSpecifications(),
				PlatformData:   string(itemBytes),
			}

			// 解析修饰符
			if len(item.Modifiers) > 0 {
				orderItem.TakeoutOrderItemModifiers = make([]takeoutModel.TakeoutOrderItemModifier, 0, len(item.Modifiers))
				for _, modifier := range item.Modifiers {
					modifierBytes, _ := json.Marshal(modifier)
					orderItemModifier := takeoutModel.TakeoutOrderItemModifier{
						Platform:           platform,
						PlatformModifierId: modifier.GetId(),
						Quantity:           int(modifier.GetQuantity()),
						Price:              modifier.GetPrice(),
						Tax:                modifier.GetTax(),
						PlatformData:       string(modifierBytes),
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
			orderCampaign.DeductedAmount = campaign.GetDeductedAmount()
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
