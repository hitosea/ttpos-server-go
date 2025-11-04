package req

import (
	"fmt"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
)

// TransferOrderListReq 调拨单列表请求
type TransferOrderListReq struct {
	dto.PageReq                // 分页参数
	OrderNo             string `json:"order_no" form:"order_no" binding:"omitempty,max=50"`                      // 单据编号
	TransferType        int    `json:"transfer_type" form:"transfer_type" binding:"omitempty,min=0,max=2"`       // 调拨类型: 1-调入 2-调出
	StatusIn            []int  `json:"status_in" form:"status_in" binding:"omitempty"`                           // 状态筛选: 0-待提交 1-待审核 2-已驳回 3-待收货 4-已完成
	OutWarehouseErpCode string `json:"out_warehouse_erp_code" form:"out_warehouse_erp_code" binding:"omitempty"` // 出库仓库ERP编码
	InWarehouseErpCode  string `json:"in_warehouse_erp_code" form:"in_warehouse_erp_code" binding:"omitempty"`   // 入库仓库ERP编码
	CreateTimeStart     int    `json:"create_time_start" form:"create_time_start" binding:"omitempty,min=0"`     // 创建时间开始
	CreateTimeEnd       int    `json:"create_time_end" form:"create_time_end" binding:"omitempty,min=0"`         // 创建时间结束
	OrderTimeStart      int    `json:"order_time_start" form:"order_time_start" binding:"omitempty,min=0"`       // 单据时间开始
	OrderTimeEnd        int    `json:"order_time_end" form:"order_time_end" binding:"omitempty,min=0"`           // 单据时间结束
}

// TransferOrderCreateReq 创建调拨单请求
type TransferOrderCreateReq struct {
	OrderTime           int64                        `json:"order_time" binding:"required,min=0"`             // 单据日期
	TransferType        int                          `json:"transfer_type" binding:"required,min=1,max=2"`    // 调拨类型: 1-调入 2-调出
	SenderCompanyUuid   uint64                       `json:"sender_company_uuid" binding:"required,min=1"`    // 发货门店UUID
	ReceiverCompanyUuid uint64                       `json:"receiver_company_uuid" binding:"required,min=1"`  // 收货门店UUID
	OutWarehouseErpCode string                       `json:"out_warehouse_erp_code" binding:"required,min=1"` // 出库仓库ERP编码
	InWarehouseErpCode  string                       `json:"in_warehouse_erp_code" binding:"required,min=1"`  // 入库仓库ERP编码
	Remark              string                       `json:"remark" binding:"omitempty,max=500"`              // 备注
	Items               []TransferOrderItemCreateReq `json:"items" binding:"required,min=1,max=500,dive"`     // 调拨明细
}

func (r *TransferOrderCreateReq) Validate() error {
	if r.SenderCompanyUuid == 0 {
		return errors.New("发货门店不能为空")
	}
	if r.ReceiverCompanyUuid == 0 {
		return errors.New("收货门店不能为空")
	}
	if r.SenderCompanyUuid == r.ReceiverCompanyUuid {
		return errors.New("发货门店和收货门店不能相同")
	}
	if r.OutWarehouseErpCode == "" {
		return errors.New("出库仓库不能为空")
	}
	if r.InWarehouseErpCode == "" {
		return errors.New("入库仓库不能为空")
	}
	if len(r.Items) == 0 {
		return errors.New("调拨明细不能为空")
	}
	for i, item := range r.Items {
		if item.MaterialUuid == 0 {
			return errors.New(fmt.Sprintf("第%d项物品不能为空", i+1))
		}
		if len(item.Units) == 0 {
			return errors.New(fmt.Sprintf("第%d项物品单位不能为空", i+1))
		}
	}
	return nil
}

// TransferOrderItemCreateReq 调拨单明细创建请求
type TransferOrderItemCreateReq struct {
	MaterialUuid uint64                           `json:"material_uuid" binding:"required,min=1"` // 物品UUID
	Units        []TransferOrderItemUnitCreateReq `json:"units" binding:"required,min=1,dive"`    // 单位列表
}

// TransferOrderItemUnitCreateReq 调拨单明细单位创建请求
type TransferOrderItemUnitCreateReq struct {
	UnitUuid uint64  `json:"unit_uuid" binding:"required,min=1"` // 单位UUID
	Num      float64 `json:"num" binding:"required,gt=0"`        // 调拨数量
}

// TransferOrderUpdateReq 更新调拨单请求（待提交状态下可更新）
type TransferOrderUpdateReq struct {
	Uuid                uint64                       `json:"uuid" binding:"required,min=1"`                   // 调拨单UUID
	OrderTime           int64                        `json:"order_time" binding:"omitempty,min=0"`            // 单据日期
	SenderCompanyUuid   uint64                       `json:"sender_company_uuid" binding:"omitempty,min=1"`   // 发货门店UUID
	ReceiverCompanyUuid uint64                       `json:"receiver_company_uuid" binding:"omitempty,min=1"` // 收货门店UUID
	OutWarehouseErpCode string                       `json:"out_warehouse_erp_code" binding:"omitempty"`      // 出库仓库ERP编码
	InWarehouseErpCode  string                       `json:"in_warehouse_erp_code" binding:"omitempty"`       // 入库仓库ERP编码
	Remark              string                       `json:"remark" binding:"omitempty,max=500"`              // 备注
	Items               []TransferOrderItemUpdateReq `json:"items" binding:"omitempty,min=1,max=500,dive"`    // 调拨明细
}

func (r *TransferOrderUpdateReq) Validate() error {
	if r.Uuid == 0 {
		return errors.New("调拨单UUID不能为空")
	}
	if r.SenderCompanyUuid != 0 && r.ReceiverCompanyUuid != 0 && r.SenderCompanyUuid == r.ReceiverCompanyUuid {
		return errors.New("发货门店和收货门店不能相同")
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

// TransferOrderItemUpdateReq 调拨单明细更新请求
type TransferOrderItemUpdateReq struct {
	MaterialUuid uint64                           `json:"material_uuid" binding:"required,min=1"` // 物品UUID
	Units        []TransferOrderItemUnitCreateReq `json:"units" binding:"required,min=1,dive"`    // 单位列表
}

// TransferOrderDetailReq 调拨单详情请求
type TransferOrderDetailReq struct {
	Uuid uint64 `json:"uuid" form:"uuid" binding:"required,min=1"` // 调拨单UUID
}

// TransferOrderSubmitReq 提交调拨单请求
type TransferOrderSubmitReq struct {
	Uuid uint64 `json:"uuid" binding:"required,min=1"` // 调拨单UUID
}

func (r *TransferOrderSubmitReq) Validate() error {
	if r.Uuid == 0 {
		return errors.New("调拨单UUID不能为空")
	}
	return nil
}

// TransferOrderApproveReq 审批调拨单请求（通过）
type TransferOrderApproveReq struct {
	Uuid   uint64 `json:"uuid" binding:"required,min=1"`      // 调拨单UUID
	Remark string `json:"remark" binding:"omitempty,max=500"` // 审批备注
}

func (r *TransferOrderApproveReq) Validate() error {
	if r.Uuid == 0 {
		return errors.New("调拨单UUID不能为空")
	}
	return nil
}

// TransferOrderRejectReq 驳回调拨单请求
type TransferOrderRejectReq struct {
	Uuid         uint64 `json:"uuid" binding:"required,min=1"`                  // 调拨单UUID
	RejectReason string `json:"reject_reason" binding:"required,min=1,max=500"` // 驳回原因
}

func (r *TransferOrderRejectReq) Validate() error {
	if r.Uuid == 0 {
		return errors.New("调拨单UUID不能为空")
	}
	if r.RejectReason == "" {
		return errors.New("驳回原因不能为空")
	}
	return nil
}

// TransferOrderReceiveReq 收货调拨单请求
type TransferOrderReceiveReq struct {
	Uuid   uint64 `json:"uuid" binding:"required,min=1"`      // 调拨单UUID
	Remark string `json:"remark" binding:"omitempty,max=500"` // 收货备注
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
