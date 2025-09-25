package printer_model

import "ttpos-server-go/app/dto"

// Products 送厨商品列表
type Products []OrderProduct
type OrderProduct struct {
	OrderProductId        uint64               `json:"order_product_id"`    // 订单商品ID
	ProductId             uint64               `json:"product_id"`          // 商品ID
	ProductName           dto.LocaleResponse   `json:"product_name"`        // 商品名称
	ProductType           uint8                `json:"product_type"`        // 商品类型
	ProductAttr           dto.LocaleResponse   `json:"product_attr"`        // 属性, 包含规格、属性、小料
	ProductAttrList       []dto.LocaleResponse `json:"product_attr_list"`   // 属性, 包含规格、属性、小料 - 列表
	Attr                  dto.LocaleResponse   `json:"attr"`                // 商品属性
	AttrList              []dto.LocaleResponse `json:"attr_list"`           // 商品属性 - 列表
	FlavorName            dto.LocaleResponse   `json:"flavor_name"`         // 商品规格
	ProductSauceNamesList []dto.LocaleResponse `json:"product_sauce_names"` // 商品小料
	ProductPrice          float64              `json:"product_price"`       // 商品价格
	TotalNum              float64              `json:"total_num"`           // 总数量
	TotalPrice            float64              `json:"total_price"`         // 总价格
	NumType               uint                 `json:"num_type"`            // 数量计算方法, 0-整数 1-小数
	IsBuffet              bool                 `json:"is_buffet"`           // 是否自助餐
	IsWrap                bool                 `json:"is_wrap"`             // 是否打包
	IsGift                bool                 `json:"is_gift"`             // 是否赠品
	Remark                string               `json:"remark"`              // 备注
	Reason                dto.LocaleResponse   `json:"reason"`              // 退菜原因
	CustomReason          string               `json:"custom_reason"`       // 自定义退菜原因
	SubProducts           []OrderProduct       `json:"sub_products"`        // 套餐子商品
}

// ProductPrinter 商品打印机
type ProductPrinter struct {
	Uuid   uint64 `json:"uuid"`   // 商品打印机uuid
	Name   string `json:"name"`   // 商品打印机名称
	Status int    `json:"status"` // 商品打印机状态
}
