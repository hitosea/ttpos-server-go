package application

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/setting/domain/entity"
	"ttpos-server-go/app/modules/setting/domain/repository"
	"ttpos-server-go/app/modules/setting/domain/service"
	"ttpos-server-go/app/modules/setting/domain/valueobject"
	memberRepo "ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	pkgctx "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"github.com/spf13/viper"
)

// getIPAndPort 获取服务器 IP 和端口（与旧服务保持一致）
func getIPAndPort() (string, string) {
	var serverIP, serverPort string
	serverIP = viper.GetString("HARDWARE_SERVER_URL")
	if serverIP == "" {
		serverIP, _ = utils.GetLocalIP()
	}
	serverIP = strings.ReplaceAll(serverIP, "addr:", "")
	serverPort = viper.GetString("HARDWARE_SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}
	return serverIP, serverPort
}

// SettingAppServiceImpl 设置应用服务实现
type SettingAppServiceImpl struct {
	settingRepo      repository.ISettingRepository
	settingDomainSvc service.ISettingDomainService
	cache            cache.Cache
	cacheKey         string
	cloudCacheKey    string
}

// NewSettingAppServiceImpl 创建设置应用服务实例
func NewSettingAppServiceImpl(settingRepo repository.ISettingRepository, settingDomainSvc service.ISettingDomainService, cache cache.Cache) ISettingAppService {
	return &SettingAppServiceImpl{
		settingRepo:      settingRepo,
		settingDomainSvc: settingDomainSvc,
		cache:            cache,
		cacheKey:         "settings:%s",
		cloudCacheKey:    "cloud_basic_setting",
	}
}

// GetStoreSetting 获取门店设置
func (s *SettingAppServiceImpl) GetStoreSetting(ctx context.Context) (entity.StoreSetting, error) {
	st := s.getSettingByKey(ctx, "store")

	var store entity.StoreSetting
	if st.Values != "{}" {
		// 解析设置
		var err error
		store, err = s.settingDomainSvc.ProcessStoreSettingJson(st.Values)
		if err != nil {
			return entity.StoreSetting{}, err
		}
	}

	// 获取默认设置并合并
	defaultStore := entity.DefaultStoreSetting()
	mergedStore := s.settingDomainSvc.MergeWithDefaultStore(*defaultStore, store)

	// 处理经纬度
	if mergedStore.Coordinates != "" {
		latLng := strings.Split(mergedStore.Coordinates, ",")
		if len(latLng) == 2 {
			// 转成float64保留6位小数，然后再转成字符串
			lat, _ := strconv.ParseFloat(latLng[0], 64)
			lng, _ := strconv.ParseFloat(latLng[1], 64)
			mergedStore.Latitude = fmt.Sprintf("%.6f", lat)
			mergedStore.Longitude = fmt.Sprintf("%.6f", lng)
		}
	}

	// 确保数组不为nil
	if len(mergedStore.TimeZoneList) == 0 {
		mergedStore.TimeZoneList = make([]valueobject.TimeZoneItem, 0)
	}
	if len(mergedStore.Language) == 0 {
		mergedStore.Language = make([]valueobject.LanguageItem, 0)
	}

	// 处理图片URL（与旧服务保持一致）
	ginContext := ctx.GetGin()
	if mergedStore.LogoURL != "" && ginContext != nil && ginContext.Request != nil {
		mergedStore.LogoURL = utils.GetBaseURL(ginContext.Request) + mergedStore.LogoURL
	}
	if mergedStore.AvatarURL != "" && ginContext != nil && ginContext.Request != nil {
		mergedStore.AvatarURL = utils.GetBaseURL(ginContext.Request) + mergedStore.AvatarURL
	}

	// 设置 IP 白名单（与旧服务保持一致）
	mergedStore.IPWhiteList = viper.GetString("PAY_SERVICE_IP")

	return mergedStore, nil
}

// GetStoreLanguageList 获取门店语言列表
func (s *SettingAppServiceImpl) GetStoreLanguageList(ctx context.Context) ([]valueobject.LanguageItem, error) {
	storeSetting, err := s.GetStoreSetting(ctx)
	if err != nil {
		return nil, err
	}
	return storeSetting.Language, nil
}

// GetStoreLanguage 获取门店语言
func (s *SettingAppServiceImpl) GetStoreLanguage(ctx context.Context) ([]string, error) {
	storeSetting, err := s.GetStoreSetting(ctx)
	if err != nil {
		return nil, err
	}
	languages := make([]string, 0, len(storeSetting.Language))
	for _, lang := range storeSetting.Language {
		languages = append(languages, lang.Name)
	}
	return languages, nil
}

// GetCashierSetting 获取收银机设置
func (s *SettingAppServiceImpl) GetCashierSetting(ctx context.Context, languageList []valueobject.LanguageItem) (entity.CashierSetting, error) {
	var err error
	if languageList == nil {
		languageList, err = s.GetStoreLanguageList(ctx)
		if err != nil {
			return entity.CashierSetting{}, err
		}
	}

	st := s.getSettingByKey(ctx, "cashier")

	var cashier entity.CashierSetting
	if st.Values != "{}" {
		// 解析设置
		cashier, err = s.settingDomainSvc.ProcessCashierSettingJson(st.Values)
		if err != nil {
			return entity.CashierSetting{}, err
		}
	}

	// 获取默认设置并合并
	defaultCashier := entity.DefaultCashierSetting(languageList)
	result := s.settingDomainSvc.MergeWithDefaultCashier(*defaultCashier, cashier)

	// 确保数组不为 nil（与旧服务保持一致）
	if result.Language == nil {
		result.Language = make([]string, 0)
	}
	if result.LanguageList == nil {
		result.LanguageList = make([]valueobject.LanguageItem, 0)
	}
	if result.Carousel == nil {
		result.Carousel = make([]valueobject.CarouselItem, 0)
	}
	if result.OrderCarousel == nil {
		result.OrderCarousel = make([]valueobject.CarouselItem, 0)
	}
	if result.RemainColor == nil {
		result.RemainColor = make([]string, 0)
	}

	return result, nil
}

// GetCashierLanguage 获取收银机语言
func (s *SettingAppServiceImpl) GetCashierLanguage(ctx context.Context) (entity.LanguageResp, error) {
	cashierSetting, err := s.GetCashierSetting(ctx, nil)
	if err != nil {
		return entity.LanguageResp{}, err
	}

	languageResp := entity.LanguageResp{
		Languages:       cashierSetting.Language,
		LanguageList:    cashierSetting.LanguageList,
		DefaultLanguage: cashierSetting.DefaultLanguage,
	}

	if len(languageResp.Languages) == 0 {
		languageResp.Languages = make([]string, 0)
	}
	if len(languageResp.LanguageList) == 0 {
		languageResp.LanguageList = make([]valueobject.LanguageItem, 0)
	}

	// 过滤有效的语言列表
	validLanguageList := make([]valueobject.LanguageItem, 0)
	languageNames := make([]string, 0)
	for _, item := range languageResp.LanguageList {
		if slices.Contains(languageResp.Languages, item.Name) && !slices.Contains(languageNames, item.Name) {
			validLanguageList = append(validLanguageList, item)
			languageNames = append(languageNames, item.Name)
		}
	}
	languageResp.LanguageList = validLanguageList

	return languageResp, nil
}

// GetCashierAd 获取收银机副屏广告
func (s *SettingAppServiceImpl) GetCashierAd(ctx context.Context) (entity.Ads, error) {
	cashierSetting, err := s.GetCashierSetting(ctx, nil)
	if err != nil {
		return entity.Ads{List: make([]valueobject.CarouselItem, 0)}, err
	}
	return entity.Ads{
		List: cashierSetting.Carousel,
	}, nil
}

// GetCashierBaseSetting 获取收银端设置（与旧服务保持一致，包括设备查询）
func (s *SettingAppServiceImpl) GetCashierBaseSetting(ctx context.Context) (entity.CashierBaseSetting, error) {
	cashierSetting, err := s.GetCashierSetting(ctx, nil)
	if err != nil {
		return entity.CashierBaseSetting{}, err
	}

	businessSetting, err := s.GetBusinessSetting(ctx)
	if err != nil {
		return entity.CashierBaseSetting{}, err
	}

	// 与旧服务保持一致：查询设备信息，如果失败返回系统内部错误
	deviceRepo := memberRepo.NewDeviceRepo(ctx.GetDB())
	_, err = deviceRepo.GetDevice(deviceRepo.WhereSn(ctx.GetDeviceSn()))
	if err != nil {
		return entity.CashierBaseSetting{}, errors.ErrInternal
	}

	// 构建接单设置
	acceptOrder := entity.AcceptOrderSetting{
		IsAutoOrder:    businessSetting.IsAutoOrder,
		AutoOrderLimit: businessSetting.AutoOrderLimit,
		IsAutoVoice:    businessSetting.IsAutoVoice,
	}

	// 构建会员接单设置
	acceptMemberOrder := entity.AcceptMemberOrderSetting{
		IsAutoMemberOrder:      businessSetting.IsAutoMemberOrder,
		AutoMemberOrderLimit:   businessSetting.AutoMemberOrderLimit,
		IsAutoVoiceMemberOrder: businessSetting.IsAutoVoiceMemberOrder,
	}

	// 构建系统设置（简化版本）
	system := entity.SystemSetting{
		IsShowAssistantSoldOut: cashierSetting.IsShowAssistantSoldOut,
		IsShowScanSoldOut:      cashierSetting.IsShowScanSoldOut,
		MenuShowSoldOut:        cashierSetting.MenuShowSoldOut,
		MemberShowSoldOut:      cashierSetting.MemberShowSoldOut,
		DishCardStyle:          businessSetting.DishCardStyle,
	}

	return entity.CashierBaseSetting{
		AcceptOrder:       acceptOrder,
		AcceptMemberOrder: acceptMemberOrder,
		System:            system,
		UsbPrinter:        entity.UsbPrinterList{}, // 暂时为空
	}, nil
}

// GetPrinterSetting 获取打印机设置
func (s *SettingAppServiceImpl) GetPrinterSetting(ctx context.Context, languageList []valueobject.LanguageItem) (entity.PrinterSetting, error) {
	var err error
	if languageList == nil {
		languageList, err = s.GetStoreLanguageList(ctx)
		if err != nil {
			return entity.PrinterSetting{}, err
		}
	}

	st := s.getSettingByKey(ctx, "printer")

	var printer entity.PrinterSetting
	if st.Values != "{}" {
		// 解析设置
		printer, err = s.settingDomainSvc.ProcessPrinterSettingJson(st.Values)
		if err != nil {
			return entity.PrinterSetting{}, err
		}
	}

	// 获取默认设置并合并
	defaultPrinter := entity.DefaultPrinterSetting(languageList)
	result := s.settingDomainSvc.MergeWithDefaultPrinter(*defaultPrinter, printer)

	// 与旧服务保持一致，Language 字段为空数组
	if len(result.Language) > 0 {
		result.Language = make([]string, 0)
	}

	return result, nil
}

// GetPrinterInfo 获取打印机信息（与旧服务保持一致）
func (s *SettingAppServiceImpl) GetPrinterInfo(ctx context.Context, printerSetting entity.PrinterSetting, deviceSn string) (entity.PrinterInfo, error) {
	// 与旧服务保持一致的默认返回值
	return entity.PrinterInfo{
		PrinterType:            "",                                // 打印机类型
		PrinterUuid:            0,                                 // 打印机UUID
		Copies:                 1,                                 // 份数
		PrinterConfig:          "",                                // 打印机配置
		IsCashierPrinter:       false,                             // 是否收银机自带打印机
		IsCashierOpen:          printerSetting.CashierOpen == "1", // 收银机是否开启
		PrinterCashierDeviceSn: "",                                // 收银机设备SN
		IsUsbPrinter:           false,                             // 是否USB打印机
		PrintMethod:            0,                                 // 打印方式
		PrinterSn:              "",                                // 打印机SN
		PrinterWidth:           80,                                // 打印机宽度，默认80mm
		EnableStatusCheck:      0,                                 // 是否启用状态检查
		EnableSound:            0,                                 // 是否启用打印提示音
		PrintSpeed:             2,                                 // 打印速度，默认稳定模式
	}, nil
}

// GetBusinessSetting 获取门店业务设置
func (s *SettingAppServiceImpl) GetBusinessSetting(ctx context.Context) (entity.BusinessSetting, error) {
	st := s.getSettingByKey(ctx, "business")

	var business entity.BusinessSetting
	if st.Values != "{}" {
		// 解析设置
		var err error
		business, err = s.settingDomainSvc.ProcessBusinessSettingJson(st.Values)
		if err != nil {
			return entity.BusinessSetting{}, err
		}
	}

	// 获取默认设置并合并
	defaultBusiness := entity.DefaultBusinessSetting()
	return s.settingDomainSvc.MergeWithDefaultBusiness(*defaultBusiness, business), nil
}

// GetShopBusinessSetting 获取商家业务设置（与旧服务保持一致，从数据库查询统计数据）
func (s *SettingAppServiceImpl) GetShopBusinessSetting(ctx context.Context) (entity.ShopBusinessSetting, error) {
	business, err := s.GetBusinessSetting(ctx)
	if err != nil {
		return entity.ShopBusinessSetting{}, err
	}

	db := ctx.GetDB()
	if db == nil {
		return entity.ShopBusinessSetting{BusinessSetting: business}, nil
	}

	// 查询各类统计数量（与旧服务保持一致）
	var freeReasonCount, returnFoodReasonCount, orderRemarkCount, orderItemRemarkCount, orderSourceCount, nationalityCount int64
	db.Model(&model.FreeReason{}).Where("delete_time = 0").Count(&freeReasonCount)
	db.Model(&model.ReturnFoodReason{}).Where("delete_time = 0").Count(&returnFoodReasonCount)
	db.Model(&model.OrderRemark{}).Where("delete_time = 0").Count(&orderRemarkCount)
	db.Model(&model.OrderItemRemark{}).Where("delete_time = 0").Count(&orderItemRemarkCount)
	db.Model(&model.OrderSource{}).Where("delete_time = 0").Count(&orderSourceCount)
	db.Model(&model.Nationality{}).Where("delete_time = 0").Count(&nationalityCount)

	// 总部业务设置（与旧服务保持一致）
	var headquarterRequiredParentCompanyApproval, headquarterViaParentCompanyWarehouse string
	companySetting, _ := s.GetCompanySetting(ctx)
	if companySetting.HeadquarterUuid > 0 {
		// 需要获取总部的业务设置
		// 这里简化处理，实际需要切换到总部数据库查询
		headquarterRequiredParentCompanyApproval = ""
		headquarterViaParentCompanyWarehouse = ""
	}

	return entity.ShopBusinessSetting{
		BusinessSetting:                          business,
		FreeReasonCount:                          int(freeReasonCount),
		ReturnFoodReasonCount:                    int(returnFoodReasonCount),
		OrderRemarkCount:                         int(orderRemarkCount),
		OrderItemRemarkCount:                     int(orderItemRemarkCount),
		HeadquarterRequiredParentCompanyApproval: headquarterRequiredParentCompanyApproval,
		HeadquarterViaParentCompanyWarehouse:     headquarterViaParentCompanyWarehouse,
		OrderSourceCount:                         int(orderSourceCount),
		NationalityCount:                         int(nationalityCount),
	}, nil
}

// GetBuffetSetting 获取自助餐设置
func (s *SettingAppServiceImpl) GetBuffetSetting(ctx context.Context, companySetting model.CompanySetting) (entity.BuffetResp, error) {
	st := s.getSettingByKey(ctx, "buffet")

	var buffet entity.Buffet
	if st.Values != "{}" {
		// 解析设置
		err := s.settingDomainSvc.ProcessBuffetSettingJson(st.Values, &buffet)
		if err != nil {
			return entity.BuffetResp{}, err
		}
	}

	// 根据公司设置调整is_open
	if companySetting.IsOpenBuffet == 0 {
		buffet.IsOpen = "0"
	}

	// 获取默认设置并合并
	defaultBuffet := entity.DefaultBuffetSetting()
	mergedBuffet := s.settingDomainSvc.MergeWithDefaultBuffet(*defaultBuffet, buffet)

	// 转换时间格式（分钟转秒）
	tabletEndTime, _ := strconv.Atoi(mergedBuffet.TabletEndTime)

	return entity.BuffetResp{
		IsOpen:                   mergedBuffet.IsOpen,
		TabletEndTime:            tabletEndTime * 60,
		IsRemainContinue:         mergedBuffet.IsRemainContinue,
		RemainContinueTime:       mergedBuffet.RemainContinueTime,
		RemainContinueNoticeTime: mergedBuffet.RemainContinueNoticeTime,
		IsBuyContinue:            mergedBuffet.IsBuyContinue,
		IsAddClock:               mergedBuffet.IsAddClock,
		IsBuffetDiscount:         mergedBuffet.IsBuffetDiscount,
		IsShowNonBuffetProduct:   mergedBuffet.IsShowNonBuffetProduct,
		AddClock:                 mergedBuffet.AddClock,
	}, nil
}

// GetTabletSetting 获取平板端设置
func (s *SettingAppServiceImpl) GetTabletSetting(ctx context.Context, languageList []valueobject.LanguageItem) (entity.TabletSetting, error) {
	var err error
	if languageList == nil {
		languageList, err = s.GetStoreLanguageList(ctx)
		if err != nil {
			return entity.TabletSetting{}, err
		}
	}

	st := s.getSettingByKey(ctx, "tablet")

	var tablet entity.TabletSetting
	if st.Values != "{}" {
		// 解析设置
		tablet, err = s.settingDomainSvc.ProcessTabletSettingJson(st.Values)
		if err != nil {
			return entity.TabletSetting{}, err
		}
	}

	// 获取默认设置并合并
	defaultTablet := entity.DefaultTabletSetting(languageList)
	mergedTablet := s.settingDomainSvc.MergeWithDefaultTablet(*defaultTablet, tablet)

	// 设置正确的服务器 IP 和端口（与旧服务保持一致）
	serverIP, serverPort := getIPAndPort()
	mergedTablet.Server.IP = serverIP
	mergedTablet.Server.Port = serverPort

	// 确保数组不为 nil（与旧服务保持一致）
	if mergedTablet.Carousel == nil {
		mergedTablet.Carousel = make([]valueobject.CarouselItem, 0)
	}
	if mergedTablet.LanguageList == nil {
		mergedTablet.LanguageList = make([]valueobject.LanguageItem, 0)
	}
	if mergedTablet.Language == nil {
		mergedTablet.Language = make([]string, 0)
	}

	// 处理语言列表过滤
	validLanguageList := make([]valueobject.LanguageItem, 0)
	languageNames := make([]string, 0)
	for _, item := range mergedTablet.LanguageList {
		if slices.Contains(mergedTablet.Language, item.Name) && !slices.Contains(languageNames, item.Name) {
			validLanguageList = append(validLanguageList, item)
			languageNames = append(languageNames, item.Name)
		}
	}
	mergedTablet.LanguageList = validLanguageList

	return mergedTablet, nil
}

// GetPaymentSetting 获取门店-支付方式设置
func (s *SettingAppServiceImpl) GetPaymentSetting(ctx context.Context, companySetting model.CompanySetting) (entity.PaymentSetting, error) {
	st := s.getSettingByKey(ctx, "payment")

	var payment entity.PaymentSetting
	if st.Values != "{}" {
		// 解析设置
		err := json.Unmarshal([]byte(st.Values), &payment)
		if err != nil {
			return entity.PaymentSetting{}, err
		}
	}

	// 会员关闭时 门店管理 支付方式 余额这个开关要关了
	if companySetting.IsOpenMember == 0 {
		payment.IsBalance = "0"
	}

	// 获取默认设置并合并
	defaultPayment := entity.DefaultPaymentSetting()
	result := s.settingDomainSvc.MergeWithDefaultPayment(*defaultPayment, payment)

	return result, nil
}

// GetCurrencySetting 获取货币单位设置
func (s *SettingAppServiceImpl) GetCurrencySetting(ctx context.Context) (entity.CurrencySetting, error) {
	st := s.getSettingByKey(ctx, "currency")

	var currency entity.CurrencySetting
	if st.Values != "{}" {
		// 解析设置
		err := json.Unmarshal([]byte(st.Values), &currency)
		if err != nil {
			return entity.CurrencySetting{}, err
		}
	}

	// 获取默认设置并合并
	defaultCurrency := entity.DefaultCurrencySetting()
	result := s.settingDomainSvc.MergeWithDefaultCurrency(*defaultCurrency, currency)

	return result, nil
}

// GetCompanySetting 获取公司设置
func (s *SettingAppServiceImpl) GetCompanySetting(ctx context.Context) (model.CompanySetting, error) {
	db := ctx.GetDB()
	if db == nil {
		return model.CompanySetting{}, fmt.Errorf("db not found")
	}
	// 调用仓储层获取公司设置
	companySetting, err := s.settingRepo.GetCompanySetting(ctx)
	if err != nil {
		return model.CompanySetting{}, err
	}
	return *companySetting, nil
}

// GetKitchenSetting 获取厨显端设置
func (s *SettingAppServiceImpl) GetKitchenSetting(ctx context.Context, companySetting model.CompanySetting, languageList []valueobject.LanguageItem) (entity.KitchenSetting, error) {
	var err error
	if languageList == nil {
		languageList, err = s.GetStoreLanguageList(ctx)
		if err != nil {
			return entity.KitchenSetting{}, err
		}
	}

	st := s.getSettingByKey(ctx, "kitchen")

	var kitchen entity.KitchenSetting
	if st.Values != "{}" {
		// 解析设置
		err = json.Unmarshal([]byte(st.Values), &kitchen)
		if err != nil {
			return entity.KitchenSetting{}, err
		}
	}

	defaultKitchen := entity.DefaultKitchenSetting(languageList)

	// 语言 不需要合并
	defaultKitchen.Language = nil

	// 使用copier进行合并
	err = copier.CopyWithOption(&defaultKitchen, kitchen, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	if err != nil {
		return entity.KitchenSetting{}, err
	}

	// 总权限 - 不开启厨显
	if companySetting.IsOpenKitchenKds == 0 {
		kitchen.IsOpen = "0"
	}

	if len(defaultKitchen.WaitColor) == 0 {
		defaultKitchen.WaitColor = make([]string, 0)
	}
	if len(defaultKitchen.WaitTimeColorRanges) == 0 {
		defaultKitchen.WaitTimeColorRanges = make([]entity.WaitTimeColorRange, 0)
	}
	if len(defaultKitchen.LanguageList) == 0 {
		defaultKitchen.LanguageList = make([]valueobject.LanguageItem, 0)
	}
	if len(defaultKitchen.Language) == 0 {
		defaultKitchen.Language = make([]string, 0)
	}

	// 转换旧格式到新格式（如果只有旧格式数据）
	if len(defaultKitchen.WaitTimeColorRanges) == 0 && len(defaultKitchen.WaitColor) > 0 {
		defaultKitchen.WaitTimeColorRanges = s.convertFromOldFormat(defaultKitchen.WaitColor)
	}

	// 设置正确的服务器 IP 和端口（与旧服务保持一致）
	serverIP, serverPort := getIPAndPort()
	defaultKitchen.Server.IP = serverIP
	defaultKitchen.Server.Port = serverPort

	validLanguageList := make([]valueobject.LanguageItem, 0)
	languageNames := make([]string, 0)
	for _, item := range defaultKitchen.LanguageList {
		if slices.Contains(defaultKitchen.Language, item.Name) && !slices.Contains(languageNames, item.Name) {
			validLanguageList = append(validLanguageList, item)
			languageNames = append(languageNames, item.Name)
		}
	}
	defaultKitchen.LanguageList = validLanguageList

	return *defaultKitchen, nil
}

// GetH5Setting 获取扫码H5设置
func (s *SettingAppServiceImpl) GetH5Setting(ctx context.Context, languageList []valueobject.LanguageItem) (entity.H5Setting, error) {
	var err error
	if languageList == nil {
		languageList, err = s.GetStoreLanguageList(ctx)
		if err != nil {
			return entity.H5Setting{}, err
		}
	}

	st := s.getSettingByKey(ctx, "h5")

	var h5 entity.H5Setting
	if st.Values != "{}" {
		// 处理JSON格式转换
		values := st.Values
		if strings.Contains(values, "\"buffet_order_limit\":[]") {
			values = strings.Replace(values, "\"buffet_order_limit\":[]", "\"buffet_order_limit\":{}", -1)
		}
		if strings.Contains(values, "\"order_limit\":[]") {
			values = strings.Replace(values, "\"order_limit\":[]", "\"order_limit\":{}", -1)
		}

		// 解析json字符串为map进行处理
		var jsonMap map[string]interface{}
		err = json.Unmarshal([]byte(values), &jsonMap)
		if err != nil {
			return entity.H5Setting{}, err
		}

		// 处理 isShowScanSoldOut
		if isShowScanSoldOut, ok := jsonMap["is_show_scan_sold_out"]; ok {
			if numVal, ok := isShowScanSoldOut.(float64); ok {
				jsonMap["is_show_scan_sold_out"] = int(numVal)
			} else if strVal, ok := isShowScanSoldOut.(string); ok {
				if intVal, err := strconv.Atoi(strVal); err == nil {
					jsonMap["is_show_scan_sold_out"] = intVal
				}
			}
		}

		// 重新序列化为JSON
		modifiedJSON, err := json.Marshal(jsonMap)
		if err != nil {
			return entity.H5Setting{}, err
		}

		// 解析json字符串
		err = json.Unmarshal(modifiedJSON, &h5)
		if err != nil {
			return entity.H5Setting{}, err
		}
	}

	defaultH5 := entity.DefaultH5Setting(languageList)
	err = copier.CopyWithOption(&defaultH5, h5, copier.Option{IgnoreEmpty: true})
	if err != nil {
		return entity.H5Setting{}, err
	}

	// 如果设置了 is_show_scan_sold_out，则读取解析后的数据，否则读取默认设置
	cashierSetting, err := s.GetCashierSetting(ctx, nil)
	if err == nil {
		defaultH5.IsShowScanSoldOut = cashierSetting.IsShowScanSoldOut
	}

	if len(defaultH5.LanguageList) == 0 {
		defaultH5.LanguageList = make([]valueobject.LanguageItem, 0)
	}

	return *defaultH5, nil
}

// GetKioskSetting 获取自助点餐机设置
func (s *SettingAppServiceImpl) GetKioskSetting(ctx context.Context) (entity.KioskSetting, error) {
	// 获取语言列表（与旧服务保持一致）
	languageList, err := s.GetStoreLanguageList(ctx)
	if err != nil {
		return entity.KioskSetting{}, err
	}

	st := s.getSettingByKey(ctx, "kiosk")

	// 如果没有设置值，返回默认值
	if st.Values == "" || st.Values == "{}" {
		return *entity.DefaultKioskSetting(languageList), nil
	}

	var kiosk entity.KioskSetting
	err = json.Unmarshal([]byte(st.Values), &kiosk)
	if err != nil {
		return entity.KioskSetting{}, err
	}

	// 合并默认值（与旧服务保持一致）
	defaultKiosk := entity.DefaultKioskSetting(languageList)
	if kiosk.AdvancedPassword == "" {
		kiosk.AdvancedPassword = defaultKiosk.AdvancedPassword
	}
	if len(kiosk.Language) == 0 {
		kiosk.Language = defaultKiosk.Language
	}
	if kiosk.DefaultLanguage == "" {
		kiosk.DefaultLanguage = defaultKiosk.DefaultLanguage
	}
	if kiosk.CallWaiterEnabled == 0 {
		kiosk.CallWaiterEnabled = defaultKiosk.CallWaiterEnabled
	}

	// 确保数组不为 nil（与旧服务保持一致）
	if kiosk.Carousel == nil {
		kiosk.Carousel = make([]valueobject.CarouselItem, 0)
	}
	if kiosk.Language == nil {
		kiosk.Language = make([]string, 0)
	}
	if kiosk.LanguageList == nil {
		kiosk.LanguageList = make([]valueobject.LanguageItem, 0)
	}

	return kiosk, nil
}

// GetAssistantSetting 获取点餐助手设置
func (s *SettingAppServiceImpl) GetAssistantSetting(ctx context.Context, languageList []valueobject.LanguageItem) (entity.AssistantSetting, error) {
	var err error
	if languageList == nil {
		languageList, err = s.GetStoreLanguageList(ctx)
		if err != nil {
			return entity.AssistantSetting{}, err
		}
	}

	st := s.getSettingByKey(ctx, "assistant")

	var assistant entity.AssistantSetting
	if st.Values != "{}" {
		// 解析json字符串为map进行处理
		var jsonMap map[string]interface{}
		err = json.Unmarshal([]byte(st.Values), &jsonMap)
		if err != nil {
			return entity.AssistantSetting{}, err
		}

		// 处理 isShowAssistantSoldOut
		if isShowAssistantSoldOut, ok := jsonMap["is_show_assistant_sold_out"]; ok {
			if numVal, ok := isShowAssistantSoldOut.(float64); ok {
				jsonMap["is_show_assistant_sold_out"] = int(numVal)
			} else if strVal, ok := isShowAssistantSoldOut.(string); ok {
				if intVal, err := strconv.Atoi(strVal); err == nil {
					jsonMap["is_show_assistant_sold_out"] = intVal
				}
			}
		}

		// 重新序列化为JSON
		modifiedJSON, err := json.Marshal(jsonMap)
		if err != nil {
			return entity.AssistantSetting{}, err
		}

		err = json.Unmarshal(modifiedJSON, &assistant)
		if err != nil {
			return entity.AssistantSetting{}, err
		}
	}

	if len(assistant.LanguageList) == 0 {
		assistant.LanguageList = nil
	}

	// 获取服务器 IP 和端口
	serverIP, serverPort := getIPAndPort()
	defaultAssistant := entity.DefaultAssistantSetting(languageList, serverIP, serverPort)
	err = copier.CopyWithOption(&defaultAssistant, assistant, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	if err != nil {
		return entity.AssistantSetting{}, err
	}

	// 如果设置了 is_show_assistant_sold_out，则读取解析后的数据，否则读取默认设置
	cashierSetting, err := s.GetCashierSetting(ctx, nil)
	if err == nil {
		defaultAssistant.IsShowAssistantSoldOut = cashierSetting.IsShowAssistantSoldOut
	}

	// 确保 RemainColor 不为 nil
	if defaultAssistant.RemainColor == nil {
		defaultAssistant.RemainColor = []string{"#E50028", "#F2A000"}
	}

	return *defaultAssistant, nil
}

// GetPointsSetting 获取积分设置（与旧服务保持一致，包括会员等级信息）
func (s *SettingAppServiceImpl) GetPointsSetting(ctx context.Context) (entity.PointsSetting, error) {
	st := s.getSettingByKey(ctx, "points")

	var points entity.PointsSetting
	if st.Values != "{}" {
		// 解析设置
		var err error
		points, err = s.settingDomainSvc.ProcessPointsSettingJson(st.Values)
		if err != nil {
			return entity.PointsSetting{}, err
		}
	}

	// 获取默认设置并合并
	defaultPoints := entity.DefaultPointsSetting()

	// 与旧服务保持一致：处理会员等级信息
	memberRepo := memberRepo.NewMemberRepo(ctx.GetDB())
	memberLevels := memberRepo.GetMemberLevels()
	var rateMemberLevels []entity.MemberLevelItem
	var quantityMemberLevels []entity.MemberLevelItem
	for _, level := range memberLevels {
		rateMemberLevels = append(rateMemberLevels, entity.MemberLevelItem{
			Uuid:  level.Uuid,
			Name:  level.Name,
			Value: level.PointsRate,
		})
		quantityMemberLevels = append(quantityMemberLevels, entity.MemberLevelItem{
			Uuid:  level.Uuid,
			Name:  level.Name,
			Value: level.PointsQuantity,
		})
	}

	// 更新默认设置中的会员等级信息
	defaultPoints.ShoppingGiftRules[0].MemberLevels = rateMemberLevels     // payment_amount 规则使用积分比例
	defaultPoints.ShoppingGiftRules[1].MemberLevels = quantityMemberLevels // desk 规则使用积分数量

	result := s.settingDomainSvc.MergeWithDefaultPoints(*defaultPoints, points)

	// 如果解析的设置中有会员等级信息，也需要更新
	if len(points.ShoppingGiftRules) >= 2 {
		if len(points.ShoppingGiftRules[0].MemberLevels) == 0 {
			result.ShoppingGiftRules[0].MemberLevels = rateMemberLevels
		}
		if len(points.ShoppingGiftRules[1].MemberLevels) == 0 {
			result.ShoppingGiftRules[1].MemberLevels = quantityMemberLevels
		}
	}

	return result, nil
}

// GetCloudBasicSetting 获取云端基础信息（与旧服务保持一致）
func (s *SettingAppServiceImpl) GetCloudBasicSetting(ctx context.Context) (entity.CloudBasicSetting, error) {
	st := s.getSettingByKey(ctx, "cloud_basic")

	var cloudBasic entity.CloudBasicSetting
	if st.Values != "{}" {
		err := json.Unmarshal([]byte(st.Values), &cloudBasic)
		if err != nil {
			return entity.CloudBasicSetting{}, err
		}
	}

	// 如果没有设置数据或数据为空，使用与旧服务完全一致的默认值
	if st.Values == "{}" || cloudBasic.BrandName == "" {
		// 与旧服务保持一致的默认值
		cloudBasic = entity.CloudBasicSetting{
			BrandName:          "Shop",
			BrandLogo:          "/image/logo/ttpos_64_64.png",
			BrandLogoLong:      "/image/logo/ttpos_146_40.png",
			BrowserLogo:        "/image/logo/ttpos_64_64.png",
			BrowserTitle:       "Shop",
			ExpirationReminder: 0, // 与旧服务保持一致
		}
	}

	// 处理图片URL（与旧服务保持一致）
	ginContext := ctx.GetGin()
	if cloudBasic.BrandLogo != "" && ginContext != nil && ginContext.Request != nil {
		cloudBasic.BrandLogo = utils.AddImageDomain(utils.RemoveDomain(cloudBasic.BrandLogo), utils.GetBaseURL(ginContext.Request), true)
	}
	if cloudBasic.BrandLogoLong != "" && ginContext != nil && ginContext.Request != nil {
		cloudBasic.BrandLogoLong = utils.AddImageDomain(utils.RemoveDomain(cloudBasic.BrandLogoLong), utils.GetBaseURL(ginContext.Request), true)
	}
	if cloudBasic.BrowserLogo != "" && ginContext != nil && ginContext.Request != nil {
		cloudBasic.BrowserLogo = utils.AddImageDomain(utils.RemoveDomain(cloudBasic.BrowserLogo), utils.GetBaseURL(ginContext.Request), true)
	}

	return cloudBasic, nil
}

// GetServiceFeeSetting 获取服务费设置
func (s *SettingAppServiceImpl) GetServiceFeeSetting(ctx context.Context) (entity.ServiceCharge, error) {
	st := s.getSettingByKey(ctx, "service_charge")

	var serviceFee entity.ServiceCharge
	if st.Values != "{}" {
		// 修改类型
		newVal, err := s.settingDomainSvc.ConvertServiceFeeFormat(st.Values)
		if err != nil {
			return entity.ServiceCharge{}, err
		}
		st.Values = string(newVal)
		err = json.Unmarshal([]byte(st.Values), &serviceFee)
		if err != nil {
			return entity.ServiceCharge{}, err
		}
	}

	if serviceFee.IsOpen == "0" {
		serviceFee.IsOpen = "0"
	}
	defaultServiceFee := entity.DefaultServiceCharge()
	result := s.settingDomainSvc.MergeWithDefaultServiceCharge(*defaultServiceFee, serviceFee)

	return result, nil
}

// GetTaxRateSetting 获取税率设置
func (s *SettingAppServiceImpl) GetTaxRateSetting(ctx context.Context) (entity.TaxRate, error) {
	st := s.getSettingByKey(ctx, "tax_rate")

	var taxRate entity.TaxRate
	if st.Values != "{}" {
		err := json.Unmarshal([]byte(st.Values), &taxRate)
		if err != nil {
			return entity.TaxRate{}, err
		}
	}

	if taxRate.IsOpen == "0" {
		taxRate.IsOpen = "0"
	}
	defaultTaxRate := entity.DefaultTaxRate()
	result := s.settingDomainSvc.MergeWithDefaultTaxRate(*defaultTaxRate, taxRate)

	if len(result.AddTaxCategory) == 0 {
		result.AddTaxCategory = make([]entity.AddTaxCategoryItem, 0)
	}

	return result, nil
}

// VerifyPassword 收银机验证密码
func (s *SettingAppServiceImpl) VerifyPassword(ctx context.Context, source, typ, password string) bool {
	passwordMap := make(map[string]string)
	switch source {
	case "cashier":
		cashierSetting, err := s.GetCashierSetting(ctx, nil)
		if err != nil {
			return false
		}
		passwordMap = map[string]string{
			"cash_box": cashierSetting.CashierPassword,
			"advanced": cashierSetting.AdvancedPassword,
			"lock":     cashierSetting.LockPassword,
		}
	case "assistant":
		assistantSetting, err := s.GetAssistantSetting(ctx, nil)
		if err != nil {
			return false
		}
		passwordMap = map[string]string{
			"advanced": assistantSetting.AdvancedPassword,
		}
	case "tablet":
		tabletSetting, err := s.GetTabletSetting(ctx, nil)
		if err != nil {
			return false
		}
		passwordMap = map[string]string{
			"advanced": tabletSetting.AdvancedPassword,
		}
	case "kitchen":
		companySetting, err := s.GetCompanySetting(ctx)
		if err != nil {
			return false
		}
		kitchenSetting, err := s.GetKitchenSetting(ctx, companySetting, nil)
		if err != nil {
			return false
		}
		passwordMap = map[string]string{
			"advanced": kitchenSetting.AdvancedPassword,
		}
	}
	if truePassword, exists := passwordMap[typ]; exists {
		return password == truePassword
	}
	return false
}

// VerifyAdvancedPassword 验证高级密码
func (s *SettingAppServiceImpl) VerifyAdvancedPassword(ctx context.Context, password string, options ...interface{}) error {
	// 获取业务设置，检查是否需要密码验证
	businessSetting, err := s.GetBusinessSetting(ctx)
	if err != nil {
		return err
	}

	if businessSetting.IsNeedPassword == "1" {
		if password == "" {
			return fmt.Errorf("请输入确认密码")
		}

		// 根据来源获取对应的设置并验证密码
		source := "cashier" // 默认来源，实际应该从context获取
		var advancedPassword string

		switch source {
		case "cashier":
			cashierSetting, err := s.GetCashierSetting(ctx, nil)
			if err != nil {
				return err
			}
			advancedPassword = cashierSetting.AdvancedPassword
		case "assistant":
			assistantSetting, err := s.GetAssistantSetting(ctx, nil)
			if err != nil {
				return err
			}
			advancedPassword = assistantSetting.AdvancedPassword
		case "kitchen":
			companySetting, err := s.GetCompanySetting(ctx)
			if err != nil {
				return err
			}
			kitchenSetting, err := s.GetKitchenSetting(ctx, companySetting, nil)
			if err != nil {
				return err
			}
			advancedPassword = kitchenSetting.AdvancedPassword
		case "tablet":
			tabletSetting, err := s.GetTabletSetting(ctx, nil)
			if err != nil {
				return err
			}
			advancedPassword = tabletSetting.AdvancedPassword
		default:
			return fmt.Errorf("未知的密码验证来源")
		}

		if password != advancedPassword {
			return fmt.Errorf("确认密码错误")
		}
	}

	return nil
}

// UpdateSetting 更新设置
func (s *SettingAppServiceImpl) UpdateSetting(ctx context.Context, settingKey string, values interface{}) error {
	valueStr := utils.ToJson(values)
	setting := &model.Setting{
		Key:      settingKey,
		Describe: constant.GetSettingDescribe(settingKey),
		Values:   valueStr,
	}

	// 先查找是否存在
	existing, err := s.settingRepo.FindByKey(ctx, settingKey)
	if err != nil {
		return err
	}

	if existing != nil {
		// 更新
		return s.settingRepo.Update(ctx, setting)
	} else {
		// 创建
		return s.settingRepo.Save(ctx, setting)
	}
}

// EditStoreSetting 修改店铺设置
func (s *SettingAppServiceImpl) EditStoreSetting(ctx context.Context, storeSetting entity.UpdateStoreSetting) error {
	// 转换为JSON并更新
	return s.UpdateSetting(ctx, "store", storeSetting)
}

// EditBusinessSetting 修改业务设置
func (s *SettingAppServiceImpl) EditBusinessSetting(ctx context.Context, businessSetting entity.UpdateBusinessSetting) error {
	// 转换为JSON并更新
	return s.UpdateSetting(ctx, "business", businessSetting)
}

// EditCashierSetting 修改收银机设置
func (s *SettingAppServiceImpl) EditCashierSetting(ctx context.Context, cashierSettingReq entity.SaveCashierSettingReq) error {
	// 转换为JSON并更新
	return s.UpdateSetting(ctx, "cashier", cashierSettingReq)
}

// EditKioskSetting 修改自助点餐机设置
func (s *SettingAppServiceImpl) EditKioskSetting(ctx context.Context, kioskSettingReq entity.SaveKioskSettingReq) error {
	// 转换为JSON并更新
	return s.UpdateSetting(ctx, "kiosk", kioskSettingReq)
}

// SaveKitchenSetting 保存厨显设置
func (s *SettingAppServiceImpl) SaveKitchenSetting(ctx context.Context, kitchenSettingReq entity.SaveKitchenSettingReq) error {
	// 转换为JSON并更新
	return s.UpdateSetting(ctx, "kitchen", kitchenSettingReq)
}

// EditSystemSetting 修改系统设置
func (s *SettingAppServiceImpl) EditSystemSetting(ctx context.Context, systemSetting entity.UpdateSystemSetting) error {
	// 更新业务设置中的系统相关字段
	return s.UpdateSetting(ctx, "system", systemSetting)
}

// GetAcceptOrderSetting 获取接单设置
func (s *SettingAppServiceImpl) GetAcceptOrderSetting(ctx context.Context) (*entity.AcceptOrderSetting, error) {
	cashierSetting, err := s.GetCashierSetting(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &entity.AcceptOrderSetting{
		IsAutoOrder:    cashierSetting.IsAutoOrder,
		AutoOrderLimit: cashierSetting.AutoOrderLimit,
		IsAutoVoice:    cashierSetting.IsAutoVoice,
	}, nil
}

// EditAcceptOrderSetting 修改自动接单设置
func (s *SettingAppServiceImpl) EditAcceptOrderSetting(ctx context.Context, orderSetting entity.UpdateAcceptOrderSetting) error {
	// 更新业务设置中的接单相关字段
	return s.UpdateSetting(ctx, "accept_order", orderSetting)
}

// EditAcceptMemberOrderSetting 修改自动接单会员订单设置
func (s *SettingAppServiceImpl) EditAcceptMemberOrderSetting(ctx context.Context, orderSetting entity.UpdateAcceptMemberOrderSetting) error {
	// 更新业务设置中的会员接单相关字段
	return s.UpdateSetting(ctx, "accept_member_order", orderSetting)
}

// GetMenuQrcode 获取电子菜单二维码
func (s *SettingAppServiceImpl) GetMenuQrcode(ctx context.Context) (string, error) {
	businessSetting, err := s.GetBusinessSetting(ctx)
	if err != nil {
		return "", err
	}
	return viper.GetString("MENU_BASE_URL") + "/home?token=" + s.getMenuQrcodeToken(ctx, businessSetting), nil
}

// GetPaymentMethodList 获取支付方式列表（与旧服务保持一致，从数据库读取）
func (s *SettingAppServiceImpl) GetPaymentMethodList(ctx context.Context) entity.PaymentMethodListResp {
	// 从数据库查询支付方式列表
	list, err := s.settingRepo.GetPaymentMethodList(ctx)
	if err != nil || len(list) == 0 {
		// 如果查询失败或没有数据，返回空列表（与旧服务保持一致）
		return entity.PaymentMethodListResp{
			List: make([]entity.PaymentMethod, 0),
		}
	}

	// 转换为响应格式
	result := make([]entity.PaymentMethod, 0, len(list))
	for _, pm := range list {
		result = append(result, entity.PaymentMethod{
			Uuid:        pm.Uuid,
			Name:        pm.Name,
			PaymentName: pm.PaymentName,
		})
	}

	return entity.PaymentMethodListResp{List: result}
}

// GetDataManageSetting 获取数据管理设置
func (s *SettingAppServiceImpl) GetDataManageSetting(ctx context.Context) entity.DataManageSetting {
	// 实际实现需要查询数据管理配置
	return entity.DataManageSetting{
		IsEnableDataManage: false, // 是否开启数据管理
	}
}

// CheckUpdate 检查更新（与旧服务保持一致）
func (s *SettingAppServiceImpl) CheckUpdate(ctx context.Context, appType int, brand, language string) (entity.UpdateInfo, error) {
	// 与旧服务保持一致，返回不支持更新的错误
	return entity.UpdateInfo{}, fmt.Errorf("当前平台暂不支持应用内更新")
}

// SymbolPosition 根据货币符号位置返回字符串
func (s *SettingAppServiceImpl) SymbolPosition(ctx context.Context, price float64) string {
	currencySetting, err := s.GetCurrencySetting(ctx)
	if err != nil {
		return fmt.Sprintf("%.2f", price)
	}

	if currencySetting.UnitPosition == "0" {
		return currencySetting.Unit + " " + fmt.Sprintf("%.2f", price)
	} else {
		return fmt.Sprintf("%.2f", price) + " " + currencySetting.Unit
	}
}

func (s *SettingAppServiceImpl) convertFromOldFormat(oldFormat []string) []entity.WaitTimeColorRange {
	var result []entity.WaitTimeColorRange
	result = append(result, entity.WaitTimeColorRange{Minute: "0", Color: "#100A05"}) // 第一区间固定黑色

	// 旧格式：["red", "yellow"] 或 ["yellow", "red"]
	// 第一个元素对应第二区间，第二个元素对应第三区间
	colorMap := map[string]string{
		"red":    "#E50028",
		"yellow": "#FFBE00",
	}

	for i, item := range oldFormat {
		if i >= 2 {
			break // 最多两个元素
		}
		minute := "10"
		if i == 1 {
			minute = "20"
		}
		color := "#FFBE00" // 默认黄色
		if mappedColor, exists := colorMap[item]; exists {
			color = mappedColor
		}
		result = append(result, entity.WaitTimeColorRange{Minute: minute, Color: color})
	}

	return result
}

func (s *SettingAppServiceImpl) getMenuQrcodeToken(ctx context.Context, businessSetting entity.BusinessSetting) string {
	type Qrcode struct {
		CompanyUuid uint64 `json:"a"`
		Qrcode      string `json:"q"`
	}
	qrcode := Qrcode{
		CompanyUuid: ctx.(*pkgctx.ContextImpl).GetCompanyUuid(), // 使用正确的context方法
		Qrcode:      businessSetting.QrCode,
	}
	qrcodeString := utils.ToJson(qrcode)
	hash := md5.Sum([]byte(qrcodeString))
	token := fmt.Sprintf("%x.%s", hash, base64.StdEncoding.EncodeToString([]byte(qrcodeString)))
	return token
}

// GetIPAndPort 获取IP和端口
func (s *SettingAppServiceImpl) getIPAndPort() (string, string) {
	// 默认值
	return "127.0.0.1", "8080"
}

// fromCache 从缓存获取所有设置
func (s *SettingAppServiceImpl) fromCache(ctx context.Context) ([]model.Setting, error) {
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
	allSettings, err := s.settingRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为[]model.Setting
	for _, setting := range allSettings {
		settings = append(settings, *setting)
	}

	data, _ := json.Marshal(settings)
	s.cache.Set(cacheKey, string(data), 0)

	return settings, nil
}

// getSettingByKey 根据key获取设置（带缓存）
func (s *SettingAppServiceImpl) getSettingByKey(ctx context.Context, key string) model.Setting {
	defaultSetting := model.Setting{Key: key, Values: "{}"}
	if key == "cloud_basic" { // 平台设置
		if data, exists := s.cache.Get(s.cloudCacheKey); exists { // 从缓存读取
			if dataValue, isString := data.(string); isString {
				return model.Setting{Key: key, Values: dataValue}
			}
		} else { // 从远程获取
			type Base struct {
				Code int    `json:"code"`
				Msg  string `json:"msg"`
				Data struct {
					Base struct {
						BrandName          string `json:"brand_name"`
						BrandLogo          string `json:"brand_logo"`
						BrandLogoLong      string `json:"brand_logo_long"`
						BrowserLogo        string `json:"browser_logo"`
						BrowserTitle       string `json:"browser_title"`
						ExpirationReminder int    `json:"expiration_reminder"`
					} `json:"base"`
				} `json:"data"`
			}
			// 这里简化实现，实际需要调用HTTP请求
			// 暂时返回默认值
			var basicCloudSetting entity.CloudBasicSetting
			dataBytes, _ := json.Marshal(basicCloudSetting)
			s.cache.Set(s.cloudCacheKey, string(dataBytes), 0)
			return model.Setting{Key: key, Values: string(dataBytes)}
		}
	}

	allSettings, _ := s.fromCache(ctx)
	for _, set := range allSettings {
		if set.Key == key {
			return set
		}
	}
	return defaultSetting
}
