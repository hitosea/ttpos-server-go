package req

type OpenDeskRequest struct {
	IsBuffet               int                  `json:"is_buffet"`            //	是否是餐桌
	BuffetIds              []uint64             `json:"buffet_ids"`           //	自助餐ID
	BuffetCustomerTypeList []BuffetCustomerType `json:"buffet_customer_type"` // 自助餐顾客类型
	MealNum                uint                 `json:"meal_num"`             // 顾客数量
}

// AddProductRemarkRequest 添加商品备注
type AddProductRemarkRequest struct {
	Remark               string `json:"remark"`           // 备注
	SaleOrderProductUuid uint64 `json:"order_product_id"` // 商品ID
}

type BuffetCustomerType struct { // 自助餐顾客类型
	NameText       string `json:"name_text"`        //	自助餐顾客类型名称
	CustomerTypeId uint64 `json:"customer_type_id"` // 自助餐顾客类型ID
	Num            uint   `json:"num"`              //	顾客数量
}
