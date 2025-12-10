package grab

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/google/uuid"
	grabfood "github.com/grab/grabfood-api-sdk-go"

	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/dao"
	"ttpos-bmp/app/ttpos-takeout/internal/model/do"
	grabDto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
	"ttpos-bmp/internal/pkg/queue"
)

const (
	// TopicProviderMenuUpdate 菜单更新 Topic
	TopicProviderMenuUpdate = "provider_menu_update"
)

// MenuService 菜单服务
// 内部使用，通过 sGrab 统一管理
type MenuService struct {
	verifier *SignatureVerifier
}

// HandleGetMenu 处理 Grab 获取菜单请求 (Partner Endpoint)
func (s *MenuService) HandleGetMenu(ctx context.Context, signature, timestamp string, partnerMerchantID string) (*grabDto.GetMenuResponse, error) {
	g.Log().Infof(ctx, "[Grab] Received GetMenu request: partnerMerchantID=%s", partnerMerchantID)

	// 1. 将 partnerMerchantID 转换为 shopUUID (uint64)
	// 假设 partnerMerchantID 是数字字符串格式的 shopUUID
	shopUUID := g.NewVar(partnerMerchantID).Uint64()
	if shopUUID == 0 {
		g.Log().Errorf(ctx, "[Grab] Invalid partnerMerchantID format: %s", partnerMerchantID)
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "invalid partnerMerchantID format")
	}

	// 2. 从数据库读取菜单快照
	menuJSON, err := service.ChannelMenu().GetChannelMenu(ctx, shopUUID, string(consts.ProviderGrab))
	if err != nil {
		g.Log().Errorf(ctx, "[Grab] Failed to get channel menu: shopUUID=%d, error: %v", shopUUID, err)
		return nil, fmt.Errorf("failed to get channel menu: %w", err)
	}

	// 3. 检查菜单是否存在
	if menuJSON == "" {
		g.Log().Warningf(ctx, "[Grab] Menu not found: shopUUID=%d", shopUUID)
		return nil, gerror.NewCode(gcode.CodeNotFound, "menu not found")
	}

	// 4. 解析 JSON 为 PushGrabMenuDTO (使用 SDK 类型)
	var pushDTO grabDto.PushGrabMenuDTO
	if err := json.Unmarshal([]byte(menuJSON), &pushDTO); err != nil {
		g.Log().Errorf(ctx, "[Grab] Failed to unmarshal menu JSON: error: %v", err)
		return nil, fmt.Errorf("failed to parse menu data: %w", err)
	}

	// 5. 转换为 GetMenuResponse (使用自定义类型)
	resp := &grabDto.GetMenuResponse{
		MerchantID:        pushDTO.MerchantID,
		PartnerMerchantID: pushDTO.PartnerMerchantID,
		Currency:          convertCurrency(pushDTO.Currency),
		SellingTimes:      convertSellingTimes(pushDTO.SellingTimes),
		Categories:        convertCategories(pushDTO.Categories),
	}

	g.Log().Infof(ctx, "[Grab] GetMenu success: merchantID=%s, partnerMerchantID=%s, categories=%d",
		resp.MerchantID, resp.PartnerMerchantID, len(resp.Categories))

	return resp, nil
}

// convertCurrency 转换 SDK Currency 到自定义 Currency
func convertCurrency(sdkCurrency grabfood.Currency) grabDto.Currency {
	return grabDto.Currency{
		Code:     sdkCurrency.GetCode(),
		Symbol:   sdkCurrency.GetSymbol(),
		Exponent: int(sdkCurrency.GetExponent()),
	}
}

// convertSellingTimes 转换 SDK SellingTime 数组到自定义 SellingTime 数组
func convertSellingTimes(sdkTimes []grabfood.SellingTime) []grabDto.SellingTime {
	if len(sdkTimes) == 0 {
		return nil
	}

	result := make([]grabDto.SellingTime, 0, len(sdkTimes))
	for _, st := range sdkTimes {
		result = append(result, grabDto.SellingTime{
			ID:           st.GetId(),
			Name:         st.GetName(),
			ServiceHours: convertServiceHours(st.GetServiceHours()),
		})
	}
	return result
}

// convertServiceHours 转换 SDK ServiceHours 到自定义 ServiceHours
func convertServiceHours(sdkHours grabfood.ServiceHours) grabDto.ServiceHours {
	return grabDto.ServiceHours{
		Mon: convertServiceHour(sdkHours.GetMon()),
		Tue: convertServiceHour(sdkHours.GetTue()),
		Wed: convertServiceHour(sdkHours.GetWed()),
		Thu: convertServiceHour(sdkHours.GetThu()),
		Fri: convertServiceHour(sdkHours.GetFri()),
		Sat: convertServiceHour(sdkHours.GetSat()),
		Sun: convertServiceHour(sdkHours.GetSun()),
	}
}

// convertServiceHour 转换 SDK ServiceHour 到自定义 DayHours
func convertServiceHour(sdkHour grabfood.ServiceHour) *grabDto.DayHours {
	periods := sdkHour.GetPeriods()
	if len(periods) == 0 {
		return nil
	}

	result := make([]grabDto.Period, 0, len(periods))
	for _, p := range periods {
		result = append(result, grabDto.Period{
			StartTime: p.GetStartTime(),
			EndTime:   p.GetEndTime(),
		})
	}

	return &grabDto.DayHours{
		Periods: result,
	}
}

// convertCategories 转换 SDK MenuCategory 数组到自定义 Category 数组
func convertCategories(sdkCategories []grabfood.MenuCategory) []grabDto.Category {
	if len(sdkCategories) == 0 {
		return nil
	}

	result := make([]grabDto.Category, 0, len(sdkCategories))
	for _, cat := range sdkCategories {
		result = append(result, grabDto.Category{
			ID:              cat.GetId(),
			Name:            cat.GetName(),
			NameTranslation: convertStringMap(cat.GetNameTranslation()),
			AvailableStatus: cat.GetAvailableStatus(),
			SellingTimeID:   cat.GetSellingTimeID(),
			Items:           convertMenuItems(cat.GetItems()),
		})
	}
	return result
}

// convertMenuItems 转换 SDK MenuItem 数组到自定义 MenuItem 数组
func convertMenuItems(sdkItems []grabfood.MenuItem) []grabDto.MenuItem {
	if len(sdkItems) == 0 {
		return nil
	}

	result := make([]grabDto.MenuItem, 0, len(sdkItems))
	for _, item := range sdkItems {
		result = append(result, grabDto.MenuItem{
			ID:                     item.GetId(),
			Name:                   item.GetName(),
			NameTranslation:        convertStringMap(item.GetNameTranslation()),
			Description:            item.GetDescription(),
			DescriptionTranslation: convertStringMap(item.GetDescriptionTranslation()),
			AvailableStatus:        item.GetAvailableStatus(),
			Price:                  item.GetPrice(),
			Photos:                 item.GetPhotos(),
			SpecialType:            item.GetSpecialType(),
			Taxable:                item.GetTaxable(),
			MaxStock:               int(item.GetMaxStock()),
			ModifierGroups:         convertModifierGroups(item.GetModifierGroups()),
		})
	}
	return result
}

// convertModifierGroups 转换 SDK ModifierGroup 数组到自定义 ModifierGroup 数组
func convertModifierGroups(sdkGroups []grabfood.ModifierGroup) []grabDto.ModifierGroup {
	if len(sdkGroups) == 0 {
		return nil
	}

	result := make([]grabDto.ModifierGroup, 0, len(sdkGroups))
	for _, mg := range sdkGroups {
		result = append(result, grabDto.ModifierGroup{
			ID:                mg.GetId(),
			Name:              mg.GetName(),
			NameTranslation:   convertStringMap(mg.GetNameTranslation()),
			AvailableStatus:   mg.GetAvailableStatus(),
			SelectionRangeMin: int(mg.GetSelectionRangeMin()),
			SelectionRangeMax: int(mg.GetSelectionRangeMax()),
			Modifiers:         convertMenuModifiers(mg.GetModifiers()),
		})
	}
	return result
}

// convertMenuModifiers 转换 SDK MenuModifier 数组到自定义 MenuModifier 数组
func convertMenuModifiers(sdkModifiers []grabfood.MenuModifier) []grabDto.MenuModifier {
	if len(sdkModifiers) == 0 {
		return nil
	}

	result := make([]grabDto.MenuModifier, 0, len(sdkModifiers))
	for _, mod := range sdkModifiers {
		var price int64
		if mod.HasPrice() {
			price = mod.GetPrice()
		}

		var nameTranslation map[string]string
		if mod.HasNameTranslation() {
			nameTranslation = mod.GetNameTranslation()
		}

		result = append(result, grabDto.MenuModifier{
			ID:              mod.GetId(),
			Name:            mod.GetName(),
			NameTranslation: convertStringMap(nameTranslation),
			AvailableStatus: mod.GetAvailableStatus(),
			Price:           price,
		})
	}
	return result
}

// convertStringMap 转换 SDK map[string]string 到自定义 map[string]string
func convertStringMap(sdkMap map[string]string) map[string]string {
	if len(sdkMap) == 0 {
		return nil
	}
	return sdkMap
}

// HandleMenuSyncState 处理菜单同步状态回调
// 使用 SDK grabfood.MenuSyncWebhookRequest
func (s *MenuService) HandleMenuSyncState(ctx context.Context, signature, timestamp string, body []byte) error {
	// 1. 验证签名
	if err := s.verifier.VerifySignature(signature, timestamp, body); err != nil {
		g.Log().Errorf(ctx, "Grab signature verification failed: %v", err)
		return fmt.Errorf("signature verification failed: %w", err)
	}

	// 2. 解析请求 - 使用 SDK Model
	var req grabfood.MenuSyncWebhookRequest
	if err := json.Unmarshal(body, &req); err != nil {
		g.Log().Errorf(ctx, "Failed to parse menu sync state request: %v", err)
		return fmt.Errorf("failed to parse request: %w", err)
	}

	requestID := req.GetRequestID()
	status := req.GetStatus()

	// 3. 更新菜单日志状态
	_, err := dao.MenuLog.Ctx(ctx).
		Where(dao.MenuLog.Columns().RequestId, requestID).
		Data(g.Map{
			dao.MenuLog.Columns().Status:    status,
			dao.MenuLog.Columns().UpdatedAt: gtime.Now(),
		}).Update()
	if err != nil {
		g.Log().Errorf(ctx, "Failed to update menu log status: %v", err)
		return fmt.Errorf("failed to update menu log: %w", err)
	}

	// 4. 如果失败，记录错误信息
	// SDK 的 Errors 是 []string，直接合并
	if status == grabDto.MenuSyncStatusFail && len(req.GetErrors()) > 0 {
		errMsg := strings.Join(req.GetErrors(), "; ")
		_, _ = dao.MenuLog.Ctx(ctx).
			Where(dao.MenuLog.Columns().RequestId, requestID).
			Data(g.Map{
				dao.MenuLog.Columns().ErrorMsg: errMsg,
			}).Update()
	}

	g.Log().Infof(ctx, "Menu sync state updated: requestId=%s, status=%s", requestID, status)
	return nil
}

// SyncMenu 主动同步菜单到 Grab
func (s *MenuService) SyncMenu(ctx context.Context, merchantID string, menu *grabDto.GetMenuResponse, notifier MenuNotifier) error {
	// 1. 保存菜单快照
	menuSnapshot, _ := json.Marshal(menu)
	logUUID := uuid.New().String()

	logDo := &do.MenuLog{
		Uuid:         logUUID,
		MerchantId:   merchantID,
		ProviderName: "grab",
		SyncType:     "FULL",
		Status:       grabDto.MenuSyncStatusQueued,
		MenuSnapshot: string(menuSnapshot),
		CreatedAt:    gtime.Now(),
		UpdatedAt:    gtime.Now(),
	}

	_, err := dao.MenuLog.Ctx(ctx).Data(logDo).Insert()
	if err != nil {
		return fmt.Errorf("failed to save menu log: %w", err)
	}

	// 2. 调用 Grab API 通知菜单更新
	requestID, err := notifier.NotifyMenuUpdate(ctx, merchantID)
	if err != nil {
		// 更新日志状态为失败
		_, _ = dao.MenuLog.Ctx(ctx).
			Where(dao.MenuLog.Columns().Uuid, logUUID).
			Data(g.Map{
				dao.MenuLog.Columns().Status:    grabDto.MenuSyncStatusFail,
				dao.MenuLog.Columns().ErrorMsg:  err.Error(),
				dao.MenuLog.Columns().UpdatedAt: gtime.Now(),
			}).Update()
		return fmt.Errorf("failed to notify Grab: %w", err)
	}

	// 3. 更新请求 ID
	_, err = dao.MenuLog.Ctx(ctx).
		Where(dao.MenuLog.Columns().Uuid, logUUID).
		Data(g.Map{
			dao.MenuLog.Columns().RequestId: requestID,
			dao.MenuLog.Columns().Status:    grabDto.MenuSyncStatusProcessing,
			dao.MenuLog.Columns().UpdatedAt: gtime.Now(),
		}).Update()
	if err != nil {
		g.Log().Warningf(ctx, "Failed to update menu log with requestID: %v", err)
	}

	g.Log().Infof(ctx, "Menu sync initiated: merchant=%s, requestId=%s", merchantID, requestID)
	return nil
}

// MenuNotifier 菜单通知接口 (仅用于 SyncMenu)
type MenuNotifier interface {
	// NotifyMenuUpdate 通知 Grab 菜单已更新
	NotifyMenuUpdate(ctx context.Context, merchantID string) (requestID string, err error)
}

// SaveMenuSnapshot 保存菜单快照到数据库
// 使用 shop_uuid + provider_name 作为唯一键，存在则更新，不存在则插入
func (s *MenuService) SaveMenuSnapshot(ctx context.Context, dto *grabDto.PushGrabMenuDTO) (string, error) {
	// 序列化菜单数据为 JSON
	menuData, err := json.Marshal(dto)
	if err != nil {
		return "", fmt.Errorf("failed to marshal menu snapshot: %w", err)
	}

	// 生成快照 UUID
	snapshotUUID := uuid.New().String()
	now := gtime.Now().Unix()

	// 使用 Save (InsertOrUpdate) 模式，基于 uk_shop_provider 唯一索引
	_, err = dao.ChannelMenuSnapshot.Ctx(ctx).Data(do.ChannelMenuSnapshot{
		Uuid:         snapshotUUID,
		ShopUuid:     dto.PartnerMerchantID,
		ProviderName: string(consts.ProviderGrab),
		MenuData:     string(menuData),
		CreateTime:   now,
		UpdateTime:   now,
	}).Save()
	if err != nil {
		return "", fmt.Errorf("failed to save menu snapshot to database: %w", err)
	}

	g.Log().Infof(ctx, "[Grab] Menu snapshot saved to database: shop_uuid=%s, provider=%s", dto.PartnerMerchantID, consts.ProviderGrab)
	return snapshotUUID, nil
}

// NotifyMenuUpdate 发送菜单更新通知 (RocketMQ)
func (s *MenuService) NotifyMenuUpdate(ctx context.Context, event *grabDto.ProviderMenuUpdateEvent) error {
	// 使用 queue 包发送消息
	if err := queue.PushWithContext(ctx, TopicProviderMenuUpdate, event); err != nil {
		return fmt.Errorf("failed to send menu update event: %w", err)
	}

	g.Log().Infof(ctx, "[Grab] Menu update event sent: topic=%s, merchant=%s", TopicProviderMenuUpdate, event.MerchantID)
	return nil
}
