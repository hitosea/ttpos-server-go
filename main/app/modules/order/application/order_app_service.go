package order

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/order/domain/entity"
	"ttpos-server-go/app/modules/order/domain/repository"
	"ttpos-server-go/app/modules/order/domain/service"
	"ttpos-server-go/pkg/context"
)

// OrderAppService 订单应用服务
type OrderAppService struct {
	orderRepo     repository.IOrderRepository
	domainService service.IOrderDomainService
}

// NewOrderAppService 创建订单应用服务
func NewOrderAppService(
	orderRepo repository.IOrderRepository,
	domainService service.IOrderDomainService,
) *OrderAppService {
	return &OrderAppService{
		orderRepo:     orderRepo,
		domainService: domainService,
	}
}

// ================ 创建订单（链式调用入口）================

// CreateOrder 创建订单（返回 Builder）
//
// 使用示例:
//
//	order, err := appService.CreateOrder(ctx).
//	    WithDesk(deskUuid).
//	    WithCustomer(memberUuid).
//	    AddItem(productUuid1, "红烧肉", 1, 68.00).
//	    AddItem(productUuid2, "米饭", 2, 3.00).
//	    ApplyPercentDiscount(10, "新客优惠").
//	    Save()
func (s *OrderAppService) CreateOrder(ctx context.Context) *service.OrderBuilder {
	return s.domainService.CreateOrder(ctx)
}

// ================ 订单查询 ================

// GetOrder 获取订单详情
func (s *OrderAppService) GetOrder(ctx context.Context, orderUuid uint64) (*OrderResponse, error) {
	order, err := s.orderRepo.FindByUuid(ctx, orderUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询订单失败")
	}
	if order == nil {
		return nil, errors.New("订单不存在")
	}

	return s.toResponse(order), nil
}

// GetOrderByOrderNo 根据订单号获取订单
func (s *OrderAppService) GetOrderByOrderNo(ctx context.Context, orderNo string) (*OrderResponse, error) {
	order, err := s.orderRepo.FindByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, errors.WithMessage(err, "查询订单失败")
	}
	if order == nil {
		return nil, errors.New("订单不存在")
	}

	return s.toResponse(order), nil
}

// GetOrderList 获取订单列表
func (s *OrderAppService) GetOrderList(ctx context.Context, req GetOrderListRequest) (*OrderListResponse, error) {
	spec := repository.NewOrderQuerySpec()

	if req.CustomerUuid > 0 {
		spec = spec.WithCustomerUuid(req.CustomerUuid)
	}
	if req.DeskUuid > 0 {
		spec = spec.WithDeskUuid(req.DeskUuid)
	}
	if req.Status != nil {
		spec = spec.WithStatus(*req.Status)
	}
	if req.Keyword != "" {
		spec = spec.WithKeyword(req.Keyword)
	}

	orders, total, err := s.orderRepo.FindWithPagination(ctx, spec, req.PageNo, req.PageSize)
	if err != nil {
		return nil, errors.WithMessage(err, "查询订单列表失败")
	}

	list := make([]*OrderResponse, 0, len(orders))
	for _, order := range orders {
		list = append(list, s.toResponse(order))
	}

	return &OrderListResponse{
		List:     list,
		Total:    total,
		PageNo:   req.PageNo,
		PageSize: req.PageSize,
	}, nil
}

// ================ 订单操作 ================

// ConfirmOrder 确认订单
func (s *OrderAppService) ConfirmOrder(ctx context.Context, orderUuid uint64) error {
	return s.domainService.ConfirmOrder(ctx, orderUuid)
}

// CancelOrder 取消订单
func (s *OrderAppService) CancelOrder(ctx context.Context, orderUuid uint64) error {
	return s.domainService.CancelOrder(ctx, orderUuid)
}

// ================ DTO 定义 ================

// GetOrderListRequest 获取订单列表请求
type GetOrderListRequest struct {
	CustomerUuid uint64 // 客户UUID
	DeskUuid     uint64 // 桌台UUID
	Status       *int   // 订单状态
	Keyword      string // 关键字搜索
	PageNo       int    // 页码
	PageSize     int    // 每页数量
}

// OrderResponse 订单响应
type OrderResponse struct {
	Uuid          uint64               `json:"uuid"`
	OrderNo       string               `json:"order_no"`
	CustomerUuid  uint64               `json:"customer_uuid"`
	DeskUuid      uint64               `json:"desk_uuid"`
	Status        int                  `json:"status"`
	StatusText    string               `json:"status_text"`
	Items         []*OrderItemResponse `json:"items"`
	Discounts     []*DiscountResponse  `json:"discounts"`
	Remark        string               `json:"remark"`
	SubTotal      float64              `json:"sub_total"`
	TotalDiscount float64              `json:"total_discount"`
	Total         float64              `json:"total"`
	ItemCount     int                  `json:"item_count"`
	CreateTime    int64                `json:"create_time"`
	UpdateTime    int64                `json:"update_time"`
}

// OrderItemResponse 订单项响应
type OrderItemResponse struct {
	ProductUuid uint64  `json:"product_uuid"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Discount    float64 `json:"discount"`
	SubTotal    float64 `json:"sub_total"`
	Total       float64 `json:"total"`
	Remark      string  `json:"remark"`
}

// DiscountResponse 优惠响应
type DiscountResponse struct {
	Type   int     `json:"type"`
	Value  float64 `json:"value"`
	Reason string  `json:"reason"`
}

// OrderListResponse 订单列表响应
type OrderListResponse struct {
	List     []*OrderResponse `json:"list"`
	Total    int64            `json:"total"`
	PageNo   int              `json:"page_no"`
	PageSize int              `json:"page_size"`
}

// ================ 转换方法 ================

func (s *OrderAppService) toResponse(order *entity.Order) *OrderResponse {
	items := make([]*OrderItemResponse, 0, len(order.Items()))
	for _, item := range order.Items() {
		items = append(items, &OrderItemResponse{
			ProductUuid: item.ProductUuid(),
			ProductName: item.ProductName(),
			Quantity:    item.Quantity(),
			UnitPrice:   item.UnitPrice(),
			Discount:    item.Discount(),
			SubTotal:    item.SubTotal(),
			Total:       item.Total(),
			Remark:      item.Remark(),
		})
	}

	discounts := make([]*DiscountResponse, 0, len(order.Discounts()))
	for _, discount := range order.Discounts() {
		discounts = append(discounts, &DiscountResponse{
			Type:   int(discount.Type()),
			Value:  discount.Value(),
			Reason: discount.Reason(),
		})
	}

	return &OrderResponse{
		Uuid:          order.Uuid(),
		OrderNo:       order.OrderNo(),
		CustomerUuid:  order.CustomerUuid(),
		DeskUuid:      order.DeskUuid(),
		Status:        order.Status().ToInt(),
		StatusText:    order.Status().String(),
		Items:         items,
		Discounts:     discounts,
		Remark:        order.Remark(),
		SubTotal:      order.SubTotal(),
		TotalDiscount: order.TotalDiscount(),
		Total:         order.Total(),
		ItemCount:     order.ItemCount(),
		CreateTime:    order.CreateTime().Unix(),
		UpdateTime:    order.UpdateTime().Unix(),
	}
}
