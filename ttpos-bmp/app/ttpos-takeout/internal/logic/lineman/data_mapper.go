// Package lineman 提供 LINE MAN 平台集成服务
package lineman

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	grabfood "github.com/grab/grabfood-api-sdk-go"

	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
)

type DataMapper struct {
	ctx context.Context
}

// NewDataMapper 创建数据映射器
func NewDataMapper(ctx context.Context) *DataMapper {
	return &DataMapper{ctx: ctx}
}

// BuildMenuPayloadFromJSON 从 TTPOS 菜单 JSON 构建 Lineman 菜单数据
// 参数:
//   - ttposMenuJSON: TTPOS 菜单 JSON 字符串（Grab 格式）
//
// 返回:
//   - *lineman.MenuSyncRequest: 菜单同步请求数据
//   - error: 错误信息
func (m *DataMapper) BuildMenuPayloadFromJSON(ttposMenuJSON string) (*lineman.MenuSyncRequest, error) {
	g.Log().Infof(m.ctx, "[DataMapper] 开始从 JSON 构建菜单数据")

	// 1. 解析为 Grab 格式（TTPOS 菜单快照使用 Grab 格式存储）
	var grabMenu grabfood.GetMenuNewResponse
	if err := gjson.Unmarshal([]byte(ttposMenuJSON), &grabMenu); err != nil {
		return nil, gerror.Wrap(err, "解析 TTPOS 格式菜单 JSON 失败")
	}

	// 2. 从 Grab 格式转换为 Lineman 格式
	menuGroups, err := m.convertGrabToLinemanMenuGroups(&grabMenu)
	if err != nil {
		return nil, gerror.Wrap(err, "转换菜单格式失败")
	}

	// 3. 构建请求
	req := &lineman.MenuSyncRequest{
		MenuGroups: menuGroups,
	}

	g.Log().Infof(m.ctx, "[DataMapper] 菜单数据构建完成, 分类数=%d", len(menuGroups))

	return req, nil
}

// ============================================================================
// 多语言处理
// ============================================================================

// buildNameTranslation 构建名称翻译
// 优先使用泰语/英语翻译，无翻译时使用中文降级
func (m *DataMapper) buildNameTranslation(nameCN, nameTH, nameEN string) lineman.NameTrans {
	// 去除首尾空格
	nameCN = strings.TrimSpace(nameCN)
	nameTH = strings.TrimSpace(nameTH)
	nameEN = strings.TrimSpace(nameEN)

	// 优先使用翻译，否则使用中文降级
	if nameTH != "" && nameEN != "" {
		return lineman.NameTrans{
			Thai:    nameTH,
			English: nameEN,
		}
	}

	// 降级：使用中文填充
	return lineman.NameTrans{
		Thai:    nameCN,
		English: nameCN,
	}
}

// buildDescTranslation 构建描述翻译
// 优先使用泰语/英语翻译，无翻译时使用中文降级
func (m *DataMapper) buildDescTranslation(descCN, descTH, descEN string) lineman.DescTrans {
	// 去除首尾空格
	descCN = strings.TrimSpace(descCN)
	descTH = strings.TrimSpace(descTH)
	descEN = strings.TrimSpace(descEN)

	// 优先使用翻译，否则使用中文降级
	if descTH != "" && descEN != "" {
		return lineman.DescTrans{
			Thai:    descTH,
			English: descEN,
		}
	}

	// 降级：使用中文填充（如果中文也为空，使用空字符串）
	return lineman.DescTrans{
		Thai:    descCN,
		English: descCN,
	}
}

// ============================================================================
// Grab 格式转换为 Lineman 格式
// ============================================================================

// convertGrabToLinemanMenuGroups 将 Grab 菜单格式转换为 Lineman 格式
// 参数:
//   - grabMenu: Grab 格式的菜单数据
//
// 返回:
//   - []*lineman.MenuGroup: Lineman 格式的菜单分组列表
//   - error: 错误信息
func (m *DataMapper) convertGrabToLinemanMenuGroups(grabMenu *grabfood.GetMenuNewResponse) ([]*lineman.MenuGroup, error) {
	if grabMenu == nil || grabMenu.Categories == nil {
		return []*lineman.MenuGroup{}, nil
	}

	menuGroups := make([]*lineman.MenuGroup, 0, len(grabMenu.Categories))

	for _, grabCategory := range grabMenu.Categories {
		// 转换分类
		linemanGroup, err := m.convertGrabCategoryToLinemanGroup(grabCategory)
		if err != nil {
			g.Log().Warningf(m.ctx, "[DataMapper] 转换分类失败: categoryID=%s, error=%v",
				grabCategory.GetId(), err)
			continue
		}
		menuGroups = append(menuGroups, linemanGroup)
	}

	g.Log().Infof(m.ctx, "[DataMapper] Grab 格式转换完成: 分类数=%d", len(menuGroups))
	return menuGroups, nil
}

// convertGrabCategoryToLinemanGroup 将 Grab 分类转换为 Lineman 菜单组
func (m *DataMapper) convertGrabCategoryToLinemanGroup(grabCat grabfood.MenuCategory) (*lineman.MenuGroup, error) {

	// 构建分类名称（多语言）
	nameTrans := m.buildNameTranslationFromGrab(grabCat.GetName(), grabCat.GetNameTranslation())

	// 转换商品
	menuItems := make([]*lineman.MenuItem, 0, len(grabCat.GetItems()))
	for _, grabItem := range grabCat.GetItems() {
		linemanItem, err := m.convertGrabItemToLinemanItem(grabItem)
		if err != nil {
			g.Log().Warningf(m.ctx, "[DataMapper] 转换商品失败: itemID=%s, error=%v",
				grabItem.GetId(), err)
			continue
		}
		menuItems = append(menuItems, linemanItem)
	}

	return &lineman.MenuGroup{
		ID:             grabCat.GetId(),
		Name:           nameTrans,
		UseSellingTime: false, // Lineman 不使用 SellingTime
		MenuItems:      menuItems,
	}, nil
}

// convertGrabItemToLinemanItem 将 Grab 商品转换为 Lineman 商品
func (m *DataMapper) convertGrabItemToLinemanItem(grabItem grabfood.MenuItem) (*lineman.MenuItem, error) {

	// 构建商品名称和描述（多语言）
	nameTrans := m.buildNameTranslationFromGrab(grabItem.GetName(), grabItem.GetNameTranslation())
	descTrans := m.buildDescTranslationFromGrab(grabItem.GetDescription(), grabItem.GetDescriptionTranslation())

	// 价格转换：Grab SDK 的 GetPrice() 返回 int64（单位：分），直接转换
	price := float64(grabItem.GetPrice())

	// 状态转换
	menuStatus := m.convertGrabAvailabilityToLinemanStatus(grabItem.GetAvailableStatus())

	// 获取图片URL：从 photos[0] 获取第一张照片
	photoUrl := ""
	if photos := grabItem.GetPhotos(); len(photos) > 0 {
		photoUrl = photos[0]
	}

	// 转换属性（modifierGroups）
	properties := make([]*lineman.Property, 0)
	for _, grabModGroup := range grabItem.GetModifierGroups() {
		linemanProp, err := m.convertGrabModifierGroupToLinemanProperty(grabModGroup)
		if err != nil {
			g.Log().Warningf(m.ctx, "[DataMapper] 转换属性失败: modifierGroupID=%s, error=%v",
				grabModGroup.GetId(), err)
			continue
		}
		properties = append(properties, linemanProp)
	}

	return &lineman.MenuItem{
		ID:          grabItem.GetId(),
		Name:        nameTrans,
		Description: descTrans,
		Price:       price,
		PhotoUrl:    photoUrl,
		MenuStatus:  menuStatus,
		Properties:  properties,
		SalesChannelsAvailability: &lineman.ChannelsAvailability{
			Delivery: true,
			Pickup:   true,
		},
	}, nil
}

// buildNameTranslationFromGrab 从 Grab 名称构建 Lineman 多语言名称
func (m *DataMapper) buildNameTranslationFromGrab(name string, translations map[string]string) lineman.NameTrans {
	thai := translations["th"]
	english := translations["en"]

	// 如果泰语或英语为空，使用原始名称（可能是中文）作为降级
	if thai == "" {
		thai = name
	}
	if english == "" {
		english = name
	}

	return lineman.NameTrans{
		Thai:    strings.TrimSpace(thai),
		English: strings.TrimSpace(english),
	}
}

// buildDescTranslationFromGrab 从 Grab 描述构建 Lineman 多语言描述
func (m *DataMapper) buildDescTranslationFromGrab(desc string, translations map[string]string) lineman.DescTrans {
	thai := translations["th"]
	english := translations["en"]

	// 如果泰语或英语为空，使用原始描述作为降级
	if thai == "" {
		thai = desc
	}
	if english == "" {
		english = desc
	}

	return lineman.DescTrans{
		Thai:    strings.TrimSpace(thai),
		English: strings.TrimSpace(english),
	}
}

// convertGrabAvailabilityToLinemanStatus 转换 Grab 可用状态到 Lineman 状态
// Grab: AVAILABLE, UNAVAILABLE, HIDDEN, SOLD_OUT
// Lineman: AVAILABLE, SOLD_OUT_TODAY, SUSPENDED
func (m *DataMapper) convertGrabAvailabilityToLinemanStatus(grabStatus string) string {
	switch grabStatus {
	case "AVAILABLE":
		return "AVAILABLE"
	case "SOLD_OUT", "UNAVAILABLE":
		return "SOLD_OUT_TODAY"
	case "HIDDEN":
		return "SUSPENDED"
	default:
		g.Log().Warningf(m.ctx, "[DataMapper] 未知的 TTPOS 状态: %s, 默认使用 AVAILABLE", grabStatus)
		return "AVAILABLE"
	}
}

// convertGrabModifierGroupToLinemanProperty 将 Grab ModifierGroup 转换为 Lineman property
func (m *DataMapper) convertGrabModifierGroupToLinemanProperty(grabModGroup grabfood.ModifierGroup) (*lineman.Property, error) {
	// 构建属性名称（多语言）
	nameTrans := m.buildNameTranslationFromGrab(grabModGroup.GetName(), grabModGroup.GetNameTranslation())

	// 确定类型：单选(1) 或 多选(2)
	propType := "2" // 默认多选
	if grabModGroup.GetSelectionRangeMax() == 1 {
		propType = "1" // 单选
	}

	// 转换属性值（modifiers 下的选项）
	propValues := make([]*lineman.PropValue, 0, len(grabModGroup.GetModifiers()))
	for _, grabMod := range grabModGroup.GetModifiers() {
		valueName := m.buildNameTranslationFromGrab(grabMod.GetName(), grabMod.GetNameTranslation())
		valuePrice := float64(grabMod.GetPrice())

		propValues = append(propValues, &lineman.PropValue{
			ID:     grabMod.GetId(),
			Name:   valueName,
			Price:  valuePrice,
			Status: m.convertGrabAvailabilityToLinemanPropStatus(grabMod.GetAvailableStatus()),
		})
	}

	return &lineman.Property{
		ID:     grabModGroup.GetId(),
		Name:   nameTrans,
		Type:   propType,
		Min:    int(grabModGroup.GetSelectionRangeMin()),
		Max:    int(grabModGroup.GetSelectionRangeMax()),
		Values: propValues,
	}, nil
}

// convertGrabAvailabilityToLinemanPropStatus 转换 Grab 可用状态到 Lineman PropValue 状态
// PropValue.Status: 1=可用 2=售罄 3=暂停
func (m *DataMapper) convertGrabAvailabilityToLinemanPropStatus(grabStatus string) int {
	switch grabStatus {
	case "AVAILABLE":
		return 1
	case "SOLD_OUT", "UNAVAILABLE":
		return 2
	case "HIDDEN":
		return 3
	default:
		g.Log().Warningf(m.ctx, "[DataMapper] 未知的 Grab 状态: %s, 默认使用可用", grabStatus)
		return 1
	}
}
