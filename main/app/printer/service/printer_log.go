package service

import (
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
	"go.uber.org/zap"
)

// IPrinterLogSrv 定义打印日志服务接口
type IPrinterLogSrv interface {
	AddLog(ctx context.Context, printer resp.PrinterInfo, printerLogData model.PrinterLog, controlDeviceId string) (model.PrinterLog, error) // 添加打印日志
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
func (s *printerLogSrv) AddLog(ctx context.Context, printer resp.PrinterInfo, printerLogData model.PrinterLog, controlDeviceId string) (model.PrinterLog, error) {
	var err error
	// 标记进行中
	printerLogData.Status = constant.PrinterLogStatusInProgress

	// // 如果是局域网部署 - 就都下放打印
	// if viper.GetBool("IS_CLOUD_DEPLOY") {
	// 	printerLogData.Type = constant.PrinterLogTypeCloud
	// }

	// 获取商家设置，判断是否开启本地打印
	companySetting := ctx.GetCompanySetting()

	// 如果是商米云打印 - 就都队列打印
	if printer.PrinterType == constant.PrinterTypeSunmiCloud && companySetting.IsOpenLocalPrint == 0 {
		printerLogData.Type = constant.PrinterLogTypeDefault
	}

	// 打印机不存在
	if printer.PrinterType == "" {
		printerLogData.Status = constant.PrinterLogStatusEnd
		printerLogData.Reason = "打印机不存在"
	}

	// 保存数据
	printerLogData.PrinterTime = time.Now().Unix()
	printerLogRepo := repository.NewPrinterLogRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	var printerLog model.PrinterLog
	copier.Copy(&printerLog, printerLogData)
	printerLog, err = printerLogRepo.Create(printerLog)
	if err != nil {
		logger.Logger.Error("保存数据打印日志失败", zap.Error(err))
		return model.PrinterLog{}, errors.WithMessage(err)
	}

	// 只保留7天的数据
	go func() {
		err = printerLogRepo.UpdateByWhere(map[string]any{"delete_time": time.Now().Unix()}, printerLogRepo.WhereCreatedBefore(7))
		if err != nil {
			logger.Logger.Error("删除n天前的打印日志失败", zap.Error(err))
		}
	}()

	return printerLog, nil
}
