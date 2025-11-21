package req

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
)

// OrderSourceCreateReq 创建外卖来源请求
//
// 任务: story-order-source-nationality Phase 2.4
// 需求: R1.1-R1.6
//
// @version v2.10.0
type OrderSourceCreateReq struct {
	LocaleName dto.LocaleResponse `json:"locale_name" binding:"required"` // 多语言名称
	Sort       int                `json:"sort"`                           // 排序
}

// Validate 验证创建请求
func (req *OrderSourceCreateReq) Validate() error {
	if req.LocaleName.IsNull() {
		return errors.New("多语言名称不能为空")
	}
	// 至少需要中文或英文名称
	if req.LocaleName.ZH == "" && req.LocaleName.EN == "" {
		return errors.New("至少需要提供中文或英文名称")
	}
	return nil
}

// OrderSourceUpdateReq 更新外卖来源请求
//
// 任务: story-order-source-nationality Phase 2.4
// 需求: R1.1-R1.6
//
// @version v2.10.0
type OrderSourceUpdateReq struct {
	Uuid       uint64             `json:"uuid" binding:"required"`        // 外卖来源UUID
	LocaleName dto.LocaleResponse `json:"locale_name" binding:"required"` // 多语言名称
	Sort       int                `json:"sort"`                           // 排序
	Status     int                `json:"status"`                         // 状态：1-启用；0-禁用
}

// Validate 验证更新请求
func (req *OrderSourceUpdateReq) Validate() error {
	if req.Uuid == 0 {
		return errors.New("UUID不能为空")
	}
	if req.LocaleName.IsNull() {
		return errors.New("多语言名称不能为空")
	}
	// 至少需要中文或英文名称
	if req.LocaleName.ZH == "" && req.LocaleName.EN == "" {
		return errors.New("至少需要提供中文或英文名称")
	}
	return nil
}

// OrderSourceDeleteReq 删除外卖来源请求
//
// 任务: story-order-source-nationality Phase 2.4
// 需求: R1.1-R1.6
//
// @version v2.10.0
type OrderSourceDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 外卖来源UUID
}
