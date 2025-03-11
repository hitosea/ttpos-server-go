package printer

import (
	"slices"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/printer_model"
	"ttpos-server-go/app/printer/template"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

/**
 * 菜品打印
 * @param int $printType 打印类型 -1-为退菜打印 0-付款打印 1-送厨打印
 */
func (p *PrinterRepoImpl) PrintingDishes(
	printType int,
	saleBillUuid uint64,
	products printer_model.Products,
) bool {
	// 设置时区
	// tz := utils.SetTimezone(p.storeSetting.TimeZone)
	// fmt.Println(tz.FormatUnixTimeDefault(1739283862))
	// fmt.Println("printerSetting")
	// fmt.Println(utils.ToJsonString(p.printerSetting))

	// 获取商品打印机列表
	productPrinters, err := p.getProductPrinterList()
	if err != nil {
		logger.Logger.Error("获取商品打印机列表失败", zap.Error(err))
		return false
	}
	if len(productPrinters) == 0 {
		return false
	}

	// 获取订单信息
	db := p.dbm.GetDB(p.ctx.GetCompanyUuid())
	billInfo, err := repository.NewOrderRepo(db).GetSaleBillInfo(saleBillUuid, constant.No)
	if err != nil {
		return false
	}

	// 当前订单对应的区域id
	regionUuid := uint64(0)
	if billInfo.Desk != nil {
		regionUuid = billInfo.Desk.RegionUuid
	}

	// 循环商品打印机
	for _, productPrinter := range productPrinters {
		// 送厨打印才走
		if productPrinter.PrintMode != printType {
			continue
		}

		// 区域对的上才走
		regionUuids := productPrinter.GetPrinterRegionUuids()
		if len(regionUuids) == 0 || !slices.Contains(regionUuids, regionUuid) {
			continue
		}

		// 选中的商品才往下走
		productIds := productPrinter.GetPrinterProductIds()
		newProducts := make(printer_model.Products, 0)
		for _, product := range products {
			if !slices.Contains(productIds, product.ProductId) {
				continue
			}
			newProducts = append(newProducts, product)
		}

		// 是否整单打印
		isCompleteOrderPrinter := productPrinter.PrintMethod == constant.No

		// 循环下拉选中的打印机一个个打印
		for _, printerItem := range productPrinter.ProductPrinterItems {
			// 判断是否删除
			if printerItem.IsDelete() {
				continue
			}
			// 退菜单打印
			if printType == constant.PrinterProductTypeBackFood {
				// $data = $this->getPrintReturnProductContent($printerConfig, $printerItem, $order);
				// if ($data) {
				// 	PrinterLog::addPrinterLog($printer, array_merge($printerData, [
				// 		"data" => $data,
				// 		// "data_type" => PrinterLog::DATA_TYPE[9]['value'],
				// 	]));
				// }
				continue
			}
			// 一菜一单打印
			if !isCompleteOrderPrinter {
				for _, product := range newProducts {
					data := p.getPrintProductOneContent(productPrinter, printerItem, billInfo, product)
					if data != "" {
						// PrinterLog::addPrinterLog($printer, array_merge($printerData, [
						// 	"data" => $data,
						// 	"data_type" => PrinterLog::DATA_TYPE[3]['value'],
						// ]));
					}
				}
				continue
			}
			// 整单打印
			// $data = $this->getPrintProductContent($printerConfig, $printerItem, $order);
			// if ($data) {
			// 	PrinterLog::addPrinterLog($printer, array_merge($printerData, [
			// 		"data" => $data,
			// 		"data_type" => PrinterLog::DATA_TYPE[4]['value'],
			// 	]));
			// }
		}
	}
	// 打印
	return true
}

// 构建订单菜品（一菜一单）打印的内容
func (p *PrinterRepoImpl) getPrintProductOneContent(
	productPrinter model.ProductPrinter,
	printerItem *model.ProductPrinterItem,
	saleBill model.SaleBill,
	product printer_model.OrderProduct,
) string {

	// 图片打印
	if p.printerSetting.KitchenPrintMethod == "2" {
		// 	return (new ImgDishesTemplate(null, $this->allSourceProductList))->oneDishOneOrder($printerConfig, $printerItem, $order, $products);
	}

	// 获取打印机类型
	var printerType string
	if printerItem.Printer != nil && printerItem.Printer.PrinterType != nil {
		printerType = printerItem.Printer.PrinterType.Key
	}

	// CODESOFT 打印机
	if printerItem.Printer != nil && slices.Contains([]string{
		constant.PrinterTypeCodesoftLan,
		constant.PrinterTypeCodesoftWifi,
	}, printerType) {
		// 创建Codesoft模板
		t := template.NewDishesCodesoftTemplate(p.ctx, p.setting, &p.storeSetting, &p.printerSetting, &p.currencySetting)
		// 调用CompleteOrder方法
		return t.CompleteOrder(printerItem, saleBill, []printer_model.OrderProduct{product})
	}

	// 商米和芯烨打印机
	if printerItem.Printer != nil {
		// return (new XprinterDishesTemplate(null, $this->allSourceProductList))->oneDishOneOrder($printerConfig, $printerItem, $order, $products);
	}

	// PrinterTypeFeiEYun       = "FEI_E_YUN"      // 飞鹅打印机
	// PrinterTypeFeiEYunTag    = "FEI_E_YUN_TAG"  // 飞鹅标签打印机
	// PrinterTypePrintCenter   = "PRINT_CENTER"   // 365云打印
	// PrinterTypeSunmiLan      = "SUNMI_LAN"      // 商米 局域网内打印
	// PrinterTypeSunmiCloud    = "SUNMI_CLOUD"    // 商米 云打印
	// PrinterTypeXPrinterLan   = "XPRINTER_LAN"   // 芯烨-有线
	// PrinterTypeXPrinterWifi  = "XPRINTER_WIFI"  // 芯烨-WIFI
	// PrinterTypeCashierCompax = "CASHIER_COMPAX" // Compax 收银打印机 80mm 自带
	// PrinterTypeCashierSunmi  = "CASHIER_SUNMI"  // SUNMI 商米 收银打印机 80mm 自带
	// PrinterTypeCodesoftLan   = "CODESOFT_LAN"   // Codesoft（网口）80mm
	// PrinterTypeCodesoftWifi  = "CODESOFT_WIFI"  //Codesoft（WIFI）80mm

	// fmt.Println(utils.ToJsonString(printerItem.Printer))

	// /* *
	// * CODESOFT 打印机
	// */
	// if ($printerItem->printer && in_array($printerType, [PrinterTypeEnum::CODESOFT_LAN, PrinterTypeEnum::CODESOFT_WIFI])) {
	// 	return (new CodesoftDishesTemplate(null, $this->allSourceProductList))->oneDishOneOrder($printerConfig, $printerItem, $order, $products);
	// }

	// /* *
	// *商米 和 芯烨 打印机
	// */
	// if ($printerItem->printer) {
	// 	return (new XprinterDishesTemplate(null, $this->allSourceProductList))->oneDishOneOrder($printerConfig, $printerItem, $order, $products);
	// }

	return ""
}
