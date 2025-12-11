package repository

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"
)

// IProductPackageRepository 商品包仓储接口
type IProductPackageRepository interface {
	// FindByUuid 根据UUID查找商品包
	FindByUuid(ctx context.Context, uuid uint64) (*model.ProductPackage, error)
}

