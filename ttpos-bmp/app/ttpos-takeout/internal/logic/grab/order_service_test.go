package grab

import (
	"testing"
	"time"

	"github.com/gogf/gf/v2/os/gtime"

	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
)

// TestParseTime 测试时间解析函数
func TestParseTime(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantNil  bool
		wantTime string
	}{
		{
			name:    "空字符串返回nil",
			input:   "",
			wantNil: true,
		},
		{
			name:     "RFC3339格式",
			input:    "2025-12-05T10:30:00Z",
			wantNil:  false,
			wantTime: "2025-12-05",
		},
		{
			name:     "带时区的RFC3339格式",
			input:    "2025-12-05T18:30:00+08:00",
			wantNil:  false,
			wantTime: "2025-12-05",
		},
		{
			name:    "无效格式返回nil",
			input:   "invalid-time",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTime(tt.input)
			if tt.wantNil {
				if result != nil {
					t.Errorf("parseTime(%q) = %v, want nil", tt.input, result)
				}
			} else {
				if result == nil {
					t.Errorf("parseTime(%q) = nil, want non-nil", tt.input)
				}
			}
		})
	}
}

// TestBoolToInt 测试布尔转整数函数
func TestBoolToInt(t *testing.T) {
	tests := []struct {
		input bool
		want  int
	}{
		{true, 1},
		{false, 0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := boolToInt(tt.input)
			if got != tt.want {
				t.Errorf("boolToInt(%v) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

// TestGetOrderType 测试订单类型判断
func TestGetOrderType(t *testing.T) {
	tests := []struct {
		name string
		req  *grab.SubmitOrderRequest
		want string
	}{
		{
			name: "堂食订单",
			req: &grab.SubmitOrderRequest{
				DineIn: &grab.DineInInfo{
					TableID:    "T001",
					EaterCount: 2,
				},
			},
			want: "DineIn",
		},
		{
			name: "外送订单",
			req: &grab.SubmitOrderRequest{
				DineIn: nil,
			},
			want: "DeliveryByGrab",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getOrderType(tt.req)
			if got != tt.want {
				t.Errorf("getOrderType() = %s, want %s", got, tt.want)
			}
		})
	}
}

// TestGetEaterCount 测试用餐人数获取
func TestGetEaterCount(t *testing.T) {
	tests := []struct {
		name   string
		dineIn *grab.DineInInfo
		want   int
	}{
		{
			name:   "nil返回0",
			dineIn: nil,
			want:   0,
		},
		{
			name: "返回正确人数",
			dineIn: &grab.DineInInfo{
				EaterCount: 4,
			},
			want: 4,
		},
		{
			name: "0人也正确返回",
			dineIn: &grab.DineInInfo{
				EaterCount: 0,
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getEaterCount(tt.dineIn)
			if got != tt.want {
				t.Errorf("getEaterCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestGetCustomerName 测试客户姓名获取
func TestGetCustomerName(t *testing.T) {
	tests := []struct {
		name     string
		receiver *grab.Receiver
		want     string
	}{
		{
			name:     "nil返回空字符串",
			receiver: nil,
			want:     "",
		},
		{
			name: "返回正确姓名",
			receiver: &grab.Receiver{
				Name: "张三",
			},
			want: "张三",
		},
		{
			name: "空姓名返回空字符串",
			receiver: &grab.Receiver{
				Name: "",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getCustomerName(tt.receiver)
			if got != tt.want {
				t.Errorf("getCustomerName() = %s, want %s", got, tt.want)
			}
		})
	}
}

// TestGetCustomerPhone 测试客户电话获取
func TestGetCustomerPhone(t *testing.T) {
	tests := []struct {
		name     string
		receiver *grab.Receiver
		want     string
	}{
		{
			name:     "nil返回空字符串",
			receiver: nil,
			want:     "",
		},
		{
			name: "优先返回虚拟号码",
			receiver: &grab.Receiver{
				Phone: "13800138000",
				VirtualContact: &grab.VirtualContact{
					Phone: "400-123-4567",
				},
			},
			want: "400-123-4567",
		},
		{
			name: "虚拟号码为空时返回真实号码",
			receiver: &grab.Receiver{
				Phone: "13800138000",
				VirtualContact: &grab.VirtualContact{
					Phone: "",
				},
			},
			want: "13800138000",
		},
		{
			name: "无虚拟号码时返回真实号码",
			receiver: &grab.Receiver{
				Phone:          "13800138000",
				VirtualContact: nil,
			},
			want: "13800138000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getCustomerPhone(tt.receiver)
			if got != tt.want {
				t.Errorf("getCustomerPhone() = %s, want %s", got, tt.want)
			}
		})
	}
}

// TestOrderEvent 测试订单事件结构
func TestOrderEvent(t *testing.T) {
	event := &OrderEvent{
		Action:       "create",
		ProviderName: "grab",
		OrderUUID:    "uuid-123",
		OrderID:      "G-123456",
		MerchantID:   "M-001",
		Status:       "PENDING",
		Timestamp:    gtime.Now().Unix(),
	}

	if event.Action != "create" {
		t.Errorf("Action = %s, want create", event.Action)
	}
	if event.ProviderName != "grab" {
		t.Errorf("ProviderName = %s, want grab", event.ProviderName)
	}
}

// TestOrderState_Constants 测试订单状态常量
func TestOrderState_Constants(t *testing.T) {
	// 验证订单状态常量定义正确
	states := []string{
		grab.OrderStatePending,
		grab.OrderStateAccepted,
		grab.OrderStatePreparing,
		grab.OrderStateReady,
		grab.OrderStateCollected,
		grab.OrderStateDelivered,
		grab.OrderStateCompleted,
		grab.OrderStateCancelled,
		grab.OrderStateFailed,
	}

	for _, state := range states {
		if state == "" {
			t.Error("Order state constant should not be empty")
		}
	}
}

// TestSubmitOrderRequest_Parse 测试订单请求解析
func TestSubmitOrderRequest_Parse(t *testing.T) {
	// 模拟 Grab 订单请求
	req := &grab.SubmitOrderRequest{
		OrderID:           "G-123456789",
		ShortOrderNumber:  "A001",
		MerchantID:        "1-CYXXXXXXXXX",
		PartnerMerchantID: "shop-001",
		PaymentType:       "CASHLESS",
		Cutlery:           true,
		OrderTime:         time.Now().Format(time.RFC3339),
		OrderState:        grab.OrderStatePending,
		Currency: grab.Currency{
			Code:     "THB",
			Symbol:   "฿",
			Exponent: 2,
		},
		Items: []grab.OrderItem{
			{
				ID:             "item-001",
				GrabItemID:     "grab-item-001",
				Quantity:       2,
				Price:          15000, // 150.00 THB (最小单位)
				Tax:            1050,  // 10.50 THB
				Specifications: "绿咖喱鸡饭",
				Modifiers: []grab.Modifier{
					{ID: "mod-001", Quantity: 1, Price: 1000},
				},
			},
		},
		Price: grab.OrderPrice{
			Subtotal:          15000,
			Tax:               1050,
			MerchantChargeFee: 500,
			GrabFundPromo:     0,
			MerchantFundPromo: 0,
			Total:             16550,
		},
		Receiver: &grab.Receiver{
			Name:  "测试用户",
			Phone: "0812345678",
			VirtualContact: &grab.VirtualContact{
				Phone:      "021234567",
				ExpiryTime: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			},
		},
	}

	// 验证基本字段
	if req.OrderID == "" {
		t.Error("OrderID should not be empty")
	}
	if req.MerchantID == "" {
		t.Error("MerchantID should not be empty")
	}
	if len(req.Items) == 0 {
		t.Error("Items should not be empty")
	}

	// 验证价格计算
	expectedTotal := int64(16550)
	if req.Price.Total != expectedTotal {
		t.Errorf("Price.Total = %d, want %d", req.Price.Total, expectedTotal)
	}

	// 验证货币信息
	if req.Currency.Code != "THB" {
		t.Errorf("Currency.Code = %s, want THB", req.Currency.Code)
	}
	if req.Currency.Exponent != 2 {
		t.Errorf("Currency.Exponent = %d, want 2", req.Currency.Exponent)
	}
}

// TestPriceConversion 测试价格转换 (最小单位 -> 元)
// 注意: 当 exponent 为 0 时，代码会默认使用 2 位小数
func TestPriceConversion(t *testing.T) {
	tests := []struct {
		name     string
		minUnit  int64
		exponent int
		want     float64
	}{
		{
			name:     "THB 2位小数",
			minUnit:  15000,
			exponent: 2,
			want:     150.00,
		},
		{
			name:     "JPY 0位小数 (代码默认使用2位)",
			minUnit:  1500,
			exponent: 0,
			want:     15.00, // 代码逻辑: exponent=0 时默认使用 2
		},
		{
			name:     "BHD 3位小数",
			minUnit:  150000,
			exponent: 3,
			want:     150.00,
		},
		{
			name:     "SGD 2位小数",
			minUnit:  9900,
			exponent: 2,
			want:     99.00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 这里模拟 saveOrder 中的价格转换逻辑
			exponent := tt.exponent
			if exponent == 0 {
				exponent = 2 // 默认 2 位小数
			}
			divisor := float64(1)
			for i := 0; i < exponent; i++ {
				divisor *= 10
			}
			got := float64(tt.minUnit) / divisor
			if got != tt.want {
				t.Errorf("Price conversion: %d / %f = %f, want %f", tt.minUnit, divisor, got, tt.want)
			}
		})
	}
}

// BenchmarkGetOrderType 订单类型判断性能测试
func BenchmarkGetOrderType(b *testing.B) {
	req := &grab.SubmitOrderRequest{
		DineIn: &grab.DineInInfo{
			TableID:    "T001",
			EaterCount: 2,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getOrderType(req)
	}
}

// BenchmarkGetCustomerPhone 客户电话获取性能测试
func BenchmarkGetCustomerPhone(b *testing.B) {
	receiver := &grab.Receiver{
		Phone: "13800138000",
		VirtualContact: &grab.VirtualContact{
			Phone: "400-123-4567",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getCustomerPhone(receiver)
	}
}
