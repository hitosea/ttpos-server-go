// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/crm"
)

type (
	ICrm interface {
		// GetAddressList 获取地址列表
		// 根据查询条件过滤并返回地址信息列表
		GetAddressList(ctx context.Context, req *crm.GetAddressListReq) (res *crm.GetAddressListResp, err error)
		// GetAddress 获取地址详情
		// 根据地址名称获取详细信息
		GetAddress(ctx context.Context, req *crm.GetAddressReq) (res *crm.GetAddressResp, err error)
		// CreateAddress 创建地址
		// 创建新的地址记录
		CreateAddress(ctx context.Context, req *crm.CreateAddressReq) (res *crm.CreateAddressResp, err error)
		// UpdateAddress 更新地址
		// 更新现有地址的信息
		UpdateAddress(ctx context.Context, req *crm.UpdateAddressReq) (res *crm.UpdateAddressResp, err error)
		// DeleteAddress 删除地址
		// 删除指定的地址记录
		DeleteAddress(ctx context.Context, req *crm.DeleteAddressReq) (res *crm.DeleteAddressResp, err error)
		// GetContactList 获取联系人列表
		// 根据查询条件过滤并返回联系人信息列表
		GetContactList(ctx context.Context, req *crm.GetContactListReq) (res *crm.GetContactListResp, err error)
		// GetContact 获取联系人详情
		// 根据联系人名称获取详细信息
		GetContact(ctx context.Context, req *crm.GetContactReq) (res *crm.GetContactResp, err error)
		// CreateContact 创建联系人
		// 创建新的联系人记录
		CreateContact(ctx context.Context, req *crm.CreateContactReq) (res *crm.CreateContactResp, err error)
		// UpdateContact 更新联系人
		// 更新现有联系人的信息
		UpdateContact(ctx context.Context, req *crm.UpdateContactReq) (res *crm.UpdateContactResp, err error)
		// DeleteContact 删除联系人
		// 删除指定的联系人记录
		DeleteContact(ctx context.Context, req *crm.DeleteContactReq) (res *crm.DeleteContactResp, err error)
	}
)

var (
	localCrm ICrm
)

func Crm() ICrm {
	if localCrm == nil {
		panic("implement not found for interface ICrm, forgot register?")
	}
	return localCrm
}

func RegisterCrm(i ICrm) {
	localCrm = i
}
