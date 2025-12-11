package repository

import (
	"context"
	"ttpos-server-go/app/model"
)

// IMenuDataRepository 菜单数据仓储接口（领域层）
// 用于从数据库加载外卖菜单相关的数据
type IMenuDataRepository interface {
	// GetTakeoutCategories 获取外卖分类列表
	// 过滤条件：is_display_in_takeout = 1 且 status = 1
	GetTakeoutCategories(ctx context.Context, companyUuid uint64, categoryIDs []uint64) ([]*model.ProductCategory, error)

	// GetTakeoutProducts 获取指定分类下的外卖商品
	// 过滤条件：status = 1（上架）且 category_uuid 匹配
	GetTakeoutProducts(ctx context.Context, companyUuid uint64, categoryUuid uint64) ([]*model.ProductPackageTakeout, error)
}
