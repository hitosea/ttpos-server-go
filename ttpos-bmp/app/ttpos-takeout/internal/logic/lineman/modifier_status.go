package lineman

import (
	"context"

	lineman_dto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"

	"github.com/gogf/gf/v2/errors/gerror"
)

// IModifierStatusClient 修饰符状态客户端接口（用于测试 Mock）
type IModifierStatusClient interface {
	UpdateModifierStatusWithRetry(ctx context.Context, storeId string, req *lineman_dto.ModifierStatusUpdateReq) (*lineman_dto.ModifierStatusUpdateResp, error)
}

// UpdateModifierStatus 更新修饰符状态
//
// 参数:
//   - ctx: 上下文
//   - storeId: 店铺 ID（Lineman storeId）
//   - modifierId: 修饰符 ID（Partner Modifier ID）
//   - status: Lineman 状态（1=AVAILABLE, 2=SOLD_OUT_TODAY, 3=SUSPENDED）
//
// 返回:
//   - error: 错误信息
func (s *sLineman) UpdateModifierStatus(ctx context.Context, storeId string, modifierId string, status int) error {
	// 1. 参数校验
	if storeId == "" {
		return gerror.New("storeId 不能为空")
	}
	if modifierId == "" {
		return gerror.New("modifierId 不能为空")
	}
	if status != 1 && status != 2 && status != 3 {
		return gerror.Newf("无效的 status: %d（必须为 1, 2, 或 3）", status)
	}

	// 2. 构造请求
	req := &lineman_dto.ModifierStatusUpdateReq{
		PropertyValues: []lineman_dto.ModifierPropertyValue{
			{
				ID:     modifierId,
				Status: status,
			},
		},
	}

	// 3. 调用 Lineman Client
	resp, err := s.modifierStatusClient.UpdateModifierStatusWithRetry(ctx, storeId, req)
	if err != nil {
		return gerror.Wrap(err, "调用 Lineman API 失败")
	}

	// 4. 检查响应
	if resp.Status != "ok" {
		return gerror.Newf("Lineman API 返回错误: %s - %s", resp.Code, resp.Message)
	}

	return nil
}
