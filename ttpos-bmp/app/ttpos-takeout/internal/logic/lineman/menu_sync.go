// Package lineman 提供 LINE MAN 平台集成服务
package lineman

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	grabfood "github.com/grab/grabfood-api-sdk-go"

	linemanClient "ttpos-bmp/app/ttpos-takeout/internal/client/lineman"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

// ============================================================================
// 菜单同步主流程
// ============================================================================

// SyncMenu 同步菜单到 Lineman
// 参数:
//   - ctx: 上下文
//   - shopUUID: 门店UUID
//
// 返回:
//   - error: 错误信息
func (s *sLineman) SyncMenu(ctx context.Context, shopUUID uint64) error {
	g.Log().Infof(ctx, "[Lineman] 开始同步菜单: shopUUID=%d", shopUUID)

	// 1. 获取门店配置（复用 shop_provider_cfg）
	// cfg, err := service.ShopProviderCfg().GetShopProviderCfg(ctx, shopUUID, string(consts.ProviderLineman))
	// if err != nil {
	// 	return gerror.Wrap(err, "[Lineman] 获取门店配置失败")
	// }
	// if cfg == nil {
	// 	return gerror.Newf("[Lineman] 门店未配置 Lineman: shopUUID=%d", shopUUID)
	// }

	// 对于 Lineman，storeId 直接使用 shopUUID
	storeId := fmt.Sprintf("%d", shopUUID)

	// 3. 获取 TTPOS 菜单数据（复用 channel_menu）
	ttposMenuJSON, err := service.ChannelMenu().GetTtposMenu(ctx, shopUUID, string(consts.ProviderLineman))
	if err != nil {
		return gerror.Wrap(err, "[Lineman] 获取 TTPOS 菜单失败")
	}
	if ttposMenuJSON == "" {
		return gerror.Newf("[Lineman] TTPOS 菜单为空: shopUUID=%d", shopUUID)
	}

	// 4. 转换菜单数据为 Lineman 格式
	menuPayload, err := s.BuildMenuPayload(ctx, ttposMenuJSON)
	if err != nil {
		return gerror.Wrap(err, "[Lineman] 构建菜单数据失败")
	}

	// 6. 调用 Client 发送请求（Client 内部会获取 auth header 和 partnerId）
	client := linemanClient.NewMenuSyncClient()
	resp, err := client.SyncMenu(ctx, storeId, menuPayload)
	if err != nil {
		// 记录失败日志（复用 channel_menu）
		menuJSON, _ := gjson.EncodeString(menuPayload)
		service.ChannelMenu().LogMenuSync(ctx, storeId, string(consts.ProviderLineman), string(consts.MenuSyncTypeFull), "", false, menuJSON, err.Error())

		return gerror.Wrap(err, "[Lineman] 调用 Lineman API 失败")
	}

	// 7. 记录成功日志（复用 channel_menu）
	menuJSON, _ := gjson.EncodeString(menuPayload)
	if err := service.ChannelMenu().LogMenuSync(ctx, storeId, string(consts.ProviderLineman), string(consts.MenuSyncTypeFull), resp.MenuSyncRequestId, true, menuJSON, ""); err != nil {
		g.Log().Warningf(ctx, "[Lineman] 记录成功日志失败: %v", err)
	}

	g.Log().Infof(ctx, "[Lineman] 菜单同步成功: requestId=%s", resp.MenuSyncRequestId)

	return nil
}

// ============================================================================
// 菜单数据构建
// ============================================================================

// BuildMenuPayload 从 TTPOS 菜单 JSON 构建 Lineman 菜单数据
// 参数:
//   - ctx: 上下文
//   - ttposMenuJSON: TTPOS 菜单 JSON 字符串（Grab 格式）
//
// 返回:
//   - *lineman.MenuSyncRequest: 菜单同步请求数据
//   - error: 错误信息
func (s *sLineman) BuildMenuPayload(ctx context.Context, ttposMenuJSON string) (*lineman.MenuSyncRequest, error) {
	g.Log().Infof(ctx, "[Lineman] 开始从 JSON 构建菜单数据")

	// 1. 解析为 Grab 格式（TTPOS 菜单快照使用 Grab 格式存储）
	var grabMenu grabfood.GetMenuNewResponse
	if err := gjson.Unmarshal([]byte(ttposMenuJSON), &grabMenu); err != nil {
		return nil, gerror.Wrap(err, "解析 TTPOS 格式菜单 JSON 失败")
	}

	// 2. 从 Grab 格式转换为 Lineman 格式
	menuGroups, err := s.convertGrabToLinemanMenuGroups(ctx, &grabMenu)
	if err != nil {
		return nil, gerror.Wrap(err, "转换菜单格式失败")
	}

	// 3. 构建请求
	req := &lineman.MenuSyncRequest{
		MenuGroups: menuGroups,
	}

	g.Log().Infof(ctx, "[Lineman] 菜单数据构建完成, 分类数=%d", len(menuGroups))

	return req, nil
}

// ============================================================================
// 多语言处理
// ============================================================================

// buildNameTranslation 构建名称翻译
// 优先使用泰语/英语翻译，无翻译时使用中文降级
func (s *sLineman) buildNameTranslation(ctx context.Context, nameCN, nameTH, nameEN string) lineman.NameTrans {
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
func (s *sLineman) buildDescTranslation(ctx context.Context, descCN, descTH, descEN string) lineman.DescTrans {
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

// buildNameTranslationFromGrab 从 Grab 名称构建 Lineman 多语言名称
func (s *sLineman) buildNameTranslationFromGrab(ctx context.Context, name string, translations map[string]string) lineman.NameTrans {
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
func (s *sLineman) buildDescTranslationFromGrab(ctx context.Context, desc string, translations map[string]string) lineman.DescTrans {
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

// ============================================================================
// Grab 格式转换为 Lineman 格式
// ============================================================================

// convertGrabToLinemanMenuGroups 将 Grab 菜单格式转换为 Lineman 格式
// 参数:
//   - ctx: 上下文
//   - grabMenu: Grab 格式的菜单数据
//
// 返回:
//   - []*lineman.MenuGroup: Lineman 格式的菜单分组列表
//   - error: 错误信息
func (s *sLineman) convertGrabToLinemanMenuGroups(ctx context.Context, grabMenu *grabfood.GetMenuNewResponse) ([]*lineman.MenuGroup, error) {
	if grabMenu == nil || grabMenu.Categories == nil {
		return []*lineman.MenuGroup{}, nil
	}

	menuGroups := make([]*lineman.MenuGroup, 0, len(grabMenu.Categories))

	for _, grabCategory := range grabMenu.Categories {
		// 转换分类
		linemanGroup, err := s.convertGrabCategoryToLinemanGroup(ctx, grabCategory)
		if err != nil {
			g.Log().Warningf(ctx, "[Lineman] 转换分类失败: categoryID=%s, error=%v",
				grabCategory.GetId(), err)
			continue
		}
		menuGroups = append(menuGroups, linemanGroup)
	}

	g.Log().Infof(ctx, "[Lineman] Grab 格式转换完成: 分类数=%d", len(menuGroups))
	return menuGroups, nil
}

// convertGrabCategoryToLinemanGroup 将 Grab 分类转换为 Lineman 菜单组
func (s *sLineman) convertGrabCategoryToLinemanGroup(ctx context.Context, grabCat grabfood.MenuCategory) (*lineman.MenuGroup, error) {

	// 构建分类名称（多语言）
	nameTrans := s.buildNameTranslationFromGrab(ctx, grabCat.GetName(), grabCat.GetNameTranslation())

	// 转换商品
	menuItems := make([]*lineman.MenuItem, 0, len(grabCat.GetItems()))
	for _, grabItem := range grabCat.GetItems() {
		linemanItem, err := s.convertGrabItemToLinemanItem(ctx, grabItem)
		if err != nil {
			g.Log().Warningf(ctx, "[Lineman] 转换商品失败: itemID=%s, error=%v",
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
func (s *sLineman) convertGrabItemToLinemanItem(ctx context.Context, grabItem grabfood.MenuItem) (*lineman.MenuItem, error) {

	// 构建商品名称和描述（多语言）
	nameTrans := s.buildNameTranslationFromGrab(ctx, grabItem.GetName(), grabItem.GetNameTranslation())
	descTrans := s.buildDescTranslationFromGrab(ctx, grabItem.GetDescription(), grabItem.GetDescriptionTranslation())

	// 价格转换：Grab SDK 的 GetPrice() 返回 int64（单位：分），转换为元
	price := float64(grabItem.GetPrice()) / 100

	// 状态转换
	menuStatus := s.convertGrabAvailabilityToLinemanStatus(ctx, grabItem.GetAvailableStatus())

	// 获取图片URL：从 photos[0] 获取第一张照片
	photoUrl := ""
	if photos := grabItem.GetPhotos(); len(photos) > 0 {
		photoUrl = photos[0]
	}

	// 转换属性（modifierGroups）
	properties := make([]*lineman.Property, 0)
	for _, grabModGroup := range grabItem.GetModifierGroups() {
		linemanProp, err := s.convertGrabModifierGroupToLinemanProperty(ctx, grabModGroup)
		if err != nil {
			g.Log().Warningf(ctx, "[Lineman] 转换属性失败: modifierGroupID=%s, error=%v",
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

// convertGrabAvailabilityToLinemanStatus 转换 Grab 可用状态到 Lineman 状态
// Grab: AVAILABLE, UNAVAILABLE, HIDDEN, SOLD_OUT
// Lineman: AVAILABLE, SOLD_OUT_TODAY, SUSPENDED
func (s *sLineman) convertGrabAvailabilityToLinemanStatus(ctx context.Context, grabStatus string) string {
	switch grabStatus {
	case "AVAILABLE":
		return "AVAILABLE"
	case "SOLD_OUT", "UNAVAILABLE":
		return "SOLD_OUT_TODAY"
	case "HIDDEN":
		return "SUSPENDED"
	default:
		g.Log().Warningf(ctx, "[Lineman] 未知的 TTPOS 状态: %s, 默认使用 AVAILABLE", grabStatus)
		return "AVAILABLE"
	}
}

// convertGrabModifierGroupToLinemanProperty 将 Grab ModifierGroup 转换为 Lineman property
func (s *sLineman) convertGrabModifierGroupToLinemanProperty(ctx context.Context, grabModGroup grabfood.ModifierGroup) (*lineman.Property, error) {
	// 构建属性名称（多语言）
	nameTrans := s.buildNameTranslationFromGrab(ctx, grabModGroup.GetName(), grabModGroup.GetNameTranslation())

	// 转换属性值（modifiers 下的选项）
	propValues := make([]*lineman.PropValue, 0, len(grabModGroup.GetModifiers()))
	for _, grabMod := range grabModGroup.GetModifiers() {
		valueName := s.buildNameTranslationFromGrab(ctx, grabMod.GetName(), grabMod.GetNameTranslation())

		// 价格转换：Grab SDK 的 GetPrice() 返回 int64（单位：分），转换为元
		valuePrice := float64(grabMod.GetPrice()) / 100

		propValues = append(propValues, &lineman.PropValue{
			ID:     grabMod.GetId(),
			Name:   valueName,
			Price:  valuePrice,
			Status: s.convertGrabAvailabilityToLinemanPropStatus(ctx, grabMod.GetAvailableStatus()),
		})
	}

	// 如果没有选项，返回错误，跳过该属性组
	if len(propValues) == 0 {
		return nil, gerror.Newf("属性组 %s 没有有效的选项", grabModGroup.GetId())
	}

	// 获取原始的 min/max 值
	originalMin := int(grabModGroup.GetSelectionRangeMin())
	originalMax := int(grabModGroup.GetSelectionRangeMax())

	// 确定类型：单选(1) 或 多选(2)
	propType := "2" // 默认多选
	if originalMax == 1 && originalMin == 1 {
		propType = "1" // 单选
	}

	return &lineman.Property{
		ID:     grabModGroup.GetId(),
		Name:   nameTrans,
		Type:   propType,
		Min:    originalMin,
		Max:    originalMax,
		Values: propValues,
	}, nil
}

// convertGrabAvailabilityToLinemanPropStatus 转换 Grab 可用状态到 Lineman PropValue 状态
// PropValue.Status: 1=可用 2=售罄 3=暂停
func (s *sLineman) convertGrabAvailabilityToLinemanPropStatus(ctx context.Context, grabStatus string) int {
	switch grabStatus {
	case "AVAILABLE":
		return 1
	case "SOLD_OUT", "UNAVAILABLE":
		return 2
	case "HIDDEN":
		return 3
	default:
		g.Log().Warningf(ctx, "[Lineman] 未知的 Grab 状态: %s, 默认使用可用", grabStatus)
		return 1
	}
}
