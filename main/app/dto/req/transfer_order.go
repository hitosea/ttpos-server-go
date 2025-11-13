package req

import (
	"fmt"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
)

// TransferOrderListReq 调拨单列表请求
type TransferOrderListReq struct {
	dto.PageReq                  // 分页参数
	OrderNo             string   `json:"order_no" form:"order_no"`                             // 单据编号
	StatusIn            []int    `json:"status_in" form:"status_in"`                           // 状态筛选: 0-待提交 1-待审核 2-已驳回 3-待收货 4-已完成
	OrderTimeStart      int      `json:"order_time_start" form:"order_time_start"`             // 单据时间开始
	OrderTimeEnd        int      `json:"order_time_end" form:"order_time_end"`                 // 单据时间结束
	OppositeCompanyUuid []uint64 `json:"opposite_company_uuids" form:"opposite_company_uuids"` // 对方机构UUID列表
	MyRole              []string `json:"my_role" form:"my_role"`                               // 我的角色: all-全部 sender-发货方 receiver-收货方 approver-上级审批
}

// TransferOrderCreateReq 创建调拨单请求
type TransferOrderCreateReq struct {
	OrderTime           int64                        `json:"order_time" binding:"required,min=0"`          // 单据日期
	TransferType        int                          `json:"transfer_type" binding:"required,min=1,max=2"` // 调拨类型: 1-调入 2-调出
	SenderCompanyUuid   uint64                       `json:"sender_company_uuid"`                          // 发货门店UUID
	ReceiverCompanyUuid uint64                       `json:"receiver_company_uuid"`                        // 收货门店UUID
	OutWarehouseErpCode string                       `json:"out_warehouse_erp_code"`                       // 出库仓库ERP编码
	InWarehouseErpCode  string                       `json:"in_warehouse_erp_code"`                        // 入库仓库ERP编码
	Remark              string                       `json:"remark"`                                       // 备注
	Items               []TransferOrderItemCreateReq `json:"items"`                                        // 调拨明细
}

func (r *TransferOrderCreateReq) Validate() error {
	if r.TransferType != 1 && r.TransferType != 2 {
		return errors.New("调拨类型错误")
	}
	if r.TransferType == 1 {
		if r.SenderCompanyUuid == 0 {
			return errors.New("发货门店不能为空")
		}
		if r.InWarehouseErpCode == "" {
			return errors.New("出库仓库不能为空")
		}
	} else {
		if r.ReceiverCompanyUuid == 0 {
			return errors.New("收货门店不能为空")
		}
		if r.OutWarehouseErpCode == "" {
			return errors.New("入库仓库不能为空")
		}
	}
	if len(r.Items) > 0 {
		for i, item := range r.Items {
			if item.MaterialUuid == 0 {
				return errors.New(fmt.Sprintf("第%d项物品不能为空", i+1))
			}
			if len(item.Units) == 0 {
				return errors.New(fmt.Sprintf("第%d项物品单位不能为空", i+1))
			}
		}
	}
	return nil
}

// TransferOrderItemCreateReq 调拨单明细创建请求
type TransferOrderItemCreateReq struct {
	MaterialUuid uint64                           `json:"material_uuid"` // 物品UUID
	Units        []TransferOrderItemUnitCreateReq `json:"units"`         // 单位列表
}

// TransferOrderItemUnitCreateReq 调拨单明细单位创建请求
type TransferOrderItemUnitCreateReq struct {
	UnitUuid uint64  `json:"unit_uuid"` // 单位UUID
	Num      float64 `json:"num"`       // 调拨数量
}

// TransferOrderUpdateReq 更新调拨单请求（待提交状态下可更新）
type TransferOrderUpdateReq struct {
	Uuid                uint64                       `json:"uuid"`                   // 调拨单UUID
	OrderTime           int64                        `json:"order_time"`             // 单据日期
	SenderCompanyUuid   uint64                       `json:"sender_company_uuid"`    // 发货门店UUID
	ReceiverCompanyUuid uint64                       `json:"receiver_company_uuid"`  // 收货门店UUID
	OutWarehouseErpCode string                       `json:"out_warehouse_erp_code"` // 出库仓库ERP编码
	InWarehouseErpCode  string                       `json:"in_warehouse_erp_code"`  // 入库仓库ERP编码
	Remark              string                       `json:"remark"`                 // 备注
	Items               []TransferOrderItemCreateReq `json:"items"`                  // 调拨明细
}

func (r *TransferOrderUpdateReq) Validate() error {
	if r.Uuid == 0 {
		return errors.New("调拨单UUID不能为空")
	}
	if r.SenderCompanyUuid != 0 && r.ReceiverCompanyUuid != 0 && r.SenderCompanyUuid == r.ReceiverCompanyUuid {
		return errors.New("发货门店和收货门店不能相同")
	}
	if len(r.Items) > 0 {
		if len(r.Items) > 0 {
			for i, item := range r.Items {
				if item.MaterialUuid == 0 {
					return errors.New(fmt.Sprintf("第%d项物品不能为空", i+1))
				}
				if len(item.Units) == 0 {
					return errors.New(fmt.Sprintf("第%d项物品单位不能为空", i+1))
				}
			}
		}
	}
	return nil
}

// TransferOrderDetailReq 调拨单详情请求
type TransferOrderDetailReq struct {
	Uuid uint64 `json:"uuid" form:"uuid" ` // 调拨单UUID
}

// TransferOrderSubmitReq 提交调拨单请求
type TransferOrderSubmitReq struct {
	Uuid      uint64 `json:"uuid"`       // 调拨单UUID
	IsConfirm bool   `json:"is_confirm"` // 是否确认提交（移除禁用物品）
}

func (r *TransferOrderSubmitReq) Validate() error {
	if r.Uuid == 0 {
		return errors.New("调拨单UUID不能为空")
	}
	return nil
}

// TransferOrderApproveReq 审批调拨单请求（通过）
type TransferOrderApproveReq struct {
	Uuid                uint64 `json:"uuid" binding:"required,min=1"`      // 调拨单UUID
	Remark              string `json:"remark" binding:"omitempty,max=500"` // 审批备注
	OutWarehouseErpCode string `json:"out_warehouse_erp_code"`             // 出库仓库ERP编码
	InWarehouseErpCode  string `json:"in_warehouse_erp_code"`              // 入库仓库ERP编码
	IsConfirm           bool   `json:"is_confirm"`                         // 是否确认审核
}

func (r *TransferOrderApproveReq) Validate() error {
	if r.Uuid == 0 {
		return errors.New("调拨单UUID不能为空")
	}
	return nil
}

// TransferOrderRejectReq 驳回调拨单请求
type TransferOrderRejectReq struct {
	Uuid         uint64 `json:"uuid" binding:"required,min=1"` // 调拨单UUID
	RejectReason string `json:"reject_reason"`                 // 驳回原因
	IsConfirm    bool   `json:"is_confirm"`                    // 是否确认驳回
}

func (r *TransferOrderRejectReq) Validate() error {
	if r.Uuid == 0 {
		return errors.New("调拨单UUID不能为空")
	}
	return nil
}

// TransferOrderReceiveReq 收货调拨单请求
type TransferOrderReceiveReq struct {
	Uuid               uint64 `json:"uuid" binding:"required,min=1"`      // 调拨单UUID
	Remark             string `json:"remark" binding:"omitempty,max=500"` // 收货备注
	InWarehouseErpCode string `json:"in_warehouse_erp_code"`              // 入库仓库ERP编码
}

func (r *TransferOrderReceiveReq) Validate() error {
	if r.Uuid == 0 {
		return errors.New("调拨单UUID不能为空")
	}
	return nil
}

// TransferOrderDeleteReq 删除调拨单请求
type TransferOrderDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required,min=1"` // 调拨单UUID
}

func (r *TransferOrderDeleteReq) Validate() error {
	if r.Uuid == 0 {
		return errors.New("调拨单UUID不能为空")
	}
	return nil
}

// TransferOrderApprovalListReq 调拨单审批流程列表请求
type TransferOrderApprovalListReq struct {
	TransferOrderUuid uint64 `json:"transfer_order_uuid" form:"transfer_order_uuid" binding:"required,min=1"` // 调拨单UUID
}

// TransferOrderLogListReq 调拨单操作日志列表请求
type TransferOrderLogListReq struct {
	dto.PageReq
	TransferOrderUuid uint64 `json:"transfer_order_uuid" form:"transfer_order_uuid" binding:"required,min=1"` // 调拨单UUID
}

// TransferOrderMaterialListReq 调拨单物品列表查询
type TransferOrderMaterialListReq struct {
	dto.PageReq                  // 分页参数
	Keyword             string   `form:"keyword" json:"keyword"`                               // 关键字
	Status              int      `form:"status" json:"status"`                                 // 状态，0-全部 1-启用 2-停用
	CategoryUuids       []uint64 `form:"category_uuids" json:"category_uuids"`                 // 分类UUID列表,多选时
	ReceiverCompanyUuid uint64   `form:"receiver_company_uuid" json:"receiver_company_uuid"`   // 收货门店UUID
	SenderCompanyUuid   uint64   `form:"sender_company_uuid" json:"sender_company_uuid"`       // 发货门店UUID
	OutWarehouseErpCode string   `form:"out_warehouse_erp_code" json:"out_warehouse_erp_code"` // 出库仓库ERP编码
}

func (r *TransferOrderMaterialListReq) Validate() error {
	if r.ReceiverCompanyUuid == 0 && r.SenderCompanyUuid == 0 {
		return errors.New("收货门店和发货门店不能同时为空")
	}
	if r.ReceiverCompanyUuid != 0 && r.SenderCompanyUuid != 0 {
		return errors.New("收货门店和发货门店不能同时有值")
	}
	return nil
}

// TransferOrderCompanyListReq 调拨单对方机构列表查询
type TransferOrderCompanyListReq struct {
	Keyword string `form:"keyword" json:"keyword"` // 关键字
}
