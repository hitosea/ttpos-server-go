package req

import "ttpos-server-go/app/dto"

type ParamCreateCompany struct {
	Name             string   `json:"name"`                // 公司名称
	Logo             string   `json:"logo"`                // 公司logo
	AuthDay          int      `json:"auth_day"`            // 授权天数
	AuthStartTime    string   `json:"auth_start_time"`     // 授权开始时间
	LinkPhone        string   `json:"link_phone"`          // 联系电话
	CashLimit        int      `json:"cash_limit"`          // 现金限制
	KitchenLimit     int      `json:"kitchen_limit"`       // 厨房限制
	TabletLimit      int      `json:"tablet_limit"`        // 平板限制
	AssistantLimit   int      `json:"assistant_limit"`     // 助手限制
	UserName         string   `json:"user_name"`           // 用户名
	Password         string   `json:"password"`            // 密码
	ConfirmPassword  string   `json:"confirm_password"`    // 确认密码
	Address          string   `json:"address"`             // 地址
	DeployMode       int      `json:"deploy_mode"`         // 部署模式
	ChainNumber      string   `json:"chain_number"`        // 连锁编号
	TimeZone         string   `json:"timezone"`            // 时区
	Languages        []string `json:"languages"`           // 语言列表
	IsOpenMember     int      `json:"is_open_member"`      // 是否开启会员
	IsOpenTablet     int      `json:"is_open_tablet"`      // 是否开启平板
	IsOpenScan       int      `json:"is_open_scan"`        // 是否开启扫码
	IsOpenAssistant  int      `json:"is_open_assistant"`   // 是否开启助手
	IsOpenKitchenKds int      `json:"is_open_kitchen_kds"` // 是否开启厨房KDS
	IsOpenBuffet     int      `json:"is_open_buffet"`      // 是否开启自助餐
	IsOpenH5Order    int      `json:"is_open_h5_order"`    // 是否接受扫码订单
	IsOpenLocalPrint int      `json:"is_open_local_print"` // 是否开启本地打印
	TableLimit       int      `json:"table_limit"`         // 桌位限制
	PrinterLimit     int      `json:"printer_limit"`       // 打印机限制
}

// UpdateCompanyInfoReq 更新门店信息请求
type UpdateCompanyInfoReq struct {
	Uuid   uint64 `json:"uuid" binding:"required"` // 门店UUID
	Name   string `json:"name"`                    // 门店名称
	Logo   string `json:"logo"`                    // 门店logo
	Status *int   `json:"status"`                  // 状态 1-启用 0-禁用
}

type GetCompanyInfoReq struct {
	Uuid uint64 `form:"uuid" binding:"required"` // 门店UUID
}

type ParamUpdateCompany struct {
	ID               int      `json:"id"`                  // 公司ID
	Name             string   `json:"name"`                // 公司名称
	Logo             string   `json:"logo"`                // 公司logo
	AuthDay          int      `json:"auth_day"`            // 授权天数
	AuthStartTime    string   `json:"auth_start_time"`     // 授权开始时间
	LinkPhone        string   `json:"link_phone"`          // 联系电话
	CashLimit        int      `json:"cash_limit"`          // 现金限制
	KitchenLimit     int      `json:"kitchen_limit"`       // 厨房限制
	TabletLimit      int      `json:"tablet_limit"`        // 平板限制
	AssistantLimit   int      `json:"assistant_limit"`     // 助手限制
	UserName         string   `json:"user_name"`           // 用户名
	Password         string   `json:"password"`            // 密码
	ConfirmPassword  string   `json:"confirm_password"`    // 确认密码
	Address          string   `json:"address"`             // 地址
	DeployMode       int      `json:"deploy_mode"`         // 部署模式
	ChainNumber      string   `json:"chain_number"`        // 连锁编号
	TimeZone         string   `json:"timezone"`            // 时区
	Languages        []string `json:"languages"`           // 语言列表
	IsOpenMember     int      `json:"is_open_member"`      // 是否开启会员
	IsOpenTablet     int      `json:"is_open_tablet"`      // 是否开启平板
	IsOpenScan       int      `json:"is_open_scan"`        // 是否开启扫码
	IsOpenAssistant  int      `json:"is_open_assistant"`   // 是否开启助手
	IsOpenKitchenKds int      `json:"is_open_kitchen_kds"` // 是否开启厨房KDS
	IsOpenBuffet     int      `json:"is_open_buffet"`      // 是否开启自助餐
	IsOpenH5Order    int      `json:"is_open_h5_order"`    // 是否接受扫码订单
	IsOpenLocalPrint int      `json:"is_open_local_print"` // 是否开启本地打印
	TableLimit       int      `json:"table_limit"`         // 桌位限制
	PrinterLimit     int      `json:"printer_limit"`       // 打印机限制
}

// 商家名称、商家LOGO、地址、经纬度、公司名称、联系电话、税号、语言、时区
type UpdateCompanySettingReq struct {
	Uuid        uint64             `json:"uuid" binding:"required"`           // 门店UUID
	Name        string             `json:"name" binding:"required,max=100"`   // 店铺名称
	LogoUrl     string             `json:"logo_url" binding:"required"`       // 店铺logo，上传后保存url，必填
	Address     string             `json:"address" binding:"max=500"`         // 地址，必填，最大500个字符
	Coordinates string             `json:"coordinates"`                       // 经纬度
	CompanyName string             `json:"company_name" binding:"max=500"`    // 公司名称，区别于店铺名称，最大500个字符
	Phone       string             `json:"phone" binding:"required,max=20"`   // 联系电话，必填，最大20个字符
	TaxNumber   string             `json:"tax_number"`                        // 税号
	Language    []dto.LanguageItem `json:"language" binding:"required,min=1"` // 系统语言，必填，至少一个
	TimeZone    string             `json:"time_zone" binding:"required"`      // 时区，必填
}
