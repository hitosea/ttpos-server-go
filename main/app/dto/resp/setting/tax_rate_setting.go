package setting

import "ttpos-server-go/app/constant"

// TaxRate 税率管理
type TaxRate struct {
	IsOpen         string               `json:"is_open"`          // 是否开启 0关闭 1开启
	CalcType       string               `json:"calc_type"`        // 计算类型 商品已含税价-1 商品未含税价-2
	AddTaxCategory []AddTaxCategoryItem `json:"add_tax_category"` // 增值税分类
}

func (resp *TaxRate) GetTaxFeeType() uint8 {
	// 销售账单税率
	if resp.IsOpen == "1" {
		if resp.CalcType == "1" {
			return constant.TaxFeeTypeTax
		}
		if resp.CalcType == "2" {
			return constant.TaxFeeTypeNoTax
		}
	}
	return constant.TaxFeeTypeNone
}

type AddTaxCategoryItem struct {
	ID      int    `json:"id"`       // 税类id
	Name    string `json:"name"`     // 名称
	TaxRate string `json:"tax_rate"` // 税率
	Action  string `json:"action"`   // 操作结果 delete-删除 edit-编辑 add-新增
}
