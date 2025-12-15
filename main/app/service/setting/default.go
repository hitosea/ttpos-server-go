package setting

import (
	"encoding/json"
	"strconv"
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/utils"

	"github.com/spf13/viper"
)

func (s *Srv) getIPAndPort() (string, string) {
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

// 默认收银机设置
func (s *Srv) getDefaultCashier(languageList []dto.LanguageItem) setting.Cashier {
	var defaultLanguage = i18n.LanguageEN
	if len(languageList) > 0 {
		defaultLanguage = languageList[0].Name
	}
	return setting.Cashier{
		CashierResp: setting.CashierResp{
			Carousel:   []setting.CarouselItem{}, // 上传后的轮播内容url（图片 + 视频）
			IsAutoSend: "0",                      // 收银结账自动送厨房 0-关闭 1-开启
			OrderMethod: setting.OrderMethod{
				IsCashierOrder: "1",
				IsTableOrder:   "1",
			}, // 用餐方式 收银-is_cashier_order（0-关闭 1-开启） 桌台-is_table_order（0-关闭 1-开启）
			IsRemainColor: "1",                            // 是否开启剩余时长颜色 0-关闭 1-开启
			RemainColor:   []string{"#E50028", "#F2A000"}, // 剩余时长颜色 10分钟-红色(#E50028) 20分钟-黄色(#F2A000)

			IsOpenCashierPassword: "1", // 是否开启钱箱密码 0-关闭 1-开启

			IsAutoLockScreen:        "1",                       // 是否开启自动锁屏 0-关闭 1-开启
			AutoLockScreen:          "300",                     // 自动锁屏（秒），默认5分钟
			IsShowScanSoldOut:       1,                         // 扫码点餐是否显示售罄商品 0-不显示 1-显示
			IsShowAssistantSoldOut:  1,                         // 点餐助手是否显示售罄商品 0-不显示 1-显示
			LanguageList:            languageList,              // 语言列表
			Language:                []string{defaultLanguage}, // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
			DefaultLanguage:         defaultLanguage,           // 默认语言
			IsAutoOrder:             "0",                       // 是否自动接单
			AutoOrderLimit:          "1000",                    // 自动接单金额上限
			IsAutoVoice:             "0",                       // 是否开启自动接单语音播报
			IsAutoMemberOrder:       "0",                       // 是否自动接单会员订单 0-关闭 1-开启
			AutoMemberOrderLimit:    "1000",                    // 自动接单会员订单金额上限
			IsAutoVoiceMemberOrder:  "0",                       // 是否开启自动接单会员订单语音播报
			MenuShowSoldOut:         "1",                       // 是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
			MemberShowSoldOut:       "1",                       // 是否显示会员端售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
			NoOrderCarouselInterval: "10",                      // 未点餐时轮播间隔(秒)
			OrderDisplayMode:        "order",                   // 点餐时展示模式 carousel/order/order_carousel
			OrderCarouselInterval:   "10",                      // 点餐时轮播间隔(秒)
		},
		AdvancedPassword: "666888", // 高级设置密码
		CashierPassword:  "666888", // 钱箱密码
		LockPassword:     "666888", // 锁屏密码
	}
}

// 默认平板端设置
func (s *Srv) getDefaultTablet(languageList []dto.LanguageItem) setting.Tablet {
	var defaultLanguage = i18n.LanguageEN
	ip, port := s.getIPAndPort()
	if len(languageList) > 0 {
		defaultLanguage = languageList[0].Name
	}
	// DefaultTabletSetting 平板端设置
	return setting.Tablet{
		TabletResp: setting.TabletResp{
			Carousel:           []setting.CarouselItem{}, // 上传后的轮播内容url（图片 + 视频）
			IsCallService:      "1",                      // 是否开启呼叫服务员 0-关闭 1-开启
			IsCustomerOrder:    "1",                      // 是否开启顾客自助下单 0-关闭 1-开启
			IsVoiceRemind:      "1",                      // 是否开启声音提醒 0-关闭 1-开启
			IsBuffetOrderLimit: "0",                      // 是否开启自助餐下单限制 0-关闭 1-开启
			BuffetOrderLimit: setting.BuffetOrderLimit{
				IsLimitTime: "1", // 是否开启时间限制 0-关闭 1-开启
				LimitTime:   "0", // 时间限制（分钟）
				IsLimitNum:  "1", // 是否开启数量限制 0-关闭 1-开启
				LimitNum:    "0", // 数量限制
			},
			IsOrderLimit: "0", // 是否开启非自助餐下单限制 0-关闭 1-开启
			OrderLimit: setting.OrderLimit{
				IsLimitTime: "1", // 是否开启时间限制 0-关闭 1-开启
				LimitTime:   "0", // 时间限制（分钟）
				IsLimitNum:  "1", // 是否开启数量限制 0-关闭 1-开启
				LimitNum:    "0", // 数量限制
			},
			Server:          setting.Server{IP: ip, Port: port}, // 平板服务器连接
			LanguageList:    languageList,                       // 语言列表
			Language:        []string{defaultLanguage},          // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
			DefaultLanguage: defaultLanguage,                    // 默认语言
		},
		AdvancedPassword: "666888", // 高级设置密码
	}
}

// 默认自助点餐机设置
func (s *Srv) getDefaultKioskSetting(languageList []dto.LanguageItem) setting.Kiosk {
	var defaultLanguage = i18n.LanguageEN
	if len(languageList) > 0 {
		defaultLanguage = languageList[0].Name
	}

	// 默认勾选所有语言
	commonLanguages := make([]string, 0, len(languageList))
	for _, lang := range languageList {
		commonLanguages = append(commonLanguages, lang.Name)
	}

	return setting.Kiosk{
		KioskResp: setting.KioskResp{
			AdvancedPassword:  "666888",                 // 高级密码（默认666888）
			CallWaiterEnabled: 1,                        // 呼叫服务员开关（默认1-开启）
			LanguageList:      languageList,             // 语言列表
			Language:          commonLanguages,          // 已设置的语言（默认所有语言）
			DefaultLanguage:   defaultLanguage,          // 默认语言（默认语言1）
			Carousel:          []setting.CarouselItem{}, // 轮播内容（默认空数组）
		},
	}
}

// 默认扫码H5设置
func (s *Srv) getDefaultH5(languageList []dto.LanguageItem) setting.H5 {
	var defaultLanguage = i18n.LanguageEN
	if len(languageList) > 0 {
		defaultLanguage = languageList[0].Name
	}
	return setting.H5{
		IsCallService:      "1", // 是否开启呼叫服务员 0-关闭 1-开启
		IsCustomerOrder:    "1", // 是否开启顾客自助下单 0-关闭 1-开启
		IsVoiceRemind:      "1", // 是否开启声音提醒 0-关闭 1-开启
		IsShowScanSoldOut:  0,   // 是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
		IsBuffetOrderLimit: "0", // 自助餐下单限制 是否开启自助餐下单限制 0-关闭 1-开启
		BuffetOrderLimit: setting.BuffetOrderLimit{
			IsLimitTime: "1", // 是否开启时间限制 0-关闭 1-开启
			LimitTime:   "0", // 时间限制（分钟）
			IsLimitNum:  "1", // 是否开启数量限制 0-关闭 1-开启
			LimitNum:    "0", // 数量限制
		},
		IsOrderLimit: "0", // 非自助餐下单限制 是否开启非自助餐下单限制 0-关闭 1-开启
		OrderLimit: setting.OrderLimit{
			IsLimitTime: "1", // 是否开启时间限制 0-关闭 1-开启
			LimitTime:   "0", // 时间限制（分钟）
			IsLimitNum:  "1", // 是否开启数量限制 0-关闭 1-开启
			LimitNum:    "0", // 数量限制
		},
		LanguageList:    languageList,              // 语言列表
		Language:        []string{defaultLanguage}, // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
		DefaultLanguage: defaultLanguage,           // 默认语言
	}
}

// 默认厨显设置
func (s *Srv) getDefaultKitchen(languageList []dto.LanguageItem) setting.Kitchen {
	var defaultLanguage = i18n.LanguageEN
	ip, port := s.getIPAndPort()
	if len(languageList) > 0 {
		defaultLanguage = languageList[0].Name
	}
	return setting.Kitchen{
		KitchenResp: setting.KitchenResp{
			IsOpen:              "1",                                // 是否开启厨显功能 0关闭 1开启
			IsComeDish:          "1",                                // 是否开启来菜提醒 0-关闭 1-开启
			IsCallService:       "1",                                // 是否开启顾客呼叫提醒 0-关闭 1-开启
			Server:              setting.Server{IP: ip, Port: port}, // 厨显服务器连接
			IsWaitColor:         "0",                                // 是否开启等待时长颜色 0-关闭 1-开启
			WaitColor:           []string{},                         // 时长颜色（旧格式：["red", "yellow"]）
			WaitTimeColorRanges: []setting.WaitTimeColorRange{},     // 等待时长颜色区间配置（新格式）
			LanguageList:        languageList,                       // 语言列表
			Language:            []string{defaultLanguage},          // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
			DefaultLanguage:     defaultLanguage,                    // 默认语言
			IsSmartKitchen:      "0",                                // 是否开启智能后厨 0-关闭 1-开启
		},
		AdvancedPassword: "666888", // 高级设置密码
	}
}

// 默认自助餐设置
func (s *Srv) getDefaultBuffet() setting.Buffet {
	return setting.Buffet{
		IsOpen:                   "1",                      // 是否开启自助餐 0-关闭 1-开启
		TabletEndTime:            "5",                      // 平板结束时间提醒（分）
		IsRemainContinue:         "0",                      // 剩余xx分不可继续点餐开关 0-关闭 1-开启
		RemainContinueTime:       "10",                     // 剩余xx分不可继续点餐
		RemainContinueNoticeTime: "10",                     // 剩余xx分提醒不可继续点餐
		IsBuyContinue:            "0",                      // 非自助餐商品到时是否能继续选购 0-关闭 1-开启
		IsAddClock:               "0",                      // 是否开启加钟 0-关闭 1-开启
		AddClock:                 []setting.AddClockItem{}, // 名称 - 加钟时间（分）- 价格
	}
}

// 默认支付方式设置
func (s *Srv) getDefaultPayment() setting.Payment {
	return setting.Payment{
		IsCash:    "0", // 是否开启现金支付 0-关闭 1-开启
		IsBalance: "0", // 是否开启余额支付 0-关闭 1-开启
		IsOther:   "1", // 是否开启其他方式支付 0-关闭 1-开启
	}
}

// 默认门店业务设置
func (s *Srv) getDefaultBusiness(language string) setting.Business {
	return setting.Business{
		ZeroingMethodList: []setting.ZeroingMethodItem{{
			Key:  "0",
			Name: i18n.Translate(language, "实款实收"),
		}, {
			Key:  "1",
			Name: i18n.Translate(language, "抹分"),
		}, {
			Key:  "2",
			Name: i18n.Translate(language, "抹角"),
		}, {
			Key:  "3",
			Name: i18n.Translate(language, "四舍五入保留一位小数"),
		}, {
			Key:  "4",
			Name: i18n.Translate(language, "四舍五入到整数"),
		}}, // 优惠折扣自动抹零方式列表
		CheckoutZeroingMethodList: []setting.CheckoutZeroingMethodItem{{
			Key:  "0",
			Name: i18n.Translate(language, "实款实收"),
		}, {
			Key:  "1",
			Name: i18n.Translate(language, "抹分"),
		}, {
			Key:  "2",
			Name: i18n.Translate(language, "抹角"),
		}, {
			Key:  "5",
			Name: i18n.Translate(language, "抹元"),
		}}, // 结账自动抹零方式列表
		ZeroingMethod:         "0", // 优惠折扣自动抹零方式
		CheckoutZeroingMethod: "0", // 结账自动抹零方式
		GiftMethodList: []setting.GiftMethodItem{{
			Key:  "10",
			Name: i18n.Translate(language, "计入总销售额、优惠折扣"),
		}, {
			Key:  "20",
			Name: i18n.Translate(language, "不计入总销售额、优惠折扣"),
		}}, // 赠菜计算方式列表
		GiftMethod: "10", // 赠菜计算方式，10-计入总销售额、优惠折扣 20-不计入总销售额、优惠折扣
		FreeMethodList: []setting.FreeMethodItem{{
			Key:  "10",
			Name: i18n.Translate(language, "计入总销售额、优惠折扣、服务费、税费"),
		}, {
			Key:  "20",
			Name: i18n.Translate(language, "不计入总销售额、优惠折扣、服务费、税费"),
		}}, // 免单计算方式列表
		FreeMethod:         "10",          // 免单计算方式，10-计入总销售额、优惠折扣、服务费、税费 20-不计入总销售额、优惠折扣、服务费、税费
		DiscountMethod:     "10",          // 折扣计算方式 10-按百分比 20-直接减免
		QrCode:             "123456",      // 电子菜单二维码校验失效值，6位数数字
		NoClearTable:       "0",           // 结账后不清台 0-清台 1-不清台
		IsNeedPassword:     "1",           // 取消订单/退菜 0-无需密码 1-需要密码
		DishCardStyle:      "0",           // 菜品卡片样式 0-无图模式 1-图片模式
		DishCardStyleTime:  "0",           // 菜品卡片样式最后更新时间
		IsInvoice:          "0",           // 开票信息 0-不需要填写 1-需要填写
		OpeningHours:       "00:00-23:59", // 营业时间 00:00-23:59
		DeliveryPriceRatio: 100,           // 外送商品价格和商品原价比例
		StartSerialNo:      "0001",        // 开始序列号
		IsBatch:            "0",           // 是否是分批商品 0-否 1-是
		BatchProductUuids:  []uint64{},    // 分批商品UUID列表
		BatchTagNum:        0,             // 分批类型数量
		SafetyStockType:    "1",           // 安全库存类型 1-门店纬度 2-仓库纬度，默认为1

		RequiredParentCompanyApproval: "0",                           // 调拨规则-经过上级门店审批 "0"-否 "1"-是, 总部和上级(有下级门店)支持此选项
		ViaParentCompanyWarehouse:     "0",                           // 调拨规则-经过上级门店仓库 "0"-否 "1"-是, 总部和上级(有下级门店)支持此选项
		BatchCookingMode:              constant.BatchCookingModePost, // 分批送厨模式: "pre" 前置 / "post" 后置，默认 "post"

		EnableOrderSource: "0", // 外卖功能开关 0-关闭 1-开启
		EnableNationality: "0", // 国籍功能开关 0-关闭 1-开启

		DiscountNeedPassword:       "0",        // 折扣操作是否需要密码 0-否 1-是
		DiscountAuthorizedStaffIds: []uint64{}, // 折扣操作授权员工ID列表
		RefundNeedPassword:         "0",        // 退款操作是否需要密码 0-否 1-是
		RefundAuthorizedStaffIds:   []uint64{}, // 退款操作授权员工ID列表
	}

}

// 默认点餐助手设置
func (s *Srv) getDefaultAssistant(languageList []dto.LanguageItem) setting.Assistant {
	var defaultLanguage = i18n.LanguageEN
	ip, port := s.getIPAndPort()
	if len(languageList) > 0 {
		defaultLanguage = languageList[0].Name
	}
	return setting.Assistant{
		AssistantResp: setting.AssistantResp{
			Server:           setting.Server{IP: ip, Port: port}, // 服务器连接
			IsRemainColor:    "1",                                // 是否开启剩余时长颜色 0-关闭 1-开启
			RemainColor:      []string{"#E50028", "#F2A000"},     // 剩余时长颜色 10分钟-红色(#E50028) 20分钟-黄色(#F2A000)
			DefaultMode:      "0",                                // 默认模式 0-服务员模式 1-顾客模式
			IsAutoLockScreen: "1",                                // 是否开启自动锁屏 0-关闭 1-开启
			IsCheckOrder:     "0",                                // 是否开启下单校验高级密码 0-关闭 1-开启. 默认关闭
			AutoLockScreen:   "300",                              // 自动锁屏（秒），默认5分钟
			LanguageList:     languageList,                       // 语言列表
			Language:         []string{defaultLanguage},          // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
			DefaultLanguage:  defaultLanguage,                    // 默认语言
		},
		AdvancedPassword: "666888", // 高级设置密码
		LockPassword:     "666888", // 锁屏密码
	}
}

// 默认商城设置
func (s *Srv) getDefaultStore(language string) setting.Store {
	return setting.Store{
		Name:          "Shop",                     // 商城名称
		AvatarURL:     "image/user/avatarUrl.png", //默认头像
		LogoURL:       "image/diy/logo.png",       //商城logo
		ZeroingMethod: "0",                        // 抹零方式，0-不抹零 1-抹分 2-抹角 3-四舍五入到角 4-四舍五入到元
		NoClearTable:  "0",                        // 结账后不清台 0-清台 1-不清台
		TimeZoneList: []setting.TimeZoneItem{
			{
				Name:  "（UTC+09:00）" + i18n.Translate(language, "东京"),
				Key:   "Asia/Tokyo",
				Value: "（UTC+09:00）" + i18n.Translate(language, "东京"),
			}, {
				Name:  "（UTC+08:00）" + i18n.Translate(language, "北京，重庆，香港特别行政区，乌鲁木齐"),
				Key:   "Asia/Shanghai",
				Value: "（UTC+08:00）" + i18n.Translate(language, "北京，重庆，香港特别行政区，乌鲁木齐"),
			}, {
				Name:  "（UTC+07:00）" + i18n.Translate(language, "曼谷，河内，雅加达"),
				Key:   "Asia/Bangkok",
				Value: "（UTC+07:00）" + i18n.Translate(language, "曼谷，河内，雅加达"),
			}, {
				Name:  "（UTC+3:00）" + i18n.Translate(language, "安卡拉"),
				Key:   "Europe/Istanbul",
				Value: "（UTC+3:00）" + i18n.Translate(language, "安卡拉"),
			}, {
				Name:  "（UTC+2:00）" + i18n.Translate(language, "斯德哥尔摩"),
				Key:   "Europe/Stockholm",
				Value: "（UTC+2:00）" + i18n.Translate(language, "斯德哥尔摩"),
			},
		}, // 时区列表
		Language: []dto.LanguageItem{}, // 语言列表
	}
}

// 小票打印机设置
func (s *Srv) getDefaultPrinter(language string, languageList []dto.LanguageItem) setting.Printer {
	var defaultLanguage = i18n.LanguageEN
	if len(languageList) > 0 {
		defaultLanguage = languageList[0].Name
	}
	return setting.Printer{
		CashierOpen:        "1",                            // 是否开启打印
		CashierPrinterID:   "-1",                           // 打印机id
		CashierPrinter:     []setting.CashierPrinterItem{}, // 打印机列表
		LanguageList:       languageList,                   // 语言列表
		LanguageMethod:     "1",                            // 语言方式（收银） 1-单语言 2-多语言
		DefaultLanguage:    defaultLanguage,                // 打印语言（收银）
		PrintMethod:        "1",                            // 打印方式（收银） 1-文本打印 2-图片打印
		KitchenLanguage:    defaultLanguage,                // 打印语言（送厨）
		KitchenPrintMethod: "1",                            // 打印方式（送厨） 1-文本打印 2-图片打印
		ConsumptionTax:     "1",                            // 消费税 1显示全部类型 2仅显示商品已含税 3仅显示商品未含税 4全部不显示
		BuffetSignOpen:     "1",                            // 自助餐标识设置（默认开启）
		MonetaryUnitOpen:   "1",                            // 货币单位（默认开启）
		CalendarList: []setting.CalendarItem{{
			Key:  "1",
			Name: i18n.Translate(language, "公历"),
		}, {
			Key:  "3",
			Name: i18n.Translate(language, "佛历"),
		}}, // 日历列表 （1-公历 2-农历 3-佛历 4-伊斯兰历 5-犹太历 ）
		PrintList: []setting.PrintItem{{
			Key:  "1",
			Name: i18n.Translate(language, "文本打印"),
		}, {
			Key:  "2",
			Name: i18n.Translate(language, "图片打印"),
		}}, // 打印方式列表 （1-文本打印 2-图片打印 ）
		DefaultCalendar: "1", // 日历类型 （1-公历 2-农历 3-佛历 4-伊斯兰历 5-犹太历 ）
	}
}

// 默认用户充值设置
func (s *Srv) getDefaultRecharge() setting.Recharge {
	return setting.Recharge{
		IsEntrance:  "1", // 是否允许用户充值
		IsCustom:    "1", // 是否允许自定义金额
		IsMatchPlan: "1", // 自定义金额是否自动匹配合适的套餐
		Describe: "1. 账户充值仅限微信在线方式支付，充值金额实时到账；\n" +
			"2. 账户充值套餐赠送的金额即时到账；\n" +
			"3. 账户余额有效期：自充值日起至用完即止；\n" +
			"4. 若有其它疑问，可拨打客服电话400-000-1234", // 充值说明
	}
}

// 默认积分设置
func (s *Srv) getDefaultPoints() setting.Points {
	return setting.Points{
		DeductionOrder:     "1",   // 扣款顺序 1-先主账户后赠送账户 2-先赠送账户后主账户 3-按比例
		DeductRatioMain:    "100", // 主账户扣款比例0-100
		DeductRatioGift:    "0",   // 赠送账户扣款比例0-100
		PointsName:         "积分",  // 积分名称自定义
		IsShoppingGift:     "0",   // 是否开启购物送积分
		GiftRatio:          "100", // 积分赠送比例
		IsShoppingDiscount: "0",
		Discount: setting.DiscountItem{
			DiscountRatio:  "0.01",
			FullOrderPrice: "100.00",
			MaxMoneyRatio:  "10",
		},
		Describe: "a) 积分不可兑现、不可转让,仅可在本平台使用;\n" +
			"b) 您在本平台参加特定活动也可使用积分,详细使用规则以具体活动时的规则为准;\n" +
			"c) 积分的数值精确到个位(小数点后全部舍弃,不进行四舍五入)\n" +
			"d) 买家在完成该笔交易(订单状态为“已签收”)后才能得到此笔交易的相应积分,如购买商品参加店铺其他优惠,则优惠的金额部分不享受积分获取;",
		ShoppingGiftRules: []setting.ShoppingGiftRule{
			{
				Type:                     setting.RuleTypePaymentAmount,
				IsOpen:                   "0",
				IsMemberLevelRelated:     "0",
				Value:                    "",
				PaymentAmountRequirement: "",
				MealType:                 []string{},
				BalancePaymentGetPoints:  "1",
				RefundReturnPoints:       "0",
				MemberLevels:             []setting.MemberLevelItem{},
			},
			{
				Type:                     setting.RuleTypeDesk,
				IsOpen:                   "0",
				IsMemberLevelRelated:     "0",
				Value:                    "",
				PaymentAmountRequirement: "",
				MealType:                 []string{"buffet"},
				BalancePaymentGetPoints:  "1",
				RefundReturnPoints:       "0",
				MemberLevels:             []setting.MemberLevelItem{},
			},
		},
		Exchange: setting.PointsExchange{
			OpenPointsExchange: "0",
			PointsExchangeRate: "",
			AutoPointsExchange: "0",
		},
	}
}

// 默认系统设置
func (s *Srv) getDefaultSysAdminConfig() setting.SysAdminConfig {
	return setting.SysAdminConfig{
		BrandName:     "Shop",                         // 商城名称
		BrandLogo:     "/image/logo/ttpos_64_64.png",  // 商城背景图
		BrandLogoLong: "/image/logo/ttpos_146_40.png", // 商城logo
		BrowserLogo:   "/image/logo/ttpos_64_64.png",  // 浏览器LOGO
		BrowserTitle:  "Shop",                         // 浏览器标题
	}
}

// 默认系统设置
func (s *Srv) getDefaultSysConfig() setting.SysConfig {
	return setting.SysConfig{
		ShopName:    "Shop", // 商城名称
		CashierName: "收银台",  // 收银台名称
	}
}

// 默认充值设置
func (s *Srv) getDefaultBalance() setting.Balance {
	return setting.Balance{
		IsOpen:   "0", // 是否开启
		IsPlan:   "1", // 是否可以自定义
		MinMoney: 1,   // 最低充值金额
		Describe: "a) 账户充值仅限在线方式支付，充值金额实时到账；\n" +
			"b) 有问题请联系客服;\n", // 充值说明
	}
}

// 默认货币单位设置
func (s *Srv) getDefaultCurrency() setting.Currency {
	return setting.Currency{
		Unit:             "฿", // 货币单位，默认泰铢
		PrintUnit:        "฿", // 货币单位 - 打印专用，默认泰铢
		UnitPosition:     "0", // 主货币显示位置 0-金额前 1-金额后
		IsOpen:           "0", // 副货币单位开关
		ViceUnitPosition: "0", // 副货币显示位置 0-金额前 1-金额后
	}
}

// 默认税率管理设置
func (s *Srv) getDefaultTaxRate() setting.TaxRate {
	// DefaultTaxRateSetting 税率管理
	return setting.TaxRate{
		IsOpen:         "0",                            // 是否开启 0关闭 1开启
		CalcType:       "1",                            // 计算类型 商品已含税价-1 商品未含税价-2
		AddTaxCategory: []setting.AddTaxCategoryItem{}, // 增值税分类
	}
}

// 默认服务费设置
func (s *Srv) getDefaultServiceCharge() setting.ServiceCharge {
	return setting.ServiceCharge{
		IsOpen:              "0",              // 是否开启 0关闭 1开启
		ChargeType:          "1",              // 服务费类型 1-固定金额 2-百分比
		IsOpenTax:           "0",              // 税费 1-收取税费 0-不收取税费
		ApplyScope:          "1",              // 适用范围 1-全部 2-部分
		ApplyScopeOrdering:  "0",              // 适用范围-点餐 0-关闭 1-开启
		ApplyScopeTable:     "0",              // 适用范围-桌台 0-关闭 1-开启
		ApplyScopeTableList: make([]int64, 0), // 适用范围-桌台id列表
	}
}

func (s *Srv) parseCashierSetting(values string) ([]byte, error) {
	// 解析json字符串为map进行处理
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(values), &jsonMap)
	if err != nil {
		return nil, errors.WithMessage(err, "解析收银机设置失败-01")
	}
	// 处理 isShowScanSoldOut
	if isShowScanSoldOut, ok := jsonMap["is_show_scan_sold_out"]; ok {
		if numVal, ok := isShowScanSoldOut.(float64); ok {
			jsonMap["is_show_scan_sold_out"] = int(numVal)
		} else if strVal, ok := isShowScanSoldOut.(string); ok {
			jsonMap["is_show_scan_sold_out"], _ = strconv.Atoi(strVal)
		}
	}
	// 处理 isShowAssistantSoldOut
	if isShowAssistantSoldOut, ok := jsonMap["is_show_assistant_sold_out"]; ok {
		if numVal, ok := isShowAssistantSoldOut.(float64); ok {
			jsonMap["is_show_assistant_sold_out"] = int(numVal)
		} else if strVal, ok := isShowAssistantSoldOut.(string); ok {
			jsonMap["is_show_assistant_sold_out"], _ = strconv.Atoi(strVal)
		}
	}
	// 处理 isAutoOrder
	if isAutoOrder, ok := jsonMap["is_auto_order"]; ok {
		switch v := isAutoOrder.(type) {
		case float64:
			jsonMap["is_auto_order"] = strconv.Itoa(int(v))
		case int:
			jsonMap["is_auto_order"] = strconv.Itoa(v)
		case string:
			jsonMap["is_auto_order"] = v
		default:
			jsonMap["is_auto_order"] = "0"
		}
	}

	// 处理 auto_order_limit
	if autoOrderLimit, ok := jsonMap["auto_order_limit"]; ok {
		if numVal, ok := autoOrderLimit.(float64); ok {
			jsonMap["auto_order_limit"] = strconv.Itoa(int(numVal))
		} else if strVal, ok := autoOrderLimit.(int); ok {
			jsonMap["auto_order_limit"] = strconv.Itoa(strVal)
		}
	}

	// 处理 menu_show_sold_out
	if menuShowSoldOut, ok := jsonMap["menu_show_sold_out"]; ok {
		if numVal, ok := menuShowSoldOut.(float64); ok {
			jsonMap["menu_show_sold_out"] = strconv.Itoa(int(numVal))
		} else if strVal, ok := menuShowSoldOut.(int); ok {
			jsonMap["menu_show_sold_out"] = strconv.Itoa(strVal)
		}
	}

	// 重新序列化为JSON
	modifiedJSON, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, errors.WithMessage(err, "重新序列化JSON失败 - 02")
	}
	return modifiedJSON, nil
}
