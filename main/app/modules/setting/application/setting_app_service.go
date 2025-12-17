package application

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/setting/domain/entity"
	"ttpos-server-go/app/modules/setting/domain/valueobject"
	"ttpos-server-go/pkg/context"
)

// ISettingAppService 设置应用服务接口
type ISettingAppService interface {
	// 门店设置相关
	GetStoreSetting(ctx context.Context) (entity.StoreSetting, error)
	GetStoreLanguageList(ctx context.Context) ([]valueobject.LanguageItem, error)
	GetStoreLanguage(ctx context.Context) ([]string, error)

	// 收银机设置相关
	GetCashierSetting(ctx context.Context, languageList []valueobject.LanguageItem) (entity.CashierSetting, error)
	GetCashierLanguage(ctx context.Context) (entity.LanguageResp, error)
	GetCashierAd(ctx context.Context) (entity.Ads, error)
	GetCashierBaseSetting(ctx context.Context) (entity.CashierBaseSetting, error)

	// 打印机设置相关
	GetPrinterSetting(ctx context.Context, languageList []valueobject.LanguageItem) (entity.PrinterSetting, error)
	GetPrinterInfo(ctx context.Context, printerSetting entity.PrinterSetting, deviceSn string) (entity.PrinterInfo, error)

	// 业务设置相关
	GetBusinessSetting(ctx context.Context) (entity.BusinessSetting, error)
	GetShopBusinessSetting(ctx context.Context) (entity.ShopBusinessSetting, error)

	// 自助餐设置相关
	GetBuffetSetting(ctx context.Context, companySetting model.CompanySetting) (entity.BuffetResp, error)

	// 平板设置相关
	GetTabletSetting(ctx context.Context, languageList []valueobject.LanguageItem) (entity.TabletSetting, error)

	// 支付设置相关
	GetPaymentSetting(ctx context.Context, companySetting model.CompanySetting) (entity.PaymentSetting, error)

	// 货币设置相关
	GetCurrencySetting(ctx context.Context) (entity.CurrencySetting, error)

	// 公司设置相关
	GetCompanySetting(ctx context.Context) (model.CompanySetting, error)

	// 厨显设置相关
	GetKitchenSetting(ctx context.Context, companySetting model.CompanySetting, languageList []valueobject.LanguageItem) (entity.KitchenSetting, error)

	// H5设置相关
	GetH5Setting(ctx context.Context, languageList []valueobject.LanguageItem) (entity.H5Setting, error)

	// 自助点餐机设置相关
	GetKioskSetting(ctx context.Context) (entity.KioskSetting, error)

	// 点餐助手设置相关
	GetAssistantSetting(ctx context.Context, languageList []valueobject.LanguageItem) (entity.AssistantSetting, error)

	// 积分设置相关
	GetPointsSetting(ctx context.Context) (entity.PointsSetting, error)

	// 云端基础信息相关
	GetCloudBasicSetting(ctx context.Context) (entity.CloudBasicSetting, error)

	// 服务费设置相关
	GetServiceFeeSetting(ctx context.Context) (entity.ServiceCharge, error)

	// 税率设置相关
	GetTaxRateSetting(ctx context.Context) (entity.TaxRate, error)

	// 密码验证相关
	VerifyPassword(ctx context.Context, source, typ, password string) bool
	VerifyAdvancedPassword(ctx context.Context, password string, options ...interface{}) error

	// 更新设置相关
	UpdateSetting(ctx context.Context, settingKey string, values interface{}) error

	// 编辑设置相关
	EditStoreSetting(ctx context.Context, storeSetting entity.UpdateStoreSetting) error
	EditBusinessSetting(ctx context.Context, businessSetting entity.UpdateBusinessSetting) error
	EditCashierSetting(ctx context.Context, cashierSettingReq entity.SaveCashierSettingReq) error
	EditKioskSetting(ctx context.Context, kioskSettingReq entity.SaveKioskSettingReq) error
	SaveKitchenSetting(ctx context.Context, kitchenSettingReq entity.SaveKitchenSettingReq) error
	EditSystemSetting(ctx context.Context, systemSetting entity.UpdateSystemSetting) error

	// 接单设置相关
	GetAcceptOrderSetting(ctx context.Context) (*entity.AcceptOrderSetting, error)
	EditAcceptOrderSetting(ctx context.Context, orderSetting entity.UpdateAcceptOrderSetting) error
	EditAcceptMemberOrderSetting(ctx context.Context, orderSetting entity.UpdateAcceptMemberOrderSetting) error

	// 菜单相关
	GetMenuQrcode(ctx context.Context) (string, error)

	// 支付方式相关
	GetPaymentMethodList(ctx context.Context) entity.PaymentMethodListResp

	// 数据管理相关
	GetDataManageSetting(ctx context.Context) entity.DataManageSetting

	// 更新检查相关
	CheckUpdate(ctx context.Context, appType int, brand, language string) (entity.UpdateInfo, error)

	// 货币符号相关
	SymbolPosition(ctx context.Context, price float64) string
}

// 响应实体定义（使用entity包中的类型以避免重复）
