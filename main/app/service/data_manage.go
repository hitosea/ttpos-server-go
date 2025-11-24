package service

import (
	"time"
	shop_req "ttpos-server-go/app/dto/req"
	setting_resp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
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
	dbm *database.DBManager
}

// IDataManageSrv数据管理服务接口
type IDataManageSrv interface {
	GetDataManage(ctx context.Context) (*setting_resp.GetDataManageResp, error) // 获取数据管理信息
	SetDataManage(ctx context.Context, req shop_req.SetDataManageReq) error     // 设置数据管理
}

// NewDataManageSrvImpl 创建数据管理服务
func NewDataManageSrv(dbm *database.DBManager) IDataManageSrv {
	return NewDataManageSrvImpl(dbm)
}

// NewDataManageSrvImpl 创建数据管理服务实现
func NewDataManageSrvImpl(dbm *database.DBManager) IDataManageSrv {
	return &DataManageSrv{
		dbm: dbm,
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
	company := ctx.GetCompany()
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
		commonRepo.WhereInDataManageSubQuery("sale_bill_uuid"),
	)
	// 总优惠折扣率 = 总优惠折扣 / 总销售额
	var discountRatio decimal.Decimal
	if statisticsData.TotalSaleAmount.Float64 > 0 {
		discountRatio = decimal.NewFromFloat(statisticsData.TotalDiscount.Float64).
			Div(decimal.NewFromFloat(statisticsData.TotalSaleAmount.Float64)).
			Mul(decimal.NewFromInt(100))
	}

	return &setting_resp.GetDataManageResp{
		IsEnableDataManage: company.IsOpenDataManage(),
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

	db := s.dbm.GetDB(ctx.GetDbId())
	err = db.Transaction(func(tx *gorm.DB) error {
		companyRepo := repository.NewCompanyRepo(tx)
		staffRepo := repository.NewStaffRepo(tx)
		dataManageRepo := repository.NewDataManageRepo(tx)

		// 更新公司数据管理状态
		err := companyRepo.UpdateCompany(ctx.GetCompanyUuid(), map[string]any{
			"is_enable_data_manage": utils.IfInt(req.IsEnableDataManage, 1, 0),
		})
		if err != nil {
			return err
		}

		// 更新员工数据管理权限
		err = staffRepo.Updates(map[string]any{
			"has_data_permission": 0,
		}, staffRepo.WhereHasDataPermission(1))
		if err != nil {
			return err
		}
		if len(req.StaffUuids) > 0 {
			err = staffRepo.Updates(map[string]any{
				"has_data_permission": 1,
			}, staffRepo.WhereUuids(req.StaffUuids))
			if err != nil {
				return err
			}
		}

		// 删除数据管理数据
		err = dataManageRepo.Delete(
			dataManageRepo.WhereByType(model.DataManageTypeOrder),
		)
		if err != nil {
			return err
		}
		if len(req.SaleBillUuids) > 0 {
			staff := ctx.GetStaff()
			dataManages := []*model.DataManage{}
			for _, saleBillUuid := range req.SaleBillUuids {
				uuid, err := utils.GetID()
				if err != nil {
					return err
				}
				dataManages = append(dataManages, &model.DataManage{
					BaseModel: model.BaseModel{
						Uuid:       uuid,
						CreateTime: time.Now().Unix(),
						UpdateTime: time.Now().Unix(),
					},
					Type:      model.DataManageTypeOrder,
					DataUuid:  saleBillUuid,
					StaffUuid: staff.Uuid,
				})
			}
			err = dataManageRepo.Creates(dataManages)
			if err != nil {
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
