// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
)

type (
	IPermission interface {
		// GetPosPermissionRuleList 获取POS权限规则列表
		// 根据查询条件过滤并返回权限规则信息列表
		GetPosPermissionRuleList(ctx context.Context, req *erp.PosPermissionRule) (res []*erp.PosPermissionRule, err error)
		// GetPosPermissionRule 获取POS权限规则详情
		// 根据名称标识获取权限规则的完整信息
		GetPosPermissionRule(ctx context.Context, ruleCode string) (res *erp.PosPermissionRule, err error)
		// CreatePosPermissionRule 创建POS权限规则
		// 创建新的权限规则记录
		CreatePosPermissionRule(ctx context.Context, req *erp.PosPermissionRule) (res *erp.PosPermissionRule, err error)
		// UpdatePosPermissionRule 更新POS权限规则
		// 更新现有的权限规则记录
		UpdatePosPermissionRule(ctx context.Context, req *erp.PosPermissionRule) (res *erp.PosPermissionRule, err error)
		// DeletePosPermissionRule 删除POS权限规则
		// 删除指定的权限规则记录
		DeletePosPermissionRule(ctx context.Context, ruleCode string) (err error)
		// CheckPermission 检查权限
		// 根据权限规则列表和公司名称检查是否有权限
		// 参数：
		//   - ctx: 上下文对象
		//   - permissionList: 权限规则列表
		//   - company: 公司名称
		//
		// 返回：
		//   - hasPermission: 是否有权限
		//   - err: 错误信息
		CheckPermission(ctx context.Context, permissionList []erp.PermissionRule, company string) (hasPermission bool, err error)
	}
)

var (
	localPermission IPermission
)

func Permission() IPermission {
	if localPermission == nil {
		panic("implement not found for interface IPermission, forgot register?")
	}
	return localPermission
}

func RegisterPermission(i IPermission) {
	localPermission = i
}
