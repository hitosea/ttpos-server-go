package grab

// GrabMenu Grab 平台菜单结构（对应 1.json）
type GrabMenu struct {
	Currency     GrabCurrency      `json:"currency"`
	SellingTimes []GrabSellingTime `json:"sellingTimes"`
	Categories   []GrabCategory    `json:"categories"`
}

// GrabCurrency Grab 货币信息
type GrabCurrency struct {
	Code     string `json:"code"`
	Symbol   string `json:"symbol"`
	Exponent int    `json:"exponent"`
}

// GrabSellingTime Grab 售卖时段
type GrabSellingTime struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Sequence     int              `json:"sequence"`
	ServiceHours GrabServiceHours `json:"serviceHours"`
	StartTime    string           `json:"startTime"`
	EndTime      string           `json:"endTime"`
}

// GrabServiceHours 每周营业时间
type GrabServiceHours struct {
	Mon GrabDayServiceHours `json:"mon"`
	Tue GrabDayServiceHours `json:"tue"`
	Wed GrabDayServiceHours `json:"wed"`
	Thu GrabDayServiceHours `json:"thu"`
	Fri GrabDayServiceHours `json:"fri"`
	Sat GrabDayServiceHours `json:"sat"`
	Sun GrabDayServiceHours `json:"sun"`
}

// GrabDayServiceHours 单日营业时间
type GrabDayServiceHours struct {
	OpenPeriodType string              `json:"openPeriodType"`
	Periods        []GrabServicePeriod `json:"periods"`
}

// GrabServicePeriod 营业时段
type GrabServicePeriod struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

// GrabCategory Grab 分类
type GrabCategory struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Sequence        int               `json:"sequence"`
	AvailableStatus string            `json:"availableStatus"`
	Items           []GrabItem        `json:"items"`
	SellingTimeID   string            `json:"sellingTimeID,omitempty"`
	NameTranslation map[string]string `json:"nameTranslation,omitempty"`
}

// GrabItem Grab 商品
type GrabItem struct {
	ID                     string              `json:"id"`
	Name                   string              `json:"name"`
	Sequence               int                 `json:"sequence"`
	AvailableStatus        string              `json:"availableStatus"`
	Price                  int64               `json:"price"`
	CampaignInfo           *GrabCampaignInfo   `json:"campaignInfo"`
	Description            string              `json:"description,omitempty"`
	DescriptionTranslation map[string]string   `json:"descriptionTranslation,omitempty"`
	Photos                 []string            `json:"photos,omitempty"`
	ModifierGroups         []GrabModifierGroup `json:"modifierGroups,omitempty"`
	SellingTimeID          string              `json:"sellingTimeID,omitempty"`
	NameTranslation        map[string]string   `json:"nameTranslation,omitempty"`
}

// GrabCampaignInfo Grab 营销活动信息
type GrabCampaignInfo struct {
	OriginalPrice int64  `json:"originalPrice"`
	DiscountType  string `json:"discountType"`
	DiscountValue int64  `json:"discountValue"`
}

// GrabModifierGroup Grab 修饰符组
type GrabModifierGroup struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	NameTranslation   map[string]string `json:"nameTranslation,omitempty"`
	Sequence          int               `json:"sequence"`
	AvailableStatus   string            `json:"availableStatus"`
	SelectionRangeMin int               `json:"selectionRangeMin"`
	SelectionRangeMax int               `json:"selectionRangeMax"`
	Modifiers         []GrabModifier    `json:"modifiers"`
}

// GrabModifier Grab 修饰符
type GrabModifier struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Sequence        int               `json:"sequence"`
	AvailableStatus string            `json:"availableStatus"`
	Price           int64             `json:"price"`
	NameTranslation map[string]string `json:"nameTranslation,omitempty"`
}
