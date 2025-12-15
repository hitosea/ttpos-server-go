package service

import (
	"ttpos-server-go/app/modules/setting/domain/entity"
	"ttpos-server-go/app/modules/setting/domain/valueobject"
	"ttpos-server-go/pkg/context"
)

// ISettingDomainService 设置域服务接口
type ISettingDomainService interface {
	// 密码验证相关
	VerifyPassword(ctx context.Context, source, passwordType, password string) bool

	// 设置值处理相关
	ProcessStoreSettingJson(jsonStr string) (entity.StoreSetting, error)
	ProcessCashierSettingJson(jsonStr string) (entity.CashierSetting, error)
	ProcessPrinterSettingJson(jsonStr string) (entity.PrinterSetting, error)
	ProcessBusinessSettingJson(jsonStr string) (entity.BusinessSetting, error)
	ProcessPointsSettingJson(jsonStr string) (entity.PointsSetting, error)
	ProcessBuffetSettingJson(jsonStr string, buffet *entity.Buffet) error
	ProcessTabletSettingJson(jsonStr string) (entity.TabletSetting, error)
	ConvertServiceFeeFormat(oldVal string) ([]byte, error)

	// 设置值合并相关
	MergeWithDefaultStore(defaultStore entity.StoreSetting, userStore entity.StoreSetting) entity.StoreSetting
	MergeWithDefaultCashier(defaultCashier entity.CashierSetting, userCashier entity.CashierSetting) entity.CashierSetting
	MergeWithDefaultPrinter(defaultPrinter entity.PrinterSetting, userPrinter entity.PrinterSetting) entity.PrinterSetting
	MergeWithDefaultBusiness(defaultBusiness entity.BusinessSetting, userBusiness entity.BusinessSetting) entity.BusinessSetting
	MergeWithDefaultPoints(defaultPoints entity.PointsSetting, userPoints entity.PointsSetting) entity.PointsSetting
	MergeWithDefaultBuffet(defaultBuffet entity.Buffet, userBuffet entity.Buffet) entity.Buffet
	MergeWithDefaultTablet(defaultTablet entity.TabletSetting, userTablet entity.TabletSetting) entity.TabletSetting
	MergeWithDefaultPayment(defaultPayment entity.PaymentSetting, userPayment entity.PaymentSetting) entity.PaymentSetting
	MergeWithDefaultCurrency(defaultCurrency entity.CurrencySetting, userCurrency entity.CurrencySetting) entity.CurrencySetting
	MergeWithDefaultServiceCharge(defaultServiceCharge entity.ServiceCharge, userServiceCharge entity.ServiceCharge) entity.ServiceCharge
	MergeWithDefaultTaxRate(defaultTaxRate entity.TaxRate, userTaxRate entity.TaxRate) entity.TaxRate

	// 语言处理相关
	FilterValidLanguages(languages []valueobject.LanguageItem, availableLanguages []string) []valueobject.LanguageItem

	// 轮播图处理相关
	AddImageDomainToCarousel(carousel []valueobject.CarouselItem, baseURL string) []valueobject.CarouselItem

	// 坐标处理相关
	ParseCoordinates(coordinates string) (latitude, longitude string)
}

// IStoreSettingDomainService 门店设置域服务接口
type IStoreSettingDomainService interface {
	// 门店设置业务逻辑
	ValidateStoreLanguages(languages []valueobject.LanguageItem) error
	UpdateStoreBasicInfo(store *entity.StoreSetting, name, company, address, phone, storeCode string)
	UpdateStoreCoordinates(store *entity.StoreSetting, coordinates string)
}

// ICashierSettingDomainService 收银机设置域服务接口
type ICashierSettingDomainService interface {
	// 收银机设置业务逻辑
	ValidateAutoOrderSettings(isAutoOrder, autoOrderLimit string) error
	UpdateAutoOrderSettings(cashier *entity.CashierSetting, isAutoOrder, autoOrderLimit, isAutoVoice string)
	UpdateAutoMemberOrderSettings(cashier *entity.CashierSetting, isAutoMemberOrder, autoMemberOrderLimit, isAutoVoice string)
}

// IPrinterSettingDomainService 打印机设置域服务接口
type IPrinterSettingDomainService interface {
	// 打印机设置业务逻辑
	ValidatePrinterConfig(printerType string, config map[string]interface{}) error
	GeneratePrinterInfo(printerSetting entity.PrinterSetting, deviceSn string) (entity.PrinterInfo, error)
}

// IBusinessSettingDomainService 业务设置域服务接口
type IBusinessSettingDomainService interface {
	// 业务设置业务逻辑
	ValidateBatchSettings(isBatch string, batchTagNum int) error
	UpdateBatchCookingMode(business *entity.BusinessSetting, cookingMode string)
}
