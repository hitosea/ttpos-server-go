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

}

// MaterialTransfer 调入调出
// 1. 调出方父公司与调入方公司相同时，直接调出到调入方公司 。 返回节点3组都相同的单号
// 2. 调出方与调入方父级公司相同时，先调入父公司，再调入到调入方公司。 返回 审核节点和调入方节点的单号都是相同的
// 3. 调出方与调入方父级公司不同时，先调出到调出方父级公司， 再调出到对方父级公司，再调出到调入方公司。返回节点的3组单号都不相同

func (s *sMaterialTransfer) MaterialTransfer(ctx context.Context, req *material_transfer.MaterialTransferReq) (*material_transfer.MaterialTransferResp, error) {
	//获取调出/入方父级公司
	fromCompany, err := service.Company().GetCompanyWithAbbr(ctx, req.FromCompanyAbbr)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取调出方公司失败")
	}
	fromParentCompanyName := fromCompany.ParentCompany

	toCompany, err := service.Company().GetCompanyWithAbbr(ctx, req.ToCompanyAbbr)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取调入方公司失败")
	}
	toParentCompanyName := toCompany.ParentCompany

	transferResp := &material_transfer.MaterialTransferResp{}

	fromParentCompany, err := service.Company().GetCompany(ctx, fromParentCompanyName)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取调出父级公司失败")
	}

	toParentCompany, err := service.Company().GetCompany(ctx, toParentCompanyName)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取调入方父级公司失败")
	}

	//case 1
	//判断调入方是否自己的父级公司
	if fromParentCompanyName == toCompany.CompanyName {
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
	if fromParentCompanyName == toParentCompanyName {
		//调出方与调入方父级公司相同时，先调入父公司，再调入到调入方公司。 返回 审核节点和调入方节点的单号都是相同的

		//获取调出方父公司的branch
		record, err := dao.ShopCashier.Ctx(ctx).Where(dao.ShopCashier.Columns().CompanyAbbr, fromParentCompany.Abbr).One()
		if err != nil {
			return nil, gerror.Wrapf(err, "获取调出方父公司的branch失败")
		}
		shopCashier := &entity.ShopCashier{}
		err = record.Struct(&shopCashier)
		if err != nil {
			return nil, gerror.Wrapf(err, "解析获取调出方父公司的branch失败")
		}
		//获取调入方父公司在途仓
		toWarehouse := ""
		warehouseList, err := service.Warehouse().GetWarehouseList(ctx, &warehouse.GetWarehouseListReq{
			Company:       fromParentCompany.Abbr,
			Branch:        shopCashier.Branch,
			WarehouseType: erp.WarehouseTypeTransit,
		})
		if err != nil || len(warehouseList.WarehouseList) == 0 {
			return nil, gerror.New("默认在途仓库不存在")
		}
		toWarehouse = warehouseList.WarehouseList[0].WarehouseName

		//调出方发起销售订单，目标是父级公司
		transferReceipt, err := s.CreateInnerTransferReceipt(ctx, &material_transfer.MaterialTransferReq{
			FromCompanyAbbr: req.FromCompanyAbbr,
			FromBranch:      req.FromBranch,
			ToCompanyAbbr:   fromParentCompany.Abbr,
			ToBranch:        shopCashier.Branch,
			FromWarehouse:   req.FromWarehouse,
			ToWarehouse:     toWarehouse,
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
		}
		//父级公司发起销售订单，目标是调入公司
		transferReceipt, err = s.CreateInnerTransferReceipt(ctx, &material_transfer.MaterialTransferReq{
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
		transferResp.AuditReceipt = &material_transfer.TransferReceipt{
			PoNo:            transferReceipt.PoNo,
			SoNo:            transferReceipt.SoNo,
			FromCompanyAbbr: req.FromCompanyAbbr,
		}
		transferResp.ToReceipt = transferResp.AuditReceipt
	}
	//case 3
	if fromParentCompanyName != toParentCompanyName {
		var (
			fromParentWarehouse = ""
			fromParentBranch    = ""
			toParentWarehouse   = ""
			toParentBranch      = ""
		)

		//调出方与调入方父级公司不同时，先调出到调出方父级公司， 再调出到对方父级公司，再调出到调入方公司。返回节点的3组单号都不相同

		//step1 调出公司发起销售订单，目标是调出方父公司
		{
			//获取调出方父公司的branch
			record, err := dao.ShopCashier.Ctx(ctx).Where(dao.ShopCashier.Columns().CompanyAbbr, fromParentCompany.Abbr).One()
			if err != nil {
				return nil, gerror.Wrapf(err, "获取调出方父公司的branch失败")
			}
			shopCashier := &entity.ShopCashier{}
			err = record.Struct(&shopCashier)
			if err != nil {
				return nil, gerror.Wrapf(err, "解析获取调出方父公司的branch失败")
			}
			fromParentBranch = shopCashier.Branch

			//获取调入方父公司在途仓
			warehouseList, err := service.Warehouse().GetWarehouseList(ctx, &warehouse.GetWarehouseListReq{
				Company:       fromParentCompany.Abbr,
				Branch:        fromParentBranch,
				WarehouseType: erp.WarehouseTypeTransit,
			})
			if err != nil || len(warehouseList.WarehouseList) == 0 {
				return nil, gerror.New("默认在途仓库不存在")
			}
			toWarehouse := warehouseList.WarehouseList[0].WarehouseName
			fromParentWarehouse = toWarehouse
			//调出方发起销售订单，目标是父级公司
			transferReceipt, err := s.CreateInnerTransferReceipt(ctx, &material_transfer.MaterialTransferReq{
				FromCompanyAbbr: req.FromCompanyAbbr,
				FromBranch:      req.FromBranch,
				ToCompanyAbbr:   fromParentCompany.Abbr,
				ToBranch:        fromParentBranch,
				FromWarehouse:   req.FromWarehouse,
				ToWarehouse:     toWarehouse,
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
			}
		}

		//step2 父级公司发起销售订单，目标是调入方父公司
		{
			//获取调出方父公司的branch
			record, err := dao.ShopCashier.Ctx(ctx).Where(dao.ShopCashier.Columns().CompanyAbbr, toParentCompany.Abbr).One()
			if err != nil {
				return nil, gerror.Wrapf(err, "获取调出方父公司的branch失败")
			}
			shopCashier := &entity.ShopCashier{}
			err = record.Struct(&shopCashier)
			if err != nil {
				return nil, gerror.Wrapf(err, "解析获取调出方父公司的branch失败")
			}
			toParentBranch = shopCashier.Branch
			//获取调入方父公司在途仓
			warehouseList, err := service.Warehouse().GetWarehouseList(ctx, &warehouse.GetWarehouseListReq{
				Company:       toParentCompany.Abbr,
				Branch:        toParentBranch,
				WarehouseType: erp.WarehouseTypeTransit,
			})
			if err != nil || len(warehouseList.WarehouseList) == 0 {
				return nil, gerror.New("默认在途仓库不存在")
			}
			toWarehouse := warehouseList.WarehouseList[0].WarehouseName
			toParentWarehouse = toWarehouse

			transferReceipt, err := s.CreateInnerTransferReceipt(ctx, &material_transfer.MaterialTransferReq{
				FromCompanyAbbr: fromParentCompany.Abbr,
				FromBranch:      fromParentBranch,
				ToCompanyAbbr:   toParentCompany.Abbr,
				ToBranch:        toParentBranch,
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
				FromCompanyAbbr: fromParentCompany.Abbr,
				ToCompanyAbbr:   toParentCompany.Abbr,
			}
		}
		{
			//step3 调入方父公司发起销售订单，目标是调入方公司
			transferReceipt, err := s.CreateInnerTransferReceipt(ctx, &material_transfer.MaterialTransferReq{
				FromCompanyAbbr: toParentCompany.Abbr,
				FromBranch:      toParentBranch,
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
				FromCompanyAbbr: toParentCompany.Abbr,
				ToCompanyAbbr:   req.ToCompanyAbbr,
			}
		}
	}

	return transferResp, nil
}

// CreateInnerTransferReceipt  实际上是通过 创建内部销售单 -> 内部采购单来实现
func (s *sMaterialTransfer) CreateInnerTransferReceipt(ctx context.Context, req *material_transfer.MaterialTransferReq) (*material_transfer.TransferReceipt, error) {
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

	customers, err := service.Selling().ListCustomers(ctx, &dtoSelling.ListCustomersReq{
		RepresentsCompany: toCompanyName,
		PageSize:          1,
	})
	if err != nil || len(customers) == 0 {
		return nil, gerror.Wrapf(err, "获取调出方供应商的交易对象失败")
	}
	containsCompany := false
	for _, companyItem := range customers[0].Companies {
		if companyItem.Company == toCompanyName {
			// 调出方公司已在调入方交易对象中
			containsCompany = true
			break
		}
	}
	if !containsCompany {
		// 调出方公司不在调入方交易对象中，默认添加
		err = service.Selling().AddCompanyToCustomer(ctx, customers[0], fromCompanyName)
		if err != nil {
			return nil, gerror.Wrapf(err, "添加调出方公司到调入方客户交易对象失败")
		}
	}

	suppliers, err := service.Supplier().ListSuppliers(ctx, &buying.ListSuppliersReq{
		RepresentsCompany: fromCompanyName,
		PageSize:          1,
	})
	if err != nil || len(suppliers.Suppliers) == 0 {
		return nil, gerror.Wrapf(err, "获取调出方供应商的交易对象失败")
	}
	//检查调入方父级公司的内部供应商的交易对象是否包含了调出方公司，如果没有默认添加
	err = service.Supplier().AddSupplerTransactCompany(ctx, &dto.AddSupplerTransactCompanyReq{
		Supplier:        suppliers.Suppliers[0].Name,
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
	//调出方发起销售订单，

	// 获取默认采购价格表
	defaultPriceList, err := service.PosPriceList().GetPosPriceListByCompany(ctx, fromCompanyName)
	if err != nil {
		g.Log().Warningf(ctx, "获取采购价格表失败，company: %s", fromCompanyName)
		defaultPriceList, err = service.PosPriceList().GetDefaultPosPriceList(ctx)
		if err != nil {
			return nil, gerror.Wrapf(err, "获取默认采购价格表失败")
		}
	}

	saleOrder, err := service.Selling().CreateSalesOrder(ctx, &dtoSelling.SalesOrder{
		Customer:         customers[0].Name,
		Company:          fromCompanyName,
		DeliveryDate:     consts.DefaultDeliveryDate,
		SetWarehouse:     req.FromWarehouse,
		Items:            saleOrderItems,
		SellingPriceList: defaultPriceList.Name,
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
		ScheduleDate:    consts.DefaultRequiredByDate,
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
