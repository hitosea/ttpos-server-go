package lineman

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	lineman_client "ttpos-bmp/app/ttpos-takeout/internal/client/lineman"
	lineman_dto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
)

// IModifierStatusClient 修饰符状态客户端接口（用于测试 Mock）
type IModifierStatusClient interface {
	UpdateModifierStatusWithRetry(ctx context.Context, storeId string, req *lineman_dto.ModifierStatusUpdateReq) (*lineman_dto.ModifierStatusUpdateResp, error)
}

// MapStatusToLinemanModifier 将 TTPOS 状态（string）映射为 Lineman 状态（int）
//
// 映射规则:
//   - "AVAILABLE" → 1 (AVAILABLE)
//   - "UNAVAILABLE" → 3 (SUSPENDED)
//   - "SOLD_OUT_TODAY" → 2 (SOLD_OUT_TODAY)
//   - 其他值 → 返回错误
//
// 参数:
//   - ttposStatus: TTPOS 状态（string）
//
// 返回:
//   - int: Lineman 状态（1, 2, 3）
//   - error: 错误信息（不支持的状态）
//
// 参考: https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=1934684079#gid=1934684079
func MapStatusToLinemanModifier(ttposStatus string) (int, error) {
	switch ttposStatus {
	case "AVAILABLE":
		return 1, nil // AVAILABLE
	case "UNAVAILABLE":
		return 3, nil // SUSPENDED
	case "SOLD_OUT_TODAY":
		return 2, nil // SOLD_OUT_TODAY
	case "":
		return 0, gerror.New("available_status 不能为空")
	default:
		return 0, gerror.Newf("不支持的状态: %s", ttposStatus)
	}
}

// ModifierStatusLogic Lineman 修饰符状态业务逻辑
type ModifierStatusLogic struct {
	client IModifierStatusClient
}

// NewModifierStatusLogic 创建 Lineman 修饰符状态业务逻辑实例
func NewModifierStatusLogic(client IModifierStatusClient) *ModifierStatusLogic {
	return &ModifierStatusLogic{
		client: client,
	}
}

// NewModifierStatusLogicDefault 创建默认 Lineman 修饰符状态业务逻辑实例
func NewModifierStatusLogicDefault() *ModifierStatusLogic {
	return &ModifierStatusLogic{
		client: lineman_client.NewModifierStatusClient(),
	}
}

// UpdateModifierStatus 更新修饰符状态
//
// 参数:
//   - ctx: 上下文
//   - storeId: 店铺 ID（对应 Lineman storeId）
//   - modifierId: 修饰符 ID（Partner Modifier ID）
//   - status: Lineman 状态（1=AVAILABLE, 2=SOLD_OUT_TODAY, 3=SUSPENDED）
//
// 返回:
//   - error: 错误信息
func (l *ModifierStatusLogic) UpdateModifierStatus(
	ctx context.Context,
	storeId string,
	modifierId string,
	status int,
) error {
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
	resp, err := l.client.UpdateModifierStatusWithRetry(ctx, storeId, req)
	if err != nil {
		return gerror.Wrap(err, "调用 Lineman API 失败")
	}

	// 4. 检查响应
	if resp.Status != "ok" {
		return gerror.Newf("Lineman API 返回错误: %s - %s", resp.Code, resp.Message)
	}

	return nil
}
