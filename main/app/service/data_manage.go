package service

import (
	"time"
	"ttpos-server-go/app/constant"
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

// IDataManageSrv数据管理服务接口
type IDataManageSrv interface {
	GetDataManage(ctx context.Context) (*setting_resp.GetDataManageResp, error) // 获取数据管理信息
	SetDataManage(ctx context.Context, req shop_req.SetDataManageReq) error     // 设置数据管理
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
	dataManages := dataManageRepo.List(
		dataManageRepo.WhereByType(model.DataManageTypeOrder),
	)
	for _, dataManage := range dataManages {
		orderUuids = append(orderUuids, dataManage.DataUuid)
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
		OrderCount:         len(orderUuids),
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
			end := i + deleteBatchSize
			if end > len(toDelete) {
				end = len(toDelete)
			}
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
