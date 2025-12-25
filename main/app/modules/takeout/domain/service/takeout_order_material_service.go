package service

import (
	"ttpos-server-go/app/errors"
	takeoutModel "ttpos-server-go/app/modules/takeout/domain/model"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// ITakeoutOrderMaterialSrv 外卖订单原料服务接口
type ITakeoutOrderMaterialSrv interface {
	// SaveOrderMaterials 保存外卖订单原料到 ttpos_takeout_order_material 表
	SaveOrderMaterials(ctx context.Context, takeoutOrderMaterials []*takeoutModel.TakeoutOrderMaterial) error
}

// takeoutOrderMaterialSrv 外卖订单原料服务实现
type takeoutOrderMaterialSrv struct {
	dbm *database.DBManager
}

// NewTakeoutOrderMaterialSrv 创建外卖订单原料服务
func NewTakeoutOrderMaterialSrv(dbm *database.DBManager) ITakeoutOrderMaterialSrv {
	return &takeoutOrderMaterialSrv{
		dbm: dbm,
	}
}

// SaveOrderMaterials 保存外卖订单原料到 ttpos_takeout_order_material 表
func (s *takeoutOrderMaterialSrv) SaveOrderMaterials(
	ctx context.Context,
	takeoutOrderMaterials []*takeoutModel.TakeoutOrderMaterial,
) error {
	db := ctx.GetDB()

	if len(takeoutOrderMaterials) == 0 {
		return nil
	}

	// 批量插入原料记录
	takeoutOrderMaterialRepo := persistence.NewTakeoutOrderMaterialRepo(db)
	if err := takeoutOrderMaterialRepo.BatchCreate(takeoutOrderMaterials); err != nil {
		return errors.WithMessage(err, "批量插入外卖订单原料记录失败")
	}

	logger.Logger.Info("保存外卖订单原料成功",
		zap.Int("materialCount", len(takeoutOrderMaterials)))

	return nil
}
