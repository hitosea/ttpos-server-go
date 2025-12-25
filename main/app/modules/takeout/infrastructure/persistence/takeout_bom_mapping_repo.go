package persistence

import (
	"gorm.io/gorm"
)

// GroupItemBomMapping 套餐商品的 BOM 映射信息
type GroupItemBomMapping struct {
	ProductBomUuid uint64  // BOM UUID
	Num            float64 // 套餐配置的数量
}

// ITakeoutBomMappingRepo 外卖商品 BOM 映射仓储接口
type ITakeoutBomMappingRepo interface {
	// GetProductBomMapping 批量查询外卖商品包UUID到BOM UUID的映射
	GetProductBomMapping(productPackageUuids []uint64) (map[uint64]uint64, error)

	// GetGroupItemBomMapping 批量查询套餐商品项UUID到BOM映射和数量
	GetGroupItemBomMapping(groupItemUuids []uint64) (map[uint64]GroupItemBomMapping, error)

	// GetProductBomInfo 批量查询 BOM UUID 获取 BOM 信息（name字段）
	GetProductBomInfo(bomUuids []uint64) (map[uint64]string, error)
}

// takeoutBomMappingRepoImpl 外卖商品 BOM 映射仓储实现
type takeoutBomMappingRepoImpl struct {
	db *gorm.DB
}

// NewTakeoutBomMappingRepo 创建外卖商品 BOM 映射仓储
func NewTakeoutBomMappingRepo(db *gorm.DB) ITakeoutBomMappingRepo {
	return &takeoutBomMappingRepoImpl{db: db}
}

// GetProductBomMapping 批量查询外卖商品包UUID到BOM UUID的映射
// 查询 ttpos_product_bom_takeout 表
// 返回: map[ProductPackageTakeoutUuid]ProductBomUuid
func (r *takeoutBomMappingRepoImpl) GetProductBomMapping(productPackageUuids []uint64) (map[uint64]uint64, error) {
	if len(productPackageUuids) == 0 {
		return make(map[uint64]uint64), nil
	}

	var results []struct {
		ProductPackageTakeoutUuid uint64 `gorm:"column:product_package_takeout_uuid"`
		ProductBomUuid            uint64 `gorm:"column:product_bom_uuid"`
	}

	err := r.db.Table("ttpos_product_bom_takeout").
		Select("product_package_takeout_uuid, product_bom_uuid").
		Where("product_package_takeout_uuid IN ?", productPackageUuids).
		Where("delete_time = 0").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	// 转换为 map
	mapping := make(map[uint64]uint64, len(results))
	for _, result := range results {
		mapping[result.ProductPackageTakeoutUuid] = result.ProductBomUuid
	}

	return mapping, nil
}

// GetGroupItemBomMapping 批量查询套餐商品项UUID到BOM UUID的映射
// 查询 ttpos_product_package_group_item 表
// 返回: map[GroupItemUuid]ProductBomUuid
// GetGroupItemBomMapping 批量查询套餐商品的 BOM 映射和数量
// 返回: map[groupItemUuid]GroupItemBomMapping
func (r *takeoutBomMappingRepoImpl) GetGroupItemBomMapping(groupItemUuids []uint64) (map[uint64]GroupItemBomMapping, error) {
	if len(groupItemUuids) == 0 {
		return make(map[uint64]GroupItemBomMapping), nil
	}

	var results []struct {
		Uuid           uint64  `gorm:"column:uuid"`
		ProductBomUuid uint64  `gorm:"column:product_bom_uuid"`
		Num            float64 `gorm:"column:num"`
	}

	err := r.db.Table("ttpos_product_package_group_item").
		Select("uuid, product_bom_uuid, num").
		Where("uuid IN ?", groupItemUuids).
		Where("delete_time = 0").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	// 转换为 map
	mapping := make(map[uint64]GroupItemBomMapping, len(results))
	for _, result := range results {
		mapping[result.Uuid] = GroupItemBomMapping{
			ProductBomUuid: result.ProductBomUuid,
			Num:            result.Num,
		}
	}

	return mapping, nil
}

// GetProductBomInfo 批量查询 BOM UUID 获取 BOM 信息（包括 name）
// 查询 ttpos_product_bom 表
// 返回: map[BomUuid]Name
func (r *takeoutBomMappingRepoImpl) GetProductBomInfo(bomUuids []uint64) (map[uint64]string, error) {
	if len(bomUuids) == 0 {
		return make(map[uint64]string), nil
	}

	var results []struct {
		Uuid uint64 `gorm:"column:uuid"`
		Name string `gorm:"column:name"`
	}

	err := r.db.Table("ttpos_product_bom").
		Select("uuid, name").
		Where("uuid IN ?", bomUuids).
		Where("delete_time = 0").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	// 转换为 map
	mapping := make(map[uint64]string, len(results))
	for _, result := range results {
		mapping[result.Uuid] = result.Name
	}

	return mapping, nil
}
