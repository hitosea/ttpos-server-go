// Package grab_menu 提供 GrabFood 菜单服务的业务逻辑
package grab_menu

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
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
	"ttpos-bmp/app/ttpos-takeout/utility"
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
	g.Log().Infof(ctx, "[Grab] 收到获取菜单请求: partnerMerchantID=%s", partnerMerchantID)

	// 1. 将 partnerMerchantID 转换为 shopUUID (uint64)
	// 假设 partnerMerchantID 是数字字符串格式的 shopUUID
	shopUUID := g.NewVar(partnerMerchantID).Uint64()
	if shopUUID == 0 {
		g.Log().Errorf(ctx, "[Grab] partnerMerchantID 格式无效: %s", partnerMerchantID)
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "partnerMerchantID 格式无效")
	}

	// 2. 优先从本地快照读取菜单
	menuJSON, err := service.ChannelMenu().GetTtposMenu(ctx, shopUUID, string(consts.ProviderGrab))
	if err != nil {
		g.Log().Errorf(ctx, "[Grab] 获取渠道菜单失败: shopUUID=%d, error: %v", shopUUID, err)
		return nil, gerror.Wrap(err, "获取渠道菜单失败")
	}

	// 3. 如果本地快照为空，回退调用 TTPOS 导出接口
	if menuJSON == "" {
		g.Log().Infof(ctx, "[Grab] 本地菜单快照不存在，回退调用 TTPOS 导出接口: shopUUID=%d", shopUUID)
		resp, err := s.fetchMenuFromTTpos(ctx, shopUUID)
		if err != nil {
			g.Log().Errorf(ctx, "[Grab] 从 TTPOS 获取菜单失败: shopUUID=%d, error=%v", shopUUID, err)
			return nil, gerror.NewCode(gcode.CodeNotFound, "菜单不存在")
		}
		g.Log().Infof(ctx, "[Grab] 获取菜单成功（来自 TTPOS）: merchantID=%v, partnerMerchantID=%v, categories=%d",
			resp.MerchantID, resp.PartnerMerchantID, len(resp.Categories))
		return resp, nil
	}

	// 4. 解析本地快照 JSON 为 PushGrabMenuDTO (使用 SDK 类型)
	var pushDTO grabDto.PushGrabMenuDTO
	if err := json.Unmarshal([]byte(menuJSON), &pushDTO); err != nil {
		g.Log().Errorf(ctx, "[Grab] 解析菜单 JSON 失败: error: %v", err)
		return nil, gerror.Wrap(err, "解析菜单数据失败")
	}

	// 5. 从 ShopProviderCfg 查询 MerchantID 和 PartnerMerchantID
	cfg, err := service.ShopProviderCfg().GetShopProviderCfg(ctx, shopUUID, string(consts.ProviderGrab))
	if err != nil {
		g.Log().Errorf(ctx, "[Grab] 获取门店第三方配置失败: shopUUID=%d, error: %v", shopUUID, err)
		return nil, gerror.Wrap(err, "获取门店第三方配置失败")
	}
	if cfg == nil {
		g.Log().Errorf(ctx, "[Grab] 门店第三方配置不存在: shopUUID=%d, provider=%s", shopUUID, consts.ProviderGrab)
		return nil, gerror.NewCode(gcode.CodeNotFound, "门店第三方配置不存在")
	}

	// 6. 构建响应结构，使用 ShopProviderCfg 中的 MerchantID，PartnerMerchantID 使用 shopUUID
	merchantID := cfg.ProviderMerchantId
	partnerMerchantIDStr := fmt.Sprintf("%d", shopUUID)
	resp := &grabfood.GetMenuNewResponse{
		MerchantID:        &merchantID,
		PartnerMerchantID: &partnerMerchantIDStr,
		Currency:          pushDTO.Currency,
		SellingTimes:      pushDTO.SellingTimes,
		Categories:        pushDTO.Categories,
	}

	g.Log().Infof(ctx, "[Grab] 获取菜单成功（来自本地）: merchantID=%v, partnerMerchantID=%v, categories=%d",
		resp.MerchantID, resp.PartnerMerchantID, len(resp.Categories))

	return resp, nil
}

// fetchMenuFromTTpos 从 TTPOS 主模块获取菜单数据
// 当本地菜单快照为空时，回退调用此方法
func (s *sGrabMenu) fetchMenuFromTTpos(ctx context.Context, shopUUID uint64) (*grabfood.GetMenuNewResponse, error) {
	// 1. 获取带认证的 Client
	client, err := utility.GetTtposClientWithAuth(ctx, fmt.Sprintf("%d", shopUUID))
	if err != nil {
		return nil, gerror.Wrap(err, "创建 TTPOS 客户端失败")
	}

	// 2. 构建请求体
	reqBody := g.Map{
		"platform":     string(consts.ProviderGrab),
		"company_uuid": shopUUID,
	}

	// 3. 发起请求
	resp := client.ContentJson().PostVar(ctx, "/api/v1/takeout/menu/export", reqBody)
	if resp == nil || resp.IsEmpty() {
		return nil, gerror.New("TTPOS 导出接口返回空响应")
	}

	// 4. 解析响应
	resultJson, err := gjson.DecodeToJson(resp)
	if err != nil {
		return nil, gerror.Wrap(err, "解析 TTPOS 导出接口响应失败")
	}

	// 5. 检查业务状态码（兼容 code=200 和 code=1 两种成功状态）
	code := resultJson.Get("code").Int()
	if code != 0 {
		message := resultJson.Get("message").String()
		return nil, gerror.Newf("TTPOS 导出接口错误: code=%d, message=%s", code, message)
	}

	// 6. 解析菜单数据
	menuData := &grabfood.GetMenuNewResponse{}
	if err := resultJson.Get("data.menuData").Struct(&menuData); err != nil {
		return nil, gerror.Wrap(err, "解析菜单数据失败")
	}

	g.Log().Infof(ctx, "[Grab] 从 TTPOS 获取菜单成功: shopUUID=%d", shopUUID)
	return menuData, nil
}

// HandleMenuSyncState 处理菜单同步状态回调
// 使用 SDK grabfood.MenuSyncWebhookRequest
func (s *sGrabMenu) HandleMenuSyncState(ctx context.Context, req *grabfood.MenuSyncWebhookRequest) error {
	requestID := req.GetRequestID()
	status := req.GetStatus()

	g.Log().Infof(ctx, "[Grab] 处理菜单同步状态: requestID=%s, merchantID=%s, partnerMerchantID=%s, status=%s",
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
			g.Log().Errorf(ctx, "[Grab] 更新渠道菜单快照状态失败: shopUUID=%d, status=%s, error: %v", shopUUID, status, err)
			// 不中断流程，继续处理 MenuLog 插入
		} else {
			g.Log().Infof(ctx, "[Grab] 渠道菜单快照状态已更新: shopUUID=%d, status=%s", shopUUID, status)
		}
	} else {
		g.Log().Warningf(ctx, "[Grab] merchantID 格式无效，无法转换为 ShopUuid: merchantID=%v", req.PartnerMerchantID)
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
		g.Log().Errorf(ctx, "[Grab] 插入菜单日志失败: %v", err)
	}
	g.Log().Infof(ctx, "[Grab] 菜单同步状态已处理: requestId=%s, status=%s, logUUID=%v", requestID, status, logUUID)
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
		return fmt.Errorf("保存菜单日志失败: %w", err)
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
		return fmt.Errorf("更新菜单快照失败: %w", err)
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
		return fmt.Errorf("通知 Grab 失败: %w", err)
	}

	g.Log().Infof(ctx, "[Grab] 菜单同步已启动: merchant=%s, requestId=%s", merchantID, requestID)
	return nil
}

// SaveMenuSnapshot 保存菜单快照到数据库
// 使用 shop_uuid + provider_name 作为唯一键，存在则更新，不存在则插入
func (s *sGrabMenu) SaveMenuSnapshot(ctx context.Context, dto *grabDto.PushGrabMenuDTO) (uint64, error) {
	// 序列化菜单数据为 JSON
	menuData, err := json.Marshal(dto)
	if err != nil {
		return 0, fmt.Errorf("序列化菜单快照失败: %w", err)
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
		return 0, fmt.Errorf("保存菜单快照到数据库失败: %w", err)
	}

	g.Log().Infof(ctx, "[Grab] 菜单快照已保存到数据库: shop_uuid=%s, provider=%s", dto.PartnerMerchantID, consts.ProviderGrab)
	return snapshotUUID, nil
}

// NotifyMenuUpdate 发送菜单更新通知 (RocketMQ)
func (s *sGrabMenu) NotifyMenuUpdate(ctx context.Context, event *grabDto.ProviderMenuUpdateEvent) error {
	// 使用 queue 包发送消息
	if err := queue.PushWithContext(ctx, TopicProviderMenuUpdate, event); err != nil {
		return fmt.Errorf("发送菜单更新事件失败: %w", err)
	}

	g.Log().Infof(ctx, "[Grab] 菜单更新事件已发送: topic=%s, merchant=%s", TopicProviderMenuUpdate, event.MerchantID)
	return nil
}

// ============================================================================
// 菜单项/修饰符单独更新 (Update Menu Record API)
// ============================================================================

// UpdateMenuItem 更新单个菜单项 (商品)
// 调用 GrabFood API PUT /partner/v1/merchants/menu/record 更新商品信息
// 支持更新：价格、可用状态、库存、高级定价配置、购买能力配置
func (s *sGrabMenu) UpdateMenuItem(ctx context.Context, req *grabDto.UpdateMenuItemReq) error {
	g.Log().Infof(ctx, "[Grab] 更新菜单项: merchantID=%s, itemID=%s", req.MerchantID, req.ItemID)

	// 1. 参数验证
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		g.Log().Errorf(ctx, "[Grab] 更新菜单项参数验证失败: %v", err)
		return gerror.NewCode(gcode.CodeValidationFailed, err.Error())
	}

	// 2. 构建 SDK 请求
	updateItem := req.ToSDKUpdateMenuItem()
	updateReq := grabfood.UpdateMenuItemAsUpdateMenuRequest(updateItem)

	// 3. 调用 Grab API
	err := service.Grab().UpdateMenuRecord(ctx, req.MerchantID, updateReq)
	if err != nil {
		// 4. 记录失败日志
		s.logMenuRecordUpdate(ctx, req.MerchantID, req.ItemID, grabDto.MenuItemUpdateFieldItem, false, err.Error())

		g.Log().Errorf(ctx, "[Grab] 更新菜单项 API 调用失败: merchantID=%s, itemID=%s, error=%v",
			req.MerchantID, req.ItemID, err)
		return gerror.Wrap(err, "调用 Grab UpdateMenuItem API 失败")
	}

	// 5. 记录成功日志
	s.logMenuRecordUpdate(ctx, req.MerchantID, req.ItemID, grabDto.MenuItemUpdateFieldItem, true, "")

	g.Log().Infof(ctx, "[Grab] 更新菜单项成功: merchantID=%s, itemID=%s", req.MerchantID, req.ItemID)
	return nil
}

// UpdateMenuModifier 更新单个修饰符
// 调用 GrabFood API PUT /partner/v1/merchants/menu/record 更新修饰符信息
// 支持更新：价格、可用状态、是否免费、高级定价配置
func (s *sGrabMenu) UpdateMenuModifier(ctx context.Context, req *grabDto.UpdateMenuModifierReq) error {
	g.Log().Infof(ctx, "[Grab] 更新菜单修饰符: merchantID=%s, modifierID=%s, modifierName=%s",
		req.MerchantID, req.ModifierID, req.ModifierName)

	// 1. 参数验证
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		g.Log().Errorf(ctx, "[Grab] 更新菜单修饰符参数验证失败: %v", err)
		return gerror.NewCode(gcode.CodeValidationFailed, err.Error())
	}

	// 2. 构建 SDK 请求
	updateModifier := req.ToSDKUpdateMenuModifier()
	updateReq := grabfood.UpdateMenuModifierAsUpdateMenuRequest(updateModifier)

	// 3. 调用 Grab API
	err := service.Grab().UpdateMenuRecord(ctx, req.MerchantID, updateReq)
	if err != nil {
		// 4. 记录失败日志
		s.logMenuRecordUpdate(ctx, req.MerchantID, req.ModifierID, grabDto.MenuItemUpdateFieldModifier, false, err.Error())

		g.Log().Errorf(ctx, "[Grab] 更新菜单修饰符 API 调用失败: merchantID=%s, modifierID=%s, error=%v",
			req.MerchantID, req.ModifierID, err)
		return gerror.Wrap(err, "调用 Grab UpdateMenuModifier API 失败")
	}

	// 5. 记录成功日志
	s.logMenuRecordUpdate(ctx, req.MerchantID, req.ModifierID, grabDto.MenuItemUpdateFieldModifier, true, "")

	g.Log().Infof(ctx, "[Grab] 更新菜单修饰符成功: merchantID=%s, modifierID=%s", req.MerchantID, req.ModifierID)
	return nil
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
		g.Log().Errorf(ctx, "[Grab] 插入菜单更新日志失败: merchantID=%s, recordID=%s, error=%v",
			merchantID, recordID, err)
	} else {
		g.Log().Debugf(ctx, "[Grab] 菜单更新日志已插入: logUUID=%d, merchantID=%s, recordID=%s, recordType=%s, success=%v",
			logUUID, merchantID, recordID, recordType, success)
	}
}
