package model

import (
	"time"
	"ttpos-server-go/app/constant"

	"github.com/shopspring/decimal"
)

// H5Order 扫码订单表 `ttpos_h5_order`
type H5Order struct {
	BaseModel
	DeskUuid      uint64 `gorm:"column:desk_uuid;not null;default:0;comment:'桌台uuid'"`                         // 桌台uuid
	SaleOrderUuid uint64 `gorm:"column:sale_order_uuid;not null;default:0;comment:'销售订单uuid'"`                 // 销售订单uuid
	SaleBillUuid  uint64 `gorm:"column:sale_bill_uuid;not null;default:0;comment:'销售账单uuid'"`                  // 销售账单uuid
	DeskNo        string `gorm:"column:desk_no;type:varchar(255);not null;default:'';comment:'桌台编号'"`          // 桌台编号
	Status        uint   `gorm:"column:status;not null;default:0;comment:'状态, 0-未下单 1-未接单 2-已接单 3-已拒单'"`       // 状态
	IsAutoAccept  uint   `gorm:"column:is_auto_accept;not null;default:0;comment:'是否自动接单, 0-否 1-是'"`           // 是否自动接单, 0-否 1-是
	IsNeedAudit   int    `gorm:"column:is_need_audit;not null;default:0;comment:'是否需要审核，0-不需要审核，直接送厨 1-需要审核'"` // 是否需要审核，0-不需要审核，直接送厨 1-需要审核
	OrderTime     int64  `gorm:"column:order_time;not null;default:0;comment:'下单时间(时间戳)'"`                     // 下单时间

	// 拒单或接单信息
	StaffUuid  uint64 `gorm:"column:staff_uuid;not null;default:0;comment:'接单或拒单员工ID'"`     // 接单或拒单员工ID
	HandleTime int64  `gorm:"column:handle_time;not null;default:0;comment:'接单或拒单时间(时间戳)'"` // 接单或拒单时间

	// 快照信息，接单和拒单后从sale_order_product表获取，不再改变
	IsBuffet               uint    `gorm:"column:is_buffet;not null;default:0;comment:'是否自助餐, 0-否 1-是'"`                                    // 是否自助餐
	MemberDiscountRate     float64 `gorm:"column:member_discount_rate;type:decimal(12,2);not null;default:1;comment:'会员折扣率(0-100%)'"`       // 会员折扣率
	MemberCardDiscountRate float64 `gorm:"column:member_card_discount_rate;type:decimal(12,2);not null;default:1;comment:'会员卡折扣率(0-100%)'"` // 会员卡折扣率
	CustomDiscountRate     float64 `gorm:"column:custom_discount_rate;type:decimal(12,2);not null;default:1;comment:'自定义折扣率(0-100%)'"`      // 自定义折扣率
	ProductTotalPrice      float64 `gorm:"column:product_total_price;type:decimal(12,2);not null;default:0;comment:'商品总价'"`                 // 商品总价
	TotalAmount            float64 `gorm:"column:total_amount;type:decimal(12,2);not null;default:0;comment:'订单金额'"`                        // 订单金额

	// 关联对象
	H5OrderProducts   []*H5OrderProduct   `gorm:"foreignKey:h5_order_uuid;references:uuid"`
	SaleOrder         *SaleOrder          `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
	SaleOrderProducts []*SaleOrderProduct `gorm:"foreignKey:H5OrderUuid;references:uuid"`
	Desk              *Desk               `gorm:"foreignKey:DeskUuid;references:uuid"`
	Staff             *Staff              `gorm:"foreignKey:StaffUuid;references:uuid"`
}

// 设置为空
func (o *H5Order) SetNil() {
	o.SaleOrder = nil
	o.H5OrderProducts = nil
	o.SaleOrderProducts = nil
	o.Desk = nil
	o.Staff = nil
}

// 获取套餐子商品列表
func (model *H5Order) GetSubProducts(saleOrderProductUuid uint64) []*SaleOrderProduct {
	subProducts := make([]*SaleOrderProduct, 0)
	for _, saleOrderProduct := range model.SaleOrderProducts {
		if saleOrderProduct.PackageUuid == saleOrderProductUuid {
			subProducts = append(subProducts, saleOrderProduct)
		}
	}
	return subProducts
}

// 拒单
func (model *H5Order) Reject(staffUuid uint64, lang string) {
	model.Handle(staffUuid, lang, constant.H5OrderStatusRejected)
}

// 接单
func (model *H5Order) Accept(staffUuid uint64, lang string) {
	model.Handle(staffUuid, lang, constant.H5OrderStatusAccepted)
}

// 处理订单
func (model *H5Order) Handle(staffUuid uint64, lang string, status uint) {
	model.Status = status
	model.HandleTime = time.Now().Unix()
	model.StaffUuid = staffUuid
	// 快照信息
	model.IsBuffet = model.SaleOrder.SaleBill.IsBuffet
	model.MemberDiscountRate = model.SaleOrder.MemberDiscountRate
	model.MemberCardDiscountRate = model.SaleOrder.MemberCardDiscountRate
	model.CustomDiscountRate = model.SaleOrder.CustomDiscountRate
	// 设置订单商品的快照信息
	for _, h5OrderProduct := range model.H5OrderProducts {
		h5OrderProduct.Handle(lang)
	}
	// 根据订单商品快照信息计算订单总价
	model.ProductTotalPrice = model.CalcProductTotalPrice()
	model.TotalAmount = model.ProductTotalPrice
}

// 计算商品总价. 商品总价=各个商品的最终单价*最终商品数量之和
func (model *H5Order) CalcProductTotalPrice() float64 {
	totalPrice := decimal.NewFromFloat(0)
	for _, h5OrderProduct := range model.H5OrderProducts {
		totalPrice = totalPrice.Add(decimal.NewFromFloat(h5OrderProduct.Price).Mul(decimal.NewFromFloat(h5OrderProduct.Num)))
	}
	return totalPrice.Truncate(2).InexactFloat64()
}

// 删除销售订单商品
func (model *H5Order) DeleteSaleOrderProduct() {
	for _, saleOrderProduct := range model.SaleOrderProducts {
		saleOrderProduct.DeleteTime = time.Now().Unix()
	}
}

// 将已下单的h5订单商品变为已接单单的h5订单商品
func (model *H5Order) ChangeToAccepted() {
	for _, saleOrderProduct := range model.SaleOrderProducts {
		saleOrderProduct.SetAcceptOrderProduct()
	}
}

// H5OrderProduct represents the ttpos_h5_order_product table in the database
type H5OrderProduct struct {
	BaseModel

	// 快照信息，接单和拒单后从sale_order_product表获取，不再改变
	Name          string  `gorm:"column:name;type:varchar(255);comment:商品名称.接单和拒单后从sale_order_product表获取，不再改变;NOT NULL" json:"name"`
	Price         float64 `gorm:"column:price;type:decimal(12,2);default:0;comment:最终单价（折后价）。接单和拒单后从sale_order_product表获取，不再改变;NOT NULL" json:"price"`
	SalePrice     float64 `gorm:"column:sale_price;type:decimal(12,2);default:0;comment:销售价（折前价）。接单和拒单后从sale_order_product表获取，不再改变;NOT NULL" json:"sale_price"`
	Num           float64 `gorm:"column:num;type:decimal(12,2);default:0;comment:最终商品数量.接单和拒单后从sale_order_product表获取，不再改变;NOT NULL" json:"num"`
	AttributeText string  `gorm:"column:attribute_text;type:varchar(500);comment:商品属性文本。接单和拒单后从sale_order_product表获取，不再改变;NOT NULL" json:"attribute_text"`
	Remark        string  `gorm:"column:remark;type:varchar(255);comment:备注。接单和拒单后从sale_order_product表获取，不再改变;NOT NULL" json:"remark"`

	// 关联uuid
	SaleOrderProductUuid uint64 `gorm:"column:sale_order_product_uuid;type:bigint(20) unsigned;default:0;comment:销售订单商品uuid;NOT NULL" json:"sale_order_product_uuid"`
	H5OrderUuid          uint64 `gorm:"column:h5_order_uuid;type:bigint(20) unsigned;default:0;comment:扫码订单uuid;NOT NULL" json:"h5_order_uuid"`
	SaleBillUuid         uint64 `gorm:"column:sale_bill_uuid;type:bigint(20) unsigned;default:0;comment:销售账单uuid;NOT NULL" json:"sale_bill_uuid"`

	SaleOrderProduct *SaleOrderProduct `gorm:"foreignKey:SaleOrderProductUuid;references:uuid"` // 关联销售订单商品
	H5Order          *H5Order          `gorm:"foreignKey:H5OrderUuid;references:uuid"`          // 关联扫码订单
}

// 设置为空
func (o *H5OrderProduct) SetNil() {
	o.SaleOrderProduct = nil
	o.H5Order = nil
}

// 拒单或接单
func (model *H5OrderProduct) Handle(lang string) {
	if model.SaleOrderProduct.IsDelete() {
		model.SetDelete()
		return
	}
	model.Name = model.SaleOrderProduct.MultiLanguageName.GetNameByLang(lang)
	model.Price = model.SaleOrderProduct.GetFinalSalePrice() // 最终售价(单价)
	model.SalePrice = model.SaleOrderProduct.SalePrice
	model.Num = model.SaleOrderProduct.Num
	model.AttributeText = model.SaleOrderProduct.GetAttributeNamesByLang(lang)
	model.Remark = model.SaleOrderProduct.Remark
}

// 是否已接单
func (model *H5OrderProduct) IsAccepted() bool {
	if model.H5OrderUuid == 0 {
		return false
	}
	// 如果H5Order为空，则认为未接单
	if model.H5Order == nil {
		return false
	}
	return model.H5Order.Status == constant.H5OrderStatusAccepted
}
