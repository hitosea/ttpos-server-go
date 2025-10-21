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
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"
)

type IPrinterSrv interface {
	GetProductPrinterList(ctx context.Context) (resp.ProductPrinterList, error)                              // 获取打印档口列表
	UsbPrinterReport(ctx context.Context, reportReq req.UsbPrinterReportReq) (resp.PrinterReportResp, error) // usb打印机上报
	// 获取打印菜单列表
	GetPrintMenuList(ctx context.Context) (resp.PrintMenuListResp, error)
	// 获取菜单详情
	GetPrintMenuDetail(ctx context.Context, id uint64) (resp.PrintMenuDetailResp, error)
	// 编辑打印机定制
	EditPrinterCustomize(ctx context.Context, id uint64, data string) error
	// 删除打印机定制
	DeletePrinterCustomize(ctx context.Context, id uint64) error
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

// GetPrintMenuList 获取打印菜单列表
func (s *printerSrv) GetPrintMenuList(ctx context.Context) (resp.PrintMenuListResp, error) {
	printerTemplateRepo := repository.NewPrinterTemplateRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	templates, err := printerTemplateRepo.GetPrinterTemplates()
	if err != nil {
		return resp.PrintMenuListResp{List: make([]resp.PrintMenuGroup, 0)}, errors.ErrInternal
	}

	// 转换为分组列表
	var groups []resp.PrintMenuGroup
	groups = append(groups, resp.PrintMenuGroup{
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
		List:      make([]resp.PrintMenu, 0),
	})
	groups = append(groups, resp.PrintMenuGroup{
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
		List:      make([]resp.PrintMenu, 0),
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
				groups[i].List = append(groups[i].List, resp.PrintMenu{
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
	return resp.PrintMenuListResp{List: groups}, nil
}

// 解析器
func (s *printerSrv) Parser(ctx context.Context, templateJSONStr string, testData map[string]interface{}) (string, error) {
	currencySetting, err := setting.NewSrv(s.dbm, s.cache).GetCurrencySetting(ctx)
	if err != nil {
		return "", errors.WithMessage(err, "获取打印设置失败")
	}

	// 创建解析器
	unitPosition, err := strconv.ParseInt(currencySetting.UnitPosition, 10, 64)
	if err != nil {
		return "", errors.WithMessage(err, "转换货币单位位置失败")
	}
	parser, err := pkg.NewImgTemplateParser(pkg.ImgBaseData{
		Language:             ctx.GetLanguage(),
		CurrencyUnit:         currencySetting.PrintUnit,
		CurrencyUnitPosition: int(unitPosition),
	}, templateJSONStr, testData)
	if err != nil {
		return "", errors.WithMessage(err, "创建模板解析器失败")
	}

	// 验证模板
	err = parser.ValidateTemplate()
	if err != nil {
		return "", errors.WithMessage(err, "验证模板失败")
	}

	// 解析模板
	img, err := parser.Parse()
	if err != nil {
		return "", errors.WithMessage(err, "解析模板失败")
	}

	// 保存测试图片
	path := "app/printer/pkg/text/tmp/printer/complex_template_test.png"

	// 设置分割高度为200000
	img.SegmentationHeight = 200000
	img.Save(path, false, 0)

	// 读取保存的图片文件并转换为base64
	imageData, err := os.ReadFile(path)
	if err != nil {
		return "", errors.WithMessage(err, "读取生成的图片文件失败")
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
	testDataBytes, err := os.ReadFile(fmt.Sprintf("app/printer/pkg/template_json/%s_data.json", templateName))
	if err != nil {
		return nil, errors.WithMessage(err, "读取测试数据文件失败")
	}
	var testData map[string]interface{}
	if err := json.Unmarshal(testDataBytes, &testData); err != nil {
		return nil, errors.WithMessage(err, "解析测试数据JSON失败")
	}
	return testData, nil
}

// GetTemplateJSONStr 获取模板JSON字符串
func (s *printerSrv) GetTemplateJSONStr(ctx context.Context, templateName string) (string, error) {
	templateJSON, err := os.ReadFile(fmt.Sprintf("app/printer/pkg/template_json/%s.json", templateName))
	if err != nil {
		return "", errors.WithMessage(err, "读取模板文件失败")
	}
	return string(templateJSON), nil
}

// GetPrintMenuDetail 获取菜单详情
func (s *printerSrv) GetPrintMenuDetail(ctx context.Context, id uint64) (resp.PrintMenuDetailResp, error) {
	db := ctx.GetDB()
	commonRepo := repository.NewCommonRepo()
	printerCustomizeRepo := repository.NewPrinterCustomizeRepo(db)

	// 获取打印模板详情
	template, err := repository.NewPrinterTemplateRepo(db).GetPrinterTemplateInfo(id)
	if err != nil {
		return resp.PrintMenuDetailResp{}, errors.WithMessage(err, "获取打印模板详情失败")
	}

	// 创建复杂的测试模板
	templateJSONStr, err := s.GetTemplateJSONStr(ctx, template.Name)

	// 获取打印机定制列表
	customizes, err := printerCustomizeRepo.GetPrinterCustomizeList(
		commonRepo.DBOption(commonRepo.WhereBySoftDelete()),
	)
	if err != nil {
		return resp.PrintMenuDetailResp{}, errors.WithMessage(err, "获取打印机定制列表失败")
	}

	// 获取测试数据
	testData, err := s.GetTestData(ctx, template.Name)
	if err != nil {
		return resp.PrintMenuDetailResp{}, errors.WithMessage(err, "获取测试数据失败")
	}

	// 默认模板
	defaultTemplate := resp.PrintMenuDetail{
		ID:      template.ID,
		Name:    "门店-默认模版",
		IsUse:   false,
		TmpUuid: template.TmpUuid,
	}

	// 高级模板列表(高级模版列表)
	advReceiptTpls := make([]resp.PrintMenuDetail, 0)
	for _, customize := range customizes {
		if customize.IsAdv == 1 {
			advReceiptTpls = append(advReceiptTpls, resp.PrintMenuDetail{
				ID:      customize.ID,
				Name:    customize.Name,
				IsUse:   customize.IsUse == 1,
				TmpUuid: customize.Uuid,
				TempImg: func() string {
					printContent, err := s.Parser(ctx, customize.Data, testData)
					if err != nil {
						return ""
					}
					return printContent
				}(),
			})
		} else if customize.Uuid == template.TmpUuid {
			defaultTemplate.Name = customize.Name
			defaultTemplate.IsUse = customize.IsUse == 1
			defaultTemplate.TmpUuid = customize.Uuid
			printContent, err := s.Parser(ctx, customize.Data, testData)
			if err != nil {
				return resp.PrintMenuDetailResp{}, errors.WithMessage(err, "解析模板失败")
			}
			defaultTemplate.TempImg = printContent
		}
	}

	// 默认模板没有设置，则使用门店默认模版解析
	if defaultTemplate.TempImg == "" {
		defaultTemplate.TempImg, err = s.Parser(ctx, templateJSONStr, testData)
		if err != nil {
			return resp.PrintMenuDetailResp{}, errors.WithMessage(err, "解析模板失败")
		}
	}

	// 返回结果
	return resp.PrintMenuDetailResp{
		DefaultTpl:      defaultTemplate,
		AdvReceiptTpls:  advReceiptTpls,
		IsAdvReceiptTpl: true,
	}, nil
}

// EditPrinterCustomize 编辑打印机定制
func (s *printerSrv) EditPrinterCustomize(ctx context.Context, id uint64, data string) error {
	db := ctx.GetDB()
	printerCustomizeRepo := repository.NewPrinterCustomizeRepo(db)
	// 检查打印机定制是否存在
	customizeInfo, err := printerCustomizeRepo.GetPrinterCustomizeInfo(id)
	if err != nil {
		return errors.WithMessage(err, "检查打印机定制是否存在失败")
	}
	//
	testData, err := s.GetTestData(ctx, customizeInfo.Name)
	if err != nil {
		return errors.WithMessage(err, "获取测试数据失败")
	}
	// 解析模板
	_, err = s.Parser(ctx, data, testData)
	if err != nil {
		return errors.WithMessage(err, "解析模板失败")
	}
	// 更新打印机定制
	return printerCustomizeRepo.UpdatePrinterCustomize(model.PrinterCustomize{
		ID:         id,
		Data:       data,
		UpdateTime: time.Now().Unix(),
	})
}

// DeletePrinterCustomize 删除打印机定制
func (s *printerSrv) DeletePrinterCustomize(ctx context.Context, id uint64) error {
	db := ctx.GetDB()
	// 检查打印机定制是否存在
	customizeInfo, err := repository.NewPrinterCustomizeRepo(db).GetPrinterCustomizeInfo(id)
	if err != nil {
		return errors.WithMessage(err, "检查打印机定制是否存在失败")
	}
	// 检查打印机定制是否正在使用中
	if customizeInfo.IsUse == 1 {
		return errors.New("打印机定制正在使用中，不能删除")
	}
	// 删除打印机定制
	err = repository.NewPrinterCustomizeRepo(db).DeletePrinterCustomize(id)
	if err != nil {
		return errors.WithMessage(err, "删除打印机定制失败")
	}
	return nil
}
