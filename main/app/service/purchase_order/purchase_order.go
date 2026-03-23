package purchase_order

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"ttpos-bmp/app/ttpos-erp/api/buying"
	"ttpos-bmp/app/ttpos-erp/api/stock"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"
	"unicode/utf8"

	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IPurchaseOrderSrv 采购申请服务接口
type IPurchaseOrderSrv interface {
	// 采购申请管理
	GetPurchaseOrderList(ctx context.Context, req req.PurchaseOrderListReq) (resp.PurchaseOrderListResp, error)            // 获取采购申请列表
	GetPurchaseOrderDetail(ctx context.Context, req req.PurchaseOrderDetailReq) (resp.PurchaseOrderDetailResp, error)      // 获取采购申请详情
	CreatePurchaseOrder(ctx context.Context, req req.PurchaseOrderCreateReq) (resp.PurchaseOrderCreateResp, error)         // 创建采购申请
	UpdatePurchaseOrder(ctx context.Context, req req.PurchaseOrderUpdateReq) error                                         // 更新采购申请
	UpdatePurchaseOrderItemUnit(ctx context.Context, req req.PurchaseOrderDetailReq) (resp.PurchaseOrderDetailResp, error) // 更新采购订单物品单位
	DeletePurchaseOrder(ctx context.Context, req req.PurchaseOrderDeleteReq) error                                         // 删除采购申请
	SubmitPurchaseOrder(ctx context.Context, req req.PurchaseOrderSubmitReq) error                                         // 提交采购申请
	ApprovePurchaseOrder(ctx context.Context, req req.PurchaseOrderApproveReq) error                                       // 审核采购申请

	// 收货管理
	CreatePurchaseReceiptOrder(ctx context.Context, req req.PurchaseReceiptCreateReq) (resp.PurchaseReceiptOrderCreateResp, error)         // 创建收货单
	GetPurchaseReceiptOrderList(ctx context.Context, req req.PurchaseReceiptOrderListReq) (resp.PurchaseReceiptOrderListResp, error)       // 获取收货单列表
	GetPurchaseReceiptOrderDetail(ctx context.Context, req req.PurchaseReceiptOrderDetailReq) (resp.PurchaseReceiptOrderDetailResp, error) // 获取收货单详情
	UpdatePurchaseReceiptOrder(ctx context.Context, req req.PurchaseReceiptOrderUpdateReq) error                                           // 更新收货单
	CancelPurchaseReceiptOrder(ctx context.Context, req req.PurchaseReceiptOrderCancelReq) error                                           // 取消收货单
	GetReceiptPendingItems(ctx context.Context, req req.ReceiptPendingItemsReq) (resp.ReceiptPendingItemsResp, error)                      // v2.16.0+ 获取待收货物品
	GetPurchaseReceiptNewList(ctx context.Context, req req.PurchaseReceiptNewListReq) (resp.PurchaseReceiptNewListResp, error)             // 新收货单列表（按采购单维度）

	// 品牌采购自动审批设置
	GetBrandPurchaseAutoApprove(ctx context.Context) (int, error)     // 获取品牌采购自动审批开关
	SetBrandPurchaseAutoApprove(ctx context.Context, value int) error // 设置品牌采购自动审批开关
}

// purchaseOrderSrv 采购申请服务实现
type purchaseOrderSrv struct {
	dbm        *database.DBManager
	validator  *purchaseOrderValidator
	helper     *purchaseOrderHelper
	receiptSrv *purchaseReceiptOrderSrv
	lock       lock.Lock
	settingSrv setting.ISrv
}

// NewPurchaseOrderSrv 创建采购申请服务
func NewPurchaseOrderSrv(dbm *database.DBManager, settingSrv setting.ISrv) IPurchaseOrderSrv {
	return NewPurchaseOrderSrvImpl(dbm, settingSrv)
}

// NewPurchaseOrderSrvImpl 创建采购申请服务实现
func NewPurchaseOrderSrvImpl(dbm *database.DBManager, settingSrv setting.ISrv) IPurchaseOrderSrv {
	return &purchaseOrderSrv{
		dbm:        dbm,
		validator:  &purchaseOrderValidator{},
		helper:     &purchaseOrderHelper{},
		receiptSrv: newPurchaseReceiptOrderSrv(dbm),
		lock:       lock.NewSystemLock(),
		settingSrv: settingSrv,
	}
}

// GetPurchaseOrderList 获取采购申请列表
func (s *purchaseOrderSrv) GetPurchaseOrderList(
	ctx context.Context,
	req req.PurchaseOrderListReq,
) (resp.PurchaseOrderListResp, error) {

	db := ctx.GetDB()
	purchaseOrderRepo := repository.NewPurchaseOrderRepo(db)

	// 构建查询选项
	var opts []repository.DBOption
	if req.OrderNo != "" {
		opts = append(opts, purchaseOrderRepo.WhereOrderNo(req.OrderNo))
	}
	if len(req.StatusIn) > 0 {
		opts = append(opts, purchaseOrderRepo.WhereStatusIn(req.StatusIn))
	}
	if req.SupplierName != "" {
		opts = append(opts, purchaseOrderRepo.WhereSupplierName(req.SupplierName))
	}
	if req.WarehouseErpCode != "" {
		opts = append(opts, purchaseOrderRepo.WhereWarehouseErpCode(req.WarehouseErpCode))
	}
	if req.CompanyUuid > 0 {
		opts = append(opts, purchaseOrderRepo.WhereCompanyUuid(req.CompanyUuid))
	}
	if req.CreateTimeStart > 0 || req.CreateTimeEnd > 0 {
		opts = append(opts, purchaseOrderRepo.WhereCreateTimeRange(req.CreateTimeStart, req.CreateTimeEnd))
	}
	if req.OrderTimeStart > 0 || req.OrderTimeEnd > 0 {
		opts = append(opts, purchaseOrderRepo.WhereOrderTimeRange(req.OrderTimeStart, req.OrderTimeEnd))
	}
	if req.ExpectArrivalTimeStart > 0 || req.ExpectArrivalTimeEnd > 0 {
		opts = append(opts, purchaseOrderRepo.WhereExpectArrivalTimeRange(req.ExpectArrivalTimeStart, req.ExpectArrivalTimeEnd))
	}
	if req.ReceiveTimeStart > 0 || req.ReceiveTimeEnd > 0 {
		opts = append(opts, purchaseOrderRepo.WhereReceiveTimeRange(req.ReceiveTimeStart, req.ReceiveTimeEnd))
	}
	if len(req.UuidIn) > 0 {
		opts = append(opts, purchaseOrderRepo.WhereUuidIn(req.UuidIn))
	}

	// 采购类型
	purchaseType := utils.IfInt(req.PurchaseType == 2, 2, 1)
	opts = append(opts, purchaseOrderRepo.WherePurchaseType(purchaseType))

	// 排序和预加载
	opts = append(opts, purchaseOrderRepo.OrderByOrderTime(true))
	opts = append(opts, purchaseOrderRepo.WithItems())

	// 查询数据
	purchaseOrders, total, err := purchaseOrderRepo.GetListWithPagination(req.PageNo, req.PageSize, opts...)
	if err != nil {
		return resp.PurchaseOrderListResp{}, errors.WithMessage(errors.New("查询采购申请列表失败"), err.Error())
	}

	// 收集需要调用ERP的新采购单（有erp_sale_order_no）
	erpOrderNoToProgress := make(map[string]float64)
	var erpOrderNos []string
	for _, po := range purchaseOrders {
		if po.ErpSaleOrderNo != "" && po.ErpOrderNo != "" {
			erpOrderNos = append(erpOrderNos, po.ErpOrderNo)
		}
	}

	// 批量调用ERP获取收货进度
	if len(erpOrderNos) > 0 {
		erpSrv := erp.NewIErpSrv(s.dbm)
		erpResp, err := erpSrv.GetPurchaseOrderList(ctx, &buying.GetPurchaseOrderListReq{
			Name: strings.Join(erpOrderNos, ","),
		})
		if err != nil {
			logger.Logger.Warn("批量调用ERP获取采购单列表失败，使用本地计算",
				zap.Int("count", len(erpOrderNos)),
				zap.Error(err),
			)
		} else {
			for _, erpOrder := range erpResp.PurchaseOrders {
				erpOrderNoToProgress[erpOrder.Name] = erpOrder.PerReceived
			}
		}
	}

	// 转换响应数据
	listResp := make([]*resp.PurchaseOrderInfo, 0, len(purchaseOrders))
	for _, po := range purchaseOrders {
		poInfo := &resp.PurchaseOrderInfo{}
		if err := copier.Copy(poInfo, &po); err != nil {
			continue
		}
		// 计算收货进度：新采购单优先使用ERP返回的进度，旧采购单使用本地计算
		if po.ErpSaleOrderNo != "" && po.ErpOrderNo != "" {
			if progress, ok := erpOrderNoToProgress[po.ErpOrderNo]; ok {
				poInfo.ReceiptProgress = fmt.Sprintf("%.2f%%", progress)
			} else {
				poInfo.ReceiptProgress = fmt.Sprintf("%.2f%%", po.GetReceiptProgress())
			}
		} else {
			poInfo.ReceiptProgress = fmt.Sprintf("%.2f%%", po.GetReceiptProgress())
		}
		listResp = append(listResp, poInfo)
	}

	return resp.PurchaseOrderListResp{
		List: listResp,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

// GetPurchaseOrderDetail 获取采购申请详情
func (s *purchaseOrderSrv) GetPurchaseOrderDetail(
	ctx context.Context,
	req req.PurchaseOrderDetailReq,
) (resp.PurchaseOrderDetailResp, error) {
	db := ctx.GetDB()
	purchaseOrderRepo := repository.NewPurchaseOrderRepo(db)

	// 查询采购申请详情
	purchaseOrder, err := purchaseOrderRepo.GetByUuid(
		req.Uuid,
		purchaseOrderRepo.WithItems(),
		purchaseOrderRepo.WithWarehouse(),
		purchaseOrderRepo.WithSupplier(),
	)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return resp.PurchaseOrderDetailResp{}, errors.New("采购申请不存在")
		}
		return resp.PurchaseOrderDetailResp{}, errors.WithMessage(errors.New("查询采购申请详情失败"), err.Error())
	}

	// 转换响应数据
	var detailResp resp.PurchaseOrderDetailResp
	if err = copier.Copy(&detailResp, purchaseOrder); err != nil {
		return resp.PurchaseOrderDetailResp{}, errors.WithMessage(errors.New("数据转换失败"), err.Error())
	}

	// 获取子店所有仓库物品库存（门店数量）\ 总店指定仓库物品库存（可采购数量）
	avaliableQuantityMap := make(map[uint64]float64)
	storeQuantityMap := make(map[uint64]float64)
	companySetting := ctx.GetCompanySetting()

	var hqUuid, subUuid uint64
	if companySetting.IsHeadquarter() { // 总店
		hqUuid = companySetting.CompanyUuid
		subUuid = purchaseOrder.CompanyUuid
	} else if companySetting.IsSubShop() { // 子店
		hqUuid = companySetting.HeadquarterUuid
		subUuid = companySetting.CompanyUuid
	}
	// 获取子店所有仓库物品库存（门店数量）
	if subUuid != 0 {
		subWarehouseItemRepo := repository.NewWarehouseItemRepo(s.dbm.GetDB(subUuid))
		warehouseItems, _ := subWarehouseItemRepo.GetItemsInActiveWarehouses()
		for _, warehouseItem := range warehouseItems {
			storeQuantityMap[warehouseItem.MaterialUuid] += warehouseItem.Stock
		}
	}
	// 总店物品默认仓库库存（可采购数量）
	if hqUuid != 0 && purchaseOrder.IsHeadquarterPurchase() {
		hqDb := s.dbm.GetDB(hqUuid)
		// 收集采购单中的物品编码和UUID映射
		itemCodes := make([]string, 0, len(purchaseOrder.Items))
		itemCodeToMaterialUuid := make(map[string]uint64)
		materialUuids := make([]uint64, 0, len(purchaseOrder.Items))
		for _, item := range purchaseOrder.Items {
			materialUuids = append(materialUuids, item.MaterialUuid)
			if item.MaterialCode != "" {
				itemCodes = append(itemCodes, item.MaterialCode)
				itemCodeToMaterialUuid[item.MaterialCode] = item.MaterialUuid
			}
		}
		// 从 ERPNext 查询每个物品的默认仓库（按总部公司简称）
		hqCompanyAbbr := companySetting.ErpnextHeadquarterAbbr
		erpSrv := erp.NewIErpSrv(s.dbm)
		itemWarehouses, err := erpSrv.GetItemDefaultWarehouses(ctx, itemCodes, hqCompanyAbbr, false)
		if err == nil && len(itemWarehouses) > 0 {
			// 收集所有不同的仓库 ErpCode
			warehouseErpCodes := make(map[string]bool)
			for _, whErpCode := range itemWarehouses {
				if whErpCode != "" {
					warehouseErpCodes[whErpCode] = true
				}
			}
			// 批量查询各仓库 ErpCode 对应的本地仓库 UUID
			warehouseRepo := repository.NewWarehouseRepo(hqDb)
			codes := make([]string, 0, len(warehouseErpCodes))
			for code := range warehouseErpCodes {
				codes = append(codes, code)
			}
			whMap, whErr := warehouseRepo.GetByErpCodes(codes)
			if whErr != nil {
				logger.Logger.Warn("批量查询仓库ERP编码失败",
					zap.Uint64("company_uuid", ctx.GetCompanyUuid()),
					zap.Error(whErr))
			}
			erpCodeToUuid := make(map[string]uint64, len(whMap))
			for code, wh := range whMap {
				erpCodeToUuid[code] = wh.Uuid
			}
			// 构建 materialUuid -> warehouseUuid 映射
			materialWarehouseMap := make(map[uint64]uint64)
			for itemCode, whErpCode := range itemWarehouses {
				if materialUuid, ok := itemCodeToMaterialUuid[itemCode]; ok {
					if warehouseUuid, ok := erpCodeToUuid[whErpCode]; ok {
						materialWarehouseMap[materialUuid] = warehouseUuid
					}
				}
			}
			// 查询各物品在总部各仓库的库存
			hqWarehouseItemRepo := repository.NewWarehouseItemRepo(hqDb)
			stockByWarehouse, _ := hqWarehouseItemRepo.GetMaterialStockByWarehouse(materialUuids)
			// 按物品的默认仓库匹配库存
			for _, stockResult := range stockByWarehouse {
				if defaultWh, ok := materialWarehouseMap[stockResult.MaterialUuid]; ok && stockResult.WarehouseUuid == defaultWh {
					avaliableQuantityMap[stockResult.MaterialUuid] = stockResult.Stock
				}
			}
		} else if err != nil {
			logger.Logger.Warn("获取物品默认仓库失败，可采购数量将为0",
				zap.Uint64("company_uuid", companySetting.CompanyUuid),
				zap.Error(err),
			)
		}
	} else if hqUuid != 0 && purchaseOrder.WarehouseErpCode != "" {
		// 兼容旧数据：如果采购单仍有指定仓库，按原逻辑查询
		hqWarehouseItemRepo := repository.NewWarehouseItemRepo(s.dbm.GetDB(hqUuid))
		warehouseItems, _ := hqWarehouseItemRepo.GetByWarehouseErpCode(purchaseOrder.WarehouseErpCode)
		for _, warehouseItem := range warehouseItems {
			avaliableQuantityMap[warehouseItem.MaterialUuid] = warehouseItem.Stock
		}
	}

	// 转换仓库名称
	detailResp.WarehouseName = *language.JsonToLocaleResponse(purchaseOrder.WarehouseName)

	// 是否可重新提交
	detailResp.CanRecommit = purchaseOrder.Status == constant.PurchaseOrderStatusRejected && purchaseOrder.ApplicantUuid == ctx.GetStaffUuid()

	// 是否更新限购方案
	detailResp.IsUpdateQuotaScheme = false

	// 初始化数组字段
	detailResp.Items = make([]resp.PurchaseOrderItemInfo, 0, len(purchaseOrder.Items))

	// 计算收货进度：新采购单调用ERP获取，旧采购单使用本地计算
	if purchaseOrder.ErpSaleOrderNo != "" && purchaseOrder.ErpOrderNo != "" {
		// 新采购单：调用ERP获取实时per_received
		erpSrv := erp.NewIErpSrv(s.dbm)
		erpResp, err := erpSrv.GetPurchaseOrder(ctx, &buying.GetPurchaseOrderReq{
			PurchaseOrderName: purchaseOrder.ErpOrderNo,
		})
		if err != nil {
			logger.Logger.Warn("调用ERP获取采购单详情失败，使用本地计算",
				zap.Uint64("uuid", purchaseOrder.Uuid),
				zap.String("erp_order_no", purchaseOrder.ErpOrderNo),
				zap.Error(err),
			)
			detailResp.ReceiptProgress = fmt.Sprintf("%.2f%%", purchaseOrder.GetReceiptProgress())
		} else if erpResp.PurchaseOrder != nil {
			detailResp.ReceiptProgress = fmt.Sprintf("%.2f%%", erpResp.PurchaseOrder.PerReceived)
		} else {
			detailResp.ReceiptProgress = fmt.Sprintf("%.2f%%", purchaseOrder.GetReceiptProgress())
		}
	} else {
		// 旧采购单：使用本地计算
		detailResp.ReceiptProgress = fmt.Sprintf("%.2f%%", purchaseOrder.GetReceiptProgress())
	}

	// 品牌采购：批量查询限购配置（避免 N+1 查询问题）
	quotaLimitMap := s.helper.getQuotaLimitMap(ctx, s.dbm, purchaseOrder)

	// 品牌采购：获取禁止采购的物品列表
	disallowedMaterials := s.helper.getDisallowedPurchaseMaterials(ctx, s.dbm, purchaseOrder)
	disallowedSet := make(map[string]bool)
	for _, code := range disallowedMaterials {
		disallowedSet[code] = true
	}

	// 品牌采购：查询上次完成的品牌采购基准单位数量和采购单位名称
	lastPurchaseInfoMap := make(map[uint64]repository.LastBrandPurchaseInfo)
	if purchaseOrder.IsHeadquarterPurchase() {
		materialUuids := make([]uint64, 0, len(purchaseOrder.Items))
		for _, item := range purchaseOrder.Items {
			materialUuids = append(materialUuids, item.MaterialUuid)
		}
		purchaseOrderItemRepo := repository.NewPurchaseOrderItemRepo(db)
		lastPurchaseInfoMap, _ = purchaseOrderItemRepo.GetLastCompletedBrandPurchaseInfo(materialUuids)
	}

	lang := ctx.GetLanguage()
	// 转换明细数据
	for _, item := range purchaseOrder.Items {
		itemInfo := resp.PurchaseOrderItemInfo{}
		copier.Copy(&itemInfo, &item)
		itemInfo.LocaleName = *language.JsonToLocaleResponse(item.MaterialName)
		itemInfo.LocaleUnitName = *language.JsonToLocaleResponse(item.UnitName)
		itemInfo.LocaleBaseUnitName = func() dto.LocaleResponse {
			if item.Material == nil {
				return dto.LocaleResponse{}
			}
			unitName := item.Material.GetBaseUnit().Name
			return *language.JsonToLocaleResponse(unitName)
		}()
		itemInfo.InternalCode = func(item model.PurchaseOrderItem) string {
			if item.Material == nil {
				return ""
			}
			return item.Material.InternalCode
		}(item)
		itemInfo.Specification = func(item model.PurchaseOrderItem) string {
			if item.Material == nil {
				return ""
			}
			return item.Material.Specification
		}(item)
		itemInfo.BarcodeValue = func(item model.PurchaseOrderItem) string {
			if item.Material == nil {
				return ""
			}
			return item.Material.BarcodeValue
		}(item)
		// 采购单位列表
		itemInfo.UnitList = func(item model.PurchaseOrderItem) []resp.PurchaseOrderItemMaterialUnit {
			unitList := []resp.PurchaseOrderItemMaterialUnit{}
			if item.Material == nil {
				return unitList
			}
			for _, unit := range item.Material.NotBaseUnitList {
				unitList = append(unitList, resp.PurchaseOrderItemMaterialUnit{
					Uuid:           unit.Uuid,
					ConversionRate: unit.ConversionRate,
					LocaleName: func() dto.LocaleResponse {
						if unit.Unit == nil {
							return dto.LocaleResponse{}
						}
						if unit.Unit.MultiLanguageName == (model.MultiLanguageName{}) {
							return dto.LocaleResponse{}
						}
						return unit.Unit.MultiLanguageName.GetNames()
					}(),
				})
			}
			return unitList
		}(item)
		// 已经选中的采购单位列表
		itemInfo.Units = func(item model.PurchaseOrderItem) []resp.PurchaseOrderItemUnit {
			unitList := []resp.PurchaseOrderItemUnit{}
			if len(item.Units) == 0 && item.BaseUnitUuid != 0 {
				unitList = append(unitList, resp.PurchaseOrderItemUnit{
					Num:            item.Num,
					PurchaseNum:    item.Num,
					ArrivalNum:     item.ArrivalNum,
					UnitUuid:       item.UnitUuid,
					ConversionRate: item.UnitConversionRate,
					LocaleName:     *language.JsonToLocaleResponse(item.UnitName),
				})
			} else {
				for _, unit := range item.Units {
					unitList = append(unitList, resp.PurchaseOrderItemUnit{
						Num:            unit.Num,
						PurchaseNum:    unit.Num,
						ArrivalNum:     unit.ArrivalNum,
						UnitUuid:       unit.UnitUuid,
						ConversionRate: unit.UnitConversionRate,
						LocaleName: func() dto.LocaleResponse {
							return *language.JsonToLocaleResponse(unit.UnitName)
						}(),
					})
				}
			}

			return unitList
		}(item)
		itemInfo.AvailableQuantity = decimal.NewFromFloat(avaliableQuantityMap[item.MaterialUuid]).Round(3).InexactFloat64()
		itemInfo.StoreQuantity = decimal.NewFromFloat(storeQuantityMap[item.MaterialUuid]).Round(3).InexactFloat64()
		// 上次采购数量（基准单位，后面转换为默认销售单位）和采购单位名称
		lastPurchaseInfo := lastPurchaseInfoMap[item.MaterialUuid]
		itemInfo.LastPurchaseQuantity = decimal.NewFromFloat(lastPurchaseInfo.BaseQty).Round(3).InexactFloat64()
		itemInfo.LastPurchaseLocaleUnitName = *language.JsonToLocaleResponse(lastPurchaseInfo.UnitName)
		// 申请时门店库存快照（已按默认销售单位存储）
		itemInfo.StoreSnapshotQuantity = decimal.NewFromFloat(item.StoreSnapshotQuantity).Round(3).InexactFloat64()
		// 物料状态：0-正常 1-禁用 2-删除
		if item.Material == nil || item.Material.IsDelete() {
			itemInfo.MaterialStatus = 2
		} else if !item.Material.Status {
			itemInfo.MaterialStatus = 1
		}

		if item.Material != nil {
			// 销售单位UUID
			itemInfo.DefaultSalesUnitUuid = item.Material.DefaultSalesUnitUuid
			for _, unit := range item.Material.NotBaseUnitList {
				if unit.Uuid == item.Material.DefaultSalesUnitUuid {
					// 销售单位名称
					itemInfo.DefaultSalesUnitLocaleName = *language.JsonToLocaleResponse(unit.Name)
					// 转成销售单位数量
					if unit.ConversionRate != 0 {
						itemInfo.AvailableQuantity = decimal.NewFromFloat(avaliableQuantityMap[item.MaterialUuid]).Div(decimal.NewFromFloat(unit.ConversionRate)).Round(3).InexactFloat64()
						itemInfo.StoreQuantity = decimal.NewFromFloat(storeQuantityMap[item.MaterialUuid]).Div(decimal.NewFromFloat(unit.ConversionRate)).Round(3).InexactFloat64()
						itemInfo.LastPurchaseQuantity = decimal.NewFromFloat(lastPurchaseInfo.BaseQty).Div(decimal.NewFromFloat(unit.ConversionRate)).Round(3).InexactFloat64()
					}
				}
			}
			// 限购配置和单位限制检查
			quotaConfig := quotaLimitMap[item.MaterialCode]
			hasQuotaLimit := quotaConfig.QuotaLimit > 0
			isDisallowed := disallowedSet[item.MaterialCode]

			// 设置是否允许采购
			isAllowPurchase := "yes"
			if isDisallowed {
				isAllowPurchase = "no"
			}

			// 获取物品名称（用于错误提示）
			materialName := language.JsonToLocaleResponse(item.MaterialName).GetLocale(lang)

			// 获取限购配置单位（单位限制逻辑不依赖限购方案，始终检查）
			quotaUnit := item.Material.GetUnitByUuidForQuotaConfig()

			// 构建 QuotaConfig
			if quotaUnit != nil && quotaUnit.Unit != nil {
				itemInfo.QuotaConfig = resp.PurchaseOrderItemQuotaConfig{
					QuotaLimit:          quotaConfig.QuotaLimit,
					MinQuotaLimit:       quotaConfig.MinQuotaLimit,
					QuotaUnitUuid:       quotaUnit.Uuid,
					QuotaUnitName:       quotaUnit.Unit.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
					QuotaUnitLocaleName: quotaUnit.Unit.MultiLanguageName.GetNames(),
					IsAllowPurchase:     isAllowPurchase,
				}

				// 检查是否需要更新（草稿或待门店审核状态）
				if purchaseOrder.IsStorePendingOrDraft() {
					// 禁止采购优先级最高
					if isDisallowed {
						detailResp.IsUpdateQuotaScheme = true
						itemInfo.QuotaConfig.ErrorMessage = fmt.Sprintf(i18n.Translate(lang, "物品 %s 禁止采购，请移除后重试"), materialName)
					} else {
						// 单位限制检查（不依赖限购方案，始终检查）
						for _, unit := range item.Units {
							if unit.UnitUuid != itemInfo.QuotaConfig.QuotaUnitUuid {
								detailResp.IsUpdateQuotaScheme = true
								itemInfo.QuotaConfig.ErrorMessage = fmt.Sprintf(
									i18n.Translate(lang, "物品%s的单位限制已变更，当前使用的单位（%s）不在允许范围内，请更新。"),
									materialName, language.JsonToLocaleResponse(unit.UnitName).GetLocale(lang),
								)
								break
							}
							// 最大限购数量检查（仅在有限购方案时检查）
							if hasQuotaLimit && itemInfo.QuotaConfig.QuotaLimit > 0 && unit.Num > itemInfo.QuotaConfig.QuotaLimit {
								detailResp.IsUpdateQuotaScheme = true
								itemInfo.QuotaConfig.ErrorMessage = fmt.Sprintf(
									i18n.Translate(lang, "物品%s申请总数（%.2f）已超过限购数（%.2f），请调整数量后提交。"),
									materialName, item.Num, itemInfo.QuotaConfig.QuotaLimit,
								)
								break
							}
							// 最小采购数量检查（仅在有限购方案时检查）
							if hasQuotaLimit && itemInfo.QuotaConfig.MinQuotaLimit > 0 && unit.Num < itemInfo.QuotaConfig.MinQuotaLimit {
								detailResp.IsUpdateQuotaScheme = true
								itemInfo.QuotaConfig.ErrorMessage = fmt.Sprintf(
									i18n.Translate(lang, "物品%s申请总数（%.2f）不能小于最小采购数（%.2f），请调整数量后提交。"),
									materialName, item.Num, itemInfo.QuotaConfig.MinQuotaLimit,
								)
								break
							}
						}
					}
				}
			} else {
				// 未找到限购配置单位时，只设置基本信息
				logger.Logger.Warn("未找到目标单位", zap.Uint64("material_uuid", item.MaterialUuid), zap.Uint64("default_sales_unit_uuid", item.Material.DefaultSalesUnitUuid))
				errorMessage := ""
				if isDisallowed && purchaseOrder.IsStorePendingOrDraft() {
					detailResp.IsUpdateQuotaScheme = true
					errorMessage = fmt.Sprintf(i18n.Translate(lang, "物品 %s 禁止采购，请移除后重试"), materialName)
				}
				itemInfo.QuotaConfig = resp.PurchaseOrderItemQuotaConfig{
					IsAllowPurchase: isAllowPurchase,
					ErrorMessage:    errorMessage,
				}
			}
		}

		detailResp.Items = append(detailResp.Items, itemInfo)
	}

	remarks := make([]resp.PurchaseOrderRemark, 0)
	var hqPurchaseOrderUuid, subPurchaseOrderUuid uint64
	if companySetting.IsHeadquarter() {
		subPurchaseOrderUuid = purchaseOrder.SubUuid
		hqPurchaseOrderUuid = purchaseOrder.Uuid
	} else if companySetting.IsSubShop() {
		subPurchaseOrderUuid = purchaseOrder.Uuid
		// 根据sub_uuid查询总店采购单UUID
		hqPurchaseOrder, _ := repository.NewPurchaseOrderRepo(s.dbm.GetDB(hqUuid)).GetBySubUuidWithoutDeleted(subPurchaseOrderUuid)
		if hqPurchaseOrder != nil {
			hqPurchaseOrderUuid = hqPurchaseOrder.Uuid
		}
	}
	if subPurchaseOrderUuid != 0 {
		subLogRepo := repository.NewPurchaseOrderLogRepo(s.dbm.GetDB(subUuid))
		subLogs, err := subLogRepo.GetList(subLogRepo.WherePurchaseOrderUuid(subPurchaseOrderUuid))
		if err != nil {
			return resp.PurchaseOrderDetailResp{}, errors.WithMessage(errors.New("查询门店操作日志失败"), err.Error())
		}
		for _, log := range subLogs {
			if (log.OldStatus != constant.PurchaseOrderStatusApproved && log.NewStatus == constant.PurchaseOrderStatusApproved) ||
				(log.OldStatus != constant.PurchaseOrderStatusRejected && log.NewStatus == constant.PurchaseOrderStatusRejected) ||
				log.NewStatus == constant.PurchaseOrderStatusRecommitted {
				remarks = append(remarks, resp.PurchaseOrderRemark{
					Source:     "store",
					Status:     log.NewStatus,
					Remark:     translatePurchaseOrderLogRemark(lang, log.Remark),
					CreateTime: log.CreateTime,
				})
			}
		}
	}

	if hqPurchaseOrderUuid != 0 {
		hqLogRepo := repository.NewPurchaseOrderLogRepo(s.dbm.GetDB(hqUuid))
		hqLogs, err := hqLogRepo.GetList(hqLogRepo.WherePurchaseOrderUuid(hqPurchaseOrderUuid))
		if err != nil {
			return resp.PurchaseOrderDetailResp{}, errors.WithMessage(errors.New("查询总店操作日志失败"), err.Error())
		}
		for _, log := range hqLogs {
			if (log.OldStatus != constant.PurchaseOrderStatusApproved && log.NewStatus == constant.PurchaseOrderStatusApproved) ||
				(log.OldStatus != constant.PurchaseOrderStatusRejected && log.NewStatus == constant.PurchaseOrderStatusRejected) ||
				log.NewStatus == constant.PurchaseOrderStatusRecommitted {
				remarks = append(remarks, resp.PurchaseOrderRemark{
					Source:     "headquarters",
					Status:     log.NewStatus,
					Remark:     translatePurchaseOrderLogRemark(lang, log.Remark),
					CreateTime: log.CreateTime,
				})
			}
		}
	}

	sort.Slice(remarks, func(i, j int) bool {
		return remarks[i].CreateTime > remarks[j].CreateTime
	})
	detailResp.Remarks = remarks

	// 统计收货批次（确认收货的记录数，status=1 已收货）
	receiptOrderRepo := repository.NewPurchaseReceiptOrderRepo(db)
	receiptBatchNum, err := receiptOrderRepo.Count(
		receiptOrderRepo.WherePurchaseOrderUuid(purchaseOrder.Uuid),
		receiptOrderRepo.WhereStatusIn([]int{constant.ReceiptOrderStatusReceived}),
	)
	if err != nil {
		logger.Logger.Warn("统计收货批次失败", zap.Error(err), zap.Uint64("purchase_order_uuid", purchaseOrder.Uuid))
		receiptBatchNum = 0
	}
	detailResp.ReceiptBatchNum = int(receiptBatchNum)

	// 初始化为空数组，确保始终返回该字段
	detailResp.ReceiptList = make([]resp.ReceiptListItem, 0)
	// 版本判断：v2.16.0+ 返回收货清单字段（仅子店）
	if ctx.Version(context.GTE, constant.ClientVersionV2160) && !companySetting.IsHeadquarter() {
		// 仅当 WithReceiptList=true 时才查询收货清单数据
		if req.WithReceiptList {
			// 检查是否需要获取收货清单数据（有ErpSaleOrderNo或是品牌采购）
			needReceiptList := purchaseOrder.ErpSaleOrderNo != "" || purchaseOrder.PurchaseType == constant.PurchaseTypeBrand
			if needReceiptList {
				receiptList, err := s.GetReceiptList(ctx, purchaseOrder)
				if err != nil {
					logger.Logger.Warn("获取收货清单失败", zap.Error(err), zap.Uint64("purchase_order_uuid", purchaseOrder.Uuid))
				} else if len(receiptList) > 0 {
					detailResp.ReceiptList = receiptList
				}
			}
			// 旧品牌采购单，如果没有确认收货记录，则返回空数据
			if purchaseOrder.ErpSaleOrderNo == "" && purchaseOrder.PurchaseType == constant.PurchaseTypeBrand && len(detailResp.ReceiptList) == 0 {
				detailResp.ReceiptList = append(detailResp.ReceiptList, resp.ReceiptListItem{
					IsLegacy:    true,
					IsCompleted: false,
				})
			}
		}
	}

	return detailResp, nil
}

// CreatePurchaseOrder 创建采购申请
func (s *purchaseOrderSrv) CreatePurchaseOrder(
	ctx context.Context,
	req req.PurchaseOrderCreateReq,
) (resp.PurchaseOrderCreateResp, error) {
	// // 加锁
	s.lock.LockUuidString(req.SupplierErpCode + strconv.FormatInt(req.OrderTime, 10))
	defer s.lock.UnlockUuidString(req.SupplierErpCode + strconv.FormatInt(req.OrderTime, 10))

	// 验证请求
	if err := req.Validate(); err != nil {
		return resp.PurchaseOrderCreateResp{}, err
	}

	// 版本检查
	if ctx.Version(context.GTE, "2.6.0") {
		if req.SupplierErpCode == "" {
			return resp.PurchaseOrderCreateResp{}, errors.New("供应商编码不能为空")
		}
	}

	db := ctx.GetDB()
	var result resp.PurchaseOrderCreateResp

	// 获取店铺编码
	var companyStoreCode string
	storeSetting, err := s.settingSrv.GetStoreSetting(ctx)
	if err == nil {
		companyStoreCode = storeSetting.StoreCode
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		purchaseOrderRepo := repository.NewPurchaseOrderRepo(tx)
		purchaseOrderItemRepo := repository.NewPurchaseOrderItemRepo(tx)

		// 获取默认仓库
		defaultWarehouse, err := repository.NewWarehouseRepo(tx).GetDefaultWarehouse()
		if err != nil {
			return errors.WithMessage(errors.New("获取默认仓库失败"), err.Error())
		}

		// 获取 saas 数据库连接
		saasDB := s.dbm.GetDB(constant.DefaultDB)
		if saasDB == nil {
			return errors.New("saas 数据库连接失败")
		}

		// 获取公司 UUID（使用总部 UUID 或当前公司 UUID）
		companyUuid := ctx.GetCompanySetting().HeadquarterUuid
		if companyUuid == 0 {
			companyUuid = ctx.GetCompanyUuid()
		}

		// 确定前缀和编号类型
		var prefix, numberType string
		if req.PurchaseType == 2 {
			// 品牌采购（内部）
			prefix = "TPHY"
			numberType = constant.NumberTypeBrandPurchase
		} else {
			// 采购申请（外部）
			prefix = "PR"
			numberType = constant.NumberTypePurchaseReq
		}

		// 生成订单编号
		orderNo, err := s.helper.generateOrderNo(
			saasDB,
			companyUuid,
			prefix,
			numberType,
			ctx.GetCompanySetting().Timezone,
		)
		if err != nil {
			return errors.WithMessage(err, "生成订单编号失败")
		}

		// 获取仓库名称
		warehouseName := ""
		if req.WarehouseErpCode != "" {
			warehouse, err := repository.NewWarehouseRepo(tx).GetByErpCode(req.WarehouseErpCode)
			if err == nil {
				warehouseName = warehouse.Name
			}
		}

		// 设置期望到货时间，如果为空则默认为2035-12-31
		expectArrivalTime := req.ExpectedDeliveryTime
		if expectArrivalTime == 0 {
			expectArrivalTime = 2082672000 // 2035-12-31的时间戳
		}

		// 创建采购申请
		purchaseOrder := &model.PurchaseOrder{
			OrderNo:           orderNo,
			SupplierName:      req.SupplierName,
			SupplierErpCode:   utils.IfString(req.SupplierErpCode != "", req.SupplierErpCode, req.SupplierName),
			Status:            constant.PurchaseOrderStatusDraft,
			Num:               float64(len(req.Items)),
			OrderTime:         time.Now().Unix(), // 单据日期，采购单提交的时间（时间戳）
			ExpectArrivalTime: expectArrivalTime,
			ApplicantUuid:     ctx.GetStaffUuid(),
			ApplicantName:     ctx.GetStaff().RealName,
			PurchaseType:      utils.IfInt(req.PurchaseType == 2, 2, 1),
			WarehouseErpCode:  req.WarehouseErpCode,
			WarehouseName:     warehouseName,
			CompanyStoreCode:  companyStoreCode,
		}

		// 设置默认仓库信息
		if defaultWarehouse != nil {
			purchaseOrder.DefaultWarehouseErpCode = defaultWarehouse.ErpCode
			purchaseOrder.DefaultWarehouseName = defaultWarehouse.Name
		}

		err = purchaseOrderRepo.Create(purchaseOrder)
		if err != nil {
			return errors.WithMessage(errors.New("创建采购申请失败"), err.Error())
		}

		// 构建采购申请明细
		itemReqs := make([]PurchaseOrderItemReq, 0, len(req.Items))
		for _, item := range req.Items {
			itemReqs = append(itemReqs, PurchaseOrderItemReq{
				MaterialUuid: item.MaterialUuid,
				UnitList:     item.UnitList,
			})
		}
		items, _, err := s.validator.buildPurchaseOrderItems(tx, purchaseOrder.Uuid, itemReqs)
		if err != nil {
			return err
		}

		// 创建采购申请明细
		err = purchaseOrderItemRepo.CreateBatch(items)
		if err != nil {
			return errors.WithMessage(errors.New("创建采购申请明细失败"), err.Error())
		}

		// 记录操作日志
		err = s.helper.createPurchaseOrderLog(
			tx,
			purchaseOrder.Uuid,
			ctx,
			"create",
			"创建采购申请",
			0,
			constant.PurchaseOrderStatusPending,
			"",
			// 记录操作日志内容
			func(order *model.PurchaseOrder, items []model.PurchaseOrderItem) string {
				order.Items = items
				return utils.ToJson(order)
			}(purchaseOrder, items),
		)
		if err != nil {
			return err
		}

		result.Uuid = purchaseOrder.Uuid
		result.OrderNo = purchaseOrder.OrderNo

		return nil
	})

	if err != nil {
		return resp.PurchaseOrderCreateResp{}, err
	}

	return result, nil
}

// UpdatePurchaseOrder 更新采购申请
func (s *purchaseOrderSrv) UpdatePurchaseOrder(
	ctx context.Context,
	req req.PurchaseOrderUpdateReq,
) error {
	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	if err := req.Validate(); err != nil {
		return err
	}

	db := ctx.GetDB()

	return db.Transaction(func(tx *gorm.DB) error {
		purchaseOrderRepo := repository.NewPurchaseOrderRepo(tx)
		purchaseOrderItemRepo := repository.NewPurchaseOrderItemRepo(tx)

		// 查询现有采购申请
		purchaseOrder, err := purchaseOrderRepo.GetByUuid(req.Uuid)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("采购申请不存在")
			}
			return errors.WithMessage(errors.New("查询采购申请失败"), err.Error())
		}

		// 检查是否可编辑，草稿状态、待审核、已驳回状态可以编辑
		if !purchaseOrder.IsEditable() && purchaseOrder.Status != constant.PurchaseOrderStatusPending && purchaseOrder.Status != constant.PurchaseOrderStatusRejected {
			return errors.New("当前状态不允许编辑")
		}

		// 检查是否需要更新（优化：避免不必要的数据库操作）
		needUpdate, err := s.helper.checkPurchaseOrderNeedUpdate(purchaseOrder, &req, purchaseOrderItemRepo)
		if err != nil {
			return err
		}

		// 如果没有任何变动，直接返回
		if !needUpdate {
			return nil
		}

		// 更新采购申请基本信息
		if purchaseOrder.Status != constant.PurchaseOrderStatusPending && purchaseOrder.Status != constant.PurchaseOrderStatusHeadquarterPending {
			// 版本检查
			if ctx.Version(context.GTE, "2.6.0") {
				if req.SupplierErpCode == "" {
					return errors.New("供应商编码不能为空")
				}
			}
			// 获取仓库名称
			warehouseName := ""
			if req.WarehouseErpCode != "" {
				warehouse, err := repository.NewWarehouseRepo(tx).GetByErpCode(req.WarehouseErpCode)
				if err == nil {
					warehouseName = warehouse.Name
				}
			}
			// 设置期望到货时间，如果为空则默认为2035-12-31
			expectArrivalTime := req.ExpectedDeliveryTime
			if expectArrivalTime == 0 {
				expectArrivalTime = 2082672000 // 2035-12-31的时间戳
			}
			purchaseOrder.SupplierName = req.SupplierName
			purchaseOrder.SupplierErpCode = req.SupplierErpCode
			purchaseOrder.ExpectArrivalTime = expectArrivalTime
			purchaseOrder.WarehouseErpCode = req.WarehouseErpCode
			purchaseOrder.WarehouseName = warehouseName
		}
		purchaseOrder.Num = float64(len(req.Items))
		err = purchaseOrderRepo.Update(purchaseOrder)
		if err != nil {
			return errors.WithMessage(errors.New("更新采购申请失败"), err.Error())
		}

		// 先删除所有现有明细项
		err = purchaseOrderItemRepo.DeleteByPurchaseOrderUuid(req.Uuid)
		if err != nil {
			return errors.WithMessage(errors.New("删除采购申请明细失败"), err.Error())
		}

		// 构建并批量创建采购申请明细
		itemReqs := make([]PurchaseOrderItemReq, 0, len(req.Items))
		for _, item := range req.Items {
			itemReqs = append(itemReqs, PurchaseOrderItemReq{
				MaterialUuid: item.MaterialUuid,
				UnitList:     item.UnitList,
			})
		}
		items, _, err := s.validator.buildPurchaseOrderItems(tx, purchaseOrder.Uuid, itemReqs)
		if err != nil {
			return err
		}

		// 创建采购申请明细
		err = purchaseOrderItemRepo.CreateBatch(items)
		if err != nil {
			return errors.WithMessage(errors.New("创建采购申请明细失败"), err.Error())
		}

		// 如果当前操作店铺不是采购单归属店铺，需要同步更新归属店铺的数据
		if purchaseOrder.IsHeadquarterPurchase() && purchaseOrder.CompanyUuid != 0 && purchaseOrder.CompanyUuid != ctx.GetCompanyUuid() {
			ctxCopy := ctx.Copy()
			ctxCopy.SetDB(tx)
			err = s.syncItemsToCompanyShop(ctxCopy, purchaseOrder, req.Items)
			if err != nil {
				return errors.WithMessage(errors.New("同步归属店铺采购明细失败"), err.Error())
			}
		}

		// 记录操作日志
		err = s.helper.createPurchaseOrderLog(
			tx,
			req.Uuid,
			ctx,
			"update",
			"更新采购申请",
			purchaseOrder.Status,
			purchaseOrder.Status,
			"",
			// 记录操作日志内容
			func() string {
				// 查询现有采购申请
				purchaseOrder, err := purchaseOrderRepo.GetByUuid(req.Uuid, purchaseOrderRepo.WithSimpleItems())
				if err != nil {
					if err == gorm.ErrRecordNotFound {
						return ""
					}
					return ""
				}
				return utils.ToJson(purchaseOrder)
			}(),
		)
		if err != nil {
			return err
		}

		return nil
	})
}

// UpdatePurchaseOrderItemUnit 更新采购订单物品单位
//
// 功能：根据最新规则更新订单中所有物品的单位
// 规则：当物品设置了销售单位时用销售单位，没有设置则使用基准单位
func (s *purchaseOrderSrv) UpdatePurchaseOrderItemUnit(
	ctx context.Context,
	req req.PurchaseOrderDetailReq,
) (resp.PurchaseOrderDetailResp, error) {
	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	db := ctx.GetDB()

	var result resp.PurchaseOrderDetailResp

	err := db.Transaction(func(tx *gorm.DB) error {
		purchaseOrderRepo := repository.NewPurchaseOrderRepo(tx)
		purchaseOrderItemUnitRepo := repository.NewPurchaseOrderItemUnitRepo(tx)
		materialRepo := repository.NewMaterialRepo(tx)

		// 1. 查询采购申请及其明细
		purchaseOrder, err := purchaseOrderRepo.GetByUuid(
			req.Uuid,
			purchaseOrderRepo.WithItems(),
		)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("采购申请不存在")
			}
			return errors.WithMessage(errors.New("查询采购申请失败"), err.Error())
		}

		// 2. 检查是否可编辑（待提交或待审核状态）
		if !purchaseOrder.IsStorePendingOrDraft() {
			return errors.New("当前状态不允许修改")
		}

		// 品牌采购：批量查询限购配置（避免 N+1 查询问题）
		quotaLimitMap := s.helper.getQuotaLimitMap(ctx, s.dbm, purchaseOrder)

		// 3. 遍历所有物品，更新单位
		for _, item := range purchaseOrder.Items {
			// 查询物品完整信息（包括单位列表）
			material, err := materialRepo.GetMaterialByUuid(
				item.MaterialUuid,
				materialRepo.WithUnit(),
				materialRepo.WithNotBaseUnitList(),
			)
			if err != nil {
				logger.Logger.Warn("查询物品失败", zap.Uint64("material_uuid", item.MaterialUuid), zap.Error(err))
				continue // 跳过无法查询的物品
			}

			// 品牌采购：检查是否超过限购数量
			quotaConfig := quotaLimitMap[item.MaterialCode]
			if quotaConfig.QuotaLimit > 0 && item.Num > quotaConfig.QuotaLimit {
				return errors.New("当前物品数量超过限购数量，请检查物品单位是否正确")
			}

			// 4. 确定应该使用的单位（销售单位优先，否则使用基准单位）
			targetUnit := material.GetUnitByUuidForQuotaConfig()
			if targetUnit == nil || targetUnit.Unit == nil {
				logger.Logger.Warn("未找到目标单位", zap.Uint64("material_uuid", item.MaterialUuid), zap.Uint64("default_sales_unit_uuid", material.DefaultSalesUnitUuid))
				continue
			}

			// 更新单位记录
			for index, unit := range item.Units {
				if index > 0 {
					// 删除旧单位记录
					err = purchaseOrderItemUnitRepo.DeleteByItemUuidAndUnitUuid(item.Uuid, unit.Uuid)
					if err != nil {
						logger.Logger.Error("删除旧单位记录失败", zap.Uint64("item_uuid", unit.Uuid), zap.Error(err))
						return errors.WithMessage(errors.New("删除旧单位记录失败"), err.Error())
					}
				} else if unit.UnitUuid != targetUnit.Uuid {
					// 如果单位不一致，则更新单位记录
					unit.UnitUuid = targetUnit.Uuid
					unit.UnitName = utils.ToJson(targetUnit.Unit.MultiLanguageName.GetNames())
					unit.UnitConversionRate = targetUnit.ConversionRate
					err = purchaseOrderItemUnitRepo.Update(unit)
					if err != nil {
						logger.Logger.Error("更新单位记录失败", zap.Uint64("item_uuid", unit.Uuid), zap.Error(err))
						return errors.WithMessage(errors.New("更新单位记录失败"), err.Error())
					}
				}
			}
		}

		// 记录日志 更新单位记录
		err = s.helper.createPurchaseOrderLog(
			tx,
			req.Uuid,
			ctx,
			"update_item_unit",
			"更新采购申请明细单位",
			purchaseOrder.Status,
			purchaseOrder.Status,
			"",
			func() string {
				purchaseOrder, err := purchaseOrderRepo.GetByUuid(
					req.Uuid,
					purchaseOrderRepo.WithSimpleItems(),
				)
				if err != nil {
					return ""
				}
				return utils.ToJson(purchaseOrder)
			}(),
		)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return resp.PurchaseOrderDetailResp{}, err
	}

	// 重新查询采购申请详情
	result, err = s.GetPurchaseOrderDetail(ctx, req)
	if err != nil {
		return resp.PurchaseOrderDetailResp{}, err
	}

	return result, nil
}

// DeletePurchaseOrder 删除采购申请
func (s *purchaseOrderSrv) DeletePurchaseOrder(
	ctx context.Context,
	req req.PurchaseOrderDeleteReq,
) error {
	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	db := s.dbm.GetDB(ctx.GetDbId())
	purchaseOrderRepo := repository.NewPurchaseOrderRepo(db)

	// 查询采购申请
	purchaseOrder, err := purchaseOrderRepo.GetByUuid(req.Uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("采购申请不存在")
		}
		return errors.WithMessage(errors.New("查询采购申请失败"), err.Error())
	}

	// 检查是否可删除（只有待提交和已驳回状态可以删除）
	if !purchaseOrder.IsEditable() {
		return errors.New("当前状态不允许删除")
	}

	// 删除采购申请（软删除）
	err = purchaseOrderRepo.Delete(req.Uuid)
	if err != nil {
		return errors.WithMessage(errors.New("删除采购申请失败"), err.Error())
	}

	return nil
}

// SubmitPurchaseOrder 提交采购申请
func (s *purchaseOrderSrv) SubmitPurchaseOrder(
	ctx context.Context,
	req req.PurchaseOrderSubmitReq,
) error {
	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	db := ctx.GetDB()

	companySetting := ctx.GetCompanySetting()

	return db.Transaction(func(tx *gorm.DB) error {
		purchaseOrderRepo := repository.NewPurchaseOrderRepo(tx)
		purchaseOrderItemRepo := repository.NewPurchaseOrderItemRepo(tx)

		// 查询采购申请
		purchaseOrder, err := purchaseOrderRepo.GetByUuid(req.Uuid, purchaseOrderRepo.WithItems())
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("采购申请不存在")
			}
			return errors.WithMessage(errors.New("查询采购申请失败"), err.Error())
		}

		// 检查供应商状态
		if err := s.validator.validateSupplierStatus(tx, purchaseOrder.SupplierErpCode); err != nil {
			return err
		}

		// 检查物品状态
		if !req.IsConfirm {
			disabledMaterials, err := s.validator.validateMaterialStatus(ctx, tx, purchaseOrder.Items, true)
			if err != nil {
				return err
			}

			// 如果有禁用的物品，返回提示信息
			if len(disabledMaterials) > 0 {
				materialNames := s.validator.joinMaterialNames(disabledMaterials)
				return errors.NewWithCodeAndData(
					constant.CodeMaterialDisabled,
					disabledMaterials,
					fmt.Sprintf(
						i18n.Translate(ctx.GetLanguage(), "物品 %s 的状态已关闭。\n\n提交后将移除该物品，是否继续提交？"),
						materialNames,
					),
				)
			}
		}

		// 删除物品为0的数据
		err = purchaseOrderItemRepo.DeleteByPurchaseOrderUuidAndNumIsZero(req.Uuid)
		if err != nil {
			return errors.WithMessage(errors.New("删除采购申请明细失败"), err.Error())
		}

		// 如果用户确认提交，删除禁用的物品
		if req.IsConfirm {
			materialRepo := repository.NewMaterialRepo(tx)
			var disabledMaterialUuids []uint64

			for _, item := range purchaseOrder.Items {
				if item.Num <= 0 {
					continue
				}
				material, err := materialRepo.GetMaterialByUuid(item.MaterialUuid)
				if err != nil || !material.Status {
					disabledMaterialUuids = append(disabledMaterialUuids, item.MaterialUuid)
				}
			}

			// 删除禁用的物品记录
			if len(disabledMaterialUuids) > 0 {
				err = purchaseOrderItemRepo.DeleteByPurchaseOrderUuidAndMaterialUuids(req.Uuid, disabledMaterialUuids)
				if err != nil {
					return errors.WithMessage(errors.New("删除禁用物品失败"), err.Error())
				}
			}
		}

		// 重新查询采购申请以获取最新的物品列表
		purchaseOrder, err = purchaseOrderRepo.GetByUuid(req.Uuid, purchaseOrderRepo.WithItems())
		if err != nil {
			return errors.WithMessage(errors.New("重新查询采购申请失败"), err.Error())
		}

		// 检查采购申请明细
		if len(purchaseOrder.Items) == 0 {
			return errors.New("申请物品数量不能为0")
		}

		// 过滤掉数量为0的项目后重新计算数量
		validItemCount := 0
		for _, item := range purchaseOrder.Items {
			if item.Num > 0 {
				validItemCount++
			}
		}

		// 🔥 新增：品牌采购限购校验（基于新的限购方案表）
		if purchaseOrder.IsHeadquarterPurchase() {
			// 检查限购方案（包含每日申请次数限制和物品数量限制）
			if err := s.checkPurchaseLimit(ctx, purchaseOrder); err != nil {
				return err
			}
		}

		// 品牌采购：记录提交时的门店库存快照（默认销售单位）
		if purchaseOrder.IsHeadquarterPurchase() {
			storeUuid := companySetting.CompanyUuid
			if storeUuid != 0 {
				// 收集物品 UUID
				materialUuids := make([]uint64, 0, len(purchaseOrder.Items))
				for _, item := range purchaseOrder.Items {
					if item.Num > 0 {
						materialUuids = append(materialUuids, item.MaterialUuid)
					}
				}
				// 查询门店所有仓库的基准单位库存
				storeWarehouseItemRepo := repository.NewWarehouseItemRepo(s.dbm.GetDB(storeUuid))
				warehouseItems, whErr := storeWarehouseItemRepo.GetItemsInActiveWarehouses()
				if whErr != nil {
					logger.Logger.Warn("查询门店仓库库存失败", zap.Uint64("store_uuid", storeUuid), zap.Error(whErr))
				}
				storeStockMap := make(map[uint64]float64)
				for _, wi := range warehouseItems {
					storeStockMap[wi.MaterialUuid] += wi.Stock
				}
				// 查询物品的默认销售单位转换率
				materialRepo := repository.NewMaterialRepo(tx)
				materials, matErr := materialRepo.GetMaterialByUuids(materialUuids, materialRepo.WithNotBaseUnitList())
				if matErr != nil {
					logger.Logger.Warn("查询物品默认销售单位失败", zap.Error(matErr))
				}
				conversionRateMap := make(map[uint64]float64)
				for _, m := range materials {
					if m.DefaultSalesUnitUuid > 0 {
						for _, unit := range m.NotBaseUnitList {
							if unit.Uuid == m.DefaultSalesUnitUuid && unit.ConversionRate != 0 {
								conversionRateMap[m.Uuid] = unit.ConversionRate
							}
						}
					}
				}
				// 设置每个物品的门店库存快照
				for i := range purchaseOrder.Items {
					item := &purchaseOrder.Items[i]
					if item.Num <= 0 {
						continue
					}
					baseStock := storeStockMap[item.MaterialUuid]
					if rate, ok := conversionRateMap[item.MaterialUuid]; ok {
						item.StoreSnapshotQuantity = decimal.NewFromFloat(baseStock).Div(decimal.NewFromFloat(rate)).Round(4).InexactFloat64()
					} else {
						item.StoreSnapshotQuantity = decimal.NewFromFloat(baseStock).Round(4).InexactFloat64()
					}
					if err := purchaseOrderItemRepo.UpdateByUuid(item.Uuid, map[string]any{
						"store_snapshot_quantity": item.StoreSnapshotQuantity,
					}); err != nil {
						return errors.WithMessage(errors.New("记录门店库存快照失败"), err.Error())
					}
				}
			}
		}

		oldStatus := purchaseOrder.Status
		purchaseOrder.Status = constant.PurchaseOrderStatusPending
		purchaseOrder.OrderTime = time.Now().Unix()
		purchaseOrder.ApplicantUuid = ctx.GetStaffUuid()
		purchaseOrder.ApplicantName = ctx.GetStaff().RealName
		purchaseOrder.Num = float64(validItemCount)

		err = purchaseOrderRepo.Update(purchaseOrder)
		if err != nil {
			return errors.WithMessage(errors.New("更新采购申请状态失败"), err.Error())
		}

		// 操作日志记录的状态
		logStatus := purchaseOrder.Status
		if req.IsResubmit {
			logStatus = constant.PurchaseOrderStatusRecommitted

			// 子店采购单驳回理由清空
			repository.NewPurchaseOrderRepo(tx).UpdateByMap(map[string]any{
				"reject_reason":      "",
				"headquarter_status": constant.HeadquarterStatusDraft,
			}, repository.CommonRepo.WhereByUuid(req.Uuid))
			// 总店子店采购单标记删除、清空驳回理由
			repository.NewPurchaseOrderRepo(s.dbm.GetDB(companySetting.HeadquarterUuid)).UpdateByMap(map[string]any{
				"delete_time":   time.Now().Unix(),
				"reject_reason": "",
			}, func(db *gorm.DB) *gorm.DB {
				return db.Where("sub_uuid = ?", req.Uuid)
			})
		}

		// 记录操作日志
		statusText := purchaseOrder.GetStatusText()
		err = s.helper.createPurchaseOrderLog(
			tx,
			req.Uuid,
			ctx,
			"status_update",
			"更新状态为"+statusText,
			oldStatus,
			logStatus,
			"",
			"{}",
		)
		if err != nil {
			return err
		}

		return nil
	})
}

// ApprovePurchaseOrder 审核采购申请
func (s *purchaseOrderSrv) ApprovePurchaseOrder(
	ctx context.Context,
	req req.PurchaseOrderApproveReq,
) error {
	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	db := ctx.GetDB()
	companySetting := ctx.GetCompanySetting()

	// 2.15.0 批注
	if utf8.RuneCountInString(req.Remark) > 100 {
		return errors.New("批注最多100个字符")
	}

	// 2.14.0 驳回原因
	if ctx.Version(context.GTE, constant.ClientVersionV2140) && req.Action == "reject" && utf8.RuneCountInString(req.RejectReason) > 100 {
		return errors.New("驳回原因最多100个字符")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		purchaseOrderRepo := repository.NewPurchaseOrderRepo(tx)

		// 查询采购申请
		purchaseOrder, err := purchaseOrderRepo.GetByUuid(req.Uuid, purchaseOrderRepo.WithItems())
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("采购申请不存在")
			}
			return errors.WithMessage(err, "查询采购申请失败")
		}

		// 检查是否可审核
		if !purchaseOrder.CanApprove() && (companySetting.IsSubShop() || companySetting.IsHeadquarter()) {
			return errors.New("当前状态不允许审核")
		}

		// 审核通过时检查
		if req.Action == "approve" {
			// 检查供应商状态
			if err := s.validator.validateSupplierStatus(tx, purchaseOrder.SupplierErpCode); err != nil {
				return err
			}
			// 检查物品状态
			_, err := s.validator.validateMaterialStatus(ctx, tx, purchaseOrder.Items, false)
			if err != nil {
				return err
			}
			// 🔥 新增：品牌采购限购校验（基于新的限购方案表）
			if purchaseOrder.IsHeadquarterPurchase() && purchaseOrder.IsStorePendingOrDraft() {
				// 检查限购方案（包含每日申请次数限制和物品数量限制）
				if err := s.checkPurchaseLimit(ctx, purchaseOrder); err != nil {
					return err
				}
			}

			// 总部审核品牌采购：校验物品默认仓库配置
			if companySetting.IsHeadquarter() && purchaseOrder.IsHeadquarterPurchase() && !req.IsConfirm {
				noWarehouseItems := s.checkItemDefaultWarehouse(ctx, purchaseOrder)
				if len(noWarehouseItems) > 0 {
					return errors.NewWithCodeAndData(
						constant.CodePurchaseOrderItemNoDefaultWarehouse,
						noWarehouseItems,
						fmt.Sprintf(i18n.Translate(ctx.GetLanguage(), "物品 %s 未配置默认发货仓，请联系管理员处理。是否继续审核？（将由系统指定物品库存仓库进行发货）"), strings.Join(noWarehouseItems, ", ")),
					)
				}
			}
		}

		// 处理审核逻辑
		oldStatus := purchaseOrder.Status
		var newStatus int
		var actionDesc string

		if req.Action == "approve" {
			newStatus = constant.PurchaseOrderStatusApproved
			actionDesc = "审核通过"
			purchaseOrder.PassTime = time.Now().Unix()
		} else if req.Action == "reject" {
			newStatus = constant.PurchaseOrderStatusRejected
			actionDesc = "审核驳回"
			purchaseOrder.RejectTime = time.Now().Unix()
			rejectReason := req.RejectReason
			if rejectReason == "" {
				rejectReason = req.Remark
			}
			purchaseOrder.RejectReason = rejectReason
		} else {
			return errors.New("无效的审核动作")
		}

		// 更新状态
		purchaseOrder.Status = newStatus
		if purchaseOrder.IsHeadquarterPurchase() && (companySetting.IsHeadquarter() || newStatus != constant.PurchaseOrderStatusRejected) {
			purchaseOrder.HeadquarterStatus = newStatus
		}

		// 子店内部采购审核通过时，状态改为待总部审核
		if companySetting.IsSubShop() && purchaseOrder.IsHeadquarterPurchase() && newStatus == constant.PurchaseOrderStatusApproved {
			purchaseOrder.Status = constant.PurchaseOrderStatusHeadquarterPending
			purchaseOrder.HeadquarterStatus = constant.HeadquarterStatusPending
		}

		// 更新状态
		err = purchaseOrderRepo.Update(purchaseOrder)
		if err != nil {
			return errors.WithMessage(errors.New("更新采购申请状态失败"), err.Error())
		}

		var remark string
		if req.RejectReason != "" {
			remark = req.RejectReason
		} else if req.Remark != "" {
			remark = req.Remark
		}
		// 记录操作日志
		err = s.helper.createPurchaseOrderLog(tx, req.Uuid, ctx, req.Action, actionDesc, oldStatus, newStatus, remark, "{}")
		if err != nil {
			return errors.WithMessage(errors.New("记录操作日志失败"), err.Error())
		}

		// 处理总部驳回场景
		if companySetting.IsHeadquarter() && purchaseOrder.IsHeadquarterPurchase() && purchaseOrder.Status == constant.PurchaseOrderStatusRejected {
			return s.handleHeadquarterReject(purchaseOrder)
		}

		// 处理子店审核通过场景
		if companySetting.IsSubShop() && purchaseOrder.IsHeadquarterPurchase() && purchaseOrder.Status == constant.PurchaseOrderStatusHeadquarterPending {
			return s.handleSubShopApproval(ctx, purchaseOrder)
		}

		// 调用ERP接口
		if ctx.GetCompany().IsOpenErp() && purchaseOrder.Status == constant.PurchaseOrderStatusApproved {
			return s.handleErpApproval(ctx, tx, &companySetting, purchaseOrder, false)
		}

		return nil
	})
}

// handleHeadquarterReject 处理总部驳回场景
func (s *purchaseOrderSrv) handleHeadquarterReject(purchaseOrder *model.PurchaseOrder) error {
	// 获取子店数据库
	subDb := s.dbm.GetDB(purchaseOrder.CompanyUuid)
	if subDb == nil {
		return errors.New("获取子店数据库失败")
	}

	subPurchaseOrder, err := repository.NewPurchaseOrderRepo(subDb).GetByUuid(purchaseOrder.SubUuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("采购申请不存在")
		}
		return errors.WithMessage(errors.New("查询采购申请失败"), err.Error())
	}

	subPurchaseOrder.Status = constant.PurchaseOrderStatusRejected
	subPurchaseOrder.HeadquarterStatus = constant.HeadquarterStatusRejected
	subPurchaseOrder.RejectTime = purchaseOrder.RejectTime
	subPurchaseOrder.RejectReason = purchaseOrder.RejectReason

	err = repository.NewPurchaseOrderRepo(subDb).Update(subPurchaseOrder)
	if err != nil {
		return errors.WithMessage(errors.New("更新子店采购申请失败"), err.Error())
	}

	return nil
}

// handleSubShopApproval 处理子店审核通过场景
func (s *purchaseOrderSrv) handleSubShopApproval(
	ctx context.Context,
	purchaseOrder *model.PurchaseOrder,
) error {
	companySetting := ctx.GetCompanySetting()

	// 获取总部数据库
	headquarterDb := s.dbm.GetDB(companySetting.HeadquarterUuid)
	if headquarterDb == nil {
		return errors.New("获取总部数据库失败")
	}

	var hqPurchaseOrderUuid uint64
	var shouldAutoApprove bool

	err := headquarterDb.Transaction(func(hqTx *gorm.DB) error {
		// 整单复制到总部
		subUuid := purchaseOrder.Uuid
		headquarterPurchaseOrder := model.PurchaseOrder{}
		err := copier.Copy(&headquarterPurchaseOrder, purchaseOrder)
		if err != nil {
			return errors.WithMessage(errors.New("复制总部采购订单失败"), err.Error())
		}

		// 判断是否已经有总部采购单
		hqPurchaseOrder, _ := repository.NewPurchaseOrderRepo(hqTx).GetBySubUuidWithoutDeleted(subUuid)
		if hqPurchaseOrder != nil {
			// 删除总部采购单
			err = repository.NewPurchaseOrderRepo(hqTx).ForceDelete(hqPurchaseOrder.Uuid)
			if err != nil {
				return errors.WithMessage(errors.New("删除总部采购申请失败"), err.Error())
			}
		}

		// 重置主键字段
		headquarterPurchaseOrder.BaseModel.ID = 0
		headquarterPurchaseOrder.BaseModel.Uuid = func() uint64 {
			uuid, _ := utils.GetID()
			return uuid
		}()
		headquarterPurchaseOrder.CompanyUuid = ctx.GetCompanyUuid()
		headquarterPurchaseOrder.CompanyName = ctx.GetCompany().Name
		headquarterPurchaseOrder.SubUuid = subUuid
		headquarterPurchaseOrder.Status = constant.PurchaseOrderStatusPending
		headquarterPurchaseOrder.HeadquarterStatus = constant.HeadquarterStatusPending
		if hqPurchaseOrder != nil {
			headquarterPurchaseOrder.BaseModel.ID = hqPurchaseOrder.ID
			headquarterPurchaseOrder.Uuid = hqPurchaseOrder.Uuid
		}
		err = repository.NewPurchaseOrderRepo(hqTx).Create(&headquarterPurchaseOrder)
		if err != nil {
			logger.Logger.Error("创建总部采购申请失败", zap.Error(err))
			return errors.WithMessage(errors.New("创建总部采购申请失败"), err.Error())
		}

		// 创建总部采购申请明细
		hqMaterialRepo := repository.NewMaterialRepo(hqTx)
		var headquarterItems []model.PurchaseOrderItem
		for _, item := range purchaseOrder.Items {
			headquarterItem := model.PurchaseOrderItem{}
			err = copier.Copy(&headquarterItem, item)
			if err != nil {
				logger.Logger.Error("复制总部物品明细失败", zap.Error(err), zap.String("物料编码", item.MaterialCode))
				return errors.WithMessage(errors.New("复制总部物品明细失败"), err.Error())
			}

			// 查询物料
			material, err := hqMaterialRepo.GetMaterialByErpCode(item.MaterialCode)
			if err != nil {
				logger.Logger.Error("总部不存在该物料", zap.String("物料编码", item.MaterialCode), zap.Error(err))
				return errors.New(fmt.Sprintf(
					i18n.Translate(ctx.GetLanguage(), "总部不存在该物料，物料编码：%s"),
					item.MaterialCode,
				))
			}

			// 重置主键字段
			headquarterItem.BaseModel.ID = 0
			headquarterItem.BaseModel.Uuid = 0
			headquarterItem.PurchaseOrderUuid = headquarterPurchaseOrder.Uuid
			headquarterItem.MaterialUuid = material.Uuid
			headquarterItem.Material = nil
			headquarterItem.Units = make([]model.PurchaseOrderItemUnit, 0)
			// 复制单位列表
			if len(item.Units) > 0 {
				for _, unit := range item.Units {
					itemUnit := model.PurchaseOrderItemUnit{}
					err := copier.Copy(&itemUnit, unit)
					if err != nil {
						logger.Logger.Error("复制总部物品单位失败", zap.Error(err))
						return errors.WithMessage(errors.New("复制总部物品单位失败"), err.Error())
					}
					itemUnit.BaseModel.ID = 0
					itemUnit.PurchaseOrderUuid = headquarterPurchaseOrder.Uuid
					itemUnit.BaseUnitUuid = material.UnitUuid
					headquarterItem.Units = append(headquarterItem.Units, itemUnit)
				}
			}
			// 添加到总部采购申请明细列表
			headquarterItems = append(headquarterItems, headquarterItem)
		}

		err = repository.NewPurchaseOrderItemRepo(hqTx).CreateBatch(headquarterItems)
		if err != nil {
			logger.Logger.Error("创建总部采购申请明细失败", zap.Error(err))
			return errors.WithMessage(errors.New("创建总部采购申请明细失败"), err.Error())
		}

		hqPurchaseOrderUuid = headquarterPurchaseOrder.Uuid
		// 在事务内查询总部设置，避免事务外额外 DB 查询
		hqSetting := repository.NewCompanySettingRepo(hqTx).Get()
		shouldAutoApprove = hqSetting.BrandPurchaseAutoApprove == 1
		return nil
	})
	if err != nil {
		return err
	}

	// 总部采购单创建成功后，检查自动审批开关，异步执行总部自动审批
	if shouldAutoApprove && ctx.GetCompany().IsOpenErp() {
		headquarterUuid := companySetting.HeadquarterUuid
		companyUuid := ctx.GetCompanyUuid()
		lang := ctx.GetLanguage()
		utils.Go(func() {
			s.autoApproveHeadquarter(headquarterUuid, companyUuid, hqPurchaseOrderUuid, lang)
		})
	}

	return nil
}

// autoApproveHeadquarter 后台任务：总部自动审批品牌采购单
// 最多重试3次，使用指数退避（2s, 4s, 8s）
func (s *purchaseOrderSrv) autoApproveHeadquarter(
	headquarterUuid uint64,
	companyUuid uint64,
	hqPurchaseOrderUuid uint64,
	lang string,
) {
	const maxRetries = 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// 指数退避: 2s, 4s, 8s
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			time.Sleep(backoff)
		}

		err := s.doAutoApproveHeadquarter(headquarterUuid, companyUuid, hqPurchaseOrderUuid, lang)
		if err == nil {
			logger.Logger.Info("品牌采购自动审批成功",
				zap.Uint64("company_uuid", companyUuid),
				zap.Uint64("headquarter_uuid", headquarterUuid),
				zap.Uint64("hq_purchase_order_uuid", hqPurchaseOrderUuid),
				zap.Int("attempt", attempt+1),
			)
			return
		}

		logger.Logger.Error("品牌采购自动审批失败",
			zap.Uint64("company_uuid", companyUuid),
			zap.Uint64("headquarter_uuid", headquarterUuid),
			zap.Uint64("hq_purchase_order_uuid", hqPurchaseOrderUuid),
			zap.Int("attempt", attempt+1),
			zap.Int("max_retries", maxRetries),
			zap.Error(err),
		)
	}

	logger.Logger.Error("品牌采购自动审批最终失败，已达最大重试次数",
		zap.Uint64("company_uuid", companyUuid),
		zap.Uint64("headquarter_uuid", headquarterUuid),
		zap.Uint64("hq_purchase_order_uuid", hqPurchaseOrderUuid),
		zap.Int("max_retries", maxRetries),
	)
}

// doAutoApproveHeadquarter 执行总部自动审批（单次尝试）
func (s *purchaseOrderSrv) doAutoApproveHeadquarter(
	headquarterUuid uint64,
	companyUuid uint64,
	hqPurchaseOrderUuid uint64,
	lang string,
) error {
	headquarterDb := s.dbm.GetDB(headquarterUuid)
	if headquarterDb == nil {
		return errors.New("获取总部数据库失败")
	}

	return headquarterDb.Transaction(func(hqTx *gorm.DB) error {
		// 加载总部采购单（含明细）
		hqPORepo := repository.NewPurchaseOrderRepo(hqTx)
		hqPO, err := hqPORepo.GetByUuid(hqPurchaseOrderUuid, hqPORepo.WithItems())
		if err != nil {
			return errors.WithMessage(errors.New("查询总部采购申请失败"), err.Error())
		}

		// 仅审批待审核状态的采购单
		if hqPO.Status != constant.PurchaseOrderStatusPending {
			logger.Logger.Info("品牌采购自动审批跳过：采购单状态非待审核",
				zap.Uint64("company_uuid", companyUuid),
				zap.Uint64("hq_purchase_order_uuid", hqPurchaseOrderUuid),
				zap.Int("status", hqPO.Status),
			)
			return nil
		}

		// 更新总部采购单状态为已通过
		hqPO.Status = constant.PurchaseOrderStatusApproved
		hqPO.HeadquarterStatus = constant.HeadquarterStatusApproved
		hqPO.PassTime = time.Now().Unix()
		err = hqPORepo.Update(hqPO)
		if err != nil {
			return errors.WithMessage(errors.New("更新总部采购申请状态失败"), err.Error())
		}

		// 查询总部公司信息（预加载 CompanySetting）
		hqCompany, companyErr := repository.NewCompanyRepo(hqTx).GetCompanyInfoByUuid(headquarterUuid)
		if companyErr != nil {
			return errors.WithMessage(errors.New("查询总部公司信息失败"), companyErr.Error())
		}
		if hqCompany == nil {
			return errors.New("查询总部公司信息失败: company not found")
		}
		if hqCompany.CompanySetting == nil {
			return errors.New("总部公司设置不存在")
		}
		hqCompanySetting := *hqCompany.CompanySetting

		// 构造 context 用于 ERP 调用和操作日志（含 Company + CompanySetting，确保下游 ctx.GetCompanySetting() 正确）
		ctx := context.NewContext(
			context.WithCompanyUuid(headquarterUuid),
			context.WithLanguage(lang),
			context.WithStaff(model.Staff{RealName: "系统自动审批"}),
			context.WithCompany(*hqCompany),
			context.WithCompanySetting(hqCompanySetting),
		)

		// 记录自动审批操作日志
		_ = s.helper.createPurchaseOrderLog(hqTx, hqPO.Uuid, ctx, "approve", "自动审批通过",
			constant.PurchaseOrderStatusPending, constant.PurchaseOrderStatusApproved, "品牌采购自动审批", "{}")

		// 调用 ERP 审核流程
		err = s.handleErpApproval(ctx, hqTx, &hqCompanySetting, hqPO, true)
		if err != nil {
			return err
		}

		// syncToSubShop: 后台任务不在子店事务内，可安全使用新连接同步
		return s.syncToSubShop(hqPO)
	})
}

// handleErpApproval 处理ERP审核
// autoApprove: 品牌采购自动审批标记，true 时 ERP 侧 SO 自动提交并生成 DN
func (s *purchaseOrderSrv) handleErpApproval(
	ctx context.Context,
	tx *gorm.DB,
	companySetting *model.CompanySetting,
	purchaseOrder *model.PurchaseOrder,
	autoApprove bool,
) error {
	purchaseOrderRepo := repository.NewPurchaseOrderRepo(tx)

	var erpOrderNo, erpSaleOrderNo string
	var err error

	if purchaseOrder.IsHeadquarterPurchase() {
		// 内部采购
		erpOrderNo, erpSaleOrderNo, err = s.handleInternalPurchaseErp(ctx, tx, purchaseOrder, autoApprove)
		if err != nil {
			return err
		}
	} else {
		// 外部采购
		erpOrderNo, err = s.handleExternalPurchaseErp(ctx, tx, companySetting, purchaseOrder)
		if err != nil {
			return err
		}
	}

	// 更新采购申请单号
	purchaseOrder.ErpOrderNo = erpOrderNo
	purchaseOrder.ErpSaleOrderNo = erpSaleOrderNo
	err = purchaseOrderRepo.Update(purchaseOrder)
	if err != nil {
		return errors.WithMessage(errors.New("更新采购申请单号失败"), err.Error())
	}

	// 同步状态到子商户采购申请
	// 自动审批路径下跳过，由 doAutoApproveHeadquarter 在 ERP 完成后统一调用 syncToSubShop
	if purchaseOrder.IsHeadquarterPurchase() && !autoApprove {
		return s.syncToSubShop(purchaseOrder)
	}

	return nil
}

// handleInternalPurchaseErp 处理内部采购ERP
// autoApprove: 品牌采购自动审批标记
func (s *purchaseOrderSrv) handleInternalPurchaseErp(
	ctx context.Context,
	tx *gorm.DB,
	purchaseOrder *model.PurchaseOrder,
	autoApprove bool,
) (string, string, error) {
	// 获取子公司数据库
	subDb := s.dbm.GetDB(purchaseOrder.CompanyUuid)
	if subDb == nil {
		return "", "", errors.New("获取子店数据库失败")
	}

	// 查询上次品牌采购数量（默认销售单位），从子店数据库查询
	lastPurchaseQtyByCode := s.helper.getLastPurchaseQtyByMaterialCode(subDb, purchaseOrder)

	// 构建物料请求项
	stockItems := make([]*stock.MaterialRequestItem, 0, len(purchaseOrder.Items))
	for _, item := range purchaseOrder.Items {
		if len(item.Units) == 0 && item.BaseUnitUuid != 0 {
			if item.Num <= 0 {
				continue
			}
			erpnextUom := item.ErpnextUom
			if erpnextUom == "" {
				materialUnit, err := repository.NewMaterialUnitRepo(tx).GetMaterialUnitsByUuid(item.UnitUuid)
				if err != nil {
					return "", "", errors.WithMessage(errors.New("查询物品单位失败"), err.Error())
				}
				if materialUnit.Unit == nil {
					return "", "", errors.New("查询物品原始单位失败")
				}
				erpnextUom = materialUnit.Unit.ErpnextUom
			}
			stockItems = append(stockItems, &stock.MaterialRequestItem{
				ItemCode:              item.MaterialCode,
				Qty:                   item.Num,
				ScheduleDate:          purchaseOrder.ExpectArrivalTime,
				Uom:                   erpnextUom,
				CustomLastPurchaseQty: lastPurchaseQtyByCode[item.MaterialCode],
			})
		} else {
			for _, unit := range item.Units {
				if unit.Num <= 0 {
					continue
				}
				stockItems = append(stockItems, &stock.MaterialRequestItem{
					ItemCode:              item.MaterialCode,
					Qty:                   unit.Num,
					ScheduleDate:          purchaseOrder.ExpectArrivalTime,
					Uom:                   unit.ErpnextUom,
					CustomLastPurchaseQty: lastPurchaseQtyByCode[item.MaterialCode],
				})
			}
		}
	}

	// 查询物品在 ERP 中的默认仓库映射（新流程：采购单无指定仓库时使用）
	itemDefaultWarehouseMap := make(map[string]uint64)
	if purchaseOrder.WarehouseErpCode == "" {
		itemCodes := make([]string, 0, len(purchaseOrder.Items))
		for _, item := range purchaseOrder.Items {
			if item.MaterialCode != "" {
				itemCodes = append(itemCodes, item.MaterialCode)
			}
		}
		companySetting := ctx.GetCompanySetting()
		hqCompanyAbbr := companySetting.ErpnextCompanyAbbr // 此处是总部审核，companySetting 为总部
		erpSrv := erp.NewIErpSrv(s.dbm)
		erpItemWarehouses, erpErr := erpSrv.GetItemDefaultWarehouses(ctx, itemCodes, hqCompanyAbbr, false)
		if erpErr != nil {
			logger.Logger.Warn("handleInternalPurchaseErp-查询ERP物品默认仓库失败",
				zap.Uint64("company_uuid", companySetting.CompanyUuid),
				zap.Error(erpErr),
			)
		} else {
			// 批量将 warehouse_erp_code 映射为本地仓库 UUID
			warehouseErpCodes := make(map[string]bool)
			for _, whErpCode := range erpItemWarehouses {
				if whErpCode != "" {
					warehouseErpCodes[whErpCode] = true
				}
			}
			codes := make([]string, 0, len(warehouseErpCodes))
			for code := range warehouseErpCodes {
				codes = append(codes, code)
			}
			warehouseRepo := repository.NewWarehouseRepo(tx)
			whMap, whErr := warehouseRepo.GetByErpCodes(codes)
			if whErr != nil {
				// 非致命：查询失败时跳过按物品默认仓库减库存，依赖订单级仓库兜底
				logger.Logger.Warn("批量查询仓库ERP编码失败",
					zap.Uint64("company_uuid", ctx.GetCompanyUuid()),
					zap.Error(whErr))
			}
			for itemCode, whErpCode := range erpItemWarehouses {
				if wh, ok := whMap[whErpCode]; ok {
					itemDefaultWarehouseMap[itemCode] = wh.Uuid
				}
			}
		}
	}

	// 减总部库存并记录出入库日志
	err := s.helper.reduceHeadquarterStockAndLog(ctx, subDb, tx, purchaseOrder, itemDefaultWarehouseMap)
	if err != nil {
		return "", "", err
	}

	// 获取子公司设置
	subCompanySetting := repository.NewCompanySettingRepo(subDb).Get()

	// 调用erp接口
	hqCompanySetting := ctx.GetCompanySetting() // 审核时为总部
	stockResp, err := erp.NewIErpSrv(s.dbm).SaveMaterialRequest(ctx, subCompanySetting, &stock.SaveMaterialRequestReq{
		TransactionDate: purchaseOrder.OrderTime,
		RequiredBy:      purchaseOrder.ExpectArrivalTime,
		Supplier: func() string {
			if purchaseOrder.SupplierErpCode != "" {
				return purchaseOrder.SupplierErpCode
			}
			return purchaseOrder.SupplierName
		}(),
		SourceWarehouse:   purchaseOrder.WarehouseErpCode,
		TargetWarehouse:   purchaseOrder.DefaultWarehouseErpCode,
		Items:             stockItems,
		RefNo:             purchaseOrder.OrderNo, // 来源单据号，用于跟踪ttpos原始订单号
		SourceCompanyAbbr: hqCompanySetting.ErpnextCompanyAbbr,
		AutoApprove:       autoApprove,
	})
	if err != nil {
		return "", "", s.helper.handleErpError(ctx, err, purchaseOrder)
	}

	return stockResp.PurchaseOrder, stockResp.SalesOrder, nil
}

// handleExternalPurchaseErp 处理外部采购ERP
func (s *purchaseOrderSrv) handleExternalPurchaseErp(
	ctx context.Context,
	tx *gorm.DB,
	companySetting *model.CompanySetting,
	purchaseOrder *model.PurchaseOrder,
) (string, error) {
	// 获取在途仓库
	transitWarehouse, _ := repository.NewWarehouseRepo(tx).GetTransitWarehouse()
	// 获取供应商ID
	supplierUuid := func() uint64 {
		supplier, err := repository.NewSupplierRepo(tx).GetByErpCode(purchaseOrder.SupplierErpCode)
		if err != nil {
			return 0
		}
		return supplier.Uuid
	}()
	// 构建采购订单项
	stockItems := make([]*buying.PurchaseOrderItemInput, 0, len(purchaseOrder.Items))
	for _, item := range purchaseOrder.Items {
		actualNum := 0.0
		if len(item.Units) == 0 && item.BaseUnitUuid != 0 {
			// 获取实际数量
			actualNum = item.GetConversionRateNum()
			if actualNum <= 0 {
				continue
			}
			//
			erpnextUom := item.ErpnextUom
			if erpnextUom == "" {
				materialUnit, err := repository.NewMaterialUnitRepo(tx).GetMaterialUnitsByUuid(item.UnitUuid)
				if err != nil {
					return "", errors.WithMessage(errors.New("查询物品单位失败"), err.Error())
				}
				if materialUnit.Unit == nil {
					return "", errors.New("查询物品原始单位失败")
				}
				erpnextUom = materialUnit.Unit.ErpnextUom
			}
			stockItems = append(stockItems, &buying.PurchaseOrderItemInput{
				ItemCode: item.MaterialCode,
				Qty:      item.Num,
				Uom:      erpnextUom,
			})
		} else {
			for _, unit := range item.Units {
				unitActualNum := unit.GetConversionRateNum()
				if unitActualNum <= 0 {
					continue
				}
				actualNum += unitActualNum
				stockItems = append(stockItems, &buying.PurchaseOrderItemInput{
					ItemCode: item.MaterialCode,
					Qty:      unit.Num,
					Uom:      unit.ErpnextUom,
				})
			}
		}

		// 添加到本店的在途仓库
		if transitWarehouse != nil {
			err := s.helper.AddToTransitWarehouse(tx, transitWarehouse, purchaseOrder, supplierUuid, &item, actualNum)
			if err != nil {
				return "", err
			}
		}
	}

	// 将时间戳转换为 Y-m-d 格式
	scheduleDate := ""
	if purchaseOrder.ExpectArrivalTime > 0 {
		scheduleDate = time.Unix(purchaseOrder.ExpectArrivalTime, 0).Format("2006-01-02")
	}

	// 调用ERP接口
	erpResp, err := erp.NewIErpSrv(s.dbm).CreatePurchaseOrder(ctx, &buying.CreatePurchaseOrderReq{
		Supplier: func() string {
			if purchaseOrder.SupplierErpCode != "" {
				return purchaseOrder.SupplierErpCode
			}
			return purchaseOrder.SupplierName
		}(),
		CompanyAbbr:     companySetting.ErpnextCompanyAbbr,
		ScheduleDate:    scheduleDate,
		TargetWarehouse: purchaseOrder.DefaultWarehouseErpCode,
		Items:           stockItems,
	})
	if err != nil {
		return "", s.helper.handleErpError(ctx, err, purchaseOrder)
	}

	return erpResp.Name, nil
}

// syncToSubShop 同步状态到子商户
func (s *purchaseOrderSrv) syncToSubShop(purchaseOrder *model.PurchaseOrder) error {
	// 获取子店数据库
	subDb := s.dbm.GetDB(purchaseOrder.CompanyUuid)
	if subDb == nil {
		return errors.New("获取子店数据库失败")
	}

	subPurchaseOrder, err := repository.NewPurchaseOrderRepo(subDb).GetByUuid(purchaseOrder.SubUuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("采购申请不存在")
		}
		return errors.WithMessage(errors.New("查询采购申请失败"), err.Error())
	}

	subPurchaseOrder.ErpOrderNo = purchaseOrder.ErpOrderNo
	subPurchaseOrder.Status = constant.PurchaseOrderStatusApproved
	subPurchaseOrder.HeadquarterStatus = constant.HeadquarterStatusApproved
	subPurchaseOrder.PassTime = purchaseOrder.PassTime
	subPurchaseOrder.ErpSaleOrderNo = purchaseOrder.ErpSaleOrderNo

	err = repository.NewPurchaseOrderRepo(subDb).Update(subPurchaseOrder)
	if err != nil {
		return errors.WithMessage(errors.New("更新子店采购申请失败"), err.Error())
	}

	// 同步物品的供应商配送字段（从总部 Material 复制到子店 PurchaseOrderItem）
	if err := s.syncItemSupplierFieldsToSubShop(subDb, purchaseOrder); err != nil {
		return errors.WithMessage(errors.New("同步物品供应商字段失败"), err.Error())
	}

	return nil
}

// syncItemSupplierFieldsToSubShop 同步物品供应商字段到子商户
func (s *purchaseOrderSrv) syncItemSupplierFieldsToSubShop(subDb *gorm.DB, purchaseOrder *model.PurchaseOrder) error {
	// 构建 MaterialCode -> 供应商字段 的映射
	materialSupplierMap := make(map[string]struct {
		DeliveredBySupplier int
		SupplierErpCode     string
	})
	for _, item := range purchaseOrder.Items {
		if item.Material != nil {
			materialSupplierMap[item.MaterialCode] = struct {
				DeliveredBySupplier int
				SupplierErpCode     string
			}{
				DeliveredBySupplier: item.Material.DeliveredBySupplier,
				SupplierErpCode:     item.Material.SupplierErpCode,
			}
		}
	}

	if len(materialSupplierMap) == 0 {
		return nil
	}

	// 获取子店采购单的物品
	subItems, err := repository.NewPurchaseOrderItemRepo(subDb).GetByPurchaseOrderUuid(purchaseOrder.SubUuid)
	if err != nil {
		return err
	}

	// 批量更新子店物品的供应商字段
	subItemRepo := repository.NewPurchaseOrderItemRepo(subDb)
	for _, subItem := range subItems {
		if supplierInfo, ok := materialSupplierMap[subItem.MaterialCode]; ok {
			if err := subItemRepo.UpdateByUuid(subItem.Uuid, map[string]any{
				"delivered_by_supplier": supplierInfo.DeliveredBySupplier,
				"supplier_erp_code":     supplierInfo.SupplierErpCode,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

// syncItemsToCompanyShop 同步采购明细到归属店铺
func (s *purchaseOrderSrv) syncItemsToCompanyShop(
	ctx context.Context,
	purchaseOrder *model.PurchaseOrder,
	reqItems []req.PurchaseOrderItemUpdateReq,
) error {
	// 获取归属店铺的数据库
	companyDb := s.dbm.GetDB(purchaseOrder.CompanyUuid)
	if companyDb == nil {
		return errors.New("获取归属店铺数据库失败")
	}

	// 获取当前操作数据库的订单最新明细（用于获取 material_code 和单位信息）
	currentItemRepo := repository.NewPurchaseOrderItemRepo(ctx.GetDB())
	currentItems, err := currentItemRepo.GetByPurchaseOrderUuid(
		purchaseOrder.Uuid,
		currentItemRepo.WithPreloadUnits(),
	)
	if err != nil {
		return errors.WithMessage(errors.New("查询当前采购明细失败"), err.Error())
	}

	return companyDb.Transaction(func(tx *gorm.DB) error {
		companyOrderRepo := repository.NewPurchaseOrderRepo(tx)
		companyItemRepo := repository.NewPurchaseOrderItemRepo(tx)
		companyItemUnitRepo := repository.NewPurchaseOrderItemUnitRepo(tx)

		// 查询归属店铺的采购单
		companyOrder, err := companyOrderRepo.GetByUuid(
			purchaseOrder.SubUuid,
			companyOrderRepo.WithItems(),
		)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 如果归属店铺没有这个采购单，不报错，直接返回
				return nil
			}
			return errors.WithMessage(errors.New("查询归属店铺采购申请失败"), err.Error())
		}

		// 检查是否需要同步（优化：避免不必要的数据库操作）
		needSync := s.helper.checkCompanyShopNeedSync(companyOrder, currentItems, reqItems)
		if !needSync {
			return nil
		}

		// 获取归属店铺现有的明细项
		// 构建现有明细的映射：MaterialCode -> Item
		existingItemMap := make(map[string]*model.PurchaseOrderItem)
		for i := range companyOrder.Items {
			existingItemMap[companyOrder.Items[i].MaterialCode] = &companyOrder.Items[i]
		}

		// 构建当前明细的映射：MaterialUuid -> MaterialCode
		materialCodeMap := make(map[uint64]string)
		for _, item := range currentItems {
			materialCodeMap[item.MaterialUuid] = item.MaterialCode
		}

		// 构建请求中的物品映射：MaterialCode -> PurchaseOrderItemReq
		reqItemMap := make(map[string]PurchaseOrderItemReq)
		for _, item := range reqItems {
			if materialCode, ok := materialCodeMap[item.MaterialUuid]; ok {
				reqItemMap[materialCode] = PurchaseOrderItemReq{
					MaterialUuid: item.MaterialUuid,
					UnitList:     item.UnitList,
				}
			}
		}

		// 1. 找出需要删除的明细（归属店铺存在，但请求中不存在）
		itemUuidsToDelete := make([]uint64, 0)
		for materialCode, existingItem := range existingItemMap {
			if _, exists := reqItemMap[materialCode]; !exists {
				itemUuidsToDelete = append(itemUuidsToDelete, existingItem.Uuid)
			}
		}

		// 2. 找出需要新增和更新的明细
		itemReqsToCreate := make([]PurchaseOrderItemReq, 0)
		for materialCode, reqItem := range reqItemMap {
			if existingItem, exists := existingItemMap[materialCode]; exists {
				// 已存在，需要更新（删除旧的，创建新的）
				itemUuidsToDelete = append(itemUuidsToDelete, existingItem.Uuid)
			}
			// 都加入创建列表
			itemReqsToCreate = append(itemReqsToCreate, reqItem)
		}

		// 执行删除操作
		if len(itemUuidsToDelete) > 0 {
			// 先删除关联的单位
			err = companyItemUnitRepo.DeleteByItemUuids(itemUuidsToDelete)
			if err != nil {
				return errors.WithMessage(errors.New("删除归属店铺明细单位失败"), err.Error())
			}
			// 再删除明细项
			err = companyItemRepo.DeleteByUuids(itemUuidsToDelete)
			if err != nil {
				return errors.WithMessage(errors.New("删除归属店铺明细失败"), err.Error())
			}
		}

		// 执行新增操作
		if len(itemReqsToCreate) > 0 {
			// 1. 尝试使用子店铺的物料配置构建明细（优先使用子店铺配置）
			itemsFromCompany, unitsFromCompany, err := s.validator.buildPurchaseOrderItems(tx, companyOrder.Uuid, itemReqsToCreate, true)
			if err != nil {
				return err
			}

			// 2. 找出在子店铺中不存在的物料（需要从总部复制）
			companyMaterialUuids := make(map[uint64]bool)
			for _, item := range itemsFromCompany {
				companyMaterialUuids[item.MaterialUuid] = true
			}

			// 3. 构建缺失物料的请求列表
			missingItemReqs := make([]PurchaseOrderItemReq, 0)
			for _, itemReq := range itemReqsToCreate {
				if !companyMaterialUuids[itemReq.MaterialUuid] {
					missingItemReqs = append(missingItemReqs, itemReq)
				}
			}

			// 4. 使用总部数据库构建缺失的物料明细
			var itemsFromHeadquarter []model.PurchaseOrderItem
			if len(missingItemReqs) > 0 {
				itemsFromHeadquarter, _, err = s.validator.buildPurchaseOrderItems(ctx.GetDB(), companyOrder.Uuid, missingItemReqs)
				if err != nil {
					return errors.WithMessage(errors.New("从总部构建明细失败"), err.Error())
				}
			}

			// 5. 检查子店铺物料是否缺少采购单位，补充缺失的单位
			// 从总部数据库查询完整的单位信息
			hqItemsWithUnits, _, err := s.validator.buildPurchaseOrderItems(ctx.GetDB(), companyOrder.Uuid, itemReqsToCreate)
			if err != nil {
				return errors.WithMessage(errors.New("从总部查询完整单位信息失败"), err.Error())
			}

			// 构建总部单位映射：MaterialUuid -> Units
			hqUnitsMap := make(map[uint64][]model.PurchaseOrderItemUnit)
			for _, item := range hqItemsWithUnits {
				hqUnitsMap[item.MaterialUuid] = item.Units
			}

			// 构建子店已有单位映射：MaterialUuid -> map[UnitUuid]bool
			companyUnitsMap := make(map[uint64]map[uint64]bool)
			for _, unit := range unitsFromCompany {
				if _, exists := companyUnitsMap[unit.ItemUuid]; !exists {
					companyUnitsMap[unit.ItemUuid] = make(map[uint64]bool)
				}
				companyUnitsMap[unit.ItemUuid][unit.UnitUuid] = true
			}

			// 为子店铺的明细补充缺失的单位
			for i := range itemsFromCompany {
				item := &itemsFromCompany[i]
				hqUnits, hasHqUnits := hqUnitsMap[item.MaterialUuid]
				if !hasHqUnits {
					continue
				}

				companyUnits, hasCompanyUnits := companyUnitsMap[item.Uuid]

				// 找出子店铺缺失的单位
				for _, hqUnit := range hqUnits {
					// 如果子店铺没有这个单位，则从总部复制
					if !hasCompanyUnits || !companyUnits[hqUnit.UnitUuid] {
						// 复制总部单位信息，但使用子店铺的ItemUuid
						newUnit := model.PurchaseOrderItemUnit{
							ItemUuid:           item.Uuid,
							PurchaseOrderUuid:  companyOrder.Uuid,
							UnitUuid:           hqUnit.UnitUuid,
							Num:                hqUnit.Num,
							UnitName:           hqUnit.UnitName,
							UnitConversionRate: hqUnit.UnitConversionRate,
							ErpnextUom:         hqUnit.ErpnextUom,
							BaseUnitUuid:       hqUnit.BaseUnitUuid,
							BaseUnitName:       hqUnit.BaseUnitName,
						}
						item.Units = append(item.Units, newUnit)
					}
				}
			}

			// 6. 合并子店铺和总部的明细
			allItems := append(itemsFromCompany, itemsFromHeadquarter...)

			// 批量创建明细
			if len(allItems) > 0 {
				err = companyItemRepo.CreateBatch(allItems)
				if err != nil {
					return errors.WithMessage(errors.New("创建归属店铺明细失败"), err.Error())
				}
			}
		}

		// 更新归属店铺采购单的物品数量
		companyOrder.Num = float64(len(reqItems))
		err = companyOrderRepo.Update(companyOrder)
		if err != nil {
			return errors.WithMessage(errors.New("更新归属店铺采购申请失败"), err.Error())
		}

		return nil
	})
}

// 收货单相关方法委托给receiptSrv
func (s *purchaseOrderSrv) CreatePurchaseReceiptOrder(
	ctx context.Context,
	req req.PurchaseReceiptCreateReq,
) (resp.PurchaseReceiptOrderCreateResp, error) {
	// 加锁
	s.lock.LockUuid(req.PurchaseOrderUuid)
	defer s.lock.UnlockUuid(req.PurchaseOrderUuid)
	// 查询采购申请
	purchaseOrder, err := repository.NewPurchaseOrderRepo(ctx.GetDB()).GetByUuid(req.PurchaseOrderUuid)
	if err != nil {
		return resp.PurchaseReceiptOrderCreateResp{}, errors.WithMessage(errors.New("采购申请不存在"), err.Error())
	}
	// 创建收货单
	result, err := s.receiptSrv.CreatePurchaseReceiptOrder(ctx, req)
	if err != nil {
		// 调用erp接口失败
		if purchaseOrder != nil && strings.Contains(err.Error(), "调用erp接口失败") {
			return resp.PurchaseReceiptOrderCreateResp{}, s.helper.handleErpError(ctx, err, purchaseOrder)
		}
		return resp.PurchaseReceiptOrderCreateResp{}, err
	}
	return result, nil
}

// 更新收货单
func (s *purchaseOrderSrv) UpdatePurchaseReceiptOrder(
	ctx context.Context,
	req req.PurchaseReceiptOrderUpdateReq,
) error {
	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)
	// 查询收货单
	receiptOrder, err := repository.NewPurchaseReceiptOrderRepo(ctx.GetDB()).GetByUuid(req.Uuid)
	if err != nil {
		return errors.WithMessage(errors.New("收货单不存在"), err.Error())
	}
	// 查询采购申请
	purchaseOrder, err := repository.NewPurchaseOrderRepo(ctx.GetDB()).GetByUuid(receiptOrder.PurchaseOrderUuid)
	if err != nil {
		return errors.WithMessage(errors.New("采购申请不存在"), err.Error())
	}
	//
	err = s.receiptSrv.UpdatePurchaseReceiptOrder(ctx, req)
	if err != nil {
		// 调用erp接口失败
		if purchaseOrder != nil && strings.Contains(err.Error(), "调用erp接口失败") {
			return s.helper.handleErpError(ctx, err, purchaseOrder)
		}
		return err
	}
	return nil
}

// 获取收货单列表
func (s *purchaseOrderSrv) GetPurchaseReceiptOrderList(
	ctx context.Context,
	reqs req.PurchaseReceiptOrderListReq,
) (resp.PurchaseReceiptOrderListResp, error) {
	if ctx.Version(context.LT, "2.7.0") {
		return s.receiptSrv.GetPurchaseReceiptOrderList(ctx, reqs)
	}

	// 时间格式兼容处理函数：前端传毫秒时间戳，数据库存秒时间戳
	convertTimestamp := func(timestamp int64) int64 {
		if timestamp <= 0 {
			return 0
		}
		// 如果是13位毫秒时间戳，转换为10位秒时间戳
		if timestamp > 9999999999 {
			return timestamp / 1000
		}
		return timestamp
	}

	// 状态筛选
	status := constant.PurchaseOrderStatusApproved
	if len(reqs.StatusIn) > 0 && reqs.StatusIn[0] != 0 {
		status = -1
	}
	if len(reqs.StatusIn) == 0 {
		reqs.StatusIn = []int{0, 1, 2}
	}

	// 通过Repository执行联合查询
	purchaseOrderRepo := repository.NewPurchaseOrderRepo(ctx.GetDB())
	results, total, err := purchaseOrderRepo.GetPurchaseReceiptOrderUnionList(repository.PurchaseReceiptUnionListParams{
		Status:           status,
		StatusIn:         reqs.StatusIn,
		ReceiptType:      reqs.ReceiptType,
		OrderNoPattern:   "%" + reqs.OrderNo + "%",
		ReceiveTimeStart: convertTimestamp(reqs.ReceiptTimeStart),
		ReceiveTimeEnd:   convertTimestamp(reqs.ReceiptTimeEnd),
		PageSize:         reqs.PageReq.PageSize,
		PageNo:           reqs.PageReq.PageNo,
	})
	if err != nil {
		return resp.PurchaseReceiptOrderListResp{}, errors.WithMessage(errors.New("查询列表失败"), err.Error())
	}

	// 分别处理采购单和收货单
	var purchaseOrderUuids []uint64
	var receiptOrderUuids []uint64
	for _, result := range results {
		if result.Type == 2 { // 采购单
			purchaseOrderUuids = append(purchaseOrderUuids, result.Uuid)
		} else if result.Type == 1 { // 收货单
			receiptOrderUuids = append(receiptOrderUuids, result.Uuid)
		}
	}

	// 构建响应数据
	response := resp.PurchaseReceiptOrderListResp{
		PurchaseOrderList: []*resp.PurchaseOrderInfo{},
		List:              []*resp.PurchaseReceiptOrderInfo{},
		Meta: dto.PageResponse{
			PageNo:   reqs.PageReq.PageNo,
			PageSize: reqs.PageReq.PageSize,
			Total:    total,
		},
	}

	// 如果有采购单，获取采购单详情
	if len(purchaseOrderUuids) > 0 {
		purchaseOrderList, err := s.GetPurchaseOrderList(ctx, req.PurchaseOrderListReq{
			UuidIn:       purchaseOrderUuids,
			PurchaseType: reqs.ReceiptType,
			PageReq: dto.PageReq{
				PageNo:   1,
				PageSize: len(purchaseOrderUuids),
			},
		})
		if err != nil {
			return resp.PurchaseReceiptOrderListResp{}, errors.WithMessage(errors.New("查询采购单列表失败"), err.Error())
		}
		response.PurchaseOrderList = purchaseOrderList.List
	}

	// 如果有收货单，获取收货单详情
	if len(receiptOrderUuids) > 0 {
		receiptOrderList, err := s.receiptSrv.GetPurchaseReceiptOrderList(ctx, req.PurchaseReceiptOrderListReq{
			UuidIn:      receiptOrderUuids,
			ReceiptType: reqs.ReceiptType,
			PageReq: dto.PageReq{
				PageNo:   1,
				PageSize: len(receiptOrderUuids),
			},
		})
		if err != nil {
			return resp.PurchaseReceiptOrderListResp{}, errors.WithMessage(errors.New("查询收货单列表失败"), err.Error())
		}
		response.List = receiptOrderList.List
	}

	return response, nil
}

// 获取收货单详情
func (s *purchaseOrderSrv) GetPurchaseReceiptOrderDetail(
	ctx context.Context,
	req req.PurchaseReceiptOrderDetailReq,
) (resp.PurchaseReceiptOrderDetailResp, error) {
	return s.receiptSrv.GetPurchaseReceiptOrderDetail(ctx, req)
}

// 取消收货单
func (s *purchaseOrderSrv) CancelPurchaseReceiptOrder(
	ctx context.Context,
	req req.PurchaseReceiptOrderCancelReq,
) error {
	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	return s.receiptSrv.CancelPurchaseReceiptOrder(ctx, req)
}

// GetPurchaseReceiptNewList 新收货单列表（按采购单维度）
// 返回采购单OrderTime、采购单收货状态、单据编号、采购单号、申请单物品数量、确认收货次数、总收货进度
func (s *purchaseOrderSrv) GetPurchaseReceiptNewList(
	ctx context.Context,
	reqData req.PurchaseReceiptNewListReq,
) (resp.PurchaseReceiptNewListResp, error) {

	db := ctx.GetDB()
	purchaseOrderRepo := repository.NewPurchaseOrderRepo(db)
	receiptOrderRepo := repository.NewPurchaseReceiptOrderRepo(db)

	// 构建查询选项 - 只查询品牌采购（内部采购）且已通过总部审核的采购单
	var opts []repository.DBOption

	// 只查询品牌采购（内部采购）
	opts = append(opts, purchaseOrderRepo.WherePurchaseType(constant.PurchaseTypeBrand))

	// 根据收货状态筛选
	if reqData.ReceiptStatus != nil {
		if *reqData.ReceiptStatus == 0 {
			// 待收货（未完全收货）：状态为 2-已通过
			opts = append(opts, purchaseOrderRepo.WhereStatusIn([]int{
				constant.PurchaseOrderStatusApproved,
			}))
		} else if *reqData.ReceiptStatus == 1 {
			// 已收货（已完全收货）：状态为 4-已完成
			opts = append(opts, purchaseOrderRepo.WhereStatusIn([]int{
				constant.PurchaseOrderStatusCompleted,
			}))
		}
	} else {
		// 默认查询所有可收货状态的采购单
		opts = append(opts, purchaseOrderRepo.WhereStatusIn([]int{
			constant.PurchaseOrderStatusApproved,
			constant.PurchaseOrderStatusCompleted,
		}))
	}

	// 单据编号筛选
	if reqData.OrderNo != "" {
		opts = append(opts, purchaseOrderRepo.WhereOrderNo(reqData.OrderNo))
	}

	// 采购单号筛选（ErpOrderNo）
	if reqData.ErpOrderNo != "" {
		opts = append(opts, func(db *gorm.DB) *gorm.DB {
			return db.Where("erp_order_no LIKE ?", "%"+reqData.ErpOrderNo+"%")
		})
	}

	// 按订单时间倒序排列
	opts = append(opts, purchaseOrderRepo.OrderByOrderTime(true))

	// 查询关联的物品明细以计算物品数量
	opts = append(opts, purchaseOrderRepo.WithSimpleItems())

	// 分页查询采购单
	purchaseOrders, total, err := purchaseOrderRepo.GetListWithPagination(
		reqData.PageNo,
		reqData.PageSize,
		opts...,
	)
	if err != nil {
		return resp.PurchaseReceiptNewListResp{}, errors.WithMessage(errors.New("查询采购单列表失败"), err.Error())
	}

	// 收集需要调用ERP的新采购单（有erp_sale_order_no）
	erpOrderNoToProgress := make(map[string]float64)
	var erpOrderNos []string
	for _, po := range purchaseOrders {
		if po.ErpSaleOrderNo != "" && po.ErpOrderNo != "" {
			erpOrderNos = append(erpOrderNos, po.ErpOrderNo)
		}
	}

	// 批量调用ERP获取收货进度
	if len(erpOrderNos) > 0 {
		erpSrv := erp.NewIErpSrv(s.dbm)
		erpResp, err := erpSrv.GetPurchaseOrderList(ctx, &buying.GetPurchaseOrderListReq{
			Name: strings.Join(erpOrderNos, ","),
		})
		if err != nil {
			logger.Logger.Warn("GetPurchaseReceiptNewList批量调用ERP获取采购单列表失败，使用本地计算",
				zap.Int("count", len(erpOrderNos)),
				zap.Error(err),
			)
		} else {
			for _, erpOrder := range erpResp.PurchaseOrders {
				erpOrderNoToProgress[erpOrder.Name] = erpOrder.PerReceived
			}
		}
	}

	// 构建响应列表
	list := make([]*resp.PurchaseReceiptNewListItem, 0, len(purchaseOrders))

	for _, po := range purchaseOrders {
		// 统计确认收货次数（status=1 已收货的收货单数量）
		confirmReceiptCount, err := receiptOrderRepo.Count(
			receiptOrderRepo.WherePurchaseOrderUuid(po.Uuid),
			receiptOrderRepo.WhereStatusIn([]int{constant.ReceiptOrderStatusReceived}),
		)
		if err != nil {
			logger.Logger.Error("统计确认收货次数失败",
				zap.Uint64("purchase_order_uuid", po.Uuid),
				zap.Error(err),
			)
			confirmReceiptCount = 0
		}

		// 计算收货状态：0-待收货（未完全收货）1-已收货（已完全收货）
		receiptStatus := 0
		if po.Status == constant.PurchaseOrderStatusCompleted {
			receiptStatus = 1
		}

		// 计算收货进度：新采购单优先使用ERP返回的进度，旧采购单使用本地计算
		var receiptProgress string
		if po.ErpSaleOrderNo != "" && po.ErpOrderNo != "" {
			// 新采购单：优先使用ERP返回的进度
			if progress, ok := erpOrderNoToProgress[po.ErpOrderNo]; ok {
				receiptProgress = fmt.Sprintf("%.2f%%", progress)
			} else {
				// ERP未返回时使用本地计算
				receiptProgress = fmt.Sprintf("%.2f%%", po.GetReceiptProgress())
			}
		} else {
			// 旧采购单：使用本地计算
			receiptProgress = fmt.Sprintf("%.2f%%", po.GetReceiptProgress())
		}

		item := &resp.PurchaseReceiptNewListItem{
			Uuid:              po.Uuid,
			OrderNo:           po.OrderNo,
			ErpOrderNo:        po.ErpOrderNo,
			OrderTime:         po.OrderTime,
			ReceiptStatus:     receiptStatus,
			ItemNum:           len(po.Items),
			ConfirmReceiptNum: int(confirmReceiptCount),
			ReceiptProgress:   receiptProgress,
		}
		list = append(list, item)
	}

	return resp.PurchaseReceiptNewListResp{
		List: list,
		Meta: dto.PageResponse{
			PageNo:   reqData.PageNo,
			PageSize: reqData.PageSize,
			Total:    total,
		},
	}, nil
}

// GetBrandPurchaseAutoApprove 获取品牌采购自动审批开关
func (s *purchaseOrderSrv) GetBrandPurchaseAutoApprove(ctx context.Context) (int, error) {
	companySetting := ctx.GetCompanySetting()
	if !companySetting.IsHeadquarter() {
		return 0, errors.New("仅总部可操作")
	}
	return companySetting.BrandPurchaseAutoApprove, nil
}

// SetBrandPurchaseAutoApprove 设置品牌采购自动审批开关
func (s *purchaseOrderSrv) SetBrandPurchaseAutoApprove(ctx context.Context, value int) error {
	companySetting := ctx.GetCompanySetting()
	if !companySetting.IsHeadquarter() {
		return errors.New("仅总部可操作")
	}
	if value != 0 && value != 1 {
		return errors.New("参数值无效")
	}
	companyUuid := ctx.GetCompanyUuid()
	db := ctx.GetDB()

	// 更新商户库
	data := map[string]any{"brand_purchase_auto_approve": value}
	if err := repository.NewCompanySettingRepo(db).UpdateByCompanyUuid(companyUuid, data); err != nil {
		return errors.WithMessage(errors.New("保存商户设置失败"), err.Error())
	}

	// 同步更新saas主库
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	if err := repository.NewCompanySettingRepo(saasDB).UpdateByCompanyUuid(companyUuid, data); err != nil {
		return errors.WithMessage(errors.New("保存saas设置失败"), err.Error())
	}

	return nil
}
