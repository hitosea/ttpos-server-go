package service

import (
	"slices"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	apperrors "ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/database"
)

// IH5OrderSrv 定义接单服务接口
type IH5OrderSrv interface {
	GetH5OrderList(companyUuid uint64, acceptOrderListReq req.H5OrderListReq) (resp.H5OrderList, error) // 获取h5订单列表
	GetH5OrderDetail(companyUuid uint64, orderUuid uint64) (resp.H5OrderDetailResp, error)              // 获取h5订单详情
}

type h5OrderSrv struct {
	dbm *database.DBManager
}

func NewH5OrderSrv(dbm *database.DBManager) IH5OrderSrv {
	return NewH5OrderSrvImpl(dbm)
}

func NewH5OrderSrvImpl(dbm *database.DBManager) IH5OrderSrv {
	return &h5OrderSrv{
		dbm: dbm,
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
	dbOptions = append(dbOptions, h5OrderRepo.WithDesk(), h5OrderRepo.WithH5OrderProducts(), h5OrderRepo.WithSaleOrderProducts())
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
				OrderUuid:  order.Uuid,
				OrderTime:  order.OrderTime,
				HandleTime: order.HandleTime,
				DeskNo:     order.DeskNo,
				Price:      price,
				Status:     order.Status,
			},
			DeskRegionUuid: regionUuid,
			Num:            num,
		})
	}

	unhandledCount, err := h5OrderRepo.GetH5OrderCount(unhandledStatusOption)
	if err != nil {
		return listResp, apperrors.ErrInternal
	}
	listResp.UnhandledCount = unhandledCount

	handledCount, err := h5OrderRepo.GetH5OrderCount(handledStatusOption)
	if err != nil {
		return listResp, apperrors.ErrInternal
	}
	listResp.HandledCount = handledCount

	listResp.List = items
	listResp.Meta = dto.PageResponse{
		PageNo:   listReq.PageNo,
		PageSize: listReq.PageSize,
		Total:    total,
	}
	return listResp, nil
}

func (s *h5OrderSrv) GetH5OrderDetail(companyUuid uint64, orderUuid uint64) (resp.H5OrderDetailResp, error) {
	h5OrderRepo := repository.NewH5OrderRepo(s.dbm.GetDB(companyUuid))
	order, err := h5OrderRepo.GetH5Order(h5OrderRepo.WhereUuid(orderUuid), h5OrderRepo.WhereNotStatus([]uint{constant.H5OrderStatusChooseProduct}),
		h5OrderRepo.WithH5OrderProducts(), h5OrderRepo.WithH5OrderProductSaleOrderProduct(), h5OrderRepo.WithH5OrderProductSaleOrderProductMultiLanguageName(),
		h5OrderRepo.WithSaleOrderProducts(), h5OrderRepo.WithSaleOrderProductMultiLanguageName(), h5OrderRepo.WithCashier())
	if err != nil {
		return resp.H5OrderDetailResp{}, apperrors.ErrInternal
	}
	newProducts := make([]resp.ProductItem, 0)
	acceptedProducts := make([]resp.ProductItem, 0)
	var price float64
	if order.Status == 1 { // 待处理
		var saleBillUuid uint64
		if len(order.SaleOrderProducts) > 0 {
			saleBillUuid = order.SaleOrderProducts[0].SaleBillUuid
		}
		for _, product := range order.SaleOrderProducts {
			newProducts = append(newProducts, resp.ProductItem{
				NameLocale: product.MultiLanguageName.GetNames(),
				Num:        product.Num,
				TotalPrice: product.Price,
			})
			price = price + product.Price
		}
		// 获取同一个销售账单，已接单的，h5订单商品
		if saleBillUuid > 0 {
			products, err := h5OrderRepo.GetH5OrderProducts(h5OrderRepo.WhereSaleBillUuid(saleBillUuid),
				h5OrderRepo.WithSaleOrderProduct(), h5OrderRepo.WithSaleOrderProductMultiLanguageName())
			if err != nil {
				return resp.H5OrderDetailResp{}, apperrors.ErrInternal
			}
			// 所有h5订单uuid
			var h5OrderUuids []uint64
			for _, product := range products {
				if !slices.Contains(h5OrderUuids, product.H5OrderUuid) {
					h5OrderUuids = append(h5OrderUuids, product.H5OrderUuid)
				}
			}
			// 已接单的h5订单uuid
			h5OrderUuids, err = h5OrderRepo.GetH5OrderUuids(h5OrderRepo.WhereUuidIn(h5OrderUuids), h5OrderRepo.WhereStatus([]uint{constant.H5OrderStatusAccepted}))
			if err != nil {
				return resp.H5OrderDetailResp{}, apperrors.ErrInternal
			}
			for _, product := range products {
				if slices.Contains(h5OrderUuids, product.H5OrderUuid) {
					acceptedProducts = append(acceptedProducts, resp.ProductItem{
						NameLocale: product.SaleOrderProduct.MultiLanguageName.GetNames(),
						Num:        product.Num,
						TotalPrice: product.Price,
					})
					price = price + product.Price
				}
			}
		}
	} else { // 已接单、拒单
		for _, product := range order.H5OrderProducts {
			newProducts = append(newProducts, resp.ProductItem{
				NameLocale: product.SaleOrderProduct.MultiLanguageName.GetNames(),
				Num:        product.Num,
				TotalPrice: product.Price,
			})
			price = price + product.Price
		}
	}

	var cashier string
	if order.Staff != nil {
		cashier = order.Staff.RealName
	}

	return resp.H5OrderDetailResp{
		H5OrderDetail: resp.H5OrderDetail{
			H5OrderInfo: resp.H5OrderInfo{
				OrderUuid:  order.Uuid,
				OrderTime:  order.OrderTime,
				HandleTime: order.HandleTime,
				DeskNo:     order.DeskNo,
				Price:      price,
				Status:     order.Status,
			},
			DeskUuid: order.DeskUuid,
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
