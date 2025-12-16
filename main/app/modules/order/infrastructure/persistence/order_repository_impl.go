package persistence

import (
	"fmt"
	"time"

	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/order/domain/entity"
	"ttpos-server-go/app/modules/order/domain/repository"
	"ttpos-server-go/app/modules/order/domain/valueobject"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

// ================ PO 定义 ================

// OrderPO 订单持久化对象
type OrderPO struct {
	Uuid         uint64  `gorm:"column:uuid;primaryKey"`
	OrderNo      string  `gorm:"column:order_no;type:varchar(64);uniqueIndex"`
	CompanyUuid  uint64  `gorm:"column:company_uuid;index"`
	CustomerUuid uint64  `gorm:"column:customer_uuid;index"`
	DeskUuid     uint64  `gorm:"column:desk_uuid;index"`
	Status       int     `gorm:"column:status;default:0"`
	Remark       string  `gorm:"column:remark;type:varchar(500)"`
	SubTotal     float64 `gorm:"column:sub_total;type:decimal(10,2)"`
	Discount     float64 `gorm:"column:discount;type:decimal(10,2)"`
	Total        float64 `gorm:"column:total;type:decimal(10,2)"`
	CreateTime   int64   `gorm:"column:create_time"`
	UpdateTime   int64   `gorm:"column:update_time"`
	DeleteTime   int64   `gorm:"column:delete_time;default:0"`
}

// TableName 表名
func (OrderPO) TableName() string {
	return "ttpos_order"
}

// OrderItemPO 订单项持久化对象
type OrderItemPO struct {
	Uuid        uint64  `gorm:"column:uuid;primaryKey"`
	OrderUuid   uint64  `gorm:"column:order_uuid;index"`
	ProductUuid uint64  `gorm:"column:product_uuid"`
	ProductName string  `gorm:"column:product_name;type:varchar(200)"`
	Quantity    int     `gorm:"column:quantity"`
	UnitPrice   float64 `gorm:"column:unit_price;type:decimal(10,2)"`
	Discount    float64 `gorm:"column:discount;type:decimal(10,2)"`
	Remark      string  `gorm:"column:remark;type:varchar(500)"`
	CreateTime  int64   `gorm:"column:create_time"`
}

// TableName 表名
func (OrderItemPO) TableName() string {
	return "ttpos_order_item"
}

// OrderDiscountPO 订单优惠持久化对象
type OrderDiscountPO struct {
	Uuid         uint64  `gorm:"column:uuid;primaryKey"`
	OrderUuid    uint64  `gorm:"column:order_uuid;index"`
	DiscountType int     `gorm:"column:discount_type"` // 1: 百分比, 2: 固定金额
	Value        float64 `gorm:"column:value;type:decimal(10,2)"`
	Reason       string  `gorm:"column:reason;type:varchar(200)"`
	CreateTime   int64   `gorm:"column:create_time"`
}

// TableName 表名
func (OrderDiscountPO) TableName() string {
	return "ttpos_order_discount"
}

// ================ Repository 实现 ================

// OrderRepositoryImpl 订单仓储实现
type OrderRepositoryImpl struct{}

// NewOrderRepository 创建订单仓储
func NewOrderRepository() repository.IOrderRepository {
	return &OrderRepositoryImpl{}
}

// Save 保存订单
func (r *OrderRepositoryImpl) Save(ctx context.Context, order *entity.Order) error {
	db := r.getDB(ctx)
	now := time.Now().Unix()

	// 转换为 PO
	po := &OrderPO{
		Uuid:         order.Uuid(),
		OrderNo:      order.OrderNo(),
		CompanyUuid:  order.CompanyUuid(),
		CustomerUuid: order.CustomerUuid(),
		DeskUuid:     order.DeskUuid(),
		Status:       order.Status().ToInt(),
		Remark:       order.Remark(),
		SubTotal:     order.SubTotal(),
		Discount:     order.TotalDiscount(),
		Total:        order.Total(),
		UpdateTime:   now,
	}

	if order.Uuid() == 0 {
		// 生成 UUID
		uuid, err := utils.GetID()
		if err != nil {
			return errors.WithMessage(err, "生成UUID失败")
		}
		po.Uuid = uuid
		po.CreateTime = now
		order.SetUuid(uuid)

		// 创建订单
		if err := db.Create(po).Error; err != nil {
			return errors.WithMessage(err, "创建订单失败")
		}

		// 保存订单项
		if err := r.saveOrderItems(db, po.Uuid, order.Items()); err != nil {
			return err
		}

		// 保存订单优惠
		if err := r.saveOrderDiscounts(db, po.Uuid, order.Discounts()); err != nil {
			return err
		}
	} else {
		// 更新订单
		if err := db.Model(&OrderPO{}).Where("uuid = ?", po.Uuid).Updates(map[string]any{
			"status":      po.Status,
			"remark":      po.Remark,
			"sub_total":   po.SubTotal,
			"discount":    po.Discount,
			"total":       po.Total,
			"update_time": po.UpdateTime,
		}).Error; err != nil {
			return errors.WithMessage(err, "更新订单失败")
		}
	}

	return nil
}

// saveOrderItems 保存订单项
func (r *OrderRepositoryImpl) saveOrderItems(db *gorm.DB, orderUuid uint64, items []*valueobject.OrderItem) error {
	now := time.Now().Unix()
	for _, item := range items {
		uuid, err := utils.GetID()
		if err != nil {
			return errors.WithMessage(err, "生成UUID失败")
		}

		itemPO := &OrderItemPO{
			Uuid:        uuid,
			OrderUuid:   orderUuid,
			ProductUuid: item.ProductUuid(),
			ProductName: item.ProductName(),
			Quantity:    item.Quantity(),
			UnitPrice:   item.UnitPrice(),
			Discount:    item.Discount(),
			Remark:      item.Remark(),
			CreateTime:  now,
		}

		if err := db.Create(itemPO).Error; err != nil {
			return errors.WithMessage(err, "创建订单项失败")
		}
	}
	return nil
}

// saveOrderDiscounts 保存订单优惠
func (r *OrderRepositoryImpl) saveOrderDiscounts(db *gorm.DB, orderUuid uint64, discounts []*valueobject.Discount) error {
	now := time.Now().Unix()
	for _, discount := range discounts {
		uuid, err := utils.GetID()
		if err != nil {
			return errors.WithMessage(err, "生成UUID失败")
		}

		discountPO := &OrderDiscountPO{
			Uuid:         uuid,
			OrderUuid:    orderUuid,
			DiscountType: int(discount.Type()),
			Value:        discount.Value(),
			Reason:       discount.Reason(),
			CreateTime:   now,
		}

		if err := db.Create(discountPO).Error; err != nil {
			return errors.WithMessage(err, "创建订单优惠失败")
		}
	}
	return nil
}

// FindByUuid 根据UUID查找订单
func (r *OrderRepositoryImpl) FindByUuid(ctx context.Context, uuid uint64) (*entity.Order, error) {
	db := r.getDB(ctx)

	var po OrderPO
	if err := db.Where("uuid = ? AND delete_time = 0", uuid).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err, "查询订单失败")
	}

	return r.toEntity(db, &po)
}

// FindByOrderNo 根据订单号查找订单
func (r *OrderRepositoryImpl) FindByOrderNo(ctx context.Context, orderNo string) (*entity.Order, error) {
	db := r.getDB(ctx)

	var po OrderPO
	if err := db.Where("order_no = ? AND delete_time = 0", orderNo).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err, "查询订单失败")
	}

	return r.toEntity(db, &po)
}

// Remove 删除订单（软删除）
func (r *OrderRepositoryImpl) Remove(ctx context.Context, uuid uint64) error {
	db := r.getDB(ctx)

	if err := db.Model(&OrderPO{}).Where("uuid = ?", uuid).Update("delete_time", time.Now().Unix()).Error; err != nil {
		return errors.WithMessage(err, "删除订单失败")
	}

	return nil
}

// GenerateOrderNo 生成订单号
func (r *OrderRepositoryImpl) GenerateOrderNo(ctx context.Context) (string, error) {
	// 格式: ORD + 年月日 + 6位随机数
	now := time.Now()
	uuid, err := utils.GetID()
	if err != nil {
		return "", errors.WithMessage(err, "生成订单号失败")
	}
	return fmt.Sprintf("ORD%s%06d", now.Format("20060102"), uuid%1000000), nil
}

// FindWithPagination 分页查询订单
func (r *OrderRepositoryImpl) FindWithPagination(ctx context.Context, spec *repository.OrderQuerySpec, pageNo, pageSize int) ([]*entity.Order, int64, error) {
	db := r.getDB(ctx)

	query := db.Model(&OrderPO{}).Where("delete_time = 0")

	// 应用查询条件
	if spec != nil {
		if spec.CustomerUuid != nil {
			query = query.Where("customer_uuid = ?", *spec.CustomerUuid)
		}
		if spec.DeskUuid != nil {
			query = query.Where("desk_uuid = ?", *spec.DeskUuid)
		}
		if spec.Status != nil {
			query = query.Where("status = ?", *spec.Status)
		}
		if spec.Keyword != nil && *spec.Keyword != "" {
			query = query.Where("order_no LIKE ?", "%"+*spec.Keyword+"%")
		}
	}

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.WithMessage(err, "统计订单数量失败")
	}

	// 分页查询
	var pos []OrderPO
	offset := (pageNo - 1) * pageSize
	if err := query.Order("create_time DESC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, errors.WithMessage(err, "查询订单列表失败")
	}

	// 转换为实体
	orders := make([]*entity.Order, 0, len(pos))
	for i := range pos {
		order, err := r.toEntity(db, &pos[i])
		if err != nil {
			return nil, 0, err
		}
		orders = append(orders, order)
	}

	return orders, total, nil
}

// ================ 转换方法 ================

// toEntity 将 PO 转换为实体
func (r *OrderRepositoryImpl) toEntity(db *gorm.DB, po *OrderPO) (*entity.Order, error) {
	// 重建订单
	order := entity.ReconstructOrder(
		po.Uuid,
		po.OrderNo,
		po.CompanyUuid,
		po.CustomerUuid,
		po.DeskUuid,
		valueobject.OrderStatus(po.Status),
		po.Remark,
		po.CreateTime,
		po.UpdateTime,
	)

	// 加载订单项
	var itemPOs []OrderItemPO
	if err := db.Where("order_uuid = ?", po.Uuid).Find(&itemPOs).Error; err != nil {
		return nil, errors.WithMessage(err, "查询订单项失败")
	}

	for _, itemPO := range itemPOs {
		item, err := valueobject.NewOrderItem(itemPO.ProductUuid, itemPO.ProductName, itemPO.Quantity, itemPO.UnitPrice)
		if err != nil {
			return nil, err
		}
		if itemPO.Discount > 0 {
			item = item.WithDiscount(itemPO.Discount)
		}
		if itemPO.Remark != "" {
			item = item.WithRemark(itemPO.Remark)
		}
		order.AddReconstructedItem(item)
	}

	// 加载订单优惠
	var discountPOs []OrderDiscountPO
	if err := db.Where("order_uuid = ?", po.Uuid).Find(&discountPOs).Error; err != nil {
		return nil, errors.WithMessage(err, "查询订单优惠失败")
	}

	for _, discountPO := range discountPOs {
		var discount *valueobject.Discount
		var err error
		if discountPO.DiscountType == int(valueobject.DiscountTypePercent) {
			discount, err = valueobject.NewPercentDiscount(discountPO.Value, discountPO.Reason)
		} else {
			discount, err = valueobject.NewFixedDiscount(discountPO.Value, discountPO.Reason)
		}
		if err != nil {
			return nil, err
		}
		order.AddReconstructedDiscount(discount)
	}

	return order, nil
}

// getDB 获取数据库连接
func (r *OrderRepositoryImpl) getDB(ctx context.Context) *gorm.DB {
	return ctx.GetDB()
}
