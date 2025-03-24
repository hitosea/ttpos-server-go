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

	"gorm.io/gorm"
)

// IH5OrderSrv 定义接单服务接口
type IH5OrderSrv interface {
	GetH5OrderList(companyUuid uint64, acceptOrderListReq req.H5OrderListReq) (resp.H5OrderList, error) // 获取h5订单列表
	GetH5OrderDetail(companyUuid uint64, orderUuid uint64) (*resp.H5OrderDetailResp, error)             // 获取h5订单详情
	RejectH5Order(ctx context.Context, h5OrderUuid uint64) error                                        // 拒单
	AcceptH5Order(ctx context.Context, orderUuid uint64) (*resp.OrderCheckServiceRes, error)            // 接单
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

func (s *h5OrderSrv) GetH5OrderList(companyUuid uint64, listReq req.H5OrderListReq) (resp.H5OrderList, error) {
	var listResp = resp.H5OrderList{
		List: make([]resp.H5OrderItem, 0),
	}
	h5OrderRepo := repository.NewH5OrderRepo(s.dbm.GetDB(companyUuid))
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
		return listResp, apperrors.ErrInternal
	}
	items := make([]resp.H5OrderItem, 0, len(orders))
	for _, order := range orders {
		var num uint
		var price float64
		if order.Status == constant.H5OrderStatusOrder { // 如果订单状态是1（待处理），读取关联的销售订单商品
			for _, product := range order.SaleOrderProducts {
				num = num + product.Num
				price = price + product.Price
			}
		} else if slices.Contains([]uint{constant.H5OrderStatusAccepted, constant.H5OrderStatusRejected}, order.Status) { // 如果订单状态是2（已接单），3（已拒单），读取关联的扫码订单商品
			for _, product := range order.H5OrderProducts {
				num = num + product.Num
				price = price + product.Price
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
		return listResp, apperrors.ErrInternal
	}
	listResp.Extra.UnhandledCount = unhandledCount

	handledCount, err := h5OrderRepo.GetH5OrderCount(handledStatusOption)
	if err != nil {
		return listResp, apperrors.ErrInternal
	}
	listResp.Extra.HandledCount = handledCount

	listResp.List = items
	listResp.Meta = dto.PageResponse{
		PageNo:   listReq.PageNo,
		PageSize: listReq.PageSize,
		Total:    total,
	}
	return listResp, nil
}

func (s *h5OrderSrv) GetH5OrderDetail(companyUuid uint64, h5OrderUuid uint64) (*resp.H5OrderDetailResp, error) {
	h5OrderRepo := repository.NewH5OrderRepo(s.dbm.GetDB(companyUuid))
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
				newProducts = append(newProducts, resp.ProductItem{
					LocaleName: product.MultiLanguageName.GetNames(),
					Num:        product.Num,
					TotalPrice: product.Price,
				})
			}

			price = price + product.Price
		}
		// 获取同一个销售账单，已接单的，h5订单商品
		if saleBillUuid > 0 {
			products, err := h5OrderRepo.GetH5OrderProductsBySaleBillUuidAndAccept(saleBillUuid)
			if err != nil {
				return nil, errors.WithMessage(apperrors.ErrInternal, "获取h5订单详情失败", err.Error())
			}
			for _, product := range products {
				if product.IsAccepted() {
					acceptedProducts = append(acceptedProducts, resp.ProductItem{
						LocaleName: product.SaleOrderProduct.MultiLanguageName.GetNames(),
						Num:        product.Num,
						TotalPrice: product.Price,
					})
					price = price + product.Price
				}
			}
		}
	} else { // 已接单、拒单
		for _, product := range h5Order.H5OrderProducts {
			newProducts = append(newProducts, resp.ProductItem{
				LocaleName: product.SaleOrderProduct.MultiLanguageName.GetNames(),
				Num:        product.Num,
				TotalPrice: product.Price,
			})
			price = price + product.Price
		}
	}

	var cashier string
	if h5Order.Staff != nil {
		cashier = h5Order.Staff.RealName
	}

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
		NewProductList: resp.ProductList{
			List: newProducts,
		},
		AcceptedProductList: resp.ProductList{
			List: acceptedProducts,
		},
		OperationLogList: resp.OperationLogList{ // ToDo 处理日志
			List: make([]resp.OperationLogItem, 0),
		},
	}, nil
}

func (s *h5OrderSrv) RejectH5Order(ctx context.Context, h5OrderUuid uint64) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	h5OrderRepo := repository.NewH5OrderRepo(db)
	// 获取h5订单
	order, err := h5OrderRepo.GetH5OrderDetail(h5OrderUuid)
	if err != nil {
		return errors.WithMessage(apperrors.ErrInternal, "获取h5订单失败", err.Error())
	}
	// 非待处理状态不可操作
	if order.Status != constant.H5OrderStatusOrder {
		return errors.WithMessage(apperrors.ErrInternal, "当前状态不可操作")
	}

	// 拒单,保证h5订单的商品快照信息
	order.Reject(ctx.GetStaffUuid(), ctx.GetLanguage())
	// 删除销售订单商品
	order.DeleteSaleOrderProduct()

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 更新h5订单
		if err := repository.NewH5OrderRepo(db).UpdateH5OrderRecord(*order); err != nil {
			return errors.WithMessage(err, "更新h5订单失败")
		}
		// 更新h5订单商品列表
		for _, h5OrderProduct := range order.H5OrderProducts {
			// 更新h5订单商品
			if err := repository.NewH5OrderRepo(db).UpdateH5OrderProductRecord(*h5OrderProduct); err != nil {
				return errors.WithMessage(err, "更新h5订单商品失败")
			}
		}
		// 删除销售订单商品.将该h5订单的商品删除
		if err := repository.NewSaleOrderProductRepo(db).DeleteSaleOrderProductList(order.SaleOrderProducts); err != nil {
			return errors.WithMessage(err, "删除销售订单商品失败")
		}

		// ToDo 增加订单操作日志
		return nil
	}); err != nil {
		return errors.WithMessage(err, "拒单失败")
	}
	return nil
}

func (s *h5OrderSrv) AcceptH5Order(ctx context.Context, h5OrderUuid uint64) (*resp.OrderCheckServiceRes, error) {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	h5OrderRepo := repository.NewH5OrderRepo(db)
	// 获取h5订单
	h5Order, err := h5OrderRepo.GetH5OrderDetail(h5OrderUuid)
	if err != nil {
		return nil, errors.WithMessage(apperrors.ErrInternal, "获取h5订单失败", err.Error())
	}
	// 非待处理状态不可操作
	if h5Order.Status != constant.H5OrderStatusOrder {
		return nil, errors.WithMessage(apperrors.ErrInternal, "当前状态不可操作")
	}

	// 接单,保证h5订单的商品快照信息
	h5Order.Accept(ctx.GetStaffUuid(), ctx.GetLanguage())
	// 将已下单的h5订单商品变为已接单单的h5订单商品
	h5Order.ChangeToAccepted()
	// 送厨已经接单的商品。送厨指定的商品列表

	{
		ignoreMust := true // 接单，送厨忽略必点方案
		// 获取销售账单信息
		saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(h5Order.SaleOrder.SaleBillUuid)
		if errSaleBill != nil {
			return nil, errors.WithMessage(errSaleBill, "repository.NewOrderRepo(db).GetSaleBillAllInfo")
		}
		ctx.Log().Debug("获取销售账单信息")

		// 获取本次接单的商品列表
		unCookingSaleOrderProducts := h5Order.SaleOrderProducts

		// 送厨
		checkServiceRes, err := s.orderSrv.ActionCooking(ctx, ignoreMust, saleBill, unCookingSaleOrderProducts)
		if err != nil {
			return nil, errors.WithMessage(err, "ActionCooking")
		}
		if checkServiceRes != nil {
			return checkServiceRes, nil
		}
	}
	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 更新h5订单
		if err := repository.NewH5OrderRepo(db).UpdateH5OrderRecord(*h5Order); err != nil {
			return errors.WithMessage(err, "更新h5订单失败")
		}
		// 更新h5订单商品列表
		for _, h5OrderProduct := range h5Order.H5OrderProducts {
			// 更新h5订单商品
			if err := repository.NewH5OrderRepo(db).UpdateH5OrderProductRecord(*h5OrderProduct); err != nil {
				return errors.WithMessage(err, "更新h5订单商品失败")
			}
		}
		// 更新销售订单商品.将该h5订单的商品变为已接单
		// for _, saleOrderProduct := range h5Order.SaleOrderProducts {
		// 	if err := repository.NewSaleOrderProductRepo(db).UpdateSaleOrderProductRecord(*saleOrderProduct); err != nil {
		// 		return errors.WithMessage(err, "将已下单的h5订单商品变为已接单单的h5订单商品失败")
		// 	}
		// }

		// ToDo 增加订单操作日志
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err, "接单失败")
	}
	return nil, nil
}
