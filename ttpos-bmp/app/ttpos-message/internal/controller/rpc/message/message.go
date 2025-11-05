// Package message 实现消息服务的 gRPC 控制器
package message

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	v1 "ttpos-bmp/app/ttpos-message/api/message"
	"ttpos-bmp/app/ttpos-message/internal/consts"
	"ttpos-bmp/app/ttpos-message/internal/model/dto"
	"ttpos-bmp/app/ttpos-message/internal/service"
)

// Controller 消息服务 gRPC 控制器
type Controller struct {
	v1.UnimplementedMessageServer
}

// New 创建消息服务控制器实例
func New() *Controller {
	return &Controller{}
}

// SendMessage 发送消息
// 接收消息发送请求，进行参数验证后调用业务逻辑层处理
// 参数：
//   - ctx: 上下文对象
//   - req: 发送消息请求参数
//
// 返回：
//   - res: 发送消息响应
//   - err: 错误信息
func (c *Controller) SendMessage(ctx context.Context, req *v1.SendMessageReq) (res *v1.SendMessageResp, err error) {
	g.Log().Info(ctx, "收到发送消息请求",
		"uuid", req.MessageUuid,
		"template_uuid", req.TemplateUuid,
		"type", req.MessageType,
		"recipient", req.Recipient,
	)

	// 参数验证
	if err = c.validateSendMessageReq(req); err != nil {
		g.Log().Warning(ctx, "发送消息参数验证失败", err)
		return &v1.SendMessageResp{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 构建输入参数
	in := &dto.SendMessageInput{
		MessageUuid:  req.MessageUuid,
		TemplateUUID: req.TemplateUuid,
		MessageArgs:  req.MessageArgs,
		MessageType:  req.MessageType,
		Recipient:    req.Recipient,
		Subject:      req.Subject,
		CompanyUuid:  req.CompanyUuid,
		OperatorUuid: req.OperatorUuid,
	}

	// 调用业务逻辑层
	out, err := service.Message().SendMessage(ctx, in)
	if err != nil {
		g.Log().Error(ctx, "发送消息失败", err)
		return &v1.SendMessageResp{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 构建响应
	res = &v1.SendMessageResp{
		Success:     out.Success,
		Message:     out.Message,
		MessageUuid: out.MessageUuid,
	}

	return res, nil
}

// GetMessageStatus 查询消息状态
// 根据消息UUID查询消息的发送状态和详细信息
// 参数：
//   - ctx: 上下文对象
//   - req: 查询消息状态请求参数
//
// 返回：
//   - res: 查询消息状态响应
//   - err: 错误信息
func (c *Controller) GetMessageStatus(ctx context.Context, req *v1.GetMessageStatusReq) (res *v1.GetMessageStatusResp, err error) {
	g.Log().Info(ctx, "收到查询消息状态请求", "uuid", req.MessageUuid)

	// 参数验证
	if req.MessageUuid == "" {
		return &v1.GetMessageStatusResp{
			Success: false,
			Message: "消息UUID不能为空",
		}, nil
	}

	// 构建输入参数
	in := &dto.GetMessageStatusInput{
		MessageUuid: req.MessageUuid,
	}

	// 调用业务逻辑层
	out, err := service.Message().GetMessageStatus(ctx, in)
	if err != nil {
		g.Log().Warning(ctx, "查询消息状态失败", err)
		return &v1.GetMessageStatusResp{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 构建响应
	res = &v1.GetMessageStatusResp{
		Success: out.Success,
		Message: out.Message,
	}

	if out.MessageInfo != nil {
		res.MessageInfo = &v1.MessageInfo{
			MessageId:    out.MessageInfo.Id,
			MessageUuid:  out.MessageInfo.Uuid,
			TemplateId:   out.MessageInfo.TemplateId,
			MessageType:  out.MessageInfo.MessageType,
			Recipient:    out.MessageInfo.Recipient,
			Status:       int32(out.MessageInfo.Status),
			ErrorMessage: out.MessageInfo.ErrorMessage,
			CreatedAt:    out.MessageInfo.CreatedAt,
			SendTime:     out.MessageInfo.SendTime,
		}
	}

	return res, nil
}

// ResendMessage 重发消息
// 针对发送失败的消息，支持重新发送
// 参数：
//   - ctx: 上下文对象
//   - req: 重发消息请求参数
//
// 返回：
//   - res: 重发消息响应
//   - err: 错误信息
func (c *Controller) ResendMessage(ctx context.Context, req *v1.ResendMessageReq) (res *v1.ResendMessageResp, err error) {
	g.Log().Info(ctx, "收到重发消息请求", "uuid", req.MessageUuid)

	// 参数验证
	if req.MessageUuid == "" {
		return &v1.ResendMessageResp{
			Success: false,
			Message: "消息UUID不能为空",
		}, nil
	}

	// 构建输入参数
	in := &dto.ResendMessageInput{
		MessageUuid: req.MessageUuid,
	}

	// 调用业务逻辑层
	out, err := service.Message().ResendMessage(ctx, in)
	if err != nil {
		g.Log().Warning(ctx, "重发消息失败", err)
		return &v1.ResendMessageResp{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 构建响应
	res = &v1.ResendMessageResp{
		Success: out.Success,
		Message: out.Message,
	}

	return res, nil
}

// validateSendMessageReq 验证发送消息请求参数
func (c *Controller) validateSendMessageReq(req *v1.SendMessageReq) error {
	if req.MessageUuid == "" {
		return gerror.New("消息UUID不能为空")
	}
	if req.TemplateUuid == "" {
		return gerror.New("模板UUID不能为空")
	}
	if req.MessageType == "" {
		return gerror.New("消息类型不能为空")
	}
	if req.MessageType != consts.MessageTypeEmail && req.MessageType != consts.MessageTypeSMS {
		return gerror.New("消息类型只能是email或sms")
	}
	if req.Recipient == "" {
		return gerror.New("接收人不能为空")
	}
	return nil
}
