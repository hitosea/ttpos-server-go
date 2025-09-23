package permission

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/permission"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
)

// Controller 权限服务控制器
type Controller struct{}

// GetPosPermissionRuleList 获取POS权限规则列表
// 参数：ctx 上下文，req 获取权限规则列表请求
// 返回：响应信息和错误
func (c *Controller) GetPosPermissionRuleList(ctx context.Context, req *permission.GetPosPermissionRuleListReq) (*api.ResponseInfo, error) {
	// 参数验证
	if err := c.validateGetPosPermissionRuleListReq(req); err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 调用服务层获取数据
	resp, err := service.Permission().GetPosPermissionRuleList(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取权限规则列表成功", resp), nil
}

// validateGetPosPermissionRuleListReq 验证获取权限规则列表请求参数
func (c *Controller) validateGetPosPermissionRuleListReq(req *permission.GetPosPermissionRuleListReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}
	// 规则类型验证
	if len(req.RuleType) > 0 && req.RuleType != "White" && req.RuleType != "Black" {
		return gerror.New("规则类型必须为White或Black")
	}
	return nil
}

// GetPosPermissionRule 获取POS权限规则详情
// 参数：ctx 上下文，req 获取权限规则详情请求
// 返回：响应信息和错误
func (c *Controller) GetPosPermissionRule(ctx context.Context, req *permission.GetPosPermissionRuleReq) (*api.ResponseInfo, error) {
	// 参数验证
	if err := c.validateGetPosPermissionRuleReq(req); err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 调用服务层获取数据
	resp, err := service.Permission().GetPosPermissionRule(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取权限规则详情成功", resp), nil
}

// validateGetPosPermissionRuleReq 验证获取权限规则详情请求参数
func (c *Controller) validateGetPosPermissionRuleReq(req *permission.GetPosPermissionRuleReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}
	if len(req.Name) == 0 {
		return gerror.New("权限规则名称不能为空")
	}
	return nil
}

// CreatePosPermissionRule 创建POS权限规则
// 参数：ctx 上下文，req 创建权限规则请求
// 返回：响应信息和错误
func (c *Controller) CreatePosPermissionRule(ctx context.Context, req *permission.CreatePosPermissionRuleReq) (*api.ResponseInfo, error) {
	// 参数验证
	if err := c.validateCreatePosPermissionRuleReq(req); err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 调用服务层创建数据
	resp, err := service.Permission().CreatePosPermissionRule(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("创建权限规则成功", resp), nil
}

// validateCreatePosPermissionRuleReq 验证创建权限规则请求参数
func (c *Controller) validateCreatePosPermissionRuleReq(req *permission.CreatePosPermissionRuleReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}
	if len(req.RuleCode) == 0 {
		return gerror.New("规则代码不能为空")
	}
	if len(req.RuleName) == 0 {
		return gerror.New("规则名称不能为空")
	}
	if len(req.RuleType) == 0 {
		return gerror.New("规则类型不能为空")
	}
	if req.RuleType != "White" && req.RuleType != "Black" {
		return gerror.New("规则类型必须为White或Black")
	}
	return nil
}

// UpdatePosPermissionRule 更新POS权限规则
// 参数：ctx 上下文，req 更新权限规则请求
// 返回：响应信息和错误
func (c *Controller) UpdatePosPermissionRule(ctx context.Context, req *permission.UpdatePosPermissionRuleReq) (*api.ResponseInfo, error) {
	// 参数验证
	if err := c.validateUpdatePosPermissionRuleReq(req); err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 调用服务层更新数据
	resp, err := service.Permission().UpdatePosPermissionRule(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("更新权限规则成功", resp), nil
}

// validateUpdatePosPermissionRuleReq 验证更新权限规则请求参数
func (c *Controller) validateUpdatePosPermissionRuleReq(req *permission.UpdatePosPermissionRuleReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}
	if len(req.Name) == 0 {
		return gerror.New("权限规则名称不能为空")
	}
	// 规则类型验证（如果提供）
	if len(req.RuleType) > 0 && req.RuleType != "White" && req.RuleType != "Black" {
		return gerror.New("规则类型必须为White或Black")
	}
	return nil
}

// DeletePosPermissionRule 删除POS权限规则
// 参数：ctx 上下文，req 删除权限规则请求
// 返回：响应信息和错误
func (c *Controller) DeletePosPermissionRule(ctx context.Context, req *permission.DeletePosPermissionRuleReq) (*api.ResponseInfo, error) {
	// 参数验证
	if err := c.validateDeletePosPermissionRuleReq(req); err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 调用服务层删除数据
	resp, err := service.Permission().DeletePosPermissionRule(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("删除权限规则成功", resp), nil
}

// validateDeletePosPermissionRuleReq 验证删除权限规则请求参数
func (c *Controller) validateDeletePosPermissionRuleReq(req *permission.DeletePosPermissionRuleReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}
	if len(req.Name) == 0 {
		return gerror.New("权限规则名称不能为空")
	}
	return nil
}
