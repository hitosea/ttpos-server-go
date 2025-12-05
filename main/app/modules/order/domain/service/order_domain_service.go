package service

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/order/domain/entity"
	"ttpos-server-go/app/modules/order/domain/repository"
	"ttpos-server-go/pkg/context"
)

// IOrderDomainService 订单领域服务接口
type IOrderDomainService interface {
	// CreateOrder 创建订单（返回 Builder）
	CreateOrder(ctx context.Context) *OrderBuilder

	// ConfirmOrder 确认订单
	ConfirmOrder(ctx context.Context, orderUuid uint64) error

	// CancelOrder 取消订单
	CancelOrder(ctx context.Context, orderUuid uint64) error
}

// orderDomainService 订单领域服务实现
type orderDomainService struct {
	orderRepo repository.IOrderRepository
}

// NewOrderDomainService 创建订单领域服务
func NewOrderDomainService(orderRepo repository.IOrderRepository) IOrderDomainService {
	return &orderDomainService{
		orderRepo: orderRepo,
	}
}

// CreateOrder 创建订单（返回 Builder）
func (s *orderDomainService) CreateOrder(ctx context.Context) *OrderBuilder {
	return NewOrderBuilder(ctx, s.orderRepo)
}

// ConfirmOrder 确认订单
func (s *orderDomainService) ConfirmOrder(ctx context.Context, orderUuid uint64) error {
	order, err := s.orderRepo.FindByUuid(ctx, orderUuid)
	if err != nil {
		return errors.WithMessage(err, "查询订单失败")
	}
	if order == nil {
		return errors.New("订单不存在")
	}

	if err := order.Confirm(); err != nil {
		return err
	}

	return s.orderRepo.Save(ctx, order)
}

// CancelOrder 取消订单
func (s *orderDomainService) CancelOrder(ctx context.Context, orderUuid uint64) error {
	order, err := s.orderRepo.FindByUuid(ctx, orderUuid)
	if err != nil {
		return errors.WithMessage(err, "查询订单失败")
	}
	if order == nil {
		return errors.New("订单不存在")
	}

	if err := order.Cancel(); err != nil {
		return err
	}

	return s.orderRepo.Save(ctx, order)
}

// ================ OrderBuilder ================

// OrderBuilder 订单构建器
type OrderBuilder struct {
	ctx       context.Context
	orderRepo repository.IOrderRepository
	order     *entity.Order
}

// NewOrderBuilder 创建订单构建器
func NewOrderBuilder(ctx context.Context, orderRepo repository.IOrderRepository) *OrderBuilder {
	return &OrderBuilder{
		ctx:       ctx,
		orderRepo: orderRepo,
		order:     entity.NewOrder(ctx.GetCompanyUuid()),
	}
}

// WithOrderNo 设置订单号
func (b *OrderBuilder) WithOrderNo(orderNo string) *OrderBuilder {
	b.order.WithOrderNo(orderNo)
	return b
}

// WithCustomer 设置客户
func (b *OrderBuilder) WithCustomer(customerUuid uint64) *OrderBuilder {
	b.order.WithCustomer(customerUuid)
	return b
}

// WithDesk 设置桌台
func (b *OrderBuilder) WithDesk(deskUuid uint64) *OrderBuilder {
	b.order.WithDesk(deskUuid)
	return b
}

// WithRemark 设置备注
func (b *OrderBuilder) WithRemark(remark string) *OrderBuilder {
	b.order.WithRemark(remark)
	return b
}

// AddItem 添加商品
func (b *OrderBuilder) AddItem(productUuid uint64, productName string, quantity int, unitPrice float64) *OrderBuilder {
	b.order.AddItem(productUuid, productName, quantity, unitPrice)
	return b
}

// AddItemWithRemark 添加商品（带备注）
func (b *OrderBuilder) AddItemWithRemark(productUuid uint64, productName string, quantity int, unitPrice float64, remark string) *OrderBuilder {
	b.order.AddItemWithRemark(productUuid, productName, quantity, unitPrice, remark)
	return b
}

// ApplyPercentDiscount 应用百分比折扣
func (b *OrderBuilder) ApplyPercentDiscount(percent float64, reason string) *OrderBuilder {
	b.order.ApplyPercentDiscount(percent, reason)
	return b
}

// ApplyFixedDiscount 应用固定金额折扣
func (b *OrderBuilder) ApplyFixedDiscount(amount float64, reason string) *OrderBuilder {
	b.order.ApplyFixedDiscount(amount, reason)
	return b
}

// Save 保存订单
func (b *OrderBuilder) Save() (*entity.Order, error) {
	// 构建并验证
	order, err := b.order.Build()
	if err != nil {
		return nil, err
	}

	// 如果没有设置订单号，自动生成
	if order.OrderNo() == "" {
		orderNo, err := b.orderRepo.GenerateOrderNo(b.ctx)
		if err != nil {
			return nil, errors.WithMessage(err, "生成订单号失败")
		}
		order.WithOrderNo(orderNo)
	}

	// 保存订单
	if err := b.orderRepo.Save(b.ctx, order); err != nil {
		return nil, errors.WithMessage(err, "保存订单失败")
	}

	return order, nil
}

// SaveAndConfirm 保存并确认订单
func (b *OrderBuilder) SaveAndConfirm() (*entity.Order, error) {
	order, err := b.Save()
	if err != nil {
		return nil, err
	}

	if err := order.Confirm(); err != nil {
		return nil, err
	}

	if err := b.orderRepo.Save(b.ctx, order); err != nil {
		return nil, errors.WithMessage(err, "确认订单失败")
	}

	return order, nil
}
