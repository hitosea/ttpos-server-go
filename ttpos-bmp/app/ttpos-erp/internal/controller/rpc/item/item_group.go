package item

import (
	"context"

	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// GroupController ItemGroupController 物品分组服务控制器
type GroupController struct {
	item.UnimplementedItemGroupServiceServer
}

// GetItemGroupList 获取物品分组列表
// 参数：ctx 上下文，req 获取物品分组列表请求
// 返回：响应信息和错误
func (c *GroupController) GetItemGroupList(ctx context.Context, req *item.GetItemGroupListReq) (*api.ResponseInfo, error) {
	// 参数验证
	if req == nil {
		return rpc.ApiError("请求参数不能为空"), nil
	}

	// 调用服务层获取数据
	dataList, err := service.ItemGroup().GetItemGroupList(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取物品分组列表成功", dataList), nil
}

// GetItemGroup 根据分组代码获取单个物品分组信息
// 参数：ctx 上下文，req 获取物品分组请求
// 返回：响应信息和错误
func (c *GroupController) GetItemGroup(ctx context.Context, req *item.GetItemGroupReq) (*api.ResponseInfo, error) {
	// 参数验证
	if req == nil {
		return rpc.ApiError("请求参数不能为空"), nil
	}
	if req.GroupCode == "" {
		return rpc.ApiError("分组代码不能为空"), nil
	}

	// 调用服务层获取数据
	itemGroupInfo, err := service.ItemGroup().GetItemGroup(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 转换为 protobuf 类型
	pbItemGroupInfo := &item.ItemGroupInfo{
		ItemGroupName:   itemGroupInfo.ItemGroupName,
		ParentItemGroup: itemGroupInfo.ParentItemGroup,
		IsGroup:         itemGroupInfo.IsGroup,
		Branch:          itemGroupInfo.Branch,
		CompanyAbbr:     "", // 需要通过公司名称反查简称
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取物品分组成功", pbItemGroupInfo), nil
}

// SaveItemGroup 保存物品分组信息
// 参数：ctx 上下文，req 保存物品分组请求
// 返回：响应信息和错误
func (c *GroupController) SaveItemGroup(ctx context.Context, req *item.SaveItemGroupReq) (*api.ResponseInfo, error) {
	// 参数验证
	if req == nil || req.ItemGroupInfo == nil {
		return rpc.ApiError("请求参数不能为空"), nil
	}
	if req.ItemGroupInfo.ItemGroupName == "" {
		return rpc.ApiError("分组名称不能为空"), nil
	}

	// 调用服务层保存数据
	savedItemGroup, err := service.ItemGroup().SaveItemGroup(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 转换为 protobuf 类型
	pbSavedItemGroup := &item.ItemGroupInfo{
		ItemGroupName:   savedItemGroup.ItemGroupName,
		ParentItemGroup: savedItemGroup.ParentItemGroup,
		IsGroup:         savedItemGroup.IsGroup,
		Branch:          savedItemGroup.Branch,
		CompanyAbbr:     req.ItemGroupInfo.CompanyAbbr,
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("保存物品分组成功", pbSavedItemGroup), nil
}

// DeleteItemGroup 删除物品分组
// 参数：ctx 上下文，req 删除物品分组请求
// 返回：响应信息和错误
func (c *GroupController) DeleteItemGroup(ctx context.Context, req *item.DeleteItemGroupReq) (*api.ResponseInfo, error) {
	// 参数验证
	if req == nil {
		return rpc.ApiError("请求参数不能为空"), nil
	}
	if req.ItemGroupName == "" {
		return rpc.ApiError("分组名称不能为空"), nil
	}

	// 调用服务层删除数据
	err := service.ItemGroup().DeleteItemGroup(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccess("删除物品分组成功"), nil
}

func (*GroupController) SaveAttributeGroup(ctx context.Context, req *item.SaveAttributeGroupReq) (*api.ResponseInfo, error) {
	// 参数验证
	if req == nil || req.AttributeGroupInfo == nil {
		return rpc.ApiError("请求参数不能为空"), nil
	}

	resp, err := service.ItemGroup().SaveAttributeGroup(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}
	return rpc.ApiSuccessWithData("保存物品属性分组成功", resp), nil
}

func (*GroupController) DeleteAttributeGroup(context.Context, *item.DeleteAttributeGroupReq) (*api.ResponseInfo, error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*GroupController) SaveAddonGroup(context.Context, *item.SaveAddonGroupReq) (*api.ResponseInfo, error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
