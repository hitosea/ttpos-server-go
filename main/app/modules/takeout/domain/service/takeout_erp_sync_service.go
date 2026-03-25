package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"ttpos-bmp/app/ttpos-erp/api/selling"

	"go.uber.org/zap"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	takeoutModel "ttpos-server-go/app/modules/takeout/domain/model"
	"ttpos-server-go/app/modules/takeout/infrastructure/adapter/rpc"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"

	appContext "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/logger"

	"gorm.io/gorm"
)

// ITakeoutErpSyncService ERP 同步领域服务接口
// @version v2.12.0
// @spec story-erp-grab-invoice-sync
type ITakeoutErpSyncService interface {
	// SyncOrderToERP 同步外卖订单到 ERP
	SyncOrderToERP(ctx appContext.Context, orderUuid uint64) error
	// SyncOrderCancelledToERP 同步订单取消到 ERP
	SyncOrderCancelledToERP(ctx appContext.Context, orderUuid uint64, remarkData string) error
	// ResyncOrderToERP 重新同步外卖订单到 ERP（订单变更后使用）
	ResyncOrderToERP(ctx appContext.Context, orderUuid uint64) error
}

// takeoutErpSyncService ERP 同步领域服务实现
type takeoutErpSyncService struct{}

// NewTakeoutErpSyncService 创建 ERP 同步领域服务
func NewTakeoutErpSyncService() ITakeoutErpSyncService {
	return &takeoutErpSyncService{}
}

// SyncOrderToERP 同步外卖订单到 ERP
// @version v2.12.0
// @spec story-erp-grab-invoice-sync
func (s *takeoutErpSyncService) SyncOrderToERP(ctx appContext.Context, orderUuid uint64) error {
	// 2. 获取数据库管理器和连接
	db := ctx.GetDB()
	if db == nil {
		return fmt.Errorf("数据库连接失败")
	}

	company := ctx.GetCompany()
	companySetting := ctx.GetCompanySetting()

	// 5. 检查 ERP 集成是否启用
	if companySetting.ErpnextSiteCode == "" {
		logger.Logger.Info("公司未启用 ERP 集成，跳过同步",
			zap.Uint64("companyUuid", ctx.GetCompanyUuid()),
			zap.Uint64("orderUuid", orderUuid))
		return nil
	}

	// 3. 查询外卖订单完整信息
	takeoutOrderRepo := persistence.NewTakeoutOrderRepo(db)
	takeoutOrder, err := takeoutOrderRepo.GetByUuid(
		orderUuid,
		takeoutOrderRepo.WithTakeoutOrderItems(),
		takeoutOrderRepo.WithTakeoutOrderItemModifiers(),
		takeoutOrderRepo.WithTakeoutOrderMaterials(),
	)
	if err != nil {
		return errors.WithMessage(err, "查询外卖订单失败")
	}
	if takeoutOrder == nil {
		return errors.WithMessage(errors.New("外卖订单不存在"), fmt.Sprintf("外卖订单不存在: %d", orderUuid))
	}

	// 检查订单是否已同步到 ERP
	if takeoutOrder.IsErpInvoiceSynced() {
		return nil // 如果订单已同步到 ERP，则不重复同步
	}

	// 判断是否使用 Sales Invoice 模式：公司配置为 SI 模式且关联班次为新版
	useSalesInvoice := false
	var erpOpenPosEntryName string
	if companySetting.IsErpSalesInvoiceMode() {
		shiftLog, _ := repository.NewShiftLogRepo(db).GetShiftLog(func(q *gorm.DB) *gorm.DB {
			return q.Where("uuid = ?", takeoutOrder.StaffShiftLogUuid)
		})
		useSalesInvoice = shiftLog.Uuid == 0 || shiftLog.IsNewShiftVersion()
	}
	if !useSalesInvoice {
		// POS Invoice 模式需要班次和 OpenPosEntry
		if !takeoutOrder.IsExistShiftLog() {
			return nil
		}
		erpOpenPosEntryName, err = takeoutOrderRepo.GetErpOpenPosEntryNameByOrderUuid(takeoutOrder.Uuid)
		if erpOpenPosEntryName == "" {
			return errors.WithMessage(err, "查询 ERP 开账名称失败")
		}
	}

	// 检查支付方式是否已存在（通过 payment_name 或 code）
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)
	existPayment := repository.NewPaymentMethodRepo(db).GetPaymentMethod(
		paymentMethodRepo.WhereCode(
			func() int {
				if takeoutOrder.IsGrabOrder() {
					return constant.PaymentMethodCodeGrab
				}
				return constant.PaymentMethodCodeLineMan
			}(),
		),
	)
	if existPayment.Uuid == 0 {
		logger.Logger.Error("支付方式不存在，跳过同步", zap.Int("paymentCode", constant.PaymentMethodCodeGrab))
		return nil
	}

	// 10. 创建 ERP 上下文
	erpCtx := appContext.NewContext(
		appContext.WithContext(ctx.GetContext()),
		appContext.WithCompanyUuid(ctx.GetCompanyUuid()),
		appContext.WithCompany(*company),
		appContext.WithCompanySetting(companySetting),
		appContext.WithStaff(ctx.GetStaff()),
		appContext.WithStaffUuid(ctx.GetStaffUuid()),
		appContext.WithLogger(logger.Logger),
	)
	erpCtx.SetDB(db)

	erpAdapter := rpc.NewErpRpcAdapter(database.GetDBManager(config.Database))

	if useSalesInvoice {
		// Sales Invoice 模式
		saveSalesInvoiceReq := buildSalesInvoiceRequest(takeoutOrder, &companySetting, &existPayment, ctx.GetCompanyUuid())
		saveSalesInvoiceResp, err := erpAdapter.SaveSalesInvoice(erpCtx, saveSalesInvoiceReq)
		if err != nil {
			logger.Logger.Error("同步外卖订单到 ERP(SI模式)失败",
				zap.Uint64("orderUuid", orderUuid),
				zap.String("platformOrderId", takeoutOrder.PlatformOrderId),
				zap.Error(err))
			return errors.WithMessage(err, "保存 Sales Invoice 失败")
		}
		respJson, err := json.Marshal(saveSalesInvoiceResp)
		if err != nil {
			logger.Logger.Error("序列化 ERP 响应数据失败", zap.Uint64("orderUuid", orderUuid), zap.Error(err))
		} else {
			if err := takeoutOrderRepo.UpdateByMap(takeoutOrder.Uuid, map[string]interface{}{
				"erp_pos_invoice_resp": string(respJson),
				"erp_sync_status":      1, // 已入队待处理，SI/PE 名称由 BMP consumer 异步回写
			}); err != nil {
				logger.Logger.Error("保存 ERP 响应数据到订单失败", zap.Uint64("orderUuid", orderUuid), zap.Error(err))
			}
		}
		logger.Logger.Info("成功同步外卖订单到 ERP(SI模式)",
			zap.Uint64("orderUuid", orderUuid),
			zap.String("platformOrderId", takeoutOrder.PlatformOrderId),
			zap.String("salesInvoiceName", saveSalesInvoiceResp.SalesInvoiceName))
	} else {
		// POS Invoice 模式（默认）
		savePosInvoiceReq := buildPosInvoiceRequest(takeoutOrder, &companySetting, erpOpenPosEntryName, &existPayment)
		savePosInvoiceResp, err := erpAdapter.SavePosInvoice(erpCtx, savePosInvoiceReq)
		if err != nil {
			logger.Logger.Error("同步外卖订单到 ERP 失败",
				zap.Uint64("orderUuid", orderUuid),
				zap.String("platformOrderId", takeoutOrder.PlatformOrderId),
				zap.Error(err))
			return errors.WithMessage(err, "保存 POS Invoice 失败")
		}
		respJson, err := json.Marshal(savePosInvoiceResp)
		if err != nil {
			logger.Logger.Error("序列化 ERP 响应数据失败", zap.Uint64("orderUuid", orderUuid), zap.Error(err))
		} else {
			if err := takeoutOrderRepo.UpdateByMap(takeoutOrder.Uuid, map[string]interface{}{
				"erp_pos_invoice_resp": string(respJson),
			}); err != nil {
				logger.Logger.Error("保存 ERP 响应数据到订单失败", zap.Uint64("orderUuid", orderUuid), zap.Error(err))
			}
		}
		logger.Logger.Info("成功同步外卖订单到 ERP",
			zap.Uint64("orderUuid", orderUuid),
			zap.String("platformOrderId", takeoutOrder.PlatformOrderId),
			zap.String("erpProductsInvoiceName", savePosInvoiceResp.ProductsInvoiceName),
			zap.String("erpMaterialInvoiceName", savePosInvoiceResp.MaterialInvoiceName))
	}

	return nil
}

// buildPosInvoiceRequest 构建 POS Invoice 请求参数
// @version v2.12.0
// @spec story-erp-grab-invoice-sync
func buildPosInvoiceRequest(
	takeoutOrder *takeoutModel.TakeoutOrder,
	companySetting *model.CompanySetting,
	erpOpenPosEntryName string,
	existPayment *model.PaymentMethod,
) req.SavePosInvoiceReq {
	// 构建商品列表
	items := buildPosInvoiceItems(takeoutOrder)

	// 构建支付列表
	payments := buildPosInvoicePayments(takeoutOrder, existPayment)

	// 构建税费列表
	taxes := buildPosInvoiceTaxes(takeoutOrder)

	// 构建原材料列表（从 takeoutOrder.TakeoutOrderMaterials 中读取）
	materialItems := buildPosInvoiceMaterialItems(takeoutOrder)

	// 会员 UUID（外卖平台订单通常无会员）
	customerUuid := ""

	takeoutProvider := takeoutOrder.GetNoSpacePlatformName()

	return req.SavePosInvoiceReq{
		SiteCode:         companySetting.ErpnextSiteCode,
		OrderNo:          strconv.FormatUint(takeoutOrder.Uuid, 10), // 使用订单UUID作为 ERP 订单号
		OpenPosEntryName: erpOpenPosEntryName,
		PostingDatetime:  takeoutOrder.AcceptedTime, // 使用接单时间作为过账时间
		CustomerUuid:     customerUuid,
		Items:            items,
		MaterialItems:    materialItems,
		Taxes:            taxes,
		Payments:         payments,
		TakeoutOrderNo:   &takeoutOrder.PlatformOrderId,
		TakeoutProvider:  &takeoutProvider, // Grab, LineMan
	}
}

// buildSalesInvoiceRequest 构建 Sales Invoice 请求参数（SI 模式）
func buildSalesInvoiceRequest(
	takeoutOrder *takeoutModel.TakeoutOrder,
	companySetting *model.CompanySetting,
	existPayment *model.PaymentMethod,
	companyUuid uint64,
) req.SaveSalesInvoiceReq {
	items := buildPosInvoiceItems(takeoutOrder)
	payments := buildPosInvoicePayments(takeoutOrder, existPayment)
	taxes := buildPosInvoiceTaxes(takeoutOrder)
	materialItems := buildPosInvoiceMaterialItems(takeoutOrder)

	takeoutProvider := takeoutOrder.GetNoSpacePlatformName()

	return req.SaveSalesInvoiceReq{
		SiteCode:          companySetting.ErpnextSiteCode,
		OrderNo:           strconv.FormatUint(takeoutOrder.Uuid, 10),
		SaleOrderUuid:     strconv.FormatUint(takeoutOrder.Uuid, 10),
		CompanyAbbr:       companySetting.ErpnextCompanyAbbr,
		Customer:          "Default",
		Currency:          "THB",
		PriceListCurrency: "THB",
		PostingDatetime:   takeoutOrder.AcceptedTime,
		Branch:            companySetting.ErpnextBranchName,
		UpdateStock:       0, // 延迟扣库存
		Items:             items,
		MaterialItems:     materialItems,
		Taxes:             taxes,
		Payments:          payments,
		TakeoutOrderNo:    &takeoutOrder.PlatformOrderId,
		TakeoutProvider:   &takeoutProvider,
		CompanyUuid:       strconv.FormatUint(companyUuid, 10),
		OrderType:         "takeout",
	}
}

// buildPosInvoiceItems 构建 POS Invoice 商品列表
// @version v2.12.0
// @spec story-erp-grab-invoice-sync
// @reference main/app/service/order.go:4307-4432
func buildPosInvoiceItems(takeoutOrder *takeoutModel.TakeoutOrder) []*selling.PosInvoiceItem {
	items := make([]*selling.PosInvoiceItem, 0)

	// 遍历订单商品
	for _, item := range takeoutOrder.TakeoutOrderItems {
		packageName := language.JsonToLocaleResponse(item.TtposItemName) // 套餐名称
		// 如果是套餐商品
		if item.IsPackage() {
			// 套餐使用固定虚拟商品编码
			items = append(items, &selling.PosInvoiceItem{
				ItemCode:    "TC001",                          // 套餐虚拟商品编码
				Qty:         float64(item.Quantity),           // 套餐数量
				Rate:        item.GetPriceNoneTax(),           // 套餐单价
				Amount:      item.GetTotalPriceNoneTaxTotal(), // 套餐总价
				Description: packageName.EN,                   // 套餐名称
				IsFreeItem: func() bool {
					if item.Price == 0 {
						return true
					}
					return false
				}(),
			})
			// 套餐子商品（commodity 类型的 modifier）
			for _, modifier := range item.TakeoutOrderItemModifiers {
				if modifier.IsCommodity() {
					modifierName := language.JsonToLocaleResponse(modifier.ModifierName) // 套餐子商品名称
					commodityItemCode := modifier.TtposModifierErpCode
					if commodityItemCode == "" {
						commodityItemCode = constant.PosInvoiceItemCodeSpareGoods // 未映射套餐子商品使用备用商品编码
					}
					items = append(items, &selling.PosInvoiceItem{
						ItemCode:    commodityItemCode,
						Qty:         float64(item.Quantity),                              // 子商品数量
						Rate:        0,                                                   // 套餐子商品没有单价
						Amount:      0,                                                   // 套餐子商品没有金额
						Description: fmt.Sprintf("Sales in package:%s", modifierName.EN), // 套餐子商品描述
						IsFreeItem:  true,                                                // 套餐子商品标记为免费
					})
				}
			}
		} else {
			// 普通商品
			itemCode := item.TtposItemErpCode
			if itemCode == "" {
				itemCode = constant.PosInvoiceItemCodeSpareGoods // 未映射商品使用备用商品编码
			}
			items = append(items, &selling.PosInvoiceItem{
				ItemCode:    itemCode,
				Qty:         float64(item.Quantity),           // 数量
				Rate:        item.GetPriceNoneTax(),           // 单价
				Amount:      item.GetTotalPriceNoneTaxTotal(), // 小计
				Description: packageName.EN,                   // 商品名称
				IsFreeItem: func() bool {
					if item.Price == 0 {
						return true
					}
					return false
				}(),
			})
		}

		// 处理修饰符（规格、加料、属性）
		// 注意：套餐的 commodity 类型已在上面处理，这里只处理其他类型
		for _, modifier := range item.TakeoutOrderItemModifiers {
			// 跳过套餐子商品（已处理）
			if modifier.IsCommodity() {
				continue
			}
			modifierName := language.JsonToLocaleResponse(modifier.ModifierName) // 修饰符名称
			// 规格(flavor)、加料(sauce)、属性(attr) 都作为独立行项添加
			// 参考 order.go:4421-4431 的小料处理逻辑
			if modifier.IsSauce() {
				sauceItemCode := modifier.TtposModifierErpCode
				if sauceItemCode == "" {
					sauceItemCode = constant.PosInvoiceItemCodeSpareGoods // 未映射加料使用备用商品编码
				}
				items = append(items, &selling.PosInvoiceItem{
					ItemCode:    sauceItemCode,
					Qty:         float64(item.Quantity), // 修饰符数量
					Rate:        0,                      // 修饰符没有单独单价（已计入主商品）
					Amount:      0,                      // 修饰符没有单独金额
					Description: modifierName.EN,        // 规格名称
					IsFreeItem:  true,                   // 修饰符标记为免费（价格已包含在主商品中）
				})
			}
		}
	}

	// 添加虚拟商品：服务
	if takeoutOrder.MerchantChargeFee > 0 {
		items = append(items, &selling.PosInvoiceItem{
			ItemCode: constant.PosInvoiceItemCodeServiceFee, // VP001 服务费
			Qty:      1,
			Rate:     takeoutOrder.MerchantChargeFee,
			Amount:   takeoutOrder.MerchantChargeFee,
		})
	}

	// 添加虚拟商品：小单费用（Small Order Fee）
	if takeoutOrder.SmallOrderFee > 0 {
		items = append(items, &selling.PosInvoiceItem{
			ItemCode: constant.PosInvoiceItemCodeServiceFee, // VP001 服务费
			Qty:      1,
			Rate:     takeoutOrder.SmallOrderFee,
			Amount:   takeoutOrder.SmallOrderFee,
		})
	}

	// 添加虚拟商品：配送费
	// 因为Grab的配送费是作为服务费同步的
	// NOTE: 产品 -》 目前模拟器不传配送费，生产传了配送费但是不记录在total。所以：所有订单都不传配送费到ERP
	// if takeoutOrder.DeliveryFee > 0 && takeoutOrder.IsRestaurantDeliveryOrder() {
	// 	items = append(items, &selling.PosInvoiceItem{
	// 		ItemCode: constant.PosInvoiceItemCodeDeliveryFee, // VP003 配送费
	// 		Qty:      1,
	// 		Rate:     takeoutOrder.DeliveryFee,
	// 		Amount:   takeoutOrder.DeliveryFee,
	// 	})
	// }

	// LINEMAN 特殊处理：计算商品合计与 restaurantRevenue 的差额
	// restaurantRevenue = Grand Total (THB)
	// 差额处理：
	// 1. 如果商品合计 > restaurantRevenue（少收了），差额作为优惠
	// 2. 如果商品合计 < restaurantRevenue（多收了），差额作为服务费
	if takeoutOrder.IsLinemanOrder() {
		diff := takeoutOrder.GetLinemanServiceFee()
		if diff < -0.01 {
			// 商品合计 < restaurantRevenue：差额作为负数服务费
			items = append(items, &selling.PosInvoiceItem{
				ItemCode:    constant.PosInvoiceItemCodeServiceFee, // VP001 服务费
				Qty:         1,
				Rate:        -diff,
				Amount:      -diff,
				Description: "LINEMAN Service Fee",
			})
		}
	}

	return items
}

// buildPosInvoiceTaxes 构建 POS Invoice 税费列表
// @version v2.12.0
// @spec story-erp-grab-invoice-sync
// @reference main/app/service/order.go:4445-4521
func buildPosInvoiceTaxes(takeoutOrder *takeoutModel.TakeoutOrder) []*selling.PosInvoiceTax {
	taxes := make([]*selling.PosInvoiceTax, 0)

	// 1. 税费（Tax）
	if takeoutOrder.Tax > 0 {
		taxes = append(taxes, &selling.PosInvoiceTax{
			TaxAmount:   takeoutOrder.Tax,
			Description: "Tax", // 消费税
		})
	}

	// 2. 商户优惠（Merchant Discount）- 作为负数税费
	if takeoutOrder.MerchantDiscount > 0 {
		taxes = append(taxes, &selling.PosInvoiceTax{
			TaxAmount:   -takeoutOrder.MerchantDiscount,
			Description: "Merchant Discount", // 商户优惠
		})
	}

	// LINEMAN 特殊处理：整单改价（Whole Order Price Adjustment）
	if takeoutOrder.IsLinemanOrder() {
		diff := takeoutOrder.GetLinemanServiceFee()
		if diff > 0.01 {
			taxes = append(taxes, &selling.PosInvoiceTax{
				TaxAmount:   -diff,
				Description: "Merchant Discount", // 商户优惠
			})
		}
	}

	return taxes
}

// buildPosInvoicePayments 构建 POS Invoice 支付列表
// @version v2.12.0
// @spec story-erp-grab-invoice-sync
func buildPosInvoicePayments(takeoutOrder *takeoutModel.TakeoutOrder, existPayment *model.PaymentMethod) []*selling.PosInvoicePayment {
	payments := make([]*selling.PosInvoicePayment, 0)
	payments = append(payments, &selling.PosInvoicePayment{
		ModeOfPayment: existPayment.ErpnextPayment, // 支付方式
		Amount:        takeoutOrder.PlatformTotal,  // 平台结算总额
	})
	return payments
}

// buildPosInvoiceMaterialItems 构建 POS Invoice 原材料列表
// @version v2.12.0
// @spec story-erp-grab-invoice-sync
// @reference main/app/service/order.go:4445-4455
func buildPosInvoiceMaterialItems(takeoutOrder *takeoutModel.TakeoutOrder) []*selling.PosInvoiceItem {
	materialItems := make([]*selling.PosInvoiceItem, 0)

	// 如果没有原材料记录，直接返回空列表
	if len(takeoutOrder.TakeoutOrderMaterials) == 0 {
		return materialItems
	}
	// 按 ErpCode 和 Unit 去重并累加数量
	splitKey := "--@--"
	materialMap := make(map[string]float64) // key: erp_code--@--unit, value: 原材料数量
	uomMap := make(map[string]string)       // key: erp_code--@--unit, value: unit name
	for _, material := range takeoutOrder.TakeoutOrderMaterials {
		// 如果原材料信息缺失或 ErpCode 为空，跳过
		if material.ErpCode == "" {
			logger.Logger.Warn("外卖订单原材料缺少 ErpCode，跳过",
				zap.Uint64("orderUuid", takeoutOrder.Uuid),
				zap.Uint64("materialUuid", material.MaterialUuid),
				zap.String("materialName", material.MaterialName))
			continue
		}
		// 构建 key
		key := fmt.Sprintf("%s%s%s", material.ErpCode, splitKey, material.BaseUnitUom)
		materialMap[key] += material.Num
		uomMap[key] = material.BaseUnitUom
	}

	// 转换为列表
	for key, num := range materialMap {
		parts := strings.Split(key, splitKey)
		erpCode := parts[0]
		materialItems = append(materialItems, &selling.PosInvoiceItem{
			ItemCode: erpCode,
			Qty:      num,         // 原材料数量
			Uom:      uomMap[key], // 原材料单位
			Rate:     0,           // 原材料没有单价
			Amount:   0,           // 原材料没有金额
		})
	}

	return materialItems
}

// SyncOrderCancelledToERP 同步订单取消到 ERP
// @version v2.16.0
// @spec story-erp-invoice-cancel-notification
//
// 参数说明：
//   - ctx: 上下文（包含数据库连接、公司信息等）
//   - orderUuid: 订单UUID
//   - remarkData: 附注数据（JSON格式字符串），包含以下字段：
//   - shop_uuid: 商户UUID（必填，用于 MQ 回调时路由到正确的数据库）
//   - order_type: 订单类型（可选，如 "takeout"）
//   - should_resync: 是否需要重新创建发票（订单变更=true，订单取消=false）
//
// 功能说明：
//  1. 调用 BMP CancelPosInvoice 接口取消发票
//  2. 在 Remark 字段中传递 remarkData（JSON格式）
//  3. BMP 取消成功后，会将 Remark 原样返回给 ttpos（通过 MQ）
//  4. ttpos 消费者解析 Remark 获取 shop_uuid 和 should_resync，决定后续动作
func (s *takeoutErpSyncService) SyncOrderCancelledToERP(ctx appContext.Context, orderUuid uint64, remarkData string) error {
	// 1. 获取数据库连接
	db := ctx.GetDB()
	if db == nil {
		return fmt.Errorf("数据库连接失败")
	}

	// 2. 获取公司信息
	company := ctx.GetCompany()
	companySetting := ctx.GetCompanySetting()

	// 3. 检查 ERP 集成是否启用
	if !company.IsOpenErp() || companySetting.ErpnextSiteCode == "" {
		logger.Logger.Info("公司未启用 ERP 集成或未配置站点编码，跳过同步", zap.Uint64("companyUuid", ctx.GetCompanyUuid()), zap.Uint64("orderUuid", orderUuid))
		return nil
	}

	// 4. 查询外卖订单信息
	takeoutOrderRepo := persistence.NewTakeoutOrderRepo(db)
	takeoutOrder, err := persistence.NewTakeoutOrderRepo(db).GetByUuid(orderUuid)
	if err != nil {
		return errors.WithMessage(err, "查询外卖订单失败")
	}
	if takeoutOrder == nil {
		return errors.WithMessage(errors.New("外卖订单不存在"), fmt.Sprintf("外卖订单不存在: %d", orderUuid))
	}

	// 5. 检查订单是否已同步到 ERP
	if len(takeoutOrder.ErpPosInvoiceResp) == 0 {
		logger.Logger.Info("外卖订单未同步到 ERP，无需取消发票",
			zap.Uint64("orderUuid", orderUuid),
			zap.String("platformOrderId", takeoutOrder.PlatformOrderId))
		return nil
	}

	erpAdapter := rpc.NewErpRpcAdapter(database.GetDBManager(config.Database))

	if companySetting.IsErpSalesInvoiceMode() {
		// SI 模式：解析 SI 响应，调用 CancelSalesInvoice
		siResp := takeoutOrder.GetErpSalesInvoiceResp()
		salesInvoiceName := ""
		var paymentEntryNames []string
		if siResp != nil {
			salesInvoiceName = siResp.SalesInvoiceName
			paymentEntryNames = siResp.PaymentEntryNames
		}
		err = erpAdapter.CancelSalesInvoice(ctx, req.CancelSalesInvoiceReq{
			SiteCode:          companySetting.ErpnextSiteCode,
			OrderNo:           strconv.FormatUint(takeoutOrder.Uuid, 10),
			SaleOrderUuid:     strconv.FormatUint(takeoutOrder.Uuid, 10),
			SalesInvoiceName:  salesInvoiceName,
			PaymentEntryNames: paymentEntryNames,
			StockDeducted:     takeoutOrder.ErpStockDeducted == 1,
			Remark:            remarkData,
		})
		if err != nil {
			logger.Logger.Error("取消外卖订单 ERP 发票失败(SI模式)",
				zap.Uint64("orderUuid", orderUuid),
				zap.String("platformOrderId", takeoutOrder.PlatformOrderId),
				zap.String("salesInvoiceName", salesInvoiceName),
				zap.Error(err))
			return errors.WithMessage(err, "取消 Sales Invoice 失败")
		}
	} else {
		// POS Invoice 模式：需要班次信息
		erpResp := takeoutOrder.GetErpPosInvoiceResp()

		if takeoutOrder.StaffShiftLogUuid == 0 {
			logger.Logger.Warn("外卖订单没有班次信息，无法取消 ERP 发票",
				zap.Uint64("orderUuid", orderUuid),
				zap.String("platformOrderId", takeoutOrder.PlatformOrderId))
			return nil
		}

		shiftLog, err := takeoutOrderRepo.GetShiftLogByOrderUuid(orderUuid)
		if err != nil {
			return errors.WithMessage(err, "查询班次信息失败")
		}
		if shiftLog.IsHandedOver() {
			return errors.New("当前班次已交班，无法取消发票")
		}

		err = erpAdapter.CancelPosInvoice(ctx, req.CancelPosInvoiceReq{
			ProductsInvoiceName: func() string {
				if erpResp != nil {
					return erpResp.ProductsInvoiceName
				}
				return ""
			}(),
			MaterialInvoiceName: func() string {
				if erpResp != nil {
					return erpResp.MaterialInvoiceName
				}
				return ""
			}(),
			OpenPosEntryName: shiftLog.ErpnextOpenPosEntryName,
			OrderNo:          strconv.FormatUint(takeoutOrder.Uuid, 10),
			Remark:           remarkData,
		})
		if err != nil {
			logger.Logger.Error("取消外卖订单 ERP 发票失败",
				zap.Uint64("orderUuid", orderUuid),
				zap.String("platformOrderId", takeoutOrder.PlatformOrderId),
				zap.Error(err))
			return errors.WithMessage(err, "取消 ERP 发票失败")
		}
	}

	// 清空订单的 erp_pos_invoice_resp 字段，使 SyncOrderToERP 能够重新同步
	if err := takeoutOrderRepo.UpdateByMap(orderUuid, map[string]interface{}{
		"erp_pos_invoice_resp": "",
	}); err != nil {
		logger.Logger.Error("清空订单 ERP 响应字段失败",
			zap.Uint64("orderUuid", orderUuid),
			zap.Error(err))
		return errors.WithMessage(err, "清空订单 ERP 响应字段失败")
	}

	return nil
}

// ResyncOrderToERP 重新同步外卖订单到 ERP（订单变更后使用）
// @version v2.16.0
// @spec story-erp-invoice-cancel-notification
//
// 异步实现策略（基于 MQ 回调）：
//  1. 调用 SyncOrderCancelledToERP 取消旧发票（在 Remark 中传递 shop_uuid 等路由信息）
//  2. BMP 取消成功后会发送 MQ 消息（topic: erp-invoice-cancel）
//  3. MQ 消费者（ErpCancelInvoiceCallbackHandler）收到消息，解析 Remark 获取 shop_uuid
//  4. 消费者调用 SyncOrderToERP 创建新发票
//
// 注意：本方法只负责发起取消请求，不等待结果。新发票的创建由 MQ 消费者异步完成。
func (s *takeoutErpSyncService) ResyncOrderToERP(ctx appContext.Context, orderUuid uint64) error {
	db := ctx.GetDB()
	if db == nil {
		return fmt.Errorf("数据库连接失败")
	}

	companySetting := ctx.GetCompanySetting()

	// 1. 检查 ERP 集成是否启用
	if companySetting.ErpnextSiteCode == "" {
		logger.Logger.Info("公司未启用 ERP 集成，跳过重新同步",
			zap.Uint64("companyUuid", ctx.GetCompanyUuid()),
			zap.Uint64("orderUuid", orderUuid))
		return nil
	}

	// 2. 查询订单，检查是否已同步到 ERP
	takeoutOrderRepo := persistence.NewTakeoutOrderRepo(db)
	takeoutOrder, err := takeoutOrderRepo.GetByUuid(orderUuid)
	if err != nil {
		return errors.WithMessage(err, "查询外卖订单失败")
	}
	if takeoutOrder == nil {
		return errors.WithMessage(errors.New("外卖订单不存在"), fmt.Sprintf("外卖订单不存在: %d", orderUuid))
	}

	// 3. 检查订单是否已同步到 ERP
	if !takeoutOrder.IsErpInvoiceSynced() {
		logger.Logger.Info("外卖订单未同步到 ERP，跳过重新同步",
			zap.Uint64("orderUuid", orderUuid),
			zap.String("platformOrderId", takeoutOrder.PlatformOrderId))
		return nil
	}

	// 4. 构建 Remark 字段（包含 shop_uuid 等路由信息，用于 BMP 回调）
	remarkData := map[string]any{
		"shop_uuid":     ctx.GetCompanyUuid(),
		"order_type":    "takeout",
		"should_resync": true, // 标记需要重新创建发票
	}
	remarkJSON, err := json.Marshal(remarkData)
	if err != nil {
		logger.Logger.Error("序列化 Remark 字段失败", zap.Uint64("orderUuid", orderUuid), zap.Error(err))
		return errors.WithMessage(err, "序列化 Remark 字段失败")
	}

	// 5. 取消旧发票（异步模式：BMP 取消成功后会发送 MQ 消息，由消费者调用 SyncOrderToERP）
	if err := s.SyncOrderCancelledToERP(ctx, orderUuid, string(remarkJSON)); err != nil {
		logger.Logger.Error("取消旧发票失败",
			zap.Uint64("orderUuid", orderUuid),
			zap.String("platformOrderId", takeoutOrder.PlatformOrderId),
			zap.Error(err))
		return errors.WithMessage(err, "取消旧发票失败")
	}

	return nil
}
