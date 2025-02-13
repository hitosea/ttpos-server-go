package setting

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type ISrv interface {
	GetAll(companyUuid uint64, language string, languageList []dto.LanguageItem, cc *gin.Context) (map[string]any, error)                 // 获取所有设置，性能问题慎用
	GetStoreSetting(companyUuid uint64, language string, cc *gin.Context) (setting.Store, error)                                          // 获取商家设置
	GetStoreLanguageList(companyUuid uint64, language string, cc *gin.Context) ([]dto.LanguageItem, error)                                // 获取商家语言
	GetPrinterSetting(companyUuid uint64, language string, cc *gin.Context, languageList []dto.LanguageItem) (setting.Printer, error)     // 获取打印机设置
	GetCashierSetting(companyUuid uint64, language string, cc *gin.Context, languageList []dto.LanguageItem) (setting.Cashier, error)     // 获取收银机设置
	GetAssistantSetting(companyUuid uint64, language string, cc *gin.Context, languageList []dto.LanguageItem) (setting.Assistant, error) // 获取点餐助手设置
	GetBusinessSetting(companyUuid uint64, language string) (setting.Business, error)                                                     // 获取门店业务设置
	GetBuffetSetting(companyUuid uint64, companySetting model.CompanySetting) (setting.Buffet, error)                                     // 获取自助餐设置
	GetCurrencySetting(companyUuid uint64) (setting.Currency, error)                                                                      // 获取货币单位设置
	GetH5Setting(companyUuid uint64, language string, cc *gin.Context, languageList []dto.LanguageItem) (setting.H5, error)               // 获取扫码H5设置
	GetCompanySetting(companyUuid uint64) (model.CompanySetting, error)                                                                   // 获取公司设置
	Updates(companyUuid uint64, settingKey string, values any) error                                                                      // 更新设置
	VerifyAdvancedPassword(companyUuid uint64, password string) error                                                                     // 验证高级密码
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
func (s *Srv) fromCache(companyUuid uint64) ([]model.Setting, error) {
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
		logger.Logger.Error("从数据库获取设置失败", zap.Error(err))
		return nil, errors.New("获取设置失败")
	}

	data, _ := json.Marshal(settings)
	s.cache.Set(cacheKey, string(data), 0)

	return settings, nil
}

// GetAll 获取所有设置
func (s *Srv) GetAll(companyUuid uint64, language string, languageList []dto.LanguageItem, cc *gin.Context) (map[string]any, error) {

	// 获取company_setting
	companySettingRepo := repository.NewCompanySettingRepo(s.dbm.GetDB(companyUuid))
	companySetting := companySettingRepo.Get()

	var retSettings = make(map[string]any)
	// 从缓存读取
	settings, err := s.fromCache(companyUuid)

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
		switch st.Key {
		case constant.SettingPrinter: // 小票打印机设置
			var printer setting.Printer
			err = json.Unmarshal([]byte(st.Values), &printer)
			if err != nil {
				logger.Logger.Error("解析小票打印机设置失败", zap.Error(err))
				return nil, errors.New("解析小票打印机设置失败")
			}
			// 过滤佛历、过滤打印方式，使用默认
			printer.CalendarList = nil
			printer.PrintList = nil
			tmp := s.getDefaultPrinter(language, languageList)
			err = copier.CopyWithOption(&tmp, printer, copier.Option{IgnoreEmpty: true})
			if err != nil {
				logger.Logger.Error("合并小票打印机设置失败", zap.Error(err))
				return nil, errors.New("合并小票打印机设置失败")
			}
			retSettings[constant.SettingPrinter] = tmp
		case constant.SettingStore: // 商城设置
			var store setting.Store
			err = json.Unmarshal([]byte(st.Values), &store)
			if err != nil {
				logger.Logger.Error("解析商城设置失败", zap.Error(err))
				return nil, errors.New("解析商城设置失败")
			}
			if store.IPWhiteList != "" {
				store.IPWhiteList = viper.GetString("PAY_SERVICE_IP")
			}

			tmp := s.getDefaultStore(language)
			err = copier.CopyWithOption(&tmp, store, copier.Option{IgnoreEmpty: true})
			if err != nil {
				logger.Logger.Error("合并商城设置失败", zap.Error(err))
				return nil, errors.New("合并商城设置失败")
			}
			if tmp.LogoURL != "" && cc != nil {
				tmp.LogoURL = utils.GetBaseURL(cc.Request) + tmp.LogoURL
			}
			if tmp.AvatarURL != "" && cc != nil {
				tmp.AvatarURL = utils.GetBaseURL(cc.Request) + tmp.AvatarURL
			}

			retSettings[constant.SettingStore] = tmp
		case constant.SettingRecharge: // 充值设置
			var recharge setting.Recharge
			err = json.Unmarshal([]byte(st.Values), &recharge)
			if err != nil {
				logger.Logger.Error("解析充值设置失败", zap.Error(err))
				return nil, errors.New("解析充值设置失败")
			}
			tmp := s.getDefaultRecharge()
			err = copier.CopyWithOption(&tmp, recharge, copier.Option{IgnoreEmpty: true})
			if err != nil {
				logger.Logger.Error("合并充值设置失败", zap.Error(err))
				return nil, errors.New("合并充值设置失败")
			}
			retSettings[constant.SettingRecharge] = tmp
		case constant.SettingPoints: // 积分设置
			var points setting.Points
			err = json.Unmarshal([]byte(st.Values), &points)
			if err != nil {
				logger.Logger.Error("解析积分设置失败", zap.Error(err))
				return nil, errors.New("解析积分设置失败")
			}
			tmp := s.getDefaultPoints()
			err = copier.CopyWithOption(&tmp, points, copier.Option{IgnoreEmpty: true})
			if err != nil {
				logger.Logger.Error("合并积分设置失败", zap.Error(err))
				return nil, errors.New("合并积分设置失败")
			}
			retSettings[constant.SettingPoints] = tmp
		case constant.SettingSysAdminConfig: // 系统配置
			var sysAdminConfig setting.SysAdminConfig
			err = json.Unmarshal([]byte(st.Values), &sysAdminConfig)
			if err != nil {
				logger.Logger.Error("解析系统配置失败", zap.Error(err))
				return nil, errors.New("解析系统配置失败")
			}
			tmp := s.getDefaultSysAdminConfig()
			err = copier.CopyWithOption(&tmp, sysAdminConfig, copier.Option{IgnoreEmpty: true})
			if err != nil {
				logger.Logger.Error("合并系统配置失败", zap.Error(err))
				return nil, errors.New("合并系统配置失败")
			}
			retSettings[constant.SettingSysAdminConfig] = tmp
		case constant.SettingSysConfig: // 系统配置
			var sysConfig setting.SysConfig
			err = json.Unmarshal([]byte(st.Values), &sysConfig)
			if err != nil {
				logger.Logger.Error("解析系统配置失败", zap.Error(err))
				return nil, errors.New("解析系统配置失败")
			}
			tmp := s.getDefaultSysConfig()
			err = copier.CopyWithOption(&tmp, sysConfig, copier.Option{IgnoreEmpty: true})
			if err != nil {
				logger.Logger.Error("合并系统配置失败", zap.Error(err))
				return nil, errors.New("合并系统配置失败")
			}
			retSettings[constant.SettingSysConfig] = tmp
		case constant.SettingBalance: // 充值设置
			var balance setting.Balance
			err = json.Unmarshal([]byte(st.Values), &balance)
			if err != nil {
				logger.Logger.Error("解析充值设置失败", zap.Error(err))
				return nil, errors.New("解析充值设置失败")
			}
			tmp := s.getDefaultBalance()
			err = copier.CopyWithOption(&tmp, balance, copier.Option{IgnoreEmpty: true})
			if err != nil {
				logger.Logger.Error("合并充值设置失败", zap.Error(err))
				return nil, errors.New("合并充值设置失败")
			}
			retSettings[constant.SettingBalance] = tmp
		case constant.SettingCurrency: // 门店-货币单位
			var currency setting.Currency
			err = json.Unmarshal([]byte(st.Values), &currency)
			if err != nil {
				logger.Logger.Error("解析门店-货币单位失败", zap.Error(err))
				return nil, errors.New("解析门店-货币单位失败")
			}
			tmp := s.getDefaultCurrency()
			err = copier.CopyWithOption(&tmp, currency, copier.Option{IgnoreEmpty: true})
			if err != nil {
				logger.Logger.Error("合并门店-货币单位失败", zap.Error(err))
				return nil, errors.New("合并门店-货币单位失败")
			}
			retSettings[constant.SettingCurrency] = tmp
		case constant.SettingTaxRate: // 门店-税率管理
			var taxRate setting.TaxRate
			err = json.Unmarshal([]byte(st.Values), &taxRate)
			if err != nil {
				logger.Logger.Error("解析门店-税率管理失败", zap.Error(err))
				return nil, errors.New("解析门店-税率管理失败")
			}
			tmp := s.getDefaultTaxRate()
			err = copier.CopyWithOption(&tmp, taxRate, copier.Option{IgnoreEmpty: true})
			if err != nil {
				logger.Logger.Error("合并门店-税率管理失败", zap.Error(err))
				return nil, errors.New("合并门店-税率管理失败")
			}
			retSettings[constant.SettingTaxRate] = tmp
		case constant.SettingServiceCharge: // 门店-服务费
			var serviceCharge setting.ServiceCharge
			err = json.Unmarshal([]byte(st.Values), &serviceCharge)
			if err != nil {
				logger.Logger.Error("解析门店-服务费失败", zap.Error(err))
				return nil, errors.New("解析门店-服务费失败")
			}
			tmp := s.getDefaultServiceCharge()
			err = copier.CopyWithOption(&tmp, serviceCharge, copier.Option{IgnoreEmpty: true})
			if err != nil {
				logger.Logger.Error("合并门店-服务费失败", zap.Error(err))
				return nil, errors.New("合并门店-服务费失败")
			}
			retSettings[constant.SettingServiceCharge] = tmp
		case constant.SettingPayment: // 门店-支付方式
			var payment setting.Payment
			err = json.Unmarshal([]byte(st.Values), &payment)
			if err != nil {
				logger.Logger.Error("解析门店-支付方式失败", zap.Error(err))
				return nil, errors.New("解析门店-支付方式失败")
			}

			// 会员关闭时 门店管理 支付方式 余额这个开关要关了
			if companySetting.IsOpenMember == 0 {
				payment.IsBalance = "0"
			}

			tmp := s.getDefaultPayment()
			err = copier.CopyWithOption(&tmp, payment, copier.Option{IgnoreEmpty: true})
			if err != nil {
				logger.Logger.Error("合并门店-支付方式失败", zap.Error(err))
				return nil, errors.New("合并门店-支付方式失败")
			}

			retSettings[constant.SettingPayment] = tmp
		case constant.SettingBusiness: // 门店-业务设置
			var business setting.Business
			err = json.Unmarshal([]byte(st.Values), &business)
			if err != nil {
				logger.Logger.Error("解析门店-业务设置失败", zap.Error(err))
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
				logger.Logger.Error("合并门店-业务设置失败", zap.Error(err))
				return nil, errors.New("合并门店-业务设置失败")
			}
			retSettings[constant.SettingBusiness] = tmp
		case constant.SettingCashier: // 各端-收银机设置
			var cashier setting.Cashier
			err = json.Unmarshal([]byte(st.Values), &cashier)
			if err != nil {
				logger.Logger.Error("解析各端-收银机设置失败", zap.Error(err))
				return nil, errors.New("解析各端-收银机设置失败")
			}

			// 滚动图/视频处理
			if len(cashier.Carousel) > 0 && cc != nil {
				for i, item := range cashier.Carousel {
					cashier.Carousel[i].FilePath = utils.AddImageDomain(item.FilePath, utils.GetBaseURL(cc.Request), true)
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
				logger.Logger.Error("合并各端-收银机设置失败", zap.Error(err))
				return nil, errors.New("合并各端-收银机设置失败")
			}
			retSettings[constant.SettingCashier] = tmp
		case constant.SettingTablet: // 各端-平板端设置
			var tablet setting.Tablet
			err = json.Unmarshal([]byte(st.Values), &tablet)
			if err != nil {
				logger.Logger.Error("解析各端-平板端设置失败", zap.Error(err))
				return nil, errors.New("解析各端-平板端设置失败")
			}
			// 滚动图/视频处理
			if len(tablet.Carousel) > 0 && cc != nil {
				for i, item := range tablet.Carousel {
					tablet.Carousel[i].FilePath = utils.AddImageDomain(item.FilePath, utils.GetBaseURL(cc.Request), true)
				}
			}
			tmp := s.getDefaultTablet(languageList)
			// 语言 不需要合并
			tmp.Language = nil
			err = copier.CopyWithOption(&tmp, tablet, copier.Option{IgnoreEmpty: true})
			if err != nil {
				logger.Logger.Error("合并各端-平板端设置失败", zap.Error(err))
				return nil, errors.New("合并各端-平板端设置失败")
			}
			retSettings[constant.SettingTablet] = tmp
		case constant.SettingH5: // 各端-扫码H5设置
			var h5 setting.H5
			err = json.Unmarshal([]byte(st.Values), &h5)
			if err != nil {
				logger.Logger.Error("解析各端-扫码H5设置失败", zap.Error(err))
				return nil, errors.New("解析各端-扫码H5设置失败")
			}
			tmp := s.getDefaultH5(languageList)
			err = copier.CopyWithOption(&tmp, h5, copier.Option{IgnoreEmpty: true})
			if err != nil {
				logger.Logger.Error("合并各端-扫码H5设置失败", zap.Error(err))
				return nil, errors.New("合并各端-扫码H5设置失败")
			}
			retSettings[constant.SettingH5] = tmp
		case constant.SettingKitchen: // 各端-厨显设置
			var kitchen setting.Kitchen
			err = json.Unmarshal([]byte(st.Values), &kitchen)
			if err != nil {
				logger.Logger.Error("解析各端-厨显设置失败", zap.Error(err))
				return nil, errors.New("解析各端-厨显设置失败")
			}

			tmp := s.getDefaultKitchen(languageList)

			// 语言 不需要合并
			tmp.Language = nil

			err = copier.CopyWithOption(&tmp, kitchen, copier.Option{IgnoreEmpty: true})
			if err != nil {
				logger.Logger.Error("合并各端-厨显设置失败", zap.Error(err))
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
				logger.Logger.Error("解析各端-点餐助手设置失败", zap.Error(err))
				return nil, errors.New("解析各端-点餐助手设置失败")
			}
			if len(assistant.LanguageList) == 0 {
				assistant.LanguageList = nil
			}
			tmp := s.getDefaultAssistant(language, languageList)
			err = copier.CopyWithOption(&tmp, assistant, copier.Option{IgnoreEmpty: true})
			if err != nil {
				logger.Logger.Error("合并各端-点餐助手设置失败", zap.Error(err))
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
func (s *Srv) GetStoreLanguageList(companyId uint64, language string, cc *gin.Context) ([]dto.LanguageItem, error) {
	set, err := s.GetStoreSetting(companyId, language, cc)
	if err != nil {
		return nil, err
	}
	return set.Language, nil
}

func (s *Srv) getSettingByKey(companyUuid uint64, key string) model.Setting {
	allSettings, _ := s.fromCache(companyUuid)
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
func (s *Srv) GetStoreSetting(companyUuid uint64, language string, cc *gin.Context) (setting.Store, error) {
	var store setting.Store
	st := s.getSettingByKey(companyUuid, constant.SettingStore)
	err := json.Unmarshal([]byte(st.Values), &store)
	if err != nil {
		logger.Logger.Error("解析商城设置失败", zap.Error(err))
		return store, errors.New("解析商城设置失败")
	}
	if store.IPWhiteList != "" {
		store.IPWhiteList = viper.GetString("PAY_SERVICE_IP")
	}

	defaultStore := s.getDefaultStore(language)
	err = copier.CopyWithOption(&defaultStore, store, copier.Option{IgnoreEmpty: true})
	if err != nil {
		logger.Logger.Error("合并商城设置失败", zap.Error(err))
		return store, errors.New("合并商城设置失败")
	}
	if defaultStore.LogoURL != "" && cc != nil {
		defaultStore.LogoURL = utils.GetBaseURL(cc.Request) + defaultStore.LogoURL
	}
	if defaultStore.AvatarURL != "" && cc != nil {
		defaultStore.AvatarURL = utils.GetBaseURL(cc.Request) + defaultStore.AvatarURL
	}
	return defaultStore, nil
}

// GetPrinterSetting 获取打印机设置
func (s *Srv) GetPrinterSetting(companyUuid uint64, language string, cc *gin.Context, languageList []dto.LanguageItem) (setting.Printer, error) {
	var (
		err     error
		printer setting.Printer
	)
	if len(languageList) == 0 {
		languageList, err = s.GetStoreLanguageList(companyUuid, language, cc)
		if err != nil {
			return printer, err
		}
	}
	st := s.getSettingByKey(companyUuid, constant.SettingPrinter)
	err = json.Unmarshal([]byte(st.Values), &printer)
	if err != nil {
		logger.Logger.Error("解析小票打印机设置失败", zap.Error(err))
		return printer, errors.New("解析小票打印机设置失败")
	}
	// 过滤佛历、过滤打印方式，使用默认
	printer.CalendarList = nil
	printer.PrintList = nil
	defaultPrinter := s.getDefaultPrinter(language, languageList)
	err = copier.CopyWithOption(&defaultPrinter, printer, copier.Option{IgnoreEmpty: true})
	if err != nil {
		logger.Logger.Error("合并小票打印机设置失败", zap.Error(err))
		return printer, errors.New("合并小票打印机设置失败")
	}
	return defaultPrinter, nil
}

// GetBusinessSetting 门店业务设置
func (s *Srv) GetBusinessSetting(companyUuid uint64, language string) (setting.Business, error) {
	st := s.getSettingByKey(companyUuid, constant.SettingBusiness)
	var business setting.Business
	err := json.Unmarshal([]byte(st.Values), &business)
	if err != nil {
		logger.Logger.Error("解析门店-业务设置失败", zap.Error(err))
		return business, errors.New("解析门店-业务设置失败")
	}
	// 门店业务-过滤列表，使用默认
	business.ZeroingMethodList = nil
	business.CheckoutZeroingMethodList = nil
	business.GiftMethodList = nil
	business.FreeMethodList = nil
	defaultBusiness := s.getDefaultBusiness(language)
	err = copier.CopyWithOption(&defaultBusiness, business, copier.Option{IgnoreEmpty: true})
	if err != nil {
		logger.Logger.Error("合并门店-业务设置失败", zap.Error(err))
		return business, errors.New("合并门店-业务设置失败")
	}
	return defaultBusiness, nil
}

// GetBuffetSetting 自助餐设置
func (s *Srv) GetBuffetSetting(companyUuid uint64, companySetting model.CompanySetting) (setting.Buffet, error) {
	st := s.getSettingByKey(companyUuid, constant.SettingBuffet)
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

// GetCurrencySetting 货币单位设置
func (s *Srv) GetCurrencySetting(companyUuid uint64) (setting.Currency, error) {
	st := s.getSettingByKey(companyUuid, constant.SettingCurrency)
	var currency setting.Currency
	err := json.Unmarshal([]byte(st.Values), &currency)

	if err != nil {
		logger.Logger.Error("解析门店-货币单位失败", zap.Error(err))
		return currency, errors.New("解析门店-货币单位失败")
	}
	defaultCurrency := s.getDefaultCurrency()
	err = copier.CopyWithOption(&defaultCurrency, currency, copier.Option{IgnoreEmpty: true})
	if err != nil {
		logger.Logger.Error("合并门店-货币单位失败", zap.Error(err))
		return currency, errors.New("合并门店-货币单位失败")
	}
	return defaultCurrency, nil
}

// GetCashierSetting 获取收银机设置
func (s *Srv) GetCashierSetting(companyUuid uint64, language string, cc *gin.Context, languageList []dto.LanguageItem) (setting.Cashier, error) {
	var (
		err     error
		cashier setting.Cashier
	)
	if len(languageList) == 0 {
		languageList, err = s.GetStoreLanguageList(companyUuid, language, cc)
		if err != nil {
			return cashier, err
		}
	}
	st := s.getSettingByKey(companyUuid, constant.SettingCashier)
	err = json.Unmarshal([]byte(st.Values), &cashier)
	if err != nil {
		logger.Logger.Error("解析各端-收银机设置失败", zap.Error(err))
		return cashier, errors.New("解析各端-收银机设置失败")
	}

	// 滚动图/视频处理
	if len(cashier.Carousel) > 0 && cc != nil {
		for i, item := range cashier.Carousel {
			cashier.Carousel[i].FilePath = utils.AddImageDomain(item.FilePath, utils.GetBaseURL(cc.Request), true)
		}
	}
	defaultCashier := s.getDefaultCashier(languageList)
	// 接单语音，设备本地处理，不需要合并
	cashier.IsAutoVoice = ""
	// 语言 不需要合并
	defaultCashier.Language = nil

	err = copier.CopyWithOption(&defaultCashier, cashier, copier.Option{IgnoreEmpty: true})
	if err != nil {
		logger.Logger.Error("合并各端-收银机设置失败", zap.Error(err))
		return cashier, errors.New("合并各端-收银机设置失败")
	}

	return defaultCashier, nil
}

// GetAssistantSetting 获取点餐助手设置
func (s *Srv) GetAssistantSetting(companyUuid uint64, language string, cc *gin.Context, languageList []dto.LanguageItem) (setting.Assistant, error) {
	var (
		err       error
		assistant setting.Assistant
	)
	if len(languageList) == 0 {
		languageList, err = s.GetStoreLanguageList(companyUuid, language, cc)
		if err != nil {
			return assistant, err
		}
	}
	st := s.getSettingByKey(companyUuid, constant.SettingAssistant)
	err = json.Unmarshal([]byte(st.Values), &assistant)
	if err != nil {
		logger.Logger.Error("解析各端-点餐助手设置失败", zap.Error(err))
		return assistant, errors.New("解析各端-点餐助手设置失败")
	}
	if len(assistant.LanguageList) == 0 {
		assistant.LanguageList = nil
	}
	defaultAssistant := s.getDefaultAssistant(language, languageList)
	err = copier.CopyWithOption(&defaultAssistant, assistant, copier.Option{IgnoreEmpty: true})
	if err != nil {
		logger.Logger.Error("合并各端-点餐助手设置失败", zap.Error(err))
		return assistant, errors.New("合并各端-点餐助手设置失败")
	}

	cashierSet := s.getSettingByKey(companyUuid, constant.SettingCashier)

	// 如果设置了 is_show_assistant_sold_out，则读取解析后的数据，否则读取默认设置
	if strings.Contains(cashierSet.Values, "\"is_show_assistant_sold_out\"") {
		var cashier setting.Cashier
		err = json.Unmarshal([]byte(st.Values), &cashier)
		if err != nil {
			logger.Logger.Error("解析各端-收银机设置失败", zap.Error(err))
			return assistant, errors.New("解析各端-收银机设置失败")
		}
		defaultAssistant.IsShowAssistantSoldOut = cashier.IsShowAssistantSoldOut
	} else {
		defaultAssistant.IsShowAssistantSoldOut = s.getDefaultCashier(languageList).IsShowAssistantSoldOut
	}
	return defaultAssistant, nil
}

// GetH5Setting 获取点餐助手设置
func (s *Srv) GetH5Setting(companyUuid uint64, language string, cc *gin.Context, languageList []dto.LanguageItem) (setting.H5, error) {
	var (
		err error
		h5  setting.H5
	)
	if len(languageList) == 0 {
		languageList, err = s.GetStoreLanguageList(companyUuid, language, cc)
		if err != nil {
			return h5, err
		}
	}
	st := s.getSettingByKey(companyUuid, constant.SettingH5)
	err = json.Unmarshal([]byte(st.Values), &h5)
	if err != nil {
		logger.Logger.Error("解析各端-扫码H5设置失败", zap.Error(err))
		return h5, errors.New("解析各端-扫码H5设置失败")
	}
	defaultH5 := s.getDefaultH5(languageList)
	err = copier.CopyWithOption(&defaultH5, h5, copier.Option{IgnoreEmpty: true})
	if err != nil {
		logger.Logger.Error("合并各端-扫码H5设置失败", zap.Error(err))
		return h5, errors.New("合并各端-扫码H5设置失败")
	}

	// 如果设置了 is_show_scan_sold_out，则读取解析后的数据，否则读取默认设置
	cashierSet := s.getSettingByKey(companyUuid, constant.SettingCashier)
	if strings.Contains(cashierSet.Values, "\"is_show_scan_sold_out\"") {
		var cashier setting.Cashier
		err = json.Unmarshal([]byte(st.Values), &cashier)
		if err != nil {
			logger.Logger.Error("解析各端-收银机设置失败", zap.Error(err))
			return h5, errors.New("解析各端-收银机设置失败")
		}
		defaultH5.IsShowScanSoldOut = cashier.IsShowScanSoldOut
	} else {
		defaultH5.IsShowScanSoldOut = s.getDefaultCashier(languageList).IsShowScanSoldOut
	}
	return defaultH5, nil
}

// GetCompanySetting 获取公司设置
func (s *Srv) GetCompanySetting(companyUuid uint64) (model.CompanySetting, error) {
	companySettingRepo := repository.NewCompanySettingRepo(s.dbm.GetDB(companyUuid))
	return companySettingRepo.Get(), nil
}

// Updates 更新设置
func (s *Srv) Updates(companyUuid uint64, settingKey string, values any) error {
	value, err := json.Marshal(values)
	if err != nil {
		return errors.New("更新设置失败")
	}
	settingRepo := repository.NewSettingRepo(s.dbm.GetDB(companyUuid))
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
	s.cache.Del(fmt.Sprintf(s.cacheKey, companyUuid))
	return nil
}
