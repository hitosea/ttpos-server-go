package model

// H5Order 扫码订单表 `ttpos_qrcode_order`
type H5Order struct {
	BaseModel
	DeskUuid               uint64  `gorm:"column:desk_uuid;not null;default:0;comment:'桌台uuid'"`                                            // 桌台uuid
	DeskNo                 string  `gorm:"column:desk_no;type:varchar(255);not null;default:'';comment:'桌台编号'"`                             // 桌台编号
	Status                 uint    `gorm:"column:status;not null;default:0;comment:'状态, 0-未下单 1-未接单 2-已接单 3-已拒单'"`                          // 状态
	IsBuffet               uint    `gorm:"column:is_buffet;not null;default:0;comment:'是否自助餐, 0-否 1-是'"`                                    // 是否自助餐
	MemberDiscountRate     float64 `gorm:"column:member_discount_rate;type:decimal(12,2);not null;default:1;comment:'会员折扣率(0-100%)'"`       // 会员折扣率
	MemberCardDiscountRate float64 `gorm:"column:member_card_discount_rate;type:decimal(12,2);not null;default:1;comment:'会员卡折扣率(0-100%)'"` // 会员卡折扣率
	CustomDiscountRate     float64 `gorm:"column:custom_discount_rate;type:decimal(12,2);not null;default:1;comment:'自定义折扣率(0-100%)'"`      // 自定义折扣率
	ProductTotalPrice      float64 `gorm:"column:product_total_price;type:decimal(12,2);not null;default:0;comment:'商品总价'"`                 // 商品总价
	TotalAmount            float64 `gorm:"column:total_amount;type:decimal(12,2);not null;default:0;comment:'订单金额'"`                        // 订单金额
	StaffUUID              uint64  `gorm:"column:staff_uuid;not null;default:0;comment:'接单或拒单员工ID'"`                                        // 接单或拒单员工ID
	HandleTime             int64   `gorm:"column:handle_time;not null;default:0;comment:'接单或拒单时间(时间戳)'"`                                    // 接单或拒单时间
	OrderTime              int64   `gorm:"column:order_time;not null;default:0;comment:'下单时间(时间戳)'"`                                        // 下单时间

	H5OrderProducts []H5OrderProduct `gorm:"foreignKey:h5_order_uuid;references:uuid"`
}

const (
	H5OrderStatusUnpaid   uint = 0 // 未下单
	H5OrderStatusPaid     uint = 1 // 未接单
	H5OrderStatusAccepted uint = 2 // 已接单
	H5OrderStatusRejected uint = 3 // 已拒单
)

// H5OrderProduct represents the ttpos_h5_order_product table in the database
type H5OrderProduct struct {
	BaseModel
	Name                 string  `gorm:"column:name;type:varchar(255);not null;default:'';comment:'商品名称.接单和拒单后从sale_order_product表获取，不再改变'"`             // 商品名称
	Price                float64 `gorm:"column:price;type:decimal(12,2);not null;default:0;comment:'最终单价（折后价）。接单和拒单后从sale_order_product表获取，不再改变'"`       // 最终单价
	SalePrice            float64 `gorm:"column:sale_price;type:decimal(12,2);not null;default:0;comment:'销售价（折前价）。接单和拒单后从sale_order_product表获取，不再改变'"`   // 销售价
	Num                  int     `gorm:"column:num;not null;default:0;comment:'最终商品数量.接单和拒单后从sale_order_product表获取，不再改变'"`                               // 最终商品数量
	AttributeText        string  `gorm:"column:attribute_text;type:varchar(500);not null;default:'';comment:'商品属性文本。接单和拒单后从sale_order_product表获取，不再改变'"` // 商品属性文本
	Remark               string  `gorm:"colummn:remark;type:varchar(255);not null;default:'';comment:'备注。接单和拒单后从sale_order_product表获取，不再改变'"`            // 备注
	SaleOrderProductUUID uint64  `gorm:"column:sale_order_product_uuid;not null;default:0;comment:'销售订单商品uuid'"`                                         // 销售订单商品uuid
	H5OrderUuid          uint64  `gorm:"column:h5_order_uuid;not null;default:0;comment:'扫码订单uuid'"`                                                     // 扫码订单uuid
}
