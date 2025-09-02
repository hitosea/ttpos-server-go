package callback

import (
	"context"
	"strings"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/internal/pkg/queue"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	v1 "ttpos-bmp/app/ttpos-erp/api/callback/v1"
)

// ItemChange 处理ERP商品变更回调
// 接收ERPNext系统的商品变更事件，验证请求参数并推送到消息队列进行异步处理
// 参数：
//   - ctx: 上下文对象，用于传递请求范围的元数据
//   - req: 商品变更请求参数，包含事件类型、数据、文档类型等
//
// 返回：
//   - res: 响应结果，包含处理状态码和消息
//   - err: 处理过程中产生的错误（若有）
func (c *ControllerV1) ItemChange(ctx context.Context, req *v1.ItemChangeReq) (res *v1.ItemChangeRes, err error) {
	// 参数验证
	// if err = c.validateItemChangeReq(req); err != nil {
	// 	return nil, gerror.Wrap(err, "商品变更请求参数验证失败")
	// }

	// 记录结构化日志
	g.Log().Infof(ctx, "接收ERP商品变更事件 - 站点: %s, 事件: %s, 文档类型: %s, 文档名称: %s",
		req.SiteCode, req.Event, req.DocType, req.DocName)

	// 推送到消息队列进行异步处理
	err = queue.PushWithContext(ctx, string(consts.TopicItemChange), req)
	if err != nil {
		return nil, gerror.Wrapf(err, "推送商品变更事件到队列失败 - 站点: %s, 事件: %s", req.SiteCode, req.Event)
	}

	// 返回成功响应
	return &v1.ItemChangeRes{
		Code: "0",       // 0表示成功
		Msg:  "success", // 成功消息
	}, nil
}

// validateItemChangeReq 验证商品变更请求参数
// 检查必填字段的有效性，确保请求数据的完整性
// 参数：req 商品变更请求参数
// 返回：验证失败时返回错误信息
func (c *ControllerV1) validateItemChangeReq(req *v1.ItemChangeReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}

	// 验证站点编码
	if strings.TrimSpace(req.SiteCode) == "" {
		return gerror.New("站点编码不能为空")
	}

	// 验证事件类型
	if strings.TrimSpace(req.Event) == "" {
		return gerror.New("事件类型不能为空")
	}

	// 验证文档类型
	if strings.TrimSpace(req.DocType) == "" {
		return gerror.New("文档类型不能为空")
	}

	// 验证文档名称
	if strings.TrimSpace(req.DocName) == "" {
		return gerror.New("文档名称不能为空")
	}

	return nil
}
