package value_object

import (
	"errors"
	"strconv"
	"strings"
)

// IDType ID类型枚举
type IDTypeEnum string

// TTPOS ID 前缀常量定义
const (
	// 分类前缀
	PrefixCategory IDTypeEnum = "TTPOS-CAT-" // 分类

	// 套餐前缀
	PrefixPackageGroup IDTypeEnum = "TTPOS-PACKAGE-GROUP-" // 套餐组
	PrefixPackageItem  IDTypeEnum = "TTPOS-PACKAGE-ITEM-"  // 套餐项

	// 商品前缀
	PrefixItem    IDTypeEnum = "TTPOS-ITEM-"    // 商品
	PrefixPackage IDTypeEnum = "TTPOS-PACKAGE-" // 套餐

	// 修饰符组前缀
	PrefixFlavorGroup IDTypeEnum = "TTPOS-FLAVOR-GROUP-" // 口味组
	PrefixSauceGroup  IDTypeEnum = "TTPOS-SAUCE-GROUP-"  // 酱料组
	PrefixAttrGroup   IDTypeEnum = "TTPOS-ATTR-GROUP-"   // 属性组

	// 修饰符前缀
	PrefixFlavor IDTypeEnum = "TTPOS-FLAVOR-" // 口味
	PrefixSauce  IDTypeEnum = "TTPOS-SAUCE-"  // 酱料
	PrefixAttr   IDTypeEnum = "TTPOS-ATTR-"   // 属性
)

// GetPrefixCategoryID 获取分类ID
func (s IDTypeEnum) GetPrefix(platform string) string {
	switch platform {
	case TakeoutPlatformGrab:
		return string(s) // PrefixCategory
	case TakeoutPlatformLineman:
		// Lineman 平台使用简写：TTPOS- 简写为 ''
		prefix := strings.Replace(string(s), "TTPOS-", "", 1)
		prefix = strings.Replace(prefix, "PACKAGE-GROUP-", "PACKAGE-S-", 1)
		prefix = strings.Replace(prefix, "PACKAGE-ITEM-", "PACKAGE-I-", 1)
		prefix = strings.Replace(prefix, "FLAVOR-GROUP-", "FLAVOR-S-", 1)
		prefix = strings.Replace(prefix, "ATTR-GROUP-", "ATTR-S-", 1)
		prefix = strings.Replace(prefix, "SAUCE-GROUP-", "SAUCE-S-", 1)
		//
		return prefix
	default:
		return ""
	}
}

// IDType ID类型枚举
type IDType string

const (
	IDTypeCategory     IDType = "category"      // 分类
	IDTypeItem         IDType = "item"          // 商品
	IDTypePackage      IDType = "package"       // 套餐
	IDTypePackageGroup IDType = "package_group" // 套餐组
	IDTypePackageItem  IDType = "package_item"  // 套餐项
	IDTypeFlavorGroup  IDType = "flavor_group"  // 口味组
	IDTypeFlavor       IDType = "flavor"        // 口味
	IDTypeSauceGroup   IDType = "sauce_group"   // 酱料组
	IDTypeSauce        IDType = "sauce"         // 酱料
	IDTypeAttrGroup    IDType = "attr_group"    // 属性组
	IDTypeAttr         IDType = "attr"          // 属性
	IDTypeCommodity    IDType = "commodity"     // 套餐商品
	IDTypeUnknown      IDType = "unknown"       // 未知类型
)

// ParseResult 解析结果
type ParseResult struct {
	UUID     uint64 // TTPOS UUID
	IDType   IDType // ID类型
	IsMapped bool   // 是否已映射
}

// ParsePlatformID 解析平台ID（统一入口）
// 支持的前缀：
// 参数：
//   - platform: 平台标识 (grab, foodpanda, lineman)
//   - platformID: 平台ID
//
// 返回：
//   - result: 解析结果
//   - err: 错误信息
func ParsePlatformID(platform, platformID string) (*ParseResult, error) {
	switch platform {
	case TakeoutPlatformGrab:
		return parseID(platform, platformID)

	case TakeoutPlatformLineman:
		return parseID(platform, platformID)

	default:
		return nil, errors.New("未知平台")
	}
}

// parseID 解析 Grab 平台的 ID
func parseID(platform string, platformID string) (*ParseResult, error) {
	// 按前缀长度从长到短匹配，避免误匹配
	// 例如：TTPOS-PACKAGE-GROUP- 必须在 TTPOS-PACKAGE- 之前匹配

	// 套餐相关（最长前缀优先）
	if strings.HasPrefix(platformID, PrefixPackageGroup.GetPrefix(platform)) {
		uuid, err := extractUUID(platformID, PrefixPackageGroup.GetPrefix(platform))
		if err != nil {
			return nil, err
		}
		return &ParseResult{UUID: uuid, IDType: IDTypePackageGroup, IsMapped: true}, nil
	}

	// 套餐项（最长前缀优先）
	if strings.HasPrefix(platformID, PrefixPackageItem.GetPrefix(platform)) {
		uuid, err := extractUUID(platformID, PrefixPackageItem.GetPrefix(platform))
		if err != nil {
			return nil, err
		}
		return &ParseResult{UUID: uuid, IDType: IDTypePackageItem, IsMapped: true}, nil
	}

	// 修饰符组（长前缀优先）
	if strings.HasPrefix(platformID, PrefixFlavorGroup.GetPrefix(platform)) {
		uuid, err := extractUUID(platformID, PrefixFlavorGroup.GetPrefix(platform))
		if err != nil {
			return nil, err
		}
		return &ParseResult{UUID: uuid, IDType: IDTypeFlavorGroup, IsMapped: true}, nil
	}

	// 酱料组（长前缀优先）
	if strings.HasPrefix(platformID, PrefixSauceGroup.GetPrefix(platform)) {
		uuid, err := extractUUID(platformID, PrefixSauceGroup.GetPrefix(platform))
		if err != nil {
			return nil, err
		}
		return &ParseResult{UUID: uuid, IDType: IDTypeSauceGroup, IsMapped: true}, nil
	}

	// 属性组（长前缀优先）
	if strings.HasPrefix(platformID, PrefixAttrGroup.GetPrefix(platform)) {
		uuid, err := extractUUID(platformID, PrefixAttrGroup.GetPrefix(platform))
		if err != nil {
			return nil, err
		}
		return &ParseResult{UUID: uuid, IDType: IDTypeAttrGroup, IsMapped: true}, nil
	}

	// 修饰符（短前缀）
	// 口味
	if strings.HasPrefix(platformID, PrefixFlavor.GetPrefix(platform)) {
		uuid, err := extractUUID(platformID, PrefixFlavor.GetPrefix(platform))
		if err != nil {
			return nil, err
		}
		return &ParseResult{UUID: uuid, IDType: IDTypeFlavor, IsMapped: true}, nil
	}

	// 酱料
	if strings.HasPrefix(platformID, PrefixSauce.GetPrefix(platform)) {
		uuid, err := extractUUID(platformID, PrefixSauce.GetPrefix(platform))
		if err != nil {
			return nil, err
		}
		return &ParseResult{UUID: uuid, IDType: IDTypeSauce, IsMapped: true}, nil
	}

	// 属性
	if strings.HasPrefix(platformID, PrefixAttr.GetPrefix(platform)) {
		uuid, err := extractUUID(platformID, PrefixAttr.GetPrefix(platform))
		if err != nil {
			return nil, err
		}
		return &ParseResult{UUID: uuid, IDType: IDTypeAttr, IsMapped: true}, nil
	}

	// 分类
	if strings.HasPrefix(platformID, PrefixCategory.GetPrefix(platform)) {
		uuid, err := extractUUID(platformID, PrefixCategory.GetPrefix(platform))
		if err != nil {
			return nil, err
		}
		return &ParseResult{UUID: uuid, IDType: IDTypeCategory, IsMapped: true}, nil
	}

	// 商品
	if strings.HasPrefix(platformID, PrefixItem.GetPrefix(platform)) {
		uuid, err := extractUUID(platformID, PrefixItem.GetPrefix(platform))
		if err != nil {
			return nil, err
		}
		return &ParseResult{UUID: uuid, IDType: IDTypeItem, IsMapped: true}, nil
	}

	// 套餐
	if strings.HasPrefix(platformID, PrefixPackage.GetPrefix(platform)) {
		uuid, err := extractUUID(platformID, PrefixPackage.GetPrefix(platform))
		if err != nil {
			return nil, err
		}
		return &ParseResult{UUID: uuid, IDType: IDTypePackage, IsMapped: true}, nil
	}

	// 未匹配到任何前缀
	return nil, errors.New("ID格式错误：必须包含有效的 TTPOS- 前缀，不能接单")
}

// extractUUID 提取UUID
func extractUUID(platformID, prefix string) (uint64, error) {
	uuidStr := strings.TrimPrefix(platformID, prefix)
	if uuidStr == "" {
		return 0, errors.New("UUID不能为空")
	}

	uuid, err := strconv.ParseUint(uuidStr, 10, 64)
	if err != nil {
		return 0, errors.New("无效的UUID格式")
	}

	return uuid, nil
}
