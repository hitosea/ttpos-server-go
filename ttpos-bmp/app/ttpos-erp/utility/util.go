package utility

import (
	"fmt"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/utility/uuid"
)

func GenItemCode(prefix string) string {
	return fmt.Sprintf("%s%d", prefix, uuid.MustGetID())
}

// ParseItemGroupFromString  从字符串解析物品分组枚举
func ParseItemGroupFromString(itemGroupStr string) item.ItemGroup {
	switch itemGroupStr {
	case string(consts.ItemGroupProducts):
		return item.ItemGroup_Products
	case string(consts.ItemGroupRawMaterial):
		return item.ItemGroup_RawMaterial
	case string(consts.ItemGroupPosAttribute):
		return item.ItemGroup_PosAttribute
	case string(consts.ItemGroupPosAddon):
		return item.ItemGroup_PosAddon
	default:
		return item.ItemGroup_Others
	}
}

// ItemGroupToString 将物品分组枚举转换为字符串
func ItemGroupToString(itemGroup item.ItemGroup) string {
	switch itemGroup {
	case item.ItemGroup_Products:
		return string(consts.ItemGroupProducts)
	case item.ItemGroup_Package:
		return string(consts.ItemGroupProducts)
	case item.ItemGroup_RawMaterial:
		return string(consts.ItemGroupRawMaterial)
	case item.ItemGroup_PosAttribute:
		return string(consts.ItemGroupPosAttribute)
	case item.ItemGroup_PosAddon:
		return string(consts.ItemGroupPosAddon)
	default:
		return ""
	}
}
