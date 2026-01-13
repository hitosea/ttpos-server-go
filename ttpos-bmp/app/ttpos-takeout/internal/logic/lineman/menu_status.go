package lineman

import (
	"context"

	lineman_dto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"

	"github.com/gogf/gf/v2/errors/gerror"
)

// IMenuStatusClient 菜单状态客户端接口（用于测试 Mock）
type IMenuStatusClient interface {
	UpdateMenuStatusWithRetry(ctx context.Context, storeId string, req *lineman_dto.MenuStatusUpdateReq) (*lineman_dto.MenuStatusUpdateResp, error)
}

// MapStatusToLineman 将 TTPOS 状态映射为 Lineman 状态
// TTPOS 状态 -> Lineman 状态:
// AVAILABLE -> AVAILABLE
// UNAVAILABLE -> SUSPENDED
// SOLD_OUT_TODAY -> SOLD_OUT_TODAY
// UNAVAILABLEHIDE -> 不支持（返回错误）
//
// 参考: https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=585076633#gid=585076633
func MapStatusToLineman(ttposStatus string) (string, error) {
	switch ttposStatus {
	case "AVAILABLE":
		return "AVAILABLE", nil
	case "UNAVAILABLE":
		return "SUSPENDED", nil
	case "SOLD_OUT_TODAY":
		return "SOLD_OUT_TODAY", nil
	case "UNAVAILABLEHIDE":
		return "", gerror.New("Lineman 平台不支持 UNAVAILABLEHIDE 状态")
	case "":
		return "", gerror.New("available_status 不能为空")
	default:
		return "", gerror.Newf("不支持的状态: %s", ttposStatus)
	}
}

// UpdateMenuItemStatus 更新菜单商品状态（单个商品）
//
// 参数:
//   - ctx: 上下文
//   - shopUuid: 店铺 UUID（内部会查询对应的 Lineman storeId）
//   - itemId: 商品 ID (partner item id)
//   - menuStatus: Lineman 状态 (AVAILABLE, SUSPENDED, SOLD_OUT_TODAY)
//
// 返回:
//   - error: 错误信息
func (s *sLineman) UpdateMenuItemStatus(ctx context.Context, shopUuid string, itemId string, menuStatus string) error {
	// 1. 参数校验
	if shopUuid == "" {
		return gerror.New("shopUuid 不能为空")
	}
	if itemId == "" {
		return gerror.New("itemId 不能为空")
	}
	if menuStatus == "" {
		return gerror.New("menuStatus 不能为空")
	}

	// 2. 构造请求
	req := &lineman_dto.MenuStatusUpdateReq{
		MenuItems: []lineman_dto.MenuItemStatus{
			{
				ID:         itemId,
				MenuStatus: menuStatus,
			},
		},
	}

	// 3. 调用通用方法（shopUuid 对应 Lineman 的 storeId）
	return s.doUpdateMenuStatus(ctx, shopUuid, req)
}

// UpdateMenuStatus 批量更新菜单商品状态
//
// 参数:
//   - ctx: 上下文
//   - storeId: 店铺 ID（Lineman storeId）
//   - req: 菜单状态更新请求
//
// 返回:
//   - error: 错误信息
func (s *sLineman) UpdateMenuStatus(ctx context.Context, storeId string, req *lineman_dto.MenuStatusUpdateReq) error {
	// 1. 参数校验
	if storeId == "" {
		return gerror.New("storeId 不能为空")
	}
	if len(req.MenuItems) == 0 {
		return gerror.New("menuItems 不能为空")
	}
	if len(req.MenuItems) > 100 {
		return gerror.New("menuItems 最多支持 100 个商品")
	}

	// 2. 调用通用方法
	return s.doUpdateMenuStatus(ctx, storeId, req)
}

// doUpdateMenuStatus 执行菜单状态更新（通用方法）
//
// 参数:
//   - ctx: 上下文
//   - storeId: 店铺 ID
//   - req: 菜单状态更新请求
//
// 返回:
//   - error: 错误信息
func (s *sLineman) doUpdateMenuStatus(ctx context.Context, storeId string, req *lineman_dto.MenuStatusUpdateReq) error {
	// 调用 Lineman Client
	resp, err := s.menuStatusClient.UpdateMenuStatusWithRetry(ctx, storeId, req)
	if err != nil {
		return gerror.Wrap(err, "调用 Lineman API 失败")
	}

	// 检查响应
	if resp.Status != "ok" {
		return gerror.Newf("Lineman API 返回错误: %s - %s", resp.Code, resp.Message)
	}

	return nil
}
