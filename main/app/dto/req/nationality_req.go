package req

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
)

// NationalityCreateReq 创建国籍请求
//
// 任务: story-order-source-nationality Phase 2.6
// 需求: R2.1-R2.7
//
// @version v2.10.0
type NationalityCreateReq struct {
	LocaleName dto.LocaleResponse `json:"locale_name" binding:"required"` // 多语言名称
	Sort       int                `json:"sort"`                           // 排序
}

// Validate 验证创建请求
func (req *NationalityCreateReq) Validate() error {
	if req.LocaleName.IsNull() {
		return errors.New("多语言名称不能为空")
	}
	// 至少需要中文或英文名称
	if req.LocaleName.ZH == "" && req.LocaleName.EN == "" {
		return errors.New("至少需要提供中文或英文名称")
	}
	return nil
}

// NationalityUpdateReq 更新国籍请求
//
// 任务: story-order-source-nationality Phase 2.6
// 需求: R2.1-R2.7
//
// @version v2.10.0
type NationalityUpdateReq struct {
	Uuid       uint64             `json:"uuid" binding:"required"`        // 国籍UUID
	LocaleName dto.LocaleResponse `json:"locale_name" binding:"required"` // 多语言名称
	Sort       int                `json:"sort"`                           // 排序
	Status     int                `json:"status"`                         // 状态：1-启用；0-禁用
}

// Validate 验证更新请求
func (req *NationalityUpdateReq) Validate() error {
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

// NationalityDeleteReq 删除国籍请求
//
// 任务: story-order-source-nationality Phase 2.6
// 需求: R2.1-R2.7
//
// @version v2.10.0
type NationalityDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 国籍UUID
}
