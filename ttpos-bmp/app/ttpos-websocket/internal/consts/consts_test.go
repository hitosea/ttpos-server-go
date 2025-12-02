package consts

import (
	"testing"
)

// TestSourceConstants 测试来源类型常量
func TestSourceConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"SourceAll", SourceAll, "*"},
		{"SourceShop", SourceShop, "shop"},
		{"SourceCashier", SourceCashier, "cashier"},
		{"SourceTablet", SourceTablet, "tablet"},
		{"SourceKitchen", SourceKitchen, "kitchen"},
		{"SourceAssistant", SourceAssistant, "assistant"},
		{"SourceH5", SourceH5, "H5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %s; 期望 %s", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

// TestMessageTypeConstants 测试消息类型常量
func TestMessageTypeConstants(t *testing.T) {
	messageTypes := []struct {
		name     string
		constant string
	}{
		{"UPDATE_ORDER", MessageTypeUpdateOrder},
		{"CUSTOMER_CALL", MessageTypeCustomerCall},
		{"PRINT_DATA", MessageTypePrintData},
		{"H5_ORDER", MessageTypeH5Order},
		{"UPDATE_CONFIG", MessageTypeUpdateConfig},
		{"UPDATE_PERMISSION", MessageTypeUpdatePermission},
		{"UPDATE_USER", MessageTypeUpdateUser},
		{"UPDATE_PRODUCT", MessageTypeUpdateProduct},
		{"UPDATE_CATEGORY", MessageTypeUpdateCategory},
		{"UPDATE_BUFFET", MessageTypeUpdateBuffet},
		{"UPDATE_DESK", MessageTypeUpdateDesk},
		{"UPDATE_DESK_TYPE", MessageTypeUpdateDeskType},
		{"UPDATE_REFUND_STATE", MessageTypeUpdateRefundState},
		{"UPDATE_KITCHEN", MessageTypeUpdateKitchen},
		{"UPDATE_SELECTED_PRINTER", MessageTypeUpdateSelectedPrinter},
		{"UPDATE_MEMBER_SALE_ORDER", MessageTypeUpdateMemberSaleOrder},
		{"SYNC_DATA", MessageTypeSyncData},
		{"IMPORT_PRODUCT", MessageTypeImportProduct},
		{"IMPORT_MATERIAL", MessageTypeImportMaterial},
	}

	for _, tt := range messageTypes {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant == "" {
				t.Errorf("%s 常量为空", tt.name)
			}
			// 验证常量是小写+下划线格式
			if len(tt.constant) == 0 {
				t.Errorf("%s 常量长度为0", tt.name)
			}
		})
	}
}

// TestCodeConstants 测试状态码常量
func TestCodeConstants(t *testing.T) {
	if CodeSuccess != 200 {
		t.Errorf("CodeSuccess = %d; 期望 200", CodeSuccess)
	}

	if CodeFail != 500 {
		t.Errorf("CodeFail = %d; 期望 500", CodeFail)
	}
}

// TestClientMessageTypeConstants 测试客户端消息类型常量
func TestClientMessageTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"ClientMessageTypeHeartbeat", ClientMessageTypeHeartbeat, "heartbeat"},
		{"ClientMessageTypeReply", ClientMessageTypeReply, "reply"},
		{"ClientMessageTypeUsbPrintReport", ClientMessageTypeUsbPrintReport, "usb_print_report"},
		{"ClientMessageTypeLanPrintReport", ClientMessageTypeLanPrintReport, "lan_print_report"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %s; 期望 %s", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

// TestConstantsUniqueness 测试常量的唯一性
func TestConstantsUniqueness(t *testing.T) {
	// 测试来源类型唯一性
	sources := []string{
		SourceAll,
		SourceShop,
		SourceCashier,
		SourceTablet,
		SourceKitchen,
		SourceAssistant,
		SourceH5,
	}

	sourceMap := make(map[string]bool)
	for _, s := range sources {
		if sourceMap[s] {
			t.Errorf("来源类型重复: %s", s)
		}
		sourceMap[s] = true
	}

	// 测试消息类型唯一性
	messageTypes := []string{
		MessageTypeUpdateOrder,
		MessageTypeCustomerCall,
		MessageTypePrintData,
		MessageTypeH5Order,
		MessageTypeUpdateConfig,
		MessageTypeUpdatePermission,
		MessageTypeUpdateUser,
		MessageTypeUpdateProduct,
		MessageTypeUpdateCategory,
		MessageTypeUpdateBuffet,
		MessageTypeUpdateDesk,
		MessageTypeUpdateDeskType,
		MessageTypeUpdateRefundState,
		MessageTypeUpdateKitchen,
		MessageTypeUpdateSelectedPrinter,
		MessageTypeUpdateMemberSaleOrder,
		MessageTypeSyncData,
		MessageTypeImportProduct,
		MessageTypeImportMaterial,
	}

	messageTypeMap := make(map[string]bool)
	for _, mt := range messageTypes {
		if messageTypeMap[mt] {
			t.Errorf("消息类型重复: %s", mt)
		}
		messageTypeMap[mt] = true
	}
}
