package service

import (
	"slices"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"

	"github.com/shopspring/decimal"
)

// IH5OrderSrv 定义接单服务接口
type IH5OrderSrv interface {
	GetH5OrderList(ctx context.Context, listReq req.H5OrderListReq) (*resp.H5OrderList, error) // 获取h5订单列表
	GetH5OrderDetail(ctx context.Context, orderUuid uint64) (*resp.H5OrderDetailResp, error)   // 获取h5订单详情
	RejectH5Order(ctx context.Context, h5OrderUuid uint64) error                               // 拒单
	AcceptH5Order(ctx context.Context, orderUuid uint64) (*resp.OrderCheckServiceRes, error)   // 接单
}

type h5OrderSrv struct {
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
	// 只查询前7天的数据
	limitTime := repository.CommonRepo.WhereCreateTimeGt(time.Now().Unix() - 7*24*60*60)
	if *listReq.Status == 0 { // 未处理
		dbOptions = append(dbOptions, unhandledStatusOption, repository.CommonRepo.SortWithCreateTime("asc"))
	} else { // 已处理：已接单、已拒单
		dbOptions = append(dbOptions, handledStatusOption, repository.CommonRepo.SortWithHandleTime("desc"), limitTime)
	}
	// 桌台区域条件
	if listReq.DeskRegionUuid > 0 {
		dbOptions = append(dbOptions, h5OrderRepo.WhereDeskRegionUuid(listReq.DeskRegionUuid))
	}
	// 订单类型条件（-1=全部、0=桌台扫码订单、1=会员端堂食订单）
	if listReq.OrderType >= 0 {
		dbOptions = append(dbOptions, h5OrderRepo.WhereOrderType(listReq.OrderType))
	}
	dbOptions = append(
		dbOptions,
		h5OrderRepo.WhereIsNeedAudit(1),
		h5OrderRepo.WithDesk(),
		repository.CommonRepo.Preload(
			repository.WithPreload{
				Query: "H5OrderProducts",
				Args: []any{
					repository.CommonRepo.DBOption(repository.CommonRepo.WhereBySoftDelete()),
				},
			},
		),
		repository.CommonRepo.Preload(
			repository.WithPreload{
				Query: "SaleOrderProducts",
				Args: []any{
					repository.CommonRepo.DBOption(repository.CommonRepo.WhereBySoftDelete()),
				},
			},
		),
		h5OrderRepo.WithSaleOrderProductsMultiLanguageName(),
	)
	orders, total, err := h5OrderRepo.PaginateGetH5Order(listReq.PageNo, listReq.PageSize, dbOptions...)
	if err != nil {
		return nil, errors.ErrInternal
	}
	items := make([]resp.H5OrderItem, 0, len(orders))
	for _, order := range orders {
		var num float64
		var price float64
		if order.Status == constant.H5OrderStatusOrder { // 如果订单状态是1（待处理），读取关联的销售订单商品
			for _, product := range order.SaleOrderProducts {
				num = num + product.Num
				totalPrice := product.GetProductFinalSalePrice()
				price = price + totalPrice
			}
			// 再加上已经接单的商品价格。同一个销售账单，已接单的，h5订单商品
			if order.SaleBillUuid > 0 {
				amount := decimal.NewFromFloat(0) // 已下单的商品金额之和
				products, err := h5OrderRepo.GetH5OrderProductsBySaleBillUuidAndAccept(order.SaleBillUuid)
				if err != nil {
					return nil, errors.WithMessage(errors.New("获取h5订单详情失败"), err.Error())
				}
				for _, product := range products {
					totalPrice := decimal.NewFromFloat(product.Price).Mul(decimal.NewFromFloat(product.Num)).Truncate(2).InexactFloat64()
					amount = amount.Add(decimal.NewFromFloat(totalPrice))
				}
				price += amount.InexactFloat64()
			}
		} else if slices.Contains([]uint{constant.H5OrderStatusAccepted, constant.H5OrderStatusRejected}, order.Status) { // 如果订单状态是2（已接单），3（已拒单），读取关联的扫码订单商品
			for _, product := range order.H5OrderProducts {
				num = num + product.Num
				totalPrice := decimal.NewFromFloat(product.Price).Mul(decimal.NewFromFloat(product.Num)).Truncate(2).InexactFloat64()
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
				OrderTime:    order.CreateTime,
				HandleTime:   order.HandleTime,
				WaitTime:     time.Now().Unix() - order.CreateTime,
				DeskNo:       order.DeskNo,
				Price:        price,
				Status:       order.Status,
				OrderType:    order.OrderType,
			},
			DeskRegionUuid: regionUuid,
			Num:            num,
		})
	}

	unhandledCount, err := h5OrderRepo.GetH5OrderCount(unhandledStatusOption, h5OrderRepo.WhereIsNeedAudit(1))
	if err != nil {
		return nil, errors.ErrInternal
	}
	listResp.Extra.UnhandledCount = unhandledCount

	handledCount, err := h5OrderRepo.GetH5OrderCount(handledStatusOption, limitTime, h5OrderRepo.WhereIsNeedAudit(1))
	if err != nil {
		return nil, errors.ErrInternal
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
		return nil, errors.WithMessage(errors.New("获取h5订单详情失败"), err.Error())
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
			if product.IsPackageSubProduct() {
				continue // 套餐子商品跳过
			}
			if !product.IsAcceptOrderBool() {
				totalPrice := product.GetProductFinalSalePrice()
				subProductList := make([]resp.SubProductItem, 0)
				if product.IsPackageProduct() {
					subProducts := h5Order.GetSubProducts(product.Uuid)
					for _, subProduct := range subProducts {
						subProductList = append(subProductList, resp.SubProductItem{
							LocaleName: subProduct.GetLocaleName(), // Requirement: story-main-product-attribute-snapshot-fix
							Num:        subProduct.Num,
						})
					}
				}

				newProducts = append(newProducts, resp.ProductItem{
					LocaleName: product.GetLocaleName(), // Requirement: story-main-product-attribute-snapshot-fix
					Num:        product.Num,
					TotalPrice: totalPrice,
					SubProducts: resp.SubProductList{
						List: subProductList,
					}, // 套餐子商品列表
				})
				price = price + totalPrice
			}
		}
		// 获取同一个销售账单，已接单的，h5订单商品
		if saleBillUuid > 0 {
			products, err := h5OrderRepo.GetH5OrderProductsBySaleBillUuidAndAccept(saleBillUuid)
			if err != nil {
				return nil, errors.WithMessage(errors.New("获取h5订单详情失败"), err.Error())
			}
			for _, product := range products {
				if product.IsAccepted() {
					totalPrice := decimal.NewFromFloat(product.Price).Mul(decimal.NewFromFloat(product.Num)).Truncate(2).InexactFloat64()
					acceptedProducts = append(acceptedProducts, resp.ProductItem{
						LocaleName: product.SaleOrderProduct.GetLocaleName(), // Requirement: story-main-product-attribute-snapshot-fix
						Num:        product.Num,
						TotalPrice: totalPrice,
					})
					price = price + totalPrice
				}
			}
		}
	} else { // 已接单、拒单
		for _, product := range h5Order.H5OrderProducts {
			totalPrice := decimal.NewFromFloat(product.Price).Mul(decimal.NewFromFloat(product.Num)).Truncate(2).InexactFloat64()
			newProducts = append(newProducts, resp.ProductItem{
				LocaleName: product.SaleOrderProduct.GetLocaleName(), // Requirement: story-main-product-attribute-snapshot-fix
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

	// 查询支付方式列表（仅会员端堂食订单）
	paymentMethodItems := make([]resp.H5OrderPaymentMethodItem, 0)
	if h5Order.OrderType == constant.H5OrderTypeMemberDineIn && h5Order.SaleOrderUuid > 0 {
		paymentOrderRepo := repository.NewPaymentOrderRepo(ctx.GetDB())
		paymentOrders, _ := paymentOrderRepo.GetPaymentOrderList(
			repository.CommonRepo.WhereByRelatedUuid(h5Order.SaleOrderUuid),
			repository.CommonRepo.WhereByRelatedType(constant.PaymentOrderRelatedTypeSaleOrder),
			repository.CommonRepo.WhereByStatus(constant.PaymentOrderStatusPaid),
			repository.CommonRepo.WhereBySoftDelete(),
			paymentOrderRepo.WithPaymentMethod(),
		)
		for _, po := range paymentOrders {
			if po.PaymentMethod != nil {
				paymentMethodItems = append(paymentMethodItems, resp.H5OrderPaymentMethodItem{
					Name: po.PaymentMethod.Name,
					Code: po.PaymentMethod.Code,
				})
			}
		}
	}

	return &resp.H5OrderDetailResp{
		H5OrderDetail: resp.H5OrderDetail{
			H5OrderInfo: resp.H5OrderInfo{
				SaleBillUuid: h5Order.SaleBillUuid,
				H5OrderUuid:  h5Order.Uuid,
				OrderTime:    h5Order.CreateTime,
				HandleTime:   h5Order.HandleTime,
				WaitTime:     time.Now().Unix() - h5Order.CreateTime,
				DeskNo:       h5Order.DeskNo,
				Price:        price,
				Status:       h5Order.Status,
				OrderType:    h5Order.OrderType,
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
		PaymentMethod: resp.H5OrderPaymentMethodList{
			List: paymentMethodItems,
		},
	}, nil
}

func (s *h5OrderSrv) RejectH5Order(ctx context.Context, h5OrderUuid uint64) error {
	return s.orderSrv.RejectH5Order(ctx, h5OrderUuid)
}

func (s *h5OrderSrv) AcceptH5Order(ctx context.Context, h5OrderUuid uint64) (*resp.OrderCheckServiceRes, error) {
	return s.orderSrv.AcceptH5Order(ctx, h5OrderUuid, false)
}
