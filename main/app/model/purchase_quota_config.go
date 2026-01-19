package model

// PurchaseQuotaConfig 品牌采购限购配置主表
//
// 表名: ttpos_purchase_quota_config
//
// 说明: 存储品牌采购的物品级限购配置，支持按物品编码+单位编码维度设置限购数量
//
// 核心字段:
//   - material_code + unit_code: 限购的物品编码和单位编码组合
//   - quota_limit: 限购数量（按 period_type 周期）
//   - apply_to_all_shops: 是否应用到全部门店（1=是 0=否）
//   - period_type: 周期类型（0=按天 1=月度）
//
// 门店范围控制:
//   - apply_to_all_shops=1: 应用到全部门店，ttpos_purchase_quota_config_shop 表无记录
//   - apply_to_all_shops=0: 应用到指定门店，ttpos_purchase_quota_config_shop 表存储关联关系
type PurchaseQuotaConfig struct {
	Id              uint64  `gorm:"column:id;primaryKey" json:"id"`
	Uuid            uint64  `gorm:"column:uuid;uniqueIndex" json:"uuid"`
	MaterialCode    string  `gorm:"column:material_code;type:varchar(100);not null" json:"material_code"`  // 物品编码
	UnitCode        string  `gorm:"column:unit_code;type:varchar(50);not null" json:"unit_code"`           // 限购单位编码
	QuotaLimit      float64 `gorm:"column:quota_limit;type:decimal(10,2);default:0.00" json:"quota_limit"` // 限购数量
	ApplyToAllShops uint8   `gorm:"column:apply_to_all_shops;default:1" json:"apply_to_all_shops"`         // 1=全部门店 0=指定门店
	PeriodType      uint8   `gorm:"column:period_type;default:0" json:"period_type"`                       // 0=按天 1=月度
	StrictMode      uint8   `gorm:"column:strict_mode;default:1" json:"strict_mode"`                       // 1=严格拒绝
	ConfigSource    uint8   `gorm:"column:config_source;default:1" json:"config_source"`                   // 1=门店 2=总部
	Status          uint8   `gorm:"column:status;default:1" json:"status"`                                 // 1=启用 0=禁用
	CreateTime      int64   `gorm:"column:create_time;default:0" json:"create_time"`
	UpdateTime      int64   `gorm:"column:update_time;default:0" json:"update_time"`
	DeleteTime      int64   `gorm:"column:delete_time;index;default:0" json:"delete_time"`

	// 关联关系（不会序列化到 JSON）
	Shops []PurchaseQuotaConfigShop `gorm:"foreignKey:ConfigUuid;references:Uuid" json:"-"`
}

func (*PurchaseQuotaConfig) TableName() string {
	return "ttpos_purchase_quota_config"
}

// PurchaseQuotaConfigShop 品牌采购限购配置门店关联表
//
// 表名: ttpos_purchase_quota_config_shop
//
// 说明: 存储限购配置与门店的多对多关联关系
//
// 使用场景:
//   - 当 ttpos_purchase_quota_config.apply_to_all_shops=0 时，此表存储该配置应用的门店列表
//   - 当 ttpos_purchase_quota_config.apply_to_all_shops=1 时，此表无记录（表示全部门店）
//
// 数据一致性:
//   - 删除主表配置时，需同步软删除此表的关联记录（通过事务保证）
type PurchaseQuotaConfigShop struct {
	Id          uint64 `gorm:"column:id;primaryKey" json:"id"`
	ConfigUuid  uint64 `gorm:"column:config_uuid;not null" json:"config_uuid"`   // 关联 ttpos_purchase_quota_config.uuid
	CompanyUuid uint64 `gorm:"column:company_uuid;not null" json:"company_uuid"` // 公司UUID（门店UUID）
	CreateTime  int64  `gorm:"column:create_time;default:0" json:"create_time"`
	DeleteTime  int64  `gorm:"column:delete_time;index;default:0" json:"delete_time"`
}

func (*PurchaseQuotaConfigShop) TableName() string {
	return "ttpos_purchase_quota_config_shop"
}

// GetShopUuidList 获取配置关联的门店UUID列表
func (p *PurchaseQuotaConfig) GetShopUuidList() []uint64 {
	uuids := make([]uint64, 0, len(p.Shops))
	for _, shop := range p.Shops {
		if shop.DeleteTime == 0 {
			uuids = append(uuids, shop.CompanyUuid)
		}
	}
	return uuids
}

// AppliesTo 检查配置是否应用到指定门店
func (p *PurchaseQuotaConfig) AppliesTo(shopUuid uint64) bool {
	// 应用到全部店铺
	if p.ApplyToAllShops == 1 {
		return true
	}

	// 检查是否在关联门店列表中
	for _, shop := range p.Shops {
		if shop.CompanyUuid == shopUuid && shop.DeleteTime == 0 {
			return true
		}
	}
	return false
}
