package buying

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/buying"
	dto "ttpos-bmp/app/ttpos-erp/internal/model/dto/buying"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

type sBuying struct {
}

var (
	Buying = new(sBuying)
)

func init() {
	service.RegisterBuying(Buying)
}

// CreatePurchaseFromMq 根据材料请求创建采购订单
func (s *sBuying) CreatePurchaseFromMq(ctx context.Context, req *dto.CreatePurchaseFromMqReq) (res *erp.PurchaseOrder, err error) {

	resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
		Method: erp.ApiMethodMakeMappedDoc,
	}, g.MapStrStr{
		"method":      "erpnext.stock.doctype.material_request.material_request.make_purchase_order",
		"source_name": req.SourceName,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "创建采购订单失败")
	}

	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析采购订单响应失败")
	}
	purchaseOrder := &erp.PurchaseOrder{}
	j.Get("data").Scan(purchaseOrder)
	//修改货币类型
	purchaseOrder.Currency = purchaseOrder.PriceListCurrency
	purchaseOrder.Supplier = req.Supplier
	purchaseOrder.ScheduleDate = req.RequiredBy

	//创建采购订单
	resp, err = service.Document().Create(ctx, "Purchase Order", purchaseOrder)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建采购订单失败")
	}

	// 解析响应数据
	j, err = gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析采购订单响应失败")
	}
	res = &erp.PurchaseOrder{}
	gconv.Scan(purchaseOrder, res)
	res.Name = j.Get("data.name").String()

	//提交采购订单
	_, err = service.Document().ChangeDocStatus(ctx, "Purchase Order", res.Name, 1)
	if err != nil {
		return nil, gerror.Wrapf(err, "提交采购订单失败")
	}
	return
}

// CreateInnerSaleOrderFromPurchaseOrder 创建内部销售订单
func (*sBuying) CreateInnerSaleOrderFromPurchaseOrder(ctx context.Context, req *dto.CreateInnerSaleOrderFromPurchaseOrderReq) (res *erp.SaleOrder, err error) {
	resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
		Method: erp.ApiMethodMakeMappedDoc,
	}, g.MapStrStr{
		"method":      "erpnext.buying.doctype.purchase_order.purchase_order.make_inter_company_sales_order",
		"source_name": req.SourceName,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "创建内部销售订单失败")
	}

	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析采购订单响应失败")
	}
	salesOrder := &erp.SaleOrder{}
	j.Get("data").Scan(salesOrder)
	// 发货时间
	salesOrder.DeliveryDate = req.DeliveryDate
	for _, item := range salesOrder.Items {
		item.DeliveryDate = req.DeliveryDate
	}

	//设置来源仓库
	warehouse, err := service.Warehouse().GetDefaultWarehouse(ctx, salesOrder.Company, "")
	if err != nil {
		return nil, gerror.Wrapf(err, "查询默认仓库失败")
	}
	salesOrder.SetWarehouse = warehouse.Name

	//创建采购订单
	resp, err = service.Document().Create(ctx, "Sales Order", salesOrder)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建内部销售订单失败")
	}

	// 解析响应数据
	j, err = gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析内部销售订单响应失败")
	}
	j.Get("data").Scan(salesOrder)

	// 提交订单
	_, err = service.Document().ChangeDocStatus(ctx, "Sales Order", salesOrder.Name, 1)
	if err != nil {
		return nil, gerror.Wrapf(err, "提交内部销售订单失败")
	}
	return salesOrder, nil
}

// GetPurchaseOrder 获取采购订单
func (*sBuying) GetPurchaseOrder(ctx context.Context, req *buying.GetPurchaseOrderReq) (*erp.PurchaseOrder, error) {

	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: "Purchase Order",
		Name:    req.PurchaseOrderName,
	}, nil)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询采购订单失败")
	}
	purchaseOrder := &erp.PurchaseOrder{}
	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析采购订单响应失败")
	}
	j.Get("data").Scan(purchaseOrder)
	return purchaseOrder, nil
}

// CreatePurchaseReceiptFromOrder 创建采购收货订单
func (*sBuying) CreatePurchaseReceiptFromOrder(ctx context.Context, req *buying.SavePurchaseReceiptReq) (*erp.PurchaseReceipt, error) {
	resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
		Method: erp.ApiMethodMakeMappedDoc,
	}, g.MapStrStr{
		"method":      "erpnext.buying.doctype.purchase_order.purchase_order.make_purchase_receipt",
		"source_name": req.PurchaseOrderName,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "创建内部销售订单失败")
	}

	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析采购订单响应失败")
	}

	receipt := &erp.PurchaseReceipt{}
	j.Get("data").Scan(receipt)

	//根据入参调整item
	receiptItems := make([]erp.PurchaseReceiptItem, 0)
	for _, item := range receipt.Items {
		for _, itemReq := range req.Items {
			if item.ItemCode == itemReq.ItemCode {
				item.Qty = itemReq.Qty
				receiptItems = append(receiptItems, item)
				break
			}
		}
	}
	receipt.Items = receiptItems

	//创建采购收货订单
	resp, err = service.Document().Create(ctx, "Purchase Receipt", receipt)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建采购订单失败")
	}

	// 解析响应数据
	j, err = gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析采购收货订单响应失败")
	}
	j.Get("data").Scan(receipt)

	// 提交订单
	_, err = service.Document().ChangeDocStatus(ctx, "Purchase Receipt", receipt.Name, 1)
	if err != nil {
		return nil, gerror.Wrapf(err, "提交采购收货订单失败")
	}
	return receipt, nil
}
