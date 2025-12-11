package inventory

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/inventory/domain/repository"
	appRepo "ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"

	"gorm.io/gorm"
)

// ProductPackageRepositoryImpl 商品包仓储实现
type ProductPackageRepositoryImpl struct {
	dbm *database.DBManager
}

// NewProductPackageRepository 创建商品包仓储
func NewProductPackageRepository(dbm *database.DBManager) repository.IProductPackageRepository {
	return &ProductPackageRepositoryImpl{
		dbm: dbm,
	}
}

// FindByUuid 根据UUID查找商品包
func (r *ProductPackageRepositoryImpl) FindByUuid(
	ctx context.Context,
	uuid uint64,
) (*model.ProductPackage, error) {
	db := r.getDB(ctx)
	repo := appRepo.NewProductPackageRepo(db)

	// 查询商品包，不预加载任何关联数据，只查询基本信息
	productPackage, err := repo.GetProductPackage(
		appRepo.CommonRepo.WhereByUuid(uuid),
	)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err, "查询商品包失败")
	}

	return productPackage, nil
}

// getDB 获取数据库连接
func (r *ProductPackageRepositoryImpl) getDB(ctx context.Context) *gorm.DB {
	return ctx.GetDB()
}
