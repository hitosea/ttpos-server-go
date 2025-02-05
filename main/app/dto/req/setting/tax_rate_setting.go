package setting

type TaxRate struct {
	IsOpen         string               `json:"is_open"`
	CalcType       string               `json:"calc_type"`
	AddTaxCategory []AddTaxCategoryItem `json:"add_tax_category"`
}

type AddTaxCategoryItem struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	TaxRate string `json:"tax_rate"`
	Action  string `json:"action"`
}
