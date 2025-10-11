package service

import (
	"fmt"
	"strconv"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	respSetting "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"
)

type IPrinterSrv interface {
	GetProductPrinterList(ctx context.Context) (resp.ProductPrinterList, error)                              // 获取打印档口列表
	UsbPrinterReport(ctx context.Context, reportReq req.UsbPrinterReportReq) (resp.PrinterReportResp, error) // usb打印机上报
}

type printerSrv struct {
	dbm   *database.DBManager
	cache cache.Cache
}

func NewPrinterSrv(dbm *database.DBManager, cache cache.Cache) IPrinterSrv {
	return NewPrinterSrvImpl(dbm, cache)
}

func NewPrinterSrvImpl(dbm *database.DBManager, cache cache.Cache) IPrinterSrv {
	return &printerSrv{
		dbm:   dbm,
		cache: cache,
	}
}

// GetProductPrinterList 获取打印档口列表
func (s *printerSrv) GetProductPrinterList(ctx context.Context) (resp.ProductPrinterList, error) {
	productPrinterRepo := repository.NewProductPrinterRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	// 如果不是商家端，则只查询开启的打印档口
	opts := []repository.DBOption{}
	if ctx.GetSource() != constant.SourceShop {
		opts = append(opts, productPrinterRepo.WhereStatus(constant.ProductPrinterStatusOpen))
	}
	// 查询打印档口列表
	printers, err := productPrinterRepo.GetProductPrinters()
	if err != nil {
		return resp.ProductPrinterList{List: make([]resp.ProductPrinter, 0)}, errors.ErrInternal
	}
	productPrinters := make([]resp.ProductPrinter, 0, len(printers))
	for _, printer := range printers {
		productPrinters = append(productPrinters, resp.ProductPrinter{
			Uuid:   printer.Uuid,
			Name:   printer.Name,
			Status: printer.Status,
		})
	}
	// 返回打印档口列表
	return resp.ProductPrinterList{List: productPrinters}, nil
}

// UsbPrinterReport 上报打印
func (s *printerSrv) UsbPrinterReport(ctx context.Context, reportReq req.UsbPrinterReportReq) (resp.PrinterReportResp, error) {
	printerRepo := repository.NewPrinterRepo(ctx.GetDB())
	dbUsbList := printerRepo.GetUsbList()
	selectedUuid := uint64(0)
	lastNewUsb := model.Printer{}

	// 优化：创建已有打印机映射，避免多次循环查询
	dbUsbMap := make(map[string]model.Printer)
	for _, usb := range dbUsbList {
		dbUsbMap[usb.ConfigJson] = usb
	}

	// 获取打印设置
	printerSetting, err := setting.NewSrv(s.dbm, s.cache).GetPrinterSetting(ctx, nil)
	if err != nil {
		return resp.PrinterReportResp{}, errors.ErrInternal
	}

	// 更新打印机状态
	{
		// 更新为离线
		if len(dbUsbList) > 0 {
			// 创建打印机配置映射
			printerConfigMap := make(map[string]bool)
			for _, printer := range reportReq.List {
				printerConfigMap[utils.JsonToStr(printer)] = true
			}
			// 检查数据库中的打印机是否在新上报列表中
			for _, usb := range dbUsbList {
				// 如果不在新列表中且状态为在线，则更新为离线
				if !printerConfigMap[usb.ConfigJson] && usb.Status == 1 {
					if err := printerRepo.UpdateBySourceDeviceSn(usb.ID, ctx.GetDeviceSn(), map[string]interface{}{
						"status": 0,
					}); err != nil {
						fmt.Printf("Error updating usb print: %v\n", err)
					}
					// 打印日志
					logger.Logger.Info(fmt.Sprintf("更新打印机状态为离线: CompanyUuid=%d, DeviceSN=%s, PrinterUuid=%d, Name=%s, 最后心跳时间: %d秒", ctx.GetCompanyUuid(), ctx.GetDeviceSn(), usb.Uuid, usb.Name, uint(time.Now().Unix())))
				}
			}
		}

		// 有新的打印机数据，遍历处理
		if len(reportReq.List) > 0 {
			// 处理每个上报的USB打印机
			for _, usbPrinter := range reportReq.List {
				printerJson := utils.JsonToStr(usbPrinter)
				if dbUsb, exists := dbUsbMap[printerJson]; exists {
					// 已存在的打印机，更新状态
					if err := printerRepo.Update(dbUsb.ID, map[string]interface{}{
						"status":              1,
						"last_heartbeat_time": uint(time.Now().Unix()),
						"source_device_sn":    ctx.GetDeviceSn(),
					}); err != nil {
						fmt.Printf("Error updating usb print: %v\n", err)
					}
					// 打印日志
					if dbUsb.Status == 0 {
						logger.Logger.Info(fmt.Sprintf("更新打印机状态为在线: CompanyUuid=%d, DeviceSN=%s, PrinterUuid=%d, Name=%s, 最后心跳时间: %d秒", ctx.GetCompanyUuid(), ctx.GetDeviceSn(), dbUsb.Uuid, dbUsb.Name, uint(time.Now().Unix())))
					}
					// 如果只有一个打印机，则选中这个打印机
					if len(reportReq.List) == 1 {
						lastNewUsb = dbUsb
					} else {
						for _, sprinter := range printerSetting.CashierPrinter {
							if sprinter.Key == ctx.GetDeviceSn() {
								if sprinter.PrinterUsbId == "0" || sprinter.PrinterUsbId == "" {
									sprinter.PrinterUsbId = strconv.FormatUint(dbUsb.Uuid, 10)
									lastNewUsb = dbUsb
								}
								break
							}
						}
					}
				} else {
					uuid, _ := utils.GetID()
					// 区分类型
					printerTypeKey := constant.PRINTER_TYPE_XPRINTER_LAN
					if usbPrinter.Vid.(float64) == 1137 && usbPrinter.Pid.(float64) == 85 {
						if usbPrinter.M_name == "Zhuhai Howbest Label Printer Co.,Ltd." {
							printerTypeKey = constant.PRINTER_TYPE_GP_C200IV
						} else if usbPrinter.M_name == "ZHU HAI HOWBEST Receipt Printer Co.,Ltd." {
							printerTypeKey = constant.PRINTER_TYPE_GP_D300I
						} else {
							printerTypeKey = constant.PRINTER_TYPE_GP_D300I
						}
					}
					// 获取打印机类型(只查询一次)
					printerType := repository.NewPrinterTypeRepository(ctx.GetDB()).GetRecordByKey(ctx.GetCompanyUuid(), printerTypeKey)
					if printerType.ID == 0 {
						fmt.Printf("Error: printer type XPRINTER_LAN not found\n")
						return resp.PrinterReportResp{}, errors.ErrInternal
					}
					// 更新映射
					dbUsbMap[printerJson] = model.Printer{
						BaseModel: model.BaseModel{
							Uuid:       uuid,
							CreateTime: time.Now().Unix(),
							UpdateTime: time.Now().Unix(),
						},
						Name:              usbPrinter.Name,
						PrinterTypeUuid:   printerType.Uuid,
						ConfigJson:        printerJson,
						Copies:            1,
						Sort:              0,
						IsUsb:             1,
						SourceDeviceSn:    ctx.GetDeviceSn(),
						Status:            1,
						LastHeartbeatTime: uint(time.Now().Unix()),
					}
					// 新打印机，创建记录
					if err := printerRepo.Create(ctx.GetCompanyUuid(), dbUsbMap[printerJson]); err != nil {
						fmt.Printf("Error creating usb print: %v\n", err)
					}
					lastNewUsb = dbUsbMap[printerJson]
					// 打印日志
					logger.Logger.Info(fmt.Sprintf("新增打印机: CompanyUuid=%d, DeviceSN=%s, PrinterUuid=%d, Name=%s, 最后心跳时间: %d秒", ctx.GetCompanyUuid(), ctx.GetDeviceSn(), uuid, usbPrinter.Name, uint(time.Now().Unix())))
				}
			}
		}
	}

	// 更新打印设置
	{

		// 更新打印设置
		isUpdate := false
		if selectedUuid != 0 {
			isExist := false
			for i, sprinter := range printerSetting.CashierPrinter {
				if sprinter.Key == ctx.GetDeviceSn() {
					printerSetting.CashierPrinter[i].PrinterUsbId = strconv.FormatUint(selectedUuid, 10)
					isExist = true
					break
				}
			}
			if !isExist {
				printerSetting.CashierPrinter = append(printerSetting.CashierPrinter, respSetting.CashierPrinterItem{
					Key:          ctx.GetDeviceSn(),
					PrinterId:    "0",
					PrinterUsbId: strconv.FormatUint(selectedUuid, 10),
				})
			}
			isUpdate = true
		} else if len(reportReq.List) > 0 && lastNewUsb.Uuid != 0 {
			isExist := false
			for i, sprinter := range printerSetting.CashierPrinter {
				if sprinter.Key == ctx.GetDeviceSn() {
					printerSetting.CashierPrinter[i].PrinterUsbId = strconv.FormatUint(lastNewUsb.Uuid, 10)
					isExist = true
					break
				}
			}
			if !isExist {
				printerSetting.CashierPrinter = append(printerSetting.CashierPrinter, respSetting.CashierPrinterItem{
					Key:          ctx.GetDeviceSn(),
					PrinterId:    "0",
					PrinterUsbId: strconv.FormatUint(lastNewUsb.Uuid, 10),
				})
			}
			isUpdate = true
		}
		if isUpdate {
			repository.NewSettingRepo(ctx.GetDB()).Updates(constant.SettingPrinter, utils.ToJson(printerSetting))
			s.cache.Del(fmt.Sprintf("setting:company_id:%d", ctx.GetCompanyUuid()))
		}

		// 返回数据
		return resp.PrinterReportResp{}, nil
	}
}
