package service

import (
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	shop_req "ttpos-server-go/app/dto/req"
	setting_resp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	settingSrv "ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DataManageSrv 数据管理服务
type DataManageSrv struct {
	dbm        *database.DBManager
	settingSrv settingSrv.ISrv
}

// IDataManageSrv 数据管理服务接口
type IDataManageSrv interface {
	GetDataManage(ctx context.Context) (*setting_resp.GetDataManageResp, error) // 获取数据管理信息
	SetDataManage(ctx context.Context, req shop_req.SetDataManageReq) error     // 设置数据管理

	// 新增：拆分接口
	SaveStatus(ctx context.Context, req shop_req.SaveDataManageStatusReq) error                                                                   // 仅保存数据管理开关状态
	SetStaff(ctx context.Context, req shop_req.SetDataManageStaffReq) error                                                                       // 设置操作人员（立即生效）
	GetOrderList(ctx context.Context, req shop_req.GetDataManageOrderListReq) (*setting_resp.DataManageOrderListResp, error)                      // 获取已选订单列表
	RestoreOrder(ctx context.Context, req shop_req.RestoreDataManageOrderReq) error                                                               // 恢复（移除）单条已选订单
	GetOrderSelect(ctx context.Context, req shop_req.GetDataManageOrderSelectReq) (*setting_resp.DataManageOrderSelectResp, error)                // 获取可选订单列表
	SubmitOrder(ctx context.Context, req shop_req.SubmitDataManageOrderReq) error                                                                 // 提交订单选择
	GetOrderSelectStats(ctx context.Context, req shop_req.GetDataManageOrderSelectStatsReq) (*setting_resp.DataManageOrderSelectStatsResp, error) // 获取可选订单统计预览
}

// NewDataManageSrvImpl 创建数据管理服务
func NewDataManageSrv(dbm *database.DBManager, settingSrv settingSrv.ISrv) IDataManageSrv {
	return NewDataManageSrvImpl(dbm, settingSrv)
}

// NewDataManageSrvImpl 创建数据管理服务实现
func NewDataManageSrvImpl(dbm *database.DBManager, settingSrv settingSrv.ISrv) IDataManageSrv {
	return &DataManageSrv{
		dbm:        dbm,
		settingSrv: settingSrv,
	}
}

// GetDataManage 获取数据管理
func (s *DataManageSrv) GetDataManage(ctx context.Context) (*setting_resp.GetDataManageResp, error) {
	// 检查权限
	err := s.checkPermission(ctx)
	if err != nil {
		return nil, err
	}

	// 获取数据管理信息
	setting := s.settingSrv.GetDataManageSetting(ctx)
	db := s.dbm.GetDB(ctx.GetDbId())
	commonRepo := repository.NewCommonRepo()
	staffRepo := repository.NewStaffRepo(db)
	dataManageRepo := repository.NewDataManageRepo(db)
	statisticsRepo := repository.NewStatisticsRepo(db)

	staffUuids := []uint64{}
	orderUuids := []uint64{}

	// 获取操作人员数量
	staffs := staffRepo.GetStaffs(
		staffRepo.WhereIsSuper(0),
		staffRepo.WhereHasDataPermission(1),
		commonRepo.WhereBySoftDelete(),
	)
	for _, staff := range staffs {
		staffUuids = append(staffUuids, staff.Uuid)
	}

	// 获取订单数量
	var orderCount int
	if utils.CompareVersion(ctx.GetVersion(), utils.VersionGTE, constant.ClientVersionV2200) {
		// >= 2.20 仅统计数量，不返回 UUID 列表
		count, _ := dataManageRepo.Count(dataManageRepo.WhereByType(model.DataManageTypeOrder))
		orderCount = int(count)
	} else {
		dataManages := dataManageRepo.List(
			dataManageRepo.WhereByType(model.DataManageTypeOrder),
		)
		for _, dataManage := range dataManages {
			orderUuids = append(orderUuids, dataManage.DataUuid)
		}
		orderCount = len(orderUuids)
	}

	// 获取统计信息
	statisticsData := statisticsRepo.CountSale(
		commonRepo.WhereInDataManageSubQuery(db, "sale_bill_uuid",
			commonRepo.WhereByType(model.DataManageTypeOrder),
			commonRepo.WhereBySoftDelete(),
		),
	)
	// 总优惠折扣率 = 总优惠折扣 / 总销售额
	var discountRatio decimal.Decimal
	if statisticsData.TotalSaleAmount.Float64 > 0 {
		discountRatio = decimal.NewFromFloat(statisticsData.TotalDiscount.Float64).
			Div(decimal.NewFromFloat(statisticsData.TotalSaleAmount.Float64)).
			Mul(decimal.NewFromInt(100))
	}

	return &setting_resp.GetDataManageResp{
		IsEnableDataManage: setting.IsEnableDataManage,
		StaffCount:         len(staffUuids),
		OrderCount:         orderCount,
		StaffUuids:         staffUuids,
		OrderUuids:         orderUuids,
		Statistics: setting_resp.DataManageStatistics{
			SaleAmount:     statisticsData.TotalSaleAmount.Float64,
			ReceivedPrice:  statisticsData.TotalReceivedAmount.Float64,
			ProductCount:   statisticsData.TotalProductNum.Float64,
			DiscountMember: statisticsData.TotalDiscountMember.Float64,
			BusinessAmount: statisticsData.TotalBusinessAmount.Float64,
			ServiceFee:     statisticsData.TotalServiceFee.Float64,
			PaymentFee:     statisticsData.TotalPaymentFee.Float64,
			Tax:            statisticsData.TotalTax.Float64,
			RefundAmount:   statisticsData.TotalRefundAmount.Float64,
			Discount:       statisticsData.TotalDiscount.Float64,
			DiscountRatio:  discountRatio.Round(2).InexactFloat64(),
			GiveAmount:     statisticsData.TotalGiftAmount.Float64,
			GiveCount:      statisticsData.TotalGiftNum.Float64,
			FreeAmount:     statisticsData.TotalFreeAmount.Float64,
			FreeCount:      statisticsData.TotalFreeNum.Float64,
		},
	}, nil
}

// SetDataManage 设置数据管理
func (s *DataManageSrv) SetDataManage(ctx context.Context, req shop_req.SetDataManageReq) error {
	// 检查权限
	err := s.checkPermission(ctx)
	if err != nil {
		return err
	}

	// 设置数据管理
	if !req.IsEnableDataManage && len(req.SaleBillUuids) > 0 {
		return errors.New("已选择订单，不可关闭")
	}

	// 获取数据管理设置
	setting := s.settingSrv.GetDataManageSetting(ctx)

	db := s.dbm.GetDB(ctx.GetDbId())
	err = db.Transaction(func(tx *gorm.DB) error {
		staffRepo := repository.NewStaffRepo(tx)
		dataManageRepo := repository.NewDataManageRepo(tx)

		// 更新公司数据管理状态
		setting.IsEnableDataManage = req.IsEnableDataManage
		if err := s.settingSrv.UpdateSetting(ctx, constant.SettingDataManage, setting); err != nil {
			return err
		}

		// 更新员工数据管理权限
		err = staffRepo.Updates(map[string]any{"has_data_permission": 0}, staffRepo.WhereHasDataPermission(1))
		if err != nil {
			return err
		}
		if len(req.StaffUuids) > 0 {
			err = staffRepo.Updates(map[string]any{"has_data_permission": 1}, staffRepo.WhereUuids(req.StaffUuids))
			if err != nil {
				return err
			}
		}

		// 增量更新数据管理数据
		commonRepo := repository.NewCommonRepo()
		existingUuids := dataManageRepo.GetDataUuids(
			dataManageRepo.WhereByType(model.DataManageTypeOrder),
			commonRepo.WhereBySoftDelete(),
		)

		existingSet := make(map[uint64]struct{}, len(existingUuids))
		for _, uid := range existingUuids {
			existingSet[uid] = struct{}{}
		}
		newSet := make(map[uint64]struct{}, len(req.SaleBillUuids))
		for _, uid := range req.SaleBillUuids {
			newSet[uid] = struct{}{}
		}

		// 需要删除的（旧有新无）
		toDelete := make([]uint64, 0)
		for _, uid := range existingUuids {
			if _, ok := newSet[uid]; !ok {
				toDelete = append(toDelete, uid)
			}
		}
		// 需要新增的（新有旧无）
		toAdd := make([]uint64, 0)
		for _, uid := range req.SaleBillUuids {
			if _, ok := existingSet[uid]; !ok {
				toAdd = append(toAdd, uid)
			}
		}

		// 批量删除
		const deleteBatchSize = 1000
		for i := 0; i < len(toDelete); i += deleteBatchSize {
			end := min(i+deleteBatchSize, len(toDelete))
			if err = dataManageRepo.Delete(
				dataManageRepo.WhereByType(model.DataManageTypeOrder),
				dataManageRepo.WhereInDataUuids(toDelete[i:end]),
			); err != nil {
				return err
			}
		}

		// 批量新增
		if len(toAdd) > 0 {
			staff := ctx.GetStaff()
			now := time.Now().Unix()
			dataManages := make([]*model.DataManage, 0, len(toAdd))
			for _, saleBillUuid := range toAdd {
				uuid, err := utils.GetID()
				if err != nil {
					return err
				}
				dataManages = append(dataManages, &model.DataManage{
					BaseModel: model.BaseModel{
						Uuid:       uuid,
						CreateTime: now,
						UpdateTime: now,
					},
					Type:      model.DataManageTypeOrder,
					DataUuid:  saleBillUuid,
					StaffUuid: staff.Uuid,
				})
			}
			if err = dataManageRepo.Creates(dataManages); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		logger.Logger.Error("设置数据管理失败", zap.Error(err))
		return errors.New("保存失败")
	}

	return nil
}

// SaveStatus 仅保存数据管理开关状态
func (s *DataManageSrv) SaveStatus(ctx context.Context, req shop_req.SaveDataManageStatusReq) error {
	if err := s.checkPermission(ctx); err != nil {
		return err
	}

	// 如果要关闭，检查是否还有已选订单
	if !req.IsEnableDataManage {
		db := s.dbm.GetDB(ctx.GetDbId())
		dataManageRepo := repository.NewDataManageRepo(db)
		count, err := dataManageRepo.Count(dataManageRepo.WhereByType(model.DataManageTypeOrder))
		if err != nil {
			return errors.New("查询失败")
		}
		if count > 0 {
			return errors.New("已选择订单，不可关闭")
		}
	}

	setting := s.settingSrv.GetDataManageSetting(ctx)
	setting.IsEnableDataManage = req.IsEnableDataManage
	if err := s.settingSrv.UpdateSetting(ctx, constant.SettingDataManage, setting); err != nil {
		logger.Logger.Error("保存数据管理状态失败", zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Error(err))
		return errors.New("保存失败")
	}
	return nil
}

// SetStaff 设置操作人员（立即生效）
func (s *DataManageSrv) SetStaff(ctx context.Context, req shop_req.SetDataManageStaffReq) error {
	if err := s.checkPermission(ctx); err != nil {
		return err
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	return db.Transaction(func(tx *gorm.DB) error {
		staffRepo := repository.NewStaffRepo(tx)

		// 先清除所有员工的数据管理权限
		if err := staffRepo.Updates(map[string]any{"has_data_permission": 0}, staffRepo.WhereHasDataPermission(1)); err != nil {
			return err
		}
		// 再设置选中的员工
		if len(req.StaffUuids) > 0 {
			if err := staffRepo.Updates(map[string]any{"has_data_permission": 1}, staffRepo.WhereUuids(req.StaffUuids)); err != nil {
				return err
			}
		}
		return nil
	})
}

// GetOrderList 获取已选订单列表（分页）
func (s *DataManageSrv) GetOrderList(ctx context.Context, req shop_req.GetDataManageOrderListReq) (*setting_resp.DataManageOrderListResp, error) {
	if err := s.checkPermission(ctx); err != nil {
		return nil, err
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	orderRepo := repository.NewOrderRepo(db)
	commonRepo := repository.NewCommonRepo()
	tz := ctx.GetCompanySetting().Timezone

	// 复用收银台订单列表查询（含完整 Preload 和计算逻辑）
	reqs := repository.GetCashierOrderListWithPaginationType{
		PageNo:           req.PageNo,
		PageSize:         req.PageSize,
		OrderNo:          req.OrderNo,
		DateType:         req.DateType,
		QueryStartDate:   req.QueryStartDate,
		QueryEndDate:     req.QueryEndDate,
		BillType:         req.BillType,
		Status:           -1,
		DiningMethod:     -1,
		IsOnlyDataManage: 1, // 仅查询 data_manage 中的订单（子查询模式）
	}
	lists, total, orderListDBOption, err := orderRepo.GetCashierOrderListWithPagination(reqs, tz)
	if err != nil {
		return nil, errors.New("查询失败")
	}

	// 统计当前筛选条件下已选订单的数量和实付金额（与订单列表筛选保持一致）
	dataManageSubQuery := commonRepo.WhereInDataManageSubQuery(db, "uuid",
		commonRepo.WhereByType(model.DataManageTypeOrder),
		commonRepo.WhereBySoftDelete(),
	)
	statsOpts := []repository.DBOption{
		commonRepo.WhereBySoftDelete(),
		commonRepo.WhereByCooking(),
		commonRepo.WhereInBillType([]uint{constant.SaleBillTypeDesk, constant.SaleBillTypeInstant}),
		dataManageSubQuery,
	}
	if orderListDBOption != nil {
		statsOpts = append(statsOpts, orderListDBOption)
	}
	orderCount, paidAmount := orderRepo.CountAndSumSaleBill(statsOpts...)
	totalPaidAmount := decimal.NewFromFloat(paidAmount).Sub(decimal.NewFromFloat(orderRepo.SumSaleBillRefundAmount(statsOpts...))).Round(2).InexactFloat64()

	// 构建响应
	list := make([]setting_resp.DataManageOrderItem, 0, len(lists))
	for _, bill := range lists {
		list = append(list, s.buildOrderItem(bill, true))
	}

	return &setting_resp.DataManageOrderListResp{
		List: list,
		Meta: setting_resp.DataManageOrderListMeta{
			PageResponse: dto.PageResponse{PageNo: req.PageNo, PageSize: req.PageSize, Total: total},
			OrderCount:   orderCount,
			PaidAmount:   totalPaidAmount,
		},
	}, nil
}

// RestoreOrder 恢复（移除）单条已选订单
func (s *DataManageSrv) RestoreOrder(ctx context.Context, req shop_req.RestoreDataManageOrderReq) error {
	if err := s.checkPermission(ctx); err != nil {
		return err
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	dataManageRepo := repository.NewDataManageRepo(db)

	// 检查记录是否存在
	if !dataManageRepo.ExistsByDataUuid(model.DataManageTypeOrder, req.SaleBillUuid) {
		return errors.New("该订单未被选中")
	}

	if err := dataManageRepo.DeleteByDataUuid(model.DataManageTypeOrder, req.SaleBillUuid); err != nil {
		logger.Logger.Error("恢复订单失败", zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Error(err))
		return errors.New("操作失败")
	}
	return nil
}

// GetOrderSelect 获取可选订单列表（筛选+分页）
func (s *DataManageSrv) GetOrderSelect(ctx context.Context, req shop_req.GetDataManageOrderSelectReq) (*setting_resp.DataManageOrderSelectResp, error) {
	if err := s.checkPermission(ctx); err != nil {
		return nil, err
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	orderRepo := repository.NewOrderRepo(db)
	tz := ctx.GetCompanySetting().Timezone

	// 复用收银台订单列表查询（含完整 Preload 和计算逻辑）
	reqs := repository.GetCashierOrderListWithPaginationType{
		PageNo:              req.PageNo,
		PageSize:            req.PageSize,
		OrderNo:             req.OrderNo,
		DateType:            req.DateType,
		QueryStartDate:      req.QueryStartDate,
		QueryEndDate:        req.QueryEndDate,
		BillType:            req.BillType,
		Status:              1, // 仅已完成订单
		DiningMethod:        -1,
		IsContainDataManage: 1, // 包含 data_manage 订单（通过 DataManage Preload 判断选中状态）
	}
	lists, total, _, err := orderRepo.GetCashierOrderListWithPagination(reqs, tz)
	if err != nil {
		return nil, errors.New("查询失败")
	}

	// 构建响应（通过 DataManage 关联判断选中状态）
	list := make([]setting_resp.DataManageOrderItem, 0, len(lists))
	for _, bill := range lists {
		isSelected := bill.DataManage != nil
		list = append(list, s.buildOrderItem(bill, isSelected))
	}

	return &setting_resp.DataManageOrderSelectResp{
		List: list,
		Meta: setting_resp.DataManageOrderSelectMeta{
			PageResponse: dto.PageResponse{PageNo: req.PageNo, PageSize: req.PageSize, Total: total},
		},
	}, nil
}

// SubmitOrder 提交订单选择（多 Filter 独立处理）
// 每个 Filter 独立决定操作：
//   - SelectAll=true：筛选范围内全部加入，DeselectedUuids 排除
//   - SelectAll=false + SelectedUuids：新增指定订单
//   - SelectAll=false + DeselectedUuids：移除指定订单
func (s *DataManageSrv) SubmitOrder(ctx context.Context, req shop_req.SubmitDataManageOrderReq) error {
	if err := s.checkPermission(ctx); err != nil {
		return err
	}

	if len(req.Filters) == 0 {
		return nil
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	commonRepo := repository.NewCommonRepo()
	orderRepo := repository.NewOrderRepo(db)
	tz := ctx.GetCompanySetting().Timezone

	// 收集所有 Filter 的新增和移除 UUID（跨 Filter 去重）
	toAddSet := make(map[uint64]struct{})
	toDeleteSet := make(map[uint64]struct{})

	for _, f := range req.Filters {
		if f.SelectAll {
			// 案例1/2：全选模式，筛选范围全部 - DeselectedUuids 新增到数据管理
			filterOpt := s.buildOrderFilterOption(f.OrderNo, f.DateType, f.QueryStartDate, f.QueryEndDate, f.BillType, tz)
			allUuids := orderRepo.GetSaleBillUuids(
				commonRepo.WhereBySoftDelete(),
				commonRepo.WhereByCooking(),
				func(db *gorm.DB) *gorm.DB { return db.Where("status = ?", 1) },
				filterOpt,
			)
			deselectedSet := make(map[uint64]struct{}, len(f.DeselectedUuids))
			for _, uid := range f.DeselectedUuids {
				deselectedSet[uid] = struct{}{}
			}
			for _, uid := range allUuids {
				if _, excluded := deselectedSet[uid]; !excluded {
					toAddSet[uid] = struct{}{}
				}
			}
		} else {
			if len(f.DeselectedUuids) > 0 {
				// 案例3：从数据管理中移除 DeselectedUuids
				for _, uid := range f.DeselectedUuids {
					toDeleteSet[uid] = struct{}{}
				}
			}
			if len(f.SelectedUuids) > 0 {
				// 案例4：手动选中新增到数据管理
				for _, uid := range f.SelectedUuids {
					toAddSet[uid] = struct{}{}
				}
			}
		}
	}

	// 冲突处理：同一 UUID 同时出现在新增和移除中，以移除优先
	for uid := range toDeleteSet {
		delete(toAddSet, uid)
	}

	if len(toAddSet) == 0 && len(toDeleteSet) == 0 {
		return nil
	}

	// 查询 toAddSet 中哪些已存在于 data_manage，避免重复插入（分批查询避免 IN 子句过大）
	toAddUuids := make([]uint64, 0, len(toAddSet))
	for uid := range toAddSet {
		toAddUuids = append(toAddUuids, uid)
	}

	dataManageRepo := repository.NewDataManageRepo(db)
	const queryBatchSize = 1000
	for i := 0; i < len(toAddUuids); i += queryBatchSize {
		end := min(i+queryBatchSize, len(toAddUuids))
		alreadyExist := dataManageRepo.GetDataUuids(
			dataManageRepo.WhereByType(model.DataManageTypeOrder),
			dataManageRepo.WhereInDataUuids(toAddUuids[i:end]),
		)
		for _, uid := range alreadyExist {
			delete(toAddSet, uid)
		}
	}

	// 最终需要新增的 UUID
	finalToAdd := make([]uint64, 0, len(toAddSet))
	for uid := range toAddSet {
		finalToAdd = append(finalToAdd, uid)
	}

	// 最终需要删除的 UUID
	toDeleteUuids := make([]uint64, 0, len(toDeleteSet))
	for uid := range toDeleteSet {
		toDeleteUuids = append(toDeleteUuids, uid)
	}

	if len(finalToAdd) == 0 && len(toDeleteUuids) == 0 {
		return nil
	}

	// 事务内批量处理
	return db.Transaction(func(tx *gorm.DB) error {
		txDataManageRepo := repository.NewDataManageRepo(tx)

		// 批量删除
		if len(toDeleteUuids) > 0 {
			if err := txDataManageRepo.BatchDeleteByDataUuids(model.DataManageTypeOrder, toDeleteUuids); err != nil {
				return err
			}
		}

		// 批量新增（分批插入，每批 500 条）
		if len(finalToAdd) > 0 {
			staff := ctx.GetStaff()
			now := time.Now().Unix()
			const batchSize = 500
			for i := 0; i < len(finalToAdd); i += batchSize {
				end := min(i+batchSize, len(finalToAdd))
				batch := make([]*model.DataManage, 0, end-i)
				for _, saleBillUuid := range finalToAdd[i:end] {
					uuid, err := utils.GetID()
					if err != nil {
						return err
					}
					batch = append(batch, &model.DataManage{
						BaseModel: model.BaseModel{
							Uuid:       uuid,
							CreateTime: now,
							UpdateTime: now,
						},
						Type:      model.DataManageTypeOrder,
						DataUuid:  saleBillUuid,
						StaffUuid: staff.Uuid,
					})
				}
				if err := txDataManageRepo.Creates(batch); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// GetOrderSelectStats 获取可选订单统计预览（不持久化）
// 只统计最后一个筛选条件范围内的订单，前面的 filters 仅贡献选择状态。
// 如果前面 filters 选中的订单也落在最后一个 filter 范围内，则合并计入。
func (s *DataManageSrv) GetOrderSelectStats(ctx context.Context, req shop_req.GetDataManageOrderSelectStatsReq) (*setting_resp.DataManageOrderSelectStatsResp, error) {
	if err := s.checkPermission(ctx); err != nil {
		return nil, err
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	commonRepo := repository.NewCommonRepo()
	orderRepo := repository.NewOrderRepo(db)
	dataManageRepo := repository.NewDataManageRepo(db)
	tz := ctx.GetCompanySetting().Timezone
	filters := s.normalizeOrderSelectStatsFilter(req)

	// 1) 所有 filters 贡献的增删集合
	toAddSet, toDeleteSet := s.buildOrderSelectOperationSets(filters, orderRepo, commonRepo, tz)

	// 2) 仅最后一个 filter 定义展示范围和基线
	lastFilter := []shop_req.DataManageOrderSubmitFilter{filters[len(filters)-1]}
	existingSet := s.buildOrderSelectExistingSet(lastFilter, db, orderRepo, dataManageRepo, commonRepo, tz)

	// 3) 最终选中集合 = existing + toAddSet - toDeleteSet
	finalSet := make(map[uint64]struct{}, len(existingSet)+len(toAddSet))
	for uid := range existingSet {
		finalSet[uid] = struct{}{}
	}
	for uid := range toAddSet {
		finalSet[uid] = struct{}{}
	}
	for uid := range toDeleteSet {
		delete(finalSet, uid)
	}

	// 4) 限定到最后一个 filter 的范围：只保留命中该筛选条件的订单
	lf := filters[len(filters)-1]
	scopeOpt := s.buildOrderFilterOption(lf.OrderNo, lf.DateType, lf.QueryStartDate, lf.QueryEndDate, lf.BillType, tz)
	scopeUuids := orderRepo.GetSaleBillUuids(
		commonRepo.WhereBySoftDelete(),
		commonRepo.WhereByCooking(),
		func(db *gorm.DB) *gorm.DB { return db.Where("status = ?", 1) },
		scopeOpt,
	)
	scopeSet := make(map[uint64]struct{}, len(scopeUuids))
	for _, uid := range scopeUuids {
		scopeSet[uid] = struct{}{}
	}
	for uid := range finalSet {
		if _, inScope := scopeSet[uid]; !inScope {
			delete(finalSet, uid)
		}
	}

	// 5) 统计
	return s.calcOrderSelectStats(finalSet, orderRepo, commonRepo), nil
}

// normalizeOrderSelectStatsFilter 规范化统计筛选条件：
// 有 Filters 时直接使用，无则默认近7天
func (s *DataManageSrv) normalizeOrderSelectStatsFilter(req shop_req.GetDataManageOrderSelectStatsReq) []shop_req.DataManageOrderSubmitFilter {
	if len(req.Filters) > 0 {
		return req.Filters
	}
	return []shop_req.DataManageOrderSubmitFilter{
		{
			DateType: constant.OrderDateTypeLastWeek,
			BillType: constant.OrderDateTypeAll,
		},
	}
}

// buildOrderSelectOperationSets 计算本次请求的新增/移除集合
func (s *DataManageSrv) buildOrderSelectOperationSets(filters []shop_req.DataManageOrderSubmitFilter, orderRepo repository.IOrderRepo, commonRepo repository.ICommonRepo, tz string) (map[uint64]struct{}, map[uint64]struct{}) {
	toAddSet := make(map[uint64]struct{})
	toDeleteSet := make(map[uint64]struct{})

	for _, f := range filters {
		if f.SelectAll {
			filterOpt := s.buildOrderFilterOption(f.OrderNo, f.DateType, f.QueryStartDate, f.QueryEndDate, f.BillType, tz)
			allUuids := orderRepo.GetSaleBillUuids(
				commonRepo.WhereBySoftDelete(),
				commonRepo.WhereByCooking(),
				func(db *gorm.DB) *gorm.DB { return db.Where("status = ?", 1) },
				filterOpt,
			)
			deselectedSet := make(map[uint64]struct{}, len(f.DeselectedUuids))
			for _, uid := range f.DeselectedUuids {
				deselectedSet[uid] = struct{}{}
			}
			for _, uid := range allUuids {
				if _, excluded := deselectedSet[uid]; !excluded {
					toAddSet[uid] = struct{}{}
				}
			}
			continue
		}

		for _, uid := range f.DeselectedUuids {
			toDeleteSet[uid] = struct{}{}
		}
		for _, uid := range f.SelectedUuids {
			toAddSet[uid] = struct{}{}
		}
	}

	// 冲突处理：同一 UUID 同时新增和移除时，以移除优先
	for uid := range toDeleteSet {
		delete(toAddSet, uid)
	}

	return toAddSet, toDeleteSet
}

// buildOrderSelectExistingSet 构建基线已选集合
func (s *DataManageSrv) buildOrderSelectExistingSet(filters []shop_req.DataManageOrderSubmitFilter, db *gorm.DB, orderRepo repository.IOrderRepo, dataManageRepo repository.IDataManageRepo, commonRepo repository.ICommonRepo, tz string) map[uint64]struct{} {
	existingSet := make(map[uint64]struct{})

	// 无筛选时，回退为全部已持久化订单（兼容旧行为）
	if len(filters) == 0 {
		existingUuids := dataManageRepo.GetDataUuids(
			dataManageRepo.WhereByType(model.DataManageTypeOrder),
			commonRepo.WhereBySoftDelete(),
		)
		for _, uid := range existingUuids {
			existingSet[uid] = struct{}{}
		}
		return existingSet
	}

	// 有筛选时：按每个 filter 查询“已持久化且命中筛选”的并集
	dataManageSubQuery := commonRepo.WhereInDataManageSubQuery(db, "uuid",
		commonRepo.WhereByType(model.DataManageTypeOrder),
		commonRepo.WhereBySoftDelete(),
	)
	for _, f := range filters {
		filterOpt := s.buildOrderFilterOption(f.OrderNo, f.DateType, f.QueryStartDate, f.QueryEndDate, f.BillType, tz)
		filteredExistingUuids := orderRepo.GetSaleBillUuids(
			commonRepo.WhereBySoftDelete(),
			commonRepo.WhereByCooking(),
			func(db *gorm.DB) *gorm.DB { return db.Where("status = ?", 1) },
			dataManageSubQuery,
			filterOpt,
		)
		for _, uid := range filteredExistingUuids {
			existingSet[uid] = struct{}{}
		}
	}

	return existingSet
}

// calcOrderSelectStats 按最终集合统计数量与实付金额
func (s *DataManageSrv) calcOrderSelectStats(finalSet map[uint64]struct{}, orderRepo repository.IOrderRepo, commonRepo repository.ICommonRepo) *setting_resp.DataManageOrderSelectStatsResp {
	if len(finalSet) == 0 {
		return &setting_resp.DataManageOrderSelectStatsResp{
			SelectedCount: 0,
			PaidAmount:    0,
			TotalCount:    0,
			IsSelectAll:   false,
			SelectedUuids: make([]uint64, 0),
		}
	}

	finalUuids := make([]uint64, 0, len(finalSet))
	for uid := range finalSet {
		finalUuids = append(finalUuids, uid)
	}

	var selectedCount int64
	totalPaidAmountDec := decimal.NewFromFloat(0)
	const statsBatchSize = 1000
	for i := 0; i < len(finalUuids); i += statsBatchSize {
		end := min(i+statsBatchSize, len(finalUuids))
		count, paidAmount := orderRepo.CountAndSumSaleBill(commonRepo.WhereInUuids(finalUuids[i:end]), commonRepo.WhereBySoftDelete(), commonRepo.WhereByCooking())
		refundAmount := orderRepo.SumSaleBillRefundAmount(commonRepo.WhereInUuids(finalUuids[i:end]), commonRepo.WhereByCooking())
		selectedCount += count
		totalPaidAmountDec = totalPaidAmountDec.Add(decimal.NewFromFloat(paidAmount).Sub(decimal.NewFromFloat(refundAmount)))
	}
	totalPaidAmount := totalPaidAmountDec.Round(2).InexactFloat64()

	return &setting_resp.DataManageOrderSelectStatsResp{
		SelectedCount: selectedCount,
		PaidAmount:    totalPaidAmount,
		TotalCount:    selectedCount,
		IsSelectAll:   false,
		SelectedUuids: finalUuids,
	}
}

// buildOrderItem 构建订单列表项
func (s *DataManageSrv) buildOrderItem(order model.SaleBill, isSelected bool) setting_resp.DataManageOrderItem {
	return setting_resp.DataManageOrderItem{
		SaleBillUuid:  order.Uuid,
		OrderNo:       order.OrderNo,
		CreateTime:    order.CreateTime,
		Amount:        order.Amount,
		PaymentAmount: order.GetPaymentAmount(),
		PaymentMethod: s.buildPaymentMethodNames(order),
		IsSelected:    isSelected,
	}
}

// buildPaymentMethodNames 提取去重的支付方式名称
func (s *DataManageSrv) buildPaymentMethodNames(order model.SaleBill) string {
	seen := make(map[string]struct{})
	names := make([]string, 0)
	for _, so := range order.SaleOrders {
		for _, po := range so.PaymentOrders {
			name := po.PaymentMethodName
			if po.PaymentMethod != nil && po.PaymentMethod.Name != "" {
				name = po.PaymentMethod.Name
			}
			if name == "" {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			names = append(names, name)
		}
	}
	return strings.Join(names, ", ")
}

// buildOrderFilterOption 构建订单筛选条件
func (s *DataManageSrv) buildOrderFilterOption(orderNo string, dateType int, queryStartDate string, queryEndDate string, billType int, tz string) repository.DBOption {
	return func(db *gorm.DB) *gorm.DB {
		// 订单编号搜索
		if orderNo != "" {
			db = db.Where("order_no LIKE ?", "%"+orderNo+"%")
		}
		// 账单类型：指定类型精确匹配，否则限制为堂食+点餐（排除外送）
		if billType != -1 {
			db = db.Where("bill_type = ?", billType)
		} else {
			db = db.Where("bill_type IN (?)", []uint{constant.SaleBillTypeDesk, constant.SaleBillTypeInstant})
		}
		// 日期类型
		var startTime, endTime int64
		if dateType >= constant.OrderDateTypeToday && dateType <= constant.OrderDateTypeLastMonth {
			switch dateType {
			case constant.OrderDateTypeToday:
				startTime, endTime, _ = utils.SetTimezone(tz).GetTimeRange(utils.DayTypeToday)
			case constant.OrderDateTypeYesterday:
				startTime, endTime, _ = utils.SetTimezone(tz).GetTimeRange(utils.DayTypeYesterday)
			case constant.OrderDateTypeWeek:
				startTime, endTime, _ = utils.SetTimezone(tz).GetTimeRange(utils.DayTypeThisWeek)
			case constant.OrderDateTypeMonth:
				startTime, endTime, _ = utils.SetTimezone(tz).GetTimeRange(utils.DayTypeThisMonth)
			case constant.OrderDateTypeYear:
				startTime, endTime, _ = utils.SetTimezone(tz).GetTimeRange(utils.DayTypeThisYear)
			case constant.OrderDateTypeLastWeek:
				// 数据管理场景“近7天”按含今天口径：today-6d 00:00:00 ~ today 23:59:59
				startTime, endTime = utils.SetTimezone(tz).Last7DaysStartEndUnix()
			case constant.OrderDateTypeLastMonth:
				startTime, endTime, _ = utils.SetTimezone(tz).GetTimeRange(utils.DayTypeLastMonth)
			}
		}
		// 自定义日期范围
		if queryStartDate != "" && queryEndDate != "" {
			timeUtil := utils.SetTimezone(tz)
			if st, err := timeUtil.FormatDateTimeToUnix(queryStartDate); err == nil {
				startTime = st
			}
			if et, err := timeUtil.FormatDateTimeToUnix(queryEndDate); err == nil {
				endTime = et
			}
		}
		if startTime > 0 && endTime > 0 {
			db = db.Where("create_time BETWEEN ? AND ?", startTime, endTime)
		} else if startTime > 0 {
			db = db.Where("create_time >= ?", startTime)
		} else if endTime > 0 {
			db = db.Where("create_time <= ? AND create_time > 0", endTime)
		}
		return db
	}
}

// checkPermission 检查权限
func (s *DataManageSrv) checkPermission(ctx context.Context) error {
	// 云平台是否开启数据管理
	setting := ctx.GetCompanySetting()
	if setting.EnableDataManagement == 0 {
		return errors.New("无此功能权限")
	}

	// 检查当前员工是否具有数据管理权限，超级管理员默认具有数据管理权限，其他员工需要具有数据管理权限
	staff := ctx.GetStaff()
	if staff.IsSuper == 1 {
		return nil
	}

	if staff.HasDataPermission == 0 {
		return errors.New("无此功能权限")
	}

	return nil
}
