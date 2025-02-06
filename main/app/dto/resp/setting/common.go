package setting

type BuffetOrderLimit OrderLimit

type OrderLimit struct {
	IsLimitTime string `json:"is_limit_time"` // 是否开启时间限制 0-关闭 1-开启
	LimitTime   string `json:"limit_time"`    // 时间限制（分钟）
	IsLimitNum  string `json:"is_limit_num"`  // 是否开启数量限制 0-关闭 1-开启
	LimitNum    string `json:"limit_num"`     // 数量限制
}

type CarouselItem struct {
	FilePath string `json:"file_path"`
	RealName string `json:"real_name"`
	Sort     string `json:"sort"`
	Type     string `json:"type"`
}

type Server struct {
	IP   string `json:"ip"`   // ip地址
	Port string `json:"port"` // 端口号
}
