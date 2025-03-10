package service

import (
	"fmt"
	"slices"
	"strconv"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"github.com/jinzhu/copier"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// IPrinterLogSrv 定义打印日志服务接口
type IPrinterLogSrv interface {
	AddLog(ctx context.Context, printer resp.PrinterInfo, printerLogData resp.PrinterLogData, controlDeviceId string) (resp.PrinterLogData, error) // 添加打印日志
}

type printerLogSrv struct {
	dbm        *database.DBManager
	settingSrv setting.ISrv
}

func NewPrinterLogSrv(dbm *database.DBManager, settingSrv setting.ISrv) IPrinterLogSrv {
	return NewPrinterLogSrvImpl(dbm, settingSrv)
}

func NewPrinterLogSrvImpl(dbm *database.DBManager, settingSrv setting.ISrv) IPrinterLogSrv {
	return &printerLogSrv{
		dbm:        dbm,
		settingSrv: settingSrv,
	}
}

// AddLog 添加打印日志
func (s *printerLogSrv) AddLog(ctx context.Context, printer resp.PrinterInfo, printerLogData resp.PrinterLogData, controlDeviceId string) (resp.PrinterLogData, error) {
	// 标记进行中
	printerLogData.Status = constant.PrinterLogStatusInProgress
	var (
		respPrinterLogData resp.PrinterLogData
		err                error
	)
	// 是否本机
	isLocal := printerLogData.CashierDeviceId == controlDeviceId
	// 如果是点餐助手操作的，就永远不是本机
	if ctx.GetSource() == constant.SourceAssistant {
		isLocal = false
	}
	// 是否队列服务
	var isQueueService bool
	// 如果是局域网部署 - 就都下放打印
	if viper.GetBool("IS_CLOUD_DEPLOY") {
		printerLogData.Type = constant.PrinterLogTypeCloud
	} else if !isLocal && printerLogData.PrinterUuid > 0 {
		isQueueService = true
	}
	// 获取商家设置，判断是否开启本地打印
	companySetting := ctx.GetCompanySetting()
	// 如果是商米云打印 -  就都队列打印
	if printer.PrinterType == constant.PrinterTypeSunmiCloud && companySetting.IsOpenLocalPrint == 0 {
		isQueueService = true
		printerLogData.Type = constant.PrinterLogTypeDefault
		printerLogData.FirstExecution = 0
	}
	if printer.PrinterType == "" { // 打印机不存在
		printerLogData.Status = constant.PrinterLogStatusEnd
		printerLogData.Reason = "打印机不存在"
	} else {
		printerSetting, err := s.settingSrv.GetPrinterSetting(ctx, nil)
		if err != nil {
			return respPrinterLogData, errors.WithMessage(err)
		}
		// 记录打印方式  = 1 文本打印，2图片打印
		// 默认文本打印
		printerLogData.PrintMethod = constant.PrinterLogPrintMethodText
		// 打印方式（收银）
		printMethodStr := printerSetting.PrintMethod
		// 这几种是打印方式（送厨）
		if slices.Contains([]int{constant.PrinterLogDataTypeOneDishOneMenu, constant.PrinterLogDataTypeEntireOrder, constant.PrinterLogDataTypeReturnDish}, printerLogData.DataType) {
			printMethodStr = printerSetting.KitchenPrintMethod
		}
		printMethod, _ := strconv.Atoi(printMethodStr)
		if printMethod > 0 && slices.Contains([]int{constant.PrinterLogPrintMethodText, constant.PrinterLogPrintMethodImage}, printMethod) {
			printerLogData.PrintMethod = printMethod
		}
		if !isQueueService { // 直接打印
			if printerLogData.PrinterUuid > 0 && printerLogData.Type == constant.PrinterLogTypeDefault && printerLogData.FirstExecution == 1 {
				// ToDo 打印驱动，打印小票，返回错误
				if err := errors.WithMessage(fmt.Errorf("未知错误")); err != nil {
					printerLogData.Status = constant.PrinterLogStatusEnd
					printerLogData.Reason = err.Error()
				} else {
					printerLogData.Status = constant.PrinterLogStatusSuccess
					printerLogData.Reason = "打印成功"
				}
				printerLogData.Num = 1
			} else if printerLogData.Type == constant.PrinterLogTypeCloud && printerLogData.FirstExecution == 1 && isLocal {
				printerLogData.Status = constant.PrinterLogStatusSuccess
			}
		}
	}
	// 保存数据
	printerLogData.PrinterTime = time.Now().Unix()
	printerLogRepo := repository.NewPrinterLogRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	var printerLog model.PrinterLog
	copier.Copy(&printerLog, printerLogData)
	printerLog, err = printerLogRepo.Create(printerLog)
	if err != nil {
		return respPrinterLogData, errors.WithMessage(err)
	}
	// 只保留7天的数据
	err = printerLogRepo.UpdateByWhere(map[string]any{"delete_time": time.Now().Unix()}, printerLogRepo.WhereCreatedBefore(7))
	if err != nil {
		logger.Logger.Error("删除n天前的打印日志失败", zap.Error(err))
	}

	if printer.PrinterType != "" && isQueueService {
		// ToDo 添加队列打印
		return respPrinterLogData, nil
	} else if printer.PrinterType != "" && printerLog.FirstExecution == 1 && printerLog.Type != constant.PrinterLogTypeDefault && isLocal {
		// 返回给前端打印
		printerLogData.Uuid = printerLog.Uuid
		printerLogData.Copies = 1
		if printer.PrintCopies > 0 {
			printerLogData.Copies = printer.PrintCopies
		}
		printerLogData.PrinterType = "CASHIER"
		if printer.PrinterType != "" {
			printerLogData.PrinterType = printer.PrinterType
		}
		printerLogData.PrinterConfig = "{}"
		if printer.PrinterConfig != "" {
			printerLogData.PrinterConfig = printer.PrinterConfig
		}
		printerLogData.CreateTime = printerLog.CreateTime
		return printerLogData, nil
	} else {
		// 执行前端定时获取打印
		if !isLocal || printerLog.Status == constant.PrinterLogStatusSuccess {
			printerLogData.Uuid = printerLog.Uuid
			return printerLogData, nil
		} else {
			return printerLogData, errors.New("未知错误")
		}
	}
}
