// Package lineman 提供 LINE MAN 平台集成服务
package lineman

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	lineman_dto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

// sLineman LINE MAN 服务（统一管理 Token 和菜单同步）
type sLineman struct {
	menuStatusLogic *MenuStatusLogic
}

func init() {
	service.RegisterLineman(New())
}

// New 创建 Lineman 服务实例
func New() *sLineman {
	return &sLineman{
		menuStatusLogic: NewMenuStatusLogicDefault(),
	}
}

// UpdateMenuItemStatus 更新菜单商品状态
//
// 参数:
//   - ctx: 上下文
//   - merchantId: 商户 ID (Grab MerchantID，对应 Lineman storeId)
//   - itemId: 商品 ID (partner item id)
//   - menuStatus: Lineman 状态 (AVAILABLE, SUSPENDED, SOLD_OUT_TODAY)
//
// 返回:
//   - error: 错误信息
func (s *sLineman) UpdateMenuItemStatus(ctx context.Context, merchantId string, itemId string, menuStatus string) error {
	if merchantId == "" {
		return gerror.New("merchantId 不能为空")
	}
	if itemId == "" {
		return gerror.New("itemId 不能为空")
	}
	if menuStatus == "" {
		return gerror.New("menuStatus 不能为空")
	}

	// 构造请求
	req := &lineman_dto.MenuStatusUpdateReq{
		MenuItems: []lineman_dto.MenuItemStatus{
			{
				ID:         itemId,
				MenuStatus: menuStatus,
			},
		},
	}

	// 调用 Logic 层
	// merchantId 在这里对应 Lineman 的 storeId
	return s.menuStatusLogic.UpdateMenuStatus(ctx, merchantId, req)
}
