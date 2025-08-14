package consts

var Limit999 = 999

type ModeOfPayment string

const (
	ModeOfPaymentCash    ModeOfPayment = "Cash"
	ModeOfPaymentBalance ModeOfPayment = "Balance"
)

type ItemGroup string

const (
	// ItemGroupRawMaterial 原材料
	ItemGroupRawMaterial ItemGroup = "Raw Material"
	// ItemGroupProducts 商品
	ItemGroupProducts ItemGroup = "Products"
	// ItemGroupOthers 其他
	ItemGroupOthers ItemGroup = ""
)

const (
	ItemCodePrefixProduct     = "SP"  //商品前缀
	ItemCodePrefixRawMaterial = "WPR" //原材料前缀
)
