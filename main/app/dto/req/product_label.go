package req

import "ttpos-server-go/app/errors"

// ProductLabelListReq 商品标签列表查询请求
type ProductLabelListReq struct {
}

// ProductLabelAddReq 商品标签新增请求
type ProductLabelAddReq struct {
	Name                string   `json:"name"`                  // 标签名称
	Style               string   `json:"style"`                 // 标签样式
	IsShowCashier       uint     `json:"is_show_cashier"`       // 是否在收银机显示, 0-否 1-是
	IsShowTablet        uint     `json:"is_show_tablet"`        // 是否在平板显示, 0-否 1-是
	IsShowAssistant     uint     `json:"is_show_assistant"`     // 是否在助手显示, 0-否 1-是
	IsShowH5            uint     `json:"is_show_h5"`            // 是否在H5显示, 0-否 1-是
	IsShowDelivery      uint     `json:"is_show_delivery"`      // 是否在外送显示, 0-否 1-是
	IsShowMenu          uint     `json:"is_show_menu"`          // 是否在电子菜单显示, 0-否 1-是
	IsShowKiosk         uint     `json:"is_show_kiosk"`         // 是否在自助点餐机显示, 0-否 1-是
	ProductPackageUuids []uint64 `json:"product_package_uuids"` // 关联商品包UUID列表
}

func (r *ProductLabelAddReq) Validate() error {
	if r.Name == "" {
		return errors.New("标签名称不能为空")
	}
	return nil
}

// ProductLabelEditReq 商品标签编辑请求
type ProductLabelEditReq struct {
	Uuid                uint64   `json:"uuid"`                  // 标签UUID
	Name                string   `json:"name"`                  // 标签名称
	Style               string   `json:"style"`                 // 标签样式
	IsShowCashier       uint     `json:"is_show_cashier"`       // 是否在收银机显示, 0-否 1-是
	IsShowTablet        uint     `json:"is_show_tablet"`        // 是否在平板显示, 0-否 1-是
	IsShowAssistant     uint     `json:"is_show_assistant"`     // 是否在助手显示, 0-否 1-是
	IsShowH5            uint     `json:"is_show_h5"`            // 是否在H5显示, 0-否 1-是
	IsShowDelivery      uint     `json:"is_show_delivery"`      // 是否在外送显示, 0-否 1-是
	IsShowMenu          uint     `json:"is_show_menu"`          // 是否在电子菜单显示, 0-否 1-是
	IsShowKiosk         uint     `json:"is_show_kiosk"`         // 是否在自助点餐机显示, 0-否 1-是
	ProductPackageUuids []uint64 `json:"product_package_uuids"` // 关联商品包UUID列表
}

func (r *ProductLabelEditReq) Validate() error {
	if r.Uuid == 0 {
		return errors.New("标签UUID不能为空")
	}
	if r.Name == "" {
		return errors.New("标签名称不能为空")
	}
	return nil
}

// ProductLabelDeleteReq 商品标签删除请求
type ProductLabelDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 标签UUID
}
