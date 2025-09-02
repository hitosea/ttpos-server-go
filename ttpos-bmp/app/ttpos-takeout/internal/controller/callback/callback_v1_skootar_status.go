package callback

import (
	"context"
	"strings"

	"ttpos-bmp/app/ttpos-takeout/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	v1 "ttpos-bmp/app/ttpos-takeout/api/callback/v1"
)

// SkootarStatus 处理Skootar配送状态变更回调
// 接收Skootar配送平台的状态变更通知，验证请求参数并处理业务逻辑
// 参数：
//   - ctx: 上下文对象，用于传递请求范围的元数据
//   - req: Skootar状态变更请求参数，包含订单ID、状态变更、回调描述等
//
// 返回：
//   - res: 响应结果，包含处理状态码和消息
//   - err: 处理过程中产生的错误（若有）
func (c *ControllerV1) SkootarStatus(ctx context.Context, req *v1.SkootarStatusReq) (res *v1.SkootarStatusRes, err error) {
	// 参数验证
	// if err = c.validateSkootarStatusReq(req); err != nil {
	// 	return nil, gerror.Wrap(err, "Skootar状态回调请求参数验证失败")
	// }

	// 记录结构化日志
	g.Log().Infof(ctx, "接收Skootar状态变更回调 - 订单ID: %s, 状态变更: %d->%d, 回调类型: %s",
		req.JobId, req.StatusBefore, req.StatusAfter, req.CallbackDesc)

	// 调用服务层处理业务逻辑
	res, err = service.Skootar().JobStatusChange(ctx, req)
	if err != nil {
		return nil, gerror.Wrapf(err, "处理Skootar状态变更失败 - 订单ID: %s", req.JobId)
	}

	return res, nil
}

// validateSkootarStatusReq 验证Skootar状态回调请求参数
// 检查必填字段的有效性，确保请求数据的完整性
// 参数：req Skootar状态变更请求参数
// 返回：验证失败时返回错误信息
func (c *ControllerV1) validateSkootarStatusReq(req *v1.SkootarStatusReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}

	// 验证订单ID
	if strings.TrimSpace(req.JobId) == "" {
		return gerror.New("订单ID不能为空")
	}

	// 验证回调描述
	if strings.TrimSpace(req.CallbackDesc) == "" {
		return gerror.New("回调类型描述不能为空")
	}

	// 验证状态值范围（0-10）
	if req.StatusBefore < 0 || req.StatusBefore > 10 {
		return gerror.New("变更前状态值必须在0-10范围内")
	}

	if req.StatusAfter < 0 || req.StatusAfter > 10 {
		return gerror.New("变更后状态值必须在0-10范围内")
	}

	// 验证状态变更的合理性
	if req.StatusBefore == req.StatusAfter {
		return gerror.New("状态变更前后不能相同")
	}

	// 验证时间戳（检查是否为零值）
	if req.StatusDatetime.IsZero() {
		return gerror.New("状态变更时间不能为空")
	}

	return nil
}
