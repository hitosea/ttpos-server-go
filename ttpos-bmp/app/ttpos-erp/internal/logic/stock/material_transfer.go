package stock

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/material_transfer"
)

/**
 * 物料转移
 * 用在华莱士内部 调入/调出。 实际上是通过 创建内部销售单 -> 内部采购单来实现
 */

var (
	MaterialTransfer = &sMaterialTransfer{}
)

type sMaterialTransfer struct {
}

func init() {

}

func (s *sMaterialTransfer) MaterialTransfer(ctx context.Context, req *material_transfer.MaterialTransferReq) (*material_transfer.MaterialTransferResp, error) {

	//获取调出方父级公司

	//判断调入方是否自己的父级公司

	//调出方发起销售订单，目标是父级公司

	//根据调出方销售订单，创建内部采购单

	//判断调入方父级公司是否是调出方父级公司的子公司

	return nil, nil
}

// CreateInnerTransferReceipt  实际上是通过 创建内部销售单 -> 内部采购单来实现
func (s *sMaterialTransfer) CreateInnerTransferReceipt(ctx context.Context, req *material_transfer.MaterialTransferReq) (*material_transfer.TransferReceipt, error) {

	//检查调出方供应商的交易对象是否包含了父级公司，如果没有默认添加

	//检查调出方父级公司的内部客户的交易对象是否包含了调出方公司，如果没有默认添加

	//调出方发起销售订单，

	//根据调出方销售订单，创建内部采购单

	return nil, nil
}
