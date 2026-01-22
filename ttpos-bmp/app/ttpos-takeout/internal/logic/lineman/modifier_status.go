package lineman

import (
	"context"
	"fmt"

	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	lineman_dto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
	"ttpos-bmp/app/ttpos-takeout/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
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
		// 记录失败日志
		s.logModifierStatusUpdate(ctx, storeId, req, false, err.Error())
		return gerror.Wrap(err, "调用 Lineman API 失败")
	}

	// 4. 检查响应
	if resp.Status != "ok" {
		errMsg := fmt.Sprintf("Lineman API 返回错误: %s - %s", resp.Code, resp.Message)
		// 记录失败日志
		s.logModifierStatusUpdate(ctx, storeId, req, false, errMsg)
		return gerror.New(errMsg)
	}

	// 5. 记录成功日志
	s.logModifierStatusUpdate(ctx, storeId, req, true, "")

	g.Log().Infof(ctx, "[Lineman] 更新修饰符状态成功: storeId=%s, modifierId=%s, status=%d", storeId, modifierId, status)

	return nil
}

// logModifierStatusUpdate 记录修饰符状态更新日志
// 内部方法，记录到 menu_log 表
func (s *sLineman) logModifierStatusUpdate(ctx context.Context, storeId string, req *lineman_dto.ModifierStatusUpdateReq, success bool, errMsg string) {
	// 将请求参数转为 JSON 作为快照
	menuSnapshot, _ := gjson.EncodeString(req)

	err := service.ChannelMenu().LogMenuSync(
		ctx,
		storeId,
		string(consts.ProviderLineman),
		"UPDATE_MODIFIER", // 参考 grab 的命名: UPDATE_ITEM, UPDATE_MODIFIER
		"",                // requestId 为空，Lineman 更新接口不返回 requestId
		success,
		menuSnapshot,
		errMsg,
	)
	if err != nil {
		g.Log().Errorf(ctx, "[Lineman] 插入修饰符状态更新日志失败: storeId=%s, error=%v",
			storeId, err)
	} else {
		g.Log().Debugf(ctx, "[Lineman] 修饰符状态更新日志已插入: storeId=%s, success=%v",
			storeId, success)
	}
}
