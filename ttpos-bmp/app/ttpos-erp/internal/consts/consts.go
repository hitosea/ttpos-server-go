package consts

var Limit100 = 100
var Limit999 = 999
var Limit9999 = 9999

type ModeOfPayment string

const (
	// ModeOfPaymentCash 现金
	ModeOfPaymentCash ModeOfPayment = "Cash"
	// ModeOfPaymentBalance 余额
	ModeOfPaymentBalance ModeOfPayment = "Balance"
	// ModeOfPaymentFreeMeal 免单
	ModeOfPaymentFreeMeal ModeOfPayment = "Free Meal"
)

type ItemGroup string

const (
	// ItemGroupRawMaterial 原材料
	ItemGroupRawMaterial ItemGroup = "Raw Material"
	// ItemGroupProducts 商品
	ItemGroupProducts ItemGroup = "Products"
	// ItemGroupPosAttribute Pos系统中特殊的item ，如 属性/加料
	ItemGroupPosAttribute ItemGroup = "Pos Attribute"
	// ItemGroupPosAddon Pos系统中特殊的item ，如 加料
	ItemGroupPosAddon ItemGroup = "Pos Addon"
	// ItemGroupOthers 其他
	ItemGroupOthers ItemGroup = ""

	ItemGroupPrefixPosAttributeGroup = "SX" // 属性前缀
	ItemGroupPrefixPosAddonGroup     = "JL" // 属性前缀

)

const (
	ItemCodePrefixProduct      = "SP"  //商品前缀
	ItemCodePrefixRawMaterial  = "WPR" //原材料前缀
	ItemCodePrefixPackage      = "TC"  //套餐前缀
	ItemCodePrefixPosAttribute = "SXZ" // 属性前缀
	ItemCodePrefixPosAddon     = "JLZ" // 加料前缀
)

// CustomerName 客户名称
const (
	DefaultCustomerName = "Default" // 默认客户
	MemberCustomerName  = "Member"  // 会员
)

const (
	ContextFakeUser = "ctx_fake_user"
	ContextSiteCode = "erp_site_code"
)

const (
	SiteCodeUat     = "0"
	SiteCodeTtpos   = "1"
	SiteCodeWallace = "4"
)
