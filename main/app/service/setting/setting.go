package setting

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/errors"
	errors2 "ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/websocket"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/nahid/gohttp"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
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
	GetStoreLanguage(ctx context.Context) ([]string, error)                                                                               // 获取商家语言
	GetPrinterSetting(ctx context.Context, languageList []dto.LanguageItem) (setting.Printer, error)                                      // 获取打印机设置
	GetPrinterInfo(ctx context.Context, printerSetting setting.Printer, deviceId string) (setting.PrinterInfo, error)                     // 获取打印机信息
	GetCashierSetting(ctx context.Context, languageList []dto.LanguageItem) (setting.Cashier, error)                                      // 获取收银机设置
	GetCloudBasicSetting(ctx context.Context) (setting.CloudBasic, error)                                                                 // 获取云端基础信息
	GetAssistantSetting(ctx context.Context, languageList []dto.LanguageItem) (setting.Assistant, error)                                  // 获取点餐助手设置
	GetPointsSetting(ctx context.Context) (setting.Points, error)                                                                         // 获取积分设置
	GetKitchenSetting(ctx context.Context, companySetting model.CompanySetting, languageList []dto.LanguageItem) (setting.Kitchen, error) // 获取厨显端设置
	GetH5Setting(ctx context.Context, languageList []dto.LanguageItem) (setting.H5, error)                                                // 获取扫码H5设置
	GetBusinessSetting(ctx context.Context) (setting.Business, error)                                                                     // 获取门店业务设置
	GetBuffetSetting(ctx context.Context, companySetting model.CompanySetting) (setting.BuffetResp, error)                                // 获取自助餐设置
	GetTabletSetting(ctx context.Context, languageList []dto.LanguageItem) (setting.Tablet, error)                                        // 获取平板端设置
	GetCurrencySetting(ctx context.Context) (setting.Currency, error)                                                                     // 获取货币单位设置
	GetCompanySetting(ctx context.Context) (model.CompanySetting, error)                                                                  // 获取公司设置
	GetPaymentSetting(ctx context.Context, companySetting model.CompanySetting) (setting.Payment, error)                                  // 获取门店-支付方式设置
	GetCashierLanguage(c context.Context) (resp.LanguageResp, error)                                                                      // 获取收银机语言
	GetCashierAd(ctx context.Context) (resp.Ads, error)                                                                                   // 获取收银机副屏广告
	GetServiceFeeSetting(ctx context.Context) (setting.ServiceCharge, error)                                                              // 获取服务费设置
	GetTaxRateSetting(ctx context.Context) (setting.TaxRate, error)                                                                       // 获取税率设置
	VerifyPassword(ctx context.Context, source string, typ string, password string) bool                                                  // 收银机验证密码
	UpdateSetting(ctx context.Context, settingKey string, values any) error                                                               // 更新设置
	VerifyAdvancedPassword(ctx context.Context, password string, options ...func(option *VerifyAdvancedPasswordOption)) error             // 验证高级密码
	CheckUpdate(ctx context.Context, appType int, brand string, language string) (resp.UpdateInfo, error)                                 // 检查更新
	EditAcceptOrderSetting(ctx context.Context, orderSetting req.UpdateAcceptOrderSetting) error                                          // 修改自动接单设置
	EditAcceptMemberOrderSetting(ctx context.Context, orderSetting req.UpdateAcceptMemberOrderSetting) error                              // 修改自动接单会员订单设置
	EditSystemSetting(ctx context.Context, systemSetting req.UpdateSystemSetting) error                                                   // 修改系统设置
	GetCashierBaseSetting(ctx context.Context) (resp.CashierBaseSetting, error)                                                           // 获取收银端设置
	GetAcceptOrderSetting(ctx context.Context) (*resp.AcceptOrderSetting, error)                                                          // 获取接单设置
	SymbolPosition(ctx context.Context, price float64) string                                                                             // 根据货币符号位置返回字符串
	EditStoreSetting(ctx context.Context, storeSetting req.UpdateStoreSetting) error                                                      // 修改店铺设置
	EditBusinessSetting(ctx context.Context, businessSetting req.UpdateBusinessSetting) error                                             // 修改业务设置
	GetShopBusinessSetting(ctx context.Context) (setting.ShopBusiness, error)                                                             // 获取商家业务设置
	GetMenuQrcode(ctx context.Context) (string, error)                                                                                    // 获取电子菜单二维码
	GetPaymentMethodList(ctx context.Context) setting.PaymentMethodListResp                                                               // 获取支付方式列表
}

func NewSrv(dbm *database.DBManager, cache cache.Cache) ISrv {
	return NewSrvImpl(dbm, cache)
}

type Srv struct {
	dbm           *database.DBManager
	cache         cache.Cache
	cacheKey      string
	cloudCacheKey string
}

func NewSrvImpl(dbm *database.DBManager, cache cache.Cache) *Srv {
	return &Srv{
		dbm:           dbm,
		cache:         cache,
		cacheKey:      "setting:company_id:%d",
		cloudCacheKey: "__CLOUD__SYNCBASEINFO__",
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
				return settings, errors.WithMessage(err)
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
		return nil, errors.WithMessage(err, "获取设置失败")
	}

	data, _ := json.Marshal(settings)
	s.cache.Set(cacheKey, string(data), 0)

	return settings, nil
}

// GetStoreLanguageList 获取商家语言列表
func (s *Srv) GetStoreLanguageList(ctx context.Context) ([]dto.LanguageItem, error) {
	set, err := s.GetStoreSetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return set.Language, nil
}

func (s *Srv) GetStoreLanguage(ctx context.Context) ([]string, error) {
	storeSetting, _ := s.GetStoreSetting(ctx)
	languages := make([]string, 0)
	for _, language := range storeSetting.Language {
		languages = append(languages, language.Name)
	}
	return languages, nil
}

func (s *Srv) getSettingByKey(ctx context.Context, key string) model.Setting {
	defaultSetting := model.Setting{Key: key, Values: "{}"}
	if key == constant.SettingCloudBasic { // 平台设置
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
			res, err := gohttp.NewRequest().Post(viper.GetString("CLOUD_PLATFORM_HOST") + "/api/admin/setting.service/info")
			if err != nil {
				return defaultSetting
			}
			defer res.GetBody().Close()
			bodyBytes, _ := res.GetBodyAsByte()
			var base Base
			if err := json.Unmarshal(bodyBytes, &base); err != nil {
				return defaultSetting
			}
			var basicCloudSetting setting.CloudBasic
			copier.Copy(&basicCloudSetting, base.Data.Base)
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

// GetStoreSetting 获取商家设置
func (s *Srv) GetStoreSetting(ctx context.Context) (setting.Store, error) {
	var store setting.Store
	st := s.getSettingByKey(ctx, constant.SettingStore)

	// 解析json字符串为map进行处理，处理language字段
	jsonStr := st.Values
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &jsonMap)
	if err != nil {
		ctx.Log().Error("解析商城设置失败-01", zap.Error(err))
		return store, errors.New("解析商城设置失败-01" + err.Error())
	}
	// 处理language数组中的key字段
	if language, ok := jsonMap["language"].([]interface{}); ok {
		for i, item := range language {
			if langItem, ok := item.(map[string]interface{}); ok {
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
		ctx.Log().Error("重新序列化JSON失败", zap.Error(err))
		return store, errors.New("重新序列化JSON失败" + err.Error())
	}

	err = json.Unmarshal(modifiedJSON, &store)
	if err != nil {
		ctx.Log().Error("解析商城设置失败", zap.Error(err))
		return store, errors.New("解析商城设置失败" + err.Error())
	}
	if store.IPWhiteList != "" {
		store.IPWhiteList = viper.GetString("PAY_SERVICE_IP")
	}

	defaultStore := s.getDefaultStore(ctx.GetLanguage())
	store.TimeZoneList = nil
	err = copier.CopyWithOption(&defaultStore, store, copier.Option{IgnoreEmpty: true})
	if err != nil {
		ctx.Log().Error("合并商城设置失败", zap.Error(err))
		return store, errors.New("合并商城设置失败")
	}
	ginContext := ctx.GetGin()
	if defaultStore.LogoURL != "" && ginContext != nil {
		defaultStore.LogoURL = utils.GetBaseURL(ginContext.Request) + defaultStore.LogoURL
	}
	if defaultStore.AvatarURL != "" && ginContext != nil {
		defaultStore.AvatarURL = utils.GetBaseURL(ginContext.Request) + defaultStore.AvatarURL
	}
	if len(defaultStore.TimeZoneList) == 0 {
		defaultStore.TimeZoneList = make([]setting.TimeZoneItem, 0)
	}
	if len(defaultStore.Language) == 0 {
		defaultStore.Language = make([]dto.LanguageItem, 0)
	}

	if defaultStore.Coordinates != "" {
		latLng := strings.Split(defaultStore.Coordinates, ",")
		if len(latLng) == 2 {
			// 转成float64保留6位小数，然后再转成字符串
			lat, _ := strconv.ParseFloat(latLng[0], 64)
			lng, _ := strconv.ParseFloat(latLng[1], 64)
			defaultStore.Latitude = fmt.Sprintf("%.6f", lat)
			defaultStore.Longitude = fmt.Sprintf("%.6f", lng)
		}
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
			return printer, errors.WithMessage(err)
		}
	}
	st := s.getSettingByKey(ctx, constant.SettingPrinter)

	// 使用正则表达式预处理JSON字符串，将language_list中的key字段从字符串转为数字
	jsonStr := st.Values

	// 解析json字符串为map进行处理
	var jsonMap map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &jsonMap)
	if err != nil {
		ctx.Log().Error("解析小票打印机设置失败", zap.Error(err))
		return printer, errors.New("解析小票打印机设置失败" + err.Error())
	}
	// 处理 cashier_printer 字段，确保它是一个数组
	if cashierPrinter, ok := jsonMap["cashier_printer"]; ok {
		// 如果是对象，转换为数组
		if _, ok := cashierPrinter.(map[string]interface{}); ok {
			// 将对象转换为数组
			cashierPrinterArray := []interface{}{cashierPrinter}
			jsonMap["cashier_printer"] = cashierPrinterArray
		}
	}

	// 处理 KitchenPrintMethod
	if kitchenPrintMethod, ok := jsonMap["kitchen_print_method"]; ok {
		// 如果是float64（JSON中的数字类型），转换为字符串
		if numVal, ok := kitchenPrintMethod.(float64); ok {
			jsonMap["kitchen_print_method"] = fmt.Sprintf("%d", int(numVal))
		} else if strVal, ok := kitchenPrintMethod.(string); ok {
			jsonMap["kitchen_print_method"] = strVal
		}
	}
	// 处理print_method
	if printMethod, ok := jsonMap["print_method"]; ok {
		// 如果是float64（JSON中的数字类型），转换为字符串
		if numVal, ok := printMethod.(float64); ok {
			jsonMap["print_method"] = fmt.Sprintf("%d", int(numVal))
		} else if strVal, ok := printMethod.(string); ok {
			jsonMap["print_method"] = strVal
		}
	}
	// 处理language_list中的key
	if languageList, ok := jsonMap["language_list"].([]interface{}); ok {
		for i, item := range languageList {
			if langItem, ok := item.(map[string]interface{}); ok {
				// 尝试将key转换为int
				if keyStr, ok := langItem["key"].(string); ok {
					langItem["key"], _ = strconv.Atoi(keyStr)
					languageList[i] = langItem
				}
			}
		}
		jsonMap["language_list"] = languageList
	}
	// 重新序列化为JSON
	modifiedJSON, err := json.Marshal(jsonMap)
	if err != nil {
		ctx.Log().Error("重新序列化JSON失败", zap.Error(err))
		return printer, errors.New("重新序列化JSON失败" + err.Error())
	}

	// 使用处理后的JSON解析
	err = json.Unmarshal(modifiedJSON, &printer)
	if err != nil {
		ctx.Log().Error("解析小票打印机设置失败", zap.Error(err))
		return printer, errors.New("解析小票打印机设置失败" + err.Error())
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

	if len(defaultPrinter.CashierPrinter) == 0 {
		defaultPrinter.CashierPrinter = make([]setting.CashierPrinterItem, 0)
	} else {
		handledCashierPrinters := make([]setting.CashierPrinterItem, 0)
		for _, item := range defaultPrinter.CashierPrinter {
			handledCashierPrinters = append(handledCashierPrinters, setting.CashierPrinterItem{
				Key:          item.Key,
				PrinterId:    utils.Uint64OrStringToString(item.PrinterId),
				PrinterUsbId: item.PrinterUsbId,
				Sn:           item.Sn,
			})
		}
		defaultPrinter.CashierPrinter = handledCashierPrinters
	}
	if len(defaultPrinter.LanguageList) == 0 {
		defaultPrinter.LanguageList = make([]dto.LanguageItem, 0)
	}
	if len(defaultPrinter.CalendarList) == 0 {
		defaultPrinter.CalendarList = make([]setting.CalendarItem, 0)
	}
	if len(defaultPrinter.PrintList) == 0 {
		defaultPrinter.PrintList = make([]setting.PrintItem, 0)
	}
	if len(defaultPrinter.Language) == 0 {
		defaultPrinter.Language = make([]string, 0)
	}

	return defaultPrinter, nil
}

// GetPrinterInfo 获取打印机设置
func (s *Srv) GetPrinterInfo(ctx context.Context, printerSetting setting.Printer, deviceSn string) (setting.PrinterInfo, error) {
	var (
		isCashierOpen          = printerSetting.CashierOpen == "1"
		printerId              string
		printerUuid            uint64
		printer                model.Printer
		err                    error
		copies                 uint = 1
		printerConfig          string
		printerType            string
		printerCashierDeviceSn string
		isCashierPrinter       bool // 是否收银机自带打印机
		isUsbPrinter           bool // 是否usb打印机
		printMethod            int  // 打印方式 1文本打印, 2图片打印
		printerSn              string
		printerWidth           int = 80 // 默认80mm打印机
		enableStatusCheck      int = 0  // 是否启用状态检查
		enableSound            int = 0  // 是否启用打印提示音
	)

	// 收银机开启
	if isCashierOpen {

		// 收银机机绑定的打印机key
		for _, cashierPrinter := range printerSetting.CashierPrinter {
			if cashierPrinter.Key == deviceSn {
				if cashierPrinter.PrinterUsbId != "" && cashierPrinter.PrinterUsbId != "0" {
					printerId = cashierPrinter.PrinterUsbId
					isUsbPrinter = true
				} else {
					printerId = utils.Uint64OrStringToString(cashierPrinter.PrinterId) // 如果是18位纯数字，说明是普通打印机
				}
				break
			}
		}

		// 普通打印机 uuid uint64 字符串
		matched, _ := regexp.MatchString(`^\d+$`, printerId)
		if len(printerId) <= 18 && matched {
			printerUuid, _ = strconv.ParseUint(printerId, 10, 64)
			printerRepo := repository.NewPrinterRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
			printer, err = printerRepo.GetPrinter(printerRepo.WhereUuid(printerUuid), printerRepo.WithPrinterType())
			if err != nil {
				return setting.PrinterInfo{}, errors.WithMessage(err)
			}
			copies = printer.Copies
			printerConfigJson := printer.GetConfigJson()
			printerConfig = utils.ToJson(printerConfigJson)
			if printer.PrinterType != nil {
				printerType = printer.PrinterType.Key
			}
			// 打印机SN
			if printer.Sn != "" {
				printerSn = printer.Sn
			} else if printer.IsUsb != 1 {
				printerSn = printerConfigJson.SN
			}
			// 由当前点击的设备进行打印
			printerCashierDeviceSn = deviceSn
			printMethod = int(printer.PrintMethod)
			printerWidth = printer.Width
			enableStatusCheck = printer.EnableStatusCheck
			enableSound = printer.EnableSound
		} else if printerId != "0" && printerId != "" {
			// 收银机内置的打印机
			printerCashierDeviceSn = printerId
			isCashierPrinter = true
			deviceRepo := repository.NewDeviceRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
			brand := deviceRepo.GetDeviceBrand(deviceRepo.WhereSn(printerId))
			if slices.Contains(constant.SunmiAllPrints, brand) {
				// 商米打印机
				printerType = constant.PrinterTypeCashierSunmi
			} else if slices.Contains([]string{constant.BrandA11510P}, brand) {
				// compax打印机
				printerType = constant.PrinterTypeCashierCompax
			} else {
				// 未知打印机
				printerType = ""
			}
		}
	}
	//
	return setting.PrinterInfo{
		PrinterType:            printerType,
		PrinterUuid:            printerUuid, // 默认为0，如果是普通打印机，则为model.Printer的Uuid
		Copies:                 copies,
		PrinterConfig:          printerConfig,
		IsCashierPrinter:       isCashierPrinter,
		IsCashierOpen:          isCashierOpen,
		PrinterCashierDeviceSn: printerCashierDeviceSn,
		IsUsbPrinter:           isUsbPrinter,
		PrintMethod:            printMethod,
		PrinterSn:              printerSn,
		PrinterWidth:           printerWidth,
		EnableStatusCheck:      enableStatusCheck,
		EnableSound:            enableSound,
	}, nil
}

// GetBusinessSetting 门店业务设置
func (s *Srv) GetBusinessSetting(ctx context.Context) (setting.Business, error) {
	st := s.getSettingByKey(ctx, constant.SettingBusiness)
	var business setting.Business

	// 兼容v1.0版本dish_card_style_time字段为数字的情况
	{
		// 正则表达式用于匹配 dish_card_style_time 后面的任意数字
		re := regexp.MustCompile(`"dish_card_style_time":(\s*)(\d+)`)
		// 替换为带引号的字符串数字
		st.Values = re.ReplaceAllString(st.Values, `"dish_card_style_time":"$2"`)
	}

	// 兼容v1.0版本dish_card_style字段为数字的情况
	{
		// 正则表达式用于匹配 dish_card_style 后面的任意数字
		re := regexp.MustCompile(`"dish_card_style":(\s*)(\d+)`)
		// 替换为带引号的字符串数字
		st.Values = re.ReplaceAllString(st.Values, `"dish_card_style":"$2"`)
	}

	// 兼容v1.0版本start_serial_no字段为数字的情况
	{
		// 正则表达式用于匹配 start_serial_no 后面的任意数字
		re := regexp.MustCompile(`"start_serial_no":(\s*)(\d+)`)
		// 替换为带引号的字符串数字
		st.Values = re.ReplaceAllString(st.Values, `"start_serial_no":"$2"`)
	}

	err := json.Unmarshal([]byte(st.Values), &business)
	if err != nil {
		ctx.Log().Error("解析门店-业务设置失败", zap.Error(err))
		return business, errors.WithMessage(errors.New("解析门店-业务设置失败"), err.Error())
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

	if len(defaultBusiness.ZeroingMethodList) == 0 {
		defaultBusiness.ZeroingMethodList = make([]setting.ZeroingMethodItem, 0)
	}
	if len(defaultBusiness.CheckoutZeroingMethodList) == 0 {
		defaultBusiness.CheckoutZeroingMethodList = make([]setting.CheckoutZeroingMethodItem, 0)
	}
	if len(defaultBusiness.GiftMethodList) == 0 {
		defaultBusiness.GiftMethodList = make([]setting.GiftMethodItem, 0)
	}
	if len(defaultBusiness.FreeMethodList) == 0 {
		defaultBusiness.FreeMethodList = make([]setting.FreeMethodItem, 0)
	}

	// 分批商品相关
	{
		db := s.dbm.GetDB(ctx.GetCompanyUuid())
		// 分批商品数量
		productPackageRepo := repository.NewProductPackageRepo(db)
		batchProductUuids, err := productPackageRepo.GetProductPackageBatchTagCount()
		if err != nil {
			return business, errors.WithMessage(err)
		}
		defaultBusiness.BatchProductUuids = batchProductUuids

		// 分批类型数量
		batchTagNum, err := repository.NewBatchTagRepo(db).GetBatchTagCount()
		if err != nil {
			return business, errors.WithMessage(err)
		}
		defaultBusiness.BatchTagNum = uint(batchTagNum)
	}

	return defaultBusiness, nil
}

// GetBuffetSetting 自助餐设置
func (s *Srv) GetBuffetSetting(ctx context.Context, companySetting model.CompanySetting) (setting.BuffetResp, error) {
	st := s.getSettingByKey(ctx, constant.SettingBuffet)
	var buffet setting.Buffet
	err := json.Unmarshal([]byte(st.Values), &buffet)
	if err != nil {
		return setting.BuffetResp{}, errors.New("解析自助餐-自助餐设置失败")
	}
	if companySetting.IsOpenBuffet == 0 {
		buffet.IsOpen = "0"
	}
	defaultBuffet := s.getDefaultBuffet()
	err = copier.CopyWithOption(&defaultBuffet, buffet, copier.Option{IgnoreEmpty: true})
	if err != nil {
		return setting.BuffetResp{}, errors.New("解析自助餐-自助餐设置失败")
	}
	if len(defaultBuffet.AddClock) == 0 {
		defaultBuffet.AddClock = make([]setting.AddClockItem, 0)
	}
	tabletEndTime, _ := strconv.Atoi(defaultBuffet.TabletEndTime)
	return setting.BuffetResp{
		IsOpen:                   defaultBuffet.IsOpen,
		TabletEndTime:            tabletEndTime * 60,
		IsRemainContinue:         defaultBuffet.IsRemainContinue,
		RemainContinueTime:       defaultBuffet.RemainContinueTime,
		RemainContinueNoticeTime: defaultBuffet.RemainContinueNoticeTime,
		IsBuyContinue:            defaultBuffet.IsBuyContinue,
		IsAddClock:               defaultBuffet.IsAddClock,
		IsBuffetDiscount:         defaultBuffet.IsBuffetDiscount,
		AddClock:                 defaultBuffet.AddClock,
	}, nil

}

// 转换平板设置指定字段类型
func (s *Srv) convertTabletFormat(oldVal string) ([]byte, error) {
	tabletMap := map[string]any{}
	err := json.Unmarshal([]byte(oldVal), &tabletMap)
	if err != nil {
		return nil, err
	}
	if v, ok := tabletMap["is_call_service"]; ok {
		switch v.(type) {
		case float64:
			tabletMap["is_call_service"] = fmt.Sprintf("%.0f", v)
		}
	}
	if v, ok := tabletMap["is_customer_order"]; ok {
		switch v.(type) {
		case float64:
			tabletMap["is_customer_order"] = fmt.Sprintf("%.0f", v)
		}
	}
	if v, ok := tabletMap["is_voice_remind"]; ok {
		switch v.(type) {
		case float64:
			tabletMap["is_voice_remind"] = fmt.Sprintf("%.0f", v)
		}
	}
	if v, ok := tabletMap["is_show_sold_out"]; ok {
		switch v.(type) {
		case float64:
			tabletMap["is_show_sold_out"] = fmt.Sprintf("%.0f", v)
		}
	}
	if v, ok := tabletMap["is_buffet_order_limit"]; ok {
		switch v.(type) {
		case float64:
			tabletMap["is_buffet_order_limit"] = fmt.Sprintf("%.0f", v)
		}
	}
	if v, ok := tabletMap["is_order_limit"]; ok {
		switch v.(type) {
		case float64:
			tabletMap["is_order_limit"] = fmt.Sprintf("%.0f", v)
		}
	}
	if v, ok := tabletMap["buffet_order_limit"]; ok {
		switch v.(type) {
		case []any:
			tabletMap["buffet_order_limit"] = struct{}{}
		}
	}
	if v, ok := tabletMap["order_limit"]; ok {
		switch v.(type) {
		case []any:
			tabletMap["order_limit"] = struct{}{}
		}
	}
	return json.Marshal(tabletMap)
}

// GetTabletSetting 平板端设置
func (s *Srv) GetTabletSetting(ctx context.Context, languageList []dto.LanguageItem) (setting.Tablet, error) {
	st := s.getSettingByKey(ctx, constant.SettingTablet)
	ginContext := ctx.GetGin()
	var (
		tablet setting.Tablet
		err    error
	)
	if languageList == nil {
		languageList, err = s.GetStoreLanguageList(ctx)
		if err != nil {
			return tablet, errors.WithMessage(err)
		}
	}
	val, err := s.convertTabletFormat(st.Values)
	if err != nil {
		ctx.Log().Error("解析各端-平板端设置失败", zap.Error(err))
		return tablet, errors.New("解析各端-平板端设置失败")
	}
	st.Values = string(val)
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
	err = copier.CopyWithOption(&defaultTablet, tablet, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	if err != nil {
		ctx.Log().Error("合并各端-平板端设置失败", zap.Error(err))
		return tablet, errors.New("合并各端-平板端设置失败")
	}
	if len(defaultTablet.Carousel) == 0 {
		defaultTablet.Carousel = make([]setting.CarouselItem, 0)
	}
	if len(defaultTablet.LanguageList) == 0 {
		defaultTablet.LanguageList = make([]dto.LanguageItem, 0)
	}
	if len(defaultTablet.Language) == 0 {
		defaultTablet.Language = make([]string, 0)
	}
	validLanguageList := make([]dto.LanguageItem, 0)
	var languageNames []string
	for _, item := range defaultTablet.LanguageList {
		if slices.Contains(defaultTablet.Language, item.Name) && !slices.Contains(languageNames, item.Name) {
			validLanguageList = append(validLanguageList, item)
			languageNames = append(languageNames, item.Name)
		}
	}
	defaultTablet.LanguageList = validLanguageList
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
			return cashier, errors.WithMessage(err)
		}
	}

	st := s.getSettingByKey(ctx, constant.SettingCashier)
	isSetIsShowScanSoldOut := strings.Contains(st.Values, "is_show_scan_sold_out")
	isSetIsShowAssistantSoldOut := strings.Contains(st.Values, "is_show_assistant_sold_out")

	// 解析json字符串为map进行处理
	modifiedJSON, err := s.parseCashierSetting(st.Values)
	if err != nil {
		return cashier, err
	}
	err = json.Unmarshal(modifiedJSON, &cashier)
	if err != nil {
		ctx.Log().Error("解析各端-收银机设置失败 - 02", zap.Error(err))
		return cashier, errors.New("解析各端-收银机设置失败 - 02" + err.Error())
	}

	// 滚动图/视频处理
	ginContext := ctx.GetGin()
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

	err = copier.CopyWithOption(&defaultCashier, cashier, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	if isSetIsShowScanSoldOut {
		defaultCashier.IsShowScanSoldOut = cashier.IsShowScanSoldOut // h5端是否显示售罄
	}
	if isSetIsShowAssistantSoldOut {
		defaultCashier.IsShowAssistantSoldOut = cashier.IsShowAssistantSoldOut // 助手端是否显示售罄
	}
	if err != nil {
		ctx.Log().Error("合并各端-收银机设置失败 - 04", zap.Error(err))
		return cashier, errors.New("合并各端-收银机设置失败 - 04")
	}

	if len(defaultCashier.Carousel) == 0 {
		defaultCashier.Carousel = make([]setting.CarouselItem, 0)
	}
	if len(defaultCashier.LanguageList) == 0 {
		defaultCashier.LanguageList = make([]dto.LanguageItem, 0)
	}
	if len(defaultCashier.Language) == 0 {
		defaultCashier.Language = make([]string, 0)
	}
	if len(defaultCashier.RemainColor) == 0 {
		defaultCashier.RemainColor = make([]string, 0)
	}
	validLanguageList := make([]dto.LanguageItem, 0)
	var languageNames []string
	for _, item := range defaultCashier.LanguageList {
		if slices.Contains(defaultCashier.Language, item.Name) && !slices.Contains(languageNames, item.Name) {
			validLanguageList = append(validLanguageList, item)
			languageNames = append(languageNames, item.Name)
		}
	}
	defaultCashier.LanguageList = validLanguageList
	return defaultCashier, nil
}

// GetCloudBasicSetting 获取云端基础信息
func (s *Srv) GetCloudBasicSetting(ctx context.Context) (setting.CloudBasic, error) {
	var (
		err        error
		cloudBasic setting.CloudBasic
	)
	st := s.getSettingByKey(ctx, constant.SettingCloudBasic)
	err = json.Unmarshal([]byte(st.Values), &cloudBasic)
	if err != nil {
		ctx.Log().Error("解析云端基础信息", zap.Error(err))
		return cloudBasic, errors.New("解析云端基础信息失败")
	}

	cloudBasic.BrandLogo = utils.AddImageDomain(utils.RemoveDomain(cloudBasic.BrandLogo), utils.GetBaseURL(ctx.GetGin().Request), true)
	cloudBasic.BrandLogoLong = utils.AddImageDomain(utils.RemoveDomain(cloudBasic.BrandLogoLong), utils.GetBaseURL(ctx.GetGin().Request), true)
	cloudBasic.BrowserLogo = utils.AddImageDomain(utils.RemoveDomain(cloudBasic.BrowserLogo), utils.GetBaseURL(ctx.GetGin().Request), true)

	return cloudBasic, nil
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
			return assistant, errors.WithMessage(err)
		}
	}
	st := s.getSettingByKey(ctx, constant.SettingAssistant)
	// 解析json字符串为map进行处理
	var jsonMap map[string]interface{}
	err = json.Unmarshal([]byte(st.Values), &jsonMap)
	if err != nil {
		ctx.Log().Error("解析点餐助手设置失败-01", zap.Error(err))
		return assistant, errors.New("解析点餐助手设置失败-01" + err.Error())
	}
	// 处理 isShowAssistantSoldOut
	if isShowAssistantSoldOut, ok := jsonMap["is_show_assistant_sold_out"]; ok {
		if numVal, ok := isShowAssistantSoldOut.(float64); ok {
			jsonMap["is_show_assistant_sold_out"] = int(numVal)
		} else if strVal, ok := isShowAssistantSoldOut.(string); ok {
			jsonMap["is_show_assistant_sold_out"], _ = strconv.Atoi(strVal)
		}
	}
	// 重新序列化为JSON
	modifiedJSON, err := json.Marshal(jsonMap)
	if err != nil {
		ctx.Log().Error("解析点餐助手设置失败 - 重新序列化JSON失败 - 02", zap.Error(err))
		return assistant, errors.New("解析点餐助手设置失败 - 重新序列化JSON失败 - 02" + err.Error())
	}
	//
	err = json.Unmarshal(modifiedJSON, &assistant)
	if err != nil {
		ctx.Log().Error("解析各端-点餐助手设置失败 - 01", zap.Error(err))
		return assistant, errors.New("解析各端-点餐助手设置失败 - 01")
	}
	if len(assistant.LanguageList) == 0 {
		assistant.LanguageList = nil
	}
	defaultAssistant := s.getDefaultAssistant(languageList)
	err = copier.CopyWithOption(&defaultAssistant, assistant, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	if err != nil {
		ctx.Log().Error("合并各端-点餐助手设置失败 - 02", zap.Error(err))
		return assistant, errors.New("合并各端-点餐助手设置失败 - 02")
	}

	// 如果设置了 is_show_assistant_sold_out，则读取解析后的数据，否则读取默认设置
	cashierSet := s.getSettingByKey(ctx, constant.SettingCashier)
	if strings.Contains(cashierSet.Values, "\"is_show_assistant_sold_out\"") {
		modifiedJSON, err := s.parseCashierSetting(cashierSet.Values)
		if err != nil {
			return assistant, err
		}
		var cashier setting.Cashier
		err = json.Unmarshal(modifiedJSON, &cashier)
		if err != nil {
			ctx.Log().Error("解析各端-点餐助手设置失败 - 02", zap.Error(err))
			return assistant, errors.New("解析各端-点餐助手设置失败 - 02")
		}
		defaultAssistant.IsShowAssistantSoldOut = cashier.IsShowAssistantSoldOut
	} else {
		defaultAssistant.IsShowAssistantSoldOut = s.getDefaultCashier(languageList).IsShowAssistantSoldOut
	}

	if len(defaultAssistant.RemainColor) == 0 {
		defaultAssistant.RemainColor = make([]string, 0)
	}
	if len(defaultAssistant.LanguageList) == 0 {
		defaultAssistant.LanguageList = make([]dto.LanguageItem, 0)
	}
	if len(defaultAssistant.Language) == 0 {
		defaultAssistant.Language = make([]string, 0)
	}

	validLanguageList := make([]dto.LanguageItem, 0)
	var languageNames []string
	for _, item := range defaultAssistant.LanguageList {
		if slices.Contains(defaultAssistant.Language, item.Name) && !slices.Contains(languageNames, item.Name) {
			validLanguageList = append(validLanguageList, item)
			languageNames = append(languageNames, item.Name)
		}
	}
	defaultAssistant.LanguageList = validLanguageList

	return defaultAssistant, nil
}

// GetPointsSetting 获取积分设置
func (s *Srv) GetPointsSetting(ctx context.Context) (setting.Points, error) {
	var (
		err    error
		points setting.Points
	)
	st := s.getSettingByKey(ctx, constant.SettingPoints)
	err = json.Unmarshal([]byte(st.Values), &points)
	if err != nil {
		ctx.Log().Error("解析积分设置失败", zap.Error(err))
		return points, errors.New("解析积分设置失败")
	}
	defaultPoints := s.getDefaultPoints()
	err = copier.CopyWithOption(&defaultPoints, points, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	if err != nil {
		ctx.Log().Error("合并积分设置失败", zap.Error(err))
		return points, errors.New("合并积分设置失败")
	}
	companySetting, err := s.GetCompanySetting(ctx)
	if err != nil {
		ctx.Log().Error("合并积分设置失败", zap.Error(err))
		return points, errors.New("获取商家信息失败")
	}
	memberRepo := repository.NewMemberRepo(ctx.GetDB())
	memberLevels := memberRepo.GetMemberLevels()
	var rateMemberLevels []setting.MemberLevelItem
	var quantityMemberLevels []setting.MemberLevelItem
	for _, level := range memberLevels {
		rateMemberLevels = append(rateMemberLevels, setting.MemberLevelItem{
			Uuid:  level.Uuid,
			Name:  level.Name,
			Value: level.PointsRate,
		})
		quantityMemberLevels = append(quantityMemberLevels, setting.MemberLevelItem{
			Uuid:  level.Uuid,
			Name:  level.Name,
			Value: level.PointsQuantity,
		})
	}
	for i, rule := range defaultPoints.ShoppingGiftRules {
		if companySetting.IsOpenBuffet == 0 { // 未开启自助餐
			newMealType := slice.Filter(rule.MealType, func(index int, item string) bool {
				return item != "buffet"
			})
			defaultPoints.ShoppingGiftRules[i].MealType = newMealType
		}

		switch rule.Type {
		case setting.RuleTypePaymentAmount:
			defaultPoints.ShoppingGiftRules[i].MemberLevels = rateMemberLevels
		case setting.RuleTypeDesk:
			defaultPoints.ShoppingGiftRules[i].MemberLevels = quantityMemberLevels
		}
	}

	return defaultPoints, nil
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
			return kitchen, errors.WithMessage(err)
		}
	}
	defaultKitchen := s.getDefaultKitchen(languageList)

	// 语言 不需要合并
	defaultKitchen.Language = nil

	err = copier.CopyWithOption(&defaultKitchen, kitchen, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	if err != nil {
		ctx.Log().Error("合并各端-厨显设置失败", zap.Error(err))
		return kitchen, errors.New("合并各端-厨显设置失败")
	}
	// 总权限 - 不开启厨显
	if companySetting.IsOpenKitchenKds == 0 {
		kitchen.IsOpen = "0"
	}

	if len(defaultKitchen.WaitColor) == 0 {
		defaultKitchen.WaitColor = make([]string, 0)
	}
	if len(defaultKitchen.LanguageList) == 0 {
		defaultKitchen.LanguageList = make([]dto.LanguageItem, 0)
	}
	if len(defaultKitchen.Language) == 0 {
		defaultKitchen.Language = make([]string, 0)
	}

	validLanguageList := make([]dto.LanguageItem, 0)
	var languageNames []string
	for _, item := range defaultKitchen.LanguageList {
		if slices.Contains(defaultKitchen.Language, item.Name) && !slices.Contains(languageNames, item.Name) {
			validLanguageList = append(validLanguageList, item)
			languageNames = append(languageNames, item.Name)
		}
	}
	defaultKitchen.LanguageList = validLanguageList

	return defaultKitchen, nil
}

// GetH5Setting 获取H5设置
func (s *Srv) GetH5Setting(ctx context.Context, languageList []dto.LanguageItem) (setting.H5, error) {
	var (
		err error
		h5  setting.H5
	)
	if languageList == nil {
		languageList, err = s.GetStoreLanguageList(ctx)
		if err != nil {
			return h5, errors.WithMessage(err)
		}
	}
	st := s.getSettingByKey(ctx, constant.SettingH5)
	if strings.Contains(st.Values, "\"buffet_order_limit\":[]") {
		st.Values = strings.Replace(st.Values, "\"buffet_order_limit\":[]", "\"buffet_order_limit\":{}", -1)
	}
	if strings.Contains(st.Values, "\"order_limit\":[]") {
		st.Values = strings.Replace(st.Values, "\"order_limit\":[]", "\"order_limit\":{}", -1)
	}
	// 解析json字符串为map进行处理
	var jsonMap map[string]interface{}
	err = json.Unmarshal([]byte(st.Values), &jsonMap)
	if err != nil {
		ctx.Log().Error("解析各端-扫码H5设置失败 - 01", zap.Error(err))
		return h5, errors.New("解析各端-扫码H5设置失败 - 01")
	}
	// 处理 isShowScanSoldOut
	if isShowScanSoldOut, ok := jsonMap["is_show_scan_sold_out"]; ok {
		if numVal, ok := isShowScanSoldOut.(float64); ok {
			jsonMap["is_show_scan_sold_out"] = int(numVal)
		} else if strVal, ok := isShowScanSoldOut.(string); ok {
			jsonMap["is_show_scan_sold_out"], _ = strconv.Atoi(strVal)
		}
	}
	// 重新序列化为JSON
	modifiedJSON, err := json.Marshal(jsonMap)
	if err != nil {
		ctx.Log().Error("重新序列化JSON失败 - 02", zap.Error(err))
		return h5, errors.New("重新序列化JSON失败 - 02")
	}
	// 解析json字符串
	err = json.Unmarshal(modifiedJSON, &h5)
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
		modifiedJSON, err := s.parseCashierSetting(cashierSet.Values)
		if err != nil {
			return defaultH5, err
		}
		var cashier setting.Cashier
		err = json.Unmarshal(modifiedJSON, &cashier)
		if err != nil {
			ctx.Log().Error("解析各端-获取H5设置 - 02", zap.Error(err))
			return defaultH5, errors.New("解析各端-获取H5设置 - 02")
		}
		defaultH5.IsShowScanSoldOut = cashier.IsShowScanSoldOut
	} else {
		defaultH5.IsShowScanSoldOut = s.getDefaultCashier(languageList).IsShowScanSoldOut
	}

	if len(defaultH5.LanguageList) == 0 {
		defaultH5.LanguageList = make([]dto.LanguageItem, 0)
	}
	if len(defaultH5.Language) == 0 {
		defaultH5.Language = make([]string, 0)
	}

	validLanguageList := make([]dto.LanguageItem, 0)
	var languageNames []string
	for _, item := range defaultH5.LanguageList {
		if slices.Contains(defaultH5.Language, item.Name) && !slices.Contains(languageNames, item.Name) {
			validLanguageList = append(validLanguageList, item)
			languageNames = append(languageNames, item.Name)
		}
	}
	defaultH5.LanguageList = validLanguageList
	return defaultH5, nil
}

// GetCompanySetting 获取公司设置
func (s *Srv) GetCompanySetting(ctx context.Context) (model.CompanySetting, error) {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	if db == nil {
		return model.CompanySetting{}, errors.New("db not found")
	}
	companySettingRepo := repository.NewCompanySettingRepo(db)
	return companySettingRepo.Get(), nil
}

// GetCashierLanguage 获取收银机语言
func (s *Srv) GetCashierLanguage(c context.Context) (resp.LanguageResp, error) {
	cashierSetting, err := s.GetCashierSetting(c, nil)
	if err != nil {
		return resp.LanguageResp{}, errors.New("获取语言失败")
	}
	languageResp := resp.LanguageResp{
		Languages:       cashierSetting.Language,
		LanguageList:    cashierSetting.LanguageList,
		DefaultLanguage: cashierSetting.DefaultLanguage,
	}
	if len(languageResp.Languages) == 0 {
		languageResp.Languages = make([]string, 0)
	}
	if len(languageResp.LanguageList) == 0 {
		languageResp.LanguageList = make([]dto.LanguageItem, 0)
	}

	validLanguageList := make([]dto.LanguageItem, 0)
	var languageNames []string
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
func (s *Srv) GetCashierAd(ctx context.Context) (resp.Ads, error) {
	cashierSetting, err := s.GetCashierSetting(ctx, nil)
	if err != nil {
		return resp.Ads{List: make([]setting.CarouselItem, 0)}, errors.New("获取副屏广告失败")
	}
	return resp.Ads{
		List: cashierSetting.Carousel,
	}, nil
}

// 转换服务费指定字段类型
func (s *Srv) convertServiceFeeFormat(oldVal string) ([]byte, error) {
	serviceFeeMap := map[string]any{}
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

// GetServiceFeeSetting 获取服务费设置
func (s *Srv) GetServiceFeeSetting(ctx context.Context) (setting.ServiceCharge, error) {
	st := s.getSettingByKey(ctx, constant.SettingServiceCharge)
	var serviceFee setting.ServiceCharge
	// 修改类型
	newVal, err := s.convertServiceFeeFormat(st.Values)
	if err != nil {
		ctx.Log().Error("解析服务费设置失败", zap.Error(err))
		return serviceFee, errors.New("解析服务费设置失败")
	}
	st.Values = string(newVal)
	err = json.Unmarshal([]byte(st.Values), &serviceFee)
	if err != nil {
		ctx.Log().Error("解析服务费设置失败", zap.Error(err))
		return serviceFee, errors.New("解析服务费设置失败")
	}
	if serviceFee.IsOpen == "0" {
		serviceFee.IsOpen = "0"
	}
	defaultServiceFee := s.getDefaultServiceCharge()
	err = copier.CopyWithOption(&defaultServiceFee, serviceFee, copier.Option{IgnoreEmpty: true})

	if len(defaultServiceFee.ApplyScopeTableList) == 0 {
		defaultServiceFee.ApplyScopeTableList = make([]int64, 0)
	}
	if err != nil {
		ctx.Log().Error("解析服务费设置失败", zap.Error(err))
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
	if len(defaultTaxRate.AddTaxCategory) == 0 {
		defaultTaxRate.AddTaxCategory = make([]setting.AddTaxCategoryItem, 0)
	}
	return defaultTaxRate, nil
}

// VerifyPassword 验证密码
func (s *Srv) VerifyPassword(ctx context.Context, source string, typ string, password string) bool {
	passwordMap := make(map[string]string)
	switch source {
	case constant.SourceCashier:
		cashierSetting, err := s.GetCashierSetting(ctx, nil)
		if err != nil {
			return false
		}
		passwordMap = map[string]string{
			constant.PasswordTypeCashBox:  cashierSetting.CashierPassword,
			constant.PasswordTypeAdvanced: cashierSetting.AdvancedPassword,
			constant.PasswordTypeLock:     cashierSetting.LockPassword,
		}
	case constant.SourceAssistant:
		assistantSetting, err := s.GetAssistantSetting(ctx, nil)
		if err != nil {
			return false
		}
		passwordMap = map[string]string{
			constant.PasswordTypeAdvanced: assistantSetting.AdvancedPassword,
			constant.PasswordTypeLock:     assistantSetting.LockPassword,
		}
	case constant.SourceTablet:
		tabletSetting, err := s.GetTabletSetting(ctx, nil)
		if err != nil {
			return false
		}
		passwordMap = map[string]string{
			constant.PasswordTypeAdvanced: tabletSetting.AdvancedPassword,
		}
	case constant.SourceKitchen:
		kitchenSetting, err := s.GetKitchenSetting(ctx, ctx.GetCompanySetting(), nil)
		if err != nil {
			return false
		}
		passwordMap = map[string]string{
			constant.PasswordTypeAdvanced: kitchenSetting.AdvancedPassword,
		}
	}
	if truePassword, exits := passwordMap[typ]; exits {
		return password == truePassword
	}
	return false
}

// CheckUpdate 检查更新
func (s *Srv) CheckUpdate(ctx context.Context, appType int, brand string, language string) (resp.UpdateInfo, error) {
	// 不等于安卓就返回空
	userAgent := ctx.GetGin().GetHeader("User-Agent") + ";" + ctx.GetGin().GetHeader("platform") // 记录平台
	if utils.GetPlatform(userAgent) != 1 {
		return resp.UpdateInfo{}, errors.NewWithCode(constant.CodeSystemError, "当前平台暂不支持应用内更新")
	}
	//
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
		return resp.UpdateInfo{}, errors.WithMessage(errors.New("获取最新版本信息失败"), err.Error())
	}
	bodyBytes, _ := res.GetBodyAsByte()
	var updateData UpdateData
	if err := json.Unmarshal(bodyBytes, &updateData); err != nil {
		ctx.Log().Error("解析版本更新信息失败", zap.Error(err))
	}

	var updateLogMultilanguage dto.LocaleResponse
	var updateLog string
	if updateData.Data.UpdateLog != "" {
		err := json.Unmarshal([]byte(updateData.Data.UpdateLog), &updateLogMultilanguage)
		if err != nil {
			updateLog = updateData.Data.UpdateLog
		} else {
			updateLog = updateLogMultilanguage.GetLocale(language)
		}
	}
	return resp.UpdateInfo{
		VersionName:  updateData.Data.VersionName,
		ForcedUpdate: updateData.Data.ForcedUpdate,
		UpdateLog:    updateLog,
		DownloadURL:  updateData.Data.DownloadURL,
	}, nil
}

// EditAcceptOrderSetting 修改自动接单参数
func (s *Srv) EditAcceptOrderSetting(ctx context.Context, orderSetting req.UpdateAcceptOrderSetting) error { // 修改自动接单设置
	cashierSetting, err := s.GetCashierSetting(ctx, nil)
	if err != nil {
		return errors.WithMessage(err)
	}
	cashierSetting.IsAutoOrder = orderSetting.IsAutoOrder
	cashierSetting.AutoOrderLimit = orderSetting.AutoOrderLimit
	return s.UpdateSetting(ctx, constant.SettingCashier, cashierSetting)
}

// EditAcceptMemberOrderSetting 修改自动接单会员订单参数
func (s *Srv) EditAcceptMemberOrderSetting(ctx context.Context, orderSetting req.UpdateAcceptMemberOrderSetting) error { // 修改自动接单会员订单设置
	cashierSetting, err := s.GetCashierSetting(ctx, nil)
	if err != nil {
		return errors.WithMessage(err)
	}
	cashierSetting.IsAutoMemberOrder = orderSetting.IsAutoMemberOrder
	cashierSetting.AutoMemberOrderLimit = orderSetting.AutoMemberOrderLimit
	return s.UpdateSetting(ctx, constant.SettingCashier, cashierSetting)
}

// EditSystemSetting 修改系统设置
func (s *Srv) EditSystemSetting(ctx context.Context, systemSetting req.UpdateSystemSetting) error {
	cashierSetting, err := s.GetCashierSetting(ctx, nil)
	if err != nil {
		return errors.WithMessage(err)
	}
	cashierSetting.IsShowAssistantSoldOut = *systemSetting.IsShowAssistantSoldOut
	cashierSetting.IsShowScanSoldOut = *systemSetting.IsShowScanSoldOut
	cashierSetting.MenuShowSoldOut = strconv.Itoa(*systemSetting.MenuShowSoldOut)
	cashierSetting.MemberShowSoldOut = strconv.Itoa(*systemSetting.MemberShowSoldOut)
	if err := s.UpdateSetting(ctx, constant.SettingCashier, cashierSetting); err != nil {
		return errors.WithMessage(err)
	}
	businessSetting, err := s.GetBusinessSetting(ctx)
	if err != nil {
		return errors.WithMessage(err)
	}
	businessSetting.DishCardStyle = systemSetting.DishCardStyle
	if err := s.UpdateSetting(ctx, constant.SettingBusiness, businessSetting); err != nil {
		return errors.WithMessage(err)
	}
	tabletSetting, err := s.GetTabletSetting(ctx, nil)
	if err != nil {
		return errors.WithMessage(err)
	}
	tabletSetting.IsShowSoldOut = strconv.Itoa(*systemSetting.IsShowSoldOut)
	if err := s.UpdateSetting(ctx, constant.SettingTablet, tabletSetting); err != nil {
		return errors.WithMessage(err)
	}
	if err := repository.NewDeviceRepo(s.dbm.GetDB(ctx.GetCompanyUuid())).UpdateDevice(ctx.GetDeviceUuid(), map[string]any{
		"remark": systemSetting.DeviceRemark,
	}); err != nil {
		return errors2.ErrInternal
	}
	return nil
}

// GetCashierBaseSetting 获取收银端设置
func (s *Srv) GetCashierBaseSetting(ctx context.Context) (resp.CashierBaseSetting, error) {
	var settingResp resp.CashierBaseSetting
	cashierSetting, err := s.GetCashierSetting(ctx, nil)
	if err != nil {
		return settingResp, errors.WithMessage(err)
	}
	businessSetting, err := s.GetBusinessSetting(ctx)
	if err != nil {
		return settingResp, errors.WithMessage(err)
	}
	tabletSetting, err := s.GetTabletSetting(ctx, nil)
	if err != nil {
		return settingResp, errors.WithMessage(err)
	}

	clientVersion := ctx.GetGin().GetHeader("Version-Name")
	if clientVersion == "" {
		clientVersion = "0.0.0"
	}

	deviceRepo := repository.NewDeviceRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))

	device, err := deviceRepo.GetDevice(deviceRepo.WhereSn(ctx.GetDeviceSn()))
	if err != nil {
		return settingResp, errors2.ErrInternal
	}

	menuShowSoldOut, _ := strconv.Atoi(cashierSetting.MenuShowSoldOut)
	isShowSoldOut, _ := strconv.Atoi(tabletSetting.IsShowSoldOut)
	memberShowSoldOut, _ := strconv.Atoi(cashierSetting.MemberShowSoldOut)

	return resp.CashierBaseSetting{
		AcceptOrder: resp.AcceptOrderSetting{
			IsAutoOrder:    cashierSetting.IsAutoOrder,
			AutoOrderLimit: cashierSetting.AutoOrderLimit,
			IsAutoVoice:    cashierSetting.IsAutoVoice,
		},
		AcceptMemberOrder: resp.AcceptMemberOrderSetting{
			IsAutoMemberOrder:      cashierSetting.IsAutoMemberOrder,
			AutoMemberOrderLimit:   cashierSetting.AutoMemberOrderLimit,
			IsAutoVoiceMemberOrder: cashierSetting.IsAutoVoiceMemberOrder,
		},
		System: resp.SystemSetting{
			IsShowScanSoldOut:      cashierSetting.IsShowScanSoldOut,
			IsShowAssistantSoldOut: cashierSetting.IsShowAssistantSoldOut,
			MenuShowSoldOut:        menuShowSoldOut,
			MemberShowSoldOut:      memberShowSoldOut,
			DishCardStyle:          businessSetting.DishCardStyle,
			IsShowSoldOut:          isShowSoldOut,
			DefaultLanguage:        cashierSetting.DefaultLanguage,
			SecondLanguage:         cashierSetting.DefaultLanguage,
			DeviceId:               ctx.GetDeviceSn(),
			DeviceRemark:           device.Remark,
			ClientVersion:          clientVersion,
			ServerVersion:          utils.GetVersion(),
		},
		UsbPrinter: resp.UsbPrinterList{
			List:       make([]resp.UsbPrinter, 0),
			SelectedSn: "",
		},
	}, nil

}

// GetAcceptOrderSetting 获取接单设置
func (s *Srv) GetAcceptOrderSetting(ctx context.Context) (*resp.AcceptOrderSetting, error) {
	cashierSetting, err := s.GetCashierSetting(ctx, nil)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &resp.AcceptOrderSetting{
		IsAutoOrder:    cashierSetting.IsAutoOrder,
		AutoOrderLimit: cashierSetting.AutoOrderLimit,
		IsAutoVoice:    cashierSetting.IsAutoVoice,
	}, nil
}

// SymbolPosition 货币符号位置
func (s *Srv) SymbolPosition(ctx context.Context, amount float64) string {
	currencySetting, _ := s.GetCurrencySetting(ctx)
	if currencySetting.UnitPosition == "0" {
		return currencySetting.Unit + " " + utils.FormatAmount(amount)
	} else {
		return utils.FormatAmount(amount) + " " + currencySetting.Unit
	}
}

// UpdateSetting 更新设置
func (s *Srv) UpdateSetting(ctx context.Context, settingKey string, values any) error {
	value := utils.ToJson(values)
	if settingKey == constant.SettingStore {
		value = strings.ReplaceAll(value, "\"logo_url\"", "\"logoUrl\"")
		value = strings.ReplaceAll(value, "\"avatar_url\"", "\"avatarUrl\"")
	}
	db := ctx.GetDB()
	if db == nil {
		db = s.dbm.GetDB(ctx.GetCompanyUuid())
	}
	settingRepo := repository.NewSettingRepo(db)
	set := settingRepo.GetByKey(settingKey)
	if set.Key == "" {
		if _, err := settingRepo.Create(model.Setting{
			Key:    settingKey,
			Values: value,
		}); err != nil {
			return errors.New("更新设置失败")
		}
	} else {
		if err := settingRepo.Updates(settingKey, value); err != nil {
			return errors.New("更新设置失败")
		}
	}

	// 删除缓存
	s.cache.Del(fmt.Sprintf(s.cacheKey, ctx.GetCompanyUuid()))
	return nil
}

func (s *Srv) EditStoreSetting(ctx context.Context, storeSettingReq req.UpdateStoreSetting) error {
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	companyUuid := ctx.GetCompanyUuid()
	companyDB := s.dbm.GetDB(companyUuid)

	// 门店设置
	storeSetting, err := s.GetStoreSetting(ctx)
	if err != nil {
		return errors.WithMessage(err)
	}

	// 公司设置
	companySetting, err := s.GetCompanySetting(ctx)
	if err != nil {
		return errors.WithMessage(err)
	}
	company := ctx.GetCompany()

	// 时区在时区列表中
	timeZoneList := storeSetting.TimeZoneList
	if !slices.ContainsFunc(timeZoneList, func(item setting.TimeZoneItem) bool {
		return item.Key == storeSettingReq.TimeZone
	}) {
		return errors.New("时区不存在")
	}

	// 传递过来的语言，必须是companySetting.GetLanguages() 且 Name和Value都在constant.Languages中的语言
	for _, language := range storeSettingReq.Language {
		if !slices.Contains(companySetting.GetLanguages(), language.Name) || !slices.ContainsFunc(constant.Languages, func(item constant.LanguageItem) bool {
			return item.Name == language.Name && item.Value == language.Value
		}) {
			return errors.New("语言不存在")
		}
	}

	// 去掉logoUrl的域名
	storeSettingReq.LogoUrl = "/" + strings.TrimLeft(utils.RemoveDomain(storeSettingReq.LogoUrl), "/")
	storeSetting.AvatarURL = "/" + strings.TrimLeft(utils.RemoveDomain(storeSetting.AvatarURL), "/")

	copier.CopyWithOption(&storeSetting, storeSettingReq, copier.Option{IgnoreEmpty: true})

	storeSetting.LogoURL = storeSettingReq.LogoUrl
	storeSetting.Company = storeSettingReq.CompanyName

	// ##### 处理 cashier tablet h5 kitchen assistant printer 各端的语言设置 #####
	// ##### 1、处理 cashier 设置 #####
	cashierSetting, err := s.GetCashierSetting(ctx, storeSettingReq.Language)
	if err != nil {
		return errors.WithMessage(err)
	}
	{
		// 去掉cashierSetting.Language中不在storeSettingReq.Language中的语言
		for _, language := range cashierSetting.Language {
			if !slices.ContainsFunc(storeSettingReq.Language, func(item dto.LanguageItem) bool {
				return item.Name == language
			}) {
				cashierSetting.Language = slices.DeleteFunc(cashierSetting.Language, func(item string) bool {
					return item == language
				})
			}
		}
		if len(cashierSetting.Language) == 0 {
			cashierSetting.Language = []string{storeSettingReq.Language[0].Name}
		}
		// 如果 cashierSetting.DefaultLanguage 不在 cashierSetting.Language 中，则设置为 cashierSetting.Language 中的第一个语言
		if !slices.ContainsFunc(cashierSetting.Language, func(item string) bool {
			return item == cashierSetting.DefaultLanguage
		}) {
			cashierSetting.DefaultLanguage = cashierSetting.Language[0]
		}
		// 清除 cashierSetting.LanguageList 中的数据
		cashierSetting.LanguageList = []dto.LanguageItem{}
	}

	// ##### 2、处理 tablet 设置 #####
	tabletSetting, err := s.GetTabletSetting(ctx, storeSettingReq.Language)
	if err != nil {
		return errors.WithMessage(err)
	}
	{
		// 去掉tabletSetting.Language中不在storeSettingReq.Language中的语言
		for _, language := range tabletSetting.Language {
			if !slices.ContainsFunc(storeSettingReq.Language, func(item dto.LanguageItem) bool {
				return item.Name == language
			}) {
				tabletSetting.Language = slices.DeleteFunc(tabletSetting.Language, func(item string) bool {
					return item == language
				})
			}
		}
		if len(tabletSetting.Language) == 0 {
			tabletSetting.Language = []string{storeSettingReq.Language[0].Name}
		}
		// 如果 tabletSetting.DefaultLanguage 不在 tabletSetting.Language 中，则设置为 tabletSetting.Language 中的第一个语言
		if !slices.ContainsFunc(tabletSetting.Language, func(item string) bool {
			return item == tabletSetting.DefaultLanguage
		}) {
			tabletSetting.DefaultLanguage = tabletSetting.Language[0]
		}
		tabletSetting.LanguageList = nil
	}

	// ##### 3、处理 h5 设置 #####
	h5Setting, err := s.GetH5Setting(ctx, storeSettingReq.Language)
	if err != nil {
		return errors.WithMessage(err)
	}
	{
		// 去掉h5Setting.Language中不在storeSettingReq.Language中的语言
		for _, language := range h5Setting.Language {
			if !slices.ContainsFunc(storeSettingReq.Language, func(item dto.LanguageItem) bool {
				return item.Name == language
			}) {
				h5Setting.Language = slices.DeleteFunc(h5Setting.Language, func(item string) bool {
					return item == language
				})
			}
		}
		if len(h5Setting.Language) == 0 {
			h5Setting.Language = []string{storeSettingReq.Language[0].Name}
		}
		// 如果 h5Setting.DefaultLanguage 不在 h5Setting.Language 中，则设置为 h5Setting.Language 中的第一个语言
		if !slices.ContainsFunc(h5Setting.Language, func(item string) bool {
			return item == h5Setting.DefaultLanguage
		}) {
			h5Setting.DefaultLanguage = h5Setting.Language[0]
		}
		h5Setting.LanguageList = nil
	}

	// ##### 4、处理 kitchen 设置 #####
	kitchenSetting, err := s.GetKitchenSetting(ctx, companySetting, storeSettingReq.Language)
	if err != nil {
		return errors.WithMessage(err)
	}
	{
		// 去掉kitchenSetting.Language中不在storeSettingReq.Language中的语言
		for _, language := range kitchenSetting.Language {
			if !slices.ContainsFunc(storeSettingReq.Language, func(item dto.LanguageItem) bool {
				return item.Name == language
			}) {
				kitchenSetting.Language = slices.DeleteFunc(kitchenSetting.Language, func(item string) bool {
					return item == language
				})
			}
		}
		if len(kitchenSetting.Language) == 0 {
			kitchenSetting.Language = []string{storeSettingReq.Language[0].Name}
		}
		// 如果 kitchenSetting.DefaultLanguage 不在 kitchenSetting.Language 中，则设置为 kitchenSetting.Language 中的第一个语言
		if !slices.ContainsFunc(kitchenSetting.Language, func(item string) bool {
			return item == kitchenSetting.DefaultLanguage
		}) {
			kitchenSetting.DefaultLanguage = kitchenSetting.Language[0]
		}
		kitchenSetting.LanguageList = nil
	}

	// ##### 5、处理 assistant 设置 #####
	assistantSetting, err := s.GetAssistantSetting(ctx, storeSettingReq.Language)
	if err != nil {
		return errors.WithMessage(err)
	}
	{
		// 去掉assistantSetting.Language中不在storeSettingReq.Language中的语言
		for _, language := range assistantSetting.Language {
			if !slices.ContainsFunc(storeSettingReq.Language, func(item dto.LanguageItem) bool {
				return item.Name == language
			}) {
				assistantSetting.Language = slices.DeleteFunc(assistantSetting.Language, func(item string) bool {
					return item == language
				})
			}
		}
		if len(assistantSetting.Language) == 0 {
			assistantSetting.Language = []string{storeSettingReq.Language[0].Name}
		}
		// 如果 assistantSetting.DefaultLanguage 不在 assistantSetting.Language 中，则设置为 assistantSetting.Language 中的第一个语言
		if !slices.ContainsFunc(assistantSetting.Language, func(item string) bool {
			return item == assistantSetting.DefaultLanguage
		}) {
			assistantSetting.DefaultLanguage = assistantSetting.Language[0]
		}
		assistantSetting.LanguageList = nil
	}

	// ##### 6、处理 printer 设置 #####
	printerSetting, err := s.GetPrinterSetting(ctx, storeSettingReq.Language)
	if err != nil {
		return errors.WithMessage(err)
	}
	{
		// 去掉printerSetting.Language中不在storeSettingReq.Language中的语言
		for _, language := range printerSetting.Language {
			if !slices.ContainsFunc(storeSettingReq.Language, func(item dto.LanguageItem) bool {
				return item.Name == language
			}) {
				printerSetting.Language = slices.DeleteFunc(printerSetting.Language, func(item string) bool {
					return item == language
				})
			}
		}
		if len(printerSetting.Language) == 0 {
			printerSetting.Language = []string{storeSettingReq.Language[0].Name}
		}
		// 如果 printerSetting.DefaultLanguage 不在 printerSetting.Language 中，则设置为 printerSetting.Language 中的第一个语言
		if !slices.ContainsFunc(printerSetting.Language, func(item string) bool {
			return item == printerSetting.DefaultLanguage
		}) {
			printerSetting.DefaultLanguage = printerSetting.Language[0]
		}
		// 如果 printerSetting.KitchenLanguage 不在 printerSetting.Language 中，则设置为 printerSetting.Language 中的第一个语言
		if !slices.ContainsFunc(printerSetting.Language, func(item string) bool {
			return item == printerSetting.KitchenLanguage
		}) {
			printerSetting.KitchenLanguage = printerSetting.Language[0]
		}
		printerSetting.LanguageList = nil
	}

	updateCompany := map[string]any{
		"name": storeSettingReq.Name,
		"logo": storeSettingReq.LogoUrl,
	}
	updateCompanySetting := map[string]any{
		"timezone":    storeSettingReq.TimeZone,
		"link_phone":  storeSettingReq.Phone,
		"address":     storeSettingReq.Address,
		"coordinates": storeSettingReq.Coordinates,
	}

	err = companyDB.Transaction(func(tx *gorm.DB) error {
		// 保存到saas.company_setting\saas.company\商家company_setting\商家company表
		err := saasDB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&model.Company{}).Where("uuid = ?", companyUuid).Updates(updateCompany).Error; err != nil {
				return errors.WithMessage(errors.New("保存saas.company设置失败"), err.Error())
			}
			if err := tx.Model(&model.CompanySetting{}).Where("company_uuid = ?", companyUuid).Updates(updateCompanySetting).Error; err != nil {
				return errors.WithMessage(errors.New("保存saas.company_setting设置失败"), err.Error())
			}
			return nil
		})
		if err != nil {
			return err
		}
		if err := tx.Model(&model.Company{}).Where("uuid = ?", companyUuid).Updates(updateCompany).Error; err != nil {
			return errors.WithMessage(errors.New("保存商家company设置失败"), err.Error())
		}
		if err := tx.Model(&model.CompanySetting{}).Where("company_uuid = ?", companyUuid).Updates(updateCompanySetting).Error; err != nil {
			return errors.WithMessage(errors.New("保存store设置失败"), err.Error())
		}

		ctx.SetDB(tx)
		// 保存设置到store_setting表
		if err := s.UpdateSetting(ctx, constant.SettingStore, storeSetting); err != nil {
			return errors.WithMessage(errors.New("保存store设置失败"), err.Error())
		}
		if err := s.UpdateSetting(ctx, constant.SettingCashier, cashierSetting); err != nil {
			return errors.WithMessage(errors.New("保存cashier设置失败"), err.Error())
		}
		if err := s.UpdateSetting(ctx, constant.SettingTablet, tabletSetting); err != nil {
			return errors.WithMessage(errors.New("保存tablet设置失败"), err.Error())
		}
		if err := s.UpdateSetting(ctx, constant.SettingH5, h5Setting); err != nil {
			return errors.WithMessage(errors.New("保存h5设置失败"), err.Error())
		}
		if err := s.UpdateSetting(ctx, constant.SettingKitchen, kitchenSetting); err != nil {
			return errors.WithMessage(errors.New("保存kitchen设置失败"), err.Error())
		}
		if err := s.UpdateSetting(ctx, constant.SettingAssistant, assistantSetting); err != nil {
			return errors.WithMessage(errors.New("保存assistant设置失败"), err.Error())
		}
		if err := s.UpdateSetting(ctx, constant.SettingPrinter, printerSetting); err != nil {
			return errors.WithMessage(errors.New("保存printer设置失败"), err.Error())
		}

		// 总店修改商家名称，更新Headquarters - Supplier供应商名称
		if company.IsOpenErp() && companySetting.IsHeadquarter() {
			erpSrv := erp.NewIErpSrv(s.dbm)
			err = erpSrv.UpdateSupplier(ctx.GetContext(), req.UpdateSupplierReq{
				CreateSupplierReq: req.CreateSupplierReq{
					SiteCode:     companySetting.ErpnextSiteCode,
					CompanyAbbr:  companySetting.ErpnextCompanyAbbr,
					Branch:       companySetting.ErpnextBranchName,
					SupplierName: storeSettingReq.Name,
				},
				CompanyUuid: companyUuid,
				Name:        constant.ErpHeadquartersSupplierCode,
			})
			if err != nil {
				return errors.WithMessage(errors.New("更新Headquarters - Supplier供应商名称失败"), err.Error())
			}
			if err := tx.Model(&model.Supplier{}).Where("headquarter_uuid = 0 AND erp_code = ?", constant.ErpHeadquartersSupplierCode).Update("name", storeSettingReq.Name).Error; err != nil {
				return errors.WithMessage(errors.New("更新Headquarters - Supplier供应商名称失败"), err.Error())
			}
		}

		return nil
	})

	if err != nil {
		logger.Logger.Error("保存设置失败", zap.Error(err))
		return errors.WithMessage(errors.New("保存失败"), err.Error())
	}

	// 删除系统设置缓存
	s.cache.Del(fmt.Sprintf(s.cacheKey, companyUuid))
	// 删除全局缓存
	tc := cache.NewTaggedCache(s.cache)
	tc.TagClear("common_get_settingLanguages")
	s.cache.Del(fmt.Sprintf("{common_get_settingLanguages}_common_setting_languages%d", companyUuid))
	tc.TagClear("cashier")

	// 推送配置更新
	utils.Go(func() {
		websocket.PushClient(companyUuid, websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_CONFIG, map[string]any{
			"update_time": time.Now().Unix(),
		})
	})

	return nil
}

func (s *Srv) EditBusinessSetting(ctx context.Context, businessSettingReq req.UpdateBusinessSetting) error {
	companyUuid := ctx.GetCompanyUuid()
	companySetting, err := s.GetCompanySetting(ctx)
	if err != nil {
		return errors.WithMessage(err)
	}
	businessSetting, err := s.GetBusinessSetting(ctx)
	if err != nil {
		return errors.WithMessage(err)
	}

	// 备份旧的businessSetting
	oldBusinessSetting := businessSetting

	// 判断菜品卡片样式是否更新，如果更新，则更新菜品卡片样式最后更新时间
	oldDishCardStyle := businessSetting.DishCardStyle
	if oldDishCardStyle != businessSetting.DishCardStyle {
		businessSetting.DishCardStyleTime = strconv.Itoa(int(time.Now().Unix()))
	}

	// 判断是否开启外送功能
	if businessSetting.DeliveryPriceRatio != businessSettingReq.DeliveryPriceRatio && companySetting.DeliveryStatus != 1 {
		return errors.New("当前没有权限使用此功能")
	}

	// 更新businessSetting
	copier.CopyWithOption(&businessSetting, businessSettingReq, copier.Option{IgnoreEmpty: true})

	// 删除不需要的列表字段
	businessSetting.ZeroingMethodList = []setting.ZeroingMethodItem{}
	businessSetting.CheckoutZeroingMethodList = []setting.CheckoutZeroingMethodItem{}
	businessSetting.GiftMethodList = []setting.GiftMethodItem{}
	businessSetting.FreeMethodList = []setting.FreeMethodItem{}

	// is_batch为“1”时，必须至少有一个分批类型
	if businessSetting.IsBatch == "1" {
		batchTagNum, err := repository.NewBatchTagRepo(s.dbm.GetDB(companyUuid)).GetBatchTagCount()
		if err != nil {
			return errors.WithMessage(err)
		}
		if batchTagNum == 0 {
			return errors.New("开启分批送厨商品时，必须至少有一个分批类型")
		}
	}

	// 覆盖oldBusinessSetting
	copier.CopyWithOption(&oldBusinessSetting, businessSetting, copier.Option{IgnoreEmpty: true})
	// 保存设置到 business_setting 表
	err = s.UpdateSetting(ctx, constant.SettingBusiness, oldBusinessSetting)
	if err != nil {
		return errors.WithMessage(err)
	}

	// 删除系统设置缓存
	s.cache.Del(fmt.Sprintf(s.cacheKey, companyUuid))
	// 删除全局缓存
	tc := cache.NewTaggedCache(s.cache)
	tc.TagClear("common_get_settingLanguages")
	s.cache.Del(fmt.Sprintf("{common_get_settingLanguages}_common_setting_languages%d", companyUuid))
	tc.TagClear("cashier")

	// 推送配置更新
	utils.Go(func() {
		websocket.PushClient(companyUuid, websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_CONFIG, map[string]any{
			"update_time": time.Now().Unix(),
		})
	})

	// 将本店非当前安全库存类型的预警记录删除
	err = s.dbm.GetDB(companyUuid).Model(&model.MaterialStockAlertLog{}).Where("alert_type != ?", businessSetting.SafetyStockType).Update("delete_time", time.Now().Unix()).Error
	if err != nil {
		logger.Logger.Error("删除本店非当前安全库存类型的预警记录失败", zap.Error(err))
	}

	return nil
}

func (s *Srv) GetShopBusinessSetting(ctx context.Context) (setting.ShopBusiness, error) {
	companyUuid := ctx.GetCompanyUuid()
	db := s.dbm.GetDB(companyUuid)

	businessSetting, err := s.GetBusinessSetting(ctx)
	if err != nil {
		return setting.ShopBusiness{}, errors.WithMessage(err)
	}

	var freeReasonCount, returnFoodReasonCount, orderRemarkCount int64
	err = db.Model(&model.FreeReason{}).Scopes(repository.NotDeleted).Select("count(*)").Scan(&freeReasonCount).Error
	if err != nil {
		return setting.ShopBusiness{}, errors.WithMessage(err)
	}
	err = db.Model(&model.ReturnFoodReason{}).Scopes(repository.NotDeleted).Select("count(*)").Scan(&returnFoodReasonCount).Error
	if err != nil {
		return setting.ShopBusiness{}, errors.WithMessage(err)
	}
	err = db.Model(&model.OrderRemark{}).Scopes(repository.NotDeleted).Select("count(*)").Scan(&orderRemarkCount).Error
	if err != nil {
		return setting.ShopBusiness{}, errors.WithMessage(err)
	}

	var headquarterRequiredParentCompanyApproval, headquarterViaParentCompanyWarehouse string
	companySetting := ctx.GetCompanySetting()
	if companySetting.HeadquarterUuid > 0 {
		ctx2 := ctx.Copy()
		ctx2.SetCompanyUuid(companySetting.HeadquarterUuid)
		headquarterBusinessSetting, err := s.GetBusinessSetting(ctx2)
		if err != nil {
			return setting.ShopBusiness{}, errors.WithMessage(err)
		}
		headquarterRequiredParentCompanyApproval = headquarterBusinessSetting.RequiredParentCompanyApproval
		headquarterViaParentCompanyWarehouse = headquarterBusinessSetting.ViaParentCompanyWarehouse
	}

	return setting.ShopBusiness{
		Business:                                 businessSetting,
		FreeReasonCount:                          int(freeReasonCount),
		ReturnFoodReasonCount:                    int(returnFoodReasonCount),
		OrderRemarkCount:                         int(orderRemarkCount),
		HeadquarterRequiredParentCompanyApproval: headquarterRequiredParentCompanyApproval,
		HeadquarterViaParentCompanyWarehouse:     headquarterViaParentCompanyWarehouse,
	}, nil
}

func (s *Srv) GetMenuQrcode(ctx context.Context) (string, error) {
	businessSetting, err := s.GetBusinessSetting(ctx)
	if err != nil {
		return "", errors.WithMessage(err)
	}
	return viper.GetString("MENU_BASE_URL") + "/home?token=" + s.getMenuQrcodeToken(ctx, businessSetting), nil
}

func (s *Srv) getMenuQrcodeToken(ctx context.Context, businessSetting setting.Business) string {
	type Qrcode struct {
		CompanyUuid uint64 `json:"a"`
		Qrcode      string `json:"q"`
	}
	qrcode := Qrcode{
		CompanyUuid: ctx.GetCompanyUuid(),
		Qrcode:      businessSetting.QrCode,
	}
	qrcodeString := utils.ToJson(qrcode)
	hash := md5.Sum([]byte(qrcodeString))
	token := fmt.Sprintf("%x.%s", hash, base64.StdEncoding.EncodeToString([]byte(qrcodeString)))

	return base64.StdEncoding.EncodeToString([]byte(token))
}

// GetPaymentMethodList 获取支付方式列表
func (s *Srv) GetPaymentMethodList(ctx context.Context) setting.PaymentMethodListResp {
	commonRepo := repository.NewCommonRepo()
	paymentRepo := repository.NewPaymentMethodRepo(ctx.GetDB())
	paymentMethodList := paymentRepo.GetPaymentMethodList(
		commonRepo.WhereBySoftDelete(),
		commonRepo.SortWithSort("asc"),
		commonRepo.SortWithCreateTime("desc"),
	)

	list := make([]setting.PaymentMethod, 0, len(paymentMethodList))
	for _, paymentMethod := range paymentMethodList {
		list = append(list, setting.PaymentMethod{
			Uuid: paymentMethod.Uuid,
			Name: paymentMethod.Name,
		})
	}
	return setting.PaymentMethodListResp{List: list}
}
