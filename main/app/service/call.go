package service

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"github.com/jinzhu/copier"
	"go.uber.org/zap"
)

// ICallSrv 定义呼叫服务接口
type ICallSrv interface {
	GetUnprocessedCallList(companyUuid uint64, listReq req.UnprocessedCallListReq) (resp.UnprocessedCallList, error) // 获取未处理的呼叫列表
	GetAbnormalPrintList(companyUuid uint64, soldOutReq req.AbnormalPrintListReq) (resp.AbnormalPrintList, error)    // 异常打印列表
	Processed(companyUuid uint64, callUuid uint64) error                                                             // 呼叫已处理
	DeletePrint(companyUuid uint64, printLogUuid uint64) error                                                       // 打印删除
	GetUnprocessed(companyUuid uint64) (resp.UnprocessedResp, error)                                                 // 获取未处理消息数量
	GetUnprocessedNotice(ctx context.Context) (resp.UnprocessedListResp, error)                                      // 获取未处理的通知
	Call(ctx context.Context, callReq req.CallReq) error                                                             // 发起呼叫
}

// callSrv 呼叫服务结构体
type callSrv struct {
	dbm *database.DBManager // 数据库管理

}

func NewCallSrv(dbm *database.DBManager) ICallSrv {
	return NewCallSrvImpl(dbm)
}

func NewCallSrvImpl(dbm *database.DBManager) ICallSrv {
	return &callSrv{
		dbm: dbm,
	}
}

// GetUnprocessedCallList 获取未处理的呼叫列表
func (s *callSrv) GetUnprocessedCallList(companyUuid uint64, listReq req.UnprocessedCallListReq) (resp.UnprocessedCallList, error) {
	var res resp.UnprocessedCallList
	callRepo := repository.NewCallRepo(s.dbm.GetDB(companyUuid))
	calls, total, err := callRepo.PaginateGet(listReq.PageNo, listReq.PageSize,
		callRepo.WhereC1Status(constant.CallStatusUnprocessed), callRepo.WhereC2IsNull())
	if err != nil {
		return res, errors.WithMessage(err, "获取呼叫列表失败")
	}
	callItems := make([]resp.UnprocessedCallItem, 0, len(calls))
	for _, call := range calls {
		var item resp.UnprocessedCallItem
		copier.Copy(&item, call)
		callItems = append(callItems, item)
	}
	return resp.UnprocessedCallList{
		List: callItems,
		Meta: dto.PageResponse{
			PageNo:   listReq.PageNo,
			PageSize: listReq.PageSize,
			Total:    total,
		},
	}, nil
}

// GetAbnormalPrintList 获取异常打印列表
func (s *callSrv) GetAbnormalPrintList(companyUuid uint64, listReq req.AbnormalPrintListReq) (resp.AbnormalPrintList, error) {
	var res resp.AbnormalPrintList
	printerLogRepo := repository.NewPrinterLogRepo(s.dbm.GetDB(companyUuid))
	printerLogs, total, err := printerLogRepo.PaginateGet(
		listReq.PageNo,
		listReq.PageSize,
		printerLogRepo.WhereStatus(constant.PrinterLogStatusEnd),
		printerLogRepo.WhereType(constant.PrinterLogTypeDefault),
		printerLogRepo.WithPrinter(),
		printerLogRepo.WithSaleOrder(),
		printerLogRepo.WithSaleBill(),
	)
	if err != nil {
		return res, errors.WithMessage(err, "获取呼叫列表失败")
	}
	abnormalPrintItems := make([]resp.AbnormalPrintItem, 0, len(printerLogs))
	for _, printerLog := range printerLogs {
		var printerName, deskNo string
		if printerLog.Printer != nil {
			printerName = printerLog.Printer.Name
		}
		if printerLog.SaleBill != nil || printerLog.SaleOrder != nil {
			if printerLog.SaleBill != nil {
				deskNo = printerLog.SaleBill.SerialNo
			} else if printerLog.SaleOrder != nil {
				deskNo = printerLog.SaleOrder.SaleBill.SerialNo
			}
		}
		var item resp.AbnormalPrintItem
		copier.Copy(&item, printerLog)

		item.PrinterName = printerName
		item.DeskNo = deskNo
		abnormalPrintItems = append(abnormalPrintItems, item)
	}
	return resp.AbnormalPrintList{
		List: abnormalPrintItems,
		Meta: dto.PageResponse{
			PageNo:   listReq.PageNo,
			PageSize: listReq.PageSize,
			Total:    total,
		},
	}, nil
}

// GetUnprocessed 获取未处理消息数量
func (s *callSrv) GetUnprocessed(companyUuid uint64) (resp.UnprocessedResp, error) {
	var (
		res                  resp.UnprocessedResp
		unprocessedCallCount int64
		abnormalPrintCount   int64
		err                  error
		db                   = s.dbm.GetDB(companyUuid)
	)
	callRepo := repository.NewCallRepo(db)
	unprocessedCallCount, err = callRepo.GetUnprocessedCallCount(callRepo.WhereC1Status(constant.CallStatusUnprocessed), callRepo.WhereC2IsNull())
	if err != nil {
		return res, errors.WithMessage(errors.New("获取未处理呼叫数量失败"), err.Error())
	}
	printerLogRepo := repository.NewPrinterLogRepo(db)
	abnormalPrintCount, err = printerLogRepo.GetPrintLogCount(printerLogRepo.WhereStatus(constant.PrinterLogStatusEnd),
		printerLogRepo.WhereType(constant.PrinterLogTypeDefault), printerLogRepo.WhereFirstExecution(0))
	if err != nil {
		logger.Logger.Error("获取异常打印数量失败", zap.Error(err))
		return res, errors.WithMessage(errors.New("获取异常打印数量失败"), err.Error())
	}
	h5OrderRepo := repository.NewH5OrderRepo(db)
	unhandledH5OrderCount, err := h5OrderRepo.GetH5OrderCount(h5OrderRepo.WhereStatus([]uint{constant.H5OrderStatusOrder}), h5OrderRepo.WhereIsNeedAudit(1))
	if err != nil {
		logger.Logger.Error("获取未处理的h5订单数量失败", zap.Error(err))
		return res, errors.WithMessage(errors.New("获取未处理的h5订单数量失败"), err.Error())
	}

	memberSaleOrderRepo := repository.NewMemberSaleOrderRepo(db)
	unhandledMemberSaleOrderCount, err := memberSaleOrderRepo.GetOrderCount(memberSaleOrderRepo.WhereStatusIn([]uint{constant.MemberSaleOrderStatusPendingMerchantAccept}))
	if err != nil {
		logger.Logger.Error("获取待接单外送订单数量失败", zap.Error(err))
		return res, errors.WithMessage(errors.New("获取待接单外送订单数量失败"), err.Error())
	}

	return resp.UnprocessedResp{
		UnprocessedCallCount:        unprocessedCallCount,
		AbnormalPrintCount:          abnormalPrintCount,
		UnprocessedH5OrderCount:     unhandledH5OrderCount,
		UnprocessedMemberOrderCount: unhandledMemberSaleOrderCount,
		UpdateTime:                  time.Now().Unix(),
	}, nil
}

// GetUnprocessedNotice 获取未处理消息数量
func (s *callSrv) GetUnprocessedNotice(ctx context.Context) (resp.UnprocessedListResp, error) {
	res := resp.UnprocessedListResp{
		Call: resp.UnprocessedCall{
			List: make([]resp.UnprocessedCallItemForNotice, 0),
		},
		AbnormalPrint: resp.UnprocessedAbnormalPrint{
			List: make([]resp.AbnormalPrintItemForNotice, 0),
		},
		H5Order: resp.UnprocessedH5Order{
			List: make([]resp.UnprocessedH5OrderItem, 0),
		},
		MemberSaleOrder: resp.UnprocessedMemberSaleOrder{
			List: make([]resp.UnprocessedMemberSaleOrderItem, 0),
		},
	}
	thirtyMinutesAgo := time.Now().Add(-30 * time.Minute).Unix()
	callRepo := repository.NewCallRepo(ctx.GetDB())
	unprocessedCalls, _, err := callRepo.PaginateGet(1, 10,
		callRepo.WhereC1Status(constant.CallStatusUnprocessed),
		callRepo.WhereC1CreateTimeGt(thirtyMinutesAgo),
		callRepo.WhereC2IsNull())
	if err != nil {
		return res, errors.WithMessage(errors.New("获取未处理呼叫失败"), err.Error())
	}
	printerLogRepo := repository.NewPrinterLogRepo(ctx.GetDB())
	abnormalPrints, _, err := printerLogRepo.PaginateGet(
		1,
		10,
		printerLogRepo.WhereStatus(constant.PrinterLogStatusEnd),
		printerLogRepo.WhereType(constant.PrinterLogTypeDefault),
		printerLogRepo.WhereCreateTimeGt(thirtyMinutesAgo),
		printerLogRepo.WithPrinter(),
		printerLogRepo.WithSaleOrder(),
		printerLogRepo.WithSaleBill())
	if err != nil {
		return res, errors.WithMessage(errors.New("获取异常打印失败"), err.Error())
	}

	h5OrderRepo := repository.NewH5OrderRepo(ctx.GetDB())
	orders, _, err := h5OrderRepo.PaginateGetH5Order(1, 10, h5OrderRepo.WhereIsNeedAudit(1),
		h5OrderRepo.WhereUnNotified(), h5OrderRepo.WhereCreateTimeGt(thirtyMinutesAgo),
		repository.CommonRepo.SortWithCreateTime("desc"))

	if err != nil {
		return res, errors.WithMessage(errors.New("获取未处理的H5订单失败"), err.Error())
	}

	memberSaleOrderRepo := repository.NewMemberSaleOrderRepo(ctx.GetDB())
	memberSaleOrders, err := memberSaleOrderRepo.GetForCall(memberSaleOrderRepo.WhereUpdateTimeGt(thirtyMinutesAgo))
	if err != nil {
		return res, errors.WithMessage(errors.New("获取未处理的外送订单失败"), err.Error())
	}

	// 未处理的呼叫
	language := ctx.GetLanguage()
	textMap := map[uint8]string{
		constant.CallTypeWaiter:   "呼叫服务员",
		constant.CallTypeCheckout: "呼叫结账",
	}
	for _, call := range unprocessedCalls {
		var item resp.UnprocessedCallItemForNotice
		copier.Copy(&item, call)
		item.CallText = i18n.Translate(language, "桌位") + " " + item.DeskNo + " " + i18n.Translate(language, textMap[item.CallType])
		res.Call.List = append(res.Call.List, item)
	}

	// 异常打印
	abnormalPrintItems := make([]resp.AbnormalPrintItemForNotice, 0, len(abnormalPrints))
	for _, printerLog := range abnormalPrints {
		var (
			deskNo string
			item   resp.AbnormalPrintItemForNotice
		)
		if printerLog.SaleBill != nil || printerLog.SaleOrder != nil {
			if printerLog.SaleBill != nil {
				deskNo = printerLog.SaleBill.SerialNo
			} else if printerLog.SaleOrder != nil {
				deskNo = printerLog.SaleOrder.SaleBill.SerialNo
			}
		}
		copier.Copy(&item, printerLog)
		item.DeskNo = deskNo
		abnormalPrintItems = append(abnormalPrintItems, item)
	}
	res.AbnormalPrint.List = abnormalPrintItems

	// 未处理的h5订单
	for _, order := range orders {
		res.H5Order.List = append(res.H5Order.List, resp.UnprocessedH5OrderItem{
			Uuid:         order.Uuid,
			DeskNo:       order.DeskNo,
			Status:       order.Status,
			IsAutoAccept: order.IsAutoAccept == 1,
		})
	}

	for _, memberSaleOrder := range memberSaleOrders {
		res.MemberSaleOrder.List = append(res.MemberSaleOrder.List, resp.UnprocessedMemberSaleOrderItem{
			Uuid:         memberSaleOrder.Uuid,
			Status:       memberSaleOrder.Status,
			UpdateTime:   memberSaleOrder.UpdateTime,
			IsAutoAccept: memberSaleOrder.IsAutoAccept == 1,
			CancelScene:  memberSaleOrder.CancelScene,
			SerialNumber: memberSaleOrder.SerialNumber,
		})
	}

	return res, nil
}

// Call 平板端呼叫
func (s *callSrv) Call(ctx context.Context, callReq req.CallReq) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())

	var deskUuid uint64
	var deskNo string

	// 如果 deskUuid 为 0，可能是 Kiosk 终端（无桌台），使用设备ID作为标识
	if ctx.GetDeskUuid() == 0 && ctx.GetSource() == constant.SourceKiosk {
		deskUuid = 0
		deskNo = "自助点餐机-" + ctx.GetDeviceSn() // 使用设备ID作为标识
	} else {
		deskRepo := repository.NewDeskRepo(db)
		desk, err := deskRepo.GetDesk(deskRepo.WhereUuid(ctx.GetDeskUuid()))
		if err != nil {
			return errors.New("桌台不存在")
		}
		deskUuid = desk.Uuid
		deskNo = desk.DeskNo
	}

	if err := repository.NewCallRepo(db).CreateCall(model.CustomerCall{
		DeskUuid: deskUuid,
		DeskNo:   deskNo,
		CallType: callReq.CallType,
	}); err != nil {
		return err
	}
	return nil
}

// Processed 消息已处理
func (s *callSrv) Processed(companyUuid uint64, callUuid uint64) error {
	callRepo := repository.NewCallRepo(s.dbm.GetDB(companyUuid))
	err := callRepo.Update(map[string]any{"status": constant.CallStatusProcessed},
		[]repository.DBOption{callRepo.WhereStatus(constant.CallStatusUnprocessed), callRepo.WhereDeskUuidByCallUuid(callUuid)})
	if err != nil {
		return errors.WithMessage(err, "处理呼叫失败")
	}
	return nil
}

// DeletePrint 删除打印
func (s *callSrv) DeletePrint(companyUuid uint64, printLogUuid uint64) error {
	err := repository.NewPrinterLogRepo(s.dbm.GetDB(companyUuid)).Update(printLogUuid, map[string]any{
		"delete_time": time.Now().Unix(),
	})
	if err != nil {
		return errors.WithMessage(err, "删除打印失败")
	}
	return nil
}
