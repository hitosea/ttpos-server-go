// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SaleOrderBuffetCustomerType is the golang structure for table sale_order_buffet_customer_type.
type SaleOrderBuffetCustomerType struct {
	Id                          uint    `json:"id"                          orm:"id"                              description:"自增ID"`                                                                                                    // 自增ID
	Uuid                        uint64  `json:"uuid"                        orm:"uuid"                            description:"销售订单顾客类型ID"`                                                                                              // 销售订单顾客类型ID
	Name                        string  `json:"name"                        orm:"name"                            description:"顾客类型名称"`                                                                                                  // 顾客类型名称
	Num                         int     `json:"num"                         orm:"num"                             description:"人数"`                                                                                                      // 人数
	SalePrice                   float64 `json:"salePrice"                   orm:"sale_price"                      description:"原始单价（单人，折前价）。自助餐顾客类型原价,下单后价格不受后台改变"`                                                                      // 原始单价（单人，折前价）。自助餐顾客类型原价,下单后价格不受后台改变
	Price                       float64 `json:"price"                       orm:"price"                           description:"最终单价（折后价），只进行自定义打折，不进行会员打折"`                                                                              // 最终单价（折后价），只进行自定义打折，不进行会员打折
	CustomDiscountRate          float64 `json:"customDiscountRate"          orm:"custom_discount_rate"            description:"自定义折扣率, 值为0-1之间(0-100%)"`                                                                                 // 自定义折扣率, 值为0-1之间(0-100%)
	CustomDiscountFee           float64 `json:"customDiscountFee"           orm:"custom_discount_fee"             description:"自定义折扣金额（单人）。自定义折扣金额（单人）=自助餐顾客类型原价*自定义折扣率"`                                                                // 自定义折扣金额（单人）。自定义折扣金额（单人）=自助餐顾客类型原价*自定义折扣率
	TaxRate                     float64 `json:"taxRate"                     orm:"tax_rate"                        description:"税率,值为0-1之间.加购时记录税率,结账时再重新核算"`                                                                             // 税率,值为0-1之间.加购时记录税率,结账时再重新核算
	ServiceTaxFee               float64 `json:"serviceTaxFee"               orm:"service_tax_fee"                 description:"服务费税费（单人）,0-不收取税费；收取时，服务费税费=服务费*税率"`                                                                      // 服务费税费（单人）,0-不收取税费；收取时，服务费税费=服务费*税率
	TaxFee                      float64 `json:"taxFee"                      orm:"tax_fee"                         description:"自助餐顾客类型税费（单人）。自助餐顾客类型已含税时，税费=自助餐顾客类型原价*(1-1/(1+税率))；自助餐顾客类型未含税时，税费=自助餐顾客类型原价*税率"`                         // 自助餐顾客类型税费（单人）。自助餐顾客类型已含税时，税费=自助餐顾客类型原价*(1-1/(1+税率))；自助餐顾客类型未含税时，税费=自助餐顾客类型原价*税率
	ServiceFee                  float64 `json:"serviceFee"                  orm:"service_fee"                     description:"服务费（单人）,0-固定服务费 大于0-按比例收服务费；自助餐顾客类型已含税时，服务费=(自助餐顾客类型原价-自助餐顾客类型税费)*服务费比例；自助餐顾客类型未含税时，服务费=自助餐顾客类型原价*服务费比例"` // 服务费（单人）,0-固定服务费 大于0-按比例收服务费；自助餐顾客类型已含税时，服务费=(自助餐顾客类型原价-自助餐顾客类型税费)*服务费比例；自助餐顾客类型未含税时，服务费=自助餐顾客类型原价*服务费比例
	TotalPrice                  float64 `json:"totalPrice"                  orm:"total_price"                     description:"应收金额(单人)。商品已含税时，应收金额(单人)=(最终单价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=最终单价+服务费+总税费"`                              // 应收金额(单人)。商品已含税时，应收金额(单人)=(最终单价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=最终单价+服务费+总税费
	SaleOrderUuid               uint64  `json:"saleOrderUuid"               orm:"sale_order_uuid"                 description:"销售订单ID"`                                                                                                  // 销售订单ID
	BuffetPackageUuid           uint64  `json:"buffetPackageUuid"           orm:"buffet_package_uuid"             description:"自助餐套餐ID"`                                                                                                 // 自助餐套餐ID
	BuffetCustomerTypePriceUuid uint64  `json:"buffetCustomerTypePriceUuid" orm:"buffet_customer_type_price_uuid" description:"自助餐客户类型价格ID"`                                                                                             // 自助餐客户类型价格ID
	CreateTime                  uint    `json:"createTime"                  orm:"create_time"                     description:"创建时间(时间戳)"`                                                                                               // 创建时间(时间戳)
	UpdateTime                  uint    `json:"updateTime"                  orm:"update_time"                     description:"更新时间(时间戳)"`                                                                                               // 更新时间(时间戳)
	DeleteTime                  uint    `json:"deleteTime"                  orm:"delete_time"                     description:"删除时间(时间戳)"`                                                                                               // 删除时间(时间戳)
}
