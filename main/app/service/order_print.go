package service

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/websocket"
)

// OrderPrint 打印
func (s *orderSrv) OrderPrint(ctx context.Context, request req.OrderPrintReq, needLock bool) (*resp.PrinterData, error) {
	// 加锁
	if ctx.NoLock() {
		s.lock.LockUuid(request.SaleBillUuid)
		defer s.lock.UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询销售账单失败")
	}

	// 当不是收银端的时候，拆单不可操作结账
	if ctx.GetSource() != constant.SourceCashier && saleBill.IsSplit() {
		return nil, errors.NewWithCode(constant.CodeOrderCheckSplit, "当前订单已经拆单，请前去收银机操作")
	}

	// 获取销售账单信息
	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 更新收银员信息。当收银员信息发生变化时，才更新。且只更新收银员信息相关字段
	needUpdateCashier := false
	if saleOrder.CashierUuid != ctx.GetStaffUuid() {
		saleOrder.CashierUuid = ctx.GetStaffUuid()
		needUpdateCashier = true
	}
	staff := ctx.GetStaff()
	if saleOrder.CashierName != staff.GetUserName() {
		saleOrder.CashierName = staff.GetUserName()
		needUpdateCashier = true
	}
	if needUpdateCashier {
		if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderCashier(ctx, saleOrder.Uuid, saleOrder.CashierUuid, saleOrder.CashierName); err != nil {
			return nil, errors.WithMessage(err)
		}
	}

	// 判断是否已支付
	printType := constant.PrinterTemplatePreBilling
	if saleOrder.IsPaid() {
		printType = constant.PrinterTemplateBilling
	}

	// 打印
	printerData, err := printer.NewPrinterRepo(ctx, request.PrintLang).PrintingStatementOrder(
		printType,
		saleBill,
		saleOrder.Uuid,
		utils.IfInt(ctx.GetSource() == constant.SourceAssistant, 0, 1),
		request.PayMethodUuid,
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 保存销售账单
	if needLock && !saleOrder.IsPaid() {
		if err := repository.NewOrderRepo(db).SetLock(saleBill.Uuid, true); err != nil {
			return nil, errors.WithMessage(err, "设置锁单失败")
		}
		// 推送桌台更新
		utils.Go(func() {
			websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_DESK, map[string]interface{}{
				"desk_uuid":   saleBill.DeskUuid,
				"update_time": time.Now().Unix(),
			})
		})
	}

	// 如果是点餐助手，不能直接打印
	if ctx.GetSource() == constant.SourceAssistant {
		return &resp.PrinterData{}, nil
	}

	return printerData, nil
}

// OrderPrintInvoice 打印发票
func (s *orderSrv) OrderPrintInvoice(ctx context.Context, req req.OrderPrintInvoiceReq) (*resp.PrinterData, error) {
	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询销售账单失败")
	}

	// 获取销售账单信息
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 更新收银员信息。当收银员信息发生变化时，才更新。且只更新收银员信息相关字段
	needUpdateCashier := false
	if saleOrder.CashierUuid != ctx.GetStaffUuid() {
		saleOrder.CashierUuid = ctx.GetStaffUuid()
		needUpdateCashier = true
	}
	staff := ctx.GetStaff()
	if saleOrder.CashierName != staff.GetUserName() {
		saleOrder.CashierName = staff.GetUserName()
		needUpdateCashier = true
	}
	if needUpdateCashier {
		if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderCashier(ctx, saleOrder.Uuid, saleOrder.CashierUuid, saleOrder.CashierName); err != nil {
			return nil, errors.WithMessage(err)
		}
	}

	// 设置发票信息
	if req.CompanyName != "" {
		// 创建发票信息对象
		invoiceInfo := model.SaleOrderInvoiceInfo{
			SaleOrderUuid:    saleOrder.Uuid,
			CompanyName:      req.CompanyName,
			CompanyAddr:      req.CompanyAddr,
			CompanyTaxNumber: req.CompanyTaxNumber,
			CompanyPhone:     req.CompanyPhone,
		}
		// 保存发票信息（不存在则创建，存在则更新）
		invoiceInfos, err := repository.NewOrderRepo(db).SaveOrUpdateInvoiceInfo(saleOrder.Uuid, invoiceInfo)
		if err != nil {
			return nil, errors.WithMessage(err, "保存发票信息失败")
		}
		// 更新内存中的发票信息
		saleOrder.InvoiceInfo = invoiceInfos
	} else {
		invoiceInfo, err := repository.NewOrderRepo(db).GetInvoiceInfo(saleOrder.Uuid)
		if err != nil {
			saleOrder.InvoiceInfo = &model.SaleOrderInvoiceInfo{}
		} else {
			saleOrder.InvoiceInfo = invoiceInfo
		}
		// 更新打印次数
		saleOrder.InvoiceInfo.PrintNum = saleOrder.InvoiceInfo.PrintNum + 1
		repository.NewOrderRepo(db).SaveOrUpdateInvoiceInfo(saleOrder.Uuid, model.SaleOrderInvoiceInfo{
			SaleOrderUuid: saleOrder.Uuid,
			PrintNum:      saleOrder.InvoiceInfo.PrintNum,
		})
	}

	// 打印
	printerData, err := printer.NewPrinterRepo(ctx, req.PrintLang).PrintingInvoice(
		saleBill,
		saleOrder.Uuid,
		utils.IfInt(ctx.GetSource() == constant.SourceAssistant, 0, 1),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 如果是点餐助手，不能直接打印
	if ctx.GetSource() == constant.SourceAssistant {
		return &resp.PrinterData{}, nil
	}

	return printerData, nil
}

// OrderPrintInvoiceInfo 获取发票信息
func (s *orderSrv) OrderPrintInvoiceInfo(ctx context.Context, req req.OrderInvoiceInfoReq) resp.SaleOrderInvoiceInfo {
	db := s.dbm.GetDB(ctx.GetDbId())
	invoiceInfo, err := repository.NewOrderRepo(db).GetInvoiceInfo(req.SaleOrderUuid)
	if err != nil {
		return resp.SaleOrderInvoiceInfo{}
	}
	return resp.SaleOrderInvoiceInfo{
		CompanyName:      invoiceInfo.CompanyName,
		CompanyAddr:      invoiceInfo.CompanyAddr,
		CompanyTaxNumber: invoiceInfo.CompanyTaxNumber,
		CompanyPhone:     invoiceInfo.CompanyPhone,
	}
}
