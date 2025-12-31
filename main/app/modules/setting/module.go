package setting

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/setting/adapter"
	"ttpos-server-go/app/modules/setting/domain/entity"
	"ttpos-server-go/app/modules/setting/domain/service"
	"ttpos-server-go/app/modules/setting/domain/valueobject"
	"ttpos-server-go/app/modules/setting/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

// ISrv 设置服务接口（向后兼容）
type ISrv interface {
	GetStoreSetting(ctx context.Context) (entity.StoreSetting, error)
	GetStoreLanguageList(ctx context.Context) ([]valueobject.LanguageItem, error)
	GetStoreLanguage(ctx context.Context) ([]string, error)
	GetPrinterSetting(ctx context.Context, languageList []valueobject.LanguageItem) (entity.PrinterSetting, error)
	GetPrinterInfo(ctx context.Context, printerSetting entity.PrinterSetting, deviceId string) (entity.PrinterInfo, error)
	GetCashierSetting(ctx context.Context, languageList []valueobject.LanguageItem) (entity.CashierSetting, error)
	GetKioskSetting(ctx context.Context) (entity.KioskSetting, error)
	GetCloudBasicSetting(ctx context.Context) (entity.CloudBasicSetting, error)
	GetAssistantSetting(ctx context.Context, languageList []valueobject.LanguageItem) (entity.AssistantSetting, error)
	GetPointsSetting(ctx context.Context) (entity.PointsSetting, error)
	GetKitchenSetting(ctx context.Context, companySetting model.CompanySetting, languageList []valueobject.LanguageItem) (entity.KitchenSetting, error)
	GetH5Setting(ctx context.Context, languageList []valueobject.LanguageItem) (entity.H5Setting, error)
	GetBusinessSetting(ctx context.Context) (entity.BusinessSetting, error)
	GetBuffetSetting(ctx context.Context, companySetting model.CompanySetting) (entity.BuffetResp, error)
	GetTabletSetting(ctx context.Context, languageList []valueobject.LanguageItem) (entity.TabletSetting, error)
	GetCurrencySetting(ctx context.Context) (entity.CurrencySetting, error)
	GetCompanySetting(ctx context.Context) (model.CompanySetting, error)
	GetPaymentSetting(ctx context.Context, companySetting model.CompanySetting) (entity.PaymentSetting, error)
	GetCashierLanguage(c context.Context) (entity.LanguageResp, error)
	GetCashierAd(ctx context.Context) (entity.Ads, error)
	GetServiceFeeSetting(ctx context.Context) (entity.ServiceCharge, error)
	GetTaxRateSetting(ctx context.Context) (entity.TaxRate, error)
	VerifyPassword(ctx context.Context, source string, typ string, password string) bool
	UpdateSetting(ctx context.Context, settingKey string, values any) error
	VerifyAdvancedPassword(ctx context.Context, password string, options ...interface{}) error
	CheckUpdate(ctx context.Context, appType int, brand string, language string) (entity.UpdateInfo, error)
	EditAcceptOrderSetting(ctx context.Context, orderSetting entity.UpdateAcceptOrderSetting) error
	EditAcceptMemberOrderSetting(ctx context.Context, orderSetting entity.UpdateAcceptMemberOrderSetting) error
	EditSystemSetting(ctx context.Context, systemSetting entity.UpdateSystemSetting) error
	EditCashierSetting(ctx context.Context, cashierSettingReq entity.SaveCashierSettingReq) error
	EditKioskSetting(ctx context.Context, kioskSettingReq entity.SaveKioskSettingReq) error
	SaveKitchenSetting(ctx context.Context, kitchenSettingReq entity.SaveKitchenSettingReq) error
	GetCashierBaseSetting(ctx context.Context) (entity.CashierBaseSetting, error)
	GetAcceptOrderSetting(ctx context.Context) (*entity.AcceptOrderSetting, error)
	SymbolPosition(ctx context.Context, price float64) string
	EditStoreSetting(ctx context.Context, storeSetting entity.UpdateStoreSetting) error
	EditBusinessSetting(ctx context.Context, businessSetting entity.UpdateBusinessSetting) error
	GetShopBusinessSetting(ctx context.Context) (entity.ShopBusinessSetting, error)
	GetMenuQrcode(ctx context.Context) (string, error)
	GetPaymentMethodList(ctx context.Context) entity.PaymentMethodListResp
	GetDataManageSetting(ctx context.Context) entity.DataManageSetting
}

// NewSrv 创建设置服务（向后兼容）
func NewSrv(db *database.DBManager, cache cache.Cache) ISrv {
	settingRepo := persistence.NewSettingRepository(db, cache)
	domainSvc := service.NewSettingDomainService()
	return adapter.NewLegacySettingSrv(settingRepo, domainSvc)
}

// NewSrvImpl 创建设置服务实现（向后兼容）
func NewSrvImpl(db *database.DBManager, cache cache.Cache) ISrv {
	return NewSrv(db, cache)
}
