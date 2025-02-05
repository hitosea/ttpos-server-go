package setting

type Buffet struct {
	IsOpen                   string         `json:"is_open"`
	TabletEndTime            string         `json:"tablet_end_time"`
	IsRemainContinue         string         `json:"is_remain_continue"`
	RemainContinueTime       string         `json:"remain_continue_time"`
	RemainContinueNoticeTime string         `json:"remain_continue_notice_time"`
	IsBuyContinue            string         `json:"is_buy_continue"`
	IsAddClock               string         `json:"is_add_clock"`
	IsBuffetDiscount         string         `json:"is_buffet_discount"`
	AddClock                 []AddClockItem `json:"add_clock"`
}

type AddClockItem struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	DelayTime string `json:"delay_time"`
	Price     string `json:"price"`
	Action    string `json:"action"`
}
