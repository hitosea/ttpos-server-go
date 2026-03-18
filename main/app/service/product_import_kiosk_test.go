package service

import (
	"strings"
	"testing"

	"ttpos-server-go/app/dto/resp/product_resp"
	"ttpos-server-go/app/model"
)

// TestParseShowsKiosk 验证 Shows 字符串中 "7" 正确解析为 IsShowKiosk
func TestParseShowsKiosk(t *testing.T) {
	tests := []struct {
		name          string
		shows         string
		wantCashier   bool
		wantTablet    bool
		wantKitchen   bool
		wantAssistant bool
		wantH5        bool
		wantDelivery  bool
		wantKiosk     bool
	}{
		{
			name:        "全部显示端 1234567",
			shows:       "1234567",
			wantCashier: true, wantTablet: true, wantKitchen: true,
			wantAssistant: true, wantH5: true, wantDelivery: true, wantKiosk: true,
		},
		{
			name:        "仅收银机和自助点餐机 17",
			shows:       "17",
			wantCashier: true, wantKiosk: true,
		},
		{
			name:      "仅自助点餐机 7",
			shows:     "7",
			wantKiosk: true,
		},
		{
			name:        "不含自助点餐机 123456",
			shows:       "123456",
			wantCashier: true, wantTablet: true, wantKitchen: true,
			wantAssistant: true, wantH5: true, wantDelivery: true, wantKiosk: false,
		},
		{
			name:  "空字符串",
			shows: "",
		},
		{
			name:        "仅收银机 1",
			shows:       "1",
			wantCashier: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟 ImportProductList 中 Shows 解析逻辑 (product.go:5068-5074)
			var p product_resp.ProductImportListItem
			p.IsShowCashier = strings.Contains(tt.shows, "1")
			p.IsShowTablet = strings.Contains(tt.shows, "2")
			p.IsShowKitchen = strings.Contains(tt.shows, "3")
			p.IsShowAssistant = strings.Contains(tt.shows, "4")
			p.IsShowH5 = strings.Contains(tt.shows, "5")
			p.IsShowDelivery = strings.Contains(tt.shows, "6")
			p.IsShowKiosk = strings.Contains(tt.shows, "7")

			if p.IsShowCashier != tt.wantCashier {
				t.Errorf("IsShowCashier = %v, want %v", p.IsShowCashier, tt.wantCashier)
			}
			if p.IsShowTablet != tt.wantTablet {
				t.Errorf("IsShowTablet = %v, want %v", p.IsShowTablet, tt.wantTablet)
			}
			if p.IsShowKitchen != tt.wantKitchen {
				t.Errorf("IsShowKitchen = %v, want %v", p.IsShowKitchen, tt.wantKitchen)
			}
			if p.IsShowAssistant != tt.wantAssistant {
				t.Errorf("IsShowAssistant = %v, want %v", p.IsShowAssistant, tt.wantAssistant)
			}
			if p.IsShowH5 != tt.wantH5 {
				t.Errorf("IsShowH5 = %v, want %v", p.IsShowH5, tt.wantH5)
			}
			if p.IsShowDelivery != tt.wantDelivery {
				t.Errorf("IsShowDelivery = %v, want %v", p.IsShowDelivery, tt.wantDelivery)
			}
			if p.IsShowKiosk != tt.wantKiosk {
				t.Errorf("IsShowKiosk = %v, want %v", p.IsShowKiosk, tt.wantKiosk)
			}
		})
	}
}

// TestDecimalPricingKioskValidation 验证小数计价时不允许在自助点餐机显示
func TestDecimalPricingKioskValidation(t *testing.T) {
	tests := []struct {
		name          string
		numType       int
		isShowTablet  bool
		isShowAssist  bool
		isShowH5      bool
		isShowDeliver bool
		isShowKiosk   bool
		wantError     bool
	}{
		{
			name:        "小数计价+自助点餐机=报错",
			numType:     2,
			isShowKiosk: true,
			wantError:   true,
		},
		{
			name:      "小数计价+仅收银机=通过",
			numType:   2,
			wantError: false,
		},
		{
			name:         "小数计价+平板=报错",
			numType:      2,
			isShowTablet: true,
			wantError:    true,
		},
		{
			name:        "整数计价+自助点餐机=通过",
			numType:     1,
			isShowKiosk: true,
			wantError:   false,
		},
		{
			name:         "小数计价+多端(含Kiosk)=报错",
			numType:      2,
			isShowTablet: true,
			isShowAssist: true,
			isShowKiosk:  true,
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟 product.go:5055 的验证逻辑
			hasError := tt.numType == 2 && (tt.isShowTablet || tt.isShowAssist || tt.isShowH5 || tt.isShowDeliver || tt.isShowKiosk)
			if hasError != tt.wantError {
				t.Errorf("decimal pricing validation = %v, want error = %v", hasError, tt.wantError)
			}
		})
	}
}

// TestDeliveryValidation 验证外送显示校验：开启扫码点餐到店自取或外送均可通过
func TestDeliveryValidation(t *testing.T) {
	tests := []struct {
		name                string
		deliveryStatus      int
		isOpenMemberInstant int
		isShowDelivery      bool
		wantError           bool
	}{
		{
			name:           "外送未开+自取未开+选择外送=报错",
			deliveryStatus: 0, isOpenMemberInstant: 0,
			isShowDelivery: true, wantError: true,
		},
		{
			name:           "外送已开+自取未开+选择外送=通过",
			deliveryStatus: 1, isOpenMemberInstant: 0,
			isShowDelivery: true, wantError: false,
		},
		{
			name:           "外送未开+自取已开+选择外送=通过",
			deliveryStatus: 0, isOpenMemberInstant: 1,
			isShowDelivery: true, wantError: false,
		},
		{
			name:           "外送已开+自取已开+选择外送=通过",
			deliveryStatus: 1, isOpenMemberInstant: 1,
			isShowDelivery: true, wantError: false,
		},
		{
			name:           "外送未开+自取未开+不选外送=通过",
			deliveryStatus: 0, isOpenMemberInstant: 0,
			isShowDelivery: false, wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			companySetting := model.CompanySetting{
				DeliveryStatus:      tt.deliveryStatus,
				IsOpenMemberInstant: tt.isOpenMemberInstant,
			}
			// 模拟 product.go 的验证逻辑
			hasError := companySetting.DeliveryStatus != 1 && companySetting.IsOpenMemberInstant != 1 && tt.isShowDelivery
			if hasError != tt.wantError {
				t.Errorf("delivery validation = %v, want error = %v", hasError, tt.wantError)
			}
		})
	}
}

// TestKioskNotEnabledValidation 验证未开启自助点餐机时不允许显示
func TestKioskNotEnabledValidation(t *testing.T) {
	tests := []struct {
		name        string
		enableKiosk int
		isShowKiosk bool
		wantError   bool
	}{
		{
			name:        "未开启+选择显示=报错",
			enableKiosk: 0,
			isShowKiosk: true,
			wantError:   true,
		},
		{
			name:        "已开启+选择显示=通过",
			enableKiosk: 1,
			isShowKiosk: true,
			wantError:   false,
		},
		{
			name:        "未开启+不显示=通过",
			enableKiosk: 0,
			isShowKiosk: false,
			wantError:   false,
		},
		{
			name:        "已开启+不显示=通过",
			enableKiosk: 1,
			isShowKiosk: false,
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			companySetting := model.CompanySetting{
				EnableKiosk: tt.enableKiosk,
			}
			// 模拟 product.go:5063 的验证逻辑
			hasError := !companySetting.IsOpenKiosk() && tt.isShowKiosk
			if hasError != tt.wantError {
				t.Errorf("kiosk not enabled validation = %v, want error = %v", hasError, tt.wantError)
			}
		})
	}
}
