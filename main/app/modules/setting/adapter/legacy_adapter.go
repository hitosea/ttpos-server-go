package adapter

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/setting/application"
	"ttpos-server-go/app/modules/setting/domain/entity"
	"ttpos-server-go/app/modules/setting/domain/repository"
	"ttpos-server-go/app/modules/setting/domain/service"
	"ttpos-server-go/app/modules/setting/domain/valueobject"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
)

// LegacySettingSrv 遗留设置服务适配器，实现原有ISrv接口
type LegacySettingSrv struct {
	appSvc application.ISettingAppService
}

// NewLegacySettingSrv 创建遗留设置服务适配器
func NewLegacySettingSrv(repo repository.ISettingRepository, domainSvc service.ISettingDomainService) *LegacySettingSrv {
	cache := cache.NewCache(cache.GoCache, cache.Config{})
	appSvc := application.NewSettingAppServiceImpl(repo, domainSvc, cache)
	return &LegacySettingSrv{
		appSvc: appSvc,
	}
}

// GetStoreSetting 获取门店设置
func (l *LegacySettingSrv) GetStoreSetting(ctx context.Context) (entity.StoreSetting, error) {
	return l.appSvc.GetStoreSetting(ctx)
}

// GetStoreLanguageList 获取门店语言列表
func (l *LegacySettingSrv) GetStoreLanguageList(ctx context.Context) ([]valueobject.LanguageItem, error) {
	return l.appSvc.GetStoreLanguageList(ctx)
}

// GetStoreLanguage 获取门店语言
func (l *LegacySettingSrv) GetStoreLanguage(ctx context.Context) ([]string, error) {
	return l.appSvc.GetStoreLanguage(ctx)
}

// GetCashierSetting 获取收银机设置
func (l *LegacySettingSrv) GetCashierSetting(ctx context.Context, languageList []valueobject.LanguageItem) (entity.CashierSetting, error) {
	return l.appSvc.GetCashierSetting(ctx, languageList)
}

// GetCashierLanguage 获取收银机语言
func (l *LegacySettingSrv) GetCashierLanguage(ctx context.Context) (entity.LanguageResp, error) {
	return l.appSvc.GetCashierLanguage(ctx)
}

// GetCashierAd 获取收银机副屏广告
func (l *LegacySettingSrv) GetCashierAd(ctx context.Context) (entity.Ads, error) {
	return l.appSvc.GetCashierAd(ctx)
}

// GetCashierBaseSetting 获取收银端设置
func (l *LegacySettingSrv) GetCashierBaseSetting(ctx context.Context) (entity.CashierBaseSetting, error) {
	return l.appSvc.GetCashierBaseSetting(ctx)
}

// GetPrinterSetting 获取打印机设置
func (l *LegacySettingSrv) GetPrinterSetting(ctx context.Context, languageList []valueobject.LanguageItem) (entity.PrinterSetting, error) {
	return l.appSvc.GetPrinterSetting(ctx, languageList)
}

// GetPrinterInfo 获取打印机信息
func (l *LegacySettingSrv) GetPrinterInfo(ctx context.Context, printerSetting entity.PrinterSetting, deviceSn string) (entity.PrinterInfo, error) {
	return l.appSvc.GetPrinterInfo(ctx, printerSetting, deviceSn)
}

// GetBusinessSetting 获取门店业务设置
func (l *LegacySettingSrv) GetBusinessSetting(ctx context.Context) (entity.BusinessSetting, error) {
	return l.appSvc.GetBusinessSetting(ctx)
}

// GetShopBusinessSetting 获取商家业务设置
func (l *LegacySettingSrv) GetShopBusinessSetting(ctx context.Context) (entity.ShopBusinessSetting, error) {
	return l.appSvc.GetShopBusinessSetting(ctx)
}

// GetBuffetSetting 获取自助餐设置
func (l *LegacySettingSrv) GetBuffetSetting(ctx context.Context, companySetting model.CompanySetting) (entity.BuffetResp, error) {
	return l.appSvc.GetBuffetSetting(ctx, companySetting)
}

// GetTabletSetting 获取平板端设置
func (l *LegacySettingSrv) GetTabletSetting(ctx context.Context, languageList []valueobject.LanguageItem) (entity.TabletSetting, error) {
	return l.appSvc.GetTabletSetting(ctx, languageList)
}

// GetPaymentSetting 获取门店-支付方式设置
func (l *LegacySettingSrv) GetPaymentSetting(ctx context.Context, companySetting model.CompanySetting) (entity.PaymentSetting, error) {
	return l.appSvc.GetPaymentSetting(ctx, companySetting)
}

// GetCurrencySetting 获取货币单位设置
func (l *LegacySettingSrv) GetCurrencySetting(ctx context.Context) (entity.CurrencySetting, error) {
	return l.appSvc.GetCurrencySetting(ctx)
}

// GetCompanySetting 获取公司设置
func (l *LegacySettingSrv) GetCompanySetting(ctx context.Context) (model.CompanySetting, error) {
	return l.appSvc.GetCompanySetting(ctx)
}

// GetKitchenSetting 获取厨显端设置
func (l *LegacySettingSrv) GetKitchenSetting(ctx context.Context, companySetting model.CompanySetting, languageList []valueobject.LanguageItem) (entity.KitchenSetting, error) {
	return l.appSvc.GetKitchenSetting(ctx, companySetting, languageList)
}

// GetH5Setting 获取扫码H5设置
func (l *LegacySettingSrv) GetH5Setting(ctx context.Context, languageList []valueobject.LanguageItem) (entity.H5Setting, error) {
	return l.appSvc.GetH5Setting(ctx, languageList)
}

// GetKioskSetting 获取自助点餐机设置
func (l *LegacySettingSrv) GetKioskSetting(ctx context.Context) (entity.KioskSetting, error) {
	return l.appSvc.GetKioskSetting(ctx)
}

// GetAssistantSetting 获取点餐助手设置
func (l *LegacySettingSrv) GetAssistantSetting(ctx context.Context, languageList []valueobject.LanguageItem) (entity.AssistantSetting, error) {
	return l.appSvc.GetAssistantSetting(ctx, languageList)
}

// GetPointsSetting 获取积分设置
func (l *LegacySettingSrv) GetPointsSetting(ctx context.Context) (entity.PointsSetting, error) {
	return l.appSvc.GetPointsSetting(ctx)
}

// GetCloudBasicSetting 获取云端基础信息
func (l *LegacySettingSrv) GetCloudBasicSetting(ctx context.Context) (entity.CloudBasicSetting, error) {
	return l.appSvc.GetCloudBasicSetting(ctx)
}

// GetServiceFeeSetting 获取服务费设置
func (l *LegacySettingSrv) GetServiceFeeSetting(ctx context.Context) (entity.ServiceCharge, error) {
	return l.appSvc.GetServiceFeeSetting(ctx)
}

// GetTaxRateSetting 获取税率设置
func (l *LegacySettingSrv) GetTaxRateSetting(ctx context.Context) (entity.TaxRate, error) {
	return l.appSvc.GetTaxRateSetting(ctx)
}

// VerifyPassword 收银机验证密码
func (l *LegacySettingSrv) VerifyPassword(ctx context.Context, source, typ, password string) bool {
	return l.appSvc.VerifyPassword(ctx, source, typ, password)
}

// UpdateSetting 更新设置
func (l *LegacySettingSrv) UpdateSetting(ctx context.Context, settingKey string, values interface{}) error {
	return l.appSvc.UpdateSetting(ctx, settingKey, values)
}

// EditStoreSetting 修改店铺设置
func (l *LegacySettingSrv) EditStoreSetting(ctx context.Context, storeSetting entity.UpdateStoreSetting) error {
	return l.appSvc.EditStoreSetting(ctx, storeSetting)
}

// EditBusinessSetting 修改业务设置
func (l *LegacySettingSrv) EditBusinessSetting(ctx context.Context, businessSetting entity.UpdateBusinessSetting) error {
	return l.appSvc.EditBusinessSetting(ctx, businessSetting)
}

// EditCashierSetting 修改收银机设置
func (l *LegacySettingSrv) EditCashierSetting(ctx context.Context, cashierSettingReq entity.SaveCashierSettingReq) error {
	return l.appSvc.EditCashierSetting(ctx, cashierSettingReq)
}

// EditKioskSetting 修改自助点餐机设置
func (l *LegacySettingSrv) EditKioskSetting(ctx context.Context, kioskSettingReq entity.SaveKioskSettingReq) error {
	return l.appSvc.EditKioskSetting(ctx, kioskSettingReq)
}

// SaveKitchenSetting 保存厨显设置
func (l *LegacySettingSrv) SaveKitchenSetting(ctx context.Context, kitchenSettingReq entity.SaveKitchenSettingReq) error {
	return l.appSvc.SaveKitchenSetting(ctx, kitchenSettingReq)
}

// EditSystemSetting 修改系统设置
func (l *LegacySettingSrv) EditSystemSetting(ctx context.Context, systemSetting entity.UpdateSystemSetting) error {
	return l.appSvc.EditSystemSetting(ctx, systemSetting)
}

// GetAcceptOrderSetting 获取接单设置
func (l *LegacySettingSrv) GetAcceptOrderSetting(ctx context.Context) (*entity.AcceptOrderSetting, error) {
	return l.appSvc.GetAcceptOrderSetting(ctx)
}

// EditAcceptOrderSetting 修改自动接单设置
func (l *LegacySettingSrv) EditAcceptOrderSetting(ctx context.Context, orderSetting entity.UpdateAcceptOrderSetting) error {
	return l.appSvc.EditAcceptOrderSetting(ctx, orderSetting)
}

// EditAcceptMemberOrderSetting 修改自动接单会员订单设置
func (l *LegacySettingSrv) EditAcceptMemberOrderSetting(ctx context.Context, orderSetting entity.UpdateAcceptMemberOrderSetting) error {
	return l.appSvc.EditAcceptMemberOrderSetting(ctx, orderSetting)
}

// GetMenuQrcode 获取电子菜单二维码
func (l *LegacySettingSrv) GetMenuQrcode(ctx context.Context) (string, error) {
	return l.appSvc.GetMenuQrcode(ctx)
}

// GetPaymentMethodList 获取支付方式列表
func (l *LegacySettingSrv) GetPaymentMethodList(ctx context.Context) entity.PaymentMethodListResp {
	return l.appSvc.GetPaymentMethodList(ctx)
}

// GetDataManageSetting 获取数据管理设置
func (l *LegacySettingSrv) GetDataManageSetting(ctx context.Context) entity.DataManageSetting {
	return l.appSvc.GetDataManageSetting(ctx)
}

// CheckUpdate 检查更新
func (l *LegacySettingSrv) CheckUpdate(ctx context.Context, appType int, brand, language string) (entity.UpdateInfo, error) {
	return l.appSvc.CheckUpdate(ctx, appType, brand, language)
}

// SymbolPosition 根据货币符号位置返回字符串
func (l *LegacySettingSrv) SymbolPosition(ctx context.Context, price float64) string {
	return l.appSvc.SymbolPosition(ctx, price)
}

// VerifyAdvancedPassword 验证高级密码
func (l *LegacySettingSrv) VerifyAdvancedPassword(ctx context.Context, password string, options ...interface{}) error {
	return l.appSvc.VerifyAdvancedPassword(ctx, password, options...)
}
