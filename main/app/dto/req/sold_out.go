package req

import (
	"fmt"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
)

// SoldOutListReq 沽清列表
type SoldOutListReq struct {
	dto.PageReq // 分页参数
}

type SoldOutItem struct {
	ProductBomUuid   uint64   `json:"product_bom_uuid" binding:"required"` // 商品规格Uuid
	IsSoldOut        *bool    `json:"is_sold_out" binding:"required"`      // 是否售罄：true-是；false-否
	UseBomCardStock  *bool    `json:"use_bom_card_stock,omitempty"`        // 是否使用成本卡库存
	IsOpenStock      *bool    `json:"is_open_stock,omitempty"`             // 是否开启可售量
	SellableQuantity *float64 `json:"sellable_quantity,omitempty"`         // 可售数量
}

type AddSoldOutReq struct {
	SoldOutData []SoldOutItem `json:"sold_out_data" binding:"required,gt=0,dive,required"` // 设置售罄数据
}

// Validate 验证请求参数
func (r *AddSoldOutReq) Validate() error {
	for i := range r.SoldOutData {
		item := &r.SoldOutData[i]
		// 如果 IsOpenStock 为 false，则设置 SellableQuantity 为无限库存常量值
		if item.IsOpenStock != nil && !*item.IsOpenStock {
			maxQuantity := float64(constant.ProductBomInfiniteStock)
			item.SellableQuantity = &maxQuantity
		}

		// 验证 SellableQuantity 最大值
		if item.SellableQuantity != nil {
			if *item.SellableQuantity > float64(constant.ProductBomInfiniteStock) {
				return errors.WithMessage(errors.New("可售数量不能超过99999999"), fmt.Sprintf("第%d项数据验证失败", i+1))
			}
		}
	}
	return nil
}

type CancelSoldOutReq struct {
	ProductBomUuid uint64 `json:"product_bom_uuid" binding:"required"` // 商品规格Uuid
}

// GetSoldOutSettingsReq 获取沽清设置请求
type GetSoldOutSettingsReq struct {
	ProductPackageUuid uint64 `json:"product_package_uuid" binding:"required" form:"product_package_uuid"` // 商品包ID
}
