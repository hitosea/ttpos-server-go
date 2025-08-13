package consts

import "ttpos-bmp/app/ttpos-erp/api/item"

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

func ItemGroupToEnum(itemGroup ItemGroup) item.ItemGroup {
	switch itemGroup {
	case ItemGroupProducts:
		return item.ItemGroup_Products
	case ItemGroupRawMaterial:
		return item.ItemGroup_RawMaterial
	default:
		return item.ItemGroup_Others
	}
}

func EnumToItemGroup(itemGroup item.ItemGroup) ItemGroup {
	switch itemGroup {
	case item.ItemGroup_Products:
		return ItemGroupProducts
	case item.ItemGroup_RawMaterial:
		return ItemGroupRawMaterial
	default:
		return ItemGroupOthers
	}
}

func ParseItemGroup(name string) ItemGroup {
	switch name {
	case string(ItemGroupProducts):
		return ItemGroupProducts
	case string(ItemGroupRawMaterial):
		return ItemGroupRawMaterial
	default:
		return ItemGroupOthers
	}
}
