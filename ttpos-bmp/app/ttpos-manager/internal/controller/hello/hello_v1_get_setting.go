package hello

import (
	"context"
	"github.com/gogf/gf/v2/errors/gerror"
	"ttpos-bmp/app/ttpos-manager/api/hello/v1"
	rpc "ttpos-bmp/app/ttpos-manager/api/manager"
	"ttpos-bmp/app/ttpos-manager/api/svc"
	"ttpos-bmp/app/ttpos-manager/internal/service"
)

// GetSetting 处理获取配置信息的HTTP请求
// 参数：
//
//	ctx: 上下文对象，用于传递请求范围的元数据（如超时、取消信号等）
//	req: 请求参数结构体，包含要查询的配置Key（来自v1.GetSettingReq）
//
// 返回：
//
//	res: 响应结构体，包含查询到的配置信息（v1.GetSettingRes）
//	err: 操作过程中产生的错误（若有）
func (c *ControllerV1) GetSetting(ctx context.Context, req *v1.GetSettingReq) (res *v1.GetSettingRes, err error) {
	// 声明用于接收配置信息的指针变量
	var setting *rpc.Setting

	// 调用逻辑层的SettingClient获取配置信息
	// 传递请求中的Key参数（来自HTTP请求）到svc.GetReq结构体
	setting, err = service.Setting().GetSetting(ctx, &svc.GetReq{
		Key: req.Key,
	})
	if err != nil {
		// 错误包装：保留原始错误信息并添加业务含义描述
		err = gerror.Wrap(err, "获取设置失败")
		return // 直接返回错误，不构造响应
	}

	// 构造响应对象，将逻辑层返回的setting指针解引用后赋值给响应体
	res = &v1.GetSettingRes{
		Setting: *setting,
	}
	return
}
