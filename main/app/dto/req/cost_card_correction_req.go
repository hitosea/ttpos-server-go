package req

import (
	"errors"
	errs "ttpos-server-go/app/errors"
)

// CostCardCorrectionReq 成本卡材料消耗修正请求
type CostCardCorrectionReq struct {
	OrderUuids []uint64 `json:"order_uuids"` // 需要修正的订单UUID列表
}

// Validate 验证修正请求参数
func (r *CostCardCorrectionReq) Validate() error {
	if len(r.OrderUuids) == 0 {
		return errs.WithMessage(errors.New("订单UUID列表不能为空"))
	}
	return nil
}

// CostCardCorrectionPreviewReq 预览修正影响请求
type CostCardCorrectionPreviewReq struct {
	OrderUuids []uint64 `json:"order_uuids"` // 预览修正影响的订单UUID列表
}

// Validate 验证预览请求参数
func (r *CostCardCorrectionPreviewReq) Validate() error {
	if len(r.OrderUuids) == 0 {
		return errs.WithMessage(errors.New("订单UUID列表不能为空"))
	}
	return nil
}

// CostCardCorrectionLogsReq 查询修正日志请求
type CostCardCorrectionLogsReq struct {
	CorrectionUuid uint64 `form:"correction_uuid"`                                  // 修正操作UUID（可选）
	OrderUuid      uint64 `form:"order_uuid"`                                       // 订单UUID（可选）
	PageNo         int    `form:"page_no,default=1" json:"page_no,default=1"`       // 页码
	PageSize       int    `form:"page_size,default=20" json:"page_size,default=20"` // 每页大小
}

// Validate 验证查询日志请求参数
func (r *CostCardCorrectionLogsReq) Validate() error {
	if r.PageNo < 1 {
		return errs.WithMessage(errors.New("页码必须大于0"))
	}
	if r.PageSize < 1 {
		return errs.WithMessage(errors.New("每页大小必须大于0"))
	}
	if r.PageSize > 100 {
		return errs.WithMessage(errors.New("每页大小不能超过100"))
	}
	return nil
}
