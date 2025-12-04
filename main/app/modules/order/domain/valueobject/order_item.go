package valueobject

import "ttpos-server-go/app/errors"

// OrderItem 订单项值对象（不可变）
type OrderItem struct {
	productUuid uint64  // 商品UUID
	productName string  // 商品名称
	quantity    int     // 数量
	unitPrice   float64 // 单价
	discount    float64 // 折扣金额
	remark      string  // 备注
}

// NewOrderItem 创建订单项
func NewOrderItem(productUuid uint64, productName string, quantity int, unitPrice float64) (*OrderItem, error) {
	if productUuid == 0 {
		return nil, errors.New("商品UUID不能为空")
	}
	if productName == "" {
		return nil, errors.New("商品名称不能为空")
	}
	if quantity <= 0 {
		return nil, errors.New("数量必须大于0")
	}
	if unitPrice < 0 {
		return nil, errors.New("单价不能为负数")
	}

	return &OrderItem{
		productUuid: productUuid,
		productName: productName,
		quantity:    quantity,
		unitPrice:   unitPrice,
		discount:    0,
		remark:      "",
	}, nil
}

// ProductUuid 获取商品UUID
func (i *OrderItem) ProductUuid() uint64 {
	return i.productUuid
}

// ProductName 获取商品名称
func (i *OrderItem) ProductName() string {
	return i.productName
}

// Quantity 获取数量
func (i *OrderItem) Quantity() int {
	return i.quantity
}

// UnitPrice 获取单价
func (i *OrderItem) UnitPrice() float64 {
	return i.unitPrice
}

// Discount 获取折扣金额
func (i *OrderItem) Discount() float64 {
	return i.discount
}

// Remark 获取备注
func (i *OrderItem) Remark() string {
	return i.remark
}

// SubTotal 小计（数量 * 单价）
func (i *OrderItem) SubTotal() float64 {
	return float64(i.quantity) * i.unitPrice
}

// Total 合计（小计 - 折扣）
func (i *OrderItem) Total() float64 {
	total := i.SubTotal() - i.discount
	if total < 0 {
		return 0
	}
	return total
}

// WithDiscount 设置折扣（返回新实例）
func (i *OrderItem) WithDiscount(discount float64) *OrderItem {
	return &OrderItem{
		productUuid: i.productUuid,
		productName: i.productName,
		quantity:    i.quantity,
		unitPrice:   i.unitPrice,
		discount:    discount,
		remark:      i.remark,
	}
}

// WithRemark 设置备注（返回新实例）
func (i *OrderItem) WithRemark(remark string) *OrderItem {
	return &OrderItem{
		productUuid: i.productUuid,
		productName: i.productName,
		quantity:    i.quantity,
		unitPrice:   i.unitPrice,
		discount:    i.discount,
		remark:      remark,
	}
}

// WithQuantity 修改数量（返回新实例）
func (i *OrderItem) WithQuantity(quantity int) (*OrderItem, error) {
	if quantity <= 0 {
		return nil, errors.New("数量必须大于0")
	}
	return &OrderItem{
		productUuid: i.productUuid,
		productName: i.productName,
		quantity:    quantity,
		unitPrice:   i.unitPrice,
		discount:    i.discount,
		remark:      i.remark,
	}, nil
}
