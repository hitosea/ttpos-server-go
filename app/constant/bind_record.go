package constant

const (
	SOURCE_CASHIER   = "cashier"   // 收银机
	SOURCE_TABLET    = "tablet"    // 平板端
	SOURCE_KITCHEN   = "kitchen"   // 厨显端
	SOURCE_ASSISTANT = "assistant" // 点餐助手
)

const (
	BRAND_A1_1500  = "A1-1500"  //不带打印机的 compax收银机器
	BRAND_A1_1510P = "A2-1510P" //自带打印机的 compax收银机器
	BRAND_T2       = "T2"       //自带打印机的 商米收银机器
	BRAND_T2S      = "T2s"      //自带打印机的 商米收银机器
	BRAND_D2S_PLUS = "D2s_PLUS" //自带打印机的 商米收银机器
	BRAND_T2MINI_S = "T2mini_s" //自带打印机的 商米小屏收银机器
)

var BRANDS_ALL = []string{BRAND_A1_1500, BRAND_A1_1510P, BRAND_T2S, BRAND_T2MINI_S, BRAND_T2, BRAND_D2S_PLUS} // 所有的机器
var BRANDS_PRINTS = []string{BRAND_A1_1510P, BRAND_T2S, BRAND_T2MINI_S, BRAND_T2, BRAND_D2S_PLUS}             // 所有带打印机的机器
var BRANDS_SUNMI_ALL_PRINTS = []string{BRAND_T2, BRAND_T2S, BRAND_T2MINI_S, BRAND_D2S_PLUS}                   // 商米所有带打印机的机器
