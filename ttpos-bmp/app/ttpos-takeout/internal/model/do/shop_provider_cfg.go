// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ShopProviderCfg is the golang structure of table takeout_shop_provider_cfg for DAO operations like Where/Data.
type ShopProviderCfg struct {
	g.Meta             `orm:"table:takeout_shop_provider_cfg, do:true"`
	Id                 any // 主键ID
	Uuid               any // 唯一标识
	ShopUuid           any // 门店UUID
	ProviderName       any // 第三方名称，如 grab
	ProviderMerchantId any // 第三方商户ID
	ProviderShopStatus any // 门店集成状态
	CreatedAt          any // 创建时间
	UpdatedAt          any // 更新时间
	DeletedAt          any // 删除时间
}
