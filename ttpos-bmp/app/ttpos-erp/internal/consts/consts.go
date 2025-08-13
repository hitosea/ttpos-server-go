package consts

var Limit999 = 999

type ModeOfPayment string

const (
	ModeOfPaymentCash    ModeOfPayment = "Cash"
	ModeOfPaymentBalance ModeOfPayment = "Balance"
)

type ItemGroup string

const (
	// ItemGroupProducts 商品
	ItemGroupProducts ItemGroup = "Products"
	// ItemGroupRawMaterial 原材料
	ItemGroupRawMaterial ItemGroup = "Raw Material"
)

const (
	ItemCodePrefixProduct     = "SP"  //商品前缀
	ItemCodePrefixRawMaterial = "WPR" //原材料前缀
)
