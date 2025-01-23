package setting

type LanguageItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	I     string `json:"i"`
	Index string `json:"index"`
}

type BuffetOrderLimit struct {
	IsLimitTime string `json:"is_limit_time"`
	LimitTime   string `json:"limit_time"`
	IsLimitNum  string `json:"is_limit_num"`
	LimitNum    string `json:"limit_num"`
}

type OrderLimit struct {
	IsLimitTime string `json:"is_limit_time"`
	LimitTime   string `json:"limit_time"`
	IsLimitNum  string `json:"is_limit_num"`
	LimitNum    string `json:"limit_num"`
}

type CarouselItem struct {
	FilePath string `json:"file_path"`
	RealName string `json:"real_name"`
	Sort     string `json:"sort"`
	Type     string `json:"type"`
}

type Server struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}
