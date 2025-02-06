package constant

import "ttpos-server-go/app/dto/req/setting"

const SettingPrinter = "printer"                 // 小票打印
const SettingStore = "store"                     // 商城设置
const SettingRecharge = "recharge"               // 充值设置
const SettingPoints = "points"                   // 积分设置
const SettingSysAdminConfig = "sys_admin_config" // 系统配置
const SettingSysConfig = "sys_config"            // 系统配置
const SettingBalance = "balance"                 // 充值设置
const SettingCurrency = "currency"               // 门店-货币单位
const SettingTaxRate = "tax_rate"                // 门店-税率管理
const SettingServiceCharge = "service_charge"    // 门店-服务费
const SettingPayment = "payment"                 // 门店-支付方式
const SettingBusiness = "business"               // 门店-业务设置
const SettingCashier = "cashier"                 // 各端-收银机设置
const SettingTablet = "tablet"                   // 各端-平板端设置
const SettingH5 = "h5"                           // 各端-扫码H5设置
const SettingKitchen = "kitchen"                 // 各端-厨显设置
const SettingAssistant = "assistant"             // 各端-点餐助手设置
const SettingBuffet = "buffet"                   // 自助餐-自助餐设置
const SettingCloudBasic = "cloud_basic"          // 云端-基础信息

// DefaultStoreSetting 商城设置
var DefaultStoreSetting = setting.Store{
	Name:          "XXX shop",                 // 商城名称
	AvatarURL:     "image/user/avatarUrl.png", //默认头像
	LogoURL:       "image/diy/logo.png",       //商城logo
	ZeroingMethod: "0",                        // 抹零方式，0-不抹零 1-抹分 2-抹角 3-四舍五入到角 4-四舍五入到元
	IPWhiteList:   "",                         // ip白名单
	TimeZone:      "",                         // 时区
	NoClearTable:  "0",                        // 结账后不清台 0-清台 1-不清台
	// todo 需要翻译
	TimeZoneList: []setting.TimeZoneItem{
		{
			Name:  "（UTC+09:00）" + "东京",
			Key:   "Asia/Tokyo",
			Value: "（UTC+09:00）" + "东京",
		}, {
			Name:  "UTC+08:00）" + "北京，重庆，香港特别行政区，乌鲁木齐",
			Key:   "Asia/Shanghai",
			Value: "UTC+08:00）" + "北京，重庆，香港特别行政区，乌鲁木齐",
		}, {
			Name:  "（UTC+07:00）" + "曼谷，河内，雅加达",
			Key:   "Asia/Bangkok",
			Value: "（UTC+07:00）" + "曼谷，河内，雅加达",
		}, {
			Name:  "（UTC+3:00）" + "安卡拉",
			Key:   "Europe/Istanbul",
			Value: "（UTC+3:00）" + "安卡拉",
		},
	}, // 时区列表
	Company:      "",                   // 公司名称
	Address:      "",                   // 地址
	Phone:        "",                   // 联系电话
	TaxNumber:    "",                   // 税号
	ChainNumber:  "",                   // 连锁编号
	Language:     []setting.Language{}, // 系统语言\语言列表
	AuthLanguage: "",                   // 授权语言
}

// DefaultPrinterSetting 小票打印机设置
var DefaultPrinterSetting = setting.Printer{
	CashierOpen:        "1",  // 是否开启打印
	CashierPrinterID:   "-1", // 打印机id
	CashierPrinter:     []interface{}{},
	LanguageList:       []setting.LanguageItem{}, // 语言列表
	LanguageMethod:     "1",                      // 语言方式（收银） 1-单语言 2-多语言
	DefaultLanguage:    "",                       // 打印语言（收银） // todo 赋值
	PrintMethod:        1,                        // 打印方式（收银） 1-文本打印 2-图片打印
	KitchenLanguage:    "",                       // 打印语言（送厨） // todo 赋值
	KitchenPrintMethod: 1,                        // 打印方式（送厨） 1-文本打印 2-图片打印
	ConsumptionTax:     "1",                      // 消费税 1显示全部类型 2仅显示商品已含税 3仅显示商品未含税 4全部不显示
	BuffetSignOpen:     "1",                      // 自助餐标识设置（默认开启）
	MonetaryUnitOpen:   "1",                      // 货币单位（默认开启）
	CalendarList: []setting.CalendarItem{{
		Key:  "1",
		Name: "公历",
	}, {
		Key:  "3",
		Name: "佛历",
	}}, // 日历列表 （1-公历 2-农历 3-佛历 4-伊斯兰历 5-犹太历 ） // ToDo 翻译
	PrintList: []setting.PrintItem{{
		Key:  "1",
		Name: "文本打印",
	}, {
		Key:  "2",
		Name: "图片打印",
	}}, // 打印方式列表 （1-文本打印 2-图片打印 ） // ToDo 翻译
	DefaultCalendar: "1", // 日历类型 （1-公历 2-农历 3-佛历 4-伊斯兰历 5-犹太历 ）
}

// DefaultRechargeSetting 用户充值设置
var DefaultRechargeSetting = setting.Recharge{
	IsEntrance:  "1", // 是否允许用户充值
	IsCustom:    "1", // 是否允许自定义金额
	IsMatchPlan: "1", // 自定义金额是否自动匹配合适的套餐
	Describe: "1. 账户充值仅限微信在线方式支付，充值金额实时到账；\n" +
		"2. 账户充值套餐赠送的金额即时到账；\n" +
		"3. 账户余额有效期：自充值日起至用完即止；\n" +
		"4. 若有其它疑问，可拨打客服电话400-000-1234", // 充值说明
}

// DefaultPointsSetting 积分设置
var DefaultPointsSetting = setting.Points{
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
	DeductOrder: "",
}

// DefaultSysAdminConfigSetting 系统设置
var DefaultSysAdminConfigSetting = setting.SysAdminConfig{
	BrandName:          "XXX shop",                   // 商城名称
	BrandLogo:          "/image/logo/jbc_64_64.png",  // 商城背景图
	BrandLogoLong:      "/image/logo/jbc_146_40.png", // 商城logo
	BrowserLogo:        "/image/logo/jbc_146_40.png", // 浏览器LOGO
	BrowserTitle:       "XXX shop",                   // 浏览器标题
	ExpirationReminder: 0,                            // 商城logo
}

// DefaultSysConfigSetting 系统设置
var DefaultSysConfigSetting = setting.SysConfig{
	ShopName:     "XXX shop", // 商城名称
	ShopBgImg:    "",         // 商城背景图
	ShopLogoImg:  "",         // 商城logo
	CashierName:  "收银台",      // 收银台名称
	CashierBgImg: "",         // 收银台背景图
}

// DefaultBalanceSetting 充值设置
var DefaultBalanceSetting = setting.Balance{
	IsOpen:   "0", // 是否开启
	IsPlan:   "1", // 是否可以自定义
	MinMoney: 1,   // 最低充值金额
	Describe: "a) 账户充值仅限在线方式支付，充值金额实时到账；\n" +
		"b) 有问题请联系客服;\n", // 充值说明
}

// DefaultCurrencySetting 货币单位
var DefaultCurrencySetting = setting.Currency{
	Unit:             "฿", // 货币单位，默认泰铢
	PrintUnit:        "฿", // 货币单位 - 打印专用，默认泰铢
	UnitPosition:     "0", // 主货币显示位置 0-金额前 1-金额后
	IsOpen:           "0", // 副货币单位开关
	ViceUnit:         "",  // 副货币单位
	ViceUnitPosition: "0", // 副货币显示位置 0-金额前 1-金额后
	UnitRate:         "",  // 单位汇率
}

// DefaultTaxRateSetting 税率管理
var DefaultTaxRateSetting = setting.TaxRate{
	IsOpen:         "0",                            // 是否开启 0关闭 1开启
	CalcType:       "1",                            // 计算类型 商品已含税价-1 商品未含税价-2
	AddTaxCategory: []setting.AddTaxCategoryItem{}, // 增值税分类
}

// DefaultServiceChargeSetting 服务费
var DefaultServiceChargeSetting = setting.ServiceCharge{
	IsOpen:            "0", // 是否开启 0关闭 1开启
	ChargeType:        "1", // 服务费类型 1-固定金额 2-百分比
	ServiceCharge:     "",  // 服务费金额
	ServiceChargeRate: "",  // 服务费率
	IsOpenTax:         "0", // 税费 1-收取税费 0-不收取税费
}

// DefaultCashierSetting 收银机设置
var DefaultCashierSetting = setting.Cashier{
	Carousel:   []setting.CarouselItem{}, // 上传后的轮播内容url（图片 + 视频）
	IsAutoSend: "0",                      // 收银结账自动送厨房 0-关闭 1-开启
	OrderMethod: setting.OrderMethodItem{
		IsCashierOrder: "1",
		IsTableOrder:   "1",
	}, // 用餐方式 收银-is_cashier_order（0-关闭 1-开启） 桌台-is_table_order（0-关闭 1-开启）
	Server: setting.Server{ // ToDo 赋值
		IP:   "",
		Port: "",
	}, // 收银机服务器连接
	IsRemainColor:          "1",                            // 是否开启剩余时长颜色 0-关闭 1-开启
	RemainColor:            []string{"#E50028", "#F2A000"}, // 剩余时长颜色 10分钟-红色(#E50028) 20分钟-黄色(#F2A000)
	AdvancedPassword:       "666888",                       // 高级设置密码
	IsOpenCashierPassword:  "1",                            // 是否开启钱箱密码 0-关闭 1-开启
	CashierPassword:        "666888",                       // 钱箱密码
	LockPassword:           "666888",                       // 锁屏密码
	IsAutoLockScreen:       "1",                            // 是否开启自动锁屏 0-关闭 1-开启
	AutoLockScreen:         "300",                          // 自动锁屏（秒），默认5分钟
	IsShowScanSoldOut:      1,                              // 扫码点餐是否显示售罄商品 0-不显示 1-显示
	IsShowAssistantSoldOut: 1,                              // 点餐助手是否显示售罄商品 0-不显示 1-显示
	LanguageList:           []setting.LanguageItem{},       // 语言列表
	Language:               []string{ /*todo 赋值默认*/ },      // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
	DefaultLanguage:        "",                             // // 默认语言 todo 赋值默认
	IsAutoOrder:            "0",                            // 是否自动接单
	AutoOrderLimit:         "1000",                         // 自动接单金额上限
	IsAutoVoice:            "0",                            // 是否开启自动接单语音播报
	MenuShowSoldOut:        "1",                            // 是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	KitchenLanguage:        "",
}

// DefaultTabletSetting 平板端设置
var DefaultTabletSetting = setting.Tablet{
	Carousel:           []setting.CarouselItem{}, // 上传后的轮播内容url（图片 + 视频）
	IsCallService:      "1",                      // 是否开启呼叫服务员 0-关闭 1-开启
	IsCustomerOrder:    "1",                      // 是否开启顾客自助下单 0-关闭 1-开启
	IsVoiceRemind:      "1",                      // 是否开启声音提醒 0-关闭 1-开启
	IsShowSoldOut:      0,                        // 是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	IsBuffetOrderLimit: "0",                      // 自助餐下单限制 // 是否开启自助餐下单限制 0-关闭 1-开启
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
	Server:           setting.Server{},         // 平板服务器连接 ToDo 赋值
	AdvancedPassword: "666888",                 // 高级设置密码
	LanguageList:     []setting.LanguageItem{}, // 语言列表
	Language:         []string{},               // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
	DefaultLanguage:  "",                       // 默认语言 todo 赋值默认
	KitchenLanguage:  "",
}

// DefaultH5Setting 扫码H5设置
var DefaultH5Setting = setting.H5{
	IsCallService:      "1", // 是否开启呼叫服务员 0-关闭 1-开启
	IsCustomerOrder:    "1", // 是否开启顾客自助下单 0-关闭 1-开启
	IsVoiceRemind:      "1", // 是否开启声音提醒 0-关闭 1-开启
	IsShowSoldOut:      "0", // 是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
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
	LanguageList:    []setting.LanguageItem{}, // 语言列表
	Language:        []string{},               // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
	DefaultLanguage: "",                       // 默认语言 todo 赋值默认
	KitchenLanguage: "",
}

// DefaultKitchenSetting 厨显设置
var DefaultKitchenSetting = setting.Kitchen{
	IsOpen:           "1",                      // 是否开启厨显功能 0关闭 1开启
	IsComeDish:       "1",                      // 是否开启来菜提醒 0-关闭 1-开启
	IsCallService:    "1",                      // 是否开启顾客呼叫提醒 0-关闭 1-开启
	Server:           setting.Server{},         // 厨显服务器连接 todo 赋值
	AdvancedPassword: "666888",                 // 高级设置密码
	IsWaitColor:      "0",                      // 是否开启等待时长颜色 0-关闭 1-开启
	WaitColor:        []string{},               // 时长颜色 10分钟-黄色#ffff00 20分钟-红色#ff0000
	LanguageList:     []setting.LanguageItem{}, // 语言列表
	Language:         []string{},               // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
	DefaultLanguage:  "",                       // 默认语言 todo 赋值默认
	KitchenLanguage:  "",
}

// DefaultAssistantSetting 点餐助手设置
var DefaultAssistantSetting = setting.Assistant{
	Server:           setting.Server{},               // 服务器连接 todo 赋值
	IsRemainColor:    "1",                            // 是否开启剩余时长颜色 0-关闭 1-开启
	RemainColor:      []string{"#E50028", "#F2A000"}, // 剩余时长颜色 10分钟-红色(#E50028) 20分钟-黄色(#F2A000)
	AdvancedPassword: "666888",                       // 高级设置密码
	LockPassword:     "666888",                       // 锁屏密码
	DefaultMode:      "0",                            // 默认模式 0-服务员模式 1-顾客模式
	SupportFunctionList: []setting.SupportFunctionListItem{{
		Key:  "add_member",
		Name: "添加会员",
	}, {
		Key:  "people",
		Name: "人数",
	}, {
		Key:  "adjust_buffet",
		Name: "调整自助餐",
	}, {
		Key:  "turn_table",
		Name: "转台",
	}, {
		Key:  "clear_table",
		Name: "清台",
	}, {
		Key:  "merge_table",
		Name: "合单结账",
	}, {
		Key:  "discount",
		Name: "优惠折扣",
	}, {
		Key:  "price",
		Name: "价格",
	}, {
		Key:  "return_dish",
		Name: "退菜",
	}, {
		Key:  "remark",
		Name: "备注",
	}, {
		Key:  "settle",
		Name: "结账",
	}, {
		Key:  "transfer_dish",
		Name: "转菜",
	}, {
		Key:  "gift_dish",
		Name: "赠菜",
	}}, // 支持功能 （添加会员 人数 调整自助餐 转台 清台 并台 优惠折扣 价格 退菜 备注 结账）
	SupportFunction:  []setting.SupportFunctionItem{},
	IsAutoLockScreen: "1",                      // 是否开启自动锁屏 0-关闭 1-开启
	AutoLockScreen:   "300",                    // 自动锁屏（秒），默认5分钟
	LanguageList:     []setting.LanguageItem{}, // 语言列表 todo 赋值默认
	Language:         nil,                      // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
	DefaultLanguage:  "",                       // 默认语言
	KitchenLanguage:  "",
	IsAutoSend:       "",
}

// DefaultBuffetSetting 自助餐设置
var DefaultBuffetSetting = setting.Buffet{
	IsOpen:                   "1",                      // 是否开启自助餐 0-关闭 1-开启
	TabletEndTime:            "5",                      // 平板结束时间提醒（分）
	IsRemainContinue:         "0",                      // 剩余xx分不可继续点餐开关 0-关闭 1-开启
	RemainContinueTime:       "10",                     // 剩余xx分不可继续点餐
	RemainContinueNoticeTime: "10",                     // 剩余xx分提醒不可继续点餐
	IsBuyContinue:            "0",                      // 非自助餐商品到时是否能继续选购 0-关闭 1-开启
	IsAddClock:               "0",                      // 是否开启加钟 0-关闭 1-开启
	IsBuffetDiscount:         "0",                      // 是否开启自助餐优惠折扣 0-关闭 1-开启
	AddClock:                 []setting.AddClockItem{}, // 名称 - 加钟时间（分）- 价格
}

// DefaultPaymentSetting 支付方式
var DefaultPaymentSetting = setting.Payment{
	IsCash:    "0", // 是否开启现金支付 0-关闭 1-开启
	IsBalance: "0", // 是否开启余额支付 0-关闭 1-开启
	IsOther:   "1", // 是否开启其他方式支付 0-关闭 1-开启
}

// DefaultBusinessSetting 门店业务设置
var DefaultBusinessSetting = setting.Business{
	// ToDo 翻译
	ZeroingMethodList: []setting.ZeroingMethodItem{{
		Key:  "0",
		Name: "实款实收",
	}, {
		Key:  "1",
		Name: "抹分",
	}, {
		Key:  "2",
		Name: "抹角",
	}, {
		Key:  "3",
		Name: "四舍五入保留一位小数",
	}, {
		Key:  "4",
		Name: "四舍五入到整数",
	}}, // 优惠折扣自动抹零方式列表
	// ToDo 翻译
	CheckoutZeroingMethodList: []setting.CheckoutZeroingMethodItem{{
		Key:  "0",
		Name: "实款实收",
	}, {
		Key:  "1",
		Name: "抹分",
	}, {
		Key:  "2",
		Name: "抹角",
	}, {
		Key:  "5",
		Name: "抹元",
	}}, // 结账自动抹零方式列表
	ZeroingMethod:         "0", // 优惠折扣自动抹零方式
	CheckoutZeroingMethod: "0", // 结账自动抹零方式
	// ToDo 翻译
	GiftMethodList: []setting.GiftMethodItem{{
		Key:  "10",
		Name: "计入总销售额、优惠折扣",
	}, {
		Key:  "20",
		Name: "不计入总销售额、优惠折扣",
	}}, // 赠菜计算方式列表
	GiftMethod: "10", // 赠菜计算方式
	// ToDo 翻译
	FreeMethodList: []setting.FreeMethodItem{{
		Key:  "10",
		Name: "计入总销售额、优惠折扣、服务费、税费",
	}, {
		Key:  "20",
		Name: "不计入总销售额、优惠折扣、服务费、税费",
	}}, // 免单计算方式列表
	FreeMethod:        "10",     // 免单计算方式
	DiscountMethod:    "10",     // 折扣计算方式 10-按百分比 20-直接减免
	QrCode:            "123456", // 电子菜单二维码校验失效值，6位数数字
	NoClearTable:      "0",      // 结账后不清台 0-清台 1-不清台
	IsNeedPassword:    "1",      // 取消订单/退菜 0-无需密码 1-需要密码
	DishCardStyle:     "0",      // 菜品卡片样式 0-无图模式 1-图片模式
	DishCardStyleTime: "0",      // 菜品卡片样式最后更新时间
	IsInvoice:         "0",      // 开票信息 0-不需要填写 1-需要填写
}
