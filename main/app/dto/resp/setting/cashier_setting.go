package setting

import (
	"strconv"
	"ttpos-server-go/app/dto"
)

// CashierResp 收银机设置，接口响应
type CashierResp struct {
	Carousel                []CarouselItem     `json:"carousel"`                   // 上传后的轮播内容url（图片 + 视频）
	IsAutoSend              string             `json:"is_auto_send"`               // 收银结账自动送厨房 0-关闭 1-开启
	OrderMethod             OrderMethod        `json:"order_method"`               // 用餐方式 收银-is_cashier_order（0-关闭 1-开启） 桌台-is_table_order（0-关闭 1-开启）
	Server                  Server             `json:"server"`                     // 收银机服务器连接
	IsRemainColor           string             `json:"is_remain_color"`            // 是否开启剩余时长颜色 0-关闭 1-开启
	RemainColor             []string           `json:"remain_color"`               // 剩余时长颜色 10分钟-红色(#E50028) 20分钟-黄色(#F2A000)
	IsOpenCashierPassword   string             `json:"is_open_cashier_password"`   // 是否开启钱箱密码 0-关闭 1-开启
	IsAutoLockScreen        string             `json:"is_auto_lock_screen"`        // 是否开启自动锁屏 0-关闭 1-开启
	AutoLockScreen          string             `json:"auto_lock_screen"`           // 自动锁屏（秒），默认5分钟
	IsShowScanSoldOut       int                `json:"is_show_scan_sold_out"`      // 扫码点餐是否显示售罄商品 0-不显示 1-显示
	IsShowAssistantSoldOut  int                `json:"is_show_assistant_sold_out"` // 点餐助手是否显示售罄商品 0-不显示 1-显示
	LanguageList            []dto.LanguageItem `json:"language_list"`              // 语言列表
	Language                []string           `json:"language"`                   // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
	DefaultLanguage         string             `json:"default_language"`           // 默认语言
	IsAutoOrder             string             `json:"is_auto_order"`              // 是否自动接单
	AutoOrderLimit          string             `json:"auto_order_limit"`           // 自动接单金额上限
	IsAutoVoice             string             `json:"is_auto_voice"`              // 是否开启自动接单语音播报
	IsAutoMemberOrder       string             `json:"is_auto_member_order"`       // 是否自动接单会员订单 0-关闭 1-开启
	AutoMemberOrderLimit    string             `json:"auto_member_order_limit"`    // 自动接单会员订单金额上限
	IsAutoVoiceMemberOrder  string             `json:"is_auto_voice_member_order"` // 是否开启自动接单会员订单语音播报
	MenuShowSoldOut         string             `json:"menu_show_sold_out"`         // 是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	MemberShowSoldOut       string             `json:"member_show_sold_out"`       // 是否显示会员端售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	NoOrderCarouselInterval string             `json:"no_order_carousel_interval"` // 未点餐时轮播间隔(秒)
	OrderDisplayMode        string             `json:"order_display_mode"`         // 点餐时展示模式 carousel/order/order_carousel
	OrderCarouselInterval   string             `json:"order_carousel_interval"`    // 点餐时轮播间隔(秒)
}

// IsAutoMemberOrderBool 是否开启自动接单会员订单
func (res *CashierResp) IsAutoMemberOrderBool() bool {
	return res.IsAutoMemberOrder == "1"
}

// AutoMemberOrderLimitValue 是否开启自动接单会员订单金额上限
func (res *CashierResp) AutoMemberOrderLimitValue() float64 {
	val, _ := strconv.ParseFloat(res.AutoMemberOrderLimit, 64)
	return val
}

// Cashier 收银机设置
type Cashier struct {
	CashierResp
	AdvancedPassword string `json:"advanced_password"` // 高级设置密码
	CashierPassword  string `json:"cashier_password"`  // 钱箱密码
	LockPassword     string `json:"lock_password"`     // 锁屏密码
}

type OrderMethod struct {
	IsCashierOrder string `json:"is_cashier_order"` // 收银用餐
	IsTableOrder   string `json:"is_table_order"`   // 桌台用餐
}
