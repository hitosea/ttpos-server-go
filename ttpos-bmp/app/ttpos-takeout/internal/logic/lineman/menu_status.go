package lineman

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	lineman_client "ttpos-bmp/app/ttpos-takeout/internal/client/lineman"
	lineman_dto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
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

// MenuStatusLogic Lineman 菜单状态业务逻辑
type MenuStatusLogic struct {
	client IMenuStatusClient
}

// NewMenuStatusLogic 创建 Lineman 菜单状态业务逻辑实例
func NewMenuStatusLogic(client IMenuStatusClient) *MenuStatusLogic {
	return &MenuStatusLogic{
		client: client,
	}
}

// NewMenuStatusLogicDefault 创建默认 Lineman 菜单状态业务逻辑实例
func NewMenuStatusLogicDefault() *MenuStatusLogic {
	return &MenuStatusLogic{
		client: lineman_client.NewMenuStatusClient(),
	}
}

// UpdateMenuStatus 更新菜单状态
//
// 参数:
//   - ctx: 上下文
//   - storeId: 店铺 ID
//   - req: 菜单状态更新请求
//
// 返回:
//   - err: 错误信息
func (l *MenuStatusLogic) UpdateMenuStatus(ctx context.Context, storeId string, req *lineman_dto.MenuStatusUpdateReq) error {
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

	// 2. 调用 Lineman Client
	resp, err := l.client.UpdateMenuStatusWithRetry(ctx, storeId, req)
	if err != nil {
		return gerror.Wrap(err, "调用 Lineman API 失败")
	}

	// 3. 检查响应
	if resp.Status != "ok" {
		return gerror.Newf("Lineman API 返回错误: %s - %s", resp.Code, resp.Message)
	}

	return nil
}
