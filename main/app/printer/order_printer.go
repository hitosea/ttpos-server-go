package printer

import (
	"fmt"
	"slices"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

/**
 * 菜品打印
 * @param $order 订单
 * @param string $printType 打印类型 -1-为退菜打印 --付款打印 1-送厨打印
 */
func (p *PrinterRepoImpl) PrintingDishes(printType int, saleBillUuid uint64, products Products) bool {
	// 设置时区
	tz := utils.SetTimezone(p.storeSetting.TimeZone)
	fmt.Println(tz.FormatUnixTimeDefault(1739283862))
	fmt.Println("printerSetting")
	fmt.Println(utils.ToJsonString(p.printerSetting))

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
	db := p.dbm.GetDB(p.Ctx.GetCompanyUuid())
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
		if !slices.Contains(productPrinter.GetPrinterRegionUuids(), regionUuid) {
			continue
		}
		// 选中的商品才往下走
		productIds := productPrinter.GetPrinterProductIds()
		newProducts := make(Products, 0)
		for _, product := range products {
			if !slices.Contains(productIds, product.ProductId) {
				continue
			}
			newProducts = append(newProducts, product)
		}
		// 是否整单打印
		isCompleteOrderPrinter := productPrinter.PrintMethod == constant.No
		//
		// 分开打印
		// if productPrinter.PrintModeScene == constant.Yes {

		// }

		fmt.Println(utils.ToJsonString(productPrinter.ProductPrinterItems))

		// 循环下拉选中的打印机 - 一个个打印
		for _, printerItem := range productPrinter.ProductPrinterItems {
			if !printerItem.IsDelete() {
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
	product OrderProduct,
) string {

	fmt.Println(utils.ToJsonString(printerItem))

	// $printerType = $printerItem->printer['printer_type']['value'] ?? '';
	// //
	// if (($printerConfig['kitchen_print_method'] ?? 1) == 2) {
	// 	return (new ImgDishesTemplate(null, $this->allSourceProductList))->oneDishOneOrder($printerConfig, $printerItem, $order, $products);
	// }
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
