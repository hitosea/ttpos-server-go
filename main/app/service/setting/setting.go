package setting

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/dto/resp/setting"
	errors2 "ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"github.com/nahid/gohttp"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type GetSettingReq struct {
	CompanyUuid  uint64             // 商家Uuid
	Language     string             // 请求头accept-language，用于实时翻译，比如门店业务获取抹零翻译"实款实收"
	Context      *gin.Context       // 请求上下文，用于给url加上域名，比如收银机副屏广告轮播图增加域名
	LanguageList []dto.LanguageItem // 如果设置为nil，则读取商家语言，否则将LanguageList传递给getDefaultXXX
}

type ISrv interface {
	GetStoreSetting(ctx context.Context) (setting.Store, error)                                                                           // 获取商家设置
	GetStoreLanguageList(ctx context.Context) ([]dto.LanguageItem, error)                                                                 // 获取商家语言
	GetPrinterSetting(ctx context.Context, languageList []dto.LanguageItem) (setting.Printer, error)                                      // 获取打印机设置
	GetPrinterInfo(ctx context.Context, printerSetting setting.Printer, deviceId string) (setting.PrinterInfo, error)                     // 获取打印机信息
	GetCashierSetting(ctx context.Context, languageList []dto.LanguageItem) (setting.Cashier, error)                                      // 获取收银机设置
	GetAssistantSetting(ctx context.Context, languageList []dto.LanguageItem) (setting.Assistant, error)                                  // 获取点餐助手设置
	GetKitchenSetting(ctx context.Context, companySetting model.CompanySetting, languageList []dto.LanguageItem) (setting.Kitchen, error) // 获取厨显端设置
	GetH5Setting(ctx context.Context, languageList []dto.LanguageItem) (setting.H5, error)                                                // 获取扫码H5设置
	GetBusinessSetting(ctx context.Context) (setting.Business, error)                                                                     // 获取门店业务设置
	GetBuffetSetting(ctx context.Context, companySetting model.CompanySetting) (setting.Buffet, error)                                    // 获取自助餐设置
	GetTabletSetting(ctx context.Context, languageList []dto.LanguageItem) (setting.Tablet, error)                                        // 获取平板端设置
	GetCurrencySetting(ctx context.Context) (setting.Currency, error)                                                                     // 获取货币单位设置
	GetCompanySetting(ctx context.Context) (model.CompanySetting, error)                                                                  // 获取公司设置
	GetPaymentSetting(ctx context.Context, companySetting model.CompanySetting) (setting.Payment, error)                                  // 获取门店-支付方式设置
	GetCashierLanguage(c context.Context) (resp.LanguageResp, error)                                                                      // 获取收银机语言
	GetCashierAd(ctx context.Context) (resp.Ads, error)                                                                                   // 获取收银机副屏广告
	GetServiceFeeSetting(ctx context.Context) (setting.ServiceCharge, error)                                                              // 获取服务费设置
	GetTaxRateSetting(ctx context.Context) (setting.TaxRate, error)                                                                       // 获取税率设置
	CashierVerifyPassword(ctx context.Context, typ string, password string, companyUuid uint64) bool                                      // 收银机验证密码
	UpdateSetting(ctx context.Context, settingKey string, values any) error                                                               // 更新设置
	VerifyAdvancedPassword(ctx context.Context, password string) error                                                                    // 验证高级密码
	CheckUpdate(ctx context.Context, appType int, brand string, language string) (resp.UpdateInfo, error)                                 // 检查更新
	EditAcceptOrderSetting(ctx context.Context, orderSetting req.UpdateAcceptOrderSetting) error                                          // 修改自动接单设置
	EditSystemSetting(ctx context.Context, systemSetting req.UpdateSystemSetting) error                                                   // 修改系统设置
	GetCashierBaseSetting(ctx context.Context) (resp.CashierBaseSetting, error)                                                           // 获取收银端设置
}

func NewSrv(dbm *database.DBManager, cache cache.Cache) ISrv {
	return NewSrvImpl(dbm, cache)
}

type Srv struct {
	dbm      *database.DBManager
	cache    cache.Cache
	cacheKey string
}

func NewSrvImpl(dbm *database.DBManager, cache cache.Cache) *Srv {
	return &Srv{
		dbm:      dbm,
		cache:    cache,
		cacheKey: "setting:company_id:%d",
	}
}

// 从缓存读取，没有则生成缓存
func (s *Srv) fromCache(ctx context.Context) ([]model.Setting, error) {
	companyUuid := ctx.GetCompanyUuid()
	var settings []model.Setting
	cacheKey := fmt.Sprintf(s.cacheKey, companyUuid)
	if data, exists := s.cache.Get(cacheKey); exists {
		if dataValue, isString := data.(string); isString {
			if err := json.Unmarshal([]byte(dataValue), &settings); err != nil {
				return settings, err
			}
			return settings, nil
		}
	}
	// 从数据库读取
	var err error
	settingRepo := repository.NewSettingRepo(s.dbm.GetDB(companyUuid))
	settings, err = settingRepo.GetAll()
	if err != nil {
		ctx.Log().Error("从数据库获取设置失败", zap.Error(err))
		return nil, errors.New("获取设置失败")
	}

	data, _ := json.Marshal(settings)
	s.cache.Set(cacheKey, string(data), 0)

	return settings, nil
}

// GetAll 获取所有设置
func (s *Srv) getAll(ctx context.Context, language string, languageList []dto.LanguageItem) (map[string]any, error) {
	// 获取company_setting
	companySettingRepo := repository.NewCompanySettingRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	companySetting := companySettingRepo.Get()

	var retSettings = make(map[string]any)
	// 从缓存读取
	settings, err := s.fromCache(ctx)

	var keys []string
	for _, m := range settings {
		keys = append(keys, m.Key)
	}
	for _, key := range []string{constant.SettingPrinter, constant.SettingStore, constant.SettingRecharge,
		constant.SettingPoints, constant.SettingSysAdminConfig, constant.SettingSysConfig, constant.SettingBalance,
		constant.SettingCurrency, constant.SettingTaxRate, constant.SettingServiceCharge, constant.SettingPayment,
		constant.SettingBusiness, constant.SettingCashier, constant.SettingTablet, constant.SettingH5, constant.SettingKitchen,
		constant.SettingAssistant, constant.SettingBuffet} {
		if !slices.Contains(keys, key) {
			settings = append(settings, model.Setting{
				Key:    key,
				Values: "{}",
			})
		}
	}

	var isShowScanSoldOut, isShowAssistantSoldOut int

	for _, st := range settings {
		ginContext := ctx.GetGinContext()
		switch st.Key {
		case constant.SettingPrinter: // 小票打印机设置
			var printer setting.Printer
			err = json.Unmarshal([]byte(st.Values), &printer)
			if err != nil {
				ctx.Log().Error("解析小票打印机设置失败", zap.Error(err))
				return nil, errors.New("解析小票打印机设置失败")
			}
			// 过滤佛历、过滤打印方式，使用默认
			printer.CalendarList = nil
			printer.PrintList = nil
			tmp := s.getDefaultPrinter(language, languageList)
			err = copier.CopyWithOption(&tmp, printer, copier.Option{IgnoreEmpty: true})
			if err != nil {
				ctx.Log().Error("合并小票打印机设置失败", zap.Error(err))
				return nil, errors.New("合并小票打印机设置失败")
			}
			retSettings[constant.SettingPrinter] = tmp
		case constant.SettingStore: // 商城设置
			var store setting.Store
			err = json.Unmarshal([]byte(st.Values), &store)
			if err != nil {
				ctx.Log().Error("解析商城设置失败", zap.Error(err))
				return nil, errors.New("解析商城设置失败")
			}
			if store.IPWhiteList != "" {
				store.IPWhiteList = viper.GetString("PAY_SERVICE_IP")
			}

			tmp := s.getDefaultStore(language)
			err = copier.CopyWithOption(&tmp, store, copier.Option{IgnoreEmpty: true})
			if err != nil {
				ctx.Log().Error("合并商城设置失败", zap.Error(err))
				return nil, errors.New("合并商城设置失败")
			}
			if tmp.LogoURL != "" && ginContext != nil {
				tmp.LogoURL = utils.GetBaseURL(ginContext.Request) + tmp.LogoURL
			}
			if tmp.AvatarURL != "" && ginContext != nil {
				tmp.AvatarURL = utils.GetBaseURL(ginContext.Request) + tmp.AvatarURL
			}

			retSettings[constant.SettingStore] = tmp
		case constant.SettingRecharge: // 充值设置
			var recharge setting.Recharge
			err = json.Unmarshal([]byte(st.Values), &recharge)
			if err != nil {
				ctx.Log().Error("解析充值设置失败", zap.Error(err))
				return nil, errors.New("解析充值设置失败")
			}
			tmp := s.getDefaultRecharge()
			err = copier.CopyWithOption(&tmp, recharge, copier.Option{IgnoreEmpty: true})
			if err != nil {
				ctx.Log().Error("合并充值设置失败", zap.Error(err))
				return nil, errors.New("合并充值设置失败")
			}
			retSettings[constant.SettingRecharge] = tmp
		case constant.SettingPoints: // 积分设置
			var points setting.Points
			err = json.Unmarshal([]byte(st.Values), &points)
			if err != nil {
				ctx.Log().Error("解析积分设置失败", zap.Error(err))
				return nil, errors.New("解析积分设置失败")
			}
			tmp := s.getDefaultPoints()
			err = copier.CopyWithOption(&tmp, points, copier.Option{IgnoreEmpty: true})
			if err != nil {
				ctx.Log().Error("合并积分设置失败", zap.Error(err))
				return nil, errors.New("合并积分设置失败")
			}
			retSettings[constant.SettingPoints] = tmp
		case constant.SettingSysAdminConfig: // 系统配置
			var sysAdminConfig setting.SysAdminConfig
			err = json.Unmarshal([]byte(st.Values), &sysAdminConfig)
			if err != nil {
				ctx.Log().Error("解析系统配置失败", zap.Error(err))
				return nil, errors.New("解析系统配置失败")
			}
			tmp := s.getDefaultSysAdminConfig()
			err = copier.CopyWithOption(&tmp, sysAdminConfig, copier.Option{IgnoreEmpty: true})
			if err != nil {
				ctx.Log().Error("合并系统配置失败", zap.Error(err))
				return nil, errors.New("合并系统配置失败")
			}
			retSettings[constant.SettingSysAdminConfig] = tmp
		case constant.SettingSysConfig: // 系统配置
			var sysConfig setting.SysConfig
			err = json.Unmarshal([]byte(st.Values), &sysConfig)
			if err != nil {
				ctx.Log().Error("解析系统配置失败", zap.Error(err))
				return nil, errors.New("解析系统配置失败")
			}
			tmp := s.getDefaultSysConfig()
			err = copier.CopyWithOption(&tmp, sysConfig, copier.Option{IgnoreEmpty: true})
			if err != nil {
				ctx.Log().Error("合并系统配置失败", zap.Error(err))
				return nil, errors.New("合并系统配置失败")
			}
			retSettings[constant.SettingSysConfig] = tmp
		case constant.SettingBalance: // 充值设置
			var balance setting.Balance
			err = json.Unmarshal([]byte(st.Values), &balance)
			if err != nil {
				ctx.Log().Error("解析充值设置失败", zap.Error(err))
				return nil, errors.New("解析充值设置失败")
			}
			tmp := s.getDefaultBalance()
			err = copier.CopyWithOption(&tmp, balance, copier.Option{IgnoreEmpty: true})
			if err != nil {
				ctx.Log().Error("合并充值设置失败", zap.Error(err))
				return nil, errors.New("合并充值设置失败")
			}
			retSettings[constant.SettingBalance] = tmp
		case constant.SettingCurrency: // 门店-货币单位
			var currency setting.Currency
			err = json.Unmarshal([]byte(st.Values), &currency)
			if err != nil {
				ctx.Log().Error("解析门店-货币单位失败", zap.Error(err))
				return nil, errors.New("解析门店-货币单位失败")
			}
			tmp := s.getDefaultCurrency()
			err = copier.CopyWithOption(&tmp, currency, copier.Option{IgnoreEmpty: true})
			if err != nil {
				ctx.Log().Error("合并门店-货币单位失败", zap.Error(err))
				return nil, errors.New("合并门店-货币单位失败")
			}
			retSettings[constant.SettingCurrency] = tmp
		case constant.SettingTaxRate: // 门店-税率管理
			var taxRate setting.TaxRate
			err = json.Unmarshal([]byte(st.Values), &taxRate)
			if err != nil {
				ctx.Log().Error("解析门店-税率管理失败", zap.Error(err))
				return nil, errors.New("解析门店-税率管理失败")
			}
			tmp := s.getDefaultTaxRate()
			err = copier.CopyWithOption(&tmp, taxRate, copier.Option{IgnoreEmpty: true})
			if err != nil {
				ctx.Log().Error("合并门店-税率管理失败", zap.Error(err))
				return nil, errors.New("合并门店-税率管理失败")
			}
			retSettings[constant.SettingTaxRate] = tmp
		case constant.SettingServiceCharge: // 门店-服务费
			var serviceCharge setting.ServiceCharge
			err = json.Unmarshal([]byte(st.Values), &serviceCharge)
			if err != nil {
				ctx.Log().Error("解析门店-服务费失败", zap.Error(err))
				return nil, errors.New("解析门店-服务费失败")
			}
			tmp := s.getDefaultServiceCharge()
			err = copier.CopyWithOption(&tmp, serviceCharge, copier.Option{IgnoreEmpty: true})
			if err != nil {
				ctx.Log().Error("合并门店-服务费失败", zap.Error(err))
				return nil, errors.New("合并门店-服务费失败")
			}
			retSettings[constant.SettingServiceCharge] = tmp
		case constant.SettingPayment: // 门店-支付方式
			var payment setting.Payment
			err = json.Unmarshal([]byte(st.Values), &payment)
			if err != nil {
				ctx.Log().Error("解析门店-支付方式失败", zap.Error(err))
				return nil, errors.New("解析门店-支付方式失败")
			}

			// 会员关闭时 门店管理 支付方式 余额这个开关要关了
			if companySetting.IsOpenMember == 0 {
				payment.IsBalance = "0"
			}

			tmp := s.getDefaultPayment()
			err = copier.CopyWithOption(&tmp, payment, copier.Option{IgnoreEmpty: true})
			if err != nil {
				ctx.Log().Error("合并门店-支付方式失败", zap.Error(err))
				return nil, errors.New("合并门店-支付方式失败")
			}

			retSettings[constant.SettingPayment] = tmp
		case constant.SettingBusiness: // 门店-业务设置
			var business setting.Business
			err = json.Unmarshal([]byte(st.Values), &business)
			if err != nil {
				ctx.Log().Error("解析门店-业务设置失败", zap.Error(err))
				return nil, errors.New("解析门店-业务设置失败")
			}
			// 门店业务-过滤列表，使用默认
			business.ZeroingMethodList = nil
			business.CheckoutZeroingMethodList = nil
			business.GiftMethodList = nil
			business.FreeMethodList = nil
			tmp := s.getDefaultBusiness(language)
			err = copier.CopyWithOption(&tmp, business, copier.Option{IgnoreEmpty: true})
			if err != nil {
				ctx.Log().Error("合并门店-业务设置失败", zap.Error(err))
				return nil, errors.New("合并门店-业务设置失败")
			}
			retSettings[constant.SettingBusiness] = tmp
		case constant.SettingCashier: // 各端-收银机设置
			var cashier setting.Cashier
			err = json.Unmarshal([]byte(st.Values), &cashier)
			if err != nil {
				ctx.Log().Error("解析各端-收银机设置失败", zap.Error(err))
				return nil, errors.New("解析各端-收银机设置失败")
			}

			// 滚动图/视频处理
			if len(cashier.Carousel) > 0 && ginContext != nil {
				for i, item := range cashier.Carousel {
					cashier.Carousel[i].FilePath = utils.AddImageDomain(item.FilePath, utils.GetBaseURL(ginContext.Request), true)
				}
			}
			tmp := s.getDefaultCashier(languageList)
			// 如果设置了 is_show_scan_sold_out，则读取解析后的数据，否则读取默认设置
			if strings.Contains(st.Values, "\"is_show_scan_sold_out\"") {
				isShowScanSoldOut = cashier.IsShowScanSoldOut
			} else {
				isShowScanSoldOut = tmp.IsShowScanSoldOut
			}
			// 如果设置了 is_show_assistant_sold_out，则读取解析后的数据，否则读取默认设置
			if strings.Contains(st.Values, "\"is_show_assistant_sold_out\"") {
				isShowAssistantSoldOut = cashier.IsShowAssistantSoldOut
			} else {
				isShowAssistantSoldOut = tmp.IsShowAssistantSoldOut
			}
			// 接单语音，设备本地处理，不需要合并
			cashier.IsAutoVoice = ""
			// 语言 不需要合并
			tmp.Language = nil

			err = copier.CopyWithOption(&tmp, cashier, copier.Option{IgnoreEmpty: true})
			if err != nil {
				ctx.Log().Error("合并各端-收银机设置失败", zap.Error(err))
				return nil, errors.New("合并各端-收银机设置失败")
			}
			retSettings[constant.SettingCashier] = tmp
		case constant.SettingTablet: // 各端-平板端设置
			var tablet setting.Tablet
			err = json.Unmarshal([]byte(st.Values), &tablet)
			if err != nil {
				ctx.Log().Error("解析各端-平板端设置失败", zap.Error(err))
				return nil, errors.New("解析各端-平板端设置失败")
			}
			// 滚动图/视频处理
			if len(tablet.Carousel) > 0 && ginContext != nil {
				for i, item := range tablet.Carousel {
					tablet.Carousel[i].FilePath = utils.AddImageDomain(item.FilePath, utils.GetBaseURL(ginContext.Request), true)
				}
			}
			tmp := s.getDefaultTablet(languageList)
			// 语言 不需要合并
			tmp.Language = nil
			err = copier.CopyWithOption(&tmp, tablet, copier.Option{IgnoreEmpty: true})
			if err != nil {
				ctx.Log().Error("合并各端-平板端设置失败", zap.Error(err))
				return nil, errors.New("合并各端-平板端设置失败")
			}
			retSettings[constant.SettingTablet] = tmp
		case constant.SettingH5: // 各端-扫码H5设置
			var h5 setting.H5
			err = json.Unmarshal([]byte(st.Values), &h5)
			if err != nil {
				ctx.Log().Error("解析各端-扫码H5设置失败", zap.Error(err))
				return nil, errors.New("解析各端-扫码H5设置失败")
			}
			tmp := s.getDefaultH5(languageList)
			err = copier.CopyWithOption(&tmp, h5, copier.Option{IgnoreEmpty: true})
			if err != nil {
				ctx.Log().Error("合并各端-扫码H5设置失败", zap.Error(err))
				return nil, errors.New("合并各端-扫码H5设置失败")
			}
			retSettings[constant.SettingH5] = tmp
		case constant.SettingKitchen: // 各端-厨显设置
			var kitchen setting.Kitchen
			err = json.Unmarshal([]byte(st.Values), &kitchen)
			if err != nil {
				ctx.Log().Error("解析各端-厨显设置失败", zap.Error(err))
				return nil, errors.New("解析各端-厨显设置失败")
			}

			tmp := s.getDefaultKitchen(languageList)

			// 语言 不需要合并
			tmp.Language = nil

			err = copier.CopyWithOption(&tmp, kitchen, copier.Option{IgnoreEmpty: true})
			if err != nil {
				ctx.Log().Error("合并各端-厨显设置失败", zap.Error(err))
				return nil, errors.New("合并各端-厨显设置失败")
			}

			// 总权限 - 不开启厨显
			if companySetting.IsOpenKitchenKds == 0 {
				kitchen.IsOpen = "0"
			}

			retSettings[constant.SettingKitchen] = tmp
		case constant.SettingAssistant: // 各端-点餐助手设置
			var assistant setting.Assistant
			err = json.Unmarshal([]byte(st.Values), &assistant)
			if err != nil {
				ctx.Log().Error("解析各端-点餐助手设置失败", zap.Error(err))
				return nil, errors.New("解析各端-点餐助手设置失败")
			}
			if len(assistant.LanguageList) == 0 {
				assistant.LanguageList = nil
			}
			tmp := s.getDefaultAssistant(language, languageList)
			err = copier.CopyWithOption(&tmp, assistant, copier.Option{IgnoreEmpty: true})
			if err != nil {
				ctx.Log().Error("合并各端-点餐助手设置失败", zap.Error(err))
				return nil, errors.New("合并各端-点餐助手设置失败")
			}

			retSettings[constant.SettingAssistant] = tmp
		case constant.SettingBuffet: // 自助餐-自助餐设置
			var buffet setting.Buffet
			err = json.Unmarshal([]byte(st.Values), &buffet)
			if err != nil {
				return nil, errors.New("解析自助餐-自助餐设置失败")
			}

			if companySetting.IsOpenBuffet == 0 {
				buffet.IsOpen = "0"
			}

			tmp := s.getDefaultBuffet()
			err = copier.CopyWithOption(&tmp, buffet, copier.Option{IgnoreEmpty: true})
			if err != nil {
				return nil, errors.New("解析自助餐-自助餐设置失败")
			}
			retSettings[constant.SettingBuffet] = tmp
		}

		// 收银机设置是否显示售罄商品
		if h5Setting, ok := retSettings[constant.SettingH5]; ok {
			if h5, yes := h5Setting.(setting.H5); yes {
				h5.IsShowScanSoldOut = isShowScanSoldOut
				retSettings[constant.SettingH5] = h5
			}
		}
		if assistantSetting, ok := retSettings[constant.SettingAssistant]; ok {
			if assistant, yes := assistantSetting.(setting.Assistant); yes {
				assistant.IsShowAssistantSoldOut = isShowAssistantSoldOut
				retSettings[constant.SettingAssistant] = assistant
			}
		}
	}

	return retSettings, nil
}

// GetStoreLanguageList 获取商家语言列表
func (s *Srv) GetStoreLanguageList(ctx context.Context) ([]dto.LanguageItem, error) {
	set, err := s.GetStoreSetting(ctx)
	if err != nil {
		return nil, err
	}
	return set.Language, nil
}

func (s *Srv) getSettingByKey(ctx context.Context, key string) model.Setting {
	allSettings, _ := s.fromCache(ctx)
	for _, set := range allSettings {
		if set.Key == key {
			return set
		}
	}
	return model.Setting{
		Key:    key,
		Values: "{}",
	}
}

// GetStoreSetting 获取商家设置
func (s *Srv) GetStoreSetting(ctx context.Context) (setting.Store, error) {
	var store setting.Store
	st := s.getSettingByKey(ctx, constant.SettingStore)
	ctx.Log().Info("", zap.String("st.Values", st.Values))
	err := json.Unmarshal([]byte(st.Values), &store)
	if err != nil {
		ctx.Log().Error("解析商城设置失败", zap.Error(err))
		return store, errors.New("解析商城设置失败")
	}
	if store.IPWhiteList != "" {
		store.IPWhiteList = viper.GetString("PAY_SERVICE_IP")
	}

	defaultStore := s.getDefaultStore(ctx.GetLanguage())
	err = copier.CopyWithOption(&defaultStore, store, copier.Option{IgnoreEmpty: true})
	if err != nil {
		ctx.Log().Error("合并商城设置失败", zap.Error(err))
		return store, errors.New("合并商城设置失败")
	}
	ginContext := ctx.GetGinContext()
	if defaultStore.LogoURL != "" && ginContext != nil {
		defaultStore.LogoURL = utils.GetBaseURL(ginContext.Request) + defaultStore.LogoURL
	}
	if defaultStore.AvatarURL != "" && ginContext != nil {
		defaultStore.AvatarURL = utils.GetBaseURL(ginContext.Request) + defaultStore.AvatarURL
	}
	return defaultStore, nil
}

// GetPrinterSetting 获取打印机设置
func (s *Srv) GetPrinterSetting(ctx context.Context, languageList []dto.LanguageItem) (setting.Printer, error) {
	var (
		err     error
		printer setting.Printer
	)
	if languageList == nil {
		languageList, err = s.GetStoreLanguageList(ctx)
		if err != nil {
			return printer, err
		}
	}
	st := s.getSettingByKey(ctx, constant.SettingPrinter)
	err = json.Unmarshal([]byte(st.Values), &printer)
	if err != nil {
		ctx.Log().Error("解析小票打印机设置失败", zap.Error(err))
		return printer, errors.New("解析小票打印机设置失败")
	}
	// 过滤佛历、过滤打印方式，使用默认
	printer.CalendarList = nil
	printer.PrintList = nil
	defaultPrinter := s.getDefaultPrinter(ctx.GetLanguage(), languageList)
	err = copier.CopyWithOption(&defaultPrinter, printer, copier.Option{IgnoreEmpty: true})
	if err != nil {
		ctx.Log().Error("合并小票打印机设置失败", zap.Error(err))
		return printer, errors.New("合并小票打印机设置失败")
	}
	return defaultPrinter, nil
}

// GetPrinterInfo 获取打印机设置
func (s *Srv) GetPrinterInfo(ctx context.Context, printerSetting setting.Printer, deviceId string) (setting.PrinterInfo, error) {
	var (
		isCashierOpen    = printerSetting.CashierOpen == "1"
		printerId        string
		printerUuid      uint64
		printer          model.Printer
		err              error
		copies           uint = 1
		printerConfig    string
		printerType      string
		cashierBindKey   string
		isCashierPrinter bool // 是否收银机自带打印机
	)
	if isCashierOpen {
		for _, cashierPrinter := range printerSetting.CashierPrinter {
			if cashierPrinter.Key == deviceId {
				printerId = cashierPrinter.PrinterId // 如果是18位纯数字，说明是普通打印机
				break
			}
		}
		matched, _ := regexp.MatchString(`^\d+$`, printerId)
		if len(printerId) == 18 && matched { // 普通打印机 uuid uint64 字符串
			printerUuid, _ = strconv.ParseUint(printerId, 10, 64)
			printerRepo := repository.NewPrinterRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
			printer, err = printerRepo.GetPrinter(printerRepo.WhereUuid(printerUuid), printerRepo.WithPrinterType())
			if err != nil {
				return setting.PrinterInfo{}, err
			}
			copies = printer.Copies
			printerConfig = printer.ConfigJson
			printerType = printer.PrinterType.Key
		} else if printerId != "0" { // 收银机内置的打印机
			deviceRepo := repository.NewDeviceRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
			printerType = deviceRepo.GetDeviceBrand(deviceRepo.WhereSn(deviceId))
			cashierBindKey = printerId
			isCashierPrinter = true
		}
	}

	return setting.PrinterInfo{
		PrinterType:      printerType,
		PrinterUuid:      printerUuid, // 默认为0，如果是普通打印机，则为model.Printer的Uuid
		Copies:           copies,
		PrinterConfig:    printerConfig,
		IsCashierPrinter: isCashierPrinter,
		IsCashierOpen:    isCashierOpen,
		CashierBindKey:   cashierBindKey,
	}, nil
}

// GetBusinessSetting 门店业务设置
func (s *Srv) GetBusinessSetting(ctx context.Context) (setting.Business, error) {
	st := s.getSettingByKey(ctx, constant.SettingBusiness)
	var business setting.Business
	err := json.Unmarshal([]byte(st.Values), &business)
	if err != nil {
		ctx.Log().Error("解析门店-业务设置失败", zap.Error(err))
		return business, errors.New("解析门店-业务设置失败")
	}
	// 门店业务-过滤列表，使用默认
	business.ZeroingMethodList = nil
	business.CheckoutZeroingMethodList = nil
	business.GiftMethodList = nil
	business.FreeMethodList = nil
	defaultBusiness := s.getDefaultBusiness(ctx.GetLanguage())
	err = copier.CopyWithOption(&defaultBusiness, business, copier.Option{IgnoreEmpty: true})
	if err != nil {
		ctx.Log().Error("合并门店-业务设置失败", zap.Error(err))
		return business, errors.New("合并门店-业务设置失败")
	}
	return defaultBusiness, nil
}

// GetBuffetSetting 自助餐设置
func (s *Srv) GetBuffetSetting(ctx context.Context, companySetting model.CompanySetting) (setting.Buffet, error) {
	st := s.getSettingByKey(ctx, constant.SettingBuffet)
	var buffet setting.Buffet
	err := json.Unmarshal([]byte(st.Values), &buffet)
	if err != nil {
		return buffet, errors.New("解析自助餐-自助餐设置失败")
	}
	if companySetting.IsOpenBuffet == 0 {
		buffet.IsOpen = "0"
	}
	defaultBuffet := s.getDefaultBuffet()
	err = copier.CopyWithOption(&defaultBuffet, buffet, copier.Option{IgnoreEmpty: true})
	if err != nil {
		return buffet, errors.New("解析自助餐-自助餐设置失败")
	}
	return defaultBuffet, nil
}

// GetTabletSetting 平板端设置
func (s *Srv) GetTabletSetting(ctx context.Context, languageList []dto.LanguageItem) (setting.Tablet, error) {
	st := s.getSettingByKey(ctx, constant.SettingTablet)
	ginContext := ctx.GetGinContext()
	var (
		tablet setting.Tablet
		err    error
	)
	if languageList == nil {
		languageList, err = s.GetStoreLanguageList(ctx)
		if err != nil {
			return tablet, err
		}
	}
	err = json.Unmarshal([]byte(st.Values), &tablet)
	if err != nil {
		ctx.Log().Error("解析各端-平板端设置失败", zap.Error(err))
		return tablet, errors.New("解析各端-平板端设置失败")
	}
	// 滚动图/视频处理
	if len(tablet.Carousel) > 0 && ginContext != nil {
		for i, item := range tablet.Carousel {
			tablet.Carousel[i].FilePath = utils.AddImageDomain(item.FilePath, utils.GetBaseURL(ginContext.Request), true)
		}
	}
	defaultTablet := s.getDefaultTablet(languageList)
	// 语言 不需要合并
	defaultTablet.Language = nil
	err = copier.CopyWithOption(&defaultTablet, tablet, copier.Option{IgnoreEmpty: true})
	if err != nil {
		ctx.Log().Error("合并各端-平板端设置失败", zap.Error(err))
		return tablet, errors.New("合并各端-平板端设置失败")
	}

	return defaultTablet, nil
}

// GetPaymentSetting 门店-支付方式
func (s *Srv) GetPaymentSetting(ctx context.Context, companySetting model.CompanySetting) (setting.Payment, error) {
	var payment setting.Payment
	st := s.getSettingByKey(ctx, constant.SettingPayment)
	err := json.Unmarshal([]byte(st.Values), &payment)
	if err != nil {
		ctx.Log().Error("解析门店-支付方式失败", zap.Error(err))
		return payment, errors.New("解析门店-支付方式失败")
	}
	// 会员关闭时 门店管理 支付方式 余额这个开关要关了
	if companySetting.IsOpenMember == 0 {
		payment.IsBalance = "0"
	}
	defaultPayment := s.getDefaultPayment()
	err = copier.CopyWithOption(&defaultPayment, payment, copier.Option{IgnoreEmpty: true})
	if err != nil {
		ctx.Log().Error("合并门店-支付方式失败", zap.Error(err))
		return payment, errors.New("合并门店-支付方式失败")
	}
	return defaultPayment, nil
}

// GetCurrencySetting 货币单位设置
func (s *Srv) GetCurrencySetting(ctx context.Context) (setting.Currency, error) {
	st := s.getSettingByKey(ctx, constant.SettingCurrency)
	var currency setting.Currency
	err := json.Unmarshal([]byte(st.Values), &currency)

	if err != nil {
		ctx.Log().Error("解析门店-货币单位失败", zap.Error(err))
		return currency, errors.New("解析门店-货币单位失败")
	}
	defaultCurrency := s.getDefaultCurrency()
	err = copier.CopyWithOption(&defaultCurrency, currency, copier.Option{IgnoreEmpty: true})
	if err != nil {
		ctx.Log().Error("合并门店-货币单位失败", zap.Error(err))
		return currency, errors.New("合并门店-货币单位失败")
	}
	return defaultCurrency, nil
}

// GetCashierSetting 获取收银机设置
func (s *Srv) GetCashierSetting(ctx context.Context, languageList []dto.LanguageItem) (setting.Cashier, error) {
	var (
		err     error
		cashier setting.Cashier
	)
	if languageList == nil {
		languageList, err = s.GetStoreLanguageList(ctx)
		if err != nil {
			return cashier, err
		}
	}
	st := s.getSettingByKey(ctx, constant.SettingCashier)
	err = json.Unmarshal([]byte(st.Values), &cashier)
	if err != nil {
		ctx.Log().Error("解析各端-收银机设置失败", zap.Error(err))
		return cashier, errors.New("解析各端-收银机设置失败")
	}

	// 滚动图/视频处理
	ginContext := ctx.GetGinContext()
	if len(cashier.Carousel) > 0 && ginContext != nil {
		for i, item := range cashier.Carousel {
			cashier.Carousel[i].FilePath = utils.AddImageDomain(item.FilePath, utils.GetBaseURL(ginContext.Request), true)
		}
	}
	defaultCashier := s.getDefaultCashier(languageList)
	// 接单语音，设备本地处理，不需要合并
	cashier.IsAutoVoice = ""
	// 语言 不需要合并
	defaultCashier.Language = nil

	err = copier.CopyWithOption(&defaultCashier, cashier, copier.Option{IgnoreEmpty: true})
	if err != nil {
		ctx.Log().Error("合并各端-收银机设置失败", zap.Error(err))
		return cashier, errors.New("合并各端-收银机设置失败")
	}

	return defaultCashier, nil
}

// GetAssistantSetting 获取点餐助手设置
func (s *Srv) GetAssistantSetting(ctx context.Context, languageList []dto.LanguageItem) (setting.Assistant, error) {
	var (
		err       error
		assistant setting.Assistant
	)
	if languageList == nil {
		languageList, err = s.GetStoreLanguageList(ctx)
		if err != nil {
			return assistant, err
		}
	}
	st := s.getSettingByKey(ctx, constant.SettingAssistant)
	err = json.Unmarshal([]byte(st.Values), &assistant)
	if err != nil {
		ctx.Log().Error("解析各端-点餐助手设置失败", zap.Error(err))
		return assistant, errors.New("解析各端-点餐助手设置失败")
	}
	if len(assistant.LanguageList) == 0 {
		assistant.LanguageList = nil
	}
	defaultAssistant := s.getDefaultAssistant(ctx.GetLanguage(), languageList)
	err = copier.CopyWithOption(&defaultAssistant, assistant, copier.Option{IgnoreEmpty: true})
	if err != nil {
		ctx.Log().Error("合并各端-点餐助手设置失败", zap.Error(err))
		return assistant, errors.New("合并各端-点餐助手设置失败")
	}

	cashierSet := s.getSettingByKey(ctx, constant.SettingCashier)

	// 如果设置了 is_show_assistant_sold_out，则读取解析后的数据，否则读取默认设置
	if strings.Contains(cashierSet.Values, "\"is_show_assistant_sold_out\"") {
		var cashier setting.Cashier
		err = json.Unmarshal([]byte(st.Values), &cashier)
		if err != nil {
			ctx.Log().Error("解析各端-收银机设置失败", zap.Error(err))
			return assistant, errors.New("解析各端-收银机设置失败")
		}
		defaultAssistant.IsShowAssistantSoldOut = cashier.IsShowAssistantSoldOut
	} else {
		defaultAssistant.IsShowAssistantSoldOut = s.getDefaultCashier(languageList).IsShowAssistantSoldOut
	}
	return defaultAssistant, nil
}

// GetKitchenSetting 获取厨显端设置
func (s *Srv) GetKitchenSetting(ctx context.Context, companySetting model.CompanySetting, languageList []dto.LanguageItem) (setting.Kitchen, error) {
	var kitchen setting.Kitchen
	st := s.getSettingByKey(ctx, constant.SettingKitchen)

	err := json.Unmarshal([]byte(st.Values), &kitchen)
	if err != nil {
		ctx.Log().Error("解析各端-厨显设置失败", zap.Error(err))
		return kitchen, errors.New("解析各端-厨显设置失败")
	}
	if languageList == nil {
		languageList, err = s.GetStoreLanguageList(ctx)
		if err != nil {
			return kitchen, err
		}
	}
	defaultKitchen := s.getDefaultKitchen(languageList)

	// 语言 不需要合并
	defaultKitchen.Language = nil

	err = copier.CopyWithOption(&defaultKitchen, kitchen, copier.Option{IgnoreEmpty: true})
	if err != nil {
		ctx.Log().Error("合并各端-厨显设置失败", zap.Error(err))
		return kitchen, errors.New("合并各端-厨显设置失败")
	}
	// 总权限 - 不开启厨显
	if companySetting.IsOpenKitchenKds == 0 {
		kitchen.IsOpen = "0"
	}
	return kitchen, nil
}

// GetH5Setting 获取点餐助手设置
func (s *Srv) GetH5Setting(ctx context.Context, languageList []dto.LanguageItem) (setting.H5, error) {
	var (
		err error
		h5  setting.H5
	)
	if languageList == nil {
		languageList, err = s.GetStoreLanguageList(ctx)
		if err != nil {
			return h5, err
		}
	}
	st := s.getSettingByKey(ctx, constant.SettingH5)
	err = json.Unmarshal([]byte(st.Values), &h5)
	if err != nil {
		ctx.Log().Error("解析各端-扫码H5设置失败", zap.Error(err))
		return h5, errors.New("解析各端-扫码H5设置失败")
	}
	defaultH5 := s.getDefaultH5(languageList)
	err = copier.CopyWithOption(&defaultH5, h5, copier.Option{IgnoreEmpty: true})
	if err != nil {
		ctx.Log().Error("合并各端-扫码H5设置失败", zap.Error(err))
		return h5, errors.New("合并各端-扫码H5设置失败")
	}

	// 如果设置了 is_show_scan_sold_out，则读取解析后的数据，否则读取默认设置
	cashierSet := s.getSettingByKey(ctx, constant.SettingCashier)
	if strings.Contains(cashierSet.Values, "\"is_show_scan_sold_out\"") {
		var cashier setting.Cashier
		err = json.Unmarshal([]byte(st.Values), &cashier)
		if err != nil {
			ctx.Log().Error("解析各端-收银机设置失败", zap.Error(err))
			return h5, errors.New("解析各端-收银机设置失败")
		}
		defaultH5.IsShowScanSoldOut = cashier.IsShowScanSoldOut
	} else {
		defaultH5.IsShowScanSoldOut = s.getDefaultCashier(languageList).IsShowScanSoldOut
	}
	return defaultH5, nil
}

// GetCompanySetting 获取公司设置
func (s *Srv) GetCompanySetting(ctx context.Context) (model.CompanySetting, error) {
	companySettingRepo := repository.NewCompanySettingRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	return companySettingRepo.Get(), nil
}

// GetCashierLanguage 获取收银机语言
func (s *Srv) GetCashierLanguage(c context.Context) (resp.LanguageResp, error) {
	cashierSetting, err := s.GetCashierSetting(c, nil)
	if err != nil {
		return resp.LanguageResp{}, errors.New("获取语言失败")
	}
	return resp.LanguageResp{
		Languages:       cashierSetting.Language,
		LanguageList:    cashierSetting.LanguageList,
		DefaultLanguage: cashierSetting.DefaultLanguage,
	}, nil
}

// GetCashierAd 获取收银机副屏广告
func (s *Srv) GetCashierAd(ctx context.Context) (resp.Ads, error) {
	cashierSetting, err := s.GetCashierSetting(ctx, nil)
	if err != nil {
		return resp.Ads{List: make([]setting.CarouselItem, 0)}, errors.New("获取副屏广告失败")
	}
	return resp.Ads{
		List: cashierSetting.Carousel,
	}, nil
}

// GetServiceFeeSetting 获取服务费设置
func (s *Srv) GetServiceFeeSetting(ctx context.Context) (setting.ServiceCharge, error) {
	st := s.getSettingByKey(ctx, constant.SettingServiceCharge)
	var serviceFee setting.ServiceCharge
	err := json.Unmarshal([]byte(st.Values), &serviceFee)
	if err != nil {
		return serviceFee, errors.New("解析服务费设置失败")
	}
	if serviceFee.IsOpen == "0" {
		serviceFee.IsOpen = "0"
	}
	defaultServiceFee := s.getDefaultServiceCharge()
	err = copier.CopyWithOption(&defaultServiceFee, serviceFee, copier.Option{IgnoreEmpty: true})
	if err != nil {
		return serviceFee, errors.New("解析服务费设置失败")
	}
	return defaultServiceFee, nil
}

// GetTaxRateSetting 获取税率设置
func (s *Srv) GetTaxRateSetting(ctx context.Context) (setting.TaxRate, error) {
	st := s.getSettingByKey(ctx, constant.SettingTaxRate)
	var taxRate setting.TaxRate
	err := json.Unmarshal([]byte(st.Values), &taxRate)
	if err != nil {
		return taxRate, errors.New("解析税率设置失败")
	}
	if taxRate.IsOpen == "0" {
		taxRate.IsOpen = "0"
	}
	defaultTaxRate := s.getDefaultTaxRate()
	err = copier.CopyWithOption(&defaultTaxRate, taxRate, copier.Option{IgnoreEmpty: true})
	if err != nil {
		return taxRate, errors.New("解析税率设置失败")
	}
	return defaultTaxRate, nil
}

// CashierVerifyPassword 收银机验证密码
func (s *Srv) CashierVerifyPassword(ctx context.Context, typ string, password string, companyUuid uint64) bool {
	cashierSetting, err := s.GetCashierSetting(ctx, nil)
	if err != nil {
		return false
	}
	switch typ {
	case constant.PasswordTypeCashBox:
		return cashierSetting.CashierPassword == password
	case constant.PasswordTypeAdvanced:
		return cashierSetting.AdvancedPassword == password
	case constant.PasswordTypeLock:
		return cashierSetting.LockPassword == password
	default:
		return false
	}
}

// CheckUpdate 检查更新
func (s *Srv) CheckUpdate(ctx context.Context, appType int, brand string, language string) (resp.UpdateInfo, error) {
	type UpdateData struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			VersionName  string `json:"version_name"`
			ForcedUpdate int    `json:"forced_update"`
			UpdateLog    string `json:"update_log"`
			DownloadURL  string `json:"download_url"`
		} `json:"data"`
	}
	url := fmt.Sprintf("%s/api/admin/client.client/getNewVersion?type=%d&brand=%s&language=%s", viper.GetString("CLOUD_PLATFORM_HOST"), appType, brand, language)
	res, err := gohttp.NewRequest().Post(url)
	if err != nil {
		return resp.UpdateInfo{}, errors.New("获取最新版本信息失败")
	}
	bodyBytes, _ := res.GetBodyAsByte()
	var updateData UpdateData
	if err := json.Unmarshal(bodyBytes, &updateData); err != nil {
		ctx.Log().Error("解析版本更新信息失败", zap.Error(err))
	}

	return resp.UpdateInfo{
		VersionName:  updateData.Data.VersionName,
		ForcedUpdate: updateData.Data.ForcedUpdate,
		UpdateLog:    updateData.Data.UpdateLog,
		DownloadURL:  updateData.Data.DownloadURL,
	}, nil
}

// EditAcceptOrderSetting 修改自动接单参数
func (s *Srv) EditAcceptOrderSetting(ctx context.Context, orderSetting req.UpdateAcceptOrderSetting) error { // 修改自动接单设置
	cashierSetting, err := s.GetCashierSetting(ctx, nil)
	if err != nil {
		return err
	}
	cashierSetting.IsAutoOrder = orderSetting.IsAutoOrder
	cashierSetting.AutoOrderLimit = orderSetting.AutoOrderLimit
	return s.UpdateSetting(ctx, constant.SettingCashier, cashierSetting)
}

// EditSystemSetting 修改系统设置
func (s *Srv) EditSystemSetting(ctx context.Context, systemSetting req.UpdateSystemSetting) error { // // 修改系统设置
	cashierSetting, err := s.GetCashierSetting(ctx, nil)
	if err != nil {
		return err
	}
	cashierSetting.IsShowAssistantSoldOut = *systemSetting.IsShowAssistantSoldOut
	cashierSetting.IsShowScanSoldOut = *systemSetting.IsShowScanSoldOut
	cashierSetting.MenuShowSoldOut = systemSetting.MenuShowSoldOut
	if err := s.UpdateSetting(ctx, constant.SettingCashier, cashierSetting); err != nil {
		return err
	}
	businessSetting, err := s.GetBusinessSetting(ctx)
	if err != nil {
		return err
	}
	businessSetting.DishCardStyle = systemSetting.DishCardStyle
	if err := s.UpdateSetting(ctx, constant.SettingBusiness, businessSetting); err != nil {
		return err
	}
	tabletSetting, err := s.GetTabletSetting(ctx, nil)
	if err != nil {
		return err
	}
	tabletSetting.IsShowSoldOut = *systemSetting.IsShowSoldOut
	if err := s.UpdateSetting(ctx, constant.SettingTablet, tabletSetting); err != nil {
		return err
	}
	return nil
}

// GetCashierBaseSetting 获取收银端设置
func (s *Srv) GetCashierBaseSetting(ctx context.Context) (resp.CashierBaseSetting, error) {
	var settingResp resp.CashierBaseSetting
	cashierSetting, err := s.GetCashierSetting(ctx, nil)
	if err != nil {
		return settingResp, err
	}
	businessSetting, err := s.GetBusinessSetting(ctx)
	if err != nil {
		return settingResp, err
	}
	tabletSetting, err := s.GetTabletSetting(ctx, nil)
	if err != nil {
		return settingResp, err
	}

	clientVersion := ctx.GetGinContext().GetHeader("Version-Name")
	if clientVersion == "" {
		clientVersion = "0.0.0"
	}

	deviceRepo := repository.NewDeviceRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))

	device, err := deviceRepo.GetDevice(deviceRepo.WhereSn(ctx.GetDeviceSn()))
	if err != nil {
		return settingResp, errors2.ErrInternal
	}
	return resp.CashierBaseSetting{
		AcceptOrder: resp.AcceptOrderSetting{
			IsAutoOrder:    cashierSetting.IsAutoOrder,
			AutoOrderLimit: cashierSetting.AutoOrderLimit,
			IsAutoVoice:    cashierSetting.IsAutoVoice,
		},
		System: resp.SystemSetting{
			IsShowScanSoldOut:      cashierSetting.IsShowScanSoldOut,
			IsShowAssistantSoldOut: cashierSetting.IsShowAssistantSoldOut,
			MenuShowSoldOut:        cashierSetting.MenuShowSoldOut,
			DishCardStyle:          businessSetting.DishCardStyle,
			IsShowSoldOut:          tabletSetting.IsShowSoldOut,
			DefaultLanguage:        cashierSetting.DefaultLanguage,
			SecondLanguage:         cashierSetting.DefaultLanguage,
			DeviceId:               ctx.GetDeviceSn(),
			DeviceRemark:           device.Remark,
			ClientVersion:          clientVersion,
			ServerVersion:          utils.GetVersion("version.json"),
		},
	}, nil

}

// UpdateSetting 更新设置
func (s *Srv) UpdateSetting(ctx context.Context, settingKey string, values any) error {
	value, err := json.Marshal(values)
	if err != nil {
		return errors.New("更新设置失败")
	}
	settingRepo := repository.NewSettingRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	set := settingRepo.GetByKey(settingKey)
	if set.Key == "" {
		if _, err = settingRepo.Create(model.Setting{
			Key:         settingKey,
			Description: "",
			Values:      string(value),
		}); err != nil {
			return errors.New("更新设置失败")
		}
	} else {
		if err = settingRepo.Updates(settingKey, string(value)); err != nil {
			return errors.New("更新设置失败")
		}
	}

	// 删除缓存
	s.cache.Del(fmt.Sprintf(s.cacheKey, ctx.GetCompanyUuid()))
	return nil
}
