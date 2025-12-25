package model

// TakeoutOrderCampaign 外卖订单活动表
// 存储外卖订单的营销活动信息，包括折扣、优惠券等
type TakeoutOrderCampaign struct {
	BaseModel
	TakeoutOrderUuid uint64 `gorm:"column:takeout_order_uuid;index:idx_takeout_order_uuid;not null;comment:外卖订单UUID" json:"takeout_order_uuid"`
	Platform         string `gorm:"column:platform;size:50;not null;comment:平台名称(grab/lineman/foodpanda等)" json:"platform"`

	// 活动基本信息
	CampaignID         string `gorm:"column:campaign_id;size:100;comment:活动ID" json:"campaign_id"`
	CampaignName       string `gorm:"column:campaign_name;size:255;comment:活动名称" json:"campaign_name"`
	CampaignNameForMex string `gorm:"column:campaign_name_for_mex;size:255;comment:商户提供的活动名称" json:"campaign_name_for_mex"`
	CampaignLevel      string `gorm:"column:campaign_level;size:50;comment:活动级别(order/item)" json:"campaign_level"`
	CampaignType       string `gorm:"column:campaign_type;size:50;comment:活动类型(discount/bundle/free_item)" json:"campaign_type"`

	// 活动使用信息
	UsageCount     int32  `gorm:"column:usage_count;default:0;comment:活动使用次数" json:"usage_count"`
	MexFundedRatio int32  `gorm:"column:mex_funded_ratio;default:0;comment:商户资金占比(%)" json:"mex_funded_ratio"`
	DeductedAmount int64  `gorm:"column:deducted_amount;default:0;comment:折扣金额(分)" json:"deducted_amount"`
	DeductedPart   string `gorm:"column:deducted_part;size:50;comment:折扣部分(subtotal/delivery_fee)" json:"deducted_part"`

	// 应用的商品
	AppliedItemIDs string `gorm:"column:applied_item_ids;type:text;comment:应用的商品ID列表(JSON数组)" json:"applied_item_ids"`

	// 赠品信息
	FreeItemID       string `gorm:"column:free_item_id;size:100;comment:赠品ID" json:"free_item_id"`
	FreeItemName     string `gorm:"column:free_item_name;size:255;comment:赠品名称" json:"free_item_name"`
	FreeItemQuantity int32  `gorm:"column:free_item_quantity;default:0;comment:赠品数量" json:"free_item_quantity"`
	FreeItemPrice    int64  `gorm:"column:free_item_price;default:0;comment:赠品价格(分)" json:"free_item_price"`
}

func (*TakeoutOrderCampaign) TableName() string {
	return "ttpos_takeout_order_campaign"
}

func (c *TakeoutOrderCampaign) GetCampaignName() string {
	if c.CampaignNameForMex != "" {
		return c.CampaignNameForMex
	}
	return c.CampaignName
}
