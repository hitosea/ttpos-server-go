// Package lineman 提供 LINE MAN 平台集成服务
package lineman

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

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
	cfg, err := service.ShopProviderCfg().GetShopProviderCfg(ctx, shopUUID, string(consts.ProviderLineman))
	if err != nil {
		return gerror.Wrap(err, "[Lineman] 获取门店配置失败")
	}
	if cfg == nil {
		return gerror.Newf("[Lineman] 门店未配置 Lineman: shopUUID=%d", shopUUID)
	}

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

// BuildMenuPayload 构建菜单数据
// 参数:
//   - ctx: 上下文
//   - ttposMenuJSON: TTPOS 菜单 JSON 字符串
//
// 返回:
//   - *lineman.MenuSyncRequest: 菜单数据
//   - error: 错误信息
func (s *sLineman) BuildMenuPayload(ctx context.Context, ttposMenuJSON string) (*lineman.MenuSyncRequest, error) {
	// 使用 data_mapper.go 中的逻辑构建数据
	mapper := NewDataMapper(ctx)
	return mapper.BuildMenuPayloadFromJSON(ttposMenuJSON)
}
