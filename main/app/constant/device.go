package constant

// 上下文场景值
const (
	SceneMemberOrder = "member_order" // 会员端订单
)

const (
	SourceShop      = "shop"      // 商家
	SourceCashier   = "cashier"   // 收银机
	SourceMember    = "member"    // 会员端
	SourceTablet    = "tablet"    // 平板端
	SourceKitchen   = "kitchen"   // 厨显端
	SourceAssistant = "assistant" // 点餐助手
	SourceH5        = "h5"        // H5
	SourceRider     = "rider"     // 骑手端. skootar
	SourceKiosk     = "kiosk"     // 自助点餐机
)

// SaleBillSource 销售账单来源（整数类型）
const (
	SaleBillSourceDefault   = 0 // 默认值
	SaleBillSourceCashier   = 1 // 收银机
	SaleBillSourceAssistant = 2 // 点餐助手
	SaleBillSourceTablet    = 3 // 平板
	SaleBillSourceH5        = 4 // H5
	SaleBillSourceMember    = 5 // 会员端
	SaleBillSourceKiosk     = 6 // 自助点餐机
)

var SourceTextMap = map[string]string{
	SourceCashier:   "收银端",
	SourceMember:    "会员端",
	SourceAssistant: "点餐助手",
	SourceShop:      "商家后台",
	SourceTablet:    "平板端",
	SourceH5:        "扫码点餐",
	SourceKiosk:     "自助点餐机",
}

const (
	BrandA11500         = "A1-1500"           //不带打印机的 compax收银机器
	BrandA11510P        = "A2-1510P"          //自带打印机的 compax收银机器
	BrandT2             = "T2"                //自带打印机的 商米收银机器
	BrandT2S            = "T2s"               //自带打印机的 商米收银机器
	BrandD2S            = "D2s"               //自带打印机的 商米收银机器
	BrandD2SPlus        = "D2s_PLUS"          //自带打印机的 商米收银机器
	BrandD2sPlus2ndStgl = "D2s_PLUS_2nd_STGL" //自带打印机的 商米收银机器
	BrandT2MINIS        = "T2mini_s"          //自带打印机的 商米小屏收银机器
	BrandD1             = "D1"                //自带打印机的 IMIN收银机器
	BrandD4             = "D4"                //自带打印机的 IMIN收银机器
)

// 所有的机器
var BrandsAll = []string{
	BrandA11500,
	BrandA11510P,
	BrandT2S,
	BrandT2MINIS,
	BrandT2,
	BrandD2SPlus,
	BrandD2S,
	BrandD2sPlus2ndStgl,
	BrandD1,
	BrandD4,
}

// 所有带打印机的机器
var BrandsPrints = []string{
	BrandA11510P,
	BrandT2S,
	BrandT2MINIS,
	BrandT2,
	BrandD2SPlus,
	BrandD2S,
	BrandD2sPlus2ndStgl,
	BrandD1,
	BrandD4,
}

// 商米所有带打印机的机器
var SunmiAllPrints = []string{
	BrandT2,
	BrandT2S,
	BrandT2MINIS,
	BrandD2SPlus,
	BrandD2S,
	BrandD2sPlus2ndStgl,
}
