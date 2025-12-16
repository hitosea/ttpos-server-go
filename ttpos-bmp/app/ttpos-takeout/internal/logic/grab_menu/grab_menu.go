// Package grab_menu 提供 GrabFood 菜单服务的业务逻辑
package grab_menu

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	grabfood "github.com/grab/grabfood-api-sdk-go"

	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/dao"
	"ttpos-bmp/app/ttpos-takeout/internal/model/do"
	grabDto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
	"ttpos-bmp/internal/pkg/queue"
	"ttpos-bmp/utility/uuid"
)

const (
	// TopicProviderMenuUpdate 菜单更新 Topic
	TopicProviderMenuUpdate = "takeout_provider_menu_update"
)

// sGrabMenu 菜单服务
type sGrabMenu struct{}

func init() {
	service.RegisterGrabMenu(New())
}

// New 创建菜单服务实例
func New() *sGrabMenu {
	return &sGrabMenu{}
}

// HandleGetMenu 处理 Grab 获取菜单请求 (Partner Endpoint)
// 签名验证已由中间件完成
func (s *sGrabMenu) HandleGetMenu(ctx context.Context, partnerMerchantID string) (*grabfood.GetMenuNewResponse, error) {
	g.Log().Infof(ctx, "[Grab] Received GetMenu request: partnerMerchantID=%s", partnerMerchantID)

	// 1. 将 partnerMerchantID 转换为 shopUUID (uint64)
	// 假设 partnerMerchantID 是数字字符串格式的 shopUUID
	shopUUID := g.NewVar(partnerMerchantID).Uint64()
	if shopUUID == 0 {
		g.Log().Errorf(ctx, "[Grab] Invalid partnerMerchantID format: %s", partnerMerchantID)
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "invalid partnerMerchantID format")
	}

	// 2. 从数据库读取菜单快照
	menuJSON, err := service.ChannelMenu().GetTtposMenu(ctx, shopUUID, string(consts.ProviderGrab))
	if err != nil {
		g.Log().Errorf(ctx, "[Grab] Failed to get channel menu: shopUUID=%d, error: %v", shopUUID, err)
		return nil, fmt.Errorf("failed to get channel menu: %w", err)
	}

	// 3. 检查菜单是否存在, 需要TTPOS 更新菜单后，才能获取到菜单快照
	if menuJSON == "" {
		g.Log().Warningf(ctx, "[Grab] TTPOS Menu not found: shopUUID=%d 。 必须在TTPOS中配置外卖菜单", shopUUID)
		return nil, gerror.NewCode(gcode.CodeNotFound, "menu not found")
	}

	// 4. 解析 JSON 为 PushGrabMenuDTO (使用 SDK 类型)
	var pushDTO grabDto.PushGrabMenuDTO
	if err := json.Unmarshal([]byte(menuJSON), &pushDTO); err != nil {
		g.Log().Errorf(ctx, "[Grab] Failed to unmarshal menu JSON: error: %v", err)
		return nil, fmt.Errorf("failed to parse menu data: %w", err)
	}

	// 5. 直接使用 SDK 响应结构
	resp := &grabfood.GetMenuNewResponse{
		MerchantID:        &pushDTO.MerchantID,
		PartnerMerchantID: &pushDTO.PartnerMerchantID,
		Currency:          pushDTO.Currency,
		SellingTimes:      pushDTO.SellingTimes,
		Categories:        pushDTO.Categories,
	}

	g.Log().Infof(ctx, "[Grab] GetMenu success: merchantID=%v, partnerMerchantID=%v, categories=%d",
		resp.MerchantID, resp.PartnerMerchantID, len(resp.Categories))

	return resp, nil
}

// HandleMenuSyncState 处理菜单同步状态回调
// 使用 SDK grabfood.MenuSyncWebhookRequest
func (s *sGrabMenu) HandleMenuSyncState(ctx context.Context, req *grabfood.MenuSyncWebhookRequest) error {
	requestID := req.GetRequestID()
	status := req.GetStatus()

	g.Log().Infof(ctx, "[Grab] HandleMenuSyncState: requestID=%s, merchantID=%s, partnerMerchantID=%s, status=%s",
		requestID, req.GetMerchantID(), req.GetPartnerMerchantID(), status)

	// 更新 channel_menu_snapshot 的状态
	shopUUID := g.NewVar(req.PartnerMerchantID).Uint64()
	if shopUUID > 0 {
		_, err := dao.ChannelMenuSnapshot.Ctx(ctx).
			Where(dao.ChannelMenuSnapshot.Columns().ShopUuid, shopUUID).
			Where(dao.ChannelMenuSnapshot.Columns().ProviderName, string(consts.ProviderGrab)).
			Data(g.Map{
				dao.ChannelMenuSnapshot.Columns().SyncState: status,
				dao.ChannelMenuSnapshot.Columns().UpdatedAt: gtime.Now(),
			}).Update()
		if err != nil {
			g.Log().Errorf(ctx, "Failed to update channel_menu_snapshot status: shopUUID=%d, status=%s, error: %v", shopUUID, status, err)
			// 不中断流程，继续处理 MenuLog 插入
		} else {
			g.Log().Infof(ctx, "Channel menu snapshot status updated: shopUUID=%d, status=%s", shopUUID, status)
		}
	} else {
		g.Log().Warningf(ctx, "Invalid merchantID format, cannot convert to ShopUuid: merchantID=%v", req.PartnerMerchantID)
	}

	// 插入新的菜单日志记录（每次状态回调都插入新记录）
	logUUID := uuid.MustGetID()
	errMsg := ""
	if status == grabDto.MenuSyncStatusFail && len(req.GetErrors()) > 0 {
		errMsg = strings.Join(req.GetErrors(), "; ")
	}

	logDo := &do.MenuLog{
		Uuid:         logUUID,
		MerchantId:   req.PartnerMerchantID,
		ProviderName: string(consts.ProviderGrab),
		Status:       status,
		ErrorMsg:     errMsg,
		CreatedAt:    gtime.Now(),
		UpdatedAt:    gtime.Now(),
	}
	_, err := dao.MenuLog.Ctx(ctx).Data(logDo).Insert()
	if err != nil {
		g.Log().Errorf(ctx, "Failed to insert menu log: %v", err)
	}
	g.Log().Infof(ctx, "Menu sync state processed: requestId=%s, status=%s, logUUID=%v", requestID, status, logUUID)
	return nil
}

// SyncMenu 主动同步菜单到 Grab
func (s *sGrabMenu) SyncMenu(ctx context.Context, merchantID string, menu *grabfood.GetMenuNewResponse, notifier grabDto.MenuNotifier) error {
	// 1. 保存菜单快照
	menuSnapshot, _ := json.Marshal(menu)
	logUUID := uuid.MustGetID()

	logDo := &do.MenuLog{
		Uuid:         logUUID,
		MerchantId:   merchantID,
		ProviderName: string(consts.ProviderGrab),
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

	// 2. 更新菜单快照表
	_, err = dao.ChannelMenuSnapshot.Ctx(ctx).
		Where(dao.ChannelMenuSnapshot.Columns().ShopUuid, g.NewVar(merchantID).Uint64()).
		Where(dao.ChannelMenuSnapshot.Columns().ProviderName, string(consts.ProviderGrab)).
		Data(g.Map{
			dao.ChannelMenuSnapshot.Columns().TtposMenuData: string(menuSnapshot),
			dao.ChannelMenuSnapshot.Columns().UpdatedAt:     gtime.Now(),
		}).Update()
	if err != nil {
		return fmt.Errorf("failed to update menu snapshot: %w", err)
	}
	// 3. 调用 Grab API 通知菜单更新
	requestID, err := notifier.NotifyMenuUpdate(ctx, *menu.MerchantID)
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

	g.Log().Infof(ctx, "Menu sync initiated: merchant=%s, requestId=%s", merchantID, requestID)
	return nil
}

// SaveMenuSnapshot 保存菜单快照到数据库
// 使用 shop_uuid + provider_name 作为唯一键，存在则更新，不存在则插入
func (s *sGrabMenu) SaveMenuSnapshot(ctx context.Context, dto *grabDto.PushGrabMenuDTO) (uint64, error) {
	// 序列化菜单数据为 JSON
	menuData, err := json.Marshal(dto)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal menu snapshot: %w", err)
	}

	// 生成快照 UUID
	snapshotUUID := uuid.MustGetID()

	// 使用 Save (InsertOrUpdate) 模式，基于 uk_shop_provider 唯一索引
	_, err = dao.ChannelMenuSnapshot.Ctx(ctx).Data(do.ChannelMenuSnapshot{
		Uuid:         snapshotUUID,
		ShopUuid:     dto.PartnerMerchantID,
		ProviderName: string(consts.ProviderGrab),
		MenuData:     string(menuData),
	}).Save()
	if err != nil {
		return 0, fmt.Errorf("failed to save menu snapshot to database: %w", err)
	}

	g.Log().Infof(ctx, "[Grab] Menu snapshot saved to database: shop_uuid=%s, provider=%s", dto.PartnerMerchantID, consts.ProviderGrab)
	return snapshotUUID, nil
}

// NotifyMenuUpdate 发送菜单更新通知 (RocketMQ)
func (s *sGrabMenu) NotifyMenuUpdate(ctx context.Context, event *grabDto.ProviderMenuUpdateEvent) error {
	// 使用 queue 包发送消息
	if err := queue.PushWithContext(ctx, TopicProviderMenuUpdate, event); err != nil {
		return fmt.Errorf("failed to send menu update event: %w", err)
	}

	g.Log().Infof(ctx, "[Grab] Menu update event sent: topic=%s, merchant=%s", TopicProviderMenuUpdate, event.MerchantID)
	return nil
}

// ============================================================================
// 菜单项/修饰符单独更新 (Update Menu Record API)
// ============================================================================

// UpdateMenuItem 更新单个菜单项 (商品)
// 调用 GrabFood API PUT /partner/v1/merchants/menu/record 更新商品信息
// 支持更新：价格、可用状态、库存、高级定价配置、购买能力配置
func (s *sGrabMenu) UpdateMenuItem(ctx context.Context, req *grabDto.UpdateMenuItemReq) (*grabDto.UpdateMenuResult, error) {
	g.Log().Infof(ctx, "[Grab] UpdateMenuItem: merchantID=%s, itemID=%s", req.MerchantID, req.ItemID)

	// 1. 参数验证
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		g.Log().Errorf(ctx, "[Grab] UpdateMenuItem validation failed: %v", err)
		return &grabDto.UpdateMenuResult{
			Success:      false,
			MerchantID:   req.MerchantID,
			RecordID:     req.ItemID,
			RecordType:   grabDto.MenuItemUpdateFieldItem,
			ErrorCode:    "VALIDATION_ERROR",
			ErrorMessage: err.Error(),
		}, gerror.NewCode(gcode.CodeValidationFailed, err.Error())
	}

	// 2. 构建 SDK 请求
	updateItem := req.ToSDKUpdateMenuItem()
	updateReq := grabfood.UpdateMenuItemAsUpdateMenuRequest(updateItem)

	// 3. 调用 Grab API
	err := service.Grab().UpdateMenuRecord(ctx, req.MerchantID, updateReq)
	if err != nil {
		// 4. 记录失败日志
		s.logMenuRecordUpdate(ctx, req.MerchantID, req.ItemID, grabDto.MenuItemUpdateFieldItem, false, err.Error())

		g.Log().Errorf(ctx, "[Grab] UpdateMenuItem API failed: merchantID=%s, itemID=%s, error=%v",
			req.MerchantID, req.ItemID, err)
		return &grabDto.UpdateMenuResult{
			Success:      false,
			MerchantID:   req.MerchantID,
			RecordID:     req.ItemID,
			RecordType:   grabDto.MenuItemUpdateFieldItem,
			ErrorCode:    "API_ERROR",
			ErrorMessage: err.Error(),
		}, gerror.Wrap(err, "调用 Grab UpdateMenuItem API 失败")
	}

	// 5. 记录成功日志
	s.logMenuRecordUpdate(ctx, req.MerchantID, req.ItemID, grabDto.MenuItemUpdateFieldItem, true, "")

	g.Log().Infof(ctx, "[Grab] UpdateMenuItem success: merchantID=%s, itemID=%s", req.MerchantID, req.ItemID)
	return &grabDto.UpdateMenuResult{
		Success:    true,
		MerchantID: req.MerchantID,
		RecordID:   req.ItemID,
		RecordType: grabDto.MenuItemUpdateFieldItem,
	}, nil
}

// UpdateMenuModifier 更新单个修饰符
// 调用 GrabFood API PUT /partner/v1/merchants/menu/record 更新修饰符信息
// 支持更新：价格、可用状态、是否免费、高级定价配置
func (s *sGrabMenu) UpdateMenuModifier(ctx context.Context, req *grabDto.UpdateMenuModifierReq) (*grabDto.UpdateMenuResult, error) {
	g.Log().Infof(ctx, "[Grab] UpdateMenuModifier: merchantID=%s, modifierID=%s, modifierName=%s",
		req.MerchantID, req.ModifierID, req.ModifierName)

	// 1. 参数验证
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		g.Log().Errorf(ctx, "[Grab] UpdateMenuModifier validation failed: %v", err)
		return &grabDto.UpdateMenuResult{
			Success:      false,
			MerchantID:   req.MerchantID,
			RecordID:     req.ModifierID,
			RecordType:   grabDto.MenuItemUpdateFieldModifier,
			ErrorCode:    "VALIDATION_ERROR",
			ErrorMessage: err.Error(),
		}, gerror.NewCode(gcode.CodeValidationFailed, err.Error())
	}

	// 2. 构建 SDK 请求
	updateModifier := req.ToSDKUpdateMenuModifier()
	updateReq := grabfood.UpdateMenuModifierAsUpdateMenuRequest(updateModifier)

	// 3. 调用 Grab API
	err := service.Grab().UpdateMenuRecord(ctx, req.MerchantID, updateReq)
	if err != nil {
		// 4. 记录失败日志
		s.logMenuRecordUpdate(ctx, req.MerchantID, req.ModifierID, grabDto.MenuItemUpdateFieldModifier, false, err.Error())

		g.Log().Errorf(ctx, "[Grab] UpdateMenuModifier API failed: merchantID=%s, modifierID=%s, error=%v",
			req.MerchantID, req.ModifierID, err)
		return &grabDto.UpdateMenuResult{
			Success:      false,
			MerchantID:   req.MerchantID,
			RecordID:     req.ModifierID,
			RecordType:   grabDto.MenuItemUpdateFieldModifier,
			ErrorCode:    "API_ERROR",
			ErrorMessage: err.Error(),
		}, gerror.Wrap(err, "调用 Grab UpdateMenuModifier API 失败")
	}

	// 5. 记录成功日志
	s.logMenuRecordUpdate(ctx, req.MerchantID, req.ModifierID, grabDto.MenuItemUpdateFieldModifier, true, "")

	g.Log().Infof(ctx, "[Grab] UpdateMenuModifier success: merchantID=%s, modifierID=%s", req.MerchantID, req.ModifierID)
	return &grabDto.UpdateMenuResult{
		Success:    true,
		MerchantID: req.MerchantID,
		RecordID:   req.ModifierID,
		RecordType: grabDto.MenuItemUpdateFieldModifier,
	}, nil
}

// logMenuRecordUpdate 记录菜单记录更新日志
// 内部方法，记录到 menu_log 表
func (s *sGrabMenu) logMenuRecordUpdate(ctx context.Context, merchantID, recordID, recordType string, success bool, errMsg string) {
	logUUID := uuid.MustGetID()
	status := grabDto.MenuSyncStatusSuccess
	if !success {
		status = grabDto.MenuSyncStatusFail
	}

	logDo := &do.MenuLog{
		Uuid:         logUUID,
		MerchantId:   merchantID,
		ProviderName: string(consts.ProviderGrab),
		SyncType:     "UPDATE_" + recordType, // UPDATE_ITEM 或 UPDATE_MODIFIER
		Status:       status,
		ErrorMsg:     errMsg,
	}

	_, err := dao.MenuLog.Ctx(ctx).Data(logDo).Insert()
	if err != nil {
		g.Log().Errorf(ctx, "[Grab] Failed to insert menu update log: merchantID=%s, recordID=%s, error=%v",
			merchantID, recordID, err)
	} else {
		g.Log().Debugf(ctx, "[Grab] Menu update log inserted: logUUID=%d, merchantID=%s, recordID=%s, recordType=%s, success=%v",
			logUUID, merchantID, recordID, recordType, success)
	}
}
