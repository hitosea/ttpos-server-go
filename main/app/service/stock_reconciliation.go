package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"
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
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 盘点单服务错误消息常量
const (
	errMsgQueryWarehouseItem   = "查询仓库物品失败"
	errMsgApproveReconcilation = "审核盘点单失败"
	errMsgItemDisabledFmt      = "物品%s状态已关闭，请修改物品状态"
)

// IStockReconciliationSrv 盘点单服务接口
type IStockReconciliationSrv interface {
	GetStockReconciliationList(ctx context.Context, req req.StockReconciliationListReq) (resp.StockReconciliationListResp, error)             // 获取盘点单列表
	GetStockReconciliationDetail(ctx context.Context, req req.StockReconciliationDetailReq) (resp.StockReconciliationDetailResp, error)       // 获取盘点单详情
	GetStockReconciliationTemplate(ctx context.Context) (resp.StockReconciliationTemplateResp, error)                                         // 获取盘点单模板
	SaveStockReconciliation(ctx context.Context, req req.StockReconciliationSaveReq) (uint64, error)                                          // 更新盘点单
	DeleteStockReconciliation(ctx context.Context, req req.StockReconciliationDeleteReq) error                                                // 删除盘点单
	ApproveStockReconciliation(ctx context.Context, req req.StockReconciliationApproveReq) ([]dto.LocaleResponse, error)                      // 审核盘点单
	RejectStockReconciliation(ctx context.Context, req req.StockReconciliationRejectReq) error                                                // 驳回盘点单
	CheckMaterials(ctx context.Context, req req.StockReconciliationCheckMaterialsReq) (resp.StockReconciliationCheckMaterialsListResp, error) // 检查物品
}

// stockReconciliationSrv 盘点单服务实现
type stockReconciliationSrv struct {
	productSrv IProductSrv
	dbm        *database.DBManager
	lock       lock.Lock
}

// NewStockReconciliationSrv 创建盘点单服务
func NewStockReconciliationSrv(dbm *database.DBManager, productSrv IProductSrv) IStockReconciliationSrv {
	return NewStockReconciliationSrvImpl(dbm, productSrv)
}

// NewStockReconciliationSrvImpl 创建盘点单服务实现
func NewStockReconciliationSrvImpl(dbm *database.DBManager, productSrv IProductSrv) IStockReconciliationSrv {
	return &stockReconciliationSrv{
		dbm:        dbm,
		productSrv: productSrv,
		lock:       lock.NewSystemLock(),
	}
}

// GetStockReconciliationList 获取盘点单列表
func (s *stockReconciliationSrv) GetStockReconciliationList(ctx context.Context, req req.StockReconciliationListReq) (resp.StockReconciliationListResp, error) {
	db := ctx.GetDB()
	stockReconciliationRepo := repository.NewStockReconciliationRepo(db)

	// 构建查询选项
	var opts []repository.DBOption

	// 多仓库筛选
	if len(req.WarehouseUuids) > 0 {
		opts = append(opts, stockReconciliationRepo.WhereWarehouseUuids(req.WarehouseUuids))
	}

	// 关键字搜索（单据编号和ERP盘点单号）
	if req.Keyword != "" {
		opts = append(opts, stockReconciliationRepo.WhereKeyword(req.Keyword))
	}

	// 创建时间范围筛选
	if req.StartCreateTime > 0 || req.EndCreateTime > 0 {
		opts = append(opts, stockReconciliationRepo.WhereCreateTimeRange(req.StartCreateTime, req.EndCreateTime))
	}

	// 状态列表筛选
	if len(req.StatusIn) > 0 {
		opts = append(opts, stockReconciliationRepo.WhereStatusIn(req.StatusIn))
	}

	opts = append(opts, stockReconciliationRepo.WithWarehouseMultiLanguageName())

	// 如果版本号大于等于v2.16.0,则按照提交时间排序
	// if ctx.Version(context.GTE, constant.ClientVersionV2160) {
	// 	opts = append(opts, func(db *gorm.DB) *gorm.DB {
	// 		return db.Order("submit_time DESC")
	// 	})
	// } else {
	// 暂时都是按照创建时间排序. 等过几期后所有商家中的调拨单都提交时间不为0时再改回去.
	opts = append(opts, func(db *gorm.DB) *gorm.DB {
		return db.Order("create_time DESC") // 默认按照创建时间排序. 由于之前版本未提交的盘点单没有提交时间,所以按照创建时间排序.
	})
	// }

	// 查询数据
	list, total, err := stockReconciliationRepo.GetStockReconciliationListWithPagination(req.PageNo, req.PageSize, opts...)
	if err != nil {
		return resp.StockReconciliationListResp{}, errors.WithMessage(err, "查询盘点单列表失败")
	}

	// 转换响应数据
	listResp := make([]*resp.StockReconciliationInfo, 0, len(list))

	stockReconciliationUuidList := make([]uint64, 0, len(list))
	for _, item := range list {
		stockReconciliationUuidList = append(stockReconciliationUuidList, item.Uuid)
	}

	// 根据盘点单获取每个盘点单的物品数量，返回map[盘点单UUID]物品数量
	itemsCountMap, err := stockReconciliationRepo.GetStockReconciliationItemCountListByReconciliationUuidList(stockReconciliationUuidList)
	if err != nil {
		return resp.StockReconciliationListResp{}, errors.WithMessage(err, "查询盘点单物品明细失败")
	}

	for _, item := range list {
		info := &resp.StockReconciliationInfo{}
		if err := copier.Copy(info, item); err != nil {
			logger.Logger.Error("转换盘点单信息失败", zap.Error(err))
			continue
		}
		info.WarehouseLocaleName = item.Warehouse.MultiLanguageName.GetNames()
		info.ItemsCount = itemsCountMap[item.Uuid]
		listResp = append(listResp, info)
	}

	return resp.StockReconciliationListResp{
		List: listResp,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

// GetStockReconciliationTemplate 获取盘点单模板
func (s *stockReconciliationSrv) GetStockReconciliationTemplate(ctx context.Context) (resp.StockReconciliationTemplateResp, error) {
	// 异步触发 Stock Entry 合并扣减（确保盘点选物品时 ERPNext 库存是最新的）
	companySetting := ctx.GetCompanySetting()
	if companySetting.IsErpSalesInvoiceMode() {
		utils.Go(func() {
			companyUuid := ctx.GetCompanyUuid()
			erpCtx := ctx.Copy()
			erpCtx.SetDB(s.dbm.GetDB(companyUuid))
			erpStockEntrySrv := NewErpStockEntrySrv(s.dbm)
			if err := erpStockEntrySrv.TriggerStockEntryDeduction(erpCtx, companyUuid); err != nil {
				logger.Logger.Error("盘点模板触发Stock Entry合并扣减失败",
					zap.Uint64("company_uuid", companyUuid),
					zap.Error(err))
			}
		})
	}

	// 调用盘点模板服务获取模板数据
	templateResp, err := s.fetchReconciliationTemplate(ctx)
	if err != nil {
		logger.Logger.Error("获取盘点单模板失败", zap.Error(err))
		return resp.StockReconciliationTemplateResp{
			Data: resp.StockReconciliationTemplateData{
				Daily:    []string{},
				Weekly:   []string{},
				Monthly:  []string{},
				Property: []string{},
			},
		}, nil
	}

	return templateResp, nil
}

// fetchReconciliationTemplate 从盘点模板服务获取模板数据
func (s *stockReconciliationSrv) fetchReconciliationTemplate(_ context.Context) (resp.StockReconciliationTemplateResp, error) {
	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 构建请求 URL
	url := viper.GetString("RECONCILIATION_TEMPLATES_URL")
	if url == "" {
		url = "http://reconciliation_templates:3000"
	}

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return resp.StockReconciliationTemplateResp{}, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	httpResp, err := client.Do(req)
	if err != nil {
		return resp.StockReconciliationTemplateResp{}, fmt.Errorf("调用盘点模板服务失败: %w", err)
	}
	defer httpResp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return resp.StockReconciliationTemplateResp{}, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查 HTTP 状态码
	if httpResp.StatusCode != http.StatusOK {
		return resp.StockReconciliationTemplateResp{}, fmt.Errorf("盘点模板服务返回错误状态码: %d, 响应: %s", httpResp.StatusCode, string(body))
	}

	// 定义包装响应结构
	var apiResp struct {
		Code    int                                  `json:"code"`
		Message string                               `json:"message"`
		Data    resp.StockReconciliationTemplateResp `json:"data"`
	}

	// 解析响应数据
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return resp.StockReconciliationTemplateResp{}, fmt.Errorf("解析响应数据失败: %w, 响应: %s", err, string(body))
	}

	// 检查业务状态码
	if apiResp.Code != 0 {
		return resp.StockReconciliationTemplateResp{}, fmt.Errorf("盘点模板服务返回业务错误: code=%d, message=%s", apiResp.Code, apiResp.Message)
	}

	// 确保切片不为 nil，避免 JSON 返回 null
	if apiResp.Data.Data.Daily == nil {
		apiResp.Data.Data.Daily = []string{}
	}
	if apiResp.Data.Data.Weekly == nil {
		apiResp.Data.Data.Weekly = []string{}
	}
	if apiResp.Data.Data.Monthly == nil {
		apiResp.Data.Data.Monthly = []string{}
	}
	if apiResp.Data.Data.Property == nil {
		apiResp.Data.Data.Property = []string{}
	}

	return apiResp.Data, nil
}

// getBookedStockMap 获取仓库物品的账面库存数量
func (s *stockReconciliationSrv) getBookedQuantityMap(db *gorm.DB, warehouseUuid uint64) (map[uint64]decimal.Decimal, error) {
	bookedStockMap := make(map[uint64]decimal.Decimal)
	warehouseItemRepo := repository.NewWarehouseItemRepo(db)
	warehouseItems, err := warehouseItemRepo.GetWarehouseMaterials(warehouseItemRepo.WhereWarehouseUuid(warehouseUuid))
	if err != nil {
		return bookedStockMap, errors.WithMessage(err, "查询仓库物品列表失败")
	}
	for _, warehouseItem := range warehouseItems {
		bookedStockMap[warehouseItem.MaterialUuid] = decimal.NewFromFloat(warehouseItem.Stock)
	}
	return bookedStockMap, nil
}

// GetStockReconciliationDetail 获取盘点单详情
func (s *stockReconciliationSrv) GetStockReconciliationDetail(ctx context.Context, req req.StockReconciliationDetailReq) (resp.StockReconciliationDetailResp, error) {
	db := ctx.GetDB()
	stockReconciliationRepo := repository.NewStockReconciliationRepo(db)
	var detailResp resp.StockReconciliationDetailResp

	opts := []repository.DBOption{
		stockReconciliationRepo.WhereUuid(req.Uuid),
		stockReconciliationRepo.WithWarehouseMultiLanguageName(),
		stockReconciliationRepo.WithStockReconciliationItemMaterialUnits(),
	}
	// 查询盘点单
	stockReconciliation, err := stockReconciliationRepo.GetStockReconciliation(opts...)
	if err != nil {
		return detailResp, errors.WithMessage(err, "查询盘点单失败")
	}
	if stockReconciliation == nil {
		return detailResp, errors.New("盘点单不存在")
	}

	// 转换响应数据
	if err := copier.Copy(&detailResp, stockReconciliation); err != nil {
		return detailResp, errors.WithMessage(err, "转换盘点单数据失败")
	}
	detailResp.WarehouseName = stockReconciliation.Warehouse.MultiLanguageName.GetNames()

	// 是否可重新提交（已驳回状态且为发起人）
	detailResp.IsCanResubmit = stockReconciliation.Status == constant.StockReconciliationStatusRejected &&
		stockReconciliation.SubmitterStaffUuid == ctx.GetStaffUuid()

	bookedQuantityMap, err := s.getBookedQuantityMap(db, stockReconciliation.WarehouseUuid)
	if err != nil {
		return detailResp, errors.WithMessage(errors.New(errMsgQueryWarehouseItem), err.Error())
	}

	// 物品单位明细
	itemsResp := make([]*resp.StockReconciliationItemInfo, 0, len(stockReconciliation.StockReconciliationItems))
	for _, item := range stockReconciliation.StockReconciliationItems {
		// 明细中的物品已删除，跳过
		if item.DeleteTime > 0 {
			continue
		}
		itemInfo := &resp.StockReconciliationItemInfo{}
		if err := copier.Copy(itemInfo, item); err != nil {
			logger.Logger.Error("转换盘点单物品信息失败", zap.Error(err))
			continue
		}

		itemInfo.BookedQuantity = item.BookedQuantity.InexactFloat64()
		itemInfo.CountedQuantity = item.CountedQuantity.InexactFloat64()

		// 查询物品信息
		if item.Material != nil {
			itemInfo.LocaleName = *language.JsonToLocaleResponse(item.MaterialName)
			if itemInfo.LocaleName.IsNull() {
				itemInfo.LocaleName = *language.JsonToLocaleResponse(item.Material.Name)
			}
			itemInfo.MaterialCode = item.Material.Code
			itemInfo.InternalCode = item.Material.InternalCode
			itemInfo.MaterialBarcode = item.Material.BarcodeValue
		}

		itemInfo.Units = make([]resp.MaterialUnitInfo, 0)
		for _, unit := range item.Material.NotBaseUnitList {
			if unit.Unit != nil {
				itemInfo.Units = append(itemInfo.Units, resp.MaterialUnitInfo{
					MaterialUnitUuid: unit.Uuid,
					UnitUuid:         unit.UnitUuid,
					UnitName:         unit.Unit.MultiLanguageName.GetNames(),
					ConversionRate:   unit.ConversionRate,
					IsDefault:        unit.IsDefault,
				})
			}
		}

		// 查询单位明细
		itemUnits, err := stockReconciliationRepo.GetStockReconciliationItemUnitListByItemUuid(item.Uuid)
		if err == nil && len(itemUnits) > 0 {
			unitsResp := make([]*resp.StockReconciliationItemUnitInfo, 0, len(itemUnits))
			for _, itemUnit := range itemUnits {
				if itemUnit.MaterialUnit == nil || itemUnit.MaterialUnit.Unit == nil || itemUnit.MaterialUnit.Unit.MultiLanguageName.Uuid == 0 ||
					(stockReconciliation.Status == constant.StockReconciliationStatusSubmitted && itemUnit.Quantity == nil) {
					continue
				}
				unitInfo := &resp.StockReconciliationItemUnitInfo{}
				if err := copier.Copy(unitInfo, itemUnit); err != nil {
					continue
				}
				unitInfo.LocaleName = *language.JsonToLocaleResponse(itemUnit.MaterialUnitName)
				if unitInfo.LocaleName.IsNull() {
					unitInfo.LocaleName = itemUnit.MaterialUnit.Unit.MultiLanguageName.GetNames()
				}
				for _, unit := range itemInfo.Units {
					if unit.MaterialUnitUuid == itemUnit.MaterialUnitUuid {
						unitInfo.ConversionRate = unit.ConversionRate
						break
					}
				}
				unitsResp = append(unitsResp, unitInfo)
			}
			itemInfo.ItemUnits = unitsResp
		}

		bookedQuantity := item.BookedQuantity
		// 已保存状态，账面库存数量要实时读取；其他状态，账面库存数量为盘点单中的数量
		if stockReconciliation.Status == constant.StockReconciliationStatusSaved {
			bookedQuantity = bookedQuantityMap[item.MaterialUuid]
		}
		// 盘盈盘亏状态
		if item.CountedQuantity.GreaterThan(bookedQuantity) {
			itemInfo.InventoryStatus = constant.StockReconciliationInventoryStatusProfit
		} else if item.CountedQuantity.LessThan(bookedQuantity) {
			itemInfo.InventoryStatus = constant.StockReconciliationInventoryStatusLoss
		} else {
			itemInfo.InventoryStatus = constant.StockReconciliationInventoryStatusNormal
		}
		// 是否盘盈盘亏异常（账面和实盘数量差值的绝对值大于20%）
		itemInfo.IsInventoryStatusException = s.getIsInventoryStatusException(bookedQuantity, item.CountedQuantity)
		itemInfo.DiffQuantity = item.CountedQuantity.Sub(bookedQuantity).Truncate(3).InexactFloat64()
		itemsResp = append(itemsResp, itemInfo)
	}
	detailResp.Items = itemsResp

	// 查询批注列表
	annotationRepo := repository.NewStockReconciliationAnnotationRepo(db)
	annotations, err := annotationRepo.GetListByStockReconciliationUuid(req.Uuid)
	if err != nil {
		logger.Logger.Error("查询批注列表失败", zap.Error(err))
	}
	annotationsResp := make([]*resp.StockReconciliationAnnotationInfo, 0, len(annotations))
	for _, annotation := range annotations {
		annotationsResp = append(annotationsResp, &resp.StockReconciliationAnnotationInfo{
			Uuid:           annotation.Uuid,
			AnnotationType: annotation.AnnotationType,
			LocaleName:     constant.GetStockReconciliationAnnotationTypeLocaleName(annotation.AnnotationType),
			Content:        annotation.Content,
			CreateTime:     int64(annotation.CreateTime),
		})
	}
	detailResp.Annotations = annotationsResp

	return detailResp, nil
}

// SaveStockReconciliation 保存盘点单
func (s *stockReconciliationSrv) SaveStockReconciliation(ctx context.Context, saveReq req.StockReconciliationSaveReq) (uint64, error) {

	db := ctx.GetDB()
	companySetting := ctx.GetCompanySetting()
	timezone := companySetting.GetTimezone()
	stockReconciliationUuid := saveReq.Uuid
	var stockReconciliation *model.StockReconciliation
	var err error

	if saveReq.Uuid == 0 { // 新建
		// 加锁保证单号唯一性（基于公司UUID和日期）
		dateStr := utils.SetTimezone(timezone).Now().Format("20060102")
		lockKey := fmt.Sprintf("stock_reconciliation_%d_%s", ctx.GetCompanyUuid(), dateStr)
		s.lock.LockUuidString(lockKey)
		defer s.lock.UnlockUuidString(lockKey)
	} else { // 修改
		// 加锁
		s.lock.LockUuid(saveReq.Uuid)
		defer s.lock.UnlockUuid(saveReq.Uuid)

		stockReconciliationRepo := repository.NewStockReconciliationRepo(db)
		// 查询盘点单
		stockReconciliation, err = stockReconciliationRepo.GetStockReconciliationByUuid(saveReq.Uuid)
		if err != nil {
			return stockReconciliationUuid, errors.WithMessage(err, "查询盘点单失败")
		}
		if stockReconciliation == nil {
			return stockReconciliationUuid, errors.New("盘点单不存在")
		}

		// 重新提交场景：只有已驳回状态才能重新提交
		if saveReq.GetIsResubmit() {
			if stockReconciliation.Status != constant.StockReconciliationStatusRejected {
				return stockReconciliationUuid, errors.New("盘点单状态不允许重新提交")
			}
			// 验证只有提交人才能重新提交
			if stockReconciliation.SubmitterStaffUuid != ctx.GetStaffUuid() {
				return stockReconciliationUuid, errors.New("只有发起人才能重新提交")
			}
		} else {
			// 只有已保存状态的盘点单才能修改
			if stockReconciliation.Status != constant.StockReconciliationStatusSaved {
				if saveReq.IsSubmit {
					return stockReconciliationUuid, errors.New("当前状态不允许提交")
				} else {
					return stockReconciliationUuid, errors.New("当前状态不允许修改")
				}
			}
		}
	}

	// A、在列表上直接提交
	if saveReq.IsSubmit && !saveReq.SubmitAfterSave && saveReq.Uuid > 0 {
		// 直接提交
		return stockReconciliationUuid, s.submitStockReconciliation(ctx, saveReq.Uuid, true)
	}

	// B、在详情中保存或者提交

	// 验证仓库和物品明细
	warehouseItems, materials, err := s.validateWarehouseAndItems(ctx, db, saveReq)
	if err != nil {
		return stockReconciliationUuid, err
	}

	bookedQuantityMap := map[uint64]float64{}
	for _, warehouseItem := range warehouseItems {
		bookedQuantityMap[warehouseItem.MaterialUuid] = warehouseItem.Stock
	}

	materialUnitMap := make(map[uint64]map[uint64]float64)
	materialNameMap := make(map[uint64]string)
	materialUnitNameMap := make(map[uint64]map[uint64]string)
	for _, material := range materials {
		materialUnitMap[material.Uuid] = make(map[uint64]float64)
		materialUnitNameMap[material.Uuid] = make(map[uint64]string)
		for _, materialUnit := range material.NotBaseUnitList {
			materialUnitMap[material.Uuid][materialUnit.Uuid] = materialUnit.ConversionRate
			materialUnitNameMap[material.Uuid][materialUnit.Uuid] = materialUnit.Name
		}
		materialNameMap[material.Uuid] = material.Name
	}

	// 获取 saas 数据库连接
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	if saasDB == nil {
		return stockReconciliationUuid, errors.New("saas 数据库连接失败")
	}

	// 获取公司 UUID（使用总部 UUID 或当前公司 UUID）
	companyUuid := companySetting.HeadquarterUuid
	if companyUuid == 0 {
		companyUuid = ctx.GetCompanyUuid()
	}

	// 开启事务
	err = db.Transaction(func(tx *gorm.DB) error {
		stockReconciliationRepo := repository.NewStockReconciliationRepo(tx)

		// 重新提交成功后添加批注记录
		if saveReq.GetIsResubmit() {
			annotationRepo := repository.NewStockReconciliationAnnotationRepo(db)
			annotation := &model.StockReconciliationAnnotation{
				StockReconciliationUuid: stockReconciliation.Uuid,
				AnnotationType:          constant.StockReconciliationAnnotationTypeResubmit,
			}
			if err := annotationRepo.Create(annotation); err != nil {
				logger.Logger.Error("保存重新提交批注失败", zap.Error(err))
				// 批注保存失败不影响主流程，仅记录日志
			}
		}

		if saveReq.Uuid == 0 { // 新建
			// 生成单据编号
			orderNo, err := s.generateOrderNo(saasDB, companyUuid, timezone)
			if err != nil {
				return errors.WithMessage(err, "生成单据编号失败")
			}
			// 创建盘点单
			stockReconciliation = &model.StockReconciliation{
				OrderNo:            orderNo,
				Type:               saveReq.Type,
				WarehouseUuid:      saveReq.WarehouseUuid,
				Purpose:            saveReq.Purpose,
				Status:             constant.StockReconciliationStatusSaved, // 0-已保存
				SubmitterStaffUuid: ctx.GetStaffUuid(),                      // 记录发起人
				SubmitTime:         int(time.Now().Unix()),                  // 记录提交时间,为了能按照时间排序(最后真正提交时会再更新提交时间)
			}
			if err := stockReconciliationRepo.CreateStockReconciliation(stockReconciliation); err != nil {
				return errors.WithMessage(errors.New("创建盘点单失败"), err.Error())
			}

			stockReconciliationUuid = stockReconciliation.Uuid
		} else { // 更新
			// 更新盘点单
			stockReconciliation.WarehouseUuid = saveReq.WarehouseUuid
			stockReconciliation.Purpose = saveReq.Purpose
			stockReconciliation.Type = saveReq.Type
			stockReconciliation.SubmitTime = int(time.Now().Unix()) // 记录提交时间,为了能按照时间排序(最后真正提交时会再更新提交时间)

			if err := stockReconciliationRepo.UpdateStockReconciliation(stockReconciliation); err != nil {
				return errors.WithMessage(err, "更新盘点单失败")
			}
			// 删除原有的物品明细
			if err := stockReconciliationRepo.DeleteStockReconciliationItemByReconciliationUuid(saveReq.Uuid); err != nil {
				return errors.WithMessage(err, "删除盘点单物品明细失败")
			}
			// 删除原有物品单位明细
			if err := stockReconciliationRepo.DeleteStockReconciliationItemUnitByReconciliationUuid(saveReq.Uuid); err != nil {
				return errors.WithMessage(err, "删除盘点单物品单位明细失败")
			}
		}

		// 步骤1：构建所有物品明细对象，同时保存每个物品对应的单位请求
		var stockReconciliationItems []*model.StockReconciliationItem
		itemUnitsMapping := make([][]*req.StockReconciliationItemUnitReq, 0, len(saveReq.Items))

		for _, reqItem := range saveReq.Items {
			// 计算实盘数量（基准单位）
			countedQuantity := decimal.Zero
			if len(reqItem.Units) > 0 {
				for _, unitItem := range reqItem.Units {
					if unitItem.Quantity == nil {
						continue
					}
					conversionRate := materialUnitMap[reqItem.MaterialUuid][unitItem.MaterialUnitUuid]
					unitQuantity := unitItem.Quantity.Mul(decimal.NewFromFloat(conversionRate))
					countedQuantity = countedQuantity.Add(unitQuantity)
				}
			}
			countedQuantity = countedQuantity.Truncate(3)

			item := &model.StockReconciliationItem{
				StockReconciliationUuid: stockReconciliation.Uuid,
				MaterialUuid:            reqItem.MaterialUuid,
				MaterialName:            materialNameMap[reqItem.MaterialUuid],
				BookedQuantity:          decimal.NewFromFloat(bookedQuantityMap[reqItem.MaterialUuid]),
				CountedQuantity:         countedQuantity,
			}
			stockReconciliationItems = append(stockReconciliationItems, item)
			itemUnitsMapping = append(itemUnitsMapping, reqItem.Units)
		}

		// 步骤2：批量插入物品明细（BeforeCreate 钩子会自动生成 UUID）
		if len(stockReconciliationItems) > 0 {
			if err := stockReconciliationRepo.CreateStockReconciliationItemBatch(stockReconciliationItems); err != nil {
				return errors.WithMessage(errors.New("创建盘点单物品明细失败"), err.Error())
			}
		}

		// 步骤3：使用回填的 UUID 构建单位明细
		var stockReconciliationItemUnits []*model.StockReconciliationItemUnit
		for i, item := range stockReconciliationItems {
			materialUuid := item.MaterialUuid
			for _, unitItem := range itemUnitsMapping[i] {
				var quantity *float64
				if unitItem.Quantity != nil {
					quantityDecimal := unitItem.Quantity.InexactFloat64()
					quantity = &quantityDecimal
				}
				stockReconciliationItemUnits = append(stockReconciliationItemUnits, &model.StockReconciliationItemUnit{
					StockReconciliationItemUuid: item.Uuid,
					MaterialUnitUuid:            unitItem.MaterialUnitUuid,
					MaterialUnitName:            materialUnitNameMap[materialUuid][unitItem.MaterialUnitUuid],
					Quantity:                    quantity,
				})
			}
		}

		// 步骤4：批量插入单位明细
		if len(stockReconciliationItemUnits) > 0 {
			if err := stockReconciliationRepo.CreateStockReconciliationItemUnitBatch(stockReconciliationItemUnits); err != nil {
				return errors.WithMessage(errors.New("创建盘点单物品单位明细失败"), err.Error())
			}
		}

		return nil
	})

	if err != nil {
		errMsg := "保存失败"
		if saveReq.IsSubmit || saveReq.GetIsResubmit() {
			errMsg = "提交失败"
		}
		return stockReconciliationUuid, errors.WithMessage(errors.New(errMsg), err.Error())
	}

	// 提交盘点单（包括首次提交和重新提交）
	if (saveReq.IsSubmit || saveReq.GetIsResubmit()) && ctx.GetCompany().IsOpenErp() {
		err = s.submitStockReconciliation(ctx, stockReconciliation.Uuid, false)
		if err != nil {
			return stockReconciliationUuid, errors.WithMessage(err, "提交盘点单失败")
		}
	}

	return stockReconciliationUuid, nil
}

// getWarehouseMaterialUuidMap 获取仓库物品UUID映射
func (s *stockReconciliationSrv) getWarehouseMaterialUuidMap(db *gorm.DB, warehouseUuid uint64) (map[uint64]bool, error) {
	warehouseMaterialUUidMap := make(map[uint64]bool)
	warehouseItemRepo := repository.NewWarehouseItemRepo(db)
	warehouseItems, err := warehouseItemRepo.GetByWarehouseUuid(warehouseUuid)
	if err != nil {
		return nil, errors.WithMessage(errors.New(errMsgQueryWarehouseItem), err.Error())
	}
	for _, item := range warehouseItems {
		warehouseMaterialUUidMap[item.MaterialUuid] = true
	}
	return warehouseMaterialUUidMap, nil
}

// 提交盘点单
// stockReconciliationUuid: 盘点单UUID
// isDirectSubmit: 是否列表上直接提交，true表示在列表上点击提交，false表示保存后提交
func (s *stockReconciliationSrv) submitStockReconciliation(ctx context.Context, stockReconciliationUuid uint64, isDirectSubmit bool) error {
	db := ctx.GetDB()
	stockReconciliationRepo := repository.NewStockReconciliationRepo(db)

	opts := []repository.DBOption{
		stockReconciliationRepo.WhereUuid(stockReconciliationUuid),
		stockReconciliationRepo.WithStockReconciliationItemsMultiLanguageName(),
		stockReconciliationRepo.WithStockReconciliationItemsUnits(),
		stockReconciliationRepo.WithWarehouse(),
	}
	stockReconciliation, err := stockReconciliationRepo.GetStockReconciliation(opts...)

	if err != nil {
		return errors.WithMessage(errors.New("查询盘点单失败"), err.Error())
	}
	if stockReconciliation == nil {
		return errors.New("盘点单不存在")
	}

	if ctx.Version(context.GTE, constant.ClientVersionV2100) && stockReconciliation.Warehouse != nil && stockReconciliation.Warehouse.IsDisabled() {
		return errors.NewWithCode(constant.CodeWarehouseDisabled, i18n.Translate(ctx.GetLanguage(), "仓库状态已关闭，请修改仓库状态"))
	}

	bookedQuantityMap := make(map[uint64]decimal.Decimal)
	if isDirectSubmit {
		var err error
		bookedQuantityMap, err = s.getBookedQuantityMap(db, stockReconciliation.WarehouseUuid)
		if err != nil {
			return errors.WithMessage(errors.New(errMsgQueryWarehouseItem), err.Error())
		}
	}

	warehouseMaterialUUidMap, err := s.getWarehouseMaterialUuidMap(db, stockReconciliation.WarehouseUuid)
	if err != nil {
		return errors.WithMessage(errors.New(errMsgQueryWarehouseItem), err.Error())
	}

	lang := ctx.GetLanguage()
	existsMaterialUuidMap := make(map[uint64]bool)
	for _, reqItem := range stockReconciliation.StockReconciliationItems {
		if reqItem.DeleteTime > 0 {
			continue
		}
		if _, exists := existsMaterialUuidMap[reqItem.MaterialUuid]; exists {
			materialName := *language.JsonToLocaleResponse(reqItem.MaterialName)
			return fmt.Errorf(i18n.Translate(lang, "物品 %s 重复"), materialName.GetLocale(lang))
		}
		existsMaterialUuidMap[reqItem.MaterialUuid] = true
	}

	companySetting := ctx.GetCompanySetting()

	// 获取业务设置，检查盘点允许估值率为0的开关
	settingSrv := setting.NewSrv(s.dbm, cache.Global)
	businessSetting, bsErr := settingSrv.GetBusinessSetting(ctx)
	if bsErr != nil {
		logger.Logger.Error("获取业务设置失败", zap.Error(bsErr), zap.Uint64("company_uuid", ctx.GetCompanyUuid()))
	}

	// 如果关闭了"盘点允许估值率为0"，需要校验物品估值率
	if !businessSetting.IsAllowZeroValuationRate() && stockReconciliation.Warehouse != nil {
		erpSrv := erp.NewIErpSrv(s.dbm)
		binItems, binErr := erpSrv.GetMaterialStockNumByBin(ctx, stockReconciliation.Warehouse.ErpCode)
		if binErr != nil {
			logger.Logger.Error("查询Bin记录失败", zap.Error(binErr), zap.Uint64("company_uuid", ctx.GetCompanyUuid()))
			return errors.WithMessage(errors.New(i18n.Translate(lang, "查询物品估值率失败")), binErr.Error())
		}

		// 构建估值率映射
		binValuationMap := make(map[string]float64)
		for _, bin := range binItems {
			binValuationMap[bin.ItemCode] = bin.ValuationRate
		}

		// 检查盘点单中估值率为0的物品
		var zeroValuationItems []string
		for _, item := range stockReconciliation.StockReconciliationItems {
			if item.DeleteTime > 0 || !item.Material.Status {
				continue
			}
			rate, hasBin := binValuationMap[item.Material.Code]
			if !hasBin || rate == 0 {
				materialName := *language.JsonToLocaleResponse(item.MaterialName)
				name := materialName.GetLocale(lang)
				if item.Material.InternalCode != "" {
					name = fmt.Sprintf("%s（%s）", name, item.Material.InternalCode)
				}
				zeroValuationItems = append(zeroValuationItems, name)
			}
		}

		if len(zeroValuationItems) > 0 {
			itemsStr := strings.Join(zeroValuationItems, "、")
			return fmt.Errorf(i18n.Translate(lang, "物品%s估值率为0，无法提交。（请联系管理员进行处理）"), itemsStr)
		}
	}

	// 根据时区获取过账日期和时间
	now := utils.SetTimezone(companySetting.GetTimezone()).Now()

	var erpItems []*stock.StockReconciliationItem
	err = db.Transaction(func(tx *gorm.DB) error {
		stockReconciliationRepo := repository.NewStockReconciliationRepo(tx)
		for _, item := range stockReconciliation.StockReconciliationItems {
			if !warehouseMaterialUUidMap[item.MaterialUuid] {
				return errors.New("盘点单中有不在此仓库的物品")
			}
			// 物品已禁用，标记item的delete_time(删除)
			if !item.Material.Status {
				if err := stockReconciliationRepo.DeleteStockReconciliationItem(item.Uuid); err != nil {
					return errors.WithMessage(errors.New("提交盘点单时移除已关闭物品失败"), err.Error())
				}
				continue
			}
			// 不往erp传递已删除的物品明细
			if item.DeleteTime > 0 {
				continue
			}

			var unitExists bool
			for _, unit := range item.StockReconciliationItemUnits {
				if unit.DeleteTime == 0 && unit.Quantity != nil {
					unitExists = true
					break
				}
			}
			if !unitExists {
				if err := stockReconciliationRepo.DeleteStockReconciliationItem(item.Uuid); err != nil {
					return errors.WithMessage(errors.New("提交盘点单时移除待盘点物品失败"), err.Error())
				}
				continue
			}

			if isDirectSubmit {
				stockReconciliationItem := *item
				stockReconciliationItem.BookedQuantity = bookedQuantityMap[item.MaterialUuid]
				if err := stockReconciliationRepo.UpdateStockReconciliationItem(&stockReconciliationItem); err != nil {
					return errors.WithMessage(errors.New("更新盘点单物品明细失败"), err.Error())
				}
			}

			erpItems = append(erpItems, &stock.StockReconciliationItem{
				ItemCode: item.Material.Code,
				Qty:      item.CountedQuantity.InexactFloat64(),
			})
		}

		if len(erpItems) == 0 {
			return errors.New("物品列表为空，请先添加物品后再操作")
		}
		erpSrv := erp.NewIErpSrv(s.dbm)
		erpReq, err := erpSrv.SubmitStockReconciliation(ctx, companySetting, &stock.SaveStockReconciliationReq{
			CompanyAbbr:   companySetting.ErpnextCompanyAbbr,
			Branch:        companySetting.ErpnextBranchName,
			PostingDate:   now.Format("2006-01-02"),
			PostingTime:   now.Format("15:04:05"),
			Warehouse:     stockReconciliation.Warehouse.ErpCode,
			Items:         erpItems,
			InventoryType: constant.StockReconciliationTypeToErpInventoryType(stockReconciliation.Type),
			Purpose:       constant.StockReconciliationPurposeToErp(stockReconciliation.Purpose),
		})
		if err != nil {
			logger.Logger.Error("提交盘点单失败", zap.Error(err))
			// 检查是否是仓库禁用错误
			if ctx.Version(context.GTE, constant.ClientVersionV2100) && strings.Contains(err.Error(), "Disabled Warehouse") {
				return errors.NewWithCode(constant.CodeWarehouseDisabled, i18n.Translate(ctx.GetLanguage(), "仓库状态已关闭，请修改仓库状态"))
			}
			// 提取物品名称
			itemName := s.extractName("Item", "is disabled", err.Error())
			for _, item := range stockReconciliation.StockReconciliationItems {
				if item.Material.Code == itemName {
					materialName := item.Material.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
					if ctx.Version(context.GTE, constant.ClientVersionV2100) {
						return errors.NewWithCode(constant.CodeItemDisabled, materialName)
					}
					message := i18n.Translate(ctx.GetLanguage(), errMsgItemDisabledFmt, materialName)
					return errors.New(message)
				}
			}
			if itemName != "" {
				if ctx.Version(context.GTE, constant.ClientVersionV2100) {
					return errors.NewWithCode(constant.CodeItemDisabled, itemName)
				}
				message := i18n.Translate(ctx.GetLanguage(), errMsgItemDisabledFmt, itemName)
				return errors.New(message)
			}
			return errors.WithMessage(errors.New("提交盘点单失败"), err.Error())
		}
		// 更新盘点单erp_code和提交时间
		stockReconciliation.ErpCode = erpReq.StockReconciliationName
		stockReconciliation.SubmitTime = int(time.Now().Unix())
		stockReconciliation.Status = constant.StockReconciliationStatusSubmitted
		stockReconciliation.SubmitterStaffUuid = ctx.GetStaffUuid() // 记录提交人
		if err := stockReconciliationRepo.UpdateStockReconciliation(stockReconciliation); err != nil {
			return errors.WithMessage(errors.New("更新盘点单状态失败"), err.Error())
		}

		return nil
	})
	if err != nil {
		return errors.WithMessage(err)
	}

	return nil
}

// DeleteStockReconciliation 删除盘点单
func (s *stockReconciliationSrv) DeleteStockReconciliation(ctx context.Context, req req.StockReconciliationDeleteReq) error {
	db := ctx.GetDB()

	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	stockReconciliationRepo := repository.NewStockReconciliationRepo(db)

	// 查询盘点单
	stockReconciliation, err := stockReconciliationRepo.GetStockReconciliationByUuid(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "查询盘点单失败")
	}
	if stockReconciliation == nil {
		return errors.New("盘点单不存在")
	}

	// 只有已保存和已驳回状态的盘点单才能删除
	if stockReconciliation.Status != constant.StockReconciliationStatusSaved {
		return errors.New("当前状态不允许删除")
	}

	// 删除盘点单以及物品明细和单位明细
	if err := stockReconciliationRepo.DeleteStockReconciliation(req.Uuid); err != nil {
		return errors.WithMessage(err, "删除盘点单失败")
	}

	// 清理UUID锁资源
	s.lock.ClearUuidLock(req.Uuid)

	return nil
}

// ApproveStockReconciliation 审核盘点单
func (s *stockReconciliationSrv) ApproveStockReconciliation(ctx context.Context, req req.StockReconciliationApproveReq) ([]dto.LocaleResponse, error) {
	db := ctx.GetDB()

	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	stockReconciliation, err := s.loadAndValidateStockReconciliation(ctx, db, req.Uuid)
	if err != nil {
		return nil, err
	}

	// 检查物品是否被禁用
	if disabledMaterials := s.checkDisabledMaterials(stockReconciliation); len(disabledMaterials) > 0 {
		return disabledMaterials, errors.New("请修改物品状态")
	}

	// 获取仓库已有物品集合
	warehouseMaterialUuids, err := s.getWarehouseMaterialSet(db, stockReconciliation.WarehouseUuid)
	if err != nil {
		return nil, err
	}

	// 开启事务
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := s.updateStatusAndAnnotation(tx, req.Uuid, req.Annotation); err != nil {
			return err
		}
		if err := s.processStockReconciliationItems(tx, stockReconciliation, warehouseMaterialUuids); err != nil {
			return err
		}
		return s.approveStockReconciliationInERP(ctx, stockReconciliation)
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// loadAndValidateStockReconciliation 加载并校验盘点单
func (s *stockReconciliationSrv) loadAndValidateStockReconciliation(ctx context.Context, db *gorm.DB, uuid uint64) (*model.StockReconciliation, error) {
	stockReconciliationRepo := repository.NewStockReconciliationRepo(db)
	stockReconciliation, err := stockReconciliationRepo.GetStockReconciliation(
		stockReconciliationRepo.WhereUuid(uuid),
		stockReconciliationRepo.WithStockReconciliationItemsMultiLanguageName(),
		stockReconciliationRepo.WithStockReconciliationItemsMaterialBaseUnit(),
		stockReconciliationRepo.WithWarehouse(),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "查询盘点单失败")
	}
	if stockReconciliation == nil {
		return nil, errors.New("盘点单不存在")
	}
	if stockReconciliation.Status != constant.StockReconciliationStatusSubmitted {
		return nil, errors.New("当前状态不允许审核")
	}
	if ctx.Version(context.GTE, constant.ClientVersionV2100) && stockReconciliation.Warehouse != nil && stockReconciliation.Warehouse.IsDisabled() {
		return nil, errors.NewWithCode(constant.CodeWarehouseDisabled, i18n.Translate(ctx.GetLanguage(), "仓库状态已关闭，请修改仓库状态"))
	}
	return stockReconciliation, nil
}

// checkDisabledMaterials 检查盘点单物品中是否有被禁用的
func (s *stockReconciliationSrv) checkDisabledMaterials(sr *model.StockReconciliation) []dto.LocaleResponse {
	var disabledMaterials []dto.LocaleResponse
	for _, item := range sr.StockReconciliationItems {
		if item.DeleteTime > 0 {
			continue
		}
		if !item.Material.Status {
			disabledMaterials = append(disabledMaterials, item.Material.MultiLanguageName.GetNames())
		}
	}
	return disabledMaterials
}

// getWarehouseMaterialSet 获取仓库已有物品的 UUID 集合
func (s *stockReconciliationSrv) getWarehouseMaterialSet(db *gorm.DB, warehouseUuid uint64) (map[uint64]struct{}, error) {
	warehouseItems, err := repository.NewWarehouseItemRepo(db).GetByWarehouseUuid(warehouseUuid)
	if err != nil {
		return nil, errors.WithMessage(errors.New(errMsgQueryWarehouseItem), err.Error())
	}
	m := make(map[uint64]struct{})
	for _, item := range warehouseItems {
		m[item.MaterialUuid] = struct{}{}
	}
	return m, nil
}

// updateStatusAndAnnotation 更新盘点单状态为已审核并保存批注
func (s *stockReconciliationSrv) updateStatusAndAnnotation(tx *gorm.DB, uuid uint64, annotation string) error {
	stockReconciliationRepo := repository.NewStockReconciliationRepo(tx)
	updateData := map[string]any{
		"status":      constant.StockReconciliationStatusApproved,
		"update_time": int(time.Now().Unix()),
	}
	if err := stockReconciliationRepo.UpdateStockReconciliationData(updateData, stockReconciliationRepo.WhereUuid(uuid)); err != nil {
		return errors.WithMessage(errors.New(errMsgApproveReconcilation), err.Error())
	}

	return errors.WithMessage(
		repository.NewStockReconciliationAnnotationRepo(tx).Create(&model.StockReconciliationAnnotation{
			StockReconciliationUuid: uuid,
			AnnotationType:          constant.StockReconciliationAnnotationTypeApprove,
			Content:                 annotation,
		}),
		"保存批注失败",
	)
}

// processStockReconciliationItems 处理盘点单物品：更新库存并生成盈亏日志
func (s *stockReconciliationSrv) processStockReconciliationItems(tx *gorm.DB, sr *model.StockReconciliation, warehouseMaterialUuids map[uint64]struct{}) error {
	stockReconciliationRepo := repository.NewStockReconciliationRepo(tx)
	txWarehouseItemRepo := repository.NewWarehouseItemRepo(tx)

	var warehouseLogs []*model.WarehouseInOutLog
	for _, item := range sr.StockReconciliationItems {
		if item.DeleteTime > 0 {
			continue
		}
		if item.Material.DeleteTime > 0 {
			if err := stockReconciliationRepo.DeleteStockReconciliationItem(item.Uuid); err != nil {
				return errors.WithMessage(errors.New("审核盘点单时移除已删除物品失败"), err.Error())
			}
			continue
		}

		if err := s.upsertWarehouseItemStock(txWarehouseItemRepo, sr.WarehouseUuid, item, warehouseMaterialUuids); err != nil {
			return err
		}

		if log := s.buildProfitLossLog(sr, item); log != nil {
			warehouseLogs = append(warehouseLogs, log)
		}
	}

	if len(warehouseLogs) > 0 {
		if err := repository.NewWarehouseInOutLogRepo(tx).CreateBatch(warehouseLogs); err != nil {
			return errors.WithMessage(errors.New("创建盘盈盘亏出入库记录失败"), err.Error())
		}
	}
	return nil
}

// upsertWarehouseItemStock 新增或更新仓库物品库存为实盘数量
func (s *stockReconciliationSrv) upsertWarehouseItemStock(repo repository.IWarehouseItemRepo, warehouseUuid uint64, item *model.StockReconciliationItem, existingMaterials map[uint64]struct{}) error {
	stockQuantity := item.CountedQuantity.Truncate(3).InexactFloat64()
	if _, exists := existingMaterials[item.MaterialUuid]; !exists {
		if err := repo.Create(&model.WarehouseItem{
			WarehouseUuid: warehouseUuid,
			MaterialUuid:  item.MaterialUuid,
			MaterialCode:  item.Material.Code,
			Stock:         stockQuantity,
			Valuation:     1.0,
		}); err != nil {
			return errors.WithMessage(errors.New("创建仓库物品失败"), err.Error())
		}
		return nil
	}
	if err := repo.UpdateStockByWarehouseAndMaterial(warehouseUuid, item.MaterialUuid, stockQuantity); err != nil {
		return errors.WithMessage(errors.New("更新仓库物品库存失败"), err.Error())
	}
	return nil
}

// buildProfitLossLog 构建盘盈盘亏出入库日志（无差异则返回 nil）
func (s *stockReconciliationSrv) buildProfitLossLog(sr *model.StockReconciliation, item *model.StockReconciliationItem) *model.WarehouseInOutLog {
	if item.CountedQuantity.Equal(item.BookedQuantity) {
		return nil
	}
	scene := constant.WarehouseInOutLogSceneProfitIn
	logType := constant.WarehouseInOutLogLogTypeIn
	if item.CountedQuantity.LessThan(item.BookedQuantity) {
		logType = constant.WarehouseInOutLogLogTypeOut
		scene = constant.WarehouseInOutLogSceneLossOut
	}
	diff := item.CountedQuantity.Sub(item.BookedQuantity).Abs()
	valuation := 0.0
	return &model.WarehouseInOutLog{
		LogType:              logType,
		Scene:                scene,
		WarehouseUuid:        sr.WarehouseUuid,
		MaterialUuid:         item.MaterialUuid,
		MaterialName:         item.MaterialName,
		MaterialBaseUnitUuid: item.Material.Unit.Uuid,
		MaterialBaseUnitName: item.Material.Unit.Name,
		Num:                  diff.Truncate(3).InexactFloat64(),
		Price:                valuation,
		Amount:               decimal.NewFromFloat(valuation).Mul(diff).Truncate(3).InexactFloat64(),
		OrderNo:              sr.OrderNo,
	}
}

// approveStockReconciliationInERP 调用 ERP 接口审核盘点单
func (s *stockReconciliationSrv) approveStockReconciliationInERP(ctx context.Context, sr *model.StockReconciliation) error {
	if !ctx.GetCompany().IsOpenErp() || sr.ErpCode == "" {
		return nil
	}

	companySetting := ctx.GetCompanySetting()
	_, err := erp.NewIErpSrv(s.dbm).ApproveStockReconciliation(ctx, companySetting, &stock.SubmitStockReconciliationReq{
		StockReconciliationName: sr.ErpCode,
	})
	if err == nil {
		return nil
	}

	logger.Logger.Error(errMsgApproveReconcilation, zap.Error(err))
	return s.handleErpApproveError(ctx, sr, err)
}

// handleErpApproveError 处理 ERP 审核盘点单失败的错误
func (s *stockReconciliationSrv) handleErpApproveError(ctx context.Context, sr *model.StockReconciliation, erpErr error) error {
	errMsg := erpErr.Error()
	isV2100 := ctx.Version(context.GTE, constant.ClientVersionV2100)

	if isV2100 && strings.Contains(errMsg, "Disabled Warehouse") {
		return errors.NewWithCode(constant.CodeWarehouseDisabled, i18n.Translate(ctx.GetLanguage(), "仓库状态已关闭，请修改仓库状态"))
	}

	itemName := s.extractName("Item", "is disabled", errMsg)
	if itemName == "" {
		return errors.WithMessage(errors.New(errMsgApproveReconcilation), errMsg)
	}

	// 在盘点单物品中找到对应的多语言名称
	for _, item := range sr.StockReconciliationItems {
		if item.Material.Code == itemName {
			itemName = item.Material.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
			break
		}
	}

	if isV2100 {
		return errors.NewWithCode(constant.CodeItemDisabled, itemName)
	}
	return errors.New(i18n.Translate(ctx.GetLanguage(), errMsgItemDisabledFmt, itemName))
}

// RejectStockReconciliation 驳回盘点单
func (s *stockReconciliationSrv) RejectStockReconciliation(ctx context.Context, req req.StockReconciliationRejectReq) error {
	db := ctx.GetDB()

	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	stockReconciliationRepo := repository.NewStockReconciliationRepo(db)

	// 查询盘点单
	stockReconciliation, err := stockReconciliationRepo.GetStockReconciliationByUuid(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "查询盘点单失败")
	}
	if stockReconciliation == nil {
		return errors.New("盘点单不存在")
	}

	// 只有已提交状态的盘点单才能驳回
	if stockReconciliation.Status != constant.StockReconciliationStatusSubmitted {
		return errors.New("盘点单状态不允许驳回")
	}

	// 开启事务
	err = db.Transaction(func(tx *gorm.DB) error {
		stockReconciliationRepo := repository.NewStockReconciliationRepo(tx)

		// 更新盘点单状态为已驳回
		updateData := map[string]any{
			"status":      constant.StockReconciliationStatusRejected, // 3-已驳回
			"update_time": int(time.Now().Unix()),
		}
		if err := stockReconciliationRepo.UpdateStockReconciliationData(updateData, stockReconciliationRepo.WhereUuid(req.Uuid)); err != nil {
			return errors.WithMessage(errors.New("驳回盘点单失败"), err.Error())
		}

		// 保存批注记录（驳回必须创建批注记录）
		annotationRepo := repository.NewStockReconciliationAnnotationRepo(tx)
		annotation := &model.StockReconciliationAnnotation{
			StockReconciliationUuid: req.Uuid,
			AnnotationType:          constant.StockReconciliationAnnotationTypeReject,
			Content:                 req.Annotation,
		}
		if err := annotationRepo.Create(annotation); err != nil {
			return errors.WithMessage(err, "保存批注失败")
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// generateOrderNo 生成盘点单编号
// 格式：ST + yyyyMMddHHmmss + 序列号（4位）
// 例如：ST202504030915120001
func (s *stockReconciliationSrv) generateOrderNo(
	saasDB *gorm.DB,
	companyUuid uint64,
	timezone string,
) (string, error) {
	// 获取秒级时间戳
	now := utils.SetTimezone(timezone).Now()
	timestamp := now.Format("20060102150405") // yyyyMMddHHmmss

	// 获取日期字符串（用于序列号表）
	dateStr := now.Format("2006-01-02") // YYYY-MM-DD

	// 从 ttpos_number_sequence 表获取下一个序列号
	seqRepo := repository.NewNumberSequenceRepo(saasDB)
	seq, err := seqRepo.GetNextSequence(companyUuid, constant.NumberTypeStockTake, dateStr)
	if err != nil {
		return "", errors.WithMessage(err, "获取序列号失败")
	}

	// 组装编号：ST + timestamp + 序列号（4位）
	orderNo := fmt.Sprintf("ST%s%04d", timestamp, seq)
	return orderNo, nil
}

// validateWarehouseAndItems 验证仓库和物品明细
// 验证内容：
// 1. 仓库是否存在且类型为 normal
// 2. 物品是否属于该仓库
// 3. 物品单位是否正确
func (s *stockReconciliationSrv) validateWarehouseAndItems(ctx context.Context, db *gorm.DB, req req.StockReconciliationSaveReq) ([]model.WarehouseItem, []*model.Material, error) {
	warehouseUuid := req.WarehouseUuid
	if warehouseUuid == 0 {
		return nil, nil, errors.New("仓库参数错误")
	}
	if req.Purpose != 1 && req.Purpose != 2 {
		return nil, nil, errors.New("盘点目的参数错误")
	}
	// 判断仓库是否存在，且类型为normal，且未被禁用
	warehouseRepo := repository.NewWarehouseRepo(db)
	warehouse, err := warehouseRepo.GetByUuid(warehouseUuid)
	if err != nil {
		return nil, nil, errors.WithMessage(errors.New("查询仓库失败"), err.Error())
	}
	if warehouse == nil || warehouse.Type != constant.WarehouseTypeNormal {
		return nil, nil, errors.New("仓库参数错误")
	}

	// 判断盘点物品明细列表是否正确，要求所有物品均为仓库内的物品，且单位均为仓库内物品的单位
	warehouseItemRepo := repository.NewWarehouseItemRepo(db)
	// 获取仓库Uuid获取仓库物品信息列表
	warehouseItems, err := warehouseItemRepo.GetByWarehouseUuid(warehouseUuid)
	if err != nil {
		return nil, nil, errors.WithMessage(errors.New(errMsgQueryWarehouseItem), err.Error())
	}

	// 获取所有物品UUID
	materialUuids := make([]uint64, 0, len(req.Items))
	for _, item := range req.Items {
		materialUuids = append(materialUuids, item.MaterialUuid)
	}

	// 批量查询物品详情
	materialRepo := repository.NewMaterialRepo(db)
	var materials []*model.Material
	materialMap := make(map[uint64]*model.Material)
	if len(materialUuids) > 0 {
		materials, err = materialRepo.GetMaterialDetailByUuids(materialUuids)
		if err != nil {
			return nil, nil, errors.WithMessage(errors.New("查询物品详情失败"), err.Error())
		}
		for _, material := range materials {
			materialMap[material.Uuid] = material
		}
	}

	language := ctx.GetLanguage()
	existsMaterialUuidMap := make(map[uint64]bool)
	for _, reqItem := range req.Items {
		if _, exists := existsMaterialUuidMap[reqItem.MaterialUuid]; exists {
			return nil, nil, fmt.Errorf(i18n.Translate(language, "物品 %s 重复"), materialMap[reqItem.MaterialUuid].MultiLanguageName.GetNameByLang(language))
		}
		existsMaterialUuidMap[reqItem.MaterialUuid] = true
	}

	// 验证请求中的物品和单位
	for _, item := range req.Items {
		material, exists := materialMap[item.MaterialUuid]
		if !exists {
			return nil, nil, errors.New("物品参数错误")
		}

		unitExists := false
		// 验证物品单位
		for _, unit := range item.Units {
			// 检查单位列表
			for _, materialUnit := range material.NotBaseUnitList {
				if unit.MaterialUnitUuid == materialUnit.Uuid && unit.Quantity != nil {
					unitExists = true
					break
				}
			}
		}
		if !unitExists && req.IsSubmit {
			return nil, nil, errors.New("物品单位参数错误")
		}
	}

	return warehouseItems, materials, nil
}

// getIsInventoryStatusException 获取是否盘盈盘亏异常
func (s *stockReconciliationSrv) getIsInventoryStatusException(bookedQuantity decimal.Decimal, countedQuantity decimal.Decimal) bool {
	if bookedQuantity.IsZero() {
		if countedQuantity.IsZero() {
			return false
		}
		return true
	}
	return countedQuantity.Sub(bookedQuantity).Abs().Div(bookedQuantity).GreaterThan(decimal.NewFromFloat(0.3)) // v2.12.0 把盈亏异常的计算由20%调整为30%
}

func (s *stockReconciliationSrv) CheckMaterials(ctx context.Context, checkReq req.StockReconciliationCheckMaterialsReq) (resp.StockReconciliationCheckMaterialsListResp, error) {

	var listResp resp.StockReconciliationCheckMaterialsListResp
	itemResp := make([]resp.StockReconciliationCheckMaterialsResp, 0)

	db := ctx.GetDB()

	// 获取业务设置，检查盘点允许估值率为0的开关
	settingSrv := setting.NewSrv(s.dbm, cache.Global)
	businessSetting, bsErr := settingSrv.GetBusinessSetting(ctx)
	if bsErr != nil {
		logger.Logger.Error("获取业务设置失败", zap.Error(bsErr), zap.Uint64("company_uuid", ctx.GetCompanyUuid()))
	}
	allowZeroValuationRate := businessSetting.IsAllowZeroValuationRate()

	// 查询 Bin 记录用于判断估值率
	binValuationMap := make(map[string]float64)
	if checkReq.WarehouseUuid != 0 {
		warehouseRepo := repository.NewWarehouseRepo(db)
		warehouse, whErr := warehouseRepo.GetByUuid(checkReq.WarehouseUuid)
		if whErr == nil && warehouse != nil && warehouse.ErpCode != "" {
			erpSrv := erp.NewIErpSrv(s.dbm)
			binItems, binErr := erpSrv.GetMaterialStockNumByBin(ctx, warehouse.ErpCode)
			if binErr != nil {
				logger.Logger.Error("查询Bin记录失败", zap.Error(binErr), zap.Uint64("company_uuid", ctx.GetCompanyUuid()))
			} else {
				for _, bin := range binItems {
					binValuationMap[bin.ItemCode] = bin.ValuationRate
				}
			}
		}
	}

	var materialUuids []uint64

	warehouseMaterialUUidMap, err := s.getWarehouseMaterialUuidMap(db, checkReq.WarehouseUuid)
	if err != nil {
		return listResp, errors.WithMessage(errors.New(errMsgQueryWarehouseItem), err.Error())
	}

	bookedQuantityMap := make(map[uint64]decimal.Decimal)

	if checkReq.Uuid != 0 {
		stockReconciliationRepo := repository.NewStockReconciliationRepo(db)
		opts := []repository.DBOption{
			stockReconciliationRepo.WhereUuid(checkReq.Uuid),
			stockReconciliationRepo.WithStockReconciliationItemsMultiLanguageName(),
			stockReconciliationRepo.WithStockReconciliationItemsUnits(),
		}
		// 查询盘点单
		stockReconciliation, err := stockReconciliationRepo.GetStockReconciliation(opts...)
		if err != nil {
			return listResp, errors.WithMessage(err, "查询盘点单失败")
		}
		if stockReconciliation == nil {
			return listResp, errors.New("盘点单不存在")
		}

		bookedQuantityMap, err = s.getBookedQuantityMap(db, stockReconciliation.WarehouseUuid)
		if err != nil {
			return listResp, errors.WithMessage(errors.New(errMsgQueryWarehouseItem), err.Error())
		}

		limitedMaterialUuids := make([]uint64, 0)
		itemCountedQuantityMap := make(map[uint64]decimal.Decimal)
		for _, item := range checkReq.Items {
			limitedMaterialUuids = append(limitedMaterialUuids, item.MaterialUuid)
			itemCountedQuantityMap[item.MaterialUuid] = item.CountedQuantity
		}
		for _, item := range stockReconciliation.StockReconciliationItems {
			if item.DeleteTime > 0 || (len(limitedMaterialUuids) > 0 && !slices.Contains(limitedMaterialUuids, item.MaterialUuid)) {
				continue
			}
			materialUuids = append(materialUuids, item.MaterialUuid)
			bookedQuantity := item.BookedQuantity
			// 已保存状态，账面库存数量要实时读取；其他状态，账面库存数量为盘点单中的数量
			if stockReconciliation.Status == constant.StockReconciliationStatusSaved {
				bookedQuantity = bookedQuantityMap[item.MaterialUuid]
			}

			// 优先使用前端传入的实盘数量（用户可能修改了草稿中的盘点数量但尚未保存）
			countedQuantity := item.CountedQuantity
			if qty, ok := itemCountedQuantityMap[item.MaterialUuid]; ok {
				countedQuantity = qty
			}

			var unitCount uint
			for _, unit := range item.StockReconciliationItemUnits {
				if unit.Quantity != nil {
					unitCount++
				}
			}

			itemResp = append(itemResp, resp.StockReconciliationCheckMaterialsResp{
				LocaleName:                 item.Material.MultiLanguageName.GetNames(),
				IsInventoryStatusException: unitCount > 0 && s.getIsInventoryStatusException(bookedQuantity, countedQuantity),
				Status:                     item.Material.Status,
				IsDeleted:                  item.Material.DeleteTime > 0,
				UnitCount:                  unitCount,
				ExistsInWarehouse:          warehouseMaterialUUidMap[item.MaterialUuid],
				IsZeroValuationRate:        s.isZeroValuationRate(binValuationMap, item.Material.Code),
				InternalCode:               item.Material.InternalCode,
			})
		}
	}

	// 没传递盘点单UUID，则根据仓库UUID查询仓库物品
	if checkReq.WarehouseUuid != 0 && checkReq.Uuid == 0 {
		var err error
		bookedQuantityMap, err = s.getBookedQuantityMap(db, checkReq.WarehouseUuid)
		if err != nil {
			return listResp, errors.WithMessage(errors.New(errMsgQueryWarehouseItem), err.Error())
		}
	}
	var warehouseDisabled bool
	if checkReq.WarehouseUuid != 0 && ctx.Version(context.GTE, constant.ClientVersionV2100) {
		warehouseRepo := repository.NewWarehouseRepo(db)
		warehouse, err := warehouseRepo.GetByUuid(checkReq.WarehouseUuid)
		if err != nil {
			return listResp, errors.WithMessage(errors.New("查询仓库失败"), err.Error())
		}
		warehouseDisabled = warehouse != nil && warehouse.IsDisabled()
	}

	if len(checkReq.Items) > 0 {
		var newMaterialUuids []uint64
		itemMap := make(map[uint64]req.CheckMaterialsItem)
		// 过滤掉在materialUuids中的物品
		for _, item := range checkReq.Items {
			if !slices.Contains(materialUuids, item.MaterialUuid) {
				newMaterialUuids = append(newMaterialUuids, item.MaterialUuid)
			}
			itemMap[item.MaterialUuid] = item
		}

		materialRepo := repository.NewMaterialRepo(db)
		materialPtrs, _ := materialRepo.GetMaterialByUuids(newMaterialUuids, materialRepo.WithMultiLanguageName())
		materials := make([]model.Material, 0, len(materialPtrs))
		for _, m := range materialPtrs {
			materials = append(materials, *m)
		}

		for _, material := range materials {
			countedQuantity := itemMap[material.Uuid].CountedQuantity
			itemResp = append(itemResp, resp.StockReconciliationCheckMaterialsResp{
				LocaleName:                 material.MultiLanguageName.GetNames(),
				Status:                     material.Status,
				IsDeleted:                  material.DeleteTime > 0,
				IsInventoryStatusException: s.getIsInventoryStatusException(bookedQuantityMap[material.Uuid], countedQuantity),
				ExistsInWarehouse:          warehouseMaterialUUidMap[material.Uuid],
				IsZeroValuationRate:        s.isZeroValuationRate(binValuationMap, material.Code),
				InternalCode:               material.InternalCode,
			})
		}
	}

	return resp.StockReconciliationCheckMaterialsListResp{
		List:                   itemResp,
		WarehouseDisabled:      warehouseDisabled,
		AllowZeroValuationRate: allowZeroValuationRate,
	}, nil
}

// isZeroValuationRate 判断物品估值率是否为0
func (s *stockReconciliationSrv) isZeroValuationRate(binValuationMap map[string]float64, materialCode string) bool {
	rate, hasBin := binValuationMap[materialCode]
	return !hasBin || rate == 0
}

// extractName 从错误信息中提取物品名称
func (s *stockReconciliationSrv) extractName(name, after, errorMsg string) string {
	// 转义正则表达式中的特殊字符
	escapedName := regexp.QuoteMeta(name)
	escapedAfter := regexp.QuoteMeta(after)
	// 使用正则表达式匹配，支持HTML标签和普通文本
	var re *regexp.Regexp
	if strings.Contains(name, "<") || strings.Contains(after, "<") {
		// HTML标签模式：不要求空格分隔
		re = regexp.MustCompile(escapedName + `(.+?)` + escapedAfter)
	} else {
		// 普通文本模式：要求空格分隔
		re = regexp.MustCompile(escapedName + `\s+(.+?)\s+` + escapedAfter)
	}
	matches := re.FindStringSubmatch(errorMsg)
	if len(matches) > 1 {
		supplierInfo := strings.TrimSpace(matches[1])
		// 如果包含编码#名称格式，提取名称部分
		if strings.Contains(supplierInfo, "#") {
			parts := strings.SplitN(supplierInfo, "#", 2)
			if len(parts) == 2 {
				return parts[1] // 返回物品erp_code
			}
		}
		return supplierInfo
	}
	return ""
}
