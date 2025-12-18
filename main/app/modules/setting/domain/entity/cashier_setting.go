package entity

import (
	"strconv"
	"ttpos-server-go/app/modules/setting/domain/valueobject"
)

// CashierSetting 收银机设置实体
type CashierSetting struct {
	Carousel                []valueobject.CarouselItem `json:"carousel"`                   // 上传后的轮播内容url（图片 + 视频）
	IsAutoSend              string                     `json:"is_auto_send"`               // 收银结账自动送厨房 0-关闭 1-开启
	OrderMethod             OrderMethod                `json:"order_method"`               // 用餐方式
	Server                  Server                     `json:"server"`                     // 收银机服务器连接
	IsRemainColor           string                     `json:"is_remain_color"`            // 是否开启剩余时长颜色 0-关闭 1-开启
	RemainColor             []string                   `json:"remain_color"`               // 剩余时长颜色 10分钟-红色(#E50028) 20分钟-黄色(#F2A000)
	IsOpenCashierPassword   string                     `json:"is_open_cashier_password"`   // 是否开启钱箱密码 0-关闭 1-开启
	IsAutoLockScreen        string                     `json:"is_auto_lock_screen"`        // 是否开启自动锁屏 0-关闭 1-开启
	AutoLockScreen          string                     `json:"auto_lock_screen"`           // 自动锁屏（秒），默认5分钟
	IsShowScanSoldOut       int                        `json:"is_show_scan_sold_out"`      // 扫码点餐是否显示售罄商品 0-不显示 1-显示
	IsShowAssistantSoldOut  int                        `json:"is_show_assistant_sold_out"` // 点餐助手是否显示售罄商品 0-不显示 1-显示
	LanguageList            []valueobject.LanguageItem `json:"language_list"`              // 语言列表
	Language                []string                   `json:"language"`                   // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
	DefaultLanguage         string                     `json:"default_language"`           // 默认语言
	IsAutoOrder             string                     `json:"is_auto_order"`              // 是否自动接单
	AutoOrderLimit          string                     `json:"auto_order_limit"`           // 自动接单金额上限
	IsAutoVoice             string                     `json:"is_auto_voice"`              // 是否开启自动接单语音播报
	IsAutoMemberOrder       string                     `json:"is_auto_member_order"`       // 是否自动接单会员订单 0-关闭 1-开启
	AutoMemberOrderLimit    string                     `json:"auto_member_order_limit"`    // 自动接单会员订单金额上限
	IsAutoVoiceMemberOrder  string                     `json:"is_auto_voice_member_order"` // 是否开启自动接单会员订单语音播报
	MenuShowSoldOut         string                     `json:"menu_show_sold_out"`         // 是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	MemberShowSoldOut       string                     `json:"member_show_sold_out"`       // 是否显示会员端售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	NoOrderCarouselInterval string                     `json:"no_order_carousel_interval"` // 未点餐时轮播间隔(秒)
	OrderDisplayMode        string                     `json:"order_display_mode"`         // 点餐时展示模式 carousel/order/order_carousel
	OrderCarousel           []valueobject.CarouselItem `json:"order_carousel"`             // 点餐时轮播内容（图片+视频，最多15个）
	OrderCarouselInterval   string                     `json:"order_carousel_interval"`    // 点餐时轮播间隔(秒)

	// 密码相关
	AdvancedPassword string `json:"advanced_password"` // 高级设置密码
	CashierPassword  string `json:"cashier_password"`  // 钱箱密码
	LockPassword     string `json:"lock_password"`     // 锁屏密码
}

// NewCashierSetting 创建新的收银机设置
func NewCashierSetting() *CashierSetting {
	return &CashierSetting{
		IsAutoSend:              "0",     // 默认关闭自动送厨房
		IsRemainColor:           "0",     // 默认关闭剩余时长颜色
		IsOpenCashierPassword:   "0",     // 默认关闭钱箱密码
		IsAutoLockScreen:        "0",     // 默认关闭自动锁屏
		AutoLockScreen:          "300",   // 默认5分钟
		IsShowScanSoldOut:       0,       // 默认不显示售罄
		IsShowAssistantSoldOut:  0,       // 默认不显示售罄
		IsAutoOrder:             "0",     // 默认关闭自动接单
		IsAutoMemberOrder:       "0",     // 默认关闭自动接单会员订单
		MenuShowSoldOut:         "0",     // 默认不显示售罄
		MemberShowSoldOut:       "0",     // 默认不显示售罄
		NoOrderCarouselInterval: "10",    // 默认10秒
		OrderDisplayMode:        "order", // 默认订单模式
		OrderCarouselInterval:   "10",    // 默认10秒
		Carousel:                make([]valueobject.CarouselItem, 0),
		OrderCarousel:           make([]valueobject.CarouselItem, 0),
		LanguageList:            make([]valueobject.LanguageItem, 0),
		Language:                make([]string, 0),
		RemainColor:             make([]string, 0),
	}
}

// DefaultCashierSetting 获取默认收银机设置
func DefaultCashierSetting(languageList []valueobject.LanguageItem) *CashierSetting {
	var defaultLanguage = "en"
	if len(languageList) > 0 {
		defaultLanguage = languageList[0].Name
	}

	return &CashierSetting{
		Carousel:   make([]valueobject.CarouselItem, 0), // 上传后的轮播内容url（图片 + 视频）
		IsAutoSend: "0",                                 // 收银结账自动送厨房 0-关闭 1-开启
		OrderMethod: OrderMethod{
			IsCashierOrder: "1",
			IsTableOrder:   "1",
		}, // 用餐方式 收银-is_cashier_order（0-关闭 1-开启） 桌台-is_table_order（0-关闭 1-开启）
		IsRemainColor:           "1",                            // 是否开启剩余时长颜色 0-关闭 1-开启
		RemainColor:             []string{"#E50028", "#F2A000"}, // 剩余时长颜色 10分钟-红色(#E50028) 20分钟-黄色(#F2A000)
		IsOpenCashierPassword:   "1",                            // 是否开启钱箱密码 0-关闭 1-开启
		IsAutoLockScreen:        "1",                            // 是否开启自动锁屏 0-关闭 1-开启
		AutoLockScreen:          "300",                          // 自动锁屏（秒），默认5分钟
		IsShowScanSoldOut:       1,                              // 扫码点餐是否显示售罄商品 0-不显示 1-显示
		IsShowAssistantSoldOut:  1,                              // 点餐助手是否显示售罄商品 0-不显示 1-显示
		LanguageList:            languageList,                   // 语言列表
		Language:                []string{defaultLanguage},      // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
		DefaultLanguage:         defaultLanguage,                // 默认语言
		IsAutoOrder:             "0",                            // 是否自动接单
		AutoOrderLimit:          "1000",                         // 自动接单金额上限
		IsAutoVoice:             "0",                            // 是否开启自动接单语音播报
		IsAutoMemberOrder:       "0",                            // 是否自动接单会员订单 0-关闭 1-开启
		AutoMemberOrderLimit:    "1000",                         // 自动接单会员订单金额上限
		IsAutoVoiceMemberOrder:  "0",                            // 是否开启自动接单会员订单语音播报
		MenuShowSoldOut:         "1",                            // 是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
		MemberShowSoldOut:       "1",                            // 是否显示会员端售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
		NoOrderCarouselInterval: "10",                           // 未点餐时轮播间隔(秒)
		OrderDisplayMode:        "order",                        // 点餐时展示模式 carousel/order/order_carousel
		OrderCarouselInterval:   "10",                           // 点餐时轮播间隔(秒)
		AdvancedPassword:        "666888",                       // 高级设置密码
		CashierPassword:         "666888",                       // 钱箱密码
		LockPassword:            "666888",                       // 锁屏密码
	}
}

// IsAutoMemberOrderBool 是否开启自动接单会员订单
func (c *CashierSetting) IsAutoMemberOrderBool() bool {
	return c.IsAutoMemberOrder == "1"
}

// AutoMemberOrderLimitValue 获取自动接单会员订单金额上限
func (c *CashierSetting) AutoMemberOrderLimitValue() float64 {
	val, _ := strconv.ParseFloat(c.AutoMemberOrderLimit, 64)
	return val
}

// UpdateCarousel 更新轮播内容
func (c *CashierSetting) UpdateCarousel(carousel []valueobject.CarouselItem) {
	c.Carousel = carousel
}

// UpdateOrderCarousel 更新点餐时轮播内容
func (c *CashierSetting) UpdateOrderCarousel(orderCarousel []valueobject.CarouselItem) {
	c.OrderCarousel = orderCarousel
}

// UpdateLanguages 更新语言设置
func (c *CashierSetting) UpdateLanguages(languages []string, defaultLanguage string, languageList []valueobject.LanguageItem) {
	c.Language = languages
	c.DefaultLanguage = defaultLanguage
	c.LanguageList = languageList
}

// UpdateAutoOrderSettings 更新自动接单设置
func (c *CashierSetting) UpdateAutoOrderSettings(isAutoOrder, autoOrderLimit, isAutoVoice string) {
	c.IsAutoOrder = isAutoOrder
	c.AutoOrderLimit = autoOrderLimit
	c.IsAutoVoice = isAutoVoice
}

// UpdateAutoMemberOrderSettings 更新自动接单会员订单设置
func (c *CashierSetting) UpdateAutoMemberOrderSettings(isAutoMemberOrder, autoMemberOrderLimit, isAutoVoiceMemberOrder string) {
	c.IsAutoMemberOrder = isAutoMemberOrder
	c.AutoMemberOrderLimit = autoMemberOrderLimit
	c.IsAutoVoiceMemberOrder = isAutoVoiceMemberOrder
}
