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
		PartnerMerchantID: stringPtrValue(req.PartnerMerchantID),
		PaymentType:       req.PaymentType,
		Cutlery:           &req.Cutlery,
		OrderTime:         req.OrderTime,
	}

	// SubmitTime, CompleteTime, ScheduledTime
	if req.SubmitTime != nil {
		submitTime := req.SubmitTime.Format(time.RFC3339)
		order.SubmitTime = &submitTime
	}

	if req.CompleteTime != nil {
		completeTime := req.CompleteTime.Format(time.RFC3339)
		order.CompleteTime = &completeTime
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

	// Currency
	if req.Currency.Code != "" {
		order.Currency = &message.TakeoutCurrency{
			Code:     req.Currency.Code,
			Symbol:   req.Currency.Symbol,
			Exponent: int(req.Currency.Exponent),
		}
	}

	// FeatureFlags
	if req.FeatureFlags.OrderAcceptedType != "" || req.FeatureFlags.OrderType != "" {
		order.FeatureFlags = &message.TakeoutFeatureFlags{
			OrderAcceptedType: req.FeatureFlags.OrderAcceptedType,
			IsMexEditOrder:    req.FeatureFlags.IsMexEditOrder != nil && *req.FeatureFlags.IsMexEditOrder,
			OrderType:         req.FeatureFlags.OrderType,
		}
	}

	// Items
	for _, item := range req.Items {
		takeoutItem := message.TakeoutOrderItem{
			ID:             item.Id,
			GrabItemID:     stringPtr(item.GrabItemID),
			Quantity:       int(item.Quantity),
			Price:          float64(item.Price) / 100, // 转换为泰铢
			Specifications: item.Specifications,
		}

		if item.Tax != nil && *item.Tax != 0 {
			tax := float64(*item.Tax) / 100
			takeoutItem.Tax = &tax
		}

		// Modifiers
		for _, mod := range item.Modifiers {
			modifier := message.TakeoutModifier{}

			if mod.Id != nil {
				modifier.ID = *mod.Id
			}

			if mod.Quantity != nil {
				modifier.Quantity = int(*mod.Quantity)
			} else {
				modifier.Quantity = 1
			}

			if mod.Price != nil {
				modifier.Price = float64(*mod.Price) / 100
			}

			takeoutItem.Modifiers = append(takeoutItem.Modifiers, modifier)
		}

		// OutOfStockInstruction
		if item.OutOfStockInstruction.IsSet() {
			inst := item.OutOfStockInstruction.Get()
			if inst != nil && inst.InstructionType != nil {
				takeoutItem.OutOfStockInstruction = &message.TakeoutOutOfStockInstruction{
					Type: *inst.InstructionType,
				}
			}
		}

		order.Items = append(order.Items, takeoutItem)
	}

	// Price
	order.Price = message.TakeoutOrderPrice{
		Subtotal: float64(req.Price.Subtotal) / 100,
	}

	if req.Price.Tax != nil {
		tax := float64(*req.Price.Tax) / 100
		order.Price.Tax = &tax
	}

	if req.Price.MerchantChargeFee != nil {
		fee := float64(*req.Price.MerchantChargeFee) / 100
		order.Price.MerchantChargeFee = &fee
	}

	if req.Price.DeliveryFee != nil {
		fee := float64(*req.Price.DeliveryFee) / 100
		order.Price.DeliveryFee = &fee
	}

	if req.Price.GrabFundPromo != nil {
		promo := float64(*req.Price.GrabFundPromo) / 100
		order.Price.GrabFundPromo = &promo
	}

	if req.Price.MerchantFundPromo != nil {
		promo := float64(*req.Price.MerchantFundPromo) / 100
		order.Price.MerchantFundPromo = &promo
	}

	if req.Price.EaterPayment != nil {
		payment := float64(*req.Price.EaterPayment) / 100
		order.Price.EaterPayment = &payment
	}

	// DineIn
	if req.DineIn.IsSet() {
		dineIn := req.DineIn.Get()
		if dineIn != nil {
			takeoutDineIn := &message.TakeoutDineIn{}

			if dineIn.TableID != nil {
				takeoutDineIn.TableID = *dineIn.TableID
			}

			if dineIn.EaterCount != nil {
				takeoutDineIn.EaterCount = int(*dineIn.EaterCount)
			}

			order.DineIn = takeoutDineIn
		}
	}

	// Receiver
	if req.Receiver.IsSet() {
		receiver := req.Receiver.Get()
		if receiver != nil {
			takeoutReceiver := &message.TakeoutReceiver{}

			if receiver.Name != nil {
				takeoutReceiver.Name = *receiver.Name
			}

			// Grab SDK 使用 Phones 字段（注意是复数）
			if receiver.Phones != nil {
				takeoutReceiver.Phone = *receiver.Phones
			}

			if receiver.Address != nil {
				takeoutReceiver.Address = &message.TakeoutDeliveryAddress{}

				if receiver.Address.UnitNumber != nil {
					takeoutReceiver.Address.UnitNumber = *receiver.Address.UnitNumber
				}

				if receiver.Address.DeliveryInstruction != nil {
					takeoutReceiver.Address.DeliveryInstruction = *receiver.Address.DeliveryInstruction
				}

				if receiver.Address.PoiSource != nil {
					takeoutReceiver.Address.PoiSource = *receiver.Address.PoiSource
				}

				if receiver.Address.PoiID != nil {
					takeoutReceiver.Address.PoiID = *receiver.Address.PoiID
				}

				if receiver.Address.Address != nil {
					takeoutReceiver.Address.Address = *receiver.Address.Address
				}

				if receiver.Address.Postcode != nil {
					takeoutReceiver.Address.Postcode = *receiver.Address.Postcode
				}

				if receiver.Address.Coordinates != nil {
					takeoutReceiver.Address.Coordinates = &message.TakeoutCoordinates{}

					if receiver.Address.Coordinates.Latitude != nil {
						takeoutReceiver.Address.Coordinates.Latitude = *receiver.Address.Coordinates.Latitude
					}

					if receiver.Address.Coordinates.Longitude != nil {
						takeoutReceiver.Address.Coordinates.Longitude = *receiver.Address.Coordinates.Longitude
					}
				}
			}

			order.Receiver = takeoutReceiver
		}
	}

	// OrderReadyEstimation
	if req.OrderReadyEstimation != nil {
		order.OrderReadyEstimation = &message.TakeoutOrderReadyEstimation{
			AllowChange:        req.OrderReadyEstimation.AllowChange,
			EstimatedReadyTime: req.OrderReadyEstimation.EstimatedOrderReadyTime.Format(time.RFC3339),
			MaxReadyTime:       req.OrderReadyEstimation.MaxOrderReadyTime.Format(time.RFC3339),
		}

		// NewOrderReadyTime 可能为空
		if req.OrderReadyEstimation.NewOrderReadyTime.IsSet() {
			newTime := req.OrderReadyEstimation.NewOrderReadyTime.Get()
			if newTime != nil {
				newTimeStr := newTime.Format(time.RFC3339)
				order.OrderReadyEstimation.NewReadyTime = newTimeStr
			}
		}
	}

	return order, nil
}

// ToGrabSDK 转换为 Grab SDK 订单请求（如需回传）
func ToGrabSDK(o *message.TakeoutOrder) *grabsdk.SubmitOrderRequest {
	if o == nil {
		return nil
	}

	req := &grabsdk.SubmitOrderRequest{
		OrderID:          o.OrderID,
		ShortOrderNumber: o.ShortOrderNumber,
		MerchantID:       o.MerchantID,
		PaymentType:      o.PaymentType,
		OrderTime:        o.OrderTime,
	}

	if o.PartnerMerchantID != "" {
		req.PartnerMerchantID = &o.PartnerMerchantID
	}

	if o.Cutlery != nil {
		req.Cutlery = *o.Cutlery
	}

	if o.SubmitTime != nil {
		t, _ := time.Parse(time.RFC3339, *o.SubmitTime)
		req.SubmitTime = &t
	}

	if o.CompleteTime != nil {
		t, _ := time.Parse(time.RFC3339, *o.CompleteTime)
		req.CompleteTime = &t
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
	if o.Currency != nil {
		req.Currency = grabsdk.Currency{
			Code:     o.Currency.Code,
			Symbol:   o.Currency.Symbol,
			Exponent: int32(o.Currency.Exponent),
		}
	}

	// FeatureFlags
	if o.FeatureFlags != nil {
		isMexEditOrder := o.FeatureFlags.IsMexEditOrder
		req.FeatureFlags = grabsdk.OrderFeatureFlags{
			OrderAcceptedType: o.FeatureFlags.OrderAcceptedType,
			IsMexEditOrder:    &isMexEditOrder,
			OrderType:         o.FeatureFlags.OrderType,
		}
	}

	// Items
	for _, item := range o.Items {
		grabItemID := item.ID
		if item.GrabItemID != nil {
			grabItemID = *item.GrabItemID
		}

		grabItem := grabsdk.OrderItem{
			Id:         item.ID,
			GrabItemID: grabItemID,
			Quantity:   int32(item.Quantity),
			Price:      int64(item.Price * 100), // 转换为最小单位
		}

		if item.Tax != nil {
			tax := int64(*item.Tax * 100)
			grabItem.Tax = &tax
		}

		if item.Specifications != nil {
			grabItem.Specifications = item.Specifications
		}

		// Modifiers
		for _, mod := range item.Modifiers {
			modID := mod.ID
			modQty := int32(mod.Quantity)
			modPrice := int64(mod.Price * 100)

			grabItem.Modifiers = append(grabItem.Modifiers, grabsdk.OrderItemModifier{
				Id:       &modID,
				Quantity: &modQty,
				Price:    &modPrice,
			})
		}

		// OutOfStockInstruction
		if item.OutOfStockInstruction != nil {
			instType := item.OutOfStockInstruction.Type
			inst := grabsdk.OutOfStockInstruction{
				InstructionType: &instType,
			}
			grabItem.OutOfStockInstruction.Set(&inst)
		}

		req.Items = append(req.Items, grabItem)
	}

	// Price
	req.Price = grabsdk.OrderPrice{
		Subtotal: int64(o.Price.Subtotal * 100),
	}

	if o.Price.Tax != nil {
		tax := int64(*o.Price.Tax * 100)
		req.Price.Tax = &tax
	}

	if o.Price.MerchantChargeFee != nil {
		fee := int64(*o.Price.MerchantChargeFee * 100)
		req.Price.MerchantChargeFee = &fee
	}

	if o.Price.DeliveryFee != nil {
		fee := int64(*o.Price.DeliveryFee * 100)
		req.Price.DeliveryFee = &fee
	}

	if o.Price.GrabFundPromo != nil {
		promo := int64(*o.Price.GrabFundPromo * 100)
		req.Price.GrabFundPromo = &promo
	}

	if o.Price.MerchantFundPromo != nil {
		promo := int64(*o.Price.MerchantFundPromo * 100)
		req.Price.MerchantFundPromo = &promo
	}

	if o.Price.EaterPayment != nil {
		payment := int64(*o.Price.EaterPayment * 100)
		req.Price.EaterPayment = &payment
	}

	// DineIn
	if o.DineIn != nil {
		dineIn := grabsdk.DineIn{}

		if o.DineIn.TableID != "" {
			tableID := o.DineIn.TableID
			dineIn.TableID = &tableID
		}

		if o.DineIn.EaterCount != 0 {
			eaterCount := int64(o.DineIn.EaterCount)
			dineIn.EaterCount = &eaterCount
		}

		req.DineIn.Set(&dineIn)
	}

	// Receiver
	if o.Receiver != nil {
		receiver := &grabsdk.Receiver{}

		if o.Receiver.Name != "" {
			name := o.Receiver.Name
			receiver.Name = &name
		}

		if o.Receiver.Phone != "" {
			phone := o.Receiver.Phone
			receiver.Phones = &phone // 注意 Grab SDK 使用 Phones 字段（复数）
		}

		if o.Receiver.Address != nil {
			receiver.Address = &grabsdk.Address{}

			if o.Receiver.Address.UnitNumber != "" {
				unitNum := o.Receiver.Address.UnitNumber
				receiver.Address.UnitNumber = &unitNum
			}

			if o.Receiver.Address.DeliveryInstruction != "" {
				instruction := o.Receiver.Address.DeliveryInstruction
				receiver.Address.DeliveryInstruction = &instruction
			}

			if o.Receiver.Address.PoiSource != "" {
				poiSource := o.Receiver.Address.PoiSource
				receiver.Address.PoiSource = &poiSource
			}

			if o.Receiver.Address.PoiID != "" {
				poiID := o.Receiver.Address.PoiID
				receiver.Address.PoiID = &poiID
			}

			if o.Receiver.Address.Address != "" {
				addr := o.Receiver.Address.Address
				receiver.Address.Address = &addr
			}

			if o.Receiver.Address.Postcode != "" {
				postcode := o.Receiver.Address.Postcode
				receiver.Address.Postcode = &postcode
			}

			if o.Receiver.Address.Coordinates != nil {
				receiver.Address.Coordinates = &grabsdk.Coordinates{}

				if o.Receiver.Address.Coordinates.Latitude != 0 {
					lat := o.Receiver.Address.Coordinates.Latitude
					receiver.Address.Coordinates.Latitude = &lat
				}

				if o.Receiver.Address.Coordinates.Longitude != 0 {
					lng := o.Receiver.Address.Coordinates.Longitude
					receiver.Address.Coordinates.Longitude = &lng
				}
			}
		}

		req.Receiver.Set(receiver)
	}

	// OrderReadyEstimation
	if o.OrderReadyEstimation != nil {
		// 解析时间字符串
		estimatedTime, err := time.Parse(time.RFC3339, o.OrderReadyEstimation.EstimatedReadyTime)
		if err != nil {
			// 如果解析失败,使用当前时间
			estimatedTime = time.Now()
		}

		maxTime, err := time.Parse(time.RFC3339, o.OrderReadyEstimation.MaxReadyTime)
		if err != nil {
			// 如果解析失败,使用当前时间
			maxTime = time.Now()
		}

		req.OrderReadyEstimation = &grabsdk.OrderReadyEstimation{
			AllowChange:             o.OrderReadyEstimation.AllowChange,
			EstimatedOrderReadyTime: estimatedTime,
			MaxOrderReadyTime:       maxTime,
		}

		// NewReadyTime 可能为空
		if o.OrderReadyEstimation.NewReadyTime != "" {
			newTime, err := time.Parse(time.RFC3339, o.OrderReadyEstimation.NewReadyTime)
			if err == nil {
				req.OrderReadyEstimation.NewOrderReadyTime.Set(&newTime)
			}
		}
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

	order := &message.TakeoutOrder{
		// 基础字段映射（参考用户要求）
		OrderID:           req.OrderId,           // orderId -> OrderID
		ShortOrderNumber:  req.OrderShortCode,    // orderShortCode -> ShortOrderNumber
		MerchantID:        "",                    // MerchantID 为空
		PartnerMerchantID: req.StoreId,           // storeId -> PartnerMerchantID (TTPOS 侧的店铺 ID)
		PaymentType:       "CASH",                // Lineman 默认为 CASH
		OrderTime:         req.OrderAcceptedTime, // orderAcceptedTime -> OrderTime
	}

	// FeatureFlags.OrderType（从 Lineman customerType 转换）
	// 参考 Google Sheets：
	// - Lineman DELIVERY -> Grab "Delivery" 或业务类型 "DeliveryByLineman"
	// - Lineman PICKUP -> Grab "Pickup" 或业务类型 "SelfPickup"
	orderTypeMapping := map[string]string{
		"DELIVERY": "Delivery",
		"PICKUP":   "Pickup",
	}

	if mappedType, ok := orderTypeMapping[req.CustomerType]; ok {
		order.FeatureFlags = &message.TakeoutFeatureFlags{
			OrderAcceptedType: "AUTO", // Lineman 默认为自动接单
			OrderType:         mappedType,
		}
	}

	// MembershipID（复用 Grab 字段名）
	// Lineman: memberId -> MembershipID
	if req.MemberId != "" {
		order.MembershipID = &req.MemberId
	}

	// AdditionalProperties（复用类似 Grab 的字段名）
	// Lineman: additionalItems[] -> AdditionalProperties
	for _, addItem := range req.AdditionalItems {
		order.AdditionalProperties = append(order.AdditionalProperties, message.TakeoutAdditionalProperty{
			Name: addItem.Name,
		})
	}

	// 收集促销信息（从商品级别提取到订单级别）
	// Lineman: items[].promotionId + items[].discount -> Grab: order.promos[]
	promoMap := make(map[string]*message.TakeoutPromo) // 使用 map 去重

	// Items（参考 Google Sheets 映射）
	for _, item := range req.Items {
		takeoutItem := message.TakeoutOrderItem{
			ID:       item.Id,        // id -> ID
			Quantity: item.Quantity,  // quantity -> Quantity
			Price:    item.UnitPrice, // unitPrice -> Price (已含选项费用和折扣)
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
			if _, exists := promoMap[item.PromotionId]; !exists {
				// 新促销，创建 Promo 结构
				promoMap[item.PromotionId] = &message.TakeoutPromo{
					Code:            item.PromotionId, // promotionId -> code
					MexFundedAmount: item.Discount,    // discount -> mexFundedAmount (商户承担金额)
				}
			} else {
				// 相同促销，累加金额
				promoMap[item.PromotionId].MexFundedAmount += item.Discount
			}
		}

		// Properties -> Modifiers 转换
		// Lineman: properties[].values[] (属性组 -> 属性值列表)
		// Grab: modifiers[] (扁平的修改项列表)
		// 需要将 Lineman 的属性组结构转换为 Grab 的扁平结构
		for _, prop := range item.Properties {
			// 为每个 property 创建一个 modifier，包含所有 values
			modifier := message.TakeoutModifier{
				ID:       prop.Id,
				Quantity: 1, // Lineman 不提供 quantity，默认为 1
				Price:    0, // 价格从 values 中累加
			}

			// 转换 values
			for _, val := range prop.Values {
				modifier.Values = append(modifier.Values, message.TakeoutModifierValue{
					ID:    val.Id,
					Price: val.Price,
				})
				modifier.Price += val.Price // 累加价格
			}

			takeoutItem.Modifiers = append(takeoutItem.Modifiers, modifier)
		}

		order.Items = append(order.Items, takeoutItem)
	}

	// 转换促销信息（从 map 转为数组）
	// Lineman: items[].promotionId + items[].discount -> Grab: order.promos[]
	for _, promo := range promoMap {
		order.Promos = append(order.Promos, message.TakeoutPromo{
			Code:            promo.Code,            // promotionId -> code
			MexFundedAmount: promo.MexFundedAmount, // discount -> mexFundedAmount
		})
	}

	// Price（参考 Google Sheets 映射）
	order.Price = message.TakeoutOrderPrice{
		Subtotal:     req.RestaurantRevenue,  // restaurantRevenue -> Subtotal（用户实付金额）
		EaterPayment: &req.RestaurantRevenue, // restaurantRevenue -> EaterPayment
	}

	return order, nil
}

// ToLinemanPlaceOrder 转换为 Lineman PlaceOrder 请求（如需回传）
func ToLinemanPlaceOrder(o *message.TakeoutOrder) *linemanv1.PlaceOrderReq {
	if o == nil {
		return nil
	}

	req := &linemanv1.PlaceOrderReq{
		PartnerId:         o.PartnerMerchantID,
		StoreId:           o.MerchantID,
		OrderId:           o.OrderID,
		OrderShortCode:    o.ShortOrderNumber,
		RestaurantRevenue: o.Price.Subtotal,
		OrderAcceptedTime: o.OrderTime,
	}

	// CustomerType（从 FeatureFlags.OrderType 推导）
	// 参考 Google Sheets 反向映射：
	// - Grab "Delivery" -> Lineman DELIVERY
	// - Grab "Pickup" -> Lineman PICKUP
	// - Grab "DineIn" -> Lineman PICKUP（堂食视为自取）
	if o.FeatureFlags != nil {
		switch o.FeatureFlags.OrderType {
		case "Delivery":
			req.CustomerType = "DELIVERY"
		case "Pickup":
			req.CustomerType = "PICKUP"
		case "DineIn":
			req.CustomerType = "PICKUP" // 堂食视为自取
		default:
			req.CustomerType = "DELIVERY"
		}
	} else {
		req.CustomerType = "DELIVERY" // 默认外送
	}

	// MemberId（复用 MembershipID）
	// Lineman: memberId -> Grab: membershipID
	if o.MembershipID != nil {
		req.MemberId = *o.MembershipID
	}

	// AdditionalItems（从 AdditionalProperties 转换）
	// Lineman: additionalItems[] -> Grab: AdditionalProperties（扩展字段）
	for _, addProp := range o.AdditionalProperties {
		req.AdditionalItems = append(req.AdditionalItems, linemanv1.OrderAdditionalItem{
			Name: addProp.Name,
		})
	}

	// Items
	for _, item := range o.Items {
		linemanItem := linemanv1.OrderItem{
			Id:        item.ID,
			Quantity:  item.Quantity,
			UnitPrice: item.Price,
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
			prop := linemanv1.OrderItemProperty{
				Id: mod.ID,
			}

			// 如果有 Values（来自 Lineman），直接转换
			if len(mod.Values) > 0 {
				for _, val := range mod.Values {
					prop.Values = append(prop.Values, linemanv1.OrderItemPropertyValue{
						Id:    val.ID,
						Price: val.Price,
					})
				}
			} else {
				// 如果没有 Values（来自 Grab），创建一个默认 value
				prop.Values = append(prop.Values, linemanv1.OrderItemPropertyValue{
					Id:    mod.ID,
					Price: mod.Price,
				})
			}

			linemanItem.Properties = append(linemanItem.Properties, prop)
		}

		req.Items = append(req.Items, linemanItem)
	}

	// 将订单级别的 Promos 转换为商品级别的 promotionId/discount
	// 策略：将第一个 Promo 的信息设置到第一个商品上（折衷方案）
	// 注意：这是反向转换的限制，因为 Grab Promos 是订单级别，Lineman 是商品级别
	if len(o.Promos) > 0 && len(req.Items) > 0 {
		firstPromo := o.Promos[0]
		req.Items[0].PromotionId = firstPromo.Code
		req.Items[0].Discount = firstPromo.MexFundedAmount
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
