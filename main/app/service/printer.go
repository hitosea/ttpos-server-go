package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	respSetting "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/pkg"
	template_info "ttpos-server-go/app/printer/pkg/template"
	"ttpos-server-go/app/printer/template"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IPrinterSrv interface {
	GetProductPrinterList(ctx context.Context) (resp.ProductPrinterList, error)                                                                       // 获取打印档口列表
	UsbPrinterReport(ctx context.Context, reportReq req.UsbPrinterReportReq) (resp.PrinterReportResp, error)                                          // usb打印机上报
	GetPrintTemplateList(ctx context.Context) (resp.PrintTemplateListResp, error)                                                                     // 获取打印模板列表
	GetPrintTemplateDetail(ctx context.Context, id uint64) (resp.PrintTemplateDetailResp, error)                                                      // 获取模板详情
	EditPrinterCustomize(ctx context.Context, editPrinterCustomizeReq req.EditPrinterCustomizeReq) error                                              // 编辑打印机定制
	PreviewPrinterCustomize(ctx context.Context, previewPrinterCustomizeReq req.PreviewPrinterCustomizeReq) (resp.PreviewPrinterCustomizeResp, error) // 预览打印机定制
	DeletePrinterCustomize(ctx context.Context, customizeUuid uint64) error                                                                           // 删除打印机定制
	CreatePrinterCustomize(ctx context.Context, createPrinterCustomizeReq req.CreatePrinterCustomizeReq) (resp.CreatePrinterCustomizeResp, error)     // 创建打印机定制
	UsePrinterCustomize(ctx context.Context, customizeUuid uint64) error                                                                              // 使用打印机定制
	GetPrinterCustomizeConfigInfo(ctx context.Context, configInfoReq req.PrinterGetConfigInfoReq) (resp.ConfigInfoResp, error)                        // 获取配置信息
}

const (
	DefaultTemplateName = "门店-默认模版"
	TemplatePngPath     = "app/printer/pkg/text/tmp/printer/complex_template_test.png"
)

const (
	TemplateNameBilling    = "结账单"
	TemplateNamePreBilling = "预结账单"
	TemplateNameInvoice    = "发票"
)

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
	// 添加公司锁 - 禁止并发操作
	companyUuid := ctx.GetCompanyUuid()
	lock.NewSystemLock().LockUuid(companyUuid)
	defer lock.NewSystemLock().UnlockUuid(companyUuid)

	// 获取打印机列表
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

// GetPrintTemplateList 获取打印模板列表
func (s *printerSrv) GetPrintTemplateList(ctx context.Context) (resp.PrintTemplateListResp, error) {
	printerTemplateRepo := repository.NewPrinterTemplateRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	templates, err := printerTemplateRepo.GetPrinterTemplates()
	if err != nil {
		return resp.PrintTemplateListResp{List: make([]resp.PrintTemplateGroup, 0)}, errors.ErrInternal
	}

	// 转换为分组列表
	var groups []resp.PrintTemplateGroup
	groups = append(groups, resp.PrintTemplateGroup{
		LocaleName: dto.LocaleResponse{
			ZH:   "收银小票",
			EN:   "Cashier Receipt",
			TH:   "รายการชำระเงิน",
			JA:   "現金レシート",
			KO:   "현금 영수증",
			MY:   "Keluaran Tunai",
			TR:   "Nakit Alış Fişi",
			SV:   "Kassafaktura",
			ZHTW: "收銀小票",
		},
		GroupType: 1,
		List:      make([]resp.PrintTemplate, 0),
	})
	// groups = append(groups, resp.PrintTemplateGroup{
	// 	LocaleName: dto.LocaleResponse{
	// 		ZH:   "厨房小票",
	// 		EN:   "Kitchen Menu",
	// 		TH:   "เมนูอาหาร",
	// 		JA:   "キッチンメニュー",
	// 		KO:   "주방 메뉴",
	// 		MY:   "Resipi Makanan",
	// 		TR:   "Yemek Menüsü",
	// 		SV:   "Maträtmeny",
	// 		ZHTW: "廚房菜單",
	// 	},
	// 	GroupType: 2,
	// 	List:      make([]resp.PrintTemplate, 0),
	// })

	// 模版ID列表
	templateOrders := []uint64{
		constant.PrinterTemplatePreBilling, // 预结账单
		constant.PrinterTemplateBilling,    // 结账单
		constant.PrinterTemplateInvoice,    // 发票
		// constant.PrinterTemplateRecharge,      // 充值单
		// constant.PrinterTemplateBusiness,      // 营业数据
		// constant.PrinterTemplateHandoverSheet, // 交班单
		// constant.PrinterTemplateTakeoutOrder,  // 外送单
	}
	templateKitchen := []uint64{
		constant.PrinterTemplateOneDishOneMenu, // 一菜一单
		constant.PrinterTemplateEntireOrder,    // 整单打印
		constant.PrinterTemplateReturnDish,     // 退菜单
		constant.PrinterTemplateOutMenu,        // 出菜单
	}

	// 创建模板ID到模板的映射
	templateMap := make(map[uint64]model.PrinterTemplate)
	for _, template := range templates {
		templateMap[template.ID] = template
	}

	// 按照预定义顺序添加打印菜单
	for i := range groups {
		var orderedTemplateIds []uint64
		if groups[i].GroupType == 1 {
			orderedTemplateIds = templateOrders
		} else if groups[i].GroupType == 2 {
			orderedTemplateIds = templateKitchen
		}

		for _, templateId := range orderedTemplateIds {
			if template, exists := templateMap[templateId]; exists {
				groups[i].List = append(groups[i].List, resp.PrintTemplate{
					ID: template.ID,
					LocaleName: dto.LocaleResponse{
						ZH:   i18n.Translate(i18n.LanguageZH, template.Name),
						EN:   i18n.Translate(i18n.LanguageEN, template.Name),
						TH:   i18n.Translate(i18n.LanguageTH, template.Name),
						JA:   i18n.Translate(i18n.LanguageJA, template.Name),
						KO:   i18n.Translate(i18n.LanguageKO, template.Name),
						MY:   i18n.Translate(i18n.LanguageMY, template.Name),
						TR:   i18n.Translate(i18n.LanguageTR, template.Name),
						SV:   i18n.Translate(i18n.LanguageSV, template.Name),
						ZHTW: i18n.Translate(i18n.LanguageZHTW, template.Name),
					},
				})
			}
		}
	}

	//
	return resp.PrintTemplateListResp{List: groups}, nil
}

// Parser 解析器
func (s *printerSrv) Parser(ctx context.Context, templateJSONStr string, testData map[string]interface{}) (string, error) {
	currencySetting, err := setting.NewSrv(s.dbm, s.cache).GetCurrencySetting(ctx)
	if err != nil {
		return "", errors.WithMessage(errors.New("获取打印设置失败"), err.Error())
	}

	// 创建解析器
	unitPosition, err := strconv.ParseInt(currencySetting.UnitPosition, 10, 64)
	if err != nil {
		return "", errors.WithMessage(errors.New("转换货币单位位置失败"), err.Error())
	}
	parser, err := pkg.NewImgTemplateParser(pkg.ImgBaseData{
		Language:             ctx.GetLanguage(),
		CurrencyUnit:         currencySetting.PrintUnit,
		CurrencyUnitPosition: int(unitPosition),
	}, templateJSONStr, testData)
	if err != nil {
		return "", errors.WithMessage(errors.New("创建模板解析器失败"), err.Error())
	}

	// 验证模板
	err = parser.ValidateTemplate()
	if err != nil {
		return "", errors.WithMessage(errors.New("验证模板失败"), err.Error())
	}

	// 解析模板
	img, err := parser.Parse()
	if err != nil {
		return "", errors.WithMessage(errors.New("解析模板失败"), err.Error())
	}

	// 设置分割高度为200000
	img.SegmentationHeight = 200000
	img.Save(TemplatePngPath, false, 0)

	// 读取保存的图片文件并转换为base64
	imageData, err := os.ReadFile(TemplatePngPath)
	if err != nil {
		return "", errors.WithMessage(errors.New("读取生成的图片文件失败"), err.Error())
	}

	// 转换为base64字符串
	base64Str := base64.StdEncoding.EncodeToString(imageData)

	// 添加data URL前缀
	dataURL := "data:image/png;base64," + base64Str

	return dataURL, nil
}

// ParserConcurrent 并发安全的模板解析器（支持多个goroutine同时调用）
// uniqueID 用于生成唯一的临时文件名，避免并发冲突
func (s *printerSrv) ParserConcurrent(ctx context.Context, templateJSONStr string, testData map[string]interface{}, uniqueID uint64) (string, error) {
	currencySetting, err := setting.NewSrv(s.dbm, s.cache).GetCurrencySetting(ctx)
	if err != nil {
		return "", errors.WithMessage(errors.New("获取打印设置失败"), err.Error())
	}

	// 创建解析器
	unitPosition, err := strconv.ParseInt(currencySetting.UnitPosition, 10, 64)
	if err != nil {
		return "", errors.WithMessage(errors.New("转换货币单位位置失败"), err.Error())
	}
	parser, err := pkg.NewImgTemplateParser(pkg.ImgBaseData{
		Language:             ctx.GetLanguage(),
		CurrencyUnit:         currencySetting.PrintUnit,
		CurrencyUnitPosition: int(unitPosition),
	}, templateJSONStr, testData)
	if err != nil {
		return "", errors.WithMessage(errors.New("创建模板解析器失败"), err.Error())
	}

	// 验证模板
	err = parser.ValidateTemplate()
	if err != nil {
		return "", errors.WithMessage(errors.New("验证模板失败"), err.Error())
	}

	// 解析模板
	img, err := parser.Parse()
	if err != nil {
		return "", errors.WithMessage(errors.New("解析模板失败"), err.Error())
	}

	// 使用唯一的临时文件名，避免并发冲突
	tempFilePath := fmt.Sprintf("/tmp/printer_template_%d_%d.png", uniqueID, time.Now().UnixNano())
	defer os.Remove(tempFilePath) // 确保清理临时文件

	// 设置分割高度并保存到临时文件
	img.SegmentationHeight = 200000
	img.Save(tempFilePath, false, 0)

	// 读取保存的图片文件并转换为base64
	imageData, err := os.ReadFile(tempFilePath)
	if err != nil {
		return "", errors.WithMessage(errors.New("读取生成的图片文件失败"), err.Error())
	}

	// 转换为base64字符串
	base64Str := base64.StdEncoding.EncodeToString(imageData)

	// 添加data URL前缀
	dataURL := "data:image/png;base64," + base64Str

	return dataURL, nil
}

// GetTestData 获取测试数据
func (s *printerSrv) GetTestData(ctx context.Context, templateName string) (map[string]interface{}, error) {
	// 从JSON文件读取测试数据
	testDataBytes, err := template_info.GetTemplateDataJsonData(templateName)
	if err != nil {
		return nil, errors.WithMessage(errors.New("读取测试数据文件失败"), err.Error())
	}
	var testData map[string]interface{}
	if err := json.Unmarshal(testDataBytes, &testData); err != nil {
		return nil, errors.WithMessage(errors.New("解析测试数据JSON失败"), err.Error())
	}

	// 初始化 settingSrv，供多处使用
	settingSrv := setting.NewSrvImpl(s.dbm, s.cache)

	if testData["store"] != nil {
		storeSetting, err := settingSrv.GetStoreSetting(ctx)
		if err != nil {
			return nil, errors.WithMessage(errors.New("获取门店设置失败"), err.Error())
		}
		printerSetting, err := settingSrv.GetPrinterSetting(ctx, nil)
		if err != nil {
			return nil, errors.WithMessage(errors.New("获取打印机设置失败"), err.Error())
		}
		currencySetting, err := settingSrv.GetCurrencySetting(ctx)
		if err != nil {
			return nil, errors.WithMessage(errors.New("获取货币设置失败"), err.Error())
		}
		if testData["store"].(map[string]interface{})["logo"] != nil {
			logoAddr := template.NewPrinterTemplate(ctx, settingSrv, &storeSetting, &printerSetting, &currencySetting, false, ctx.GetLanguage()).GetLogoAddr()
			if logoAddr != "" {
				testData["store"].(map[string]interface{})["logo"] = logoAddr
			}
		}
		if testData["store"].(map[string]interface{})["name"] != nil {
			testData["store"].(map[string]interface{})["name"] = i18n.Translate(ctx.GetLanguage(), "店铺名称")
		}
		if testData["store"].(map[string]interface{})["address"] != nil {
			testData["store"].(map[string]interface{})["address"] = i18n.Translate(ctx.GetLanguage(), "商家地址商家地址商家地址商家地址商家地址")
		}
		if testData["store"].(map[string]interface{})["company"] != nil {
			testData["store"].(map[string]interface{})["company"] = i18n.Translate(ctx.GetLanguage(), "公司名称公司名称公司名称公司名称公司名称公司名称公司名称")
		}
		if testData["store"].(map[string]interface{})["company_addr"] != nil {
			testData["store"].(map[string]interface{})["company_addr"] = i18n.Translate(ctx.GetLanguage(), "公司地址公司地址公司地址公司地址")
		}
	}
	if testData["invoice"] != nil {
		if testData["invoice"].(map[string]interface{})["company_name"] != nil {
			testData["invoice"].(map[string]interface{})["company_name"] = i18n.Translate(ctx.GetLanguage(), "公司名称公司名称公司名称公司名称公司名称公司名称公司名称")
		}
		if testData["invoice"].(map[string]interface{})["company_address"] != nil {
			testData["invoice"].(map[string]interface{})["company_address"] = i18n.Translate(ctx.GetLanguage(), "商家地址商家地址商家地址商家地址商家地址")
		}
	}
	if testData["order"] != nil {
		// 备注
		if testData["order"].(map[string]interface{})["remark"] != nil {
			testData["order"].(map[string]interface{})["remark"] = i18n.Translate(ctx.GetLanguage(), "开桌备注")
		}
		// 商品
		if testData["order"].(map[string]interface{})["products"] != nil {
			switch products := testData["order"].(map[string]interface{})["products"].(type) {
			case []interface{}:
				for _, p := range products {
					product, ok := p.(map[string]interface{})
					if !ok {
						continue
					}
					if v, ok := product["attrs"].(string); ok {
						product["attrs"] = i18n.Translate(ctx.GetLanguage(), v)
					}
					if v, ok := product["attr"].(string); ok {
						product["attr"] = i18n.Translate(ctx.GetLanguage(), v)
					}
					if v, ok := product["sauce_names"].(string); ok {
						product["sauce_names"] = i18n.Translate(ctx.GetLanguage(), v)
					}
					if v, ok := product["flavor_name"].(string); ok {
						product["flavor_name"] = i18n.Translate(ctx.GetLanguage(), v)
					}
					if v, ok := product["name"].(string); ok {
						product["name"] = i18n.Translate(ctx.GetLanguage(), v)
					}
				}
			}
		}
		// 支付方式
		if testData["order"].(map[string]interface{})["payment_methods"] != nil {
			switch methods := testData["order"].(map[string]interface{})["payment_methods"].(type) {
			case []interface{}:
				for _, m := range methods {
					paymentMethod, ok := m.(map[string]interface{})
					if !ok {
						continue
					}
					if v, ok := paymentMethod["name"].(string); ok {
						paymentMethod["name"] = i18n.Translate(ctx.GetLanguage(), v)
					}
				}
			}
		}
		// 动态设置 order.is_contain_tax 根据商家配置（与实际打印单逻辑保持一致）
		if orderData, ok := testData["order"].(map[string]interface{}); ok {
			// 获取税率设置和打印机设置
			taxRateSetting, err := settingSrv.GetTaxRateSetting(ctx)
			if err == nil {
				printerSetting, err := settingSrv.GetPrinterSetting(ctx, nil)
				if err == nil {
					taxFeeType := taxRateSetting.GetTaxFeeType()
					// 打印机消费税设置
					consumptionTax, _ := strconv.Atoi(printerSetting.ConsumptionTax)
					// 与实际打印单相同的逻辑
					// 商品未含税：需要TaxFeeType == 1 && (ConsumptionTax == 1 || 3)
					if taxFeeType == 1 && (consumptionTax == 1 || consumptionTax == 3) {
						orderData["is_contain_tax"] = uint(1)
					} else if taxFeeType == 2 && (consumptionTax == 1 || consumptionTax == 2) {
						// 商品已含税：需要 TaxFee > 0 && TaxFeeType == 2 && (ConsumptionTax == 1 || 2)
						orderData["is_contain_tax"] = uint(2)
					} else {
						// 关闭消费税
						orderData["is_contain_tax"] = uint(0)
					}
				}
			}
		}
	}
	return testData, nil
}

// GetTemplateJSONStr 获取模板JSON字符串
func (s *printerSrv) GetTemplateJSONStr(ctx context.Context, templateName string) (string, error) {
	templateJSON, err := template_info.GetTemplateJsonData(templateName)
	if err != nil {
		return "", errors.WithMessage(errors.New("读取模板文件失败"), err.Error())
	}
	return string(templateJSON), nil
}

// GetTemplateConfigInfo 获取模板配置信息
func (s *printerSrv) GetTemplateConfigInfo(ctx context.Context, templateName string, isAdv bool) (string, error) {
	templateJSON, err := template_info.GetTemplateConfigJsonData(templateName)
	if err != nil {
		return "", errors.WithMessage(errors.New("读取模板文件失败"), err.Error())
	}
	return string(templateJSON), nil
}

// GetPrintTemplateDetail 获取菜单详情（并发优化版本）
func (s *printerSrv) GetPrintTemplateDetail(ctx context.Context, id uint64) (resp.PrintTemplateDetailResp, error) {
	db := ctx.GetDB()
	companySetting := ctx.GetCompanySetting()
	commonRepo := repository.NewCommonRepo()
	printerCustomizeRepo := repository.NewPrinterCustomizeRepo(db)

	// 获取打印模板详情
	template, err := repository.NewPrinterTemplateRepo(db).GetPrinterTemplateInfo(id)
	if err != nil {
		return resp.PrintTemplateDetailResp{}, errors.WithMessage(errors.New("获取打印模板详情失败"), err.Error())
	}

	// 创建复杂的测试模板
	templateJSONStr, err := s.GetTemplateJSONStr(ctx, template.Name)

	// 获取打印机定制列表
	customizes, err := printerCustomizeRepo.GetPrinterCustomizeList(
		commonRepo.WhereBySoftDelete(),
		printerCustomizeRepo.WhereByTemplateId(template.ID),
	)
	if err != nil {
		return resp.PrintTemplateDetailResp{}, errors.WithMessage(errors.New("获取打印机定制列表失败"), err.Error())
	}

	// 获取测试数据
	testData, err := s.GetTestData(ctx, template.Name)
	if err != nil {
		return resp.PrintTemplateDetailResp{}, errors.WithMessage(errors.New("获取测试数据失败"), err.Error())
	}

	// 定义结果结构
	type templateResult struct {
		customize model.PrinterCustomize
		imgUrl    string
		err       error
		isDefault bool
	}

	// 创建结果通道和等待组
	resultChan := make(chan templateResult, len(customizes)+1)
	var wg sync.WaitGroup

	// 并发生成所有模板预览图
	for _, customize := range customizes {
		wg.Add(1)
		go func(cust model.PrinterCustomize) {
			defer wg.Done()

			// 使用并发安全的Parser
			imgUrl, err := s.ParserConcurrent(ctx, cust.Data, testData, cust.Uuid)

			resultChan <- templateResult{
				customize: cust,
				imgUrl:    imgUrl,
				err:       err,
				isDefault: cust.IsAdv == 0 && cust.TemplateId == template.ID,
			}
		}(customize)
	}

	// 检查是否需要生成默认模板
	hasDefault := false
	for _, customize := range customizes {
		if customize.IsAdv == 0 && customize.TemplateId == template.ID {
			hasDefault = true
			break
		}
	}

	// 如果没有默认模板，并发生成一个
	if !hasDefault {
		wg.Add(1)
		go func() {
			defer wg.Done()
			imgUrl, err := s.ParserConcurrent(ctx, templateJSONStr, testData, 0)
			resultChan <- templateResult{
				imgUrl:    imgUrl,
				err:       err,
				isDefault: true,
			}
		}()
	}

	// 等待所有任务完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 初始化默认模板
	defaultTemplate := resp.PrintTemplateDetail{
		ID:    template.ID,
		Name:  i18n.Translate(ctx.GetLanguage(), DefaultTemplateName),
		IsUse: false,
	}

	// 收集结果
	advReceiptTpls := make([]resp.PrintTemplateDetail, 0)
	needCreateDefault := false

	for result := range resultChan {
		if result.err != nil {
			// 记录错误但不中断（容错处理）
			zap.L().Error("生成模板预览图失败",
				zap.Uint("customize_id", result.customize.ID),
				zap.Error(result.err))
			continue
		}

		if result.isDefault {
			// 处理默认模板
			defaultTemplate.ImgUrl = result.imgUrl
			if result.customize.ID > 0 {
				defaultTemplate.Name = utils.IfString(result.customize.Name == DefaultTemplateName,
					i18n.Translate(ctx.GetLanguage(), DefaultTemplateName),
					result.customize.Name)
				defaultTemplate.IsUse = result.customize.IsUse == 1
				defaultTemplate.CustomizeUuid = result.customize.Uuid
			} else {
				// 没有customize记录，需要创建
				needCreateDefault = true
			}
		} else {
			// 处理高级模板
			advReceiptTpls = append(advReceiptTpls, resp.PrintTemplateDetail{
				ID:            template.ID,
				Name:          result.customize.Name,
				IsUse:         result.customize.IsUse == 1,
				CustomizeUuid: result.customize.Uuid,
				ImgUrl:        result.imgUrl,
			})
		}
	}

	// 如果需要创建默认模板记录
	if needCreateDefault && defaultTemplate.ImgUrl != "" {
		uuid, err := utils.GetID()
		if err != nil {
			return resp.PrintTemplateDetailResp{}, errors.WithMessage(errors.New("生成雪花ID失败"), err.Error())
		}
		err = printerCustomizeRepo.CreatePrinterCustomize(model.PrinterCustomize{
			BaseModel:  model.BaseModel{Uuid: uuid},
			Name:       DefaultTemplateName,
			Data:       templateJSONStr,
			TemplateId: template.ID,
			IsAdv:      0,
		})
		if err != nil {
			return resp.PrintTemplateDetailResp{}, errors.WithMessage(errors.New("创建打印机定制失败"), err.Error())
		}
		defaultTemplate.CustomizeUuid = uuid
	}

	// 返回结果
	return resp.PrintTemplateDetailResp{
		DefaultTpl:      defaultTemplate,
		AdvReceiptTpls:  advReceiptTpls,
		IsAdvReceiptTpl: companySetting.IsOpenAdvancedTicketPrint == 1,
	}, nil
}

// EditPrinterCustomize 编辑打印机定制
func (s *printerSrv) EditPrinterCustomize(ctx context.Context, editPrinterCustomizeReq req.EditPrinterCustomizeReq) error {
	db := ctx.GetDB()
	printerCustomizeRepo := repository.NewPrinterCustomizeRepo(db)
	// 检查打印机定制是否存在
	customizeInfo, err := printerCustomizeRepo.GetPrinterCustomizeInfo(editPrinterCustomizeReq.CustomizeUuid)
	if err != nil {
		return errors.WithMessage(errors.New("检查打印机定制是否存在失败"), err.Error())
	}
	// 获取打印模板详情
	template, err := repository.NewPrinterTemplateRepo(db).GetPrinterTemplateInfo(customizeInfo.TemplateId)
	if err != nil {
		return errors.WithMessage(errors.New("获取打印模板详情失败"), err.Error())
	}
	//
	if customizeInfo.IsAdv == 1 {
		if ctx.GetCompanySetting().IsOpenAdvancedTicketPrint == 0 {
			return errors.New("未开启高级模版打印")
		}
	}
	//
	testData, err := s.GetTestData(ctx, template.Name)
	if err != nil {
		return errors.WithMessage(errors.New("获取测试数据失败"), err.Error())
	}
	// 解析模板
	_, err = s.Parser(ctx, editPrinterCustomizeReq.Data, testData)
	if err != nil {
		return errors.WithMessage(errors.New("解析模板失败"), err.Error())
	}
	// 更新打印机定制
	customizeInfo.Name = func() string {
		if customizeInfo.IsAdv == 1 {
			return editPrinterCustomizeReq.Name
		}
		return customizeInfo.Name
	}()
	customizeInfo.Data = editPrinterCustomizeReq.Data

	// 更新打印机定制
	if err := db.Transaction(func(tx *gorm.DB) error {
		err = repository.NewPrinterCustomizeRepo(tx).UpdatePrinterCustomize(customizeInfo)
		if err != nil {
			return errors.WithMessage(errors.New("更新打印机定制失败"), err.Error())
		}
		// 更新打印机模板
		if customizeInfo.IsUse == 1 {
			err = repository.NewPrinterTemplateRepo(tx).UpdatePrinterTemplate(model.PrinterTemplate{
				ID:      customizeInfo.TemplateId,
				TmpUuid: customizeInfo.Uuid,
				TmpData: customizeInfo.Data,
			})
			if err != nil {
				return errors.WithMessage(errors.New("更新打印机模板失败"), err.Error())
			}
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	return nil
}

// PreviewPrinterCustomize 预览打印机定制
func (s *printerSrv) PreviewPrinterCustomize(ctx context.Context, previewPrinterCustomizeReq req.PreviewPrinterCustomizeReq) (resp.PreviewPrinterCustomizeResp, error) {
	db := ctx.GetDB()
	// 获取模板详情
	template, err := repository.NewPrinterTemplateRepo(db).GetPrinterTemplateInfo(previewPrinterCustomizeReq.TemplateId)
	if err != nil {
		return resp.PreviewPrinterCustomizeResp{}, errors.WithMessage(errors.New("获取模板详情失败"), err.Error())
	}
	// 获取测试数据
	testData, err := s.GetTestData(ctx, template.Name)
	if err != nil {
		return resp.PreviewPrinterCustomizeResp{}, errors.WithMessage(errors.New("获取测试数据失败"), err.Error())
	}
	// 解析模板
	printContent, err := s.Parser(ctx, previewPrinterCustomizeReq.Data, testData)
	if err != nil {
		return resp.PreviewPrinterCustomizeResp{}, errors.WithMessage(errors.New("解析模板失败"), err.Error())
	}
	return resp.PreviewPrinterCustomizeResp{ImageUrl: printContent}, nil
}

// DeletePrinterCustomize 删除打印机定制
func (s *printerSrv) DeletePrinterCustomize(ctx context.Context, customizeUuid uint64) error {
	db := ctx.GetDB()
	// 检查打印机定制是否存在
	customizeInfo, err := repository.NewPrinterCustomizeRepo(db).GetPrinterCustomizeInfo(customizeUuid)
	if err != nil {
		return errors.WithMessage(errors.New("检查打印机定制是否存在失败"), err.Error())
	}
	if customizeInfo.IsAdv == 0 {
		return errors.New("默认模版不能删除")
	}
	// 检查打印机定制是否正在使用中
	if customizeInfo.IsUse == 1 {
		return errors.New("打印机定制正在使用中，不能删除")
	}
	// 删除打印机定制
	err = repository.NewPrinterCustomizeRepo(db).DeletePrinterCustomize(customizeInfo.Uuid)
	if err != nil {
		return errors.WithMessage(errors.New("删除打印机定制失败"), err.Error())
	}
	return nil
}

// CreatePrinterCustomize 创建打印机定制
func (s *printerSrv) CreatePrinterCustomize(ctx context.Context, createPrinterCustomizeReq req.CreatePrinterCustomizeReq) (resp.CreatePrinterCustomizeResp, error) {
	db := ctx.GetDB()
	// 检查是否开启高级模版打印
	if ctx.GetCompanySetting().IsOpenAdvancedTicketPrint == 0 {
		return resp.CreatePrinterCustomizeResp{}, errors.WithMessage(errors.New("未开启高级模版打印"))
	}
	// 检查模板是否存在
	printerCustomizeRepo := repository.NewPrinterCustomizeRepo(db)
	template, err := repository.NewPrinterTemplateRepo(db).GetPrinterTemplateInfo(createPrinterCustomizeReq.TemplateId)
	if err != nil {
		return resp.CreatePrinterCustomizeResp{}, errors.WithMessage(errors.New("检查模板是否存在失败"), err.Error())
	}
	// 创建打印机定制
	customizeUuid, err := utils.GetID()
	if err != nil {
		return resp.CreatePrinterCustomizeResp{}, errors.WithMessage(errors.New("生成雪花ID失败"), err.Error())
	}
	err = printerCustomizeRepo.CreatePrinterCustomize(model.PrinterCustomize{
		BaseModel:  model.BaseModel{Uuid: customizeUuid},
		Name:       createPrinterCustomizeReq.Name,
		Data:       createPrinterCustomizeReq.Data,
		TemplateId: template.ID,
		IsAdv:      1,
	})
	if err != nil {
		return resp.CreatePrinterCustomizeResp{}, errors.WithMessage(errors.New("创建打印机定制失败"), err.Error())
	}
	return resp.CreatePrinterCustomizeResp{CustomizeUuid: customizeUuid}, nil
}

// UsePrinterCustomize 使用打印机定制
func (s *printerSrv) UsePrinterCustomize(ctx context.Context, customizeUuid uint64) error {
	db := ctx.GetDB()
	printerCustomizeRepo := repository.NewPrinterCustomizeRepo(db)
	// 检查打印机定制是否存在
	customizeInfo, err := printerCustomizeRepo.GetPrinterCustomizeInfo(customizeUuid)
	if err != nil {
		return errors.WithMessage(errors.New("检查打印机定制是否存在失败"), err.Error())
	}
	// 检查打印机定制是否正在使用中
	return db.Transaction(func(tx *gorm.DB) error {
		// 按template_id更新is_use为0, 其他字段不变
		err = repository.NewPrinterCustomizeRepo(tx).UpdatePrinterCustomizeByTemplateId(customizeInfo.TemplateId)
		if err != nil {
			return errors.WithMessage(errors.New("使用打印机定制失败"), err.Error())
		}

		// 使用打印机定制
		customizeInfo.IsUse = 1
		customizeInfo.UpdateTime = time.Now().Unix()
		err = repository.NewPrinterCustomizeRepo(tx).UpdatePrinterCustomize(customizeInfo)
		if err != nil {
			return errors.WithMessage(errors.New("更新打印机定制失败"), err.Error())
		}

		// 更新打印机模板
		err = repository.NewPrinterTemplateRepo(tx).UpdatePrinterTemplate(model.PrinterTemplate{
			ID:      customizeInfo.TemplateId,
			TmpUuid: customizeInfo.Uuid,
			TmpData: customizeInfo.Data,
		})
		if err != nil {
			return errors.WithMessage(errors.New("更新打印机定制失败"), err.Error())
		}

		return nil
	})
}

// GetConfigInfo 获取配置信息
func (s *printerSrv) GetPrinterCustomizeConfigInfo(ctx context.Context, configInfoReq req.PrinterGetConfigInfoReq) (resp.ConfigInfoResp, error) {
	db := ctx.GetDB()
	// 检查模板是否存在
	template, err := repository.NewPrinterTemplateRepo(db).GetPrinterTemplateInfo(configInfoReq.TemplateId)
	if err != nil {
		return resp.ConfigInfoResp{}, errors.WithMessage(errors.New("检查模板不存在"), err.Error())
	}

	// 获取模板名称
	templateName := utils.IfString(configInfoReq.IsAdv == 0, i18n.Translate(ctx.GetLanguage(), DefaultTemplateName), template.Name)

	// 获取模板JSON字符串
	defaultJsonStr, err := s.GetTemplateJSONStr(ctx, template.Name)
	if err != nil {
		return resp.ConfigInfoResp{}, errors.WithMessage(errors.New("获取模板JSON字符串失败"), err.Error())
	}

	// 格式化模板JSON字符串
	var templateObj map[string]interface{}
	if err := json.Unmarshal([]byte(defaultJsonStr), &templateObj); err == nil {
		if formattedJSON, err := json.Marshal(templateObj); err == nil {
			defaultJsonStr = string(formattedJSON)
		}
	}

	// 获取打印机定制
	templateJSONStr := defaultJsonStr
	if configInfoReq.CustomizeUuid != 0 {
		// 获取打印机定制信息
		customizeInfo, err := repository.NewPrinterCustomizeRepo(db).GetPrinterCustomizeInfo(configInfoReq.CustomizeUuid)
		if err != nil {
			return resp.ConfigInfoResp{}, errors.WithMessage(errors.New("检查打印机定制是否存在失败"), err.Error())
		}
		// 格式化模板JSON字符串
		var templateObj map[string]interface{}
		if err := json.Unmarshal([]byte(customizeInfo.Data), &templateObj); err == nil {
			if formattedJSON, err := json.Marshal(templateObj); err == nil {
				templateJSONStr = string(formattedJSON)
			}
		}
		//
		templateName = customizeInfo.Name
	}

	// 获取测试数据
	testData, err := s.GetTestData(ctx, template.Name)
	if err != nil {
		return resp.ConfigInfoResp{}, errors.WithMessage(errors.New("获取测试数据失败"), err.Error())
	}
	_, err = s.Parser(ctx, templateJSONStr, testData)
	if err != nil {
		return resp.ConfigInfoResp{}, errors.WithMessage(errors.New("解析模板失败"), err.Error())
	}
	testDataJSON, err := json.Marshal(testData)
	if err != nil {
		return resp.ConfigInfoResp{}, errors.WithMessage(errors.New("序列化测试数据失败"), err.Error())
	}

	// 获取模板配置信息
	templateConfigInfoStr, err := s.GetTemplateConfigInfo(ctx, template.Name, configInfoReq.IsAdv == 1)
	if err != nil {
		return resp.ConfigInfoResp{}, errors.WithMessage(errors.New("获取模板配置信息失败"), err.Error())
	}

	// 从模板JSON中提取 block 数据，并追加到配置信息的 group_blocks 的 data_rows 中
	extractedBlocks, err := s.ExtractBlocksFromTemplate(defaultJsonStr)
	if err != nil {
		logger.Logger.Warn("提取模板blocks失败", zap.Error(err))
	}

	// 格式化模板JSON字符串并添加 data_rows
	var templateConfigObj map[string]interface{}
	if err := json.Unmarshal([]byte(templateConfigInfoStr), &templateConfigObj); err == nil {
		// 如果有提取的blocks，将它们添加到配置信息的 data_rows 中
		if len(extractedBlocks) > 0 {
			if rows, ok := templateConfigObj["rows"].([]interface{}); ok {
				// 遍历每个 group
				for _, row := range rows {
					rowMap, ok := row.(map[string]interface{})
					if !ok {
						continue
					}

					groupBlocks, ok := rowMap["group_blocks"].([]interface{})
					if !ok {
						continue
					}

					// 遍历该组的每个 group_block
					for _, groupBlockInterface := range groupBlocks {
						groupBlockMap, ok := groupBlockInterface.(map[string]interface{})
						if !ok {
							continue
						}

						blockId, exists := groupBlockMap["block_id"].(string)
						if !exists {
							continue
						}

						// 从提取的blocks中找到匹配的 block_id，添加到 data_rows
						if templateBlocks, found := extractedBlocks[blockId]; found && len(templateBlocks) > 0 {
							dataRows, _ := groupBlockMap["data_rows"].([]interface{})
							if dataRows == nil {
								dataRows = make([]interface{}, 0)
							}

							// templateBlocks 是二维数组：[[第1行blocks], [第2行blocks]]
							// 直接展开添加到 data_rows，保持每一行作为一个数组
							// 结果：data_rows = [[第1行blocks], [第2行blocks]]
							dataRows = append(dataRows, templateBlocks...)

							groupBlockMap["data_rows"] = dataRows
						}
					}
				}
				templateConfigObj["rows"] = rows
			}
		}

		// 格式化JSON字符串
		if formattedJSON, err := json.Marshal(templateConfigObj); err == nil {
			templateConfigInfoStr = string(formattedJSON)
		}
	} else {
		return resp.ConfigInfoResp{}, errors.WithMessage(errors.New("格式化模板配置信息失败"), err.Error())
	}

	// 结果
	result := resp.ConfigInfoResp{
		ConfigListJson: templateConfigInfoStr,
		DefaultJson:    defaultJsonStr,
		CustomizeJson:  templateJSONStr,
		CustomizeName:  templateName,
		CustomizeData:  string(testDataJSON),
	}

	// 返回
	return result, nil
}

// ExtractBlocksFromTemplate 从模板JSON中提取 block 数据
// 返回 map[string][]interface{}，key 是 group_block_id（对应配置中的 block_id），value 是该 group_block_id 的所有 block 数据数组
func (s *printerSrv) ExtractBlocksFromTemplate(templateJSONStr string) (map[string][]interface{}, error) {
	var templateObj map[string]interface{}
	if err := json.Unmarshal([]byte(templateJSONStr), &templateObj); err != nil {
		return nil, errors.WithMessage(errors.New("解析模板JSON失败"), err.Error())
	}

	// 存储按 group_block_id 索引的blocks（因为配置中的 block_id 对应的是模板的 group_block_id）
	blocksByGroupBlockId := make(map[string][]interface{})

	// 获取 rows 数组
	rows, ok := templateObj["rows"].([]interface{})
	if !ok {
		return blocksByGroupBlockId, nil
	}

	// 递归遍历所有 blocks
	s.extractBlocksRecursive(rows, blocksByGroupBlockId)

	return blocksByGroupBlockId, nil
}

// extractBlocksRecursive 递归提取 blocks
func (s *printerSrv) extractBlocksRecursive(rows []interface{}, blocksByGroupBlockId map[string][]interface{}) {
	for _, row := range rows {
		rowArray, ok := row.([]interface{})
		if !ok {
			continue
		}

		// 检查当前行是否所有blocks都有相同的 group_block_id
		var currentGroupId string
		allSameGroup := true
		blocksInRow := make([]interface{}, 0)

		// 第一遍遍历：检查是否同组，并收集blocks
		for _, blockInterface := range rowArray {
			block, ok := blockInterface.(map[string]interface{})
			if !ok {
				allSameGroup = false
				break
			}

			if block["block_type"] == "blank_line" {
				continue
			}

			groupBlockId, exists := block["group_block_id"].(string)
			if !exists || groupBlockId == "" {
				allSameGroup = false
				break
			}

			if currentGroupId == "" {
				currentGroupId = groupBlockId
			} else if currentGroupId != groupBlockId {
				allSameGroup = false
				break
			}

			blocksInRow = append(blocksInRow, block)
		}

		// 如果当前行所有blocks都是同一个 group_block_id，将这一行作为一个数组添加
		if allSameGroup && currentGroupId != "" && len(blocksInRow) > 0 {
			// 将当前行的blocks作为一个独立的数组添加到该组
			// 格式：blocksByGroupBlockId["commodity"] = [[第1行blocks], [第2行blocks]]
			blocksByGroupBlockId[currentGroupId] = append(blocksByGroupBlockId[currentGroupId], blocksInRow)

			// 处理嵌套的 rows
			for _, blockInterface := range rowArray {
				block, _ := blockInterface.(map[string]interface{})
				if nestedRows, ok := block["rows"].([]interface{}); ok {
					s.extractBlocksRecursive(nestedRows, blocksByGroupBlockId)
				}
			}
		} else {
			// 如果不是同一组，按原来的逻辑单独处理每个block
			for _, blockInterface := range rowArray {
				block, ok := blockInterface.(map[string]interface{})
				if !ok {
					continue
				}

				if block["block_type"] == "blank_line" {
					continue
				}

				// 获取 group_block_id（这是配置文件中 block_id 的值）
				groupBlockId, exists := block["group_block_id"].(string)
				if !exists || groupBlockId == "" {
					// 如果没有 group_block_id，尝试递归检查是否有嵌套的 rows
					if nestedRows, ok := block["rows"].([]interface{}); ok {
						s.extractBlocksRecursive(nestedRows, blocksByGroupBlockId)
					}
					continue
				}

				// 单个block也作为一个数组添加（保持格式统一）
				blocksByGroupBlockId[groupBlockId] = append(blocksByGroupBlockId[groupBlockId], []interface{}{block})

				// 如果有嵌套的 rows，也递归处理
				if nestedRows, ok := block["rows"].([]interface{}); ok {
					s.extractBlocksRecursive(nestedRows, blocksByGroupBlockId)
				}
			}
		}
	}
}
