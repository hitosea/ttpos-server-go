// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"

	"ttpos-bmp/app/ttpos-websocket/internal/model/dto"
)

// IWebsocket WebSocket 服务接口定义
// 提供 WebSocket 连接管理和消息推送功能
type IWebsocket interface {
	// PushMessage 推送消息到客户端
	// 根据条件筛选连接并推送消息
	// 参数：
	//   - ctx: 上下文对象
	//   - in: 推送消息输入参数
	// 返回：
	//   - out: 推送消息输出参数
	//   - err: 错误信息
	PushMessage(ctx context.Context, in *dto.PushMessageInput) (out *dto.PushMessageOutput, err error)

	// GetConnectionStats 获取连接统计信息
	// 返回当前所有 WebSocket 连接的统计数据
	// 参数：
	//   - ctx: 上下文对象
	//   - in: 获取连接统计输入参数
	// 返回：
	//   - out: 获取连接统计输出参数
	//   - err: 错误信息
	GetConnectionStats(ctx context.Context, in *dto.GetConnectionStatsInput) (out *dto.GetConnectionStatsOutput, err error)

	// CheckDeviceOnline 检查设备是否在线
	// 根据公司UUID和设备ID检查设备是否在线
	// 参数：
	//   - ctx: 上下文对象
	//   - in: 检查设备在线输入参数
	// 返回：
	//   - out: 检查设备在线输出参数
	//   - err: 错误信息
	CheckDeviceOnline(ctx context.Context, in *dto.CheckDeviceOnlineInput) (out *dto.CheckDeviceOnlineOutput, err error)

	// CloseConnection 关闭指定连接
	// 根据条件关闭 WebSocket 连接
	// 参数：
	//   - ctx: 上下文对象
	//   - in: 关闭连接输入参数
	// 返回：
	//   - out: 关闭连接输出参数
	//   - err: 错误信息
	CloseConnection(ctx context.Context, in *dto.CloseConnectionInput) (out *dto.CloseConnectionOutput, err error)

	// HandleConnections 处理 WebSocket 连接
	// 这是 HTTP 升级到 WebSocket 的处理函数
	// 参数：
	//   - r: GoFrame HTTP 请求对象
	HandleConnections(r *ghttp.Request)
}

var localWebsocket IWebsocket

// Websocket 获取 WebSocket 服务实例
func Websocket() IWebsocket {
	if localWebsocket == nil {
		panic("implement not found for interface IWebsocket, forgot register?")
	}
	return localWebsocket
}

// RegisterWebsocket 注册 WebSocket 服务实现
func RegisterWebsocket(i IWebsocket) {
	localWebsocket = i
}
