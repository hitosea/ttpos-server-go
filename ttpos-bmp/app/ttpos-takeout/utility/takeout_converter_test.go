package utility

import (
	"encoding/json"
	"testing"
	"time"

	grabsdk "github.com/grab/grabfood-api-sdk-go"
	"ttpos-api/ttpos-takeout/message"
	linemanv1 "ttpos-bmp/app/ttpos-takeout/api/lineman/v1"
)

// TestFromGrabSDK 测试 Grab SDK 订单转换
func TestFromGrabSDK(t *testing.T) {
	// 构造 Grab SDK 订单请求
	cutlery := false
	partnerMerchantID := "partner-123"
	orderState := "NEW"
	scheduledTime := "2025-12-26T10:00:00Z"
	membershipID := "member-456"
	orderAcceptedType := "AUTO"
	isMexEditOrder := false
	orderType := "DineIn"

	submitTime := time.Now()
	completeTime := time.Now().Add(30 * time.Minute)

	tax := int64(0)
	deliveryFee := int64(0)
	grabFundPromo := int64(0)
	merchantFundPromo := int64(0)
	eaterPayment := int64(33389)
	specifications := "Large"

	req := &grabsdk.SubmitOrderRequest{
		OrderID:           "123456789-C7WHJNBGNPDXR2",
		ShortOrderNumber:  "GF-9869",
		MerchantID:        "GFSBPOS-721-058",
		PartnerMerchantID: &partnerMerchantID,
		PaymentType:       "CASH",
		Cutlery:           cutlery,
		OrderTime:         "2025-12-26T09:57:06Z",
		SubmitTime:        &submitTime,
		CompleteTime:      &completeTime,
		ScheduledTime:     &scheduledTime,
		OrderState:        &orderState,
		MembershipID:      &membershipID,
		Currency: grabsdk.Currency{
			Code:     "THB",
			Symbol:   "฿",
			Exponent: 2,
		},
		FeatureFlags: grabsdk.OrderFeatureFlags{
			OrderAcceptedType: orderAcceptedType,
			IsMexEditOrder:    &isMexEditOrder,
			OrderType:         orderType,
		},
		Items: []grabsdk.OrderItem{
			{
				Id:             "TTPOS-ITEM-592",
				GrabItemID:     "THITE20251226095706648826",
				Quantity:       1,
				Price:          33389, // 333.89 THB (最小单位)
				Tax:            &tax,
				Specifications: &specifications,
				Modifiers: []grabsdk.OrderItemModifier{
					{
						Id:       stringPtr("TTPOS-FLAVOR-592"),
						Quantity: int32Ptr(1),
						Price:    int64Ptr(0),
					},
				},
			},
		},
		Price: grabsdk.OrderPrice{
			Subtotal:          33389,
			Tax:               &tax,
			DeliveryFee:       &deliveryFee,
			GrabFundPromo:     &grabFundPromo,
			MerchantFundPromo: &merchantFundPromo,
			EaterPayment:      &eaterPayment,
		},
	}

	// 转换
	order, err := FromGrabSDK(req)
	if err != nil {
		t.Fatalf("FromGrabSDK failed: %v", err)
	}

	// 验证基础字段
	if order.OrderID != "123456789-C7WHJNBGNPDXR2" {
		t.Errorf("Expected OrderID '123456789-C7WHJNBGNPDXR2', got '%s'", order.OrderID)
	}

	if order.ShortOrderNumber != "GF-9869" {
		t.Errorf("Expected ShortOrderNumber 'GF-9869', got '%s'", order.ShortOrderNumber)
	}

	if order.MerchantID != "GFSBPOS-721-058" {
		t.Errorf("Expected MerchantID 'GFSBPOS-721-058', got '%s'", order.MerchantID)
	}

	// PartnerMerchantID is now *string
	if order.PartnerMerchantID == nil || *order.PartnerMerchantID != "partner-123" {
		t.Errorf("Expected PartnerMerchantID 'partner-123', got '%v'", order.PartnerMerchantID)
	}

	if order.PaymentType != "CASH" {
		t.Errorf("Expected PaymentType 'CASH', got '%s'", order.PaymentType)
	}

	// Cutlery is now bool (not pointer)
	if order.Cutlery != false {
		t.Errorf("Expected Cutlery false, got %v", order.Cutlery)
	}

	// 验证 Currency (now required, not pointer)
	if order.Currency.Code != "THB" {
		t.Errorf("Expected Currency.Code 'THB', got '%s'", order.Currency.Code)
	}

	// 验证 FeatureFlags (now required, not pointer)
	if order.FeatureFlags.OrderAcceptedType != "AUTO" {
		t.Errorf("Expected FeatureFlags.OrderAcceptedType 'AUTO', got '%s'", order.FeatureFlags.OrderAcceptedType)
	}

	// 验证 Items
	if len(order.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(order.Items))
	}

	item := order.Items[0]
	if item.ID != "TTPOS-ITEM-592" {
		t.Errorf("Expected Item.ID 'TTPOS-ITEM-592', got '%s'", item.ID)
	}

	if item.Quantity != 1 {
		t.Errorf("Expected Item.Quantity 1, got %d", item.Quantity)
	}

	// Price is now int64 (minor unit), no longer needs conversion
	expectedPrice := int64(33389)
	if item.Price != expectedPrice {
		t.Errorf("Expected Item.Price %d, got %d", expectedPrice, item.Price)
	}

	// 验证 Modifiers
	if len(item.Modifiers) != 1 {
		t.Fatalf("Expected 1 modifier, got %d", len(item.Modifiers))
	}

	modifier := item.Modifiers[0]
	if modifier.ID == nil || *modifier.ID != "TTPOS-FLAVOR-592" {
		t.Errorf("Expected Modifier.ID 'TTPOS-FLAVOR-592', got '%v'", modifier.ID)
	}

	// 验证 Price (now int64, minor unit)
	if order.Price.Subtotal != 33389 {
		t.Errorf("Expected Price.Subtotal 33389, got %d", order.Price.Subtotal)
	}

	if order.Price.EaterPayment == nil || *order.Price.EaterPayment != 33389 {
		t.Errorf("Expected Price.EaterPayment 33389, got %v", order.Price.EaterPayment)
	}
}

// TestFromLinemanPlaceOrder 测试 Lineman PlaceOrder 订单转换
func TestFromLinemanPlaceOrder(t *testing.T) {
	// 构造 Lineman PlaceOrder 请求（参考 Google Sheets）
	req := &linemanv1.PlaceOrderReq{
		PartnerId:         "partner-001",
		StoreId:           "store-123",
		OrderId:           "LMF-260112-338798091",
		OrderShortCode:    "8091",
		RestaurantRevenue: 333.89,
		OrderAcceptedTime: "2022-11-01T13:08:06+07:00",
		MemberId:          "member-789",
		CustomerType:      "DELIVERY", // DELIVERY 或 PICKUP
		Items: []linemanv1.OrderItem{
			{
				Id:          "TTPOS-ITEM-592",
				Quantity:    1,
				UnitPrice:   333.89,
				Memo:        "No spicy",
				PromotionId: "promo-001",
				Discount:    10.0,
				Properties: []linemanv1.OrderItemProperty{
					{
						Id: "TTPOS-FLAVOR-592",
						Values: []linemanv1.OrderItemPropertyValue{
							{
								Id:    "FLAVOR-MILD",
								Price: 0,
							},
						},
					},
					{
						Id: "TTPOS-SIZE-001",
						Values: []linemanv1.OrderItemPropertyValue{
							{
								Id:    "SIZE-LARGE",
								Price: 20.0,
							},
						},
					},
				},
			},
		},
		AdditionalItems: []linemanv1.OrderAdditionalItem{
			{
				Name: "ไม่รับช้อนส้อมพลาสติก", // 不需要塑料餐具
			},
		},
	}

	// 转换
	order, err := FromLinemanPlaceOrder(req)
	if err != nil {
		t.Fatalf("FromLinemanPlaceOrder failed: %v", err)
	}

	// 验证基础字段映射（参考 Google Sheets）
	if order.OrderID != "LMF-260112-338798091" {
		t.Errorf("Expected OrderID 'LMF-260112-338798091', got '%s'", order.OrderID)
	}

	if order.ShortOrderNumber != "8091" {
		t.Errorf("Expected ShortOrderNumber '8091', got '%s'", order.ShortOrderNumber)
	}

	// 验证 MerchantID 和 PartnerMerchantID 映射（根据用户要求）
	// Lineman: storeId -> PartnerMerchantID (TTPOS 侧的店铺 ID)
	// MerchantID 为空
	if order.MerchantID != "" {
		t.Errorf("Expected MerchantID to be empty, got '%s'", order.MerchantID)
	}

	// PartnerMerchantID is now *string
	if order.PartnerMerchantID == nil || *order.PartnerMerchantID != "store-123" {
		t.Errorf("Expected PartnerMerchantID 'store-123', got '%v'", order.PartnerMerchantID)
	}

	if order.PaymentType != "CASH" {
		t.Errorf("Expected PaymentType 'CASH', got '%s'", order.PaymentType)
	}

	if order.OrderTime != "2022-11-01T13:08:06+07:00" {
		t.Errorf("Expected OrderTime '2022-11-01T13:08:06+07:00', got '%s'", order.OrderTime)
	}

	// 验证 FeatureFlags.OrderType（从 Lineman customerType 转换）
	if order.FeatureFlags.OrderType != "Delivery" {
		t.Errorf("Expected FeatureFlags.OrderType 'Delivery', got '%s'", order.FeatureFlags.OrderType)
	}

	// 验证 MembershipID（复用 Grab 字段名）
	if order.MembershipID == nil || *order.MembershipID != "member-789" {
		t.Errorf("Expected MembershipID 'member-789', got %v", order.MembershipID)
	}

	// 验证 AdditionalProperties（现在是 map[string]any）
	if order.AdditionalProperties == nil {
		t.Fatal("Expected AdditionalProperties not nil")
	}

	additionalItems, ok := order.AdditionalProperties["additionalItems"].([]map[string]any)
	if !ok || len(additionalItems) != 1 {
		t.Fatalf("Expected 1 additional item, got %v", order.AdditionalProperties["additionalItems"])
	}

	if additionalItems[0]["name"] != "ไม่รับช้อนส้อมพลาสติก" {
		t.Errorf("Expected additionalItems[0].name 'ไม่รับช้อนส้อมพลาสติก', got '%v'", additionalItems[0]["name"])
	}

	// 验证 Items
	if len(order.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(order.Items))
	}

	item := order.Items[0]
	if item.ID != "TTPOS-ITEM-592" {
		t.Errorf("Expected Item.ID 'TTPOS-ITEM-592', got '%s'", item.ID)
	}

	if item.Quantity != 1 {
		t.Errorf("Expected Item.Quantity 1, got %d", item.Quantity)
	}

	// Price is now int64 (minor unit): 333.89 THB = 33389 satang
	expectedPrice := int64(33389)
	if item.Price != expectedPrice {
		t.Errorf("Expected Item.Price %d, got %d", expectedPrice, item.Price)
	}

	// 验证 Specifications（Lineman memo -> Grab specifications）
	if item.Specifications == nil || *item.Specifications != "No spicy" {
		t.Errorf("Expected Item.Specifications 'No spicy', got %v", item.Specifications)
	}

	// 验证促销信息（从商品级别转换到订单级别）
	// Lineman: items[].promotionId + items[].discount -> Grab: order.promos[]
	if len(order.Promos) != 1 {
		t.Fatalf("Expected 1 promo, got %d", len(order.Promos))
	}

	promo := order.Promos[0]
	if promo.Code == nil || *promo.Code != "promo-001" {
		t.Errorf("Expected Promo.Code 'promo-001', got '%v'", promo.Code)
	}

	// MexFundedAmount is now *int64 (minor unit): 10.0 THB = 1000 satang
	expectedDiscount := int64(1000)
	if promo.MexFundedAmount == nil || *promo.MexFundedAmount != expectedDiscount {
		t.Errorf("Expected Promo.MexFundedAmount %d, got %v", expectedDiscount, promo.MexFundedAmount)
	}

	// 验证 Properties -> Modifiers 转换
	if len(item.Modifiers) != 2 {
		t.Fatalf("Expected 2 modifiers, got %d", len(item.Modifiers))
	}

	// 第一个 Modifier（TTPOS-FLAVOR-592）
	modifier1 := item.Modifiers[0]
	if modifier1.ID == nil || *modifier1.ID != "TTPOS-FLAVOR-592" {
		t.Errorf("Expected Modifier[0].ID 'TTPOS-FLAVOR-592', got '%v'", modifier1.ID)
	}

	// 第二个 Modifier（TTPOS-SIZE-001）
	modifier2 := item.Modifiers[1]
	if modifier2.ID == nil || *modifier2.ID != "TTPOS-SIZE-001" {
		t.Errorf("Expected Modifier[1].ID 'TTPOS-SIZE-001', got '%v'", modifier2.ID)
	}

	// Price is now *int64 (minor unit): 20.0 THB = 2000 satang
	expectedModPrice := int64(2000)
	if modifier2.Price == nil || *modifier2.Price != expectedModPrice {
		t.Errorf("Expected Modifier[1].Price %d, got %v", expectedModPrice, modifier2.Price)
	}

	// 验证 Price（参考 Google Sheets：restaurantRevenue -> Subtotal/EaterPayment）
	// Price is now int64 (minor unit): 333.89 THB = 33389 satang
	expectedSubtotal := int64(33389)
	if order.Price.Subtotal != expectedSubtotal {
		t.Errorf("Expected Price.Subtotal %d, got %d", expectedSubtotal, order.Price.Subtotal)
	}

	if order.Price.EaterPayment == nil || *order.Price.EaterPayment != expectedSubtotal {
		t.Errorf("Expected Price.EaterPayment %d, got %v", expectedSubtotal, order.Price.EaterPayment)
	}
}

// TestRoundTripGrabSDK 测试 Grab SDK 双向转换
func TestRoundTripGrabSDK(t *testing.T) {
	// 原始订单
	cutlery := true
	partnerMerchantID := "partner-456"
	orderAcceptedType := "MANUAL"
	isMexEditOrder := false
	orderType := "Delivery"

	tax := int64(5000)
	deliveryFee := int64(2000)
	grabFundPromo := int64(500)
	merchantFundPromo := int64(500)
	eaterPayment := int64(56000)

	original := &grabsdk.SubmitOrderRequest{
		OrderID:           "order-001",
		ShortOrderNumber:  "0001",
		MerchantID:        "merchant-001",
		PartnerMerchantID: &partnerMerchantID,
		PaymentType:       "CASHLESS",
		Cutlery:           cutlery,
		OrderTime:         "2025-12-26T09:57:06Z",
		Currency: grabsdk.Currency{
			Code:     "THB",
			Symbol:   "฿",
			Exponent: 2,
		},
		FeatureFlags: grabsdk.OrderFeatureFlags{
			OrderAcceptedType: orderAcceptedType,
			IsMexEditOrder:    &isMexEditOrder,
			OrderType:         orderType,
		},
		Items: []grabsdk.OrderItem{
			{
				Id:         "item-001",
				GrabItemID: "grab-item-001",
				Quantity:   2,
				Price:      50000, // 500.00 THB
				Tax:        &tax,  // 50.00 THB
				Modifiers: []grabsdk.OrderItemModifier{
					{
						Id:       stringPtr("mod-001"),
						Quantity: int32Ptr(1),
						Price:    int64Ptr(1000), // 10.00 THB
					},
				},
			},
		},
		Price: grabsdk.OrderPrice{
			Subtotal:          50000,
			Tax:               &tax,
			DeliveryFee:       &deliveryFee,
			GrabFundPromo:     &grabFundPromo,
			MerchantFundPromo: &merchantFundPromo,
			EaterPayment:      &eaterPayment,
		},
	}

	// 转换为统一模型
	unified, err := FromGrabSDK(original)
	if err != nil {
		t.Fatalf("FromGrabSDK failed: %v", err)
	}

	// 转换回 Grab SDK
	converted := ToGrabSDK(unified)

	// 验证关键字段
	if converted.OrderID != original.OrderID {
		t.Errorf("OrderID mismatch: expected '%s', got '%s'", original.OrderID, converted.OrderID)
	}

	if converted.ShortOrderNumber != original.ShortOrderNumber {
		t.Errorf("ShortOrderNumber mismatch: expected '%s', got '%s'", original.ShortOrderNumber, converted.ShortOrderNumber)
	}

	if converted.PaymentType != original.PaymentType {
		t.Errorf("PaymentType mismatch: expected '%s', got '%s'", original.PaymentType, converted.PaymentType)
	}

	if len(converted.Items) != len(original.Items) {
		t.Fatalf("Items count mismatch: expected %d, got %d", len(original.Items), len(converted.Items))
	}

	// 验证商品
	origItem := original.Items[0]
	convItem := converted.Items[0]

	if convItem.Id != origItem.Id {
		t.Errorf("Item.Id mismatch: expected '%s', got '%s'", origItem.Id, convItem.Id)
	}

	if convItem.Quantity != origItem.Quantity {
		t.Errorf("Item.Quantity mismatch: expected %d, got %d", origItem.Quantity, convItem.Quantity)
	}

	if convItem.Price != origItem.Price {
		t.Errorf("Item.Price mismatch: expected %d, got %d", origItem.Price, convItem.Price)
	}
}

// TestRoundTripLineman 测试 Lineman 双向转换
func TestRoundTripLineman(t *testing.T) {
	// 原始订单
	original := &linemanv1.PlaceOrderReq{
		PartnerId:         "partner-002",
		StoreId:           "store-456",
		OrderId:           "LMF-260112-123456",
		OrderShortCode:    "6789",
		RestaurantRevenue: 450.50,
		OrderAcceptedTime: "2022-11-01T14:00:00+07:00",
		MemberId:          "member-999",
		CustomerType:      "PICKUP",
		Items: []linemanv1.OrderItem{
			{
				Id:          "item-002",
				Quantity:    3,
				UnitPrice:   150.16,
				Memo:        "Extra sauce",
				PromotionId: "promo-002",
				Discount:    5.0,
				Properties: []linemanv1.OrderItemProperty{
					{
						Id: "prop-001",
						Values: []linemanv1.OrderItemPropertyValue{
							{
								Id:    "val-001",
								Price: 15.0,
							},
						},
					},
				},
			},
		},
		AdditionalItems: []linemanv1.OrderAdditionalItem{
			{
				Name: "Special instructions",
			},
		},
	}

	// 转换为统一模型
	unified, err := FromLinemanPlaceOrder(original)
	if err != nil {
		t.Fatalf("FromLinemanPlaceOrder failed: %v", err)
	}

	// 转换回 Lineman
	converted := ToLinemanPlaceOrder(unified)

	// 验证关键字段
	if converted.OrderId != original.OrderId {
		t.Errorf("OrderId mismatch: expected '%s', got '%s'", original.OrderId, converted.OrderId)
	}

	if converted.OrderShortCode != original.OrderShortCode {
		t.Errorf("OrderShortCode mismatch: expected '%s', got '%s'", original.OrderShortCode, converted.OrderShortCode)
	}

	if converted.RestaurantRevenue != original.RestaurantRevenue {
		t.Errorf("RestaurantRevenue mismatch: expected %.2f, got %.2f", original.RestaurantRevenue, converted.RestaurantRevenue)
	}

	if converted.CustomerType != original.CustomerType {
		t.Errorf("CustomerType mismatch: expected '%s', got '%s'", original.CustomerType, converted.CustomerType)
	}

	if len(converted.Items) != len(original.Items) {
		t.Fatalf("Items count mismatch: expected %d, got %d", len(original.Items), len(converted.Items))
	}

	// 验证商品
	origItem := original.Items[0]
	convItem := converted.Items[0]

	if convItem.Id != origItem.Id {
		t.Errorf("Item.Id mismatch: expected '%s', got '%s'", origItem.Id, convItem.Id)
	}

	if convItem.Quantity != origItem.Quantity {
		t.Errorf("Item.Quantity mismatch: expected %d, got %d", origItem.Quantity, convItem.Quantity)
	}

	// Note: There may be slight floating point differences due to conversion
	// 150.16 * 100 = 15016, 15016 / 100 = 150.16
	if convItem.UnitPrice != origItem.UnitPrice {
		t.Errorf("Item.UnitPrice mismatch: expected %.2f, got %.2f", origItem.UnitPrice, convItem.UnitPrice)
	}

	if convItem.Memo != origItem.Memo {
		t.Errorf("Item.Memo mismatch: expected '%s', got '%s'", origItem.Memo, convItem.Memo)
	}
}

// TestJSONSerialization 测试 JSON 序列化
func TestJSONSerialization(t *testing.T) {
	membershipID := "member-001"
	partnerMerchantID := "partner-001"
	modID := "mod-001"
	modQty := int32(1)
	modPrice := int64(1000)
	eaterPayment := int64(20000)

	order := &message.TakeoutOrder{
		OrderID:           "test-order-001",
		ShortOrderNumber:  "0001",
		MerchantID:        "merchant-001",
		PartnerMerchantID: &partnerMerchantID,
		PaymentType:       "CASH",
		Cutlery:           true,
		OrderTime:         "2025-12-26T09:57:06Z",
		MembershipID:      &membershipID,
		Currency: message.TakeoutCurrency{
			Code:     "THB",
			Symbol:   "฿",
			Exponent: 2,
		},
		FeatureFlags: message.TakeoutFeatureFlags{
			OrderAcceptedType: "AUTO",
			OrderType:         "Delivery",
		},
		Items: []message.TakeoutOrderItem{
			{
				ID:         "item-001",
				GrabItemID: "grab-item-001",
				Quantity:   2,
				Price:      10050, // 100.50 in minor unit
				Modifiers: []message.TakeoutModifier{
					{
						ID:       &modID,
						Quantity: &modQty,
						Price:    &modPrice,
					},
				},
			},
		},
		Price: message.TakeoutOrderPrice{
			Subtotal:     20000, // 200.00 in minor unit
			EaterPayment: &eaterPayment,
		},
		AdditionalProperties: map[string]any{
			"testKey": "testValue",
		},
	}

	// 序列化
	data, err := json.Marshal(order)
	if err != nil {
		t.Fatalf("JSON Marshal failed: %v", err)
	}

	// 反序列化
	var decoded message.TakeoutOrder
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("JSON Unmarshal failed: %v", err)
	}

	// 验证
	if decoded.OrderID != order.OrderID {
		t.Errorf("OrderID mismatch after JSON round-trip: expected '%s', got '%s'", order.OrderID, decoded.OrderID)
	}

	if len(decoded.Items) != len(order.Items) {
		t.Errorf("Items count mismatch after JSON round-trip: expected %d, got %d", len(order.Items), len(decoded.Items))
	}

	// Verify currency
	if decoded.Currency.Code != order.Currency.Code {
		t.Errorf("Currency.Code mismatch after JSON round-trip: expected '%s', got '%s'", order.Currency.Code, decoded.Currency.Code)
	}

	// Verify feature flags
	if decoded.FeatureFlags.OrderType != order.FeatureFlags.OrderType {
		t.Errorf("FeatureFlags.OrderType mismatch after JSON round-trip: expected '%s', got '%s'", order.FeatureFlags.OrderType, decoded.FeatureFlags.OrderType)
	}
}

// TestNullableTypes 测试 Nullable 类型
func TestNullableTypes(t *testing.T) {
	// Test NullableTakeoutDineIn
	tableID := "T001"
	eaterCount := int64(4)
	dineIn := &message.TakeoutDineIn{
		TableID:    &tableID,
		EaterCount: &eaterCount,
	}

	nullable := message.NullableTakeoutDineIn{}
	if nullable.IsSet() {
		t.Error("Expected NullableTakeoutDineIn to be unset initially")
	}

	nullable.Set(dineIn)
	if !nullable.IsSet() {
		t.Error("Expected NullableTakeoutDineIn to be set after Set()")
	}

	got := nullable.Get()
	if got == nil || got.TableID == nil || *got.TableID != "T001" {
		t.Errorf("Expected TableID 'T001', got '%v'", got)
	}

	nullable.Unset()
	if nullable.IsSet() {
		t.Error("Expected NullableTakeoutDineIn to be unset after Unset()")
	}
}

// TestNullableTimeJSON 测试 NullableTime JSON 序列化
func TestNullableTimeJSON(t *testing.T) {
	now := time.Now()
	nullable := message.NewNullableTime(&now)

	// 序列化
	data, err := json.Marshal(nullable)
	if err != nil {
		t.Fatalf("JSON Marshal failed: %v", err)
	}

	// 反序列化
	var decoded message.NullableTime
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("JSON Unmarshal failed: %v", err)
	}

	if !decoded.IsSet() {
		t.Error("Expected NullableTime to be set after unmarshal")
	}

	got := decoded.Get()
	if got == nil {
		t.Error("Expected NullableTime value not nil")
	}
}

// 辅助函数

func int32Ptr(i int32) *int32 {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}
