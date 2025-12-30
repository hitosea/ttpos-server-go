package service

import (
	"encoding/json"
	"fmt"
	"strconv"

	"ttpos-bmp/app/ttpos-erp/api/selling"

	"go.uber.org/zap"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/model"
	takeoutModel "ttpos-server-go/app/modules/takeout/domain/model"
	valueobject "ttpos-server-go/app/modules/takeout/domain/value_object"
	"ttpos-server-go/app/modules/takeout/infrastructure/adapter/rpc"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/context"
	appContext "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/logger"
)

// ITakeoutErpSyncService ERP 同步领域服务接口
// @version v2.12.0
// @spec story-erp-grab-invoice-sync
type ITakeoutErpSyncService interface {
	// SyncOrderToERP 同步外卖订单到 ERP
	SyncOrderToERP(ctx context.Context, orderUuid uint64) error
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
	)
	if err != nil {
		return fmt.Errorf("查询外卖订单失败: %w", err)
	}
	if takeoutOrder == nil {
		return fmt.Errorf("外卖订单不存在: %d", orderUuid)
	}
	if takeoutOrder.OrderState != valueobject.TakeoutOrderStateAccepted && takeoutOrder.OrderState != valueobject.TakeoutOrderStateRiderPending {
		return fmt.Errorf("外卖订单状态不正确: %d", takeoutOrder.OrderState)
	}
	// 检查订单是否已同步到 ERP
	if len(takeoutOrder.ErpPosInvoiceResp) > 0 {
		return nil // 如果订单已同步到 ERP，则不重复同步
	}
	// 如果订单没有班次，则不同步 ERP
	if takeoutOrder.StaffShiftLogUuid == 0 {
		return nil
	}

	// 4. 查询 ERP 开账名称
	erpOpenPosEntryName, err := takeoutOrderRepo.GetErpOpenPosEntryNameByOrderUuid(takeoutOrder.Uuid)
	if erpOpenPosEntryName == "" {
		return fmt.Errorf("查询 ERP 开账名称失败: %w", err)
	}

	// 9. 构建 POS Invoice 请求参数
	savePosInvoiceReq := buildPosInvoiceRequest(
		takeoutOrder,
		company.CompanySetting,
		erpOpenPosEntryName,
	)

	// 10. 创建 ERP 上下文
	erpCtx := appContext.NewContext(
		appContext.WithContext(ctx.GetContext()),
		appContext.WithCompanyUuid(ctx.GetCompanyUuid()),
		appContext.WithCompany(*company),
		appContext.WithCompanySetting(*company.CompanySetting),
		appContext.WithStaff(ctx.GetStaff()),
		appContext.WithStaffUuid(ctx.GetStaffUuid()),
		appContext.WithLogger(logger.Logger),
	)
	erpCtx.SetDB(db)

	// 11. 调用 ERP RPC 适配器保存 POS Invoice
	savePosInvoiceResp, err := rpc.NewErpRpcAdapter(database.GetDBManager(config.Database)).SavePosInvoice(erpCtx, savePosInvoiceReq)
	if err != nil {
		logger.Logger.Error("同步 Grab 订单到 ERP 失败",
			zap.Uint64("orderUuid", orderUuid),
			zap.String("platformOrderId", takeoutOrder.PlatformOrderId),
			zap.Error(err))
		return fmt.Errorf("保存 POS Invoice 失败: %w", err)
	}

	// 12. 保存 ERP 响应数据到订单
	respJson, err := json.Marshal(savePosInvoiceResp)
	if err != nil {
		logger.Logger.Error("序列化 ERP 响应数据失败",
			zap.Uint64("orderUuid", orderUuid),
			zap.Error(err))
	} else {
		// 更新订单的 ERP 响应字段
		if err := takeoutOrderRepo.UpdateByMap(takeoutOrder.Uuid, map[string]interface{}{
			"erp_pos_invoice_resp": string(respJson),
		}); err != nil {
			logger.Logger.Error("保存 ERP 响应数据到订单失败",
				zap.Uint64("orderUuid", orderUuid),
				zap.Error(err))
		}
	}

	// 13. 记录 ERP 发票信息到日志
	logger.Logger.Info("成功同步 Grab 订单到 ERP",
		zap.Uint64("orderUuid", orderUuid),
		zap.String("platformOrderId", takeoutOrder.PlatformOrderId),
		zap.String("erpProductsInvoiceName", savePosInvoiceResp.ProductsInvoiceName),
		zap.String("erpMaterialInvoiceName", savePosInvoiceResp.MaterialInvoiceName))

	return nil
}

// buildPosInvoiceRequest 构建 POS Invoice 请求参数
// @version v2.12.0
// @spec story-erp-grab-invoice-sync
func buildPosInvoiceRequest(
	takeoutOrder *takeoutModel.TakeoutOrder,
	companySetting *model.CompanySetting,
	erpOpenPosEntryName string,
) req.SavePosInvoiceReq {
	// 构建商品列表
	items := buildPosInvoiceItems(takeoutOrder)

	// 构建支付列表
	payments := buildPosInvoicePayments(takeoutOrder)

	// 构建税费列表
	taxes := buildPosInvoiceTaxes(takeoutOrder)

	// 构建原材料列表（外卖订单暂不支持原材料扣减）
	materialItems := make([]*selling.PosInvoiceItem, 0)

	// 会员 UUID（外卖平台订单通常无会员）
	customerUuid := ""

	takeoutProvider := takeoutOrder.GetCapitalPlatform()

	return req.SavePosInvoiceReq{
		SiteCode:         companySetting.ErpnextSiteCode,
		OrderNo:          takeoutOrder.PlatformOrderId, // 使用平台订单号作为 ERP 订单号
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
				ItemCode:    "TC001",                // 套餐虚拟商品编码
				Qty:         float64(item.Quantity), // 套餐数量
				Rate:        item.Price,             // 套餐单价
				Amount:      item.GetTotalPrice(),   // 套餐总价
				Description: packageName.EN,         // 套餐名称
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
					items = append(items, &selling.PosInvoiceItem{
						ItemCode:    modifier.TtposModifierErpCode,                       // 使用 ERP Code
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
			items = append(items, &selling.PosInvoiceItem{
				ItemCode:    item.TtposItemErpCode,  // 使用 ERP Code
				Qty:         float64(item.Quantity), // 数量
				Rate:        item.Price,             // 单价
				Amount:      item.GetTotalPrice(),   // 小计
				Description: packageName.EN,         // 商品名称
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
				items = append(items, &selling.PosInvoiceItem{
					ItemCode:    modifier.TtposModifierErpCode, // 使用 ERP Code
					Qty:         float64(item.Quantity),        // 修饰符数量
					Rate:        0,                             // 修饰符没有单独单价（已计入主商品）
					Amount:      0,                             // 修饰符没有单独金额
					Description: modifierName.EN,               // 规格名称
					IsFreeItem:  true,                          // 修饰符标记为免费（价格已包含在主商品中）
				})
			}
		}
	}

	// 添加虚拟商品：配送费
	if takeoutOrder.DeliveryFee > 0 {
		items = append(items, &selling.PosInvoiceItem{
			ItemCode: constant.PosInvoiceItemCodeDeliveryFee, // VP003
			Qty:      1,
			Rate:     takeoutOrder.DeliveryFee,
			Amount:   takeoutOrder.DeliveryFee,
		})
	}

	// 添加虚拟商品：小单费用（Small Order Fee）
	if takeoutOrder.SmallOrderFee > 0 {
		items = append(items, &selling.PosInvoiceItem{
			ItemCode: constant.PosInvoiceItemCodeSmallOrderFee, // 小单费虚拟商品编码
			Qty:      1,
			Rate:     takeoutOrder.SmallOrderFee,
			Amount:   takeoutOrder.SmallOrderFee,
		})
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

	// 2. 商户服务费（Merchant Charge Fee）
	if takeoutOrder.MerchantChargeFee > 0 {
		taxes = append(taxes, &selling.PosInvoiceTax{
			TaxAmount:   takeoutOrder.MerchantChargeFee,
			Description: "Merchant Service Fee", // 商户服务费
		})
	}

	// 3. 平台优惠（Platform Discount）- 作为负数税费
	if takeoutOrder.PlatformDiscount > 0 {
		taxes = append(taxes, &selling.PosInvoiceTax{
			TaxAmount:   -takeoutOrder.PlatformDiscount,
			Description: "Platform Discount", // 平台优惠
		})
	}

	// 4. 商户优惠（Merchant Discount）- 作为负数税费
	if takeoutOrder.MerchantDiscount > 0 {
		taxes = append(taxes, &selling.PosInvoiceTax{
			TaxAmount:   -takeoutOrder.MerchantDiscount,
			Description: "Merchant Discount", // 商户优惠
		})
	}

	// 5. 购物车优惠（Basket Promo）- 作为负数税费
	if takeoutOrder.BasketPromo > 0 {
		taxes = append(taxes, &selling.PosInvoiceTax{
			TaxAmount:   -takeoutOrder.BasketPromo,
			Description: "Basket Promo", // 购物车优惠
		})
	}

	return taxes
}

// buildPosInvoicePayments 构建 POS Invoice 支付列表
// @version v2.12.0
// @spec story-erp-grab-invoice-sync
func buildPosInvoicePayments(takeoutOrder *takeoutModel.TakeoutOrder) []*selling.PosInvoicePayment {
	payments := make([]*selling.PosInvoicePayment, 0)

	if takeoutOrder.IsGrabOrder() {
		paymentID := strconv.Itoa(constant.PaymentMethodCodeGrab)
		payments = append(payments, &selling.PosInvoicePayment{
			ModeOfPayment: constant.PaymentMethodNameGrab, // 支付方式
			Amount:        takeoutOrder.EaterPayment,      // 实付金额
			PaymentId:     &paymentID,                     // 支付方式唯一标识
		})
	}
	if takeoutOrder.IsLinemanOrder() {
		paymentID := strconv.Itoa(constant.PaymentMethodCodeLineMan)
		payments = append(payments, &selling.PosInvoicePayment{
			ModeOfPayment: constant.PaymentMethodNameLineMan, // 支付方式
			Amount:        takeoutOrder.EaterPayment,         // 实付金额
			PaymentId:     &paymentID,                        // 支付方式唯一标识
		})
	}

	return payments
}
