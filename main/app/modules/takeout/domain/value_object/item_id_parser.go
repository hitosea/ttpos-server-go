package value_object

import (
	"errors"
	"strconv"
	"strings"
)

// TTPOS ID 前缀常量定义
const (
	// 分类前缀
	PrefixCategory = "TTPOS-CAT-" // 分类

	// 商品前缀
	PrefixItem    = "TTPOS-ITEM-"    // 商品
	PrefixPackage = "TTPOS-PACKAGE-" // 套餐

	// 套餐前缀
	PrefixPackageGroup = "TTPOS-PACKAGE-GROUP-" // 套餐组
	PrefixPackageItem  = "TTPOS-PACKAGE-ITEM-"  // 套餐项

	// 修饰符组前缀
	PrefixFlavorGroup = "TTPOS-FLAVOR-GROUP-" // 口味组
	PrefixSauceGroup  = "TTPOS-SAUCE-GROUP-"  // 酱料组
	PrefixAttrGroup   = "TTPOS-ATTR-GROUP-"   // 属性组

	// 修饰符前缀
	PrefixFlavor = "TTPOS-FLAVOR-" // 口味
	PrefixSauce  = "TTPOS-SAUCE-"  // 酱料
	PrefixAttr   = "TTPOS-ATTR-"   // 属性
)

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
//
// 支持的前缀：
//   - TTPOS-CAT-            → 分类
//   - TTPOS-ITEM-           → 商品
//   - TTPOS-PACKAGE-        → 套餐
//   - TTPOS-PACKAGE-GROUP-  → 套餐组
//   - TTPOS-PACKAGE-ITEM-   → 套餐项
//   - TTPOS-FLAVOR-GROUP-   → 口味组
//   - TTPOS-FLAVOR-         → 口味
//   - TTPOS-SAUCE-GROUP-    → 酱料组
//   - TTPOS-SAUCE-          → 酱料
//   - TTPOS-ATTR-GROUP-     → 属性组
//   - TTPOS-ATTR-           → 属性
//
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
		return parseGrabID(platformID)

	case TakeoutPlatformLineman, TakeoutPlatformShopeefood:
		// TODO: 其他平台的解析规则
		return nil, errors.New("暂不支持该平台")

	default:
		return nil, errors.New("未知平台")
	}
}

// parseGrabID 解析 Grab 平台的 ID
func parseGrabID(platformID string) (*ParseResult, error) {
	// 按前缀长度从长到短匹配，避免误匹配
	// 例如：TTPOS-PACKAGE-GROUP- 必须在 TTPOS-PACKAGE- 之前匹配

	// 套餐相关（最长前缀优先）
	if strings.HasPrefix(platformID, PrefixPackageGroup) {
		uuid, err := extractUUID(platformID, PrefixPackageGroup)
		if err != nil {
			return nil, err
		}
		return &ParseResult{UUID: uuid, IDType: IDTypePackageGroup, IsMapped: true}, nil
	}

	if strings.HasPrefix(platformID, PrefixPackageItem) {
		uuid, err := extractUUID(platformID, PrefixPackageItem)
		if err != nil {
			return nil, err
		}
		return &ParseResult{UUID: uuid, IDType: IDTypePackageItem, IsMapped: true}, nil
	}

	// 修饰符组（长前缀优先）
	if strings.HasPrefix(platformID, PrefixFlavorGroup) {
		uuid, err := extractUUID(platformID, PrefixFlavorGroup)
		if err != nil {
			return nil, err
		}
		return &ParseResult{UUID: uuid, IDType: IDTypeFlavorGroup, IsMapped: true}, nil
	}

	if strings.HasPrefix(platformID, PrefixSauceGroup) {
		uuid, err := extractUUID(platformID, PrefixSauceGroup)
		if err != nil {
			return nil, err
		}
		return &ParseResult{UUID: uuid, IDType: IDTypeSauceGroup, IsMapped: true}, nil
	}

	if strings.HasPrefix(platformID, PrefixAttrGroup) {
		uuid, err := extractUUID(platformID, PrefixAttrGroup)
		if err != nil {
			return nil, err
		}
		return &ParseResult{UUID: uuid, IDType: IDTypeAttrGroup, IsMapped: true}, nil
	}

	// 修饰符（短前缀）
	if strings.HasPrefix(platformID, PrefixFlavor) {
		uuid, err := extractUUID(platformID, PrefixFlavor)
		if err != nil {
			return nil, err
		}
		return &ParseResult{UUID: uuid, IDType: IDTypeFlavor, IsMapped: true}, nil
	}

	if strings.HasPrefix(platformID, PrefixSauce) {
		uuid, err := extractUUID(platformID, PrefixSauce)
		if err != nil {
			return nil, err
		}
		return &ParseResult{UUID: uuid, IDType: IDTypeSauce, IsMapped: true}, nil
	}

	if strings.HasPrefix(platformID, PrefixAttr) {
		uuid, err := extractUUID(platformID, PrefixAttr)
		if err != nil {
			return nil, err
		}
		return &ParseResult{UUID: uuid, IDType: IDTypeAttr, IsMapped: true}, nil
	}

	// 分类
	if strings.HasPrefix(platformID, PrefixCategory) {
		uuid, err := extractUUID(platformID, PrefixCategory)
		if err != nil {
			return nil, err
		}
		return &ParseResult{UUID: uuid, IDType: IDTypeCategory, IsMapped: true}, nil
	}

	// 商品
	if strings.HasPrefix(platformID, PrefixItem) {
		uuid, err := extractUUID(platformID, PrefixItem)
		if err != nil {
			return nil, err
		}
		return &ParseResult{UUID: uuid, IDType: IDTypeItem, IsMapped: true}, nil
	}

	if strings.HasPrefix(platformID, PrefixPackage) {
		uuid, err := extractUUID(platformID, PrefixPackage)
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

// ParseModifierId 解析平台修饰符ID（向后兼容）
//
// Deprecated: 请使用 ParsePlatformID 替代
func ParseModifierId(platform, modifierId string) (ttposAttributeUuid uint64, isMapped bool, err error) {
	result, err := ParsePlatformID(platform, modifierId)
	if err != nil {
		return 0, false, err
	}

	// 接受所有修饰符类型
	switch result.IDType {
	case IDTypeFlavor, IDTypeSauce, IDTypeAttr:
		return result.UUID, result.IsMapped, nil
	default:
		return 0, false, errors.New("ID类型错误：期望修饰符类型")
	}
}
