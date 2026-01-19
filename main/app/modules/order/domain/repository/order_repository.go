package repository

import (
	"ttpos-server-go/app/modules/order/domain/entity"
	"ttpos-server-go/pkg/context"
)

// IOrderRepository 订单仓储接口
type IOrderRepository interface {
	// Save 保存订单（创建或更新）
	Save(ctx context.Context, order *entity.Order) error

	// FindByUuid 根据UUID查找订单
	FindByUuid(ctx context.Context, uuid uint64) (*entity.Order, error)

	// FindByOrderNo 根据订单号查找订单
	FindByOrderNo(ctx context.Context, orderNo string) (*entity.Order, error)

	// Remove 删除订单（软删除）
	Remove(ctx context.Context, uuid uint64) error

	// GenerateOrderNo 生成订单号
	GenerateOrderNo(ctx context.Context) (string, error)

	// FindWithPagination 分页查询订单
	FindWithPagination(ctx context.Context, spec *OrderQuerySpec, pageNo, pageSize int) ([]*entity.Order, int64, error)
}

// OrderQuerySpec 订单查询规格
type OrderQuerySpec struct {
	CustomerUuid *uint64 // 客户UUID
	DeskUuid     *uint64 // 桌台UUID
	Status       *int    // 订单状态
	Keyword      *string // 关键字搜索
}

// NewOrderQuerySpec 创建查询规格
func NewOrderQuerySpec() *OrderQuerySpec {
	return &OrderQuerySpec{}
}

// WithCustomerUuid 设置客户UUID
func (s *OrderQuerySpec) WithCustomerUuid(customerUuid uint64) *OrderQuerySpec {
	s.CustomerUuid = &customerUuid
	return s
}

// WithDeskUuid 设置桌台UUID
func (s *OrderQuerySpec) WithDeskUuid(deskUuid uint64) *OrderQuerySpec {
	s.DeskUuid = &deskUuid
	return s
}

// WithStatus 设置订单状态
func (s *OrderQuerySpec) WithStatus(status int) *OrderQuerySpec {
	s.Status = &status
	return s
}

// WithKeyword 设置关键字
func (s *OrderQuerySpec) WithKeyword(keyword string) *OrderQuerySpec {
	s.Keyword = &keyword
	return s
}
