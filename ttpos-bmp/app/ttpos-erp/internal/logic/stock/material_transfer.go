package stock

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/buying"
	"ttpos-bmp/app/ttpos-erp/api/material_transfer"
	"ttpos-bmp/app/ttpos-erp/api/warehouse"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/dao"
	dto "ttpos-bmp/app/ttpos-erp/internal/model/dto/buying"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	dtoSelling "ttpos-bmp/app/ttpos-erp/internal/model/dto/selling"
	"ttpos-bmp/app/ttpos-erp/internal/model/entity"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

/**
 * 物料转移
 * 用在华莱士内部 调入/调出。 实际上是通过 创建内部销售单 -> 内部采购单来实现
 */

var (
	MaterialTransfer = &sMaterialTransfer{}
)

type sMaterialTransfer struct {
}

func init() {
	service.RegisterMaterialTransfer(MaterialTransfer)
}

// getCompanyBranchAndWarehouse 获取公司的分支和在途仓库
// 参数：
//   - ctx: 上下文对象
//   - companyAbbr: 公司简称
//
// 返回：
//   - branch: 分支名称
//   - warehouseName: 在途仓库名称
//   - err: 错误信息
func (s *sMaterialTransfer) getCompanyBranchAndWarehouse(ctx context.Context, companyAbbr string) (branch string, warehouseName string, err error) {
	// 获取公司的 branch
	record, err := dao.ShopCashier.Ctx(ctx).Where(dao.ShopCashier.Columns().CompanyAbbr, companyAbbr).One()
	if err != nil {
		return "", "", gerror.Wrapf(err, "获取公司[%s]的branch失败", companyAbbr)
	}

	shopCashier := &entity.ShopCashier{}
	err = record.Struct(&shopCashier)
	if err != nil {
		g.Log().Error(ctx, "解析公司branch信息失败", err)
		//return "", "", gerror.Wrapf(err, "解析公司[%s]的branch失败", companyAbbr)
	}
	branch = shopCashier.Branch

	// 获取在途仓库列表
	warehouseList, err := service.Warehouse().GetWarehouseList(ctx, &warehouse.GetWarehouseListReq{
		Company:       companyAbbr,
		WarehouseType: erp.WarehouseTypeTransit,
	})
	if err != nil || len(warehouseList.WarehouseList) == 0 {
		return "", "", gerror.Newf("公司[%s]的默认在途仓库不存在", companyAbbr)
	}

	// 优先选择与分支相同的在途仓
	for _, warehouseInfo := range warehouseList.WarehouseList {
		if warehouseInfo.Branch == branch {
			warehouseName = warehouseInfo.WarehouseName
			return branch, warehouseName, nil
		}
		warehouseName = warehouseInfo.WarehouseName
	}

	return branch, warehouseName, nil
}

// getTransitWarehouse 获取在途仓库
// 参数：
//   - ctx: 上下文对象
//   - companyAbbr: 公司简称
//   - branch: 分支名称
//
// 返回：
//   - warehouseName: 在途仓库名称
//   - err: 错误信息
func (s *sMaterialTransfer) getTransitWarehouse(ctx context.Context, companyAbbr string, branch string) (warehouseName string, err error) {
	companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, companyAbbr)
	if err != nil {
		return "", gerror.Wrapf(err, "获取公司[%s]的名称失败", companyAbbr)
	}
	// 获取在途仓库列表
	warehouseList, err := service.Warehouse().GetWarehouseList(ctx, &warehouse.GetWarehouseListReq{
		Company:       companyName,
		WarehouseType: erp.WarehouseTypeTransit,
	})
	if err != nil || len(warehouseList.WarehouseList) == 0 {
		return "", gerror.Newf("公司[%s]的默认在途仓库不存在", companyAbbr)
	}

	// 优先选择与分支相同的在途仓
	for _, warehouseInfo := range warehouseList.WarehouseList {
		if warehouseInfo.Branch == branch {
			warehouseName = warehouseInfo.Name
			return warehouseName, nil
		}
		warehouseName = warehouseInfo.Name
	}

	return warehouseName, nil
}

// MaterialTransfer 调入调出
// 1. 调出方父公司与调入方公司相同时，直接调出到调入方公司 。 返回节点3组都相同的单号
// 2. 调出方与调入方父级公司相同时，先调入父公司，再调入到调入方公司。 返回 审核节点和调入方节点的单号都是相同的
// 3. 调出方与调入方父级公司不同时，先调出到调出方父级公司， 再调出到对方父级公司，再调出到调入方公司。返回节点的3组单号都不相同

func (s *sMaterialTransfer) MaterialTransfer(ctx context.Context, req *material_transfer.MaterialTransferReq) (*material_transfer.MaterialTransferResp, error) {

	transferResp := &material_transfer.MaterialTransferResp{}
	//case 1
	//判断调入方是否自己的父级公司, 或者 调出方与调入方父级公司为空时
	if (req.FromParentCompanyAbbr == "" && req.ToParentCompanyAbbr == "") || req.FromParentCompanyAbbr == req.ToCompanyAbbr {
		transferReceipt, err := s.CreateInnerTransferReceipt(ctx, &material_transfer.MaterialTransferReq{
			FromCompanyAbbr: req.FromCompanyAbbr,
			FromBranch:      req.FromBranch,
			ToCompanyAbbr:   req.ToCompanyAbbr,
			ToBranch:        req.ToBranch,
			FromWarehouse:   req.FromWarehouse,
			ToWarehouse:     req.ToWarehouse,
			RequiredDate:    req.RequiredDate,
			DeliveryDate:    req.DeliveryDate,
			Items:           req.Items,
		})
		if err != nil {
			return nil, gerror.Wrapf(err, "创建内部采购单失败")
		}
		transferResp.FromReceipt = &material_transfer.TransferReceipt{
			PoNo:            transferReceipt.PoNo,
			SoNo:            transferReceipt.SoNo,
			FromCompanyAbbr: req.FromCompanyAbbr,
			ToCompanyAbbr:   req.ToCompanyAbbr,
		}
		transferResp.AuditReceipt = transferResp.FromReceipt
		transferResp.ToReceipt = transferResp.FromReceipt

	}
	//case 2
	//调出方与调入方父级公司相同、调入调出方父级任意一个为空时，先调入审核公司，再调入到调入方公司。 返回 审核节点和调入方节点的单号都是相同的
	if (req.FromParentCompanyAbbr != "" && req.ToParentCompanyAbbr == "") ||
		(req.FromParentCompanyAbbr == "" && req.ToParentCompanyAbbr != "") ||
		((req.FromParentCompanyAbbr != "" && req.ToParentCompanyAbbr != "") && req.FromParentCompanyAbbr == req.ToParentCompanyAbbr) {
		var (
			auditCompanyAbbr, auditBranch string
		)
		if req.FromParentCompanyAbbr != "" && req.ToParentCompanyAbbr != "" {
			auditCompanyAbbr = req.FromParentCompanyAbbr
			auditBranch = req.FromParentBranch
		} else {
			//合并空字符串
			auditCompanyAbbr = req.FromParentCompanyAbbr + req.ToParentCompanyAbbr
			auditBranch = req.ToParentCompanyAbbr + req.ToParentBranch
		}

		// 获取调出方父公司的分支和在途仓
		auditWarehouse, err := s.getTransitWarehouse(ctx, auditCompanyAbbr, auditBranch)
		if err != nil {
			return nil, err
		}

		// 调出方发起销售订单，目标是父级公司
		transferReceipt, err := s.CreateInnerTransferReceipt(ctx, &material_transfer.MaterialTransferReq{
			FromCompanyAbbr: req.FromCompanyAbbr,
			FromBranch:      req.FromBranch,
			ToCompanyAbbr:   auditCompanyAbbr,
			ToBranch:        auditBranch,
			FromWarehouse:   req.FromWarehouse,
			ToWarehouse:     auditWarehouse,
			RequiredDate:    req.RequiredDate,
			DeliveryDate:    req.DeliveryDate,
			Items:           req.Items,
		})
		if err != nil {
			return nil, gerror.Wrapf(err, "创建内部采购单失败")
		}
		transferResp.FromReceipt = &material_transfer.TransferReceipt{
			PoNo:            transferReceipt.PoNo,
			SoNo:            transferReceipt.SoNo,
			FromCompanyAbbr: req.FromCompanyAbbr,
			ToCompanyAbbr:   auditCompanyAbbr,
		}
		//父级公司发起销售订单，目标是调入公司
		transferReceipt, err = s.CreateInnerTransferReceipt(ctx, &material_transfer.MaterialTransferReq{
			FromCompanyAbbr: auditCompanyAbbr,
			FromBranch:      auditBranch,
			ToCompanyAbbr:   req.ToCompanyAbbr,
			ToBranch:        req.ToBranch,
			FromWarehouse:   auditWarehouse,
			ToWarehouse:     req.ToWarehouse,
			RequiredDate:    req.RequiredDate,
			DeliveryDate:    req.DeliveryDate,
			Items:           req.Items,
		})
		if err != nil {
			return nil, gerror.Wrapf(err, "创建内部采购单失败")
		}
		transferResp.AuditReceipt = &material_transfer.TransferReceipt{
			PoNo:            transferReceipt.PoNo,
			SoNo:            transferReceipt.SoNo,
			FromCompanyAbbr: auditCompanyAbbr,
			ToCompanyAbbr:   req.ToCompanyAbbr,
		}
		transferResp.ToReceipt = transferResp.AuditReceipt
	}
	//case 3
	//调出方与调入方父级公司不同时，先调出到调出方父级公司， 再调出到对方父级公司，再调出到调入方公司。返回节点的3组单号都不相同
	if (req.FromParentCompanyAbbr != "" && req.ToParentCompanyAbbr != "") && req.FromParentCompanyAbbr != req.ToParentCompanyAbbr {
		var (
			fromParentWarehouse, toParentWarehouse string
			err                                    error
		)
		//step1 调出公司发起销售订单，目标是调出方父公司
		{
			// 获取调出方父公司的分支和在途仓
			fromParentWarehouse, err = s.getTransitWarehouse(ctx, req.FromParentCompanyAbbr, req.FromParentBranch)
			if err != nil {
				return nil, err
			}

			// 调出方发起销售订单，目标是父级公司
			transferReceipt, err := s.CreateInnerTransferReceipt(ctx, &material_transfer.MaterialTransferReq{
				FromCompanyAbbr: req.FromCompanyAbbr,
				FromBranch:      req.FromBranch,
				ToCompanyAbbr:   req.FromParentCompanyAbbr,
				ToBranch:        req.FromParentBranch,
				FromWarehouse:   req.FromWarehouse,
				ToWarehouse:     fromParentWarehouse,
				RequiredDate:    req.RequiredDate,
				DeliveryDate:    req.DeliveryDate,
				Items:           req.Items,
			})
			if err != nil {
				return nil, gerror.Wrapf(err, "创建内部采购单失败")
			}
			transferResp.FromReceipt = &material_transfer.TransferReceipt{
				PoNo:            transferReceipt.PoNo,
				SoNo:            transferReceipt.SoNo,
				FromCompanyAbbr: req.FromCompanyAbbr,
				ToCompanyAbbr:   req.FromParentCompanyAbbr,
			}
		}

		//step2 父级公司发起销售订单，目标是调入方父公司
		{
			// 获取调入方父公司的分支和在途仓
			toParentWarehouse, err = s.getTransitWarehouse(ctx, req.ToParentCompanyAbbr, req.ToParentBranch)
			if err != nil {
				return nil, err
			}

			transferReceipt, err := s.CreateInnerTransferReceipt(ctx, &material_transfer.MaterialTransferReq{
				FromCompanyAbbr: req.FromParentCompanyAbbr,
				FromBranch:      req.FromParentBranch,
				ToCompanyAbbr:   req.ToParentCompanyAbbr,
				ToBranch:        req.ToParentBranch,
				FromWarehouse:   fromParentWarehouse,
				ToWarehouse:     toParentWarehouse,
				RequiredDate:    req.RequiredDate,
				DeliveryDate:    req.DeliveryDate,
				Items:           req.Items,
			})
			if err != nil {
				return nil, gerror.Wrapf(err, "创建内部采购单失败")
			}
			transferResp.AuditReceipt = &material_transfer.TransferReceipt{
				PoNo:            transferReceipt.PoNo,
				SoNo:            transferReceipt.SoNo,
				FromCompanyAbbr: req.FromParentCompanyAbbr,
				ToCompanyAbbr:   req.ToParentCompanyAbbr,
			}
		}
		{
			//step3 调入方父公司发起销售订单，目标是调入方公司
			transferReceipt, err := s.CreateInnerTransferReceipt(ctx, &material_transfer.MaterialTransferReq{
				FromCompanyAbbr: req.ToParentCompanyAbbr,
				FromBranch:      req.ToParentBranch,
				ToCompanyAbbr:   req.ToCompanyAbbr,
				ToBranch:        req.ToBranch,
				FromWarehouse:   toParentWarehouse,
				ToWarehouse:     req.ToWarehouse,
				RequiredDate:    req.RequiredDate,
				DeliveryDate:    req.DeliveryDate,
				Items:           req.Items,
			})
			if err != nil {
				return nil, gerror.Wrapf(err, "创建内部采购单失败")
			}
			transferResp.ToReceipt = &material_transfer.TransferReceipt{
				PoNo:            transferReceipt.PoNo,
				SoNo:            transferReceipt.SoNo,
				FromCompanyAbbr: req.ToParentCompanyAbbr,
				ToCompanyAbbr:   req.ToCompanyAbbr,
			}
		}
	}

	return transferResp, nil
}

// CreateInnerTransferReceipt  实际上是通过 创建内部销售单 -> 内部采购单来实现
func (s *sMaterialTransfer) CreateInnerTransferReceipt(ctx context.Context, req *material_transfer.MaterialTransferReq) (*material_transfer.TransferReceipt, error) {
	//判断 需求时间和发货时间是否为空， 默认使用
	//          RequiredDate:    consts.DefaultRequiredByDate,
	//			DeliveryDate:    consts.DefaultDeliveryDate,
	if req.RequiredDate == "" {
		req.RequiredDate = consts.DefaultRequiredByDate
	}
	if req.DeliveryDate == "" {
		req.DeliveryDate = consts.DefaultDeliveryDate
	}

	transferReceipt := &material_transfer.TransferReceipt{}
	//检查调入方的交易对象是否包含了调出公司，如果没有默认添加
	toCompanyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.ToCompanyAbbr)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取调出方公司失败")
	}
	//检查调出方父级公司的内部客户的交易对象是否包含了调出方公司，如果没有默认添加
	fromCompanyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.FromCompanyAbbr)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取调出方公司失败")
	}

	var customer *erp.Customer
	customers, err := service.Selling().ListCustomers(ctx, &dtoSelling.ListCustomersReq{
		RepresentsCompany: toCompanyName,
		PageSize:          1,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "获取调出方供应商的交易对象失败")
	}
	if len(customers) == 0 {
		//创建内部客户
		customer, err = service.Selling().CreateCustomer(ctx, &erp.Customer{
			CustomerName:       toCompanyName + "-" + req.ToCompanyAbbr,
			CustomerType:       "Company",
			IsInternalCustomer: 1,
			RepresentsCompany:  toCompanyName,
			CustomerGroup:      consts.DefaultCustomerGroupName,
		})
		if err != nil {
			return nil, gerror.Wrapf(err, "创建内部客户失败")
		}
	} else {
		customer = customers[0]
	}
	containsCompany := false
	for _, companyItem := range customer.Companies {
		if companyItem.Company == toCompanyName {
			// 调出方公司已在调入方交易对象中
			containsCompany = true
			break
		}
	}
	if !containsCompany {
		// 调出方公司不在调入方交易对象中，默认添加
		err = service.Selling().AddCompanyToCustomer(ctx, customer, fromCompanyName)
		if err != nil {
			return nil, gerror.Wrapf(err, "添加调出方公司到调入方客户交易对象失败")
		}
	}

	var supplier *buying.SupplierData
	suppliers, err := service.Supplier().ListSuppliers(ctx, &buying.ListSuppliersReq{
		RepresentsCompany: fromCompanyName,
		PageSize:          1,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "获取调出方供应商的交易对象失败")
	}
	if len(suppliers.Suppliers) == 0 {
		resp, err := service.Supplier().CreateSupplier(ctx, &buying.CreateSupplierReq{
			Supplier: &buying.SupplierData{
				SupplierName:       req.FromBranch + "-" + req.FromCompanyAbbr,
				AliasName:          req.FromBranch,
				Branch:             req.FromBranch,
				CompanyAbbr:        req.FromCompanyAbbr,
				IsInternalSupplier: true,
			},
		})
		if err != nil {
			return nil, gerror.Wrapf(err, "创建内部供应商失败")
		}
		supplier = resp.Supplier
	} else {
		supplier = suppliers.Suppliers[0]
	}
	//检查调入方父级公司的内部供应商的交易对象是否包含了调出方公司，如果没有默认添加
	err = service.Supplier().AddSupplerTransactCompany(ctx, &dto.AddSupplerTransactCompanyReq{
		Supplier:        supplier.Name,
		WithCompanyAbbr: req.ToCompanyAbbr,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "添加调出方公司到调入方供应商交易对象失败")
	}

	//处理销售和采购的物品
	saleOrderItems := make([]dtoSelling.SalesOrderItem, 0)
	purchaseOrderItems := make([]*buying.PurchaseOrderItemInput, 0)

	for _, item := range req.Items {
		saleOrderItems = append(saleOrderItems, dtoSelling.SalesOrderItem{
			ItemCode: item.ItemCode,
			Qty:      item.Qty,
			Uom:      item.Uom,
			Rate:     item.Rate,
		})
		purchaseOrderItems = append(purchaseOrderItems, &buying.PurchaseOrderItemInput{
			ItemCode: item.ItemCode,
			Qty:      item.Qty,
			Uom:      item.Uom,
			Rate:     item.Rate,
		})
	}
	//调出方发起销售订单

	// 获取默认采购价格表
	//defaultPriceList, err := service.PosPriceList().GetPosPriceListByCompany(ctx, fromCompanyName)
	//if err != nil {
	//	g.Log().Warningf(ctx, "获取采购价格表失败，company: %s", fromCompanyName)
	//	defaultPriceList, err = service.PosPriceList().GetDefaultPosPriceList(ctx)
	//	if err != nil {
	//		return nil, gerror.Wrapf(err, "获取默认采购价格表失败")
	//	}
	//}

	saleOrder, err := service.Selling().CreateSalesOrder(ctx, &dtoSelling.SalesOrder{
		Customer:         customers[0].Name,
		Company:          fromCompanyName,
		DeliveryDate:     req.DeliveryDate,
		SetWarehouse:     req.FromWarehouse,
		Items:            saleOrderItems,
		SellingPriceList: consts.DefaultTransferPriceList,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "创建销售订单失败")
	}
	transferReceipt.SoNo = saleOrder.Name

	//根据调出方销售订单，创建内部采购单
	purchaseOrder, err := service.Buying().CreatePurchaseOrderFromSalesOrder(ctx, &dto.CreatePurchaseOrderFromSalesOrderReq{
		SourceName:      saleOrder.Name,
		Supplier:        suppliers.Suppliers[0].Name,
		TargetWarehouse: req.ToWarehouse,
		ScheduleDate:    req.RequiredDate,
		BuyingPriceList: consts.DefaultTransferPriceList,
	})
	if err != nil {
		//失败时取消销售订单
		err2 := service.Selling().CancelSalesOrder(ctx, saleOrder.Name)
		if err2 != nil {
			g.Log().Warningf(ctx, "调拨时取消销售订单失败，销售订单名称：%s", saleOrder.Name)
		}
		return nil, gerror.Wrapf(err, "创建采购订单失败")
	}
	transferReceipt.PoNo = purchaseOrder.Name
	return transferReceipt, nil
}
