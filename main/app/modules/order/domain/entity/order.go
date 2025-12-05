package entity

import (
	"time"

	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/order/domain/valueobject"
)

// Order 订单聚合根
type Order struct {
	uuid         uint64                   // 订单UUID
	orderNo      string                   // 订单号
	companyUuid  uint64                   // 商家UUID
	customerUuid uint64                   // 客户UUID（可选）
	deskUuid     uint64                   // 桌台UUID（可选）
	status       valueobject.OrderStatus  // 订单状态
	items        []*valueobject.OrderItem // 订单项列表
	discounts    []*valueobject.Discount  // 订单优惠列表
	remark       string                   // 订单备注
	createTime   time.Time                // 创建时间
	updateTime   time.Time                // 更新时间

	// 链式调用错误收集
	err error
}

// ================ 创建方法（Builder 入口）================

// NewOrder 创建新订单（链式调用入口）
func NewOrder(companyUuid uint64) *Order {
	now := time.Now()
	return &Order{
		companyUuid: companyUuid,
		status:      valueobject.StatusPending,
		items:       make([]*valueobject.OrderItem, 0),
		discounts:   make([]*valueobject.Discount, 0),
		createTime:  now,
		updateTime:  now,
	}
}

// ================ 链式设置方法（Fluent Interface）================

// WithOrderNo 设置订单号
func (o *Order) WithOrderNo(orderNo string) *Order {
	if o.err != nil {
		return o
	}
	if orderNo == "" {
		o.err = errors.New("订单号不能为空")
		return o
	}
	o.orderNo = orderNo
	o.updateTime = time.Now()
	return o
}

// WithCustomer 设置客户
func (o *Order) WithCustomer(customerUuid uint64) *Order {
	if o.err != nil {
		return o
	}
	o.customerUuid = customerUuid
	o.updateTime = time.Now()
	return o
}

// WithDesk 设置桌台
func (o *Order) WithDesk(deskUuid uint64) *Order {
	if o.err != nil {
		return o
	}
	o.deskUuid = deskUuid
	o.updateTime = time.Now()
	return o
}

// WithRemark 设置备注
func (o *Order) WithRemark(remark string) *Order {
	if o.err != nil {
		return o
	}
	o.remark = remark
	o.updateTime = time.Now()
	return o
}

// AddItem 添加商品
func (o *Order) AddItem(productUuid uint64, productName string, quantity int, unitPrice float64) *Order {
	if o.err != nil {
		return o
	}

	if !o.status.CanAddItem() {
		o.err = errors.New("当前订单状态不允许添加商品")
		return o
	}

	item, err := valueobject.NewOrderItem(productUuid, productName, quantity, unitPrice)
	if err != nil {
		o.err = err
		return o
	}

	o.items = append(o.items, item)
	o.updateTime = time.Now()
	return o
}

// AddItemWithRemark 添加商品（带备注）
func (o *Order) AddItemWithRemark(productUuid uint64, productName string, quantity int, unitPrice float64, remark string) *Order {
	if o.err != nil {
		return o
	}

	if !o.status.CanAddItem() {
		o.err = errors.New("当前订单状态不允许添加商品")
		return o
	}

	item, err := valueobject.NewOrderItem(productUuid, productName, quantity, unitPrice)
	if err != nil {
		o.err = err
		return o
	}

	o.items = append(o.items, item.WithRemark(remark))
	o.updateTime = time.Now()
	return o
}

// ApplyPercentDiscount 应用百分比折扣
func (o *Order) ApplyPercentDiscount(percent float64, reason string) *Order {
	if o.err != nil {
		return o
	}

	if !o.status.CanApplyDiscount() {
		o.err = errors.New("当前订单状态不允许应用优惠")
		return o
	}

	discount, err := valueobject.NewPercentDiscount(percent, reason)
	if err != nil {
		o.err = err
		return o
	}

	o.discounts = append(o.discounts, discount)
	o.updateTime = time.Now()
	return o
}

// ApplyFixedDiscount 应用固定金额折扣
func (o *Order) ApplyFixedDiscount(amount float64, reason string) *Order {
	if o.err != nil {
		return o
	}

	if !o.status.CanApplyDiscount() {
		o.err = errors.New("当前订单状态不允许应用优惠")
		return o
	}

	discount, err := valueobject.NewFixedDiscount(amount, reason)
	if err != nil {
		o.err = err
		return o
	}

	o.discounts = append(o.discounts, discount)
	o.updateTime = time.Now()
	return o
}

// Build 完成构建并验证
func (o *Order) Build() (*Order, error) {
	if o.err != nil {
		return nil, o.err
	}

	// 基本验证
	if o.companyUuid == 0 {
		return nil, errors.New("商家UUID不能为空")
	}

	return o, nil
}

// ================ 状态变更方法 ================

// Confirm 确认订单
func (o *Order) Confirm() error {
	if o.status != valueobject.StatusPending {
		return errors.New("只有待处理的订单才能确认")
	}
	if len(o.items) == 0 {
		return errors.New("订单必须包含至少一个商品")
	}
	o.status = valueobject.StatusConfirmed
	o.updateTime = time.Now()
	return nil
}

// StartPreparing 开始制作
func (o *Order) StartPreparing() error {
	if o.status != valueobject.StatusConfirmed {
		return errors.New("只有已确认的订单才能开始制作")
	}
	o.status = valueobject.StatusPreparing
	o.updateTime = time.Now()
	return nil
}

// Complete 完成订单
func (o *Order) Complete() error {
	if o.status != valueobject.StatusPreparing {
		return errors.New("只有制作中的订单才能完成")
	}
	o.status = valueobject.StatusCompleted
	o.updateTime = time.Now()
	return nil
}

// Cancel 取消订单
func (o *Order) Cancel() error {
	if !o.status.CanCancel() {
		return errors.New("当前订单状态不允许取消")
	}
	o.status = valueobject.StatusCancelled
	o.updateTime = time.Now()
	return nil
}

// ================ Getter 方法 ================

// Uuid 获取订单UUID
func (o *Order) Uuid() uint64 { return o.uuid }

// OrderNo 获取订单号
func (o *Order) OrderNo() string { return o.orderNo }

// CompanyUuid 获取商家UUID
func (o *Order) CompanyUuid() uint64 { return o.companyUuid }

// CustomerUuid 获取客户UUID
func (o *Order) CustomerUuid() uint64 { return o.customerUuid }

// DeskUuid 获取桌台UUID
func (o *Order) DeskUuid() uint64 { return o.deskUuid }

// Status 获取订单状态
func (o *Order) Status() valueobject.OrderStatus { return o.status }

// Items 获取订单项列表
func (o *Order) Items() []*valueobject.OrderItem { return o.items }

// Discounts 获取订单优惠列表
func (o *Order) Discounts() []*valueobject.Discount { return o.discounts }

// Remark 获取订单备注
func (o *Order) Remark() string { return o.remark }

// CreateTime 获取创建时间
func (o *Order) CreateTime() time.Time { return o.createTime }

// UpdateTime 获取更新时间
func (o *Order) UpdateTime() time.Time { return o.updateTime }

// ItemCount 获取商品数量
func (o *Order) ItemCount() int { return len(o.items) }

// ================ 计算方法 ================

// SubTotal 商品小计（不含优惠）
func (o *Order) SubTotal() float64 {
	var total float64
	for _, item := range o.items {
		total += item.SubTotal()
	}
	return total
}

// TotalDiscount 总优惠金额
func (o *Order) TotalDiscount() float64 {
	subTotal := o.SubTotal()
	var totalDiscount float64

	for _, discount := range o.discounts {
		totalDiscount += discount.Calculate(subTotal)
	}

	// 优惠金额不能超过小计
	if totalDiscount > subTotal {
		return subTotal
	}
	return totalDiscount
}

// Total 订单总金额（小计 - 优惠）
func (o *Order) Total() float64 {
	total := o.SubTotal() - o.TotalDiscount()
	if total < 0 {
		return 0
	}
	return total
}

// ================ 辅助方法 ================

// SetUuid 设置UUID（仓储层使用）
func (o *Order) SetUuid(uuid uint64) {
	o.uuid = uuid
}

// ReconstructOrder 重建订单（从数据库加载）
func ReconstructOrder(
	uuid uint64,
	orderNo string,
	companyUuid uint64,
	customerUuid uint64,
	deskUuid uint64,
	status valueobject.OrderStatus,
	remark string,
	createTime, updateTime int64,
) *Order {
	return &Order{
		uuid:         uuid,
		orderNo:      orderNo,
		companyUuid:  companyUuid,
		customerUuid: customerUuid,
		deskUuid:     deskUuid,
		status:       status,
		items:        make([]*valueobject.OrderItem, 0),
		discounts:    make([]*valueobject.Discount, 0),
		remark:       remark,
		createTime:   time.Unix(createTime, 0),
		updateTime:   time.Unix(updateTime, 0),
	}
}

// AddReconstructedItem 添加重建的订单项（仓储层使用）
func (o *Order) AddReconstructedItem(item *valueobject.OrderItem) {
	o.items = append(o.items, item)
}

// AddReconstructedDiscount 添加重建的优惠（仓储层使用）
func (o *Order) AddReconstructedDiscount(discount *valueobject.Discount) {
	o.discounts = append(o.discounts, discount)
}
