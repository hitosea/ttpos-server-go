// Package websocket 实现 WebSocket 服务的 gRPC 控制器
package websocket

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	v1 "ttpos-bmp/app/ttpos-websocket/api/websocket"
	"ttpos-bmp/app/ttpos-websocket/internal/model/dto"
	"ttpos-bmp/app/ttpos-websocket/internal/service"
)

// Controller WebSocket 服务 gRPC 控制器
type Controller struct {
	v1.UnimplementedWebSocketServer
}

// New 创建 WebSocket 服务控制器实例
func New() *Controller {
	return &Controller{}
}

// PushMessage 推送消息到客户端
// 根据条件筛选连接并推送消息
// 参数：
//   - ctx: 上下文对象
//   - req: 推送消息请求参数
//
// 返回：
//   - res: 推送消息响应
//   - err: 错误信息
func (c *Controller) PushMessage(ctx context.Context, req *v1.PushMessageReq) (res *v1.PushMessageResp, err error) {
	g.Log().Info(ctx, "收到推送消息请求",
		"company_uuid", req.CompanyUuid,
		"message_type", req.MessageType,
		"source_client", req.SourceClient,
		"device_id", req.DeviceId,
	)

	// 参数验证
	if err = c.validatePushMessageReq(req); err != nil {
		g.Log().Warning(ctx, "推送消息参数验证失败", err)
		return &v1.PushMessageResp{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 构建输入参数
	in := &dto.PushMessageInput{
		CompanyUuid:  req.CompanyUuid,
		StaffUuid:    req.StaffUuid,
		NotStaffUuid: req.NotStaffUuid,
		SourceClient: req.SourceClient,
		DeviceId:     req.DeviceId,
		NotDeviceId:  req.NotDeviceId,
		MessageType:  req.MessageType,
		Data:         req.Data,
	}

	// 调用业务逻辑层
	out, err := service.Websocket().PushMessage(ctx, in)
	if err != nil {
		g.Log().Error(ctx, "推送消息失败", err)
		return &v1.PushMessageResp{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 构建响应
	res = &v1.PushMessageResp{
		Success:   out.Success,
		Message:   out.Message,
		PushCount: out.PushCount,
	}

	return res, nil
}

// GetConnectionStats 获取连接统计信息
// 返回当前所有 WebSocket 连接的统计数据
// 参数：
//   - ctx: 上下文对象
//   - req: 获取连接统计请求参数
//
// 返回：
//   - res: 获取连接统计响应
//   - err: 错误信息
func (c *Controller) GetConnectionStats(ctx context.Context, req *v1.GetConnectionStatsReq) (res *v1.GetConnectionStatsResp, err error) {
	g.Log().Info(ctx, "收到获取连接统计请求", "company_uuid", req.CompanyUuid)

	// 构建输入参数
	in := &dto.GetConnectionStatsInput{
		CompanyUuid: req.CompanyUuid,
	}

	// 调用业务逻辑层
	out, err := service.Websocket().GetConnectionStats(ctx, in)
	if err != nil {
		g.Log().Warning(ctx, "获取连接统计失败", err)
		return &v1.GetConnectionStatsResp{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 构建响应
	res = &v1.GetConnectionStatsResp{
		Success: out.Success,
		Message: out.Message,
		Stats: &v1.ConnectionStats{
			TotalConnections: out.TotalConnections,
			ByCompany:        out.ByCompany,
			BySource:         out.BySource,
			ByDevice:         out.ByDevice,
		},
	}

	return res, nil
}

// CheckDeviceOnline 检查设备是否在线
// 根据公司UUID和设备ID检查设备是否在线
// 参数：
//   - ctx: 上下文对象
//   - req: 检查设备在线请求参数
//
// 返回：
//   - res: 检查设备在线响应
//   - err: 错误信息
func (c *Controller) CheckDeviceOnline(ctx context.Context, req *v1.CheckDeviceOnlineReq) (res *v1.CheckDeviceOnlineResp, err error) {
	g.Log().Info(ctx, "收到检查设备在线请求",
		"company_uuid", req.CompanyUuid,
		"device_id", req.DeviceId,
	)

	// 参数验证
	if err = c.validateCheckDeviceOnlineReq(req); err != nil {
		g.Log().Warning(ctx, "检查设备在线参数验证失败", err)
		return &v1.CheckDeviceOnlineResp{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 构建输入参数
	in := &dto.CheckDeviceOnlineInput{
		CompanyUuid:  req.CompanyUuid,
		SourceClient: req.SourceClient,
		DeviceId:     req.DeviceId,
	}

	// 调用业务逻辑层
	out, err := service.Websocket().CheckDeviceOnline(ctx, in)
	if err != nil {
		g.Log().Warning(ctx, "检查设备在线失败", err)
		return &v1.CheckDeviceOnlineResp{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 构建响应
	res = &v1.CheckDeviceOnlineResp{
		Success:       out.Success,
		Message:       out.Message,
		IsOnline:      out.IsOnline,
		LastHeartbeat: out.LastHeartbeat,
	}

	return res, nil
}

// CloseConnection 关闭指定连接
// 根据条件关闭 WebSocket 连接
// 参数：
//   - ctx: 上下文对象
//   - req: 关闭连接请求参数
//
// 返回：
//   - res: 关闭连接响应
//   - err: 错误信息
func (c *Controller) CloseConnection(ctx context.Context, req *v1.CloseConnectionReq) (res *v1.CloseConnectionResp, err error) {
	g.Log().Info(ctx, "收到关闭连接请求",
		"company_uuid", req.CompanyUuid,
		"source_client", req.SourceClient,
		"device_id", req.DeviceId,
	)

	// 参数验证
	if req.CompanyUuid == 0 {
		return &v1.CloseConnectionResp{
			Success: false,
			Message: "公司UUID不能为空",
		}, nil
	}

	// 构建输入参数
	in := &dto.CloseConnectionInput{
		CompanyUuid:  req.CompanyUuid,
		SourceClient: req.SourceClient,
		DeviceId:     req.DeviceId,
	}

	// 调用业务逻辑层
	out, err := service.Websocket().CloseConnection(ctx, in)
	if err != nil {
		g.Log().Warning(ctx, "关闭连接失败", err)
		return &v1.CloseConnectionResp{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 构建响应
	res = &v1.CloseConnectionResp{
		Success:     out.Success,
		Message:     out.Message,
		ClosedCount: out.ClosedCount,
	}

	return res, nil
}

// validatePushMessageReq 验证推送消息请求参数
func (c *Controller) validatePushMessageReq(req *v1.PushMessageReq) error {
	if req.CompanyUuid == 0 {
		return gerror.New("公司UUID不能为空")
	}
	if req.MessageType == "" {
		return gerror.New("消息类型不能为空")
	}
	if req.Data == "" {
		return gerror.New("消息数据不能为空")
	}
	return nil
}

// validateCheckDeviceOnlineReq 验证检查设备在线请求参数
func (c *Controller) validateCheckDeviceOnlineReq(req *v1.CheckDeviceOnlineReq) error {
	if req.CompanyUuid == 0 {
		return gerror.New("公司UUID不能为空")
	}
	if req.DeviceId == "" {
		return gerror.New("设备ID不能为空")
	}
	return nil
}
