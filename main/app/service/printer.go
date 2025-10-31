package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	respSetting "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/pkg"
	"ttpos-server-go/app/printer/pkg/template_json"
	"ttpos-server-go/app/printer/template"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
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
	groups = append(groups, resp.PrintTemplateGroup{
		LocaleName: dto.LocaleResponse{
			ZH:   "厨房小票",
			EN:   "Kitchen Menu",
			TH:   "เมนูอาหาร",
			JA:   "キッチンメニュー",
			KO:   "주방 메뉴",
			MY:   "Resipi Makanan",
			TR:   "Yemek Menüsü",
			SV:   "Maträtmeny",
			ZHTW: "廚房菜單",
		},
		GroupType: 2,
		List:      make([]resp.PrintTemplate, 0),
	})

	// 模版ID列表
	templateOrders := []uint64{
		constant.PrinterTemplatePreBilling,    // 预结账单
		constant.PrinterTemplateBilling,       // 结账单
		constant.PrinterTemplateInvoice,       // 发票
		constant.PrinterTemplateRecharge,      // 充值单
		constant.PrinterTemplateBusiness,      // 营业数据
		constant.PrinterTemplateHandoverSheet, // 交班单
		constant.PrinterTemplateTakeoutOrder,  // 外送单
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

// GetTestData 获取测试数据
func (s *printerSrv) GetTestData(ctx context.Context, templateName string) (map[string]interface{}, error) {
	// 从JSON文件读取测试数据
	testDataBytes, err := template_json.GetTemplateJsonData(templateName + "_data.json")
	if err != nil {
		return nil, errors.WithMessage(errors.New("读取测试数据文件失败"), err.Error())
	}
	var testData map[string]interface{}
	if err := json.Unmarshal(testDataBytes, &testData); err != nil {
		return nil, errors.WithMessage(errors.New("解析测试数据JSON失败"), err.Error())
	}
	if testData["store"] != nil && testData["store"].(map[string]interface{})["logo"] != nil {
		settingSrv := setting.NewSrvImpl(s.dbm, s.cache)
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
		logoAddr := template.NewPrinterTemplate(ctx, settingSrv, &storeSetting, &printerSetting, &currencySetting, false, ctx.GetLanguage()).GetLogoAddr()
		if logoAddr != "" {
			testData["store"].(map[string]interface{})["logo"] = logoAddr
		}
	}
	return testData, nil
}

// GetTemplateJSONStr 获取模板JSON字符串
func (s *printerSrv) GetTemplateJSONStr(ctx context.Context, templateName string) (string, error) {
	templateJSON, err := template_json.GetTemplateJsonData(templateName + "_tmp.json")
	if err != nil {
		return "", errors.WithMessage(errors.New("读取模板文件失败"), err.Error())
	}
	return string(templateJSON), nil
}

// GetTemplateConfigInfo 获取模板配置信息
func (s *printerSrv) GetTemplateConfigInfo(ctx context.Context, templateName string, isAdv bool) (string, error) {
	templateJSON, err := template_json.GetTemplateJsonData(templateName + "_config.json")
	if err != nil {
		return "", errors.WithMessage(errors.New("读取模板文件失败"), err.Error())
	}
	return string(templateJSON), nil
}

// GetPrintTemplateDetail 获取菜单详情
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

	// 默认模板
	defaultTemplate := resp.PrintTemplateDetail{
		ID:    template.ID,
		Name:  i18n.Translate(ctx.GetLanguage(), DefaultTemplateName),
		IsUse: false,
	}

	// 高级模板列表(高级模版列表)
	advReceiptTpls := make([]resp.PrintTemplateDetail, 0)
	for _, customize := range customizes {
		if customize.IsAdv == 1 {
			advReceiptTpls = append(advReceiptTpls, resp.PrintTemplateDetail{
				ID:            template.ID,
				Name:          customize.Name,
				IsUse:         customize.IsUse == 1,
				CustomizeUuid: customize.Uuid,
				ImgUrl: func() string {
					printContent, err := s.Parser(ctx, customize.Data, testData)
					if err != nil {
						return ""
					}
					return printContent
				}(),
			})
		} else if customize.TemplateId == template.ID {
			defaultTemplate.Name = utils.IfString(customize.Name == DefaultTemplateName, i18n.Translate(ctx.GetLanguage(), DefaultTemplateName), customize.Name)
			defaultTemplate.IsUse = customize.IsUse == 1
			defaultTemplate.CustomizeUuid = customize.Uuid
			printContent, err := s.Parser(ctx, customize.Data, testData)
			if err != nil {
				return resp.PrintTemplateDetailResp{}, errors.WithMessage(errors.New("解析模板失败"), err.Error())
			}
			defaultTemplate.ImgUrl = printContent
		}
	}

	// 默认模板没有设置，则使用门店默认模版解析
	if defaultTemplate.ImgUrl == "" && defaultTemplate.CustomizeUuid == 0 {
		defaultTemplate.ImgUrl, err = s.Parser(ctx, templateJSONStr, testData)
		if err != nil {
			return resp.PrintTemplateDetailResp{}, errors.WithMessage(errors.New("解析模板失败"), err.Error())
		}
		// 创建打印机定制
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
		exists, err := printerCustomizeRepo.CheckPrinterCustomizeNameExists(customizeInfo.Uuid, editPrinterCustomizeReq.Name)
		if err != nil {
			return errors.WithMessage(errors.New("检查打印机定制名称是否存在失败"), err.Error())
		}
		if exists {
			return errors.New("高级模版名称已存在")
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
	db.Transaction(func(tx *gorm.DB) error {
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
	})

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
	// 检查打印机定制名称是否存在
	exists, err := printerCustomizeRepo.CheckPrinterCustomizeNameExists(0, createPrinterCustomizeReq.Name)
	if err != nil {
		return resp.CreatePrinterCustomizeResp{}, errors.WithMessage(errors.New("检查打印机定制名称是否存在失败"), err.Error())
	}
	if exists {
		return resp.CreatePrinterCustomizeResp{}, errors.WithMessage(errors.New("高级模版名称已存在"))
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
		if extractedBlocks != nil && len(extractedBlocks) > 0 {
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

							// 将每个 template block 作为单独的数组添加到 data_rows
							// 格式：data_rows = [[block1], [block2], [block3]]
							for _, templateBlock := range templateBlocks {
								dataRows = append(dataRows, []interface{}{templateBlock})
							}

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

		// 遍历每行的blocks
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

			// 将 block 添加到对应的 group_block_id 数组中
			blocksByGroupBlockId[groupBlockId] = append(blocksByGroupBlockId[groupBlockId], block)

			// 如果有嵌套的 rows，也递归处理
			if nestedRows, ok := block["rows"].([]interface{}); ok {
				s.extractBlocksRecursive(nestedRows, blocksByGroupBlockId)
			}
		}
	}
}
