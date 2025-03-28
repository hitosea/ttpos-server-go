package constant

const (
	SourceShop      = "shop"      // 商家
	SourceCashier   = "cashier"   // 收银机
	SourceTablet    = "tablet"    // 平板端
	SourceKitchen   = "kitchen"   // 厨显端
	SourceAssistant = "assistant" // 点餐助手
	SourceH5        = "h5"        // H5
)

var SourceTextMap = map[string]string{
	SourceCashier:   "收银端",
	SourceAssistant: "点餐助手",
	SourceShop:      "商家后台",
	SourceTablet:    "平板端",
	SourceH5:        "扫码点餐",
}

const (
	BrandA11500  = "A1-1500"  //不带打印机的 compax收银机器
	BrandA11510P = "A2-1510P" //自带打印机的 compax收银机器
	BrandT2      = "T2"       //自带打印机的 商米收银机器
	BrandT2S     = "T2s"      //自带打印机的 商米收银机器
	BrandD2SPlus = "D2s_PLUS" //自带打印机的 商米收银机器
	BrandT2MINIS = "T2mini_s" //自带打印机的 商米小屏收银机器
)

var BrandsAll = []string{BrandA11500, BrandA11510P, BrandT2S, BrandT2MINIS, BrandT2, BrandD2SPlus} // 所有的机器
var BrandsPrints = []string{BrandA11510P, BrandT2S, BrandT2MINIS, BrandT2, BrandD2SPlus}           // 所有带打印机的机器
var SunmiAllPrints = []string{BrandT2, BrandT2S, BrandT2MINIS, BrandD2SPlus}                       // 商米所有带打印机的机器
