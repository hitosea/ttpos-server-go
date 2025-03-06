package service

import (
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
)

// ICallSrv 定义呼叫服务接口
type ICallSrv interface {
	GetUnprocessedCallList(companyUuid uint64, listReq req.UnprocessedCallListReq) (resp.UnprocessedCallList, error) // 获取未处理的呼叫列表
	GetAbnormalPrintList(companyUuid uint64, soldOutReq req.AbnormalPrintListReq) (resp.AbnormalPrintList, error)    // 异常打印列表
	Processed(companyUuid uint64, callUuid uint64) error                                                             // 呼叫已处理
	DeletePrint(companyUuid uint64, printLogUuid uint64) error                                                       // 打印删除
	Reprint(ctx context.Context, printerLogUuid uint64) (resp.ReprintResp, error)                                    // 重新打印
	GetUnprocessed(companyUuid uint64) (resp.UnprocessedResp, error)                                                 // 获取未处理消息数量
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
	printerLogs, total, err := printerLogRepo.PaginateGet(listReq.PageNo, listReq.PageSize,
		printerLogRepo.WhereStatus(constant.PrinterLogStatusEnd), printerLogRepo.WhereType(constant.PrinterLogTypeDefault),
		printerLogRepo.WithPrinter(), printerLogRepo.WithSaleBill(), printerLogRepo.WithSaleBillDesk())
	if err != nil {
		return res, errors.WithMessage(err, "获取呼叫列表失败")
	}
	abnormalPrintItems := make([]resp.AbnormalPrintItem, 0, len(printerLogs))
	for _, printerLog := range printerLogs {
		var printerName, deskNo string
		if printerLog.Printer != nil {
			printerName = printerLog.Printer.Name
		}
		if printerLog.SaleBill != nil {
			deskNo = printerLog.SaleBill.Desk.DeskNo
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

// Reprint 重新打印
func (s *callSrv) Reprint(ctx context.Context, printerLogUuid uint64) (resp.ReprintResp, error) {
	var res resp.ReprintResp
	companyUuid := ctx.GetCompanyUuid()
	deviceId := ctx.GetDeviceSn()
	printerLogRepo := repository.NewPrinterLogRepo(s.dbm.GetDB(companyUuid))
	printerLog := printerLogRepo.GetPrinterLog(printerLogRepo.WhereUuid(printerLogUuid),
		printerLogRepo.WithPrinter(), printerLogRepo.WithPrinterPrinterType())
	if printerLog.PrinterUuid > 0 && printerLog.Printer == nil && printerLog.Printer.IsDelete() {
		return res, errors.New("打印失败，打印机已不存在")
	}
	if printerLog.Data == "" {
		return res, errors.New("打印失败，打印机已不存在")
	}
	if printerLog.PrinterUuid > 0 && printerLog.Type == constant.PrinterLogTypeDefault {
		// todo 调用打印接口，实例化驱动对象传参 打印内容，打印机类型，打印机配置，打印份数
	} else if printerLog.Type == constant.PrinterLogTypeCloud && printerLog.CashierDeviceId != "" && printerLog.CashierDeviceId != deviceId {
		err := s.dbm.GetDB(companyUuid).Transaction(func(tx *gorm.DB) error {
			if err := repository.NewPrinterReadLogRepo(tx).
				Update(deviceId, map[string]any{"delete_time": time.Now().Unix()}); err != nil {
				return errors.WithMessage(err)
			}
			if err := repository.NewPrinterLogRepo(tx).Update(printerLogUuid, map[string]any{"status": constant.PrinterLogStatusInProgress}); err != nil {
				return errors.WithMessage(err)
			}
			return nil
		})
		if err != nil {
			return res, errors.WithMessage(err, "打印失败")
		}
		res = resp.ReprintResp{PrinterLogUuid: printerLogUuid}
	} else {
		var printerName, printerType, printerConfig string
		if printerLog.Printer != nil {
			printerName = printerLog.Printer.Name
			if printerLog.Printer.PrinterType != nil {
				printerType = printerLog.Printer.PrinterType.Name
			}
			printerConfig = printerLog.Printer.ConfigJson
		}
		res = resp.ReprintResp{
			Data:          printerLog.Data,
			PrintMethod:   printerLog.PrintMethod,
			PrinterUuid:   printerLog.PrinterUuid,
			PrinterTime:   time.Now().Unix(),
			PrinterName:   printerName,
			PrinterType:   printerType,
			PrinterConfig: printerConfig,
			PrintTimes:    1,
		}
	}
	return res, nil
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
