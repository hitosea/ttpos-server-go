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

// CustomerName 客户名称
const (
	DefaultCustomerName = "Default" // 默认客户
	MemberCustomerName  = "Member"  // 会员
)
