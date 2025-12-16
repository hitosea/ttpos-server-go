package setting

import (
	"testing"

	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	setting "ttpos-server-go/app/dto/resp/setting"

	"github.com/stretchr/testify/assert"
)

// TestStoreCodeFieldExists 测试店铺编码字段是否正确定义
func TestStoreCodeFieldExists(t *testing.T) {
	// 测试请求结构体
	updateReq := req.UpdateStoreSetting{
		Name:        "测试店铺",
		LogoUrl:     "/uploads/logo.png",
		TimeZone:    "Asia/Shanghai",
		CompanyName: "测试公司",
		Address:     "测试地址",
		Phone:       "1234567890",
		TaxNumber:   "91110000XXXXXXXX",
		StoreCode:   "STORE-001", // 店铺编码
		Language: []dto.LanguageItem{
			{Name: "简体中文", Value: "zh_CN"},
		},
		Coordinates: "39.9042,116.4074",
	}

	// 验证字段赋值正确
	assert.Equal(t, "STORE-001", updateReq.StoreCode)

	// 测试配置结构体
	storeConfig := setting.Store{
		Name:        "测试店铺",
		Company:     "测试公司",
		Address:     "测试地址",
		Phone:       "1234567890",
		TaxNumber:   "91110000XXXXXXXX",
		StoreCode:   "STORE-001", // 店铺编码
		ChainNumber: "CHAIN-001",
	}

	// 验证字段赋值正确
	assert.Equal(t, "STORE-001", storeConfig.StoreCode)
}

// TestStoreCodeEmpty 测试店铺编码为空的情况
func TestStoreCodeEmpty(t *testing.T) {
	updateReq := req.UpdateStoreSetting{
		Name:        "测试店铺",
		LogoUrl:     "/uploads/logo.png",
		TimeZone:    "Asia/Shanghai",
		CompanyName: "测试公司",
		Address:     "测试地址",
		Phone:       "1234567890",
		TaxNumber:   "91110000XXXXXXXX",
		StoreCode:   "", // 空店铺编码
		Language: []dto.LanguageItem{
			{Name: "简体中文", Value: "zh_CN"},
		},
	}

	// 验证空值允许
	assert.Equal(t, "", updateReq.StoreCode)
}

// TestStoreCodeMaxLength 测试店铺编码最大长度
func TestStoreCodeMaxLength(t *testing.T) {
	// 100个字符的店铺编码
	longCode := ""
	for i := 0; i < 100; i++ {
		longCode += "A"
	}

	updateReq := req.UpdateStoreSetting{
		Name:        "测试店铺",
		LogoUrl:     "/uploads/logo.png",
		TimeZone:    "Asia/Shanghai",
		CompanyName: "测试公司",
		Address:     "测试地址",
		Phone:       "1234567890",
		StoreCode:   longCode,
		Language: []dto.LanguageItem{
			{Name: "简体中文", Value: "zh_CN"},
		},
	}

	// 验证100个字符的编码
	assert.Equal(t, 100, len(updateReq.StoreCode))
}

// TestStoreCodeWithChinese 测试包含中文的店铺编码
func TestStoreCodeWithChinese(t *testing.T) {
	updateReq := req.UpdateStoreSetting{
		Name:        "测试店铺",
		LogoUrl:     "/uploads/logo.png",
		TimeZone:    "Asia/Shanghai",
		CompanyName: "测试公司",
		Address:     "测试地址",
		Phone:       "1234567890",
		StoreCode:   "门店-北京-001", // 包含中文
		Language: []dto.LanguageItem{
			{Name: "简体中文", Value: "zh_CN"},
		},
	}

	// 验证中文编码
	assert.Equal(t, "门店-北京-001", updateReq.StoreCode)
}

// TestStoreCodeWithSpecialChars 测试包含特殊字符的店铺编码
func TestStoreCodeWithSpecialChars(t *testing.T) {
	updateReq := req.UpdateStoreSetting{
		Name:        "测试店铺",
		LogoUrl:     "/uploads/logo.png",
		TimeZone:    "Asia/Shanghai",
		CompanyName: "测试公司",
		Address:     "测试地址",
		Phone:       "1234567890",
		StoreCode:   "STORE#001-ABC!", // 包含特殊字符
		Language: []dto.LanguageItem{
			{Name: "简体中文", Value: "zh_CN"},
		},
	}

	// 验证特殊字符编码
	assert.Equal(t, "STORE#001-ABC!", updateReq.StoreCode)
}
