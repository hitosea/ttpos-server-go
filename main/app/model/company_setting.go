package model

// CompanySetting 公司设置表
type CompanySetting struct {
	ID                uint    `gorm:"primary_key;AUTO_INCREMENT;comment:主键id"`
	ParentId          int     `gorm:"default:0;comment:上级商家id"`
	Name              string  `gorm:"default:'';comment:商家姓名"`
	RealName          string  `gorm:"default:'';comment:真实姓名"`
	LinkName          string  `gorm:"default:'';comment:联系人"`
	LinkPhone         string  `gorm:"default:'';comment:联系电话"`
	Logo              string  `gorm:"default:'';comment:logo"`
	Level             int     `gorm:"default:1;comment:商家等级: 1开始"`
	SaleStock         int     `gorm:"default:0;comment:进销存: 0不开启, 1开启"`
	Reserve           int     `gorm:"default:0;comment:预订: 0不开启, 1开启"`
	IsOpenMember      int     `gorm:"default:0;comment:是否开启会员: 0不开启, 1开启"`
	IsOpenTablet      int     `gorm:"default:0;comment:是否开启平板: 0不开启, 1开启"`
	IsOpenScan        int     `gorm:"default:0;comment:是否开启扫码H5: 0不开启, 1开启"`
	IsOpenAssistant   int     `gorm:"default:0;comment:是否开启点餐助手: 0不开启, 1开启"`
	IsOpenKitchenKds  int     `gorm:"default:0;comment:是否开启后厨KDS: 0不开启, 1开启"`
	IsOpenBuffet      int     `gorm:"default:0;comment:是否开启自助餐: 0不开启, 1开启"`
	IsAcceptScanOrder int     `gorm:"default:0;comment:是否开启扫码点餐接单 0不开启, 1开启"`
	IsOpenLocalPrint  int     `gorm:"default:1;comment:是否开启本地打印服务 0不开启, 1开启"`
	CashLimit         int     `gorm:"default:0;comment:收银机上限"`
	KitchenLimit      int     `gorm:"default:0;comment:厨显上限"`
	TabletLimit       int     `gorm:"default:0;comment:平板上限"`
	AssistantLimit    int     `gorm:"default:0;comment:点餐助手上限"`
	TableLimit        int     `gorm:"default:0;comment:桌台上限"`
	PrinterLimit      int     `gorm:"default:0;comment:打印机上限"`
	Timezone          string  `gorm:"default:Asia/Shanghai;comment:时区"`
	Languages         string  `gorm:"default:'';comment:支持语言"`
	Address           string  `gorm:"default:'';comment:联系地址"`
	DeployMode        int     `gorm:"default:0;comment:部署方式 0局域网部署, 1云部署"`
	MacAddr           string  `gorm:"default:'';comment:mac地址"`
	SerialNumber      string  `gorm:"default:'';comment:服务序列号"`
	ChainNumber       string  `gorm:"default:'';comment:连锁编号"`
	BusinessId        uint    `gorm:"default:0;comment:营业执照"`
	Description       string  `gorm:"default:'';comment:商家介绍"`
	TotalMoney        float64 `gorm:"default:0.00;comment:总货款"`
	Money             float64 `gorm:"default:0.00;comment:当前可提现金额"`
	FreezeMoney       float64 `gorm:"default:0.00;comment:已冻结金额"`
	CashMoney         float64 `gorm:"default:0.00;comment:累积提现佣金"`
	DepositMoney      float64 `gorm:"default:0.00;comment:保证金"`
	UserId            int     `gorm:"default:0;comment:会员id"`
	FavCount          int     `gorm:"default:0;comment:关注人数"`
	Status            int     `gorm:"default:0;comment:店铺状态0营业中1停止营业"`
	StoreType         int     `gorm:"default:10;comment:店铺类型10加盟20自营"`
	TotalGift         int     `gorm:"default:0;comment:收到的礼物币总数"`
	IsRecycle         uint    `gorm:"default:1;comment:是否禁用0否1是"`
	IsMain            int     `gorm:"default:0;comment:是否总店，0否1是"`
	ProvinceId        uint    `gorm:"default:0;comment:所在省份id"`
	CityId            uint    `gorm:"default:0;comment:所在城市id"`
	RegionId          uint    `gorm:"default:0;comment:所在辖区id"`
	Longitude         string  `gorm:"default:'';comment:门店坐标经度"`
	Latitude          string  `gorm:"default:'';comment:门店坐标纬度"`
	ShippingFee       float64 `gorm:"default:0.00;comment:配送费"`
	BagType           int     `gorm:"default:0;comment:包装费类型0按商品收费1按单收费"`
	BagPrice          float64 `gorm:"default:0.00;comment:包装费"`
	StoreBagType      int     `gorm:"default:0;comment:店内包装费类型0按商品收费1按单收费"`
	StoreBagPrice     float64 `gorm:"default:0.00;comment:店内包装费"`
	DeliveryTime      string  `gorm:"default:'';comment:外卖营业时间"`
	PickTime          string  `gorm:"default:'';comment:自提营业时间"`
	StoreTime         string  `gorm:"default:'';comment:店内营业时间"`
	DeliveryDistance  float64 `gorm:"default:0.00;comment:配送范围km"`
	DeliverySet       string  `gorm:"default:'';comment:外卖配送方式"`
	StoreSet          string  `gorm:"default:'';comment:店内用餐方式"`
	MinMoney          float64 `gorm:"default:0.00;comment:最低消费"`
	SettleType        int     `gorm:"default:10;comment:计算模式10先结账后用餐20先用餐后结账"`
	ServiceType       int     `gorm:"default:0;comment:服务费类型0按就餐人数1按桌台收费"`
	ServiceMoney      float64 `gorm:"default:0.00;comment:服务费"`
	AutoClose         int     `gorm:"default:1;comment:0定时清台1立即清台"`
	CloseTime         int     `gorm:"default:0;comment:0分钟清台"`
	CategorySet       int     `gorm:"default:10;comment:商品分类设置10同步主店20分店创建"`
	IsDelete          int     `gorm:"default:0;comment:是否删除0，否1是"`
	CompanyId         uint    `gorm:"default:0;comment:集团id"`
	CreateTime        int64   `gorm:"autoCreateTime;comment:创建时间"`
	UpdateTime        int64   `gorm:"autoUpdateTime;comment:更新时间"`
	DeleteTime        int64   `gorm:"default:0;comment:删除时间"`
}
