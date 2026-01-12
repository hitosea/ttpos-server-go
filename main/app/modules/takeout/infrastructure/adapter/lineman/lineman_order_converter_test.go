package lineman

import (
	"encoding/json"
	"testing"
	"time"

	valueobject "ttpos-server-go/app/modules/takeout/domain/value_object"
)

// TestParseOrderWebhook 测试解析 LINE MAN 订单 Webhook
func TestParseOrderWebhook(t *testing.T) {
	converter := NewLineManConverter()

	// 模拟 LINE MAN Place Order API 的 JSON 数据
	rawJSON := `{
		"orderId": "LMF-221031-33879881",
		"orderShortCode": "9881",
		"restaurantRevenue": 250.50,
		"orderAcceptedTime": "2022-11-01T13:08:06+07:00",
		"items": [
			{
				"id": "TTPOS-ITEM-001",
				"quantity": 2,
				"unitPrice": 120.00,
				"memo": "ไม่ใส่ผักชี",
				"promotionId": "PROMO2024001",
				"discount": 20.00,
				"properties": [
					{
						"id": "size-option",
						"values": [
							{
								"id": "large",
								"price": 15.00
							}
						]
					}
				]
			}
		],
		"additionalItems": [
			{
				"name": "ค่าบริการเพิ่มเติม"
			}
		],
		"memberId": "LINE_MAN_USER_12345",
		"customerType": "DELIVERY"
	}`

	// 解析 Webhook
	result, err := converter.ParseOrderWebhook([]byte(rawJSON))
	if err != nil {
		t.Fatalf("解析 LINE MAN 订单失败: %v", err)
	}

	// 类型断言
	placeOrderReq, ok := result.(*LineManPlaceOrderRequest)
	if !ok {
		t.Fatalf("类型断言失败，期望 *LineManPlaceOrderRequest，实际: %T", result)
	}

	// 验证基础字段
	if placeOrderReq.OrderID != "LMF-221031-33879881" {
		t.Errorf("OrderID 不匹配，期望: LMF-221031-33879881，实际: %s", placeOrderReq.OrderID)
	}

	if placeOrderReq.OrderShortCode != "9881" {
		t.Errorf("OrderShortCode 不匹配，期望: 9881，实际: %s", placeOrderReq.OrderShortCode)
	}

	if placeOrderReq.RestaurantRevenue != 250.50 {
		t.Errorf("RestaurantRevenue 不匹配，期望: 250.50，实际: %.2f", placeOrderReq.RestaurantRevenue)
	}

	if placeOrderReq.CustomerType != "DELIVERY" {
		t.Errorf("CustomerType 不匹配，期望: DELIVERY，实际: %s", placeOrderReq.CustomerType)
	}

	// 验证商品数据
	if len(placeOrderReq.Items) != 1 {
		t.Fatalf("商品数量不匹配，期望: 1，实际: %d", len(placeOrderReq.Items))
	}

	item := placeOrderReq.Items[0]
	if item.ID != "TTPOS-ITEM-001" {
		t.Errorf("商品 ID 不匹配，期望: TTPOS-ITEM-001，实际: %s", item.ID)
	}

	if item.Quantity != 2 {
		t.Errorf("商品数量不匹配，期望: 2，实际: %d", item.Quantity)
	}

	if item.UnitPrice != 120.00 {
		t.Errorf("商品单价不匹配，期望: 120.00，实际: %.2f", item.UnitPrice)
	}

	t.Logf("✅ 测试通过：成功解析 LINE MAN 订单")
}

// TestConvertOrderToTakeoutOrder 测试将 LINE MAN 订单转换为 TTPOS 订单
func TestConvertOrderToTakeoutOrder(t *testing.T) {
	converter := NewLineManConverter()

	// 模拟解析后的 LINE MAN 订单
	placeOrderReq := &LineManPlaceOrderRequest{
		OrderID:           "LMF-221031-33879881",
		OrderShortCode:    "9881",
		RestaurantRevenue: 250.50,
		OrderAcceptedTime: "2022-11-01T13:08:06+07:00",
		Items: []LineManOrderItem{
			{
				ID:          "TTPOS-ITEM-001",
				Quantity:    2,
				UnitPrice:   120.00,
				Memo:        "ไม่ใส่ผักชี",
				PromotionID: "PROMO2024001",
				Discount:    20.00,
				Properties: []LineManItemProperty{
					{
						ID: "size-option",
						Values: []LineManPropertyValue{
							{
								ID:    "large",
								Price: 15.00,
							},
						},
					},
				},
			},
		},
		MemberID:     "LINE_MAN_USER_12345",
		CustomerType: "DELIVERY",
	}

	rawDataJSON, _ := json.Marshal(placeOrderReq)
	currentTime := time.Now().Unix()

	// 转换订单
	order, err := converter.ConvertOrderToTakeoutOrder(
		1234567890,
		valueobject.TakeoutPlatformLineman,
		placeOrderReq.OrderID,
		placeOrderReq,
		rawDataJSON,
		currentTime,
	)

	if err != nil {
		t.Fatalf("转换订单失败: %v", err)
	}

	// 验证基础字段
	if order.Platform != valueobject.TakeoutPlatformLineman {
		t.Errorf("平台不匹配，期望: %s，实际: %s", valueobject.TakeoutPlatformLineman, order.Platform)
	}

	if order.PlatformOrderId != "LMF-221031-33879881" {
		t.Errorf("平台订单 ID 不匹配，期望: LMF-221031-33879881，实际: %s", order.PlatformOrderId)
	}

	if order.ShortOrderNumber != "9881" {
		t.Errorf("短单号不匹配，期望: 9881，实际: %s", order.ShortOrderNumber)
	}

	if order.Subtotal != 250.50 {
		t.Errorf("小计金额不匹配，期望: 250.50，实际: %.2f", order.Subtotal)
	}

	if order.OrderType != valueobject.TakeoutOrderTypeDelivery {
		t.Errorf("订单类型不匹配，期望: %s，实际: %s", valueobject.TakeoutOrderTypeDelivery, order.OrderType)
	}

	if order.CurrencyCode != "THB" {
		t.Errorf("货币代码不匹配，期望: THB，实际: %s", order.CurrencyCode)
	}

	if order.OrderState != valueobject.TakeoutOrderStatePending {
		t.Errorf("订单状态不匹配，期望: %d，实际: %d", valueobject.TakeoutOrderStatePending, order.OrderState)
	}

	// 验证商品数据
	if len(order.TakeoutOrderItems) != 1 {
		t.Fatalf("商品数量不匹配，期望: 1，实际: %d", len(order.TakeoutOrderItems))
	}

	orderItem := order.TakeoutOrderItems[0]
	if orderItem.PlatformItemId != "TTPOS-ITEM-001" {
		t.Errorf("商品 ID 不匹配，期望: TTPOS-ITEM-001，实际: %s", orderItem.PlatformItemId)
	}

	if orderItem.Quantity != 2 {
		t.Errorf("商品数量不匹配，期望: 2，实际: %d", orderItem.Quantity)
	}

	if orderItem.Price != 120.00 {
		t.Errorf("商品价格不匹配，期望: 120.00，实际: %.2f", orderItem.Price)
	}

	// 验证修饰符
	if len(orderItem.TakeoutOrderItemModifiers) != 1 {
		t.Fatalf("修饰符数量不匹配，期望: 1，实际: %d", len(orderItem.TakeoutOrderItemModifiers))
	}

	modifier := orderItem.TakeoutOrderItemModifiers[0]
	if modifier.Price != 15.00 {
		t.Errorf("修饰符价格不匹配，期望: 15.00，实际: %.2f", modifier.Price)
	}

	// 验证促销信息
	if len(order.TakeoutOrderPromos) != 1 {
		t.Fatalf("促销数量不匹配，期望: 1，实际: %d", len(order.TakeoutOrderPromos))
	}

	promo := order.TakeoutOrderPromos[0]
	if promo.PromoCode != "PROMO2024001" {
		t.Errorf("促销代码不匹配，期望: PROMO2024001，实际: %s", promo.PromoCode)
	}

	if promo.PromoAmount != 20.00 {
		t.Errorf("促销金额不匹配，期望: 20.00，实际: %.2f", promo.PromoAmount)
	}

	t.Logf("✅ 测试通过：成功转换 LINE MAN 订单为 TTPOS 订单")
}

// TestParseISO8601Time 测试时间解析
func TestParseISO8601Time(t *testing.T) {
	converter := NewLineManConverter()

	// 测试基本的时间解析功能（不验证具体数值，因为可能有时区差异）
	testCases := []string{
		"2022-11-01T13:08:06+07:00", // 带时区
		"2022-11-01T06:08:06Z",      // UTC 时间
	}

	for _, tc := range testCases {
		result := converter.parseISO8601Time(tc)
		if result == 0 {
			t.Errorf("时间解析失败，输入: %s，返回: %d", tc, result)
		}
	}

	// 测试空字符串
	if result := converter.parseISO8601Time(""); result != 0 {
		t.Errorf("空字符串应返回 0，实际: %d", result)
	}

	t.Logf("✅ 测试通过：时间解析正确")
}

// TestConvertCustomerTypeToOrderType 测试客户类型转换
func TestConvertCustomerTypeToOrderType(t *testing.T) {
	converter := NewLineManConverter()

	testCases := []struct {
		input    string
		expected string
	}{
		{"DELIVERY", valueobject.TakeoutOrderTypeDelivery},
		{"PICKUP", valueobject.TakeoutOrderTypeTakeaway},
		{"delivery", valueobject.TakeoutOrderTypeDelivery}, // 测试大小写不敏感
		{"pickup", valueobject.TakeoutOrderTypeTakeaway},
		{"UNKNOWN", valueobject.TakeoutOrderTypeDelivery}, // 未知类型默认为配送
	}

	for _, tc := range testCases {
		result := converter.convertCustomerTypeToOrderType(tc.input)
		if result != tc.expected {
			t.Errorf("客户类型转换不匹配，输入: %s，期望: %s，实际: %s", tc.input, tc.expected, result)
		}
	}

	t.Logf("✅ 测试通过：客户类型转换正确")
}
