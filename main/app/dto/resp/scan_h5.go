package resp

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp/setting"
)

type H5Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

type H5Shop struct {
	ShopSupplierID    uint64   `json:"shop_supplier_id"`     // 商家ID
	ParentID          int      `json:"parent_id"`            // 父级ID
	Name              string   `json:"name"`                 // 商家名称
	RealName          string   `json:"real_name"`            // 商家真实名称
	LinkName          string   `json:"link_name"`            // 联系人名称
	LinkPhone         string   `json:"link_phone"`           // 联系人电话
	Logo              string   `json:"logo"`                 // 商家logo
	Level             int      `json:"level"`                // 商家等级
	SaleStock         int      `json:"sale_stock"`           // 销售库存
	Reserve           int      `json:"reserve"`              // 预定
	IsOpenMember      int      `json:"is_open_member"`       // 是否开启会员
	IsOpenTablet      int      `json:"is_open_tablet"`       // 是否开启桌台
	IsOpenScan        int      `json:"is_open_scan"`         // 是否开启扫码
	IsOpenAssistant   int      `json:"is_open_assistant"`    // 是否开启助手
	IsOpenKitchenKds  int      `json:"is_open_kitchen_kds"`  // 是否开启厨房KDS
	IsOpenBuffet      int      `json:"is_open_buffet"`       // 是否开启自助餐
	IsAcceptScanOrder int      `json:"is_accept_scan_order"` // 是否接受扫码订单
	IsOpenLocalPrint  int      `json:"is_open_local_print"`  // 是否开启本地打印
	CashLimit         int      `json:"cash_limit"`           // 现金限额
	KitchenLimit      int      `json:"kitchen_limit"`        // 厨房限额
	TabletLimit       int      `json:"tablet_limit"`         // 桌台限额
	AssistantLimit    int      `json:"assistant_limit"`      // 助手限额
	TableLimit        int      `json:"table_limit"`          // 桌台限额
	PrinterLimit      int      `json:"printer_limit"`        // 打印机限额
	Timezone          string   `json:"timezone"`             // 时区
	Languages         string   `json:"languages"`            // 语言
	Address           string   `json:"address"`              // 地址
	DeployMode        int      `json:"deploy_mode"`          // 部署模式
	MacAddr           string   `json:"mac_addr"`             // 设备mac地址
	SerialNumber      string   `json:"serial_number"`        // 设备序列号
	ChainNumber       string   `json:"chain_number"`         // 连锁编号
	BusinessID        int      `json:"business_id"`          // 商家ID
	Description       string   `json:"description"`          // 描述
	TotalMoney        string   `json:"total_money"`          // 总金额
	Money             string   `json:"money"`                // 金额
	FreezeMoney       string   `json:"freeze_money"`         // 冻结金额
	CashMoney         string   `json:"cash_money"`           // 现金金额
	DepositMoney      string   `json:"deposit_money"`        // 押金金额
	UserID            int      `json:"user_id"`              // 用户ID
	FavCount          int      `json:"fav_count"`            // 收藏数量
	Status            int      `json:"status"`               // 状态
	StoreType         int      `json:"store_type"`           // 店铺类型
	TotalGift         int      `json:"total_gift"`           // 总赠菜
	IsRecycle         int      `json:"is_recycle"`           // 是否回收
	IsMain            int      `json:"is_main"`              // 是否主店
	ProvinceID        int      `json:"province_id"`          // 省份ID
	CityID            int      `json:"city_id"`              // 城市ID
	RegionID          int      `json:"region_id"`            // 区域ID
	Longitude         string   `json:"longitude"`            // 经度
	Latitude          string   `json:"latitude"`             // 纬度
	ShippingFee       string   `json:"shipping_fee"`         // 运费
	BagType           int      `json:"bag_type"`             // 包装类型
	BagPrice          string   `json:"bag_price"`            // 包装价格
	StorebagType      int      `json:"storebag_type"`        // 店铺包装类型
	StorebagPrice     string   `json:"storebag_price"`       // 店铺包装价格
	DeliveryTime      []string `json:"delivery_time"`        // 配送时间
	PickTime          []string `json:"pick_time"`            // 自提时间
	StoreTime         []string `json:"store_time"`           // 店铺营业时间
	DeliveryDistance  int      `json:"delivery_distance"`    // 配送距离
	DeliverySet       []string `json:"delivery_set"`         // 配送设置
	StoreSet          []string `json:"store_set"`            // 店铺设置
	MinMoney          string   `json:"min_money"`            // 最小金额
	SettleType        int      `json:"settle_type"`          // 结算方式
	ServiceType       int      `json:"service_type"`         // 服务类型
	ServiceMoney      string   `json:"service_money"`        // 服务金额
	AutoClose         int      `json:"auto_close"`           // 自动关闭
	CloseTime         int      `json:"close_time"`           // 关闭时间
	CategorySet       int      `json:"category_set"`         // 分类设置
	IsDelete          int      `json:"is_delete"`            // 是否删除
	AppID             uint64   `json:"app_id"`               // 应用ID
	CreateTime        string   `json:"create_time"`          // 创建时间
	UpdateTime        string   `json:"update_time"`          // 更新时间
}

type H5Currency struct {
	Unit         string `json:"unit"`          // 单位
	IsOpen       string `json:"is_open"`       // 是否启用
	UnitPosition string `json:"unit_position"` // 单位位置
	Vices        struct {
		ViceUnit         string `json:"vice_unit"`          // 辅币单位
		ViceUnitPosition string `json:"vice_unit_position"` // 辅币单位位置
		UnitRate         string `json:"unit_rate"`          // 单位换算率
	} `json:"vices"`
}

type H5 struct {
	IsCallService      string `json:"is_call_service"`       // 是否呼叫服务
	IsCustomerOrder    string `json:"is_customer_order"`     // 是否顾客订单
	IsVoiceRemind      string `json:"is_voice_remind"`       // 是否语音提醒
	IsShowSoldOut      string `json:"is_show_sold_out"`      // 是否显示售罄
	IsBuffetOrderLimit string `json:"is_buffet_order_limit"` // 是否自助餐订单限制
	BuffetOrderLimit   struct {
		IsLimitTime string `json:"is_limit_time"` // 是否限制时间
		LimitTime   string `json:"limit_time"`    // 限制时间
		IsLimitNum  string `json:"is_limit_num"`  // 是否限制数量
		LimitNum    string `json:"limit_num"`     // 限制数量
	} `json:"buffet_order_limit"` // 自助餐订单限制
	IsOrderLimit string `json:"is_order_limit"` // 是否订单限制
	OrderLimit   struct {
		IsLimitTime string `json:"is_limit_time"` // 是否限制时间
		LimitTime   string `json:"limit_time"`    // 限制时间
		IsLimitNum  string `json:"is_limit_num"`  // 是否限制数量
		LimitNum    string `json:"limit_num"`     // 限制数量
	} `json:"order_limit"` // 订单限制
	Language          []string   `json:"language"`              // 语言
	DefaultLanguage   string     `json:"default_language"`      // 默认语言
	IsShowScanSoldOut int        `json:"is_show_scan_sold_out"` // 是否显示扫码售罄
	KitchenLanguage   string     `json:"kitchen_language"`      // 厨房语言
	LanguageList      []Language `json:"language_list"`         // 语言列表
}

type H5BaseInfo struct {
	IsOpenH5Order int                 `json:"is_open_h5_order"` // 是否开启扫码点餐接单 0不开启, 1开启
	Desk          Desk                `json:"desk"`             // 桌台信息
	Company       Company             `json:"company"`          // 商家信息
	H5            setting.H5          `json:"h5"`               // 扫码H5设置
	Buffet        setting.BuffetResp  `json:"buffet"`           // 自助餐设置
	Currency      setting.Currency    `json:"currency"`         // 货币设置
	CloudBasic    setting.CloudBasic  `json:"cloud"`            // 云端基础信息
	Business      setting.Business    `json:"business"`         // 门店业务设置.为了前端业务统一，这个字段实际未返回任何东西
	Kitchen       setting.KitchenResp `json:"kitchen"`          // 厨显设置
}

type Menu struct {
	LanguageList    []dto.LanguageItem `json:"language_list"`    // 语言列表
	Language        []string           `json:"language"`         // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
	DefaultLanguage string             `json:"default_language"` // 默认语言
}

type MenuBaseInfo struct {
	Currency      setting.Currency   `json:"currency"`         // 货币设置
	CloudBasic    setting.CloudBasic `json:"cloud"`            // 云端基础信息
	Company       Company            `json:"company"`          // 商家信息
	IsShowSoldOut bool               `json:"is_show_sold_out"` // 电子菜单是否显示售罄商品
	Menu          Menu               `json:"menu"`             // 菜单设置
}

type GetBaseInfoResponse struct {
	TableInfo struct {
		TableID        uint64 `json:"table_id"`
		TableNo        string `json:"table_no"`
		Sort           int    `json:"sort"`
		AreaID         uint64 `json:"area_id"`
		TypeID         uint64 `json:"type_id"`
		Status         int    `json:"status"`
		SwitchStatus   int    `json:"switch_status"`
		AreaName       string `json:"area_name"`
		TypeName       string `json:"type_name"`
		ShopSupplierID uint64 `json:"shop_supplier_id"`
		MinNum         int    `json:"min_num"`
		MaxNum         int    `json:"max_num"`
		AppID          uint64 `json:"app_id"`
		BindInfo       string `json:"bind_info"`
		QrcodeValue    string `json:"qrcode_value"`
		IsBind         int    `json:"is_bind"`
		CreateTime     string `json:"create_time"`
		UpdateTime     string `json:"update_time"`
		UnderwayOrder  struct {
			StateText           string `json:"state_text"`
			OrderSourceText     string `json:"order_source_text"`
			OrderTypeText       string `json:"order_type_text"`
			DeliverText         string `json:"deliver_text"`
			ElapsedTime         int    `json:"elapsed_time"`
			PayTimeText         string `json:"pay_time_text"`
			BuffetRemainingTime int    `json:"buffet_remaining_time"`
			ActualPrice         int    `json:"actual_price"`
			ActualReceivePrice  string `json:"actual_receive_price"`
			LockElapsedTime     int    `json:"lock_elapsed_time"`
			OrderID             int    `json:"order_id"`
			ParentID            int    `json:"parent_id"`
			OrderName           string `json:"order_name"`
			IsMustNotice        int    `json:"is_must_notice"`
			ExtraTimes          int    `json:"extra_times"`
			MergeID             string `json:"merge_id"`
			IsMerge             int    `json:"is_merge"`
			MergeParentID       int    `json:"merge_parent_id"`
			OrderNo             string `json:"order_no"`
			DeviceID            string `json:"device_id"`
			SettleDeviceID      string `json:"settle_device_id"`
			IsBuffet            int    `json:"is_buffet"`
			TotalPrice          string `json:"total_price"`
			TotalProductPrice   string `json:"total_product_price"`
			OrderPrice          string `json:"order_price"`
			StayTime            int    `json:"stay_time"`
			IsStay              int    `json:"is_stay"`
			CouponID            int    `json:"coupon_id"`
			CouponMoney         string `json:"coupon_money"`
			CouponIDSys         int    `json:"coupon_id_sys"`
			CouponMoneySys      string `json:"coupon_money_sys"`
			PointsMoney         string `json:"points_money"`
			PointsNum           int    `json:"points_num"`
			PayPrice            string `json:"pay_price"`
			FreePayPrice        string `json:"free_pay_price"`
			ChangeDue           string `json:"change_due"`
			OriginalPrice       string `json:"original_price"`
			UpdatePrice         struct {
				Symbol string `json:"symbol"`
				Value  int    `json:"value"`
			} `json:"update_price"`
			BuyerRemark string `json:"buyer_remark"`
			PayType     int    `json:"pay_type"`
			PaySource   string `json:"pay_source"`
			PayStatus   struct {
				Text  string `json:"text"`
				Value int    `json:"value"`
			} `json:"pay_status"`
			IsFree       int    `json:"is_free"`
			FreeRemark   string `json:"free_remark"`
			PayTime      int    `json:"pay_time"`
			PayEndTime   int    `json:"pay_end_time"`
			DeliveryType struct {
				Text  string `json:"text"`
				Value int    `json:"value"`
			} `json:"delivery_type"`
			DeliveryStatus struct {
				Text  string `json:"text"`
				Value int    `json:"value"`
			} `json:"delivery_status"`
			DeliveryTime  int `json:"delivery_time"`
			ReceiptStatus struct {
				Text  string `json:"text"`
				Value int    `json:"value"`
			} `json:"receipt_status"`
			ReceiptTime int `json:"receipt_time"`
			OrderStatus struct {
				Text  string `json:"text"`
				Value int    `json:"value"`
			} `json:"order_status"`
			PointsBonus                       string `json:"points_bonus"`
			IsSettled                         int    `json:"is_settled"`
			TransactionID                     string `json:"transaction_id"`
			OrderSource                       int    `json:"order_source"`
			UserID                            int    `json:"user_id"`
			ShopSupplierID                    uint64 `json:"shop_supplier_id"`
			SupplierMoney                     string `json:"supplier_money"`
			SysMoney                          string `json:"sys_money"`
			RoomID                            int    `json:"room_id"`
			CancelRemark                      string `json:"cancel_remark"`
			VirtualAuto                       int    `json:"virtual_auto"`
			VirtualContent                    string `json:"virtual_content"`
			BagPrice                          string `json:"bag_price"`
			Mealtime                          string `json:"mealtime"`
			RefundMoney                       string `json:"refund_money"`
			RefundConsumptionTax              string `json:"refund_consumption_tax"`
			OrderType                         int    `json:"order_type"`
			FinanceID                         int    `json:"finance_id"`
			TableNo                           string `json:"table_no"`
			IsDelete                          int    `json:"is_delete"`
			DeliverSource                     int    `json:"deliver_source"`
			DeliverStatus                     int    `json:"deliver_status"`
			DriverID                          int    `json:"driver_id"`
			TakeFee                           string `json:"take_fee"`
			DiscountMoney                     string `json:"discount_money"`
			DiscountChangePrice               string `json:"discount_change_price"`
			IsChangePrice                     int    `json:"is_change_price"`
			SmallDiscountType                 int    `json:"small_discount_type"`
			SmallDiffMoney                    string `json:"small_diff_money"`
			SmallAuto                         int    `json:"small_auto"`
			CheckoutDiffMoney                 string `json:"checkout_diff_money"`
			CheckoutDiscountType              int    `json:"checkout_discount_type"`
			DiscountRatio                     int    `json:"discount_ratio"`
			DiscountMethod                    int    `json:"discount_method"`
			CashierID                         int    `json:"cashier_id"`
			CallNo                            string `json:"call_no"`
			EatType                           int    `json:"eat_type"`
			TableID                           uint64 `json:"table_id"`
			TableRemark                       string `json:"table_remark"`
			MealNum                           int    `json:"meal_num"`
			ServiceMoney                      string `json:"service_money"`
			ServiceType                       int    `json:"service_type"`
			SettleType                        int    `json:"settle_type"`
			AutoClose                         int    `json:"auto_close"`
			CloseTime                         int    `json:"close_time"`
			SurplusBalance                    string `json:"surplus_balance"`
			Balance                           string `json:"balance"`
			OnlineMoney                       string `json:"online_money"`
			TradeNo                           string `json:"trade_no"`
			AppID                             uint64 `json:"app_id"`
			SettingServiceMoney               string `json:"setting_service_money"`
			ConsumptionTaxMoney               string `json:"consumption_tax_money"`
			OriginalConsumptionTaxMoney       string `json:"original_consumption_tax_money"`
			TotalProductConsumptionTax        string `json:"total_product_consumption_tax"`
			TotalProductServiceConsumptionTax string `json:"total_product_service_consumption_tax"`
			TotalProductServiceFee            string `json:"total_product_service_fee"`
			ConsumptionTaxType                int    `json:"consumption_tax_type"`
			UserDiscountMoney                 string `json:"user_discount_money"`
			PayFeeMoney                       string `json:"pay_fee_money"`
			IsLock                            int    `json:"is_lock"`
			LockTime                          int    `json:"lock_time"`
			PrintNum                          int    `json:"print_num"`
			BuffetExpiredTime                 int    `json:"buffet_expired_time"`
			BuffetStartTime                   int    `json:"buffet_start_time"`
			LastBuffetTimeLimit               int    `json:"last_buffet_time_limit"`
			CreateTime                        string `json:"create_time"`
			UpdateTime                        string `json:"update_time"`
			DeleteTime                        int    `json:"delete_time"`
		} `json:"underwayOrder"`
	} `json:"tableInfo"` // 桌台信息
	BaseInfo struct {
		Shop     H5Shop     `json:"shop"`
		Currency H5Currency `json:"currency"` // 货币信息
		H5       H5H5       `json:"h5"`
		Name     string     `json:"name"`
		Logo     string     `json:"logo"` // 商家logo
		Buffet   struct {
			IsOpen                   string   `json:"is_open"`                     // 是否开启自助餐
			TabletEndTime            string   `json:"tablet_end_time"`             // 桌台结束时间
			IsRemainContinue         string   `json:"is_remain_continue"`          // 是否继续剩余
			RemainContinueTime       string   `json:"remain_continue_time"`        // 继续剩余时间
			RemainContinueNoticeTime string   `json:"remain_continue_notice_time"` // 继续剩余通知时间
			IsBuyContinue            string   `json:"is_buy_continue"`             // 是否继续购买
			IsAddClock               string   `json:"is_add_clock"`                // 是否添加时钟
			IsBuffetDiscount         string   `json:"is_buffet_discount"`          // 是否自助餐折扣
			AddClock                 []string `json:"add_clock"`                   // 添加时钟
		} `json:"buffet"`
		LicenseRemainingDays int `json:"license_remaining_days"` // 许可证剩余天数
		IsAcceptScanOrder    int `json:"is_accept_scan_order"`   // 是否接受扫码订单
		CloudBasic           struct {
			Base struct {
				BrandName                  string `json:"brand_name"`                     // 品牌名称
				BrandLogo                  string `json:"brand_logo"`                     // 品牌logo
				BrandLogoLong              string `json:"brand_logo_long"`                // 品牌logo长
				BrowserLogo                string `json:"browser_logo"`                   // 浏览器logo
				BrowserTitle               string `json:"browser_title"`                  // 浏览器标题
				ExpirationReminder         int    `json:"expiration_reminder"`            // 过期提醒
				MemberDefaultAvatar        string `json:"member_default_avatar"`          // 会员默认头像
				AuthCodeBindValidityPeriod string `json:"auth_code_bind_validity_period"` // 授权码绑定有效期
			} `json:"base"`
		} `json:"cloud_basic"`
		Kitchen struct {
			IsOpen string `json:"is_open"` // 是否开启厨房
		} `json:"kitchen"`
		IsCloudDeploy bool `json:"is_cloud_deploy"` // 是否云部署
	} `json:"baseInfo"`
	OrderInfo struct {
		IsBuffet int `json:"is_buffet"` // 是否自助餐
	} `json:"orderInfo"`
}

// 商家信息
type Shop struct {
	Uuid              uint64 `json:"uuid"`                 // 商家ID
	Name              string `json:"name"`                 // 商家名称
	RealName          string `json:"real_name"`            // 商家真实名称
	LinkName          string `json:"link_name"`            // 联系人名称
	LinkPhone         string `json:"link_phone"`           // 联系人电话
	Logo              string `json:"logo"`                 // 商家logo
	SaleStock         int    `json:"sale_stock"`           // 销售库存
	IsOpenMember      int    `json:"is_open_member"`       // 是否开启会员
	IsOpenTablet      int    `json:"is_open_tablet"`       // 是否开启桌台
	IsOpenH5          int    `json:"is_open_scan"`         // 是否开启扫码
	IsOpenAssistant   int    `json:"is_open_assistant"`    // 是否开启助手
	IsOpenKitchenKds  int    `json:"is_open_kitchen_kds"`  // 是否开启厨房KDS
	IsOpenBuffet      int    `json:"is_open_buffet"`       // 是否开启自助餐
	IsAcceptScanOrder int    `json:"is_accept_scan_order"` // 是否接受扫码订单
	IsOpenLocalPrint  int    `json:"is_open_local_print"`  // 是否开启本地打印
	CashLimit         int    `json:"cash_limit"`           // 现金限额
	KitchenLimit      int    `json:"kitchen_limit"`        // 厨房限额
	TabletLimit       int    `json:"tablet_limit"`         // 桌台限额
	AssistantLimit    int    `json:"assistant_limit"`      // 助手限额
	TableLimit        int    `json:"table_limit"`          // 桌台限额
	PrinterLimit      int    `json:"printer_limit"`        // 打印机限额
	Timezone          string `json:"timezone"`             // 时区
	Languages         string `json:"languages"`            // 语言
	Address           string `json:"address"`              // 地址

}

// 货币信息
type Currency struct {
	Unit         string `json:"unit"`          // 单位
	IsOpen       string `json:"is_open"`       // 是否启用
	UnitPosition string `json:"unit_position"` // 单位位置
	Vices        Vices  `json:"vices"`
}

// 辅币信息
type Vices struct {
	ViceUnit         string `json:"vice_unit"`          // 辅币单位
	ViceUnitPosition string `json:"vice_unit_position"` // 辅币单位位置
	UnitRate         string `json:"unit_rate"`          // 单位换算率
}

// H5信息
type H5H5 struct {
	IsCallService      string `json:"is_call_service"`       // 是否呼叫服务
	IsCustomerOrder    string `json:"is_customer_order"`     // 是否顾客订单
	IsVoiceRemind      string `json:"is_voice_remind"`       // 是否语音提醒
	IsShowSoldOut      string `json:"is_show_sold_out"`      // 是否显示售罄
	IsBuffetOrderLimit string `json:"is_buffet_order_limit"` // 是否自助餐订单限制
	BuffetOrderLimit   struct {
		IsLimitTime string `json:"is_limit_time"` // 是否限制时间
		LimitTime   string `json:"limit_time"`    // 限制时间
		IsLimitNum  string `json:"is_limit_num"`  // 是否限制数量
		LimitNum    string `json:"limit_num"`     // 限制数量
	} `json:"buffet_order_limit"` // 自助餐订单限制
	IsOrderLimit string `json:"is_order_limit"` // 是否订单限制
	OrderLimit   struct {
		IsLimitTime string `json:"is_limit_time"` // 是否限制时间
		LimitTime   string `json:"limit_time"`    // 限制时间
		IsLimitNum  string `json:"is_limit_num"`  // 是否限制数量
		LimitNum    string `json:"limit_num"`     // 限制数量
	} `json:"order_limit"` // 订单限制
	Language          []string   `json:"language"`              // 语言
	DefaultLanguage   string     `json:"default_language"`      // 默认语言
	IsShowScanSoldOut int        `json:"is_show_scan_sold_out"` // 是否显示扫码售罄
	KitchenLanguage   string     `json:"kitchen_language"`      // 厨房语言
	LanguageList      []Language `json:"language_list"`         // 语言列表
}

type Language struct {
	Key   string `json:"key"`                 // 语言键
	Value string `json:"value"`               // 语言值
	I     string `json:"i"`                   // 语言索引
	Index string `json:"index" copier:"name"` // 语言索引
}

type H5BuffetResponse struct {
	IsOpen                   string   `json:"is_open"`                     // 是否开启自助餐
	TabletEndTime            string   `json:"tablet_end_time"`             // 桌台结束时间
	IsRemainContinue         string   `json:"is_remain_continue"`          // 是否继续剩余
	RemainContinueTime       string   `json:"remain_continue_time"`        // 继续剩余时间
	RemainContinueNoticeTime string   `json:"remain_continue_notice_time"` // 继续剩余通知时间
	IsBuyContinue            string   `json:"is_buy_continue"`             // 是否继续购买
	IsAddClock               string   `json:"is_add_clock"`                // 是否添加时钟
	IsBuffetDiscount         string   `json:"is_buffet_discount"`          // 是否自助餐折扣
	AddClock                 []string `json:"add_clock"`                   // 添加时钟
}

type H5DeskInfo struct {
	TableID        uint64 `json:"table_id"`
	TableNo        string `json:"table_no"`
	Sort           int    `json:"sort"`
	AreaID         uint64 `json:"area_id"`
	TypeID         uint64 `json:"type_id"`
	Status         uint   `json:"status"`
	SwitchStatus   uint   `json:"switch_status"`
	AreaName       string `json:"area_name"`
	TypeName       string `json:"type_name"`
	ShopSupplierID uint64 `json:"shop_supplier_id"`
	MinNum         int    `json:"min_num"`
	MaxNum         int    `json:"max_num"`
	AppID          uint64 `json:"app_id"`
	BindInfo       string `json:"bind_info"`
	QrcodeValue    string `json:"qrcode_value"`
	IsBind         int    `json:"is_bind"`
	CreateTime     string `json:"create_time"`
	UpdateTime     string `json:"update_time"`
	UnderwayOrder  struct {
		StateText           string `json:"state_text"`
		OrderSourceText     string `json:"order_source_text"`
		OrderTypeText       string `json:"order_type_text"`
		DeliverText         string `json:"deliver_text"`
		ElapsedTime         int    `json:"elapsed_time"`
		PayTimeText         string `json:"pay_time_text"`
		BuffetRemainingTime int    `json:"buffet_remaining_time"`
		ActualPrice         int    `json:"actual_price"`
		ActualReceivePrice  string `json:"actual_receive_price"`
		LockElapsedTime     int    `json:"lock_elapsed_time"`
		OrderID             int    `json:"order_id"`
		ParentID            int    `json:"parent_id"`
		OrderName           string `json:"order_name"`
		IsMustNotice        int    `json:"is_must_notice"`
		ExtraTimes          int    `json:"extra_times"`
		MergeID             string `json:"merge_id"`
		IsMerge             int    `json:"is_merge"`
		MergeParentID       int    `json:"merge_parent_id"`
		OrderNo             string `json:"order_no"`
		DeviceID            string `json:"device_id"`
		SettleDeviceID      string `json:"settle_device_id"`
		IsBuffet            int    `json:"is_buffet"`
		TotalPrice          string `json:"total_price"`
		TotalProductPrice   string `json:"total_product_price"`
		OrderPrice          string `json:"order_price"`
		StayTime            int    `json:"stay_time"`
		IsStay              int    `json:"is_stay"`
		CouponID            int    `json:"coupon_id"`
		CouponMoney         string `json:"coupon_money"`
		CouponIDSys         int    `json:"coupon_id_sys"`
		CouponMoneySys      string `json:"coupon_money_sys"`
		PointsMoney         string `json:"points_money"`
		PointsNum           int    `json:"points_num"`
		PayPrice            string `json:"pay_price"`
		FreePayPrice        string `json:"free_pay_price"`
		ChangeDue           string `json:"change_due"`
		OriginalPrice       string `json:"original_price"`
		UpdatePrice         struct {
			Symbol string `json:"symbol"`
			Value  int    `json:"value"`
		} `json:"update_price"`
		BuyerRemark string `json:"buyer_remark"`
		PayType     int    `json:"pay_type"`
		PaySource   string `json:"pay_source"`
		PayStatus   struct {
			Text  string `json:"text"`
			Value int    `json:"value"`
		} `json:"pay_status"`
		IsFree       int    `json:"is_free"`
		FreeRemark   string `json:"free_remark"`
		PayTime      int    `json:"pay_time"`
		PayEndTime   int    `json:"pay_end_time"`
		DeliveryType struct {
			Text  string `json:"text"`
			Value int    `json:"value"`
		} `json:"delivery_type"`
		DeliveryStatus struct {
			Text  string `json:"text"`
			Value int    `json:"value"`
		} `json:"delivery_status"`
		DeliveryTime  int `json:"delivery_time"`
		ReceiptStatus struct {
			Text  string `json:"text"`
			Value int    `json:"value"`
		} `json:"receipt_status"`
		ReceiptTime int `json:"receipt_time"`
		OrderStatus struct {
			Text  string `json:"text"`
			Value int    `json:"value"`
		} `json:"order_status"`
		PointsBonus                       string `json:"points_bonus"`
		IsSettled                         int    `json:"is_settled"`
		TransactionID                     string `json:"transaction_id"`
		OrderSource                       int    `json:"order_source"`
		UserID                            int    `json:"user_id"`
		ShopSupplierID                    uint64 `json:"shop_supplier_id"`
		SupplierMoney                     string `json:"supplier_money"`
		SysMoney                          string `json:"sys_money"`
		RoomID                            int    `json:"room_id"`
		CancelRemark                      string `json:"cancel_remark"`
		VirtualAuto                       int    `json:"virtual_auto"`
		VirtualContent                    string `json:"virtual_content"`
		BagPrice                          string `json:"bag_price"`
		Mealtime                          string `json:"mealtime"`
		RefundMoney                       string `json:"refund_money"`
		RefundConsumptionTax              string `json:"refund_consumption_tax"`
		OrderType                         int    `json:"order_type"`
		FinanceID                         int    `json:"finance_id"`
		TableNo                           string `json:"table_no"`
		IsDelete                          int    `json:"is_delete"`
		DeliverSource                     int    `json:"deliver_source"`
		DeliverStatus                     int    `json:"deliver_status"`
		DriverID                          int    `json:"driver_id"`
		TakeFee                           string `json:"take_fee"`
		DiscountMoney                     string `json:"discount_money"`
		DiscountChangePrice               string `json:"discount_change_price"`
		IsChangePrice                     int    `json:"is_change_price"`
		SmallDiscountType                 int    `json:"small_discount_type"`
		SmallDiffMoney                    string `json:"small_diff_money"`
		SmallAuto                         int    `json:"small_auto"`
		CheckoutDiffMoney                 string `json:"checkout_diff_money"`
		CheckoutDiscountType              int    `json:"checkout_discount_type"`
		DiscountRatio                     int    `json:"discount_ratio"`
		DiscountMethod                    int    `json:"discount_method"`
		CashierID                         int    `json:"cashier_id"`
		CallNo                            string `json:"call_no"`
		EatType                           int    `json:"eat_type"`
		TableID                           uint64 `json:"table_id"`
		TableRemark                       string `json:"table_remark"`
		MealNum                           int    `json:"meal_num"`
		ServiceMoney                      string `json:"service_money"`
		ServiceType                       int    `json:"service_type"`
		SettleType                        int    `json:"settle_type"`
		AutoClose                         int    `json:"auto_close"`
		CloseTime                         int    `json:"close_time"`
		SurplusBalance                    string `json:"surplus_balance"`
		Balance                           string `json:"balance"`
		OnlineMoney                       string `json:"online_money"`
		TradeNo                           string `json:"trade_no"`
		AppID                             uint64 `json:"app_id"`
		SettingServiceMoney               string `json:"setting_service_money"`
		ConsumptionTaxMoney               string `json:"consumption_tax_money"`
		OriginalConsumptionTaxMoney       string `json:"original_consumption_tax_money"`
		TotalProductConsumptionTax        string `json:"total_product_consumption_tax"`
		TotalProductServiceConsumptionTax string `json:"total_product_service_consumption_tax"`
		TotalProductServiceFee            string `json:"total_product_service_fee"`
		ConsumptionTaxType                int    `json:"consumption_tax_type"`
		UserDiscountMoney                 string `json:"user_discount_money"`
		PayFeeMoney                       string `json:"pay_fee_money"`
		IsLock                            int    `json:"is_lock"`
		LockTime                          int    `json:"lock_time"`
		PrintNum                          int    `json:"print_num"`
		BuffetExpiredTime                 int    `json:"buffet_expired_time"`
		BuffetStartTime                   int    `json:"buffet_start_time"`
		LastBuffetTimeLimit               int    `json:"last_buffet_time_limit"`
		CreateTime                        string `json:"create_time"`
		UpdateTime                        string `json:"update_time"`
		DeleteTime                        int    `json:"delete_time"`
	} `json:"underwayOrder"`
}

type H5BuffetList []H5Buffet
type H5Buffet struct {
	NameText                 string `json:"name_text"`
	Id                       uint64 `json:"id"`
	Name                     string `json:"name"`
	Price                    string `json:"price"`
	BuyLimitStatus           int    `json:"buy_limit_status"`
	IsComb                   int    `json:"is_comb"`
	SaleNum                  int    `json:"sale_num"`
	TimeLimit                int    `json:"time_limit"`
	Status                   int    `json:"status"`
	Sort                     int    `json:"sort"`
	IsRemainContinue         int    `json:"is_remain_continue"`
	RemainContinueTime       uint   `json:"remain_continue_time"`
	RemainContinueNoticeTime uint   `json:"remain_continue_notice_time"`
	AppId                    int    `json:"app_id"`
	ShopSupplierId           int    `json:"shop_supplier_id"`
	CreateTime               string `json:"create_time"`
	UpdateTime               string `json:"update_time"`
	DeleteTime               int    `json:"delete_time"`
	BuffetTaxes              []struct {
		TaxRate       int    `json:"tax_rate"`
		Id            int    `json:"id"`
		BuffetId      int    `json:"buffet_id"`
		BuffetTaxType int    `json:"buffet_tax_type"`
		TaxCategoryId int    `json:"tax_category_id"`
		AppId         int    `json:"app_id"`
		CreateTime    string `json:"create_time"`
		UpdateTime    string `json:"update_time"`
	} `json:"buffetTaxes"`
	BuffetCustomerType []H5BuffetCustomerType `json:"buffetCustomerType"`
}

type H5BuffetCustomerType struct {
	NameText       string `json:"name_text"`
	Id             uint64 `json:"id"`
	BuffetId       int    `json:"buffet_id"`
	CustomerTypeId uint64 `json:"customer_type_id"`
	Price          string `json:"price"`
	AppId          int    `json:"app_id"`
	ShopSupplierId int    `json:"shop_supplier_id"`
	CreateTime     string `json:"create_time"`
	UpdateTime     string `json:"update_time"`
}
