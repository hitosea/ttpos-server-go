package service

import (
	"slices"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	apperrors "ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"

	"github.com/shopspring/decimal"
)

// IH5OrderSrv 定义接单服务接口
type IH5OrderSrv interface {
	GetH5OrderList(ctx context.Context, listReq req.H5OrderListReq) (*resp.H5OrderList, error)                 // 获取h5订单列表
	GetH5OrderDetail(ctx context.Context, orderUuid uint64) (*resp.H5OrderDetailResp, error)                   // 获取h5订单详情
	RejectH5Order(ctx context.Context, h5OrderUuid uint64) error                                               // 拒单
	AcceptH5Order(ctx context.Context, orderUuid uint64, isAutoOrder bool) (*resp.OrderCheckServiceRes, error) // 接单
}

type h5OrderSrv struct {
	bus      *event.SystemEventBus
	dbm      *database.DBManager
	orderSrv IOrderSrv
}

func NewH5OrderSrv(dbm *database.DBManager, orderSrv IOrderSrv) IH5OrderSrv {
	return NewH5OrderSrvImpl(dbm, orderSrv)
}

func NewH5OrderSrvImpl(dbm *database.DBManager, orderSrv IOrderSrv) IH5OrderSrv {
	return &h5OrderSrv{
		dbm:      dbm,
		orderSrv: orderSrv,
	}
}

func (s *h5OrderSrv) GetH5OrderList(ctx context.Context, listReq req.H5OrderListReq) (*resp.H5OrderList, error) {
	var listResp = resp.H5OrderList{
		List: make([]resp.H5OrderItem, 0),
	}
	h5OrderRepo := repository.NewH5OrderRepo(ctx.GetDB())
	var dbOptions []repository.DBOption
	unhandledStatusOption := h5OrderRepo.WhereStatus([]uint{constant.H5OrderStatusOrder})
	handledStatusOption := h5OrderRepo.WhereStatus([]uint{constant.H5OrderStatusAccepted, constant.H5OrderStatusRejected})
	if *listReq.Status == 0 { // 未处理
		dbOptions = append(dbOptions, unhandledStatusOption)
	} else { // 已处理：已接单、已拒单
		dbOptions = append(dbOptions, handledStatusOption)
	}
	// 桌台区域条件
	if listReq.DeskRegionUuid > 0 {
		dbOptions = append(dbOptions, h5OrderRepo.WhereDeskRegionUuid(listReq.DeskRegionUuid))
	}
	dbOptions = append(
		dbOptions,
		repository.CommonRepo.SortWithCreateTime("desc"),
		h5OrderRepo.WithDesk(),
		h5OrderRepo.WithH5OrderProducts(),
		h5OrderRepo.WithSaleOrderProducts(),
		h5OrderRepo.WithSaleOrderProductsMultiLanguageName(),
	)
	orders, total, err := h5OrderRepo.PaginateGetH5Order(listReq.PageNo, listReq.PageSize, dbOptions...)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	items := make([]resp.H5OrderItem, 0, len(orders))
	for _, order := range orders {
		var num uint
		var price float64
		if order.Status == constant.H5OrderStatusOrder { // 如果订单状态是1（待处理），读取关联的销售订单商品
			for _, product := range order.SaleOrderProducts {
				num = num + product.Num
				totalPrice := decimal.NewFromFloat(product.Price).Mul(decimal.NewFromInt(int64(product.Num))).Truncate(2).InexactFloat64()
				price = price + totalPrice
			}
			// 再加上已经接单的商品价格。同一个销售账单，已接单的，h5订单商品
			if order.SaleBillUuid > 0 {
				amount := decimal.NewFromFloat(0) // 已下单的商品金额之和
				products, err := h5OrderRepo.GetH5OrderProductsBySaleBillUuidAndAccept(order.SaleBillUuid)
				if err != nil {
					return nil, errors.WithMessage(apperrors.ErrInternal, "获取h5订单详情失败", err.Error())
				}
				for _, product := range products {
					totalPrice := decimal.NewFromFloat(product.Price).Mul(decimal.NewFromInt(int64(product.Num))).Truncate(2).InexactFloat64()
					amount = amount.Add(decimal.NewFromFloat(totalPrice))
				}
				price += amount.InexactFloat64()
			}
		} else if slices.Contains([]uint{constant.H5OrderStatusAccepted, constant.H5OrderStatusRejected}, order.Status) { // 如果订单状态是2（已接单），3（已拒单），读取关联的扫码订单商品
			for _, product := range order.H5OrderProducts {
				num = num + product.Num
				totalPrice := decimal.NewFromFloat(product.Price).Mul(decimal.NewFromInt(int64(product.Num))).Truncate(2).InexactFloat64()
				price = price + totalPrice
			}
		}
		var regionUuid uint64
		if order.Desk != nil {
			regionUuid = order.Desk.RegionUuid
		}
		items = append(items, resp.H5OrderItem{
			H5OrderInfo: resp.H5OrderInfo{
				SaleBillUuid: order.SaleBillUuid,
				H5OrderUuid:  order.Uuid,
				OrderTime:    order.OrderTime,
				HandleTime:   order.HandleTime,
				WaitTime:     time.Now().Unix() - order.OrderTime,
				DeskNo:       order.DeskNo,
				Price:        price,
				Status:       order.Status,
			},
			DeskRegionUuid: regionUuid,
			Num:            num,
		})
	}

	unhandledCount, err := h5OrderRepo.GetH5OrderCount(unhandledStatusOption)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	listResp.Extra.UnhandledCount = unhandledCount

	handledCount, err := h5OrderRepo.GetH5OrderCount(handledStatusOption)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	listResp.Extra.HandledCount = handledCount

	listResp.List = items
	listResp.Meta = dto.PageResponse{
		PageNo:   listReq.PageNo,
		PageSize: listReq.PageSize,
		Total:    total,
	}
	return &listResp, nil
}

func (s *h5OrderSrv) GetH5OrderDetail(ctx context.Context, h5OrderUuid uint64) (*resp.H5OrderDetailResp, error) {
	h5OrderRepo := repository.NewH5OrderRepo(ctx.GetDB())
	h5Order, err := h5OrderRepo.GetH5OrderDetailOrdered(h5OrderUuid)
	if err != nil {
		return nil, errors.WithMessage(apperrors.ErrInternal, "获取h5订单详情失败", err.Error())
	}
	newProducts := make([]resp.ProductItem, 0)
	acceptedProducts := make([]resp.ProductItem, 0)
	var price float64
	if h5Order.Status == constant.H5OrderStatusOrder { // 待处理
		var saleBillUuid uint64
		if len(h5Order.SaleOrderProducts) > 0 {
			saleBillUuid = h5Order.SaleOrderProducts[0].SaleBillUuid
		}
		for _, product := range h5Order.SaleOrderProducts {
			if !product.IsAcceptOrderBool() {
				totalPrice := decimal.NewFromFloat(product.Price).Mul(decimal.NewFromInt(int64(product.Num))).Truncate(2).InexactFloat64()
				newProducts = append(newProducts, resp.ProductItem{
					LocaleName: product.MultiLanguageName.GetNames(),
					Num:        product.Num,
					TotalPrice: totalPrice,
				})
				price = price + totalPrice
			}
		}
		// 获取同一个销售账单，已接单的，h5订单商品
		if saleBillUuid > 0 {
			products, err := h5OrderRepo.GetH5OrderProductsBySaleBillUuidAndAccept(saleBillUuid)
			if err != nil {
				return nil, errors.WithMessage(apperrors.ErrInternal, "获取h5订单详情失败", err.Error())
			}
			for _, product := range products {
				if product.IsAccepted() {
					totalPrice := decimal.NewFromFloat(product.Price).Mul(decimal.NewFromInt(int64(product.Num))).Truncate(2).InexactFloat64()
					acceptedProducts = append(acceptedProducts, resp.ProductItem{
						LocaleName: product.SaleOrderProduct.MultiLanguageName.GetNames(),
						Num:        product.Num,
						TotalPrice: totalPrice,
					})
					price = price + totalPrice
				}
			}
		}
	} else { // 已接单、拒单
		for _, product := range h5Order.H5OrderProducts {
			totalPrice := decimal.NewFromFloat(product.Price).Mul(decimal.NewFromInt(int64(product.Num))).Truncate(2).InexactFloat64()
			newProducts = append(newProducts, resp.ProductItem{
				LocaleName: product.SaleOrderProduct.MultiLanguageName.GetNames(),
				Num:        product.Num,
				TotalPrice: totalPrice,
			})
			price = price + totalPrice
		}
	}

	var cashier string
	if h5Order.Staff != nil {
		cashier = h5Order.Staff.RealName
	}

	operationLogs, _ := s.orderSrv.GetRecordList(ctx, h5Order.SaleBillUuid, h5OrderUuid)
	return &resp.H5OrderDetailResp{
		H5OrderDetail: resp.H5OrderDetail{
			H5OrderInfo: resp.H5OrderInfo{
				SaleBillUuid: h5Order.SaleBillUuid,
				H5OrderUuid:  h5Order.Uuid,
				OrderTime:    h5Order.OrderTime,
				HandleTime:   h5Order.HandleTime,
				WaitTime:     time.Now().Unix() - h5Order.OrderTime,
				DeskNo:       h5Order.DeskNo,
				Price:        price,
				Status:       h5Order.Status,
			},
			DeskUuid: h5Order.DeskUuid,
			Cashier:  cashier,
		},
		NewProduct: resp.ProductList{
			List: newProducts,
		},
		AcceptedProduct: resp.ProductList{
			List: acceptedProducts,
		},
		OperationLog: resp.OperationLog{
			List: operationLogs,
		},
	}, nil
}

func (s *h5OrderSrv) RejectH5Order(ctx context.Context, h5OrderUuid uint64) error {
	return s.orderSrv.RejectH5Order(ctx, h5OrderUuid)
}

func (s *h5OrderSrv) AcceptH5Order(ctx context.Context, h5OrderUuid uint64, isAutoOrder bool) (*resp.OrderCheckServiceRes, error) {
	return s.orderSrv.AcceptH5Order(ctx, h5OrderUuid, isAutoOrder)
}
