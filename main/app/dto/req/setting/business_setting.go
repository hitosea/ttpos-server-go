package setting

type Business struct {
	ZeroingMethodList         []ZeroingMethodItem         `json:"zeroing_method_list"`
	CheckoutZeroingMethodList []CheckoutZeroingMethodItem `json:"checkout_zeroing_method_list"`
	ZeroingMethod             string                      `json:"zeroing_method"`
	CheckoutZeroingMethod     string                      `json:"checkout_zeroing_method"`
	GiftMethodList            []GiftMethodItem            `json:"gift_method_list"`
	GiftMethod                string                      `json:"gift_method"`
	FreeMethodList            []FreeMethodItem            `json:"free_method_list"`
	FreeMethod                string                      `json:"free_method"`
	DiscountMethod            string                      `json:"discount_method"`
	QrCode                    string                      `json:"qr_code"`
	NoClearTable              string                      `json:"no_clear_table"`
	IsNeedPassword            string                      `json:"is_need_password"`
	DishCardStyle             string                      `json:"dish_card_style"`
	DishCardStyleTime         string                      `json:"dish_card_style_time"`
	IsInvoice                 string                      `json:"is_invoice"`
}

type ZeroingMethodItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type CheckoutZeroingMethodItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type GiftMethodItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type FreeMethodItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}
