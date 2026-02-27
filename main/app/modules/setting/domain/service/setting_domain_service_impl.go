package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"ttpos-server-go/app/modules/setting/domain/entity"
	"ttpos-server-go/app/modules/setting/domain/valueobject"
	"ttpos-server-go/pkg/context"

	"github.com/jinzhu/copier"
)

// SettingDomainServiceImpl 设置域服务实现
type SettingDomainServiceImpl struct{}

// NewSettingDomainService 创建设置域服务
func NewSettingDomainService() ISettingDomainService {
	return &SettingDomainServiceImpl{}
}

// VerifyPassword 验证密码
func (s *SettingDomainServiceImpl) VerifyPassword(ctx context.Context, source, passwordType, password string) bool {
	// 根据来源获取对应的设置，然后验证密码
	// 这里需要根据业务逻辑实现密码验证
	return false
}

// ProcessStoreSettingJson 处理门店设置JSON
func (s *SettingDomainServiceImpl) ProcessStoreSettingJson(jsonStr string) (entity.StoreSetting, error) {
	var store entity.StoreSetting

	// 解析json字符串为map进行处理，处理language字段
	var jsonMap map[string]any
	err := json.Unmarshal([]byte(jsonStr), &jsonMap)
	if err != nil {
		return store, err
	}

	// 处理language数组中的key字段
	if language, ok := jsonMap["language"].([]any); ok {
		for i, item := range language {
			if langItem, ok := item.(map[string]any); ok {
				// 尝试将key转换为string
				if keyNum, ok := langItem["key"].(string); ok {
					langItem["key"], _ = strconv.Atoi(keyNum)
					language[i] = langItem
				}
			}
		}
		jsonMap["language"] = language
	}

	if logoUrl, ok := jsonMap["logoUrl"].(string); ok {
		jsonMap["logo_url"] = logoUrl
	}

	if avatarUrl, ok := jsonMap["avatarUrl"].(string); ok {
		jsonMap["avatar_url"] = avatarUrl
	}

	// 重新序列化为JSON
	modifiedJSON, err := json.Marshal(jsonMap)
	if err != nil {
		return store, err
	}

	err = json.Unmarshal(modifiedJSON, &store)
	return store, err
}

// ProcessCashierSettingJson 处理收银机设置JSON
func (s *SettingDomainServiceImpl) ProcessCashierSettingJson(jsonStr string) (entity.CashierSetting, error) {
	var cashier entity.CashierSetting
	err := json.Unmarshal([]byte(jsonStr), &cashier)
	return cashier, err
}

// ProcessPrinterSettingJson 处理打印机设置JSON
func (s *SettingDomainServiceImpl) ProcessPrinterSettingJson(jsonStr string) (entity.PrinterSetting, error) {
	var printer entity.PrinterSetting
	err := json.Unmarshal([]byte(jsonStr), &printer)
	return printer, err
}

// ProcessBusinessSettingJson 处理业务设置JSON
func (s *SettingDomainServiceImpl) ProcessBusinessSettingJson(jsonStr string) (entity.BusinessSetting, error) {
	var business entity.BusinessSetting
	err := json.Unmarshal([]byte(jsonStr), &business)
	return business, err
}

// ProcessPointsSettingJson 处理积分设置JSON
func (s *SettingDomainServiceImpl) ProcessPointsSettingJson(jsonStr string) (entity.PointsSetting, error) {
	var points entity.PointsSetting
	err := json.Unmarshal([]byte(jsonStr), &points)
	return points, err
}

// ProcessBuffetSettingJson 处理自助餐设置JSON
func (s *SettingDomainServiceImpl) ProcessBuffetSettingJson(jsonStr string, buffet *entity.Buffet) error {
	return json.Unmarshal([]byte(jsonStr), buffet)
}

// ProcessTabletSettingJson 处理平板设置JSON
func (s *SettingDomainServiceImpl) ProcessTabletSettingJson(jsonStr string) (entity.TabletSetting, error) {
	var tablet entity.TabletSetting
	err := json.Unmarshal([]byte(jsonStr), &tablet)
	return tablet, err
}

// ConvertServiceFeeFormat 转换服务费格式
func (s *SettingDomainServiceImpl) ConvertServiceFeeFormat(oldVal string) ([]byte, error) {
	serviceFeeMap := map[string]interface{}{}
	err := json.Unmarshal([]byte(oldVal), &serviceFeeMap)
	if err != nil {
		return nil, err
	}
	if v, ok := serviceFeeMap["service_charge"]; ok {
		switch v.(type) {
		case float64:
			serviceFeeMap["service_charge"] = fmt.Sprintf("%f", v)
		}
	}
	return json.Marshal(serviceFeeMap)
}

// MergeWithDefaultStore 与默认门店设置合并
func (s *SettingDomainServiceImpl) MergeWithDefaultStore(defaultStore entity.StoreSetting, userStore entity.StoreSetting) entity.StoreSetting {
	// 与旧服务保持一致：将 TimeZoneList 设为 nil 以便合并后使用默认值
	userStore.TimeZoneList = nil
	_ = copier.CopyWithOption(&defaultStore, userStore, copier.Option{IgnoreEmpty: true})
	return defaultStore
}

// MergeWithDefaultCashier 与默认收银机设置合并
func (s *SettingDomainServiceImpl) MergeWithDefaultCashier(defaultCashier entity.CashierSetting, userCashier entity.CashierSetting) entity.CashierSetting {
	// 与旧服务保持一致：不合并语言设置
	defaultCashier.Language = nil
	// 接单语音，设备本地处理，不需要合并
	userCashier.IsAutoVoice = ""
	_ = copier.CopyWithOption(&defaultCashier, userCashier, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	return defaultCashier
}

// MergeWithDefaultPrinter 与默认打印机设置合并
func (s *SettingDomainServiceImpl) MergeWithDefaultPrinter(defaultPrinter entity.PrinterSetting, userPrinter entity.PrinterSetting) entity.PrinterSetting {
	// 与旧服务保持一致
	defaultPrinter.Language = nil
	_ = copier.CopyWithOption(&defaultPrinter, userPrinter, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	return defaultPrinter
}

// MergeWithDefaultBusiness 与默认业务设置合并
func (s *SettingDomainServiceImpl) MergeWithDefaultBusiness(defaultBusiness entity.BusinessSetting, userBusiness entity.BusinessSetting) entity.BusinessSetting {
	// 与旧服务保持一致：将列表设为 nil 以便使用默认值
	userBusiness.ZeroingMethodList = nil
	userBusiness.CheckoutZeroingMethodList = nil
	userBusiness.GiftMethodList = nil
	userBusiness.FreeMethodList = nil
	_ = copier.CopyWithOption(&defaultBusiness, userBusiness, copier.Option{IgnoreEmpty: true})
	return defaultBusiness
}

// FilterValidLanguages 过滤有效语言
func (s *SettingDomainServiceImpl) FilterValidLanguages(languages []valueobject.LanguageItem, availableLanguages []string) []valueobject.LanguageItem {
	var validLanguages []valueobject.LanguageItem
	for _, lang := range languages {
		for _, available := range availableLanguages {
			if lang.Name == available {
				validLanguages = append(validLanguages, lang)
				break
			}
		}
	}
	return validLanguages
}

// AddImageDomainToCarousel 为轮播图添加域名
func (s *SettingDomainServiceImpl) AddImageDomainToCarousel(carousel []valueobject.CarouselItem, baseURL string) []valueobject.CarouselItem {
	result := make([]valueobject.CarouselItem, len(carousel))
	for i, item := range carousel {
		result[i] = item
		// 这里应该实现添加域名的逻辑
		// result[i].FilePath = utils.AddImageDomain(item.FilePath, baseURL, true)
	}
	return result
}

// ParseCoordinates 解析坐标
func (s *SettingDomainServiceImpl) ParseCoordinates(coordinates string) (latitude, longitude string) {
	// 解析经纬度坐标
	return "", ""
}

// MergeWithDefaultPoints 与默认积分设置合并
func (s *SettingDomainServiceImpl) MergeWithDefaultPoints(defaultPoints entity.PointsSetting, userPoints entity.PointsSetting) entity.PointsSetting {
	result := defaultPoints

	// 如果用户有设置，则覆盖默认值
	if userPoints.DeductionOrder != "" {
		result.DeductionOrder = userPoints.DeductionOrder
	}
	if userPoints.DeductRatioMain != "" {
		result.DeductRatioMain = userPoints.DeductRatioMain
	}
	if userPoints.DeductRatioGift != "" {
		result.DeductRatioGift = userPoints.DeductRatioGift
	}
	if userPoints.PointsName != "" {
		result.PointsName = userPoints.PointsName
	}
	if userPoints.IsShoppingGift != "" {
		result.IsShoppingGift = userPoints.IsShoppingGift
	}
	if userPoints.GiftRatio != "" {
		result.GiftRatio = userPoints.GiftRatio
	}
	if userPoints.IsShoppingDiscount != "" {
		result.IsShoppingDiscount = userPoints.IsShoppingDiscount
	}
	if userPoints.Describe != "" {
		result.Describe = userPoints.Describe
	}

	return result
}

// MergeWithDefaultBuffet 与默认自助餐设置合并
func (s *SettingDomainServiceImpl) MergeWithDefaultBuffet(defaultBuffet entity.Buffet, userBuffet entity.Buffet) entity.Buffet {
	result := defaultBuffet

	// 如果用户有设置，则覆盖默认值
	if userBuffet.IsOpen != "" {
		result.IsOpen = userBuffet.IsOpen
	}
	if userBuffet.TabletEndTime != "" {
		result.TabletEndTime = userBuffet.TabletEndTime
	}
	if userBuffet.IsRemainContinue != "" {
		result.IsRemainContinue = userBuffet.IsRemainContinue
	}
	if userBuffet.RemainContinueTime != "" {
		result.RemainContinueTime = userBuffet.RemainContinueTime
	}
	if userBuffet.RemainContinueNoticeTime != "" {
		result.RemainContinueNoticeTime = userBuffet.RemainContinueNoticeTime
	}
	if userBuffet.IsBuyContinue != "" {
		result.IsBuyContinue = userBuffet.IsBuyContinue
	}
	if userBuffet.IsAddClock != "" {
		result.IsAddClock = userBuffet.IsAddClock
	}
	if userBuffet.IsBuffetDiscount != "" {
		result.IsBuffetDiscount = userBuffet.IsBuffetDiscount
	}
	if userBuffet.IsShowNonBuffetProduct != "" {
		result.IsShowNonBuffetProduct = userBuffet.IsShowNonBuffetProduct
	}
	if len(userBuffet.AddClock) > 0 {
		result.AddClock = userBuffet.AddClock
	}

	return result
}

// MergeWithDefaultTablet 与默认平板设置合并
func (s *SettingDomainServiceImpl) MergeWithDefaultTablet(defaultTablet entity.TabletSetting, userTablet entity.TabletSetting) entity.TabletSetting {
	// 与旧服务保持一致：语言不需要合并
	defaultTablet.Language = nil
	_ = copier.CopyWithOption(&defaultTablet, userTablet, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	return defaultTablet
}

// MergeWithDefaultPayment 与默认支付设置合并
func (s *SettingDomainServiceImpl) MergeWithDefaultPayment(defaultPayment entity.PaymentSetting, userPayment entity.PaymentSetting) entity.PaymentSetting {
	result := defaultPayment

	// 如果用户有设置，则覆盖默认值
	if userPayment.IsBalance != "" {
		result.IsBalance = userPayment.IsBalance
	}
	if userPayment.IsCash != "" {
		result.IsCash = userPayment.IsCash
	}
	if userPayment.IsBankCard != "" {
		result.IsBankCard = userPayment.IsBankCard
	}
	if userPayment.IsWechat != "" {
		result.IsWechat = userPayment.IsWechat
	}
	if userPayment.IsAlipay != "" {
		result.IsAlipay = userPayment.IsAlipay
	}
	if userPayment.IsUnionPay != "" {
		result.IsUnionPay = userPayment.IsUnionPay
	}
	if userPayment.IsOther != "" {
		result.IsOther = userPayment.IsOther
	}

	return result
}

// MergeWithDefaultCurrency 与默认货币设置合并
func (s *SettingDomainServiceImpl) MergeWithDefaultCurrency(defaultCurrency entity.CurrencySetting, userCurrency entity.CurrencySetting) entity.CurrencySetting {
	result := defaultCurrency

	// 如果用户有设置，则覆盖默认值
	if userCurrency.Unit != "" {
		result.Unit = userCurrency.Unit
	}
	if userCurrency.PrintUnit != "" {
		result.PrintUnit = userCurrency.PrintUnit
	}
	if userCurrency.UnitPosition != "" {
		result.UnitPosition = userCurrency.UnitPosition
	}
	if userCurrency.IsOpen != "" {
		result.IsOpen = userCurrency.IsOpen
	}
	if userCurrency.ViceUnitPosition != "" {
		result.ViceUnitPosition = userCurrency.ViceUnitPosition
	}

	return result
}

// MergeWithDefaultServiceCharge 与默认服务费设置合并
func (s *SettingDomainServiceImpl) MergeWithDefaultServiceCharge(defaultServiceCharge entity.ServiceCharge, userServiceCharge entity.ServiceCharge) entity.ServiceCharge {
	result := defaultServiceCharge

	// 如果用户有设置，则覆盖默认值
	if userServiceCharge.IsOpen != "" {
		result.IsOpen = userServiceCharge.IsOpen
	}
	if userServiceCharge.ChargeType != "" {
		result.ChargeType = userServiceCharge.ChargeType
	}
	if userServiceCharge.ServiceCharge != "" {
		result.ServiceCharge = userServiceCharge.ServiceCharge
	}
	if userServiceCharge.ServiceChargeRate != "" {
		result.ServiceChargeRate = userServiceCharge.ServiceChargeRate
	}
	if userServiceCharge.IsOpenTax != "" {
		result.IsOpenTax = userServiceCharge.IsOpenTax
	}
	if userServiceCharge.ApplyScope != "" {
		result.ApplyScope = userServiceCharge.ApplyScope
	}
	if userServiceCharge.ApplyScopeOrdering != "" {
		result.ApplyScopeOrdering = userServiceCharge.ApplyScopeOrdering
	}
	if userServiceCharge.ApplyScopeTable != "" {
		result.ApplyScopeTable = userServiceCharge.ApplyScopeTable
	}
	if len(userServiceCharge.ApplyScopeTableList) > 0 {
		result.ApplyScopeTableList = userServiceCharge.ApplyScopeTableList
	}
	if userServiceCharge.ServiceFeeBase != "" {
		result.ServiceFeeBase = userServiceCharge.ServiceFeeBase
	}

	return result
}

// MergeWithDefaultTaxRate 与默认税率设置合并
func (s *SettingDomainServiceImpl) MergeWithDefaultTaxRate(defaultTaxRate entity.TaxRate, userTaxRate entity.TaxRate) entity.TaxRate {
	result := defaultTaxRate

	// 如果用户有设置，则覆盖默认值
	if userTaxRate.IsOpen != "" {
		result.IsOpen = userTaxRate.IsOpen
	}
	if userTaxRate.CalcType != "" {
		result.CalcType = userTaxRate.CalcType
	}
	if len(userTaxRate.AddTaxCategory) > 0 {
		result.AddTaxCategory = userTaxRate.AddTaxCategory
	}

	return result
}
