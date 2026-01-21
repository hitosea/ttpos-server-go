package controller

import (
	"fmt"
	"sync"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/adapter"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SaleBillQueryFunc 单个 SaleBill 查询函数类型
type SaleBillQueryFunc func(db *gorm.DB, uuid uint64) (*model.SaleBill, error)

// SaleBillBatchQueryFunc 批量 SaleBill 查询函数类型
type SaleBillBatchQueryFunc func(db *gorm.DB, uuids []uint64) ([]*model.SaleBill, error)

// SaleBillController SaleBill 专用控制器
// 通过嵌入 CacheObjectController 继承其所有方法，并扩展 SaleBill 特有的方法
type SaleBillController struct {
	*CacheObjectController[*model.SaleBill]
}

// saleBillControllerInstance SaleBill 控制器单例实例
var (
	saleBillControllerInstance *SaleBillController
	saleBillControllerOnce     sync.Once
)

// InitSaleBillController 初始化 SaleBill 对象的缓存控制器（单例模式）
// 参数：
//   - underlyingCache: 底层缓存实例
//   - ttl: 缓存 TTL（默认 10 分钟）
//   - queryFunc: 单个 SaleBill 查询函数
//   - batchQueryFunc: 批量 SaleBill 查询函数
func InitSaleBillController(
	underlyingCache cache.Cache,
	ttl time.Duration,
	queryFunc SaleBillQueryFunc,
	batchQueryFunc SaleBillBatchQueryFunc,
) {
	// 构建缓存层选项,禁用 L1 本地缓存
	cacheOpts := []func(*adapter.CacheLayerOption){adapter.WithEnableL1Cache(false)}
	saleBillControllerOnce.Do(func() {
		baseController := InitCacheObjectController[*model.SaleBill](
			underlyingCache,
			ttl,
			func(db *gorm.DB, uuid uint64) (*model.SaleBill, error) {
				return queryFunc(db, uuid)
			},
			func(db *gorm.DB, uuids []uint64) ([]*model.SaleBill, error) {
				return batchQueryFunc(db, uuids)
			},
			func(obj *model.SaleBill) uint64 {
				return obj.Uuid
			},
			persistence.ObjectTypeSaleBill,
			cacheOpts...,
		)
		saleBillControllerInstance = &SaleBillController{
			CacheObjectController: baseController,
		}
	})
}

// GetSaleBillController 获取 SaleBill 对象的缓存控制器单例实例
func GetSaleBillController() *SaleBillController {
	return saleBillControllerInstance
}

// SaveToDb 保存 SaleBill 到数据库
// 参数：
//   - ctx: 上下文
//   - saleBill: 要保存的 SaleBill 对象
//
// 返回：
//   - error: 错误信息
func (c *SaleBillController) SaveToDb(ctx context.Context, saleBill *model.SaleBill) error {
	if saleBill == nil {
		return fmt.Errorf("SaleBill 对象不能为 nil")
	}

	db := ctx.GetDB()

	// 同步更新缓存
	if err := c.Update(ctx, db, []uint64{saleBill.Uuid}, WithUpdateValue(map[uint64]interface{}{
		saleBill.Uuid: saleBill,
	})); err != nil {
		return fmt.Errorf("更新缓存失败: %w", err)
	}

	// 异步保存到数据库
	utils.Go(func() {
		if err := db.Transaction(func(tx *gorm.DB) error {
			// 创建SaleBill到数据库
			if saleBillId, err := createSaleBill(tx, *saleBill); err != nil {
				logger.Logger.Error("创建 SaleBill 失败", zap.Error(err))
				return err
			} else {
				saleBill.ID = saleBillId
			}
			// 创建SaleBillSetting到数据库
			if saleBillSettingId, err := createSaleBillSetting(tx, *saleBill.SaleBillSetting); err != nil {
				logger.Logger.Error("创建 SaleBillSetting 失败", zap.Error(err))
				return err
			} else {
				saleBill.SaleBillSetting.ID = saleBillSettingId
			}
			// 创建SaleOrders到数据库
			for _, saleOrder := range saleBill.SaleOrders {
				if saleOrderId, err := createSaleOrder(tx, *saleOrder); err != nil {
					logger.Logger.Error("创建 SaleOrder 失败", zap.Error(err))
					return err
				} else {
					saleOrder.ID = saleOrderId
				}
			}
			return nil
		}); err != nil {
			logger.Logger.Error("保存 SaleBill 到数据库失败", zap.Error(err))
		}
		// 更新缓存
		if err := c.Update(ctx, db, []uint64{saleBill.Uuid}, WithUpdateValue(map[uint64]interface{}{
			saleBill.Uuid: saleBill,
		})); err != nil {
			logger.Logger.Error("更新缓存失败", zap.Error(err))
		}
	})
	return nil
}

// createSaleBill 创建销售单（避免循环导入，在 controller 包内实现）
func createSaleBill(db *gorm.DB, saleBill model.SaleBill) (uint, error) {
	saleBill.SetNil()
	if err := db.Model(&model.SaleBill{}).Create(&saleBill).Error; err != nil {
		return 0, fmt.Errorf("CreateSaleBill: %w", err)
	}
	return saleBill.ID, nil
}

// createSaleBillSetting 创建销售账单设置（避免循环导入，在 controller 包内实现）
func createSaleBillSetting(db *gorm.DB, setting model.SaleBillSetting) (uint, error) {
	if err := db.Model(&model.SaleBillSetting{}).Create(&setting).Error; err != nil {
		return 0, fmt.Errorf("CreateSaleBillSetting: %w", err)
	}
	return setting.ID, nil
}

// createSaleOrder 创建销售订单（避免循环导入，在 controller 包内实现）
func createSaleOrder(db *gorm.DB, saleOrder model.SaleOrder) (uint, error) {
	saleOrder.SetNil()
	if err := db.Model(&model.SaleOrder{}).Create(&saleOrder).Error; err != nil {
		return 0, fmt.Errorf("CreateSaleOrder: %w", err)
	}
	return saleOrder.ID, nil
}
