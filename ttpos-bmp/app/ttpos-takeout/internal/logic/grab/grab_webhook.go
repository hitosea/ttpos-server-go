// Package grab 提供 GrabFood API 集成的业务逻辑
package grab

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"ttpos-bmp/app/ttpos-takeout/internal/dao"
	"ttpos-bmp/app/ttpos-takeout/internal/model/do"
	"ttpos-bmp/app/ttpos-takeout/utility"
	"ttpos-bmp/utility/uuid"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	grabfood "github.com/grab/grabfood-api-sdk-go"

	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	grabDto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
	ttposDto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/ttpos"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

// ============================================================================
// Webhook 处理 (包装层 - 处理一些额外逻辑)
// ============================================================================

// HandleGetMenu 处理 Grab 获取菜单请求 (Partner Endpoint)
// 签名验证已由中间件完成
func (s *sGrab) HandleGetMenu(ctx context.Context, merchantID string) (*grabfood.GetMenuNewResponse, error) {
	g.Log().Infof(ctx, "[Grab] 收到获取菜单请求: partnerMerchantID=%s", merchantID)

	// 1. 将 partnerMerchantID 转换为 shopUUID (uint64)
	// 假设 partnerMerchantID 是数字字符串格式的 shopUUID
	shopUUID := g.NewVar(merchantID).Uint64()
	if shopUUID == 0 {
		g.Log().Errorf(ctx, "[Grab] partnerMerchantID 格式无效: %s", merchantID)
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "partnerMerchantID 格式无效")
	}

	// 2. 从本地快照读取菜单
	menuJSON, err := service.ChannelMenu().GetTtposMenu(ctx, shopUUID, string(consts.ProviderGrab))
	if err != nil {
		g.Log().Errorf(ctx, "[Grab] 获取渠道菜单失败: shopUUID=%d, error: %v", shopUUID, err)
		return nil, gerror.Wrap(err, "获取渠道菜单失败")
	}

	// 3. 如果本地快照为空，回退调用 TTPOS 导出接口
	if menuJSON == "" {
		g.Log().Infof(ctx, "[Grab] 本地菜单快照不存在，回退调用 TTPOS 导出接口: shopUUID=%d", shopUUID)
		return s.fetchMenuFromTTpos(ctx, shopUUID)
	}

	// 4. 解析本地快照 JSON 为 PushGrabMenuDTO (使用 SDK 类型)
	var pushDTO grabDto.PushGrabMenuDTO
	if err := gjson.New([]byte(menuJSON)).Scan(&pushDTO); err != nil {
		g.Log().Errorf(ctx, "[Grab] 解析菜单数据失败: error: %v", err)
		return nil, gerror.Wrap(err, "解析菜单数据失败")
	}

	// 5. 构建响应结构
	menuResp := &grabfood.GetMenuNewResponse{
		Currency:     pushDTO.Currency,
		SellingTimes: pushDTO.SellingTimes,
		Categories:   pushDTO.Categories,
	}

	g.Log().Infof(ctx, "[Grab] 获取菜单成功（来自本地）: categories=%d", len(menuResp.Categories))

	// 6. 设置 MerchantID 和 PartnerMerchantID
	return s.setMenuResponseIdentifiers(ctx, shopUUID, menuResp)
}

// HandlePushGrabMenu 处理 Grab 菜单推送 Webhook
// 此方法组合了 SaveMenuSnapshot 和 NotifyMenuUpdate 两个操作
func (s *sGrab) HandlePushGrabMenu(ctx context.Context, dto *grabDto.PushGrabMenuDTO) error {
	// 1. Save Snapshot
	menuUuid, err := s.SaveMenuSnapshot(ctx, dto)
	if err != nil {
		return err
	}

	// 2. Notify
	event := &grabDto.ProviderMenuUpdateEvent{
		ProviderName:      string(consts.ProviderGrab),
		MerchantID:        dto.MerchantID,
		PartnerMerchantID: dto.PartnerMerchantID,
		ShopUuid:          dto.PartnerMerchantID,
		Uuid:              menuUuid,
		ReceivedAt:        gtime.Timestamp(),
	}
	return s.NotifyMenuUpdateEvent(ctx, event)
}

// HandleMenuSyncState 处理菜单同步状态回调
// 使用 SDK grabfood.MenuSyncWebhookRequest
func (s *sGrab) HandleMenuSyncState(ctx context.Context, req *grabfood.MenuSyncWebhookRequest) error {
	requestID := req.GetRequestID()
	status := req.GetStatus()

	g.Log().Infof(ctx, "[Grab] 处理菜单同步状态: requestID=%s, merchantID=%s, status=%s",
		requestID, req.GetMerchantID(), status)

	// 通过 Grab MerchantID 查询 shopUUID
	merchantID := req.GetMerchantID()
	cfg, err := service.ShopProviderCfg().GetShopProviderCfgByMerchantID(ctx, merchantID, string(consts.ProviderGrab))
	if err != nil {
		g.Log().Warningf(ctx, "[Grab] 获取门店配置失败: merchantID=%s, error: %v", merchantID, err)
		// 不中断流程，继续处理 MenuLog 插入
	} else if cfg == nil || cfg.Id == 0 {
		g.Log().Warningf(ctx, "[Grab] 未找到对应的门店配置: merchantID=%s", merchantID)
		// 不中断流程，继续处理 MenuLog 插入
	} else {
		shopUUID := cfg.ShopUuid
		// 更新 channel_menu_snapshot 的状态
		_, err := dao.ChannelMenuSnapshot.Ctx(ctx).
			Where(dao.ChannelMenuSnapshot.Columns().ShopUuid, shopUUID).
			Where(dao.ChannelMenuSnapshot.Columns().ProviderName, string(consts.ProviderGrab)).
			Data(g.Map{
				dao.ChannelMenuSnapshot.Columns().SyncState: status,
				dao.ChannelMenuSnapshot.Columns().UpdatedAt: gtime.Now().Unix(),
			}).Update()
		if err != nil {
			g.Log().Errorf(ctx, "[Grab] 更新渠道菜单快照状态失败: shopUUID=%d, status=%s, error: %v", shopUUID, status, err)
			// 不中断流程，继续处理 MenuLog 插入
		} else {
			g.Log().Infof(ctx, "[Grab] 渠道菜单快照状态已更新: shopUUID=%d, status=%s", shopUUID, status)
		}
	}

	// 插入新的菜单日志记录（每次状态回调都插入新记录）
	logUUID := uuid.MustGetID()
	errMsg := ""
	if status == grabDto.MenuSyncStatusFail && len(req.GetErrors()) > 0 {
		errMsg = strings.Join(req.GetErrors(), "; ")
	}

	logDo := &do.MenuLog{
		Uuid:         logUUID,
		MerchantId:   req.GetMerchantID(),
		ProviderName: string(consts.ProviderGrab),
		Status:       status,
		ErrorMsg:     errMsg,
	}
	_, err = dao.MenuLog.Ctx(ctx).Data(logDo).Insert()
	if err != nil {
		g.Log().Errorf(ctx, "[Grab] 插入菜单日志失败: %v", err)
	}
	g.Log().Infof(ctx, "[Grab] 菜单同步状态已处理: requestId=%s, status=%s, logUUID=%v", requestID, status, logUUID)
	return nil
}

// fetchMenuFromTTpos 从 TTPOS 主模块获取菜单数据
// 当本地菜单快照为空时，回退调用此方法
func (s *sGrab) fetchMenuFromTTpos(ctx context.Context, shopUUID uint64) (*grabfood.GetMenuNewResponse, error) {
	// 1. 使用 utility 工具获取带认证的客户端
	client, err := utility.GetTtposClientWithAuth(ctx, fmt.Sprintf("%d", shopUUID))
	if err != nil {
		return nil, gerror.Wrap(err, "生成 TTPOS 认证头失败")
	}

	// 2. 构建请求体
	reqBody := g.Map{
		"platform":    string(consts.ProviderGrab),
		"companyUuid": shopUUID,
	}

	// 3. 发起 HTTP 请求
	resp, err := client.Post(ctx, "/api/v1/takeout/menu/export", reqBody)
	if err != nil {
		return nil, gerror.Wrap(err, "调用 TTPOS 导出接口失败")
	}
	defer resp.Close()

	// 4. 检查 HTTP 状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, gerror.Newf("TTPOS 导出接口返回错误状态码 %d: %s", resp.StatusCode, resp.ReadAllString())
	}

	// 5. 解析响应
	var result ttposDto.GetMenuExportResp
	if err := json.Unmarshal(resp.ReadAll(), &result); err != nil {
		return nil, gerror.Wrap(err, "解析 TTPOS 导出接口响应失败")
	}

	// 6. 检查业务状态码（兼容 code=200 和 code=1 两种成功状态）
	if result.Code != 200 && result.Code != 1 {
		return nil, gerror.Newf("TTPOS 导出接口业务错误: code=%d, message=%s", result.Code, result.Message)
	}

	g.Log().Infof(ctx, "[Grab] 从 TTPOS 获取菜单成功（原始）: shopUUID=%d, categories=%d",
		shopUUID, len(result.Data.MenuData.Categories))

	// 7. 设置 MerchantID 和 PartnerMerchantID
	return s.setMenuResponseIdentifiers(ctx, shopUUID, &result.Data.MenuData)
}

// setMenuResponseIdentifiers 为菜单响应设置 MerchantID 和 PartnerMerchantID
// 通过查询 ShopProviderCfg 获取 Grab MerchantID，并设置到响应对象中
func (s *sGrab) setMenuResponseIdentifiers(ctx context.Context, shopUUID uint64, menuResp *grabfood.GetMenuNewResponse) (*grabfood.GetMenuNewResponse, error) {
	// 1. 查询 ShopProviderCfg 获取 Grab MerchantID
	cfg, err := service.ShopProviderCfg().GetShopProviderCfg(ctx, shopUUID, string(consts.ProviderGrab))
	if err != nil {
		g.Log().Errorf(ctx, "[Grab] 获取门店第三方配置失败: shopUUID=%d, error: %v", shopUUID, err)
		return nil, gerror.Wrap(err, "获取门店第三方配置失败")
	}
	if cfg == nil {
		g.Log().Errorf(ctx, "[Grab] 门店第三方配置不存在: shopUUID=%d, provider=%s", shopUUID, consts.ProviderGrab)
		return nil, gerror.NewCode(gcode.CodeNotFound, "门店第三方配置不存在")
	}

	// 2. 设置 MerchantID 和 PartnerMerchantID
	merchantIDStr := cfg.ProviderMerchantId
	partnerMerchantIDStr := fmt.Sprintf("%d", shopUUID)
	menuResp.MerchantID = &merchantIDStr
	menuResp.PartnerMerchantID = &partnerMerchantIDStr

	g.Log().Infof(ctx, "[Grab] 菜单标识符已设置: merchantID=%v, partnerMerchantID=%v, categories=%d",
		menuResp.MerchantID, menuResp.PartnerMerchantID, len(menuResp.Categories))

	return menuResp, nil
}
