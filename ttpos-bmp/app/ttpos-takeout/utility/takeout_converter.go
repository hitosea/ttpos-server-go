package utility

import (
	"time"

	"ttpos-api/ttpos-takeout/message"
	linemanv1 "ttpos-bmp/app/ttpos-takeout/api/lineman/v1"

	grabsdk "github.com/grab/grabfood-api-sdk-go"
)

// ============================================================================
// Grab SDK 转换方法
// ============================================================================

// FromGrabSDK 从 Grab SDK 订单请求转换为统一模型
func FromGrabSDK(req *grabsdk.SubmitOrderRequest) (*message.TakeoutOrder, error) {
	if req == nil {
		return nil, nil
	}

	order := &message.TakeoutOrder{
		OrderID:           req.OrderID,
		ShortOrderNumber:  req.ShortOrderNumber,
		MerchantID:        req.MerchantID,
		PartnerMerchantID: req.PartnerMerchantID,
		PaymentType:       req.PaymentType,
		Cutlery:           req.Cutlery,
		OrderTime:         req.OrderTime,
	}

	// SubmitTime, CompleteTime
	if req.SubmitTime != nil {
		order.SubmitTime = req.SubmitTime
	}

	if req.CompleteTime != nil {
		order.CompleteTime = req.CompleteTime
	}

	if req.ScheduledTime != nil {
		order.ScheduledTime = req.ScheduledTime
	}

	if req.OrderState != nil {
		order.OrderState = req.OrderState
	}

	if req.MembershipID != nil {
		order.MembershipID = req.MembershipID
	}

	// Currency (required in new struct)
	order.Currency = message.TakeoutCurrency{
		Code:     req.Currency.Code,
		Symbol:   req.Currency.Symbol,
		Exponent: req.Currency.Exponent,
	}

	// FeatureFlags (required in new struct)
	order.FeatureFlags = message.TakeoutFeatureFlags{
		OrderAcceptedType: req.FeatureFlags.OrderAcceptedType,
		IsMexEditOrder:    req.FeatureFlags.IsMexEditOrder,
		OrderType:         req.FeatureFlags.OrderType,
	}

	// Items
	for _, item := range req.Items {
		takeoutItem := message.TakeoutOrderItem{
			ID:         item.Id,
			GrabItemID: item.GrabItemID,
			Quantity:   item.Quantity,
			Price:      item.Price, // Already in minor unit from Grab SDK
		}

		if item.Tax != nil {
			takeoutItem.Tax = item.Tax
		}

		if item.Specifications != nil {
			takeoutItem.Specifications = item.Specifications
		}

		// Modifiers
		for _, mod := range item.Modifiers {
			modifier := message.TakeoutModifier{
				ID:       mod.Id,
				Price:    mod.Price,
				Tax:      mod.Tax,
				Quantity: mod.Quantity,
			}
			takeoutItem.Modifiers = append(takeoutItem.Modifiers, modifier)
		}

		// OutOfStockInstruction
		if item.OutOfStockInstruction.IsSet() {
			inst := item.OutOfStockInstruction.Get()
			if inst != nil {
				takeoutInst := &message.TakeoutOutOfStockInstruction{
					Title:                 inst.Title,
					InstructionType:       inst.InstructionType,
					ReplacementItemID:     inst.ReplacementItemID,
					ReplacementGrabItemID: inst.ReplacementGrabItemID,
				}
				takeoutItem.OutOfStockInstruction.Set(takeoutInst)
			}
		}

		order.Items = append(order.Items, takeoutItem)
	}

	// Price (required in new struct, all values in minor unit)
	// 计算 Total: Subtotal + MerchantChargeFee - MerchantFundPromo
	// 注意: MerchantChargeFee 和 MerchantFundPromo 是可选的 *int64 类型
	var merchantChargeFee, merchantFundPromo int64
	if req.Price.MerchantChargeFee != nil {
		merchantChargeFee = *req.Price.MerchantChargeFee
	}
	if req.Price.MerchantFundPromo != nil {
		merchantFundPromo = *req.Price.MerchantFundPromo
	}
	calculatedTotal := req.Price.Subtotal + merchantChargeFee - merchantFundPromo

	order.Price = message.TakeoutOrderPrice{
		Subtotal:          req.Price.Subtotal,
		Tax:               req.Price.Tax,
		MerchantChargeFee: req.Price.MerchantChargeFee,
		GrabFundPromo:     req.Price.GrabFundPromo,
		MerchantFundPromo: req.Price.MerchantFundPromo,
		BasketPromo:       req.Price.BasketPromo,
		DeliveryFee:       req.Price.DeliveryFee,
		SmallOrderFee:     req.Price.SmallOrderFee,
		EaterPayment:      req.Price.EaterPayment,
		Total:             &calculatedTotal, // 特殊处理：Grab SDK 不支持此字段，通过计算生成
	}

	// DineIn (nullable)
	if req.DineIn.IsSet() {
		dineIn := req.DineIn.Get()
		if dineIn != nil {
			takeoutDineIn := &message.TakeoutDineIn{
				TableID:    dineIn.TableID,
				EaterCount: dineIn.EaterCount,
			}
			order.DineIn.Set(takeoutDineIn)
		}
	}

	// Receiver (nullable)
	if req.Receiver.IsSet() {
		receiver := req.Receiver.Get()
		if receiver != nil {
			takeoutReceiver := &message.TakeoutReceiver{
				Name:   receiver.Name,
				Phones: receiver.Phones,
			}

			if receiver.Address != nil {
				takeoutReceiver.Address = &message.TakeoutAddress{
					UnitNumber:          receiver.Address.UnitNumber,
					DeliveryInstruction: receiver.Address.DeliveryInstruction,
					PoiSource:           receiver.Address.PoiSource,
					PoiID:               receiver.Address.PoiID,
					Address:             receiver.Address.Address,
					Postcode:            receiver.Address.Postcode,
				}

				if receiver.Address.Coordinates != nil {
					takeoutReceiver.Address.Coordinates = &message.TakeoutCoordinates{
						Latitude:  receiver.Address.Coordinates.Latitude,
						Longitude: receiver.Address.Coordinates.Longitude,
					}
				}
			}

			order.Receiver.Set(takeoutReceiver)
		}
	}

	// OrderReadyEstimation
	if req.OrderReadyEstimation != nil {
		order.OrderReadyEstimation = &message.TakeoutOrderReadyEstimation{
			AllowChange:             req.OrderReadyEstimation.AllowChange,
			EstimatedOrderReadyTime: req.OrderReadyEstimation.EstimatedOrderReadyTime,
			MaxOrderReadyTime:       req.OrderReadyEstimation.MaxOrderReadyTime,
		}

		// NewOrderReadyTime (nullable)
		if req.OrderReadyEstimation.NewOrderReadyTime.IsSet() {
			newTime := req.OrderReadyEstimation.NewOrderReadyTime.Get()
			if newTime != nil {
				order.OrderReadyEstimation.NewOrderReadyTime.Set(newTime)
			}
		}
	}

	// Campaigns
	for _, campaign := range req.Campaigns {
		takeoutCampaign := message.TakeoutCampaign{
			ID:                 campaign.Id,
			Name:               campaign.Name,
			CampaignNameForMex: campaign.CampaignNameForMex,
			Level:              campaign.Level,
			Type:               campaign.Type,
			UsageCount:         campaign.UsageCount,
			MexFundedRatio:     campaign.MexFundedRatio,
			DeductedAmount:     campaign.DeductedAmount,
			DeductedPart:       campaign.DeductedPart,
			AppliedItemIDs:     campaign.AppliedItemIDs,
		}

		if campaign.FreeItem != nil {
			takeoutCampaign.FreeItem = &message.TakeoutFreeItem{
				ID:       campaign.FreeItem.Id,
				Name:     campaign.FreeItem.Name,
				Quantity: campaign.FreeItem.Quantity,
				Price:    campaign.FreeItem.Price,
			}
		}

		order.Campaigns = append(order.Campaigns, takeoutCampaign)
	}

	// Promos
	for _, promo := range req.Promos {
		takeoutPromo := message.TakeoutPromo{
			Code:             promo.Code,
			Description:      promo.Description,
			Name:             promo.Name,
			PromoAmount:      promo.PromoAmount,
			MexFundedRatio:   promo.MexFundedRatio,
			MexFundedAmount:  promo.MexFundedAmount,
			TargetedPrice:    promo.TargetedPrice,
			PromoAmountInMin: promo.PromoAmountInMin,
		}
		order.Promos = append(order.Promos, takeoutPromo)
	}

	return order, nil
}

// ToGrabSDK 转换为 Grab SDK 订单请求（如需回传）
func ToGrabSDK(o *message.TakeoutOrder) *grabsdk.SubmitOrderRequest {
	if o == nil {
		return nil
	}

	req := &grabsdk.SubmitOrderRequest{
		OrderID:           o.OrderID,
		ShortOrderNumber:  o.ShortOrderNumber,
		MerchantID:        o.MerchantID,
		PartnerMerchantID: o.PartnerMerchantID,
		PaymentType:       o.PaymentType,
		Cutlery:           o.Cutlery,
		OrderTime:         o.OrderTime,
	}

	if o.SubmitTime != nil {
		req.SubmitTime = o.SubmitTime
	}

	if o.CompleteTime != nil {
		req.CompleteTime = o.CompleteTime
	}

	if o.ScheduledTime != nil {
		req.ScheduledTime = o.ScheduledTime
	}

	if o.OrderState != nil {
		req.OrderState = o.OrderState
	}

	if o.MembershipID != nil {
		req.MembershipID = o.MembershipID
	}

	// Currency
	req.Currency = grabsdk.Currency{
		Code:     o.Currency.Code,
		Symbol:   o.Currency.Symbol,
		Exponent: o.Currency.Exponent,
	}

	// FeatureFlags
	req.FeatureFlags = grabsdk.OrderFeatureFlags{
		OrderAcceptedType: o.FeatureFlags.OrderAcceptedType,
		IsMexEditOrder:    o.FeatureFlags.IsMexEditOrder,
		OrderType:         o.FeatureFlags.OrderType,
	}

	// Items
	for _, item := range o.Items {
		grabItem := grabsdk.OrderItem{
			Id:             item.ID,
			GrabItemID:     item.GrabItemID,
			Quantity:       item.Quantity,
			Price:          item.Price, // Already in minor unit
			Tax:            item.Tax,
			Specifications: item.Specifications,
		}

		// Modifiers
		for _, mod := range item.Modifiers {
			grabItem.Modifiers = append(grabItem.Modifiers, grabsdk.OrderItemModifier{
				Id:       mod.ID,
				Quantity: mod.Quantity,
				Price:    mod.Price,
				Tax:      mod.Tax,
			})
		}

		// OutOfStockInstruction
		if item.OutOfStockInstruction.IsSet() {
			inst := item.OutOfStockInstruction.Get()
			if inst != nil {
				grabInst := grabsdk.OutOfStockInstruction{
					Title:                 inst.Title,
					InstructionType:       inst.InstructionType,
					ReplacementItemID:     inst.ReplacementItemID,
					ReplacementGrabItemID: inst.ReplacementGrabItemID,
				}
				grabItem.OutOfStockInstruction.Set(&grabInst)
			}
		}

		req.Items = append(req.Items, grabItem)
	}

	// Price
	req.Price = grabsdk.OrderPrice{
		Subtotal:          o.Price.Subtotal,
		Tax:               o.Price.Tax,
		MerchantChargeFee: o.Price.MerchantChargeFee,
		GrabFundPromo:     o.Price.GrabFundPromo,
		MerchantFundPromo: o.Price.MerchantFundPromo,
		BasketPromo:       o.Price.BasketPromo,
		DeliveryFee:       o.Price.DeliveryFee,
		SmallOrderFee:     o.Price.SmallOrderFee,
		EaterPayment:      o.Price.EaterPayment,
	}

	// DineIn
	if o.DineIn.IsSet() {
		dineIn := o.DineIn.Get()
		if dineIn != nil {
			grabDineIn := grabsdk.DineIn{
				TableID:    dineIn.TableID,
				EaterCount: dineIn.EaterCount,
			}
			req.DineIn.Set(&grabDineIn)
		}
	}

	// Receiver
	if o.Receiver.IsSet() {
		receiver := o.Receiver.Get()
		if receiver != nil {
			grabReceiver := &grabsdk.Receiver{
				Name:   receiver.Name,
				Phones: receiver.Phones,
			}

			if receiver.Address != nil {
				grabReceiver.Address = &grabsdk.Address{
					UnitNumber:          receiver.Address.UnitNumber,
					DeliveryInstruction: receiver.Address.DeliveryInstruction,
					PoiSource:           receiver.Address.PoiSource,
					PoiID:               receiver.Address.PoiID,
					Address:             receiver.Address.Address,
					Postcode:            receiver.Address.Postcode,
				}

				if receiver.Address.Coordinates != nil {
					grabReceiver.Address.Coordinates = &grabsdk.Coordinates{
						Latitude:  receiver.Address.Coordinates.Latitude,
						Longitude: receiver.Address.Coordinates.Longitude,
					}
				}
			}

			req.Receiver.Set(grabReceiver)
		}
	}

	// OrderReadyEstimation
	if o.OrderReadyEstimation != nil {
		req.OrderReadyEstimation = &grabsdk.OrderReadyEstimation{
			AllowChange:             o.OrderReadyEstimation.AllowChange,
			EstimatedOrderReadyTime: o.OrderReadyEstimation.EstimatedOrderReadyTime,
			MaxOrderReadyTime:       o.OrderReadyEstimation.MaxOrderReadyTime,
		}

		// NewOrderReadyTime
		if o.OrderReadyEstimation.NewOrderReadyTime.IsSet() {
			newTime := o.OrderReadyEstimation.NewOrderReadyTime.Get()
			if newTime != nil {
				req.OrderReadyEstimation.NewOrderReadyTime.Set(newTime)
			}
		}
	}

	// Campaigns
	for _, campaign := range o.Campaigns {
		grabCampaign := grabsdk.OrderCampaign{
			Id:                 campaign.ID,
			Name:               campaign.Name,
			CampaignNameForMex: campaign.CampaignNameForMex,
			Level:              campaign.Level,
			Type:               campaign.Type,
			UsageCount:         campaign.UsageCount,
			MexFundedRatio:     campaign.MexFundedRatio,
			DeductedAmount:     campaign.DeductedAmount,
			DeductedPart:       campaign.DeductedPart,
			AppliedItemIDs:     campaign.AppliedItemIDs,
		}

		if campaign.FreeItem != nil {
			grabCampaign.FreeItem = &grabsdk.OrderFreeItem{
				Id:       campaign.FreeItem.ID,
				Name:     campaign.FreeItem.Name,
				Quantity: campaign.FreeItem.Quantity,
				Price:    campaign.FreeItem.Price,
			}
		}

		req.Campaigns = append(req.Campaigns, grabCampaign)
	}

	// Promos
	for _, promo := range o.Promos {
		grabPromo := grabsdk.OrderPromo{
			Code:             promo.Code,
			Description:      promo.Description,
			Name:             promo.Name,
			PromoAmount:      promo.PromoAmount,
			MexFundedRatio:   promo.MexFundedRatio,
			MexFundedAmount:  promo.MexFundedAmount,
			TargetedPrice:    promo.TargetedPrice,
			PromoAmountInMin: promo.PromoAmountInMin,
		}
		req.Promos = append(req.Promos, grabPromo)
	}

	return req
}

// ============================================================================
// Lineman API 转换方法
// ============================================================================

// FromLinemanPlaceOrder 从 Lineman PlaceOrder 请求转换为统一模型
// 参考映射: https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=182890165#gid=182890165
func FromLinemanPlaceOrder(req *linemanv1.PlaceOrderReq) (*message.TakeoutOrder, error) {
	if req == nil {
		return nil, nil
	}

	// Parse orderAcceptedTime to time.Time for SubmitTime
	var submitTime *time.Time
	if req.OrderAcceptedTime != "" {
		t, err := time.Parse(time.RFC3339, req.OrderAcceptedTime)
		if err == nil {
			submitTime = &t
		}
	}

	order := &message.TakeoutOrder{
		// 基础字段映射（参考用户要求）
		OrderID:          req.OrderId,           // orderId -> OrderID
		ShortOrderNumber: req.OrderShortCode,    // orderShortCode -> ShortOrderNumber
		MerchantID:       "",                    // MerchantID 为空
		PaymentType:      "CASH",                // Lineman 默认为 CASH
		Cutlery:          false,                 // Lineman 不提供，默认 false
		OrderTime:        req.OrderAcceptedTime, // orderAcceptedTime -> OrderTime
		SubmitTime:       submitTime,            // 转换后的 time.Time
	}

	// PartnerMerchantID
	if req.StoreId != "" {
		order.PartnerMerchantID = &req.StoreId // storeId -> PartnerMerchantID (TTPOS 侧的店铺 ID)
	}

	// Currency - 设置 THB 默认值
	order.Currency = message.TakeoutCurrency{
		Code:     "THB",
		Symbol:   "฿",
		Exponent: 2, // THB 使用 satang，exponent = 2
	}

	// FeatureFlags.OrderType（从 Lineman customerType 转换）
	// 参考 Google Sheets：
	// - Lineman DELIVERY -> Grab "Delivery" 或业务类型 "DeliveryByLineman"
	// - Lineman PICKUP -> Grab "Pickup" 或业务类型 "SelfPickup"
	orderTypeMapping := map[string]string{
		"DELIVERY": "Delivery",
		"PICKUP":   "Pickup",
	}

	orderType := "Delivery" // 默认值
	if mappedType, ok := orderTypeMapping[req.CustomerType]; ok {
		orderType = mappedType
	}

	order.FeatureFlags = message.TakeoutFeatureFlags{
		OrderAcceptedType: "AUTO", // Lineman 默认为自动接单
		OrderType:         orderType,
	}

	// MembershipID（复用 Grab 字段名）
	// Lineman: memberId -> MembershipID
	if req.MemberId != "" {
		order.MembershipID = &req.MemberId
	}

	// AdditionalProperties（复用类似 Grab 的字段名）
	// Lineman: additionalItems[] -> AdditionalProperties
	if len(req.AdditionalItems) > 0 {
		order.AdditionalProperties = make(map[string]any)
		additionalItems := make([]map[string]any, 0, len(req.AdditionalItems))
		for _, addItem := range req.AdditionalItems {
			additionalItems = append(additionalItems, map[string]any{
				"name": addItem.Name,
			})
		}
		order.AdditionalProperties["additionalItems"] = additionalItems
	}

	// 收集促销信息（从商品级别提取到订单级别）
	// Lineman: items[].promotionId + items[].discount -> Grab: order.promos[]
	promoMap := make(map[string]*message.TakeoutPromo) // 使用 map 去重

	// Items（参考 Google Sheets 映射）
	for _, item := range req.Items {
		// 转换价格：Lineman 使用泰铢（float64），需要转换为最小单位（satang）
		priceInMinor := int64(item.UnitPrice * 100)

		takeoutItem := message.TakeoutOrderItem{
			ID:         item.Id, // id -> ID
			GrabItemID: item.Id, // Lineman 没有单独的 GrabItemID，使用 id
			Quantity:   int32(item.Quantity),
			Price:      priceInMinor, // unitPrice -> Price (转换为最小单位)
		}

		// Specifications（商品备注/规格说明）
		// Lineman: memo -> Specifications（参考 Google Sheets 映射）
		if item.Memo != "" {
			takeoutItem.Specifications = &item.Memo
		}

		// 收集促销信息（转换为订单级别的 Promos）
		// Lineman: items[].promotionId -> Grab: promos[].code
		// Lineman: items[].discount -> Grab: promos[].mexFundedAmount
		if item.PromotionId != "" && item.Discount != 0 {
			// 转换折扣金额为最小单位
			discountInMinor := int64(item.Discount * 100)
			if _, exists := promoMap[item.PromotionId]; !exists {
				// 新促销，创建 Promo 结构
				promoCode := item.PromotionId
				promoMap[item.PromotionId] = &message.TakeoutPromo{
					Code:            &promoCode,       // promotionId -> code
					MexFundedAmount: &discountInMinor, // discount -> mexFundedAmount (商户承担金额)
				}
			} else {
				// 相同促销，累加金额
				existingAmount := *promoMap[item.PromotionId].MexFundedAmount
				newAmount := existingAmount + discountInMinor
				promoMap[item.PromotionId].MexFundedAmount = &newAmount
			}
		}

		// Properties -> Modifiers 转换
		// Lineman: properties[].values[] (属性组 -> 属性值列表)
		// Grab: modifiers[] (扁平的修改项列表)
		// 需要将 Lineman 的属性组结构转换为 Grab 的扁平结构
		for _, prop := range item.Properties {
			// 为每个 property 创建一个 modifier，包含所有 values
			modID := prop.Id
			var modPrice int64 = 0
			modQty := int32(1) // Lineman 不提供 quantity，默认为 1

			// 累加 values 的价格
			for _, val := range prop.Values {
				modPrice += int64(val.Price * 100) // 转换为最小单位
			}

			modifier := message.TakeoutModifier{
				ID:       &modID,
				Quantity: &modQty,
				Price:    &modPrice,
			}

			takeoutItem.Modifiers = append(takeoutItem.Modifiers, modifier)
		}

		order.Items = append(order.Items, takeoutItem)
	}

	// 转换促销信息（从 map 转为数组）
	// Lineman: items[].promotionId + items[].discount -> Grab: order.promos[]
	for _, promo := range promoMap {
		order.Promos = append(order.Promos, *promo)
	}

	// Price（参考 Google Sheets 映射）
	// 转换价格为最小单位
	subtotalInMinor := int64(req.RestaurantRevenue * 100)
	eaterPaymentInMinor := int64(req.RestaurantRevenue * 100)

	order.Price = message.TakeoutOrderPrice{
		Subtotal:     subtotalInMinor,      // restaurantRevenue -> Subtotal（用户实付金额）
		EaterPayment: &eaterPaymentInMinor, // restaurantRevenue -> EaterPayment
	}

	return order, nil
}

// ToLinemanPlaceOrder 转换为 Lineman PlaceOrder 请求（如需回传）
func ToLinemanPlaceOrder(o *message.TakeoutOrder) *linemanv1.PlaceOrderReq {
	if o == nil {
		return nil
	}

	// 从最小单位转换回泰铢
	subtotalInTHB := float64(o.Price.Subtotal)

	req := &linemanv1.PlaceOrderReq{
		StoreId:           o.MerchantID,
		OrderId:           o.OrderID,
		OrderShortCode:    o.ShortOrderNumber,
		RestaurantRevenue: subtotalInTHB,
		OrderAcceptedTime: o.OrderTime,
	}

	// PartnerMerchantID -> PartnerId
	if o.PartnerMerchantID != nil {
		req.PartnerId = *o.PartnerMerchantID
	}

	// CustomerType（从 FeatureFlags.OrderType 推导）
	// 参考 Google Sheets 反向映射：
	// - Grab "Delivery" -> Lineman DELIVERY
	// - Grab "Pickup" -> Lineman PICKUP
	// - Grab "DineIn" -> Lineman PICKUP（堂食视为自取）
	switch o.FeatureFlags.OrderType {
	case "Delivery":
		req.CustomerType = "DELIVERY"
	case "Pickup":
		req.CustomerType = "PICKUP"
	case "DineIn":
		req.CustomerType = "PICKUP" // 堂食视为自取
	default:
		req.CustomerType = "DELIVERY" // 默认外送
	}

	// MemberId（复用 MembershipID）
	// Lineman: memberId -> Grab: membershipID
	if o.MembershipID != nil {
		req.MemberId = *o.MembershipID
	}

	// AdditionalItems（从 AdditionalProperties 转换）
	if o.AdditionalProperties != nil {
		if additionalItems, ok := o.AdditionalProperties["additionalItems"].([]map[string]any); ok {
			for _, addItem := range additionalItems {
				if name, ok := addItem["name"].(string); ok {
					req.AdditionalItems = append(req.AdditionalItems, linemanv1.OrderAdditionalItem{
						Name: name,
					})
				}
			}
		}
	}

	// Items
	for _, item := range o.Items {
		// 从最小单位转换回泰铢
		priceInTHB := float64(item.Price)

		linemanItem := linemanv1.OrderItem{
			Id:        item.ID,
			Quantity:  int(item.Quantity),
			UnitPrice: priceInTHB,
		}

		// Memo（从 Specifications 转换）
		// 参考 Google Sheets 第 24 行：Lineman memo -> Grab specifications
		if item.Specifications != nil {
			linemanItem.Memo = *item.Specifications
		}

		// 注意：promotionId 和 discount 现在在订单级别的 Promos 中
		// 反向转换时，将订单级别的 Promos 信息分摊到第一个商品上
		// （这是一个折衷方案，因为 Grab Promos 是订单级别，而 Lineman 是商品级别）

		// Modifiers -> Properties 转换
		// 将 Grab 的扁平 modifiers 结构转换回 Lineman 的属性组结构
		for _, mod := range item.Modifiers {
			var propID string
			if mod.ID != nil {
				propID = *mod.ID
			}

			prop := linemanv1.OrderItemProperty{
				Id: propID,
			}

			// 转换价格从最小单位到泰铢
			var priceInTHB float64 = 0
			if mod.Price != nil {
				priceInTHB = float64(*mod.Price)
			}

			// 创建一个默认 value
			prop.Values = append(prop.Values, linemanv1.OrderItemPropertyValue{
				Id:    propID,
				Price: priceInTHB,
			})

			linemanItem.Properties = append(linemanItem.Properties, prop)
		}

		req.Items = append(req.Items, linemanItem)
	}

	// 将订单级别的 Promos 转换为商品级别的 promotionId/discount
	// 策略：将第一个 Promo 的信息设置到第一个商品上（折衷方案）
	// 注意：这是反向转换的限制，因为 Grab Promos 是订单级别，Lineman 是商品级别
	if len(o.Promos) > 0 && len(req.Items) > 0 {
		firstPromo := o.Promos[0]
		if firstPromo.Code != nil {
			req.Items[0].PromotionId = *firstPromo.Code
		}
		if firstPromo.MexFundedAmount != nil {
			// 从最小单位转换回泰铢
			req.Items[0].Discount = float64(*firstPromo.MexFundedAmount)
		}
	}

	return req
}

// ============================================================================
// 辅助方法
// ============================================================================

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func stringPtrValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
