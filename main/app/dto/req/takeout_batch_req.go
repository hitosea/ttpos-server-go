package req

import (
	"ttpos-server-go/app/errors"
)

// TakeoutBatchCreateReq 批量创建外卖商品请求
type TakeoutBatchCreateReq struct {
	Platform     string   `json:"platform" binding:"required,oneof=grab lineman"` // 外卖平台标识: grab/lineman
	ProductUuids []uint64 `json:"product_uuids" binding:"required"`               // 商品UUID列表,最多100个
}

// Validate 验证批量创建请求
func (r *TakeoutBatchCreateReq) Validate() error {
	if r.Platform == "" {
		return errors.WithMessage(errors.New("平台标识不能为空"))
	}
	if r.Platform != "grab" && r.Platform != "lineman" {
		return errors.WithMessage(errors.New("平台标识必须是 grab 或 lineman"))
	}
	if len(r.ProductUuids) == 0 {
		return errors.WithMessage(errors.New("商品UUID列表不能为空"))
	}
	return nil
}

// TakeoutBatchOnlineReq 批量上架外卖商品请求
type TakeoutBatchOnlineReq struct {
	Platform     string   `json:"platform" binding:"required,oneof=grab lineman"` // 外卖平台标识: grab/lineman
	ProductUuids []uint64 `json:"product_uuids" binding:"required"`               // 商品UUID列表,最多100个
}

// Validate 验证批量上架请求
func (r *TakeoutBatchOnlineReq) Validate() error {
	if r.Platform == "" {
		return errors.WithMessage(errors.New("平台标识不能为空"))
	}
	if r.Platform != "grab" && r.Platform != "lineman" {
		return errors.WithMessage(errors.New("平台标识必须是 grab 或 lineman"))
	}
	if len(r.ProductUuids) == 0 {
		return errors.WithMessage(errors.New("商品UUID列表不能为空"))
	}
	return nil
}

// TakeoutBatchOfflineReq 批量下架外卖商品请求
type TakeoutBatchOfflineReq struct {
	Platform     string   `json:"platform" binding:"required,oneof=grab lineman"` // 外卖平台标识: grab/lineman
	ProductUuids []uint64 `json:"product_uuids" binding:"required"`               // 商品UUID列表,最多100个
}

// Validate 验证批量下架请求
func (r *TakeoutBatchOfflineReq) Validate() error {
	if r.Platform == "" {
		return errors.WithMessage(errors.New("平台标识不能为空"))
	}
	if r.Platform != "grab" && r.Platform != "lineman" {
		return errors.WithMessage(errors.New("平台标识必须是 grab 或 lineman"))
	}
	if len(r.ProductUuids) == 0 {
		return errors.WithMessage(errors.New("商品UUID列表不能为空"))
	}
	return nil
}

// TakeoutBatchDeleteReq 批量删除外卖商品请求
type TakeoutBatchDeleteReq struct {
	Platform     string   `json:"platform" binding:"required,oneof=grab lineman"` // 外卖平台标识: grab/lineman
	ProductUuids []uint64 `json:"product_uuids" binding:"required"`               // 商品UUID列表,最多100个
}

// Validate 验证批量删除请求
func (r *TakeoutBatchDeleteReq) Validate() error {
	if r.Platform == "" {
		return errors.WithMessage(errors.New("平台标识不能为空"))
	}
	if r.Platform != "grab" && r.Platform != "lineman" {
		return errors.WithMessage(errors.New("平台标识必须是 grab 或 lineman"))
	}
	if len(r.ProductUuids) == 0 {
		return errors.WithMessage(errors.New("商品UUID列表不能为空"))
	}
	return nil
}
