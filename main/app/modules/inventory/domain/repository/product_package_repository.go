package repository

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"
)

// IProductPackageRepository 商品包仓储接口
type IProductPackageRepository interface {
	// FindByUuid 根据UUID查找商品包
	FindByUuid(ctx context.Context, uuid uint64) (*model.ProductPackage, error)

	// FindByUuids 根据UUID列表批量查找商品包列表
	FindByUuids(ctx context.Context, uuids []uint64) ([]*model.ProductPackage, error)
}
