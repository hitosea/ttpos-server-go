package setting

type Store struct {
	Name          string         `json:"name"`
	AvatarURL     string         `json:"avatarUrl"`
	LogoURL       string         `json:"logoUrl"`
	ZeroingMethod string         `json:"zeroing_method"`
	IPWhiteList   string         `json:"ip_white_list"`
	TimeZone      string         `json:"time_zone"`
	NoClearTable  string         `json:"no_clear_table"`
	TimeZoneList  []TimeZoneItem `json:"time_zone_list"`
	Company       string         `json:"company"`
	Address       string         `json:"address"`
	Phone         string         `json:"phone"`
	TaxNumber     string         `json:"tax_number"`
	ChainNumber   string         `json:"chain_number"`
	Language      []Language     `json:"language"`
	AuthLanguage  string         `json:"auth_language"`
}

type TimeZoneItem struct {
	Name  string `json:"name"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Language struct {
	Key   string `json:"key"`
	Name  string `json:"name"`
	Value string `json:"value"`
}
