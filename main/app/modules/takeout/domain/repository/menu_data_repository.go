package repository

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"
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

	// GetProductNameByUuid 根据商品UUID获取多语言名称
	// 优先返回中文名称，如果没有则返回英文名称
	GetProductNameByUuid(ctx context.Context, productUuid uint64, productType int) (string, error)

	// GetProductNamesByUuids 批量根据商品UUID获取多语言名称
	// 返回 map[productUuid]name
	GetProductNamesByUuids(ctx context.Context, productUuids []uint64, productTypes map[uint64]int) map[uint64]string

	// GetModifierNamesByUuids 批量根据修饰符UUID和类型获取多语言名称
	// 返回 map[modifierUuid]name
	GetModifierNamesByUuids(ctx context.Context, modifierUuids []uint64, modifierTypes map[uint64]string) map[uint64]string

	// GetMenuNamesByPlatformItemIds 批量根据平台商品ID获取菜单名称
	// 从 ttpos_takeout 表的 menu JSON 字段获取
	// 返回 map[platformItemId]menuName
	GetMenuNamesByPlatformItemIds(ctx context.Context, platform string, platformItemIds []string) map[string]string

	// GetModifierNamesByPlatformIds 批量根据平台修饰符ID获取修饰符名称
	// 从 ttpos_takeout 表的 menu JSON 字段获取
	// 返回 map[platformModifierId]modifierName
	GetModifierNamesByPlatformIds(ctx context.Context, platform string, platformModifierIds []string) map[string]string
}
