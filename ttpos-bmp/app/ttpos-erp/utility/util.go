package utility

import (
	"fmt"
	"strings"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/utility/uuid"

	"github.com/gogf/gf/v2/text/gstr"
)

func GenItemCode(prefix string) string {
	return fmt.Sprintf("%s%d", prefix, uuid.MustGetID())
}

// ParseItemGroupFromString  从字符串解析物品分组枚举
func ParseItemGroupFromString(itemGroupStr string) item.ItemGroup {

	//特殊处理属性组和加料组
	if gstr.Pos(itemGroupStr, consts.ItemGroupPrefixPosAttributeGroup) > -1 {
		return item.ItemGroup_PosAttribute
	}
	if gstr.Pos(itemGroupStr, consts.ItemGroupPrefixPosAddonGroup) > -1 {
		return item.ItemGroup_PosAddon
	}

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

// GetFilterOperator 根据值是否包含 % 返回对应的操作符
// 参数：
//   - value: 查询值
//
// 返回：
//   - operator: "like" 或 "="
func GetFilterOperator(value string) string {
	if strings.Contains(value, "%") {
		return "like"
	}
	return "="
}
