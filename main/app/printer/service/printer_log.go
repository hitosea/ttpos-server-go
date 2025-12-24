package service

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"ttpos-server-go/app/printer/printer_tasks"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IPrinterLogSrv 定义打印日志服务接口
type IPrinterLogSrv interface {
	AddLog(ctx context.Context, printer resp.PrinterInfo, printerLogData model.PrinterLog, controlDeviceId string) (model.PrinterLog, error) // 添加打印日志
	GetPrinterBase(ctx context.Context) (*resp.PrinterBaseResp, error)                                                                       // 获取基础数据
	GetPrinterLogList(ctx context.Context, req req.PrinterListReq) (*resp.PrinterListPaginationResp, error)                                  // 获取打印列表
	GetPrinterData(ctx context.Context) (*resp.PrinterDataList, error)                                                                       // 获取打印数据
	PrinterPrint(ctx context.Context, req req.PrinterPrintReq) (*resp.PrinterData, error)                                                    // 打印
	PrinterReport(ctx context.Context, req req.PrinterReportReqs) error                                                                      // 打印报告
	GetStaticOpenCashBoxPrinterConfig(ctx context.Context) (*resp.PrinterData, error)                                                        // 获取静态打印机配置
	GetOldOrderPrinterConfig(ctx context.Context, data string) (*resp.PrinterData, error)                                                    // 获取旧订单打印配置
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

// GetPrinterBase 获取打印查询条件数据
func (s *printerLogSrv) GetPrinterBase(ctx context.Context) (*resp.PrinterBaseResp, error) {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	// 获取打印机列表
	printerLists, err := base.NewPrinterRepo(db).GetPrinterList()
	if err != nil {
		logger.Logger.Error("获取打印机列表失败", zap.Error(err))
		return nil, err
	}
	printerList := make([]resp.PrinterBase, 0, len(printerLists))
	for _, printer := range printerLists {
		printerList = append(printerList, resp.PrinterBase{
			Uuid: printer.Uuid,
			Name: printer.Name,
		})
	}
	//
	language := ctx.GetLanguage()
	printerTypes := make([]resp.PrinterBase, 0)
	printerTypes = append(printerTypes, resp.PrinterBase{
		Uuid: constant.PrinterTemplateHandoverSheet,
		Name: i18n.Translate(language, "交班单"),
	})
	printerTypes = append(printerTypes, resp.PrinterBase{
		Uuid: constant.PrinterTemplateBilling,
		Name: i18n.Translate(language, "结账单"),
	})
	printerTypes = append(printerTypes, resp.PrinterBase{
		Uuid: constant.PrinterTemplatePreBilling,
		Name: i18n.Translate(language, "预结账单"),
	})
	printerTypes = append(printerTypes, resp.PrinterBase{
		Uuid: constant.PrinterTemplateOneDishOneMenu,
		Name: i18n.Translate(language, "一菜一单"),
	})
	printerTypes = append(printerTypes, resp.PrinterBase{
		Uuid: constant.PrinterTemplateBusiness,
		Name: i18n.Translate(language, "营业数据"),
	})
	printerTypes = append(printerTypes, resp.PrinterBase{
		Uuid: constant.PrinterTemplateEntireOrder,
		Name: i18n.Translate(language, "整单打印"),
	})
	printerTypes = append(printerTypes, resp.PrinterBase{
		Uuid: constant.PrinterTemplateInvoice,
		Name: i18n.Translate(language, "打印发票"),
	})
	printerTypes = append(printerTypes, resp.PrinterBase{
		Uuid: constant.PrinterTemplateRecharge,
		Name: i18n.Translate(language, "充值单"),
	})
	printerTypes = append(printerTypes, resp.PrinterBase{
		Uuid: constant.PrinterTemplateReturnDish,
		Name: i18n.Translate(language, "退菜单"),
	})
	//
	return &resp.PrinterBaseResp{
		PrinterList:  printerList,
		PrinterTypes: printerTypes,
	}, nil
}

// GetPrinterLogList 获取打印列表
func (s *printerLogSrv) GetPrinterLogList(ctx context.Context, req req.PrinterListReq) (*resp.PrinterListPaginationResp, error) {
	// 获取打印日志
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	printerLogRepo := repository.NewPrinterLogRepo(db)

	// 准备查询选项
	queryOpts := []repository.DBOption{
		printerLogRepo.WithPrinter(),
		printerLogRepo.WithPrinterPrinterType(),
		printerLogRepo.WithSaleOrder(),
		printerLogRepo.WithSaleBill(),
		printerLogRepo.WithProductPrinter(),
		printerLogRepo.WithMemberRechargeOrder(),
	}

	// 添加时间范围查询
	if req.QueryStartTime > 0 || req.QueryEndTime > 0 {
		queryOpts = append(queryOpts, printerLogRepo.WhereTimeRange(req.QueryStartTime, req.QueryEndTime))
	}

	// 添加状态查询
	if req.Status != -1 {
		// 状态, -1=全都、0=失败, 1=成功, 2=补打成功, 3=补打失败
		// 补打成功： Status = 2 && num > 0
		// 补打失败： Status = 0 && num > 0
		switch req.Status {
		case 0: // 失败
			queryOpts = append(queryOpts, func(db *gorm.DB) *gorm.DB {
				return db.Where("status = ?", 0)
			})
		case 1: // 成功
			queryOpts = append(queryOpts, func(db *gorm.DB) *gorm.DB {
				return db.Where("status = ?", 2)
			})
		case 2: // 补打成功
			queryOpts = append(queryOpts, func(db *gorm.DB) *gorm.DB {
				return db.Where("status = ? AND num > 1", 2)
			})
		case 3: // 补打失败
			queryOpts = append(queryOpts, func(db *gorm.DB) *gorm.DB {
				return db.Where("status = ? AND num > 1", 0)
			})
		}
	}

	// 添加打印机查询
	if req.PrinterUuid > 0 {
		queryOpts = append(queryOpts, func(db *gorm.DB) *gorm.DB {
			return db.Where("printer_uuid = ?", req.PrinterUuid)
		})
	}

	// 添加数据类型查询
	if req.DataType != -1 {
		queryOpts = append(queryOpts, func(db *gorm.DB) *gorm.DB {
			return db.Where("data_type = ?", req.DataType)
		})
	}

	// 添加排序
	queryOpts = append(queryOpts, func(db *gorm.DB) *gorm.DB {
		return db.Order("create_time desc")
	})

	// 只查询某些字段
	queryOpts = append(queryOpts, func(db *gorm.DB) *gorm.DB {
		return db.Select("uuid, product_printer_uuid, related_uuid, data_type, printer_time, print_method, first_execution,  status, num, printer_uuid, reason, create_time")
	})

	// 排除数据管理
	companySetting := ctx.GetCompanySetting()
	dataSetting := s.settingSrv.GetDataManageSetting(ctx)
	excludeDataManage := companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
	if excludeDataManage {
		dataManageUuid := []uint64{}
		db.Model(&model.DataManage{}).Where("type = ? AND delete_time = 0", model.DataManageTypeOrder).Pluck("data_uuid", &dataManageUuid)
		saleOrderUuid := []uint64{}
		db.Model(&model.SaleOrder{}).Where("sale_bill_uuid IN ? AND delete_time = 0", dataManageUuid).Pluck("uuid", &saleOrderUuid)
		relatedUuid := append(dataManageUuid, saleOrderUuid...)
		if len(relatedUuid) > 0 {
			queryOpts = append(queryOpts, func(db *gorm.DB) *gorm.DB {
				return db.Where("related_uuid NOT IN ?", relatedUuid)
			})
		}
	}

	// 执行查询
	printerLogList, total, err := printerLogRepo.PaginateGet(
		req.PageNo,
		req.PageSize,
		queryOpts...,
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 转换为响应数据
	language := ctx.GetLanguage()
	results := make([]resp.PrinterLogData, 0, len(printerLogList))
	for _, log := range printerLogList {
		results = append(results, resp.PrinterLogData{
			Uuid: log.Uuid,
			RuleName: func() string {
				if log.ProductPrinter == nil {
					return "-"
				}
				return log.ProductPrinter.Name
			}(),
			SerialNo: func() string {
				if log.SaleBill == nil && log.SaleOrder == nil {
					return "-"
				}
				if log.SaleBill != nil {
					return log.SaleBill.SerialNo
				}
				return log.SaleOrder.SaleBill.SerialNo
			}(),
			OrderNo: func() string {
				if log.SaleOrder == nil && log.SaleBill == nil && log.MemberRechargeOrder == nil {
					return "-"
				}
				if log.MemberRechargeOrder != nil {
					return log.MemberRechargeOrder.OrderNo
				}
				if log.SaleBill != nil {
					return log.SaleBill.OrderNo
				}
				return log.SaleOrder.OrderNo
			}(),
			DataTypeName: func() string {
				switch log.DataType {
				case constant.PrinterTemplateHandoverSheet:
					return i18n.Translate(language, "交班单")
				case constant.PrinterTemplateBilling:
					return i18n.Translate(language, "结账单")
				case constant.PrinterTemplatePreBilling:
					return i18n.Translate(language, "预结账单")
				case constant.PrinterTemplateOneDishOneMenu:
					return i18n.Translate(language, "一菜一单")
				case constant.PrinterTemplateBusiness:
					return i18n.Translate(language, "营业数据")
				case constant.PrinterTemplateEntireOrder:
					return i18n.Translate(language, "整单打印")
				case constant.PrinterTemplateInvoice:
					return i18n.Translate(language, "打印发票")
				case constant.PrinterTemplateRecharge:
					return i18n.Translate(language, "充值单")
				case constant.PrinterTemplateReturnDish:
					return i18n.Translate(language, "退菜单")
				default:
					return "-"
				}
			}(),
			PrinterName: func() string {
				if log.Printer != nil {
					return log.Printer.Name
				}
				return "-"
			}(),
			CreateTime: log.CreateTime,
			Status:     log.Status,
			StatusText: func() string {
				switch log.Status {
				case 0:
					if log.Num > 1 {
						return i18n.Translate(language, "补打失败")
					}
					return i18n.Translate(language, "失败")
				case 1:
					return i18n.Translate(language, "进行中")
				case 2:
					if log.Num > 1 {
						return i18n.Translate(language, "补打成功")
					}
					return i18n.Translate(language, "成功")
				default:
					return "-"
				}
			}(),
			PrinterTime: log.PrinterTime,
			Reason: func() string {
				if log.Status == 0 {
					return log.Reason
				}
				return ""
			}(),
		})
	}

	return &resp.PrinterListPaginationResp{
		List: results,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

// GetPrinterData 获取打印数据
func (s *printerLogSrv) GetPrinterData(ctx context.Context) (*resp.PrinterDataList, error) {
	companyUuid := ctx.GetCompanyUuid()

	//禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(companyUuid)
		defer lock.NewSystemLock().UnlockUuid(companyUuid)
		ctx.AddLock()
	}

	printerSetting, err := s.settingSrv.GetPrinterSetting(ctx, nil)
	if err != nil {
		return nil, errors.WithMessage(errors.New("获取打印设置失败"), err.Error())
	}

	// 获取设备 - 判断网页版的设备 不能获取
	deviceRepo := repository.NewDeviceRepo(s.dbm.GetDB(companyUuid))
	device, errDevice := deviceRepo.GetDevice(deviceRepo.WhereSource(ctx.GetSource()), deviceRepo.WhereSn(ctx.GetDeviceSn()))
	if errDevice != nil {
		return nil, errors.WithMessage(errDevice, "deviceRepo.GetDevice failed")
	}
	if device.IsDelete() {
		return nil, errors.NewWithCode(constant.CodeParamError, "设备不存在")
	}
	// 网页版设备不能获取打印数据
	if device.Platform == 0 {
		return &resp.PrinterDataList{List: []resp.PrinterData{}}, nil
	}

	// 获取打印日志
	printerLogRepo := repository.NewPrinterLogRepo(s.dbm.GetDB(companyUuid))
	printerLogList, err := printerLogRepo.GetPrinterData(ctx.GetDeviceSn())
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 是否开启自定义打印联数
	enableCustomCopies := printerSetting.EnableCustomCopies == "1"

	// 转换为响应数据
	printerDataList := make([]resp.PrinterData, 0, len(printerLogList))
	for _, log := range printerLogList {
		// 如果打印类型为结账单，且启用了自定义打印联数，且打印份数为0，则不打印
		if log.DataType == 2 && enableCustomCopies && log.Copies == 0 {
			continue
		}
		printerDataList = append(printerDataList, resp.PrinterData{
			Uuid: log.Uuid,
			Data: func() string {
				if log.PrinterLogData == nil {
					if log.Data != "-" && log.Data != "" {
						return log.Data
					}
					return ""
				}
				return log.PrinterLogData.GetData(true)
			}(),
			PrintMethod: log.PrintMethod,
			Copies:      log.GetCopies(),
			PrinterType: log.PrinterType,
			PrinterConfig: func() string {
				if log.Printer == nil {
					return ""
				}
				configJson, err := json.Marshal(log.Printer.GetConfigJson())
				if err != nil {
					return ""
				}
				return string(configJson)
			}(),
			IsCashierPrinter: log.IsCashierPrinter(),
			IsUsbPrinter:     log.IsUsbPrinter(),
			PrintingTime:     log.PrintingTime,
			EnableStatusCheck: func() int {
				if log.Printer == nil {
					return 0
				}
				return log.Printer.EnableStatusCheck
			}(),
			TradeNo:        log.GetTradeNo(companyUuid),
			PrintChunkSize: log.GetPrintChunkSize(),
		})
	}

	return &resp.PrinterDataList{List: printerDataList}, nil
}

// PrinterPrint 根据打印日志ID进行打印
func (s *printerLogSrv) PrinterPrint(ctx context.Context, req req.PrinterPrintReq) (*resp.PrinterData, error) {
	// 获取打印日志仓库
	printerLogRepo := repository.NewPrinterLogRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	printerLog := printerLogRepo.GetPrinterLog(
		printerLogRepo.WhereUuid(req.Uuid),
		printerLogRepo.WithPrinter(),
		printerLogRepo.WithPrinterLogData(),
	)
	if printerLog.PrinterUuid > 0 && (printerLog.Printer == nil || printerLog.Printer.IsDelete()) {
		return nil, errors.New("打印失败，打印机已不存在")
	}
	if printerLog.Data == "" {
		return nil, errors.New("打印失败，打印数据为空")
	}
	if printerLog.PrinterType == "" {
		return nil, errors.New("打印失败，打印机类型为空")
	}

	// 去除开箱指令
	if printerLog.PrinterLogData != nil {
		data := printerLog.PrinterLogData.DecompressData()
		dataLen := len(data)
		if dataLen >= 10 {
			if strings.HasSuffix(data, "1b700019fa") {
				data = data[:dataLen-10]
			} else if strings.HasSuffix(data, "1014010001") {
				data = data[:dataLen-10]
			}
		}
		printerLog.PrinterLogData.Data = data
	} else if printerLog.Data != "-" && printerLog.Data != "" {
		data := printerLog.DecompressData()
		dataLen := len(data)
		if dataLen >= 10 {
			if strings.HasSuffix(data, "1b700019fa") {
				data = data[:dataLen-10]
			} else if strings.HasSuffix(data, "1014010001") {
				data = data[:dataLen-10]
			}
		}
		printerLog.Data = data
	}

	return &resp.PrinterData{
		Uuid: req.Uuid,
		Data: func() string {
			if printerLog.PrinterLogData != nil {
				return printerLog.PrinterLogData.CompressData()
			} else if printerLog.Data != "-" && printerLog.Data != "" {
				return printerLog.CompressData()
			}
			return ""
		}(),
		PrintMethod:      printerLog.PrintMethod,
		PrinterType:      printerLog.PrinterType,
		IsCashierPrinter: printerLog.IsCashierPrinter(),
		IsUsbPrinter:     printerLog.IsUsbPrinter(),
		Copies:           printerLog.GetCopies(),
		PrinterConfig: func() string {
			if printerLog.Printer == nil {
				return ""
			}
			configJson, err := json.Marshal(printerLog.Printer.GetConfigJson())
			if err != nil {
				return ""
			}
			return string(configJson)
		}(),
		PrintingTime: printerLog.PrintingTime,
		EnableStatusCheck: func() int {
			if printerLog.Printer == nil {
				return 0
			}
			return printerLog.Printer.EnableStatusCheck
		}(),
		TradeNo:        printerLog.GetRandomTradeNo(),
		PrintChunkSize: printerLog.GetPrintChunkSize(),
	}, nil
}

// AddLog 添加打印日志
func (s *printerLogSrv) AddLog(ctx context.Context, printer resp.PrinterInfo, printerLogData model.PrinterLog, controlDeviceId string) (model.PrinterLog, error) {
	companyUuid := ctx.GetCompanyUuid()
	// 标记进行中
	printerLogData.Status = constant.PrinterLogStatusInProgress
	// 获取商家设置，判断是否开启本地打印
	companySetting := ctx.GetCompanySetting()

	// 如果是商米云打印 - 就都队列打印
	if printer.PrinterType == constant.PrinterTypeSunmiCloud && companySetting.IsOpenLocalPrint == 0 {
		printerLogData.Type = constant.PrinterLogTypeDefault
		printerLogData.FirstExecution = 0
	}

	// 打印机不存在
	if printer.PrinterType == "" {
		printerLogData.Status = constant.PrinterLogStatusEnd
		printerLogData.Reason = "打印机不存在"
	}

	// 开启事务
	tx := s.dbm.GetDB(companyUuid).Begin()
	if tx.Error != nil {
		return model.PrinterLog{}, errors.WithMessage(tx.Error)
	}

	// 保存打印日志数据
	var LogData model.PrinterLogData
	LogData.LogUuid, _ = utils.GetID()
	LogData.SetData(companyUuid, printerLogData.Data)
	LogData, err := repository.NewPrinterLogDataRepo(tx).Create(LogData)
	if err != nil {
		tx.Rollback()
		logger.Logger.Error("保存数据打印日志数据失败", zap.Error(err))
		return model.PrinterLog{}, errors.WithMessage(err)
	}

	// 保存日志
	var printerLog model.PrinterLog
	copier.Copy(&printerLog, printerLogData)
	printerLog.BaseModel.Uuid = LogData.LogUuid
	printerLog.PrinterType = printer.PrinterType
	printerLog.PrinterTime = time.Now().Unix()
	printerLog.PrintingTime = printerLog.CalculationTime(printerLogData.Data)
	printerLog.Data = "-"
	printerLog, err = repository.NewPrinterLogRepo(tx).Create(printerLog)
	if err != nil {
		tx.Rollback()
		logger.Logger.Error("保存数据打印日志失败", zap.Error(err))
		return model.PrinterLog{}, errors.WithMessage(err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		logger.Logger.Error("提交事务失败", zap.Error(err))
		return model.PrinterLog{}, errors.WithMessage(err)
	}

	//关联打印日志数据
	printerLog.Data = LogData.GetData(printerLog.FirstExecution == 1)
	printerLog.PrinterLogData = &LogData

	// 进行队列打印
	if viper.GetString("CHECK_PRINT") == "false" || printerLog.Type == constant.PrinterLogTypeDefault {
		utils.Go(func() {
			if printerLog.Printer == nil {
				printerLogRepo := repository.NewPrinterLogRepo(s.dbm.GetDB(companyUuid))
				printerLog.Printer = printerLogRepo.GetPrinter(printerLogRepo.WhereUuid(printerLog.PrinterUuid))
			}
			printer_tasks.NewPrinterTask(s.dbm, cache.Global).ExecutePrinter(companyUuid, printerLog)
		})
	}

	return printerLog, nil
}

// PrinterPrint 获取静态打开钱箱配置
func (s *printerLogSrv) GetStaticOpenCashBoxPrinterConfig(ctx context.Context) (*resp.PrinterData, error) {
	// 获取打印设置
	printerSetting, err := s.settingSrv.GetPrinterSetting(ctx, nil)
	if err != nil {
		logger.Logger.Error("获取打印机设置失败", zap.Error(err))
		fmt.Println("获取打印机设置失败", zap.Error(err))
	}

	// 获取打印设置
	settingPrinterInfo, err := s.settingSrv.GetPrinterInfo(ctx, printerSetting, ctx.GetDeviceSn())
	if err != nil {
		return nil, errors.WithMessage(err, "获取打印设置失败")
	}

	data := "1014010001"
	// 商米打印机
	if slices.Contains([]string{
		constant.PrinterTypeSunmiLan,
		constant.PrinterTypeSunmiCloud,
		constant.PrinterTypeCashierSunmi,
	}, settingPrinterInfo.PrinterType) {
		data = "1014010001"
	}

	// 如果是收银打印机 - 不用返回
	if settingPrinterInfo.IsCashierPrinter {
		return &resp.PrinterData{}, nil
	}

	//
	return &resp.PrinterData{
		Uuid:              0,
		Data:              data,
		PrintMethod:       1,
		Copies:            settingPrinterInfo.Copies,
		PrinterType:       settingPrinterInfo.PrinterType,
		PrinterConfig:     settingPrinterInfo.PrinterConfig,
		IsCashierPrinter:  settingPrinterInfo.IsCashierPrinter,
		PrintingTime:      200,
		IsUsbPrinter:      settingPrinterInfo.IsUsbPrinter,
		EnableStatusCheck: settingPrinterInfo.EnableStatusCheck,
		TradeNo:           fmt.Sprintf("%d%s", time.Now().Unix(), hex.EncodeToString([]byte(strings.Replace(uuid.New().String(), "-", "", -1)))[:8]),
		PrintChunkSize:    4096 * 2, // 4KB
	}, nil
}

// PrinterReport 打印报告
func (s *printerLogSrv) PrinterReport(ctx context.Context, req req.PrinterReportReqs) error {
	if err := req.Validate(); err != nil {
		return err
	}
	// 获取现有的打印日志
	printerLogRepo := repository.NewPrinterLogRepo(ctx.GetDB())
	existingLogs, err := printerLogRepo.GetByUuids(req.Uuids())
	if err != nil {
		return errors.WithMessage(err, "获取打印日志失败")
	}
	// 创建UUID到现有日志的映射
	existingLogMap := make(map[uint64]*model.PrinterLog)
	for i := range existingLogs {
		existingLogMap[existingLogs[i].Uuid] = &existingLogs[i]
	}
	// 更新打印日志
	printerLogs := []model.PrinterLog{}
	for _, report := range req.Data {
		// 获取当前Num值并自增1
		currentNum := 1
		if existingLog, exists := existingLogMap[report.Uuid]; exists {
			currentNum = existingLog.Num + 1
		}
		printerLogs = append(printerLogs, model.PrinterLog{
			BaseModel: model.BaseModel{
				Uuid: report.Uuid,
			},
			Num:    currentNum,
			Status: utils.IfInt(report.Status == 0, constant.PrinterLogStatusEnd, constant.PrinterLogStatusSuccess),
			Reason: report.Reason,
		})
	}
	//
	err = printerLogRepo.BatchUpdate(printerLogs)
	if err != nil {
		return errors.WithMessage(err, "更新打印日志失败")
	}
	//
	return nil
}

// GetOldOrderPrinterConfig 获取旧订单打印配置
func (s *printerLogSrv) GetOldOrderPrinterConfig(ctx context.Context, data string) (*resp.PrinterData, error) {
	// 获取打印设置
	printerSetting, err := s.settingSrv.GetPrinterSetting(ctx, nil)
	if err != nil {
		logger.Logger.Error("获取打印机设置失败", zap.Error(err))
		fmt.Println("获取打印机设置失败", zap.Error(err))
		return nil, errors.WithMessage(err, "获取打印机设置失败")
	}
	// 获取打印设置
	settingPrinterInfo, err := s.settingSrv.GetPrinterInfo(ctx, printerSetting, ctx.GetDeviceSn())
	if err != nil {
		return nil, errors.WithMessage(err, "获取打印设置失败")
	}
	//
	return &resp.PrinterData{
		Uuid:              0,
		Data:              utils.GzipCompressData(data),
		PrintMethod:       2,
		Copies:            settingPrinterInfo.Copies,
		PrinterType:       settingPrinterInfo.PrinterType,
		PrinterConfig:     settingPrinterInfo.PrinterConfig,
		IsCashierPrinter:  settingPrinterInfo.IsCashierPrinter,
		IsUsbPrinter:      settingPrinterInfo.IsUsbPrinter,
		PrintingTime:      200,
		EnableStatusCheck: settingPrinterInfo.EnableStatusCheck,
		TradeNo:           fmt.Sprintf("%d%s", time.Now().Unix(), hex.EncodeToString([]byte(strings.Replace(uuid.New().String(), "-", "", -1)))[:8]),
		PrintChunkSize:    10 * 1024 * 1024, // 10MB
	}, nil
}
