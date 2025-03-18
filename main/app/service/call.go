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
			}
			deskNo = printerLog.SaleOrder.SaleBill.SerialNo
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
	)
	callRepo := repository.NewCallRepo(s.dbm.GetDB(companyUuid))
	unprocessedCallCount, err = callRepo.GetUnprocessedCallCount(callRepo.WhereC1Status(constant.CallStatusUnprocessed), callRepo.WhereC2IsNull())
	if err != nil {
		return res, errors.WithMessage(err, "获取未处理呼叫数量失败")
	}
	printerLogRepo := repository.NewPrinterLogRepo(s.dbm.GetDB(companyUuid))
	abnormalPrintCount, err = printerLogRepo.GetPrintLogCount(printerLogRepo.WhereStatus(constant.PrinterLogStatusEnd),
		printerLogRepo.WhereType(constant.PrinterLogTypeDefault), printerLogRepo.WhereFirstExecution(0))
	if err != nil {
		logger.Logger.Error("获取异常打印数量失败", zap.Error(err))
		return res, errors.WithMessage(err, "获取异常打印数量失败")
	}
	return resp.UnprocessedResp{
		UnprocessedCallCount: unprocessedCallCount,
		AbnormalPrintCount:   abnormalPrintCount,
	}, nil
}

// Call 平板端呼叫
func (s *callSrv) Call(ctx context.Context, callReq req.CallReq) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	deskRepo := repository.NewDeskRepo(db)
	desk, err := deskRepo.GetDesk(deskRepo.WhereUuid(ctx.GetDeskUuid()))
	if err != nil {
		return errors.New("桌台不存在")
	}
	if err := repository.NewCallRepo(db).CreateCall(model.CustomerCall{
		DeskUuid: desk.Uuid,
		DeskNo:   desk.DeskNo,
		CallType: callReq.CallType,
	}); err != nil {
		return errors.ErrInternal
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
