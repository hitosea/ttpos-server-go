package setting

// Buffet 自助餐设置
type Buffet struct {
	IsOpen                   string         `json:"is_open"`                     // 是否开启自助餐 0-关闭 1-开启
	TabletEndTime            string         `json:"tablet_end_time"`             // 平板结束时间提醒（分）
	IsRemainContinue         string         `json:"is_remain_continue"`          // 剩余xx分不可继续点餐开关 0-关闭 1-开启
	RemainContinueTime       string         `json:"remain_continue_time"`        // 剩余xx分不可继续点餐
	RemainContinueNoticeTime string         `json:"remain_continue_notice_time"` // 剩余xx分提醒不可继续点餐
	IsBuyContinue            string         `json:"is_buy_continue"`             // 非自助餐商品到时是否能继续选购 0-关闭 1-开启
	IsAddClock               string         `json:"is_add_clock"`                // 是否开启加钟 0-关闭 1-开启
	IsBuffetDiscount         string         `json:"is_buffet_discount"`          // 是否开启自助餐优惠折扣 0-关闭 1-开启
	IsShowNonBuffetProduct   string         `json:"is_show_non_buffet_product"`  // 是否显示非自助餐商品 0-关闭 1-开启
	AddClock                 []AddClockItem `json:"add_clock"`                   // 名称 - 加钟时间（分）- 价格
}

// BuffetResp 自助餐设置
type BuffetResp struct {
	IsOpen                   string         `json:"is_open"`                     // 是否开启自助餐 0-关闭 1-开启
	TabletEndTime            int            `json:"tablet_end_time"`             // 平板结束时间提醒（秒）
	IsRemainContinue         string         `json:"is_remain_continue"`          // 剩余xx分不可继续点餐开关 0-关闭 1-开启
	RemainContinueTime       string         `json:"remain_continue_time"`        // 剩余xx秒不可继续点餐
	RemainContinueNoticeTime string         `json:"remain_continue_notice_time"` // 剩余xx秒提醒不可继续点餐
	IsBuyContinue            string         `json:"is_buy_continue"`             // 非自助餐商品到时是否能继续选购 0-关闭 1-开启 (用餐时间到后可继续选购非自助餐商品)
	IsAddClock               string         `json:"is_add_clock"`                // 是否开启加钟 0-关闭 1-开启
	IsBuffetDiscount         string         `json:"is_buffet_discount"`          // 是否开启自助餐优惠折扣 0-关闭 1-开启
	IsShowNonBuffetProduct   string         `json:"is_show_non_buffet_product"`  // 是否显示非自助餐商品 0-关闭 1-开启
	AddClock                 []AddClockItem `json:"add_clock"`                   // 名称 - 加钟时间（分）- 价格
}

type AddClockItem struct {
	ID        int    `json:"id"`         // 加钟id
	Name      string `json:"name"`       // 名称
	DelayTime string `json:"delay_time"` // 时间（分）
	Price     string `json:"price"`      // 价格
	Action    string `json:"action"`     // 操作结果 delete-删除 edit-编辑 add-新增
}
