package service

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"ttpos-bmp/app/ttpos-erp/api/stock"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IStockLossSrv 报损单服务接口
type IStockLossSrv interface {
	GetStockLossList(ctx context.Context, req req.StockLossListReq) (resp.StockLossListResp, error)                               // 获取报损单列表
	GetStockLossDetail(ctx context.Context, req req.StockLossDetailReq) (resp.StockLossDetailResp, error)                         // 获取报损单详情
	SaveStockLoss(ctx context.Context, req req.StockLossSaveReq) (uint64, error)                                                  // 保存/提交报损单
	DeleteStockLoss(ctx context.Context, req req.StockLossDeleteReq) error                                                        // 删除报损单
	ApproveStockLoss(ctx context.Context, req req.StockLossApproveReq) error                                                      // 审核通过报损单
	RejectStockLoss(ctx context.Context, req req.StockLossRejectReq) error                                                        // 驳回报损单
	GetStockLossAnnotationList(ctx context.Context, req req.StockLossAnnotationListReq) (resp.StockLossAnnotationListResp, error) // 获取批注列表
}

// stockLossSrv 报损单服务实现
type stockLossSrv struct {
	dbm  *database.DBManager
	lock lock.Lock
}

// NewStockLossSrv 创建报损单服务
func NewStockLossSrv(dbm *database.DBManager) IStockLossSrv {
	return &stockLossSrv{
		dbm:  dbm,
		lock: lock.NewSystemLock(),
	}
}

// GetStockLossList 获取报损单列表
func (s *stockLossSrv) GetStockLossList(ctx context.Context, listReq req.StockLossListReq) (resp.StockLossListResp, error) {
	db := ctx.GetDB()
	stockLossRepo := repository.NewStockLossRepo(db)

	// 构建查询选项
	var opts []repository.DBOption

	// 多仓库筛选
	if len(listReq.WarehouseUuids) > 0 {
		opts = append(opts, stockLossRepo.WhereWarehouseUuids(listReq.WarehouseUuids))
	}

	// 关键字搜索（单据编号和ERP单据编号）
	if listReq.Keyword != "" {
		opts = append(opts, stockLossRepo.WhereKeyword(listReq.Keyword))
	}

	// 创建时间范围筛选
	if listReq.StartCreateTime > 0 || listReq.EndCreateTime > 0 {
		opts = append(opts, stockLossRepo.WhereCreateTimeRange(listReq.StartCreateTime, listReq.EndCreateTime))
	}

	// 状态列表筛选
	if len(listReq.StatusIn) > 0 {
		opts = append(opts, stockLossRepo.WhereStatusIn(listReq.StatusIn))
	}

	// 报损类型筛选
	if listReq.LossType > 0 {
		opts = append(opts, stockLossRepo.WhereLossType(listReq.LossType))
	}

	// 预加载仓库多语言名称
	opts = append(opts, stockLossRepo.WithWarehouse())

	// 按照创建时间倒序排序
	opts = append(opts, func(db *gorm.DB) *gorm.DB {
		return db.Order("create_time DESC")
	})

	// 查询数据
	list, total, err := stockLossRepo.GetStockLossListWithPagination(listReq.PageNo, listReq.PageSize, opts...)
	if err != nil {
		return resp.StockLossListResp{}, errors.WithMessage(err, "查询报损单列表失败")
	}

	// 转换响应数据
	listResp := make([]*resp.StockLossInfo, 0, len(list))

	stockLossUuids := make([]uint64, 0, len(list))
	for _, item := range list {
		stockLossUuids = append(stockLossUuids, item.Uuid)
	}

	// 获取每个报损单的物品数量
	itemsCountMap, err := stockLossRepo.GetStockLossItemCountByUuids(stockLossUuids)
	if err != nil {
		return resp.StockLossListResp{}, errors.WithMessage(err, "查询报损单物品明细数量失败")
	}

	for _, item := range list {
		info := &resp.StockLossInfo{}
		if err := copier.Copy(info, item); err != nil {
			logger.Logger.Error("转换报损单信息失败", zap.Error(err))
			continue
		}
		if item.Warehouse != nil && item.Warehouse.MultiLanguageName != nil {
			info.WarehouseLocaleName = item.Warehouse.MultiLanguageName.GetNames()
		}
		info.ItemsCount = itemsCountMap[item.Uuid]
		listResp = append(listResp, info)
	}

	return resp.StockLossListResp{
		List: listResp,
		Meta: dto.PageResponse{
			PageNo:   listReq.PageNo,
			PageSize: listReq.PageSize,
			Total:    total,
		},
	}, nil
}

// GetStockLossDetail 获取报损单详情
func (s *stockLossSrv) GetStockLossDetail(ctx context.Context, detailReq req.StockLossDetailReq) (resp.StockLossDetailResp, error) {
	db := ctx.GetDB()
	stockLossRepo := repository.NewStockLossRepo(db)
	var detailResp resp.StockLossDetailResp

	opts := []repository.DBOption{
		stockLossRepo.WhereUuid(detailReq.Uuid),
		stockLossRepo.WithWarehouse(),
		stockLossRepo.WithItems(),
		stockLossRepo.WithItemsMaterial(),
		stockLossRepo.WithItemsMaterialNotBaseUnitList(),
		stockLossRepo.WithItemsItemUnits(),
		stockLossRepo.WithItemsItemUnitsMaterialUnit(),
		stockLossRepo.WithFiles(),
		stockLossRepo.WithFilesFile(),
		stockLossRepo.WithAnnotations(),
	}

	// 查询报损单
	stockLoss, err := stockLossRepo.GetStockLoss(opts...)
	if err != nil {
		return detailResp, errors.WithMessage(err, "查询报损单失败")
	}
	if stockLoss == nil {
		return detailResp, errors.New("报损单不存在")
	}

	// 转换响应数据
	if err := copier.Copy(&detailResp, stockLoss); err != nil {
		return detailResp, errors.WithMessage(err, "转换报损单数据失败")
	}

	// 仓库名称
	if stockLoss.Warehouse != nil && stockLoss.Warehouse.MultiLanguageName != nil {
		detailResp.WarehouseName = stockLoss.Warehouse.MultiLanguageName.GetNames()
	}

	// 是否可重新提交（已驳回状态且为发起人）
	detailResp.IsCanResubmit = stockLoss.Status == model.StockLossStatusRejected &&
		stockLoss.SubmitterUuid == ctx.GetStaffUuid()

	// 物品明细
	itemsResp := make([]*resp.StockLossItemInfo, 0, len(stockLoss.StockLossItems))
	for _, item := range stockLoss.StockLossItems {
		itemInfo := &resp.StockLossItemInfo{
			MaterialUuid: item.MaterialUuid,
			BaseQuantity: item.BaseQuantity.InexactFloat64(),
			CreateTime:   int(item.CreateTime),
		}

		// 物品信息
		if item.Material != nil {
			itemInfo.LocaleName = *language.JsonToLocaleResponse(item.MaterialName)
			if itemInfo.LocaleName.IsNull() && item.Material.MultiLanguageName.Uuid > 0 {
				itemInfo.LocaleName = item.Material.MultiLanguageName.GetNames()
			}
			itemInfo.MaterialCode = item.Material.Code
			itemInfo.InternalCode = item.Material.InternalCode
			itemInfo.MaterialBarcode = item.Material.BarcodeValue

			// 构建物品的所有单位列表（用于前端选择）
			itemInfo.Units = make([]resp.MaterialUnitInfo, 0, len(item.Material.NotBaseUnitList))
			for _, materialUnit := range item.Material.NotBaseUnitList {
				unitInfo := resp.MaterialUnitInfo{
					MaterialUnitUuid: materialUnit.Uuid,
					UnitUuid:         materialUnit.UnitUuid,
					ConversionRate:   materialUnit.ConversionRate,
					IsDefault:        materialUnit.IsDefault,
				}
				if materialUnit.Unit != nil && materialUnit.Unit.MultiLanguageName.Uuid > 0 {
					unitInfo.UnitName = materialUnit.Unit.MultiLanguageName.GetNames()
				}
				itemInfo.Units = append(itemInfo.Units, unitInfo)
			}
		}

		// 物品单位明细
		itemInfo.ItemUnits = make([]*resp.StockLossItemUnitInfo, 0, len(item.StockLossItemUnits))
		for _, itemUnit := range item.StockLossItemUnits {
			itemUnitInfo := &resp.StockLossItemUnitInfo{
				MaterialUnitUuid: itemUnit.MaterialUnitUuid,
				Quantity:         itemUnit.Quantity,
			}
			// 单位名称
			itemUnitInfo.LocaleName = *language.JsonToLocaleResponse(itemUnit.MaterialUnitName)
			if itemUnitInfo.LocaleName.IsNull() && itemUnit.MaterialUnit != nil && itemUnit.MaterialUnit.Unit != nil && itemUnit.MaterialUnit.Unit.MultiLanguageName.Uuid > 0 {
				itemUnitInfo.LocaleName = itemUnit.MaterialUnit.Unit.MultiLanguageName.GetNames()
			}
			// 转换率
			if itemUnit.MaterialUnit != nil {
				itemUnitInfo.ConversionRate = itemUnit.MaterialUnit.ConversionRate
			}
			itemInfo.ItemUnits = append(itemInfo.ItemUnits, itemUnitInfo)
		}

		itemsResp = append(itemsResp, itemInfo)
	}
	detailResp.Items = itemsResp

	// 附件列表
	filesResp := make([]*resp.StockLossFileInfo, 0, len(stockLoss.StockLossFiles))
	for _, file := range stockLoss.StockLossFiles {
		if file.File != nil {
			filesResp = append(filesResp, &resp.StockLossFileInfo{
				Uuid:      file.FileUuid,
				Name:      file.File.FileName,
				Url:       file.File.FileUrl,
				FileType:  file.File.FileType,
				SortOrder: file.SortOrder,
			})
		}
	}
	detailResp.Files = filesResp

	// 批注列表
	annotationsResp := make([]*resp.StockLossAnnotationInfo, 0, len(stockLoss.StockLossAnnotations))
	for _, annotation := range stockLoss.StockLossAnnotations {
		annotationsResp = append(annotationsResp, &resp.StockLossAnnotationInfo{
			Uuid:         annotation.Uuid,
			Action:       annotation.Action,
			LocaleName:   getStockLossAnnotationLocaleName(annotation.Action),
			Content:      annotation.Content,
			OperatorUuid: annotation.OperatorUuid,
			OperatorName: annotation.OperatorName,
			CreateTime:   int(annotation.CreateTime),
		})
	}
	detailResp.Annotations = annotationsResp

	return detailResp, nil
}

// SaveStockLoss 保存/提交报损单
// 当 Uuid > 0 且 Items 为空时，表示直接提交已保存的报损单，不更新数据
func (s *stockLossSrv) SaveStockLoss(ctx context.Context, saveReq req.StockLossSaveReq) (uint64, error) {
	db := ctx.GetDB()
	companySetting := ctx.GetCompanySetting()
	timezone := companySetting.GetTimezone()
	stockLossUuid := saveReq.Uuid
	var stockLoss *model.StockLoss
	var err error

	// 条件验证：新建时必填字段检查
	if saveReq.Uuid == 0 {
		if saveReq.WarehouseUuid == 0 {
			return 0, errors.New("仓库UUID不能为空")
		}
		if saveReq.LossType == 0 {
			return 0, errors.New("报损类型不能为空")
		}
		if len(saveReq.Items) == 0 {
			return 0, errors.New("报损物品明细不能为空")
		}
	}

	// 判断是否为"仅提交"模式（Uuid > 0 且 Items 为空）
	isSubmitOnly := saveReq.Uuid > 0 && len(saveReq.Items) == 0

	if saveReq.Uuid == 0 { // 新建
		// 加锁保证单号唯一性（基于公司UUID和日期）
		dateStr := utils.SetTimezone(timezone).Now().Format("20060102")
		lockKey := fmt.Sprintf("stock_loss_%d_%s", ctx.GetCompanyUuid(), dateStr)
		s.lock.LockUuidString(lockKey)
		defer s.lock.UnlockUuidString(lockKey)
	} else { // 修改
		// 加锁
		s.lock.LockUuid(saveReq.Uuid)
		defer s.lock.UnlockUuid(saveReq.Uuid)

		stockLossRepo := repository.NewStockLossRepo(db)
		// 查询报损单
		stockLoss, err = stockLossRepo.GetStockLossByUuid(saveReq.Uuid)
		if err != nil {
			return stockLossUuid, errors.WithMessage(err, "查询报损单失败")
		}
		if stockLoss == nil {
			return stockLossUuid, errors.New("报损单不存在")
		}

		// 重新提交场景：只有已驳回状态才能重新提交
		if saveReq.GetIsResubmit() {
			if stockLoss.Status != model.StockLossStatusRejected {
				return stockLossUuid, errors.New("报损单状态不允许重新提交")
			}
			// 验证只有提交人才能重新提交
			if stockLoss.SubmitterUuid != ctx.GetStaffUuid() {
				return stockLossUuid, errors.New("只有发起人才能重新提交")
			}
		} else {
			// 只有已保存状态的报损单才能修改
			if stockLoss.Status != model.StockLossStatusSaved {
				if saveReq.IsSubmit {
					return stockLossUuid, errors.New("当前状态不允许提交")
				} else {
					return stockLossUuid, errors.New("当前状态不允许修改")
				}
			}
		}
	}

	// 仅提交模式：直接提交已保存的报损单，不更新数据
	if isSubmitOnly {
		if !saveReq.IsSubmit {
			return stockLossUuid, errors.New("参数错误")
		}
		// 直接提交，跳过数据验证和更新
		staffName := ""
		if staff := ctx.GetStaff(); staff.Uuid > 0 {
			staffName = staff.RealName
		}
		err = s.submitStockLoss(ctx, stockLoss.Uuid, staffName)
		if err != nil {
			return stockLossUuid, errors.WithMessage(err, "提交报损单失败")
		}
		return stockLossUuid, nil
	}

	// 验证报损类型
	if saveReq.LossType < model.StockLossTypeDamage || saveReq.LossType > model.StockLossTypeExpired {
		return stockLossUuid, errors.New("报损类型参数错误")
	}

	// 验证报损原因长度（最多500个字符）
	if utf8.RuneCountInString(saveReq.Reason) > 500 {
		return stockLossUuid, errors.New("报损原因最多500个字符")
	}

	// 验证仓库和物品
	materials, err := s.validateWarehouseAndItems(db, saveReq, ctx.GetLanguage())
	if err != nil {
		return stockLossUuid, err
	}

	// 构建物品名称、单位名称和转换率映射
	materialNameMap := make(map[uint64]string)
	materialUnitNameMap := make(map[uint64]map[uint64]string)
	materialUnitMap := make(map[uint64]map[uint64]float64) // 转换率映射
	for _, material := range materials {
		materialNameMap[material.Uuid] = material.Name
		materialUnitNameMap[material.Uuid] = make(map[uint64]string)
		materialUnitMap[material.Uuid] = make(map[uint64]float64)
		for _, materialUnit := range material.NotBaseUnitList {
			materialUnitNameMap[material.Uuid][materialUnit.Uuid] = materialUnit.Name
			materialUnitMap[material.Uuid][materialUnit.Uuid] = materialUnit.ConversionRate
		}
	}

	// 获取 saas 数据库连接
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	if saasDB == nil {
		return stockLossUuid, errors.New("saas 数据库连接失败")
	}

	// 获取公司 UUID（使用总部 UUID 或当前公司 UUID）
	companyUuid := companySetting.HeadquarterUuid
	if companyUuid == 0 {
		companyUuid = ctx.GetCompanyUuid()
	}

	// 获取员工信息
	staffName := ""
	if staff := ctx.GetStaff(); staff.Uuid > 0 {
		staffName = staff.RealName
	}

	// 开启事务
	err = db.Transaction(func(tx *gorm.DB) error {
		stockLossRepo := repository.NewStockLossRepo(tx)

		// 重新提交成功后添加批注记录（包含报损原因）
		if saveReq.GetIsResubmit() {
			annotation := &model.StockLossAnnotation{
				StockLossUuid: stockLoss.Uuid,
				Action:        model.StockLossActionResubmit,
				Content:       saveReq.Reason,
				OperatorUuid:  ctx.GetStaffUuid(),
				OperatorName:  staffName,
			}
			if err := stockLossRepo.CreateAnnotation(annotation); err != nil {
				logger.Logger.Error("保存重新提交批注失败", zap.Error(err))
				// 批注保存失败不影响主流程，仅记录日志
			}
		}

		if saveReq.Uuid == 0 { // 新建
			// 生成单据编号
			code, err := s.generateCode(saasDB, companyUuid, timezone)
			if err != nil {
				return errors.WithMessage(err, "生成单据编号失败")
			}
			// 创建报损单
			stockLoss = &model.StockLoss{
				Code:          code,
				LossType:      saveReq.LossType,
				WarehouseUuid: saveReq.WarehouseUuid,
				Reason:        saveReq.Reason,
				Status:        model.StockLossStatusSaved, // 0-已保存
				SubmitterUuid: ctx.GetStaffUuid(),         // 记录发起人
			}
			if err := stockLossRepo.CreateStockLoss(stockLoss); err != nil {
				return errors.WithMessage(errors.New("创建报损单失败"), err.Error())
			}

			stockLossUuid = stockLoss.Uuid
		} else { // 更新
			// 更新报损单
			stockLoss.LossType = saveReq.LossType
			stockLoss.WarehouseUuid = saveReq.WarehouseUuid
			stockLoss.Reason = saveReq.Reason

			if err := stockLossRepo.UpdateStockLoss(stockLoss); err != nil {
				return errors.WithMessage(err, "更新报损单失败")
			}
			// 删除原有的物品单位明细（必须在删除物品明细之前执行）
			if err := stockLossRepo.DeleteItemUnitsByStockLossUuid(saveReq.Uuid); err != nil {
				return errors.WithMessage(err, "删除报损单物品单位明细失败")
			}
			// 删除原有的物品明细
			if err := stockLossRepo.DeleteItemsByStockLossUuid(saveReq.Uuid); err != nil {
				return errors.WithMessage(err, "删除报损单物品明细失败")
			}
			// 删除原有的附件关联
			if err := stockLossRepo.DeleteFilesByStockLossUuid(saveReq.Uuid); err != nil {
				return errors.WithMessage(err, "删除报损单附件失败")
			}
		}

		// 创建物品明细和单位明细
		var stockLossItemUnits []*model.StockLossItemUnit
		for _, reqItem := range saveReq.Items {
			// 计算基准单位数量（所有单位转换后的总量）
			baseQuantity := decimal.Zero
			if len(reqItem.Units) > 0 {
				for _, unitReq := range reqItem.Units {
					if unitReq.Quantity == nil {
						continue
					}
					conversionRate := materialUnitMap[reqItem.MaterialUuid][unitReq.MaterialUnitUuid]
					unitQuantity := unitReq.Quantity.Mul(decimal.NewFromFloat(conversionRate))
					baseQuantity = baseQuantity.Add(unitQuantity)
				}
			}
			baseQuantity = baseQuantity.Truncate(4)

			// 创建物品明细
			item := &model.StockLossItem{
				StockLossUuid: stockLoss.Uuid,
				MaterialUuid:  reqItem.MaterialUuid,
				MaterialName:  materialNameMap[reqItem.MaterialUuid],
				BaseQuantity:  baseQuantity,
			}
			if err := stockLossRepo.CreateItem(item); err != nil {
				return errors.WithMessage(errors.New("创建报损单物品明细失败"), err.Error())
			}

			// 收集单位明细
			for _, unitReq := range reqItem.Units {
				var quantity *float64
				if unitReq.Quantity != nil {
					q := unitReq.Quantity.InexactFloat64()
					quantity = &q
				}
				stockLossItemUnits = append(stockLossItemUnits, &model.StockLossItemUnit{
					StockLossItemUuid: item.Uuid,
					MaterialUnitUuid:  unitReq.MaterialUnitUuid,
					MaterialUnitName:  materialUnitNameMap[reqItem.MaterialUuid][unitReq.MaterialUnitUuid],
					Quantity:          quantity,
				})
			}
		}

		// 批量创建物品单位明细
		if len(stockLossItemUnits) > 0 {
			if err := stockLossRepo.CreateItemUnits(stockLossItemUnits); err != nil {
				return errors.WithMessage(errors.New("创建报损单物品单位明细失败"), err.Error())
			}
		}

		// 创建附件关联
		for i, fileUuid := range saveReq.FileUuids {
			file := &model.StockLossFile{
				StockLossUuid: stockLoss.Uuid,
				FileUuid:      fileUuid,
				SortOrder:     i,
			}
			if err := stockLossRepo.CreateFile(file); err != nil {
				return errors.WithMessage(errors.New("创建报损单附件关联失败"), err.Error())
			}
		}

		return nil
	})

	if err != nil {
		errMsg := "保存失败"
		if saveReq.IsSubmit || saveReq.GetIsResubmit() {
			errMsg = "提交失败"
		}
		return stockLossUuid, errors.WithMessage(errors.New(errMsg), err.Error())
	}

	// 提交报损单（包括首次提交和重新提交）
	if saveReq.IsSubmit || saveReq.GetIsResubmit() {
		err = s.submitStockLoss(ctx, stockLoss.Uuid, staffName)
		if err != nil {
			return stockLossUuid, errors.WithMessage(err, "提交报损单失败")
		}
	}

	return stockLossUuid, nil
}

// submitStockLoss 提交报损单
func (s *stockLossSrv) submitStockLoss(ctx context.Context, stockLossUuid uint64, staffName string) error {
	db := ctx.GetDB()
	stockLossRepo := repository.NewStockLossRepo(db)

	opts := []repository.DBOption{
		stockLossRepo.WhereUuid(stockLossUuid),
		stockLossRepo.WithWarehouse(),
		stockLossRepo.WithItems(),
		stockLossRepo.WithItemsMaterial(),
		stockLossRepo.WithItemsItemUnits(),
		stockLossRepo.WithFiles(),
	}
	stockLoss, err := stockLossRepo.GetStockLoss(opts...)
	if err != nil {
		return errors.WithMessage(errors.New("查询报损单失败"), err.Error())
	}
	if stockLoss == nil {
		return errors.New("报损单不存在")
	}

	// 检查是否上传了附件
	if len(stockLoss.StockLossFiles) == 0 {
		return errors.New("请上传附件")
	}

	// 检查仓库状态
	if stockLoss.Warehouse != nil && stockLoss.Warehouse.IsDisabled() {
		return errors.New("仓库状态已关闭，请修改仓库状态")
	}

	// 检查物品是否有效
	if len(stockLoss.StockLossItems) == 0 {
		return errors.New("物品列表为空，请先添加物品后再操作")
	}

	// 验证物品状态和库存
	if err := s.validateItemsForSubmit(db, stockLoss, ctx.GetLanguage()); err != nil {
		return err
	}

	// 开启事务
	err = db.Transaction(func(tx *gorm.DB) error {
		stockLossRepo := repository.NewStockLossRepo(tx)

		// 收集数量为0的单位明细UUID和没有有效单位的物品明细UUID
		var zeroQuantityUnitUuids []uint64
		var emptyItemUuids []uint64
		var validItems []*model.StockLossItem

		for _, item := range stockLoss.StockLossItems {
			hasValidUnit := false
			for _, itemUnit := range item.StockLossItemUnits {
				// 数量为 nil 或 0 的单位需要删除
				if itemUnit.Quantity == nil || *itemUnit.Quantity == 0 {
					zeroQuantityUnitUuids = append(zeroQuantityUnitUuids, itemUnit.Uuid)
				} else {
					hasValidUnit = true
				}
			}
			// 如果物品没有任何有效单位，标记为待删除
			if !hasValidUnit {
				emptyItemUuids = append(emptyItemUuids, item.Uuid)
			} else {
				validItems = append(validItems, item)
			}
		}

		// 删除数量为0的单位明细
		if len(zeroQuantityUnitUuids) > 0 {
			if err := stockLossRepo.DeleteItemUnitsByUuids(zeroQuantityUnitUuids); err != nil {
				return errors.WithMessage(err, "删除0数量单位明细失败")
			}
		}

		// 删除没有有效单位的物品明细
		if len(emptyItemUuids) > 0 {
			if err := stockLossRepo.DeleteItemsByUuids(emptyItemUuids); err != nil {
				return errors.WithMessage(err, "删除空物品明细失败")
			}
			// 更新内存中的物品列表
			stockLoss.StockLossItems = validItems
		}

		// 删除后检查物品列表是否为空
		if len(stockLoss.StockLossItems) == 0 {
			return errors.New("物品列表为空，请先添加物品后再操作")
		}

		// 更新报损单状态为已提交
		stockLoss.Status = model.StockLossStatusSubmit
		stockLoss.SubmitTime = int(time.Now().Unix())
		stockLoss.SubmitterUuid = ctx.GetStaffUuid()
		if err := stockLossRepo.UpdateStockLoss(stockLoss); err != nil {
			return errors.WithMessage(errors.New("更新报损单状态失败"), err.Error())
		}

		// 创建提交批注（包含报损原因）
		annotation := &model.StockLossAnnotation{
			StockLossUuid: stockLoss.Uuid,
			Action:        model.StockLossActionSubmit,
			Content:       stockLoss.Reason,
			OperatorUuid:  ctx.GetStaffUuid(),
			OperatorName:  staffName,
		}
		if err := stockLossRepo.CreateAnnotation(annotation); err != nil {
			logger.Logger.Error("保存提交批注失败", zap.Error(err))
		}

		return nil
	})

	return err
}

// DeleteStockLoss 删除报损单
func (s *stockLossSrv) DeleteStockLoss(ctx context.Context, deleteReq req.StockLossDeleteReq) error {
	db := ctx.GetDB()

	// 加锁
	s.lock.LockUuid(deleteReq.Uuid)
	defer s.lock.UnlockUuid(deleteReq.Uuid)

	stockLossRepo := repository.NewStockLossRepo(db)

	// 查询报损单
	stockLoss, err := stockLossRepo.GetStockLossByUuid(deleteReq.Uuid)
	if err != nil {
		return errors.WithMessage(err, "查询报损单失败")
	}
	if stockLoss == nil {
		return errors.New("报损单不存在")
	}

	// 只有已保存状态的报损单才能删除
	if stockLoss.Status != model.StockLossStatusSaved {
		return errors.New("当前状态不允许删除")
	}

	// 删除报损单
	if err := stockLossRepo.DeleteStockLoss(deleteReq.Uuid); err != nil {
		return errors.WithMessage(err, "删除报损单失败")
	}

	// 清理UUID锁资源
	s.lock.ClearUuidLock(deleteReq.Uuid)

	return nil
}

// ApproveStockLoss 审核通过报损单
func (s *stockLossSrv) ApproveStockLoss(ctx context.Context, approveReq req.StockLossApproveReq) error {
	db := ctx.GetDB()

	// 加锁
	s.lock.LockUuid(approveReq.Uuid)
	defer s.lock.UnlockUuid(approveReq.Uuid)

	stockLossRepo := repository.NewStockLossRepo(db)

	opts := []repository.DBOption{
		stockLossRepo.WhereUuid(approveReq.Uuid),
		stockLossRepo.WithWarehouse(),
		stockLossRepo.WithItems(),
		stockLossRepo.WithItemsMaterial(),
	}

	// 查询报损单
	stockLoss, err := stockLossRepo.GetStockLoss(opts...)
	if err != nil {
		return errors.WithMessage(err, "查询报损单失败")
	}
	if stockLoss == nil {
		return errors.New("报损单不存在")
	}

	// 只有已提交状态的报损单才能审核
	if stockLoss.Status != model.StockLossStatusSubmit {
		return errors.New("当前状态不允许审核")
	}

	// 检查仓库状态
	if stockLoss.Warehouse != nil && stockLoss.Warehouse.IsDisabled() {
		return errors.New("仓库状态已关闭，请修改仓库状态")
	}

	// 验证物品状态和库存
	if err := s.validateItemsForSubmit(db, stockLoss, ctx.GetLanguage()); err != nil {
		return err
	}

	// 获取员工信息
	staffName := ""
	if staff := ctx.GetStaff(); staff.Uuid > 0 {
		staffName = staff.RealName
	}

	// 获取仓库物品信息
	warehouseItemRepo := repository.NewWarehouseItemRepo(db)
	warehouseItems, err := warehouseItemRepo.GetByWarehouseUuid(stockLoss.WarehouseUuid)
	if err != nil {
		return errors.WithMessage(errors.New("查询仓库物品失败"), err.Error())
	}
	warehouseItemMap := make(map[uint64]model.WarehouseItem)
	for _, item := range warehouseItems {
		warehouseItemMap[item.MaterialUuid] = item
	}

	// 获取历史批注（用于 ERP remarks）
	annotations, err := stockLossRepo.GetAnnotationsByStockLossUuid(approveReq.Uuid)
	if err != nil {
		logger.Logger.Error("获取报损单历史批注失败", zap.Error(err), zap.Uint64("stock_loss_uuid", approveReq.Uuid))
		// 不影响主流程，继续执行
		annotations = make([]*model.StockLossAnnotation, 0)
	}

	// 开启事务
	err = db.Transaction(func(tx *gorm.DB) error {
		stockLossRepo := repository.NewStockLossRepo(tx)

		// 调用 ERP 接口创建出库单（Stock Entry - Material Issue）
		// 策略："先 ERP 后库存"，审核通过时调用 ERP 创建出库单
		if ctx.GetCompany().IsOpenErp() && stockLoss.Warehouse != nil && stockLoss.Warehouse.ErpCode != "" {
			companySetting := ctx.GetCompanySetting()
			erpSrv := erp.NewIErpSrv(s.dbm)

			// 构建 ERP 请求参数
			erpItems := make([]*stock.StockEntryItem, 0, len(stockLoss.StockLossItems))
			for _, item := range stockLoss.StockLossItems {
				// 获取物品编码
				itemCode := ""
				if item.Material != nil {
					itemCode = item.Material.Code
				}
				if itemCode == "" {
					continue
				}
				// 跳过数量为 0 的物品
				if item.BaseQuantity.IsZero() {
					continue
				}
				erpItems = append(erpItems, &stock.StockEntryItem{
					ItemCode:  itemCode,
					Qty:       item.BaseQuantity.InexactFloat64(),
					Warehouse: stockLoss.Warehouse.ErpCode,
					ItemName:  item.MaterialName,
				})
			}

			if len(erpItems) > 0 {
				// 构建 ERP remarks（批注按倒序排列）
				approveTime := int(time.Now().Unix())
				erpRemarks := buildStockLossErpRemarks(annotations, model.StockLossActionApprove, approveReq.Annotation, approveTime)

				now := utils.SetTimezone(companySetting.GetTimezone()).Now()
				erpResp, err := erpSrv.SubmitStockEntry(ctx, companySetting, &stock.SubmitStockEntryReq{
					StockEntryType:  "Material Issue", // 物料出库
					SourceWarehouse: stockLoss.Warehouse.ErpCode,
					Items:           erpItems,
					Remarks:         erpRemarks,
					PostingDate:     now.Format("2006-01-02"),
					PostingTime:     now.Format("15:04:05"),
				})
				if err != nil {
					logger.Logger.Error("调用ERP创建出库单失败", zap.Error(err), zap.Uint64("stock_loss_uuid", stockLoss.Uuid))
					return s.parseErpErrorForStockLoss(tx, err.Error(), ctx.GetLanguage())
				}
				// 保存 ERP 单号
				stockLoss.ErpCode = erpResp.StockEntryName
			}
		}

		// 更新报损单状态为已审核通过
		stockLoss.Status = model.StockLossStatusApproved
		stockLoss.ApproveTime = int(time.Now().Unix())
		stockLoss.ApproverUuid = ctx.GetStaffUuid()
		if err := stockLossRepo.UpdateStockLoss(stockLoss); err != nil {
			return errors.WithMessage(errors.New("审核报损单失败"), err.Error())
		}

		// 保存批注记录
		annotation := &model.StockLossAnnotation{
			StockLossUuid: approveReq.Uuid,
			Action:        model.StockLossActionApprove,
			Content:       approveReq.Annotation,
			OperatorUuid:  ctx.GetStaffUuid(),
			OperatorName:  staffName,
		}
		if err := stockLossRepo.CreateAnnotation(annotation); err != nil {
			return errors.WithMessage(err, "保存批注失败")
		}

		// 扣减仓库库存并记录出入库日志
		var warehouseLogs []*model.WarehouseInOutLog
		for _, item := range stockLoss.StockLossItems {
			// 检查物品是否在仓库中
			warehouseItem, exists := warehouseItemMap[item.MaterialUuid]
			if !exists {
				continue // 物品不在仓库中，跳过
			}

			// 直接使用 BaseQuantity（已经是转换后的基准单位数量）
			deductQty := item.BaseQuantity.InexactFloat64()

			// 更新仓库物品库存
			newStock := warehouseItem.Stock - deductQty
			if newStock < 0 {
				newStock = 0 // 防止负库存
			}
			if err := tx.Model(&model.WarehouseItem{}).
				Where("warehouse_uuid = ?", stockLoss.WarehouseUuid).
				Where("material_uuid = ?", item.MaterialUuid).
				Update("stock", newStock).Error; err != nil {
				return errors.WithMessage(errors.New("更新仓库物品库存失败"), err.Error())
			}

			// 获取物品的基准单位信息
			var baseUnitUuid uint64
			var baseUnitName string
			if item.Material != nil {
				for _, materialUnit := range item.Material.NotBaseUnitList {
					if materialUnit.IsDefault == 1 {
						baseUnitUuid = materialUnit.Uuid
						baseUnitName = materialUnit.Name
						break
					}
				}
			}

			// 记录出库日志
			warehouseLogs = append(warehouseLogs, &model.WarehouseInOutLog{
				LogType:              constant.WarehouseInOutLogLogTypeOut,        // 出库
				Scene:                constant.WarehouseInOutLogSceneStockLossOut, // 报损出库
				WarehouseUuid:        stockLoss.WarehouseUuid,
				MaterialUuid:         item.MaterialUuid,
				MaterialName:         item.MaterialName,
				MaterialBaseUnitUuid: baseUnitUuid,
				MaterialBaseUnitName: baseUnitName,
				Num:                  deductQty,
				OrderNo:              stockLoss.Code,
			})
		}

		if len(warehouseLogs) > 0 {
			if err := tx.Create(&warehouseLogs).Error; err != nil {
				return errors.WithMessage(errors.New("创建报损出库记录失败"), err.Error())
			}
		}

		return nil
	})

	return err
}

// RejectStockLoss 驳回报损单
func (s *stockLossSrv) RejectStockLoss(ctx context.Context, rejectReq req.StockLossRejectReq) error {
	// 验证驳回原因必填
	if strings.TrimSpace(rejectReq.Annotation) == "" {
		return errors.New("请填写驳回原因")
	}

	db := ctx.GetDB()

	// 加锁
	s.lock.LockUuid(rejectReq.Uuid)
	defer s.lock.UnlockUuid(rejectReq.Uuid)

	stockLossRepo := repository.NewStockLossRepo(db)

	// 查询报损单
	stockLoss, err := stockLossRepo.GetStockLossByUuid(rejectReq.Uuid)
	if err != nil {
		return errors.WithMessage(err, "查询报损单失败")
	}
	if stockLoss == nil {
		return errors.New("报损单不存在")
	}

	// 只有已提交状态的报损单才能驳回
	if stockLoss.Status != model.StockLossStatusSubmit {
		return errors.New("报损单状态不允许驳回")
	}

	// 获取员工信息
	staffName := ""
	if staff := ctx.GetStaff(); staff.Uuid > 0 {
		staffName = staff.RealName
	}

	// 开启事务
	err = db.Transaction(func(tx *gorm.DB) error {
		stockLossRepo := repository.NewStockLossRepo(tx)

		// 更新报损单状态为已驳回
		stockLoss.Status = model.StockLossStatusRejected
		stockLoss.RejectTime = int(time.Now().Unix())
		if err := stockLossRepo.UpdateStockLoss(stockLoss); err != nil {
			return errors.WithMessage(errors.New("驳回报损单失败"), err.Error())
		}

		// 保存批注记录
		annotation := &model.StockLossAnnotation{
			StockLossUuid: rejectReq.Uuid,
			Action:        model.StockLossActionReject,
			Content:       rejectReq.Annotation,
			OperatorUuid:  ctx.GetStaffUuid(),
			OperatorName:  staffName,
		}
		if err := stockLossRepo.CreateAnnotation(annotation); err != nil {
			return errors.WithMessage(err, "保存批注失败")
		}

		return nil
	})

	return err
}

// GetStockLossAnnotationList 获取批注列表
func (s *stockLossSrv) GetStockLossAnnotationList(ctx context.Context, annotationReq req.StockLossAnnotationListReq) (resp.StockLossAnnotationListResp, error) {
	db := ctx.GetDB()
	stockLossRepo := repository.NewStockLossRepo(db)

	annotations, err := stockLossRepo.GetAnnotationsByStockLossUuid(annotationReq.StockLossUuid)
	if err != nil {
		return resp.StockLossAnnotationListResp{}, errors.WithMessage(err, "查询批注列表失败")
	}

	listResp := make([]*resp.StockLossAnnotationInfo, 0, len(annotations))
	for _, annotation := range annotations {
		listResp = append(listResp, &resp.StockLossAnnotationInfo{
			Uuid:         annotation.Uuid,
			Action:       annotation.Action,
			LocaleName:   getStockLossAnnotationLocaleName(annotation.Action),
			Content:      annotation.Content,
			OperatorUuid: annotation.OperatorUuid,
			OperatorName: annotation.OperatorName,
			CreateTime:   int(annotation.CreateTime),
		})
	}

	return resp.StockLossAnnotationListResp{
		List: listResp,
	}, nil
}

// generateCode 生成报损单编号
// 格式：SL + yyyyMMddHHmmss + 序列号（4位）
// 例如：SL202504030915120001
func (s *stockLossSrv) generateCode(
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
	seq, err := seqRepo.GetNextSequence(companyUuid, constant.NumberTypeStockLoss, dateStr)
	if err != nil {
		return "", errors.WithMessage(err, "获取序列号失败")
	}

	// 组装编号：SL + timestamp + 序列号（4位）
	code := fmt.Sprintf("SL%s%04d", timestamp, seq)
	return code, nil
}

// validateWarehouseAndItems 验证仓库和物品明细
func (s *stockLossSrv) validateWarehouseAndItems(db *gorm.DB, saveReq req.StockLossSaveReq, defaultLang string) ([]*model.Material, error) {
	warehouseUuid := saveReq.WarehouseUuid
	if warehouseUuid == 0 {
		return nil, errors.New("仓库参数错误")
	}

	// 判断仓库是否存在，且类型为normal
	warehouseRepo := repository.NewWarehouseRepo(db)
	warehouse, err := warehouseRepo.GetByUuid(warehouseUuid)
	if err != nil {
		return nil, errors.WithMessage(errors.New("查询仓库失败"), err.Error())
	}
	if warehouse == nil || warehouse.Type != constant.WarehouseTypeNormal {
		return nil, errors.New("仓库参数错误")
	}

	// 获取所有物品UUID
	materialUuids := make([]uint64, 0, len(saveReq.Items))
	for _, item := range saveReq.Items {
		materialUuids = append(materialUuids, item.MaterialUuid)
	}

	// 批量查询物品详情（包含已删除的，用于区分删除和禁用，并获取名称和内部编码）
	materialRepo := repository.NewMaterialRepo(db)
	var allMaterials []*model.Material
	materialMap := make(map[uint64]*model.Material)
	if len(materialUuids) > 0 {
		allMaterials, err = materialRepo.GetMaterialDetailContainsDeletedByUuids(materialUuids)
		if err != nil {
			return nil, errors.WithMessage(errors.New("查询物品详情失败"), err.Error())
		}
		for _, material := range allMaterials {
			materialMap[material.Uuid] = material
		}
	}

	// 收集未找到的物品（不存在或被删除）
	var notFoundItems []string
	// 收集被禁用的物品
	var disabledItems []string
	// 有效的物品列表（未删除且启用）
	var validMaterials []*model.Material

	// 验证请求中的物品和单位
	for _, item := range saveReq.Items {
		material, exists := materialMap[item.MaterialUuid]
		if !exists {
			// 物品完全不存在
			notFoundItems = append(notFoundItems, fmt.Sprintf("%d", item.MaterialUuid))
			continue
		}

		// 检查物品是否被软删除
		if material.DeleteTime > 0 {
			notFoundItems = append(notFoundItems, formatMaterialName(material.Name, material.InternalCode, defaultLang))
			continue
		}

		// 检查物品是否启用
		if !material.Status {
			disabledItems = append(disabledItems, formatMaterialName(material.Name, material.InternalCode, defaultLang))
			continue
		}

		// 验证物品的所有单位
		for _, unitReq := range item.Units {
			unitExists := false
			for _, materialUnit := range material.NotBaseUnitList {
				if unitReq.MaterialUnitUuid == materialUnit.Uuid {
					unitExists = true
					break
				}
			}
			if !unitExists {
				return nil, errors.New("物品单位参数错误")
			}
		}

		validMaterials = append(validMaterials, material)
	}

	// 优先返回未找到的物品错误
	if len(notFoundItems) > 0 {
		return nil, errors.New(fmt.Sprintf(i18n.Translate(defaultLang, "物品%s未找到"), strings.Join(notFoundItems, "、")))
	}

	// 返回禁用的物品错误
	if len(disabledItems) > 0 {
		return nil, errors.New(fmt.Sprintf(i18n.Translate(defaultLang, "物品%s的状态已关闭。请修改物品状态"), strings.Join(disabledItems, "、")))
	}

	return validMaterials, nil
}

// formatMaterialName 格式化物品名称，格式：名称（内部编码）
// nameJson: 多语言 JSON 字符串
// internalCode: 内部编码
// defaultLang: 商家默认语言
func formatMaterialName(nameJson string, internalCode string, defaultLang string) string {
	// 解析多语言 JSON，获取指定语言的名称
	locale := language.JsonToLocaleResponse(nameJson)
	name := locale.GetLocale(defaultLang)
	if name == "" {
		name = nameJson // 如果解析失败，使用原始字符串
	}
	if internalCode != "" {
		return name + "（" + internalCode + "）"
	}
	return name
}

// validateItemsForSubmit 提交时验证物品状态和库存
func (s *stockLossSrv) validateItemsForSubmit(db *gorm.DB, stockLoss *model.StockLoss, defaultLang string) error {
	// 收集被删除的物品（Material 为 nil 或 DeleteTime > 0）
	var notFoundItems []string
	// 收集被禁用的物品
	var disabledItems []string
	// 收集库存不足的物品
	var insufficientStockItems []string

	// 获取仓库物品库存信息
	warehouseItemRepo := repository.NewWarehouseItemRepo(db)
	warehouseItems, err := warehouseItemRepo.GetByWarehouseUuid(stockLoss.WarehouseUuid)
	if err != nil {
		return errors.WithMessage(errors.New("查询仓库物品失败"), err.Error())
	}
	warehouseItemMap := make(map[uint64]model.WarehouseItem)
	for _, item := range warehouseItems {
		warehouseItemMap[item.MaterialUuid] = item
	}

	// 收集 Material 为 nil 的物品 UUID，单独查询以获取名称和内部编码
	var nilMaterialUuids []uint64
	for _, item := range stockLoss.StockLossItems {
		if item.Material == nil {
			nilMaterialUuids = append(nilMaterialUuids, item.MaterialUuid)
		}
	}

	// 查询软删除的物品详情（包含名称和内部编码）
	deletedMaterialMap := make(map[uint64]*model.Material)
	if len(nilMaterialUuids) > 0 {
		materialRepo := repository.NewMaterialRepo(db)
		deletedMaterials, _ := materialRepo.GetMaterialDetailContainsDeletedByUuids(nilMaterialUuids)
		for _, m := range deletedMaterials {
			deletedMaterialMap[m.Uuid] = m
		}
	}

	// 按物品 UUID 汇总报损数量（基准单位）
	materialBaseQtyMap := make(map[uint64]float64)
	for _, item := range stockLoss.StockLossItems {
		materialBaseQtyMap[item.MaterialUuid] += item.BaseQuantity.InexactFloat64()
	}

	// 验证每个物品
	for _, item := range stockLoss.StockLossItems {
		// 检查物品是否存在（被删除）
		if item.Material == nil {
			// 尝试从软删除物品中获取名称和内部编码
			if deletedMaterial, exists := deletedMaterialMap[item.MaterialUuid]; exists {
				notFoundItems = append(notFoundItems, formatMaterialName(deletedMaterial.Name, deletedMaterial.InternalCode, defaultLang))
			} else {
				// 完全不存在，使用保存时备份的名称
				notFoundItems = append(notFoundItems, formatMaterialName(item.MaterialName, "", defaultLang))
			}
			continue
		}

		// 检查物品是否被软删除
		if item.Material.DeleteTime > 0 {
			notFoundItems = append(notFoundItems, formatMaterialName(item.Material.Name, item.Material.InternalCode, defaultLang))
			continue
		}

		// 检查物品是否禁用
		if !item.Material.Status {
			disabledItems = append(disabledItems, formatMaterialName(item.Material.Name, item.Material.InternalCode, defaultLang))
			continue
		}

		// 检查库存是否充足（同一物品只检查一次）
		if _, checked := materialBaseQtyMap[item.MaterialUuid]; checked {
			totalQty := materialBaseQtyMap[item.MaterialUuid]
			warehouseItem, exists := warehouseItemMap[item.MaterialUuid]
			if !exists || warehouseItem.Stock < totalQty {
				insufficientStockItems = append(insufficientStockItems, formatMaterialName(item.Material.Name, item.Material.InternalCode, defaultLang))
			}
			// 标记为已检查，避免重复
			delete(materialBaseQtyMap, item.MaterialUuid)
		}
	}

	// 优先返回未找到的物品错误
	if len(notFoundItems) > 0 {
		return errors.New(fmt.Sprintf(i18n.Translate(defaultLang, "物品%s未找到"), strings.Join(notFoundItems, "、")))
	}

	// 返回禁用的物品错误
	if len(disabledItems) > 0 {
		return errors.New(fmt.Sprintf(i18n.Translate(defaultLang, "物品%s的状态已关闭。请修改物品状态"), strings.Join(disabledItems, "、")))
	}

	// 返回库存不足的物品错误
	if len(insufficientStockItems) > 0 {
		return errors.New(fmt.Sprintf(i18n.Translate(defaultLang, "物品%s的库存不足"), strings.Join(insufficientStockItems, "、")))
	}

	return nil
}

// getStockLossAnnotationLocaleName 获取批注操作类型的多语言名称
func getStockLossAnnotationLocaleName(action string) dto.LocaleResponse {
	switch action {
	case model.StockLossActionSubmit:
		return dto.LocaleResponse{
			ZH: "提交", ZHTW: "提交", EN: "Submit", JA: "提出", KO: "제출",
			TH: "ส่ง", MY: "တင်သွင်းရန်", TR: "Gönder", SV: "Skicka",
		}
	case model.StockLossActionResubmit:
		return dto.LocaleResponse{
			ZH: "重新提交", ZHTW: "重新提交", EN: "Resubmit", JA: "再提出", KO: "재제출",
			TH: "ส่งใหม่", MY: "ပြန်လည်တင်သွင်းရန်", TR: "Yeniden Gönder", SV: "Skicka igen",
		}
	case model.StockLossActionApprove:
		return dto.LocaleResponse{
			ZH: "审核通过", ZHTW: "審核通過", EN: "Approved", JA: "承認済み", KO: "승인됨",
			TH: "อนุมัติ", MY: "အတည်ပြုပြီး", TR: "Onaylandı", SV: "Godkänd",
		}
	case model.StockLossActionReject:
		return dto.LocaleResponse{
			ZH: "驳回", ZHTW: "駁回", EN: "Rejected", JA: "却下", KO: "반려됨",
			TH: "ปฏิเสธ", MY: "ငြင်းပယ်ပြီး", TR: "Reddedildi", SV: "Avvisad",
		}
	default:
		return dto.LocaleResponse{}
	}
}

// getStockLossAnnotationErpActionName 获取批注操作类型的 ERP 英文名称
func getStockLossAnnotationErpActionName(action string) string {
	switch action {
	case model.StockLossActionSubmit:
		return "Pending Review"
	case model.StockLossActionResubmit:
		return "Resubmitted"
	case model.StockLossActionApprove:
		return "Approved"
	case model.StockLossActionReject:
		return "Rejected"
	default:
		return ""
	}
}

// buildStockLossErpRemarks 构建报损单 ERP remarks（倒序排列批注）
func buildStockLossErpRemarks(annotations []*model.StockLossAnnotation, currentAction string, currentContent string, currentTime int) string {
	var lines []string

	// 先添加当前操作（最新的在最前面）
	actionName := getStockLossAnnotationErpActionName(currentAction)
	if actionName != "" {
		timeStr := time.Unix(int64(currentTime), 0).Format("2006-01-02 15:04")
		lines = append(lines, actionName+" "+timeStr)
		if currentContent != "" {
			// submit/resubmit 操作的内容是报损原因，需要添加前缀
			content := currentContent
			if currentAction == model.StockLossActionSubmit || currentAction == model.StockLossActionResubmit {
				content = "Loss Reason: " + content
			}
			lines = append(lines, content)
		}
	}

	// 正序遍历历史批注（数据库已按 create_time DESC 排序，最新在前）
	for _, annotation := range annotations {
		actionName := getStockLossAnnotationErpActionName(annotation.Action)
		if actionName == "" {
			continue
		}
		timeStr := time.Unix(int64(annotation.CreateTime), 0).Format("2006-01-02 15:04")
		lines = append(lines, actionName+" "+timeStr)
		if annotation.Content != "" {
			// submit/resubmit 操作的内容是报损原因，需要添加前缀
			content := annotation.Content
			if annotation.Action == model.StockLossActionSubmit || annotation.Action == model.StockLossActionResubmit {
				content = "Loss Reason: " + content
			}
			lines = append(lines, content)
		}
	}

	return strings.Join(lines, "\n")
}

// parseErpErrorForStockLoss 解析 ERP 错误，返回友好的错误信息
// 支持两种错误模式：
// 1. "Item XXX is not active or end of life has been reached" -> 物品状态已关闭
// 2. "XXX is not a stock Item" -> 物品未找到
func (s *stockLossSrv) parseErpErrorForStockLoss(db *gorm.DB, errMsg string, defaultLang string) error {
	// 模式1: Item XXX is not active or end of life has been reached
	notActivePattern := regexp.MustCompile(`Item\s+(\S+)\s+is not active`)
	if matches := notActivePattern.FindStringSubmatch(errMsg); len(matches) > 1 {
		erpCode := matches[1]
		name, internalCode := s.getMaterialInfoByCode(db, erpCode, defaultLang)
		displayName := formatMaterialNameForError(name, internalCode)
		// 翻译: 物品XXX状态已关闭，请修改物品状态（使用现有翻译键）
		return errors.New(i18n.Translate(defaultLang, "物品%s状态已关闭，请修改物品状态", displayName))
	}

	// 模式2: XXX is not a stock Item
	notStockPattern := regexp.MustCompile(`(\S+)\s+is not a stock Item`)
	if matches := notStockPattern.FindStringSubmatch(errMsg); len(matches) > 1 {
		erpCode := matches[1]
		name, internalCode := s.getMaterialInfoByCode(db, erpCode, defaultLang)
		displayName := formatMaterialNameForError(name, internalCode)
		// 翻译: 物品XXX未找到
		return errors.New(i18n.Translate(defaultLang, "物品%s未找到", displayName))
	}

	// 无法解析，返回原始错误
	return errors.New(i18n.Translate(defaultLang, "审核报损单失败") + ": " + errMsg)
}

// getMaterialInfoByCode 根据物品ERP编码获取物品名称和内部编码
func (s *stockLossSrv) getMaterialInfoByCode(db *gorm.DB, erpCode string, defaultLang string) (name string, internalCode string) {
	materialRepo := repository.NewMaterialRepo(db)
	material, err := materialRepo.GetMaterialByCode(erpCode)
	if err != nil || material.Uuid == 0 {
		return "", ""
	}
	locale := language.JsonToLocaleResponse(material.Name)
	name = locale.GetLocale(defaultLang)
	if name == "" {
		name = material.Name
	}
	return name, material.InternalCode
}

// formatMaterialNameForError 格式化物品名称用于错误信息，格式：名称（内部编码）
func formatMaterialNameForError(name string, internalCode string) string {
	if internalCode != "" {
		return name + "（" + internalCode + "）"
	}
	return name
}
