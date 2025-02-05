package model

// CompanySetting 公司设置表
type CompanySetting struct {
	ID                uint    `gorm:"column:id;type:int(11);primary_key;AUTO_INCREMENT;comment:主键id" json:"id"`
	ParentId          int     `gorm:"column:parent_id;type:int(11);default:0;comment:上级商家id" json:"parent_id"`
	Name              string  `gorm:"column:name;type:varchar(150);comment:商家姓名" json:"name"`
	RealName          string  `gorm:"column:real_name;type:varchar(50);comment:真实姓名" json:"real_name"`
	LinkName          string  `gorm:"column:link_name;type:varchar(50);comment:联系人" json:"link_name"`
	LinkPhone         string  `gorm:"column:link_phone;type:varchar(25);comment:联系电话" json:"link_phone"`
	Logo              string  `gorm:"column:logo;type:text;comment:logo" json:"logo"`
	Level             int     `gorm:"column:level;type:int(11);default:1;comment:商家等级: 1开始" json:"level"`
	SaleStock         int     `gorm:"column:sale_stock;type:int(11);default:0;comment:进销存: 0不开启, 1开启" json:"sale_stock"`
	Reserve           int     `gorm:"column:reserve;type:int(11);default:0;comment:预订: 0不开启, 1开启" json:"reserve"`
	IsOpenMember      int     `gorm:"column:is_open_member;type:int(11);default:0;comment:是否开启会员: 0不开启, 1开启" json:"is_open_member"`
	IsOpenTablet      int     `gorm:"column:is_open_tablet;type:int(11);default:0;comment:是否开启平板: 0不开启, 1开启" json:"is_open_tablet"`
	IsOpenScan        int     `gorm:"column:is_open_scan;type:int(11);default:0;comment:是否开启扫码H5: 0不开启, 1开启" json:"is_open_scan"`
	IsOpenAssistant   int     `gorm:"column:is_open_assistant;type:int(11);default:0;comment:是否开启点餐助手: 0不开启, 1开启" json:"is_open_assistant"`
	IsOpenKitchenKds  int     `gorm:"column:is_open_kitchen_kds;type:int(11);default:0;comment:是否开启后厨KDS: 0不开启, 1开启" json:"is_open_kitchen_kds"`
	IsOpenBuffet      int     `gorm:"column:is_open_buffet;type:int(11);default:0;comment:是否开启自助餐: 0不开启, 1开启" json:"is_open_buffet"`
	IsAcceptScanOrder int     `gorm:"column:is_accept_scan_order;type:int(11);default:0;comment:是否开启扫码点餐接单 0不开启, 1开启" json:"is_accept_scan_order"`
	IsOpenLocalPrint  int     `gorm:"column:is_open_local_print;type:int(11);default:1;comment:是否开启本地打印服务 0不开启, 1开启" json:"is_open_local_print"`
	CashLimit         int     `gorm:"column:cash_limit;type:int(11);default:0;comment:收银机上限" json:"cash_limit"`
	KitchenLimit      int     `gorm:"column:kitchen_limit;type:int(11);default:0;comment:厨显上限" json:"kitchen_limit"`
	TabletLimit       int     `gorm:"column:tablet_limit;type:int(11);default:0;comment:平板上限" json:"tablet_limit"`
	AssistantLimit    int     `gorm:"column:assistant_limit;type:int(11);default:0;comment:点餐助手上限" json:"assistant_limit"`
	TableLimit        int     `gorm:"column:table_limit;type:int(11);default:0;comment:桌台上限" json:"table_limit"`
	PrinterLimit      int     `gorm:"column:printer_limit;type:int(11);default:0;comment:打印机上限" json:"printer_limit"`
	Timezone          string  `gorm:"column:timezone;type:varchar(50);default:Asia/Shanghai;comment:时区" json:"timezone"`
	Languages         string  `gorm:"column:languages;type:longtext;comment:支持语言" json:"languages"`
	Address           string  `gorm:"column:address;type:text;comment:联系地址" json:"address"`
	DeployMode        int     `gorm:"column:deploy_mode;type:tinyint(4);default:0;comment:部署方式 0局域网部署, 1云部署;NOT NULL" json:"deploy_mode"`
	MacAddr           string  `gorm:"column:mac_addr;type:varchar(100);comment:mac地址" json:"mac_addr"`
	SerialNumber      string  `gorm:"column:serial_number;type:varchar(100);comment:服务序列号" json:"serial_number"`
	ChainNumber       string  `gorm:"column:chain_number;type:varchar(100);comment:连锁编号" json:"chain_number"`
	BusinessId        uint    `gorm:"column:business_id;type:int(11);default:0;comment:营业执照;NOT NULL" json:"business_id"`
	Description       string  `gorm:"column:description;type:text;comment:商家介绍" json:"description"`
	TotalMoney        float64 `gorm:"column:total_money;type:decimal(12,2);default:0.00;comment:总货款" json:"total_money"`
	Money             float64 `gorm:"column:money;type:decimal(12,2);default:0.00;comment:当前可提现金额" json:"money"`
	FreezeMoney       float64 `gorm:"column:freeze_money;type:decimal(12,2);default:0.00;comment:已冻结金额" json:"freeze_money"`
	CashMoney         float64 `gorm:"column:cash_money;type:decimal(12,2);default:0.00;comment:累积提现佣金" json:"cash_money"`
	DepositMoney      float64 `gorm:"column:deposit_money;type:decimal(12,2);default:0.00;comment:保证金" json:"deposit_money"`
	UserId            int     `gorm:"column:user_id;type:int(11);default:0;comment:会员id;NOT NULL" json:"user_id"`
	FavCount          int     `gorm:"column:fav_count;type:int(11);default:0;comment:关注人数" json:"fav_count"`
	Status            int     `gorm:"column:status;type:tinyint(3);default:0;comment:店铺状态0营业中1停止营业;NOT NULL" json:"status"`
	StoreType         int     `gorm:"column:store_type;type:tinyint(3);default:10;comment:店铺类型10加盟20自营;NOT NULL" json:"store_type"`
	TotalGift         int     `gorm:"column:total_gift;type:int(11);default:0;comment:收到的礼物币总数" json:"total_gift"`
	IsRecycle         uint    `gorm:"column:is_recycle;type:tinyint(3);default:1;comment:是否禁用0否1是;NOT NULL" json:"is_recycle"`
	IsMain            int     `gorm:"column:is_main;type:tinyint(3);default:0;comment:是否总店，0否1是" json:"is_main"`
	ProvinceId        uint    `gorm:"column:province_id;type:int(11);default:0;comment:所在省份id;NOT NULL" json:"province_id"`
	CityId            uint    `gorm:"column:city_id;type:int(11);default:0;comment:所在城市id;NOT NULL" json:"city_id"`
	RegionId          uint    `gorm:"column:region_id;type:int(11);default:0;comment:所在辖区id;NOT NULL" json:"region_id"`
	Longitude         string  `gorm:"column:longitude;type:varchar(50);comment:门店坐标经度;NOT NULL" json:"longitude"`
	Latitude          string  `gorm:"column:latitude;type:varchar(50);comment:门店坐标纬度;NOT NULL" json:"latitude"`
	ShippingFee       float64 `gorm:"column:shipping_fee;type:decimal(12,2);default:0.00;comment:配送费;NOT NULL" json:"shipping_fee"`
	BagType           int     `gorm:"column:bag_type;type:tinyint(1);default:0;comment:包装费类型0按商品收费1按单收费;NOT NULL" json:"bag_type"`
	BagPrice          float64 `gorm:"column:bag_price;type:decimal(12,2);default:0.00;comment:包装费;NOT NULL" json:"bag_price"`
	StoreBagType      int     `gorm:"column:store_bag_type;type:tinyint(1);default:0;comment:店内包装费类型0按商品收费1按单收费;NOT NULL" json:"store_bag_type"`
	StoreBagPrice     float64 `gorm:"column:store_bag_price;type:decimal(12,2);default:0.00;comment:店内包装费;NOT NULL" json:"store_bag_price"`
	DeliveryTime      string  `gorm:"column:delivery_time;type:varchar(100);comment:外卖营业时间;NOT NULL" json:"delivery_time"`
	PickTime          string  `gorm:"column:pick_time;type:varchar(100);comment:自提营业时间;NOT NULL" json:"pick_time"`
	StoreTime         string  `gorm:"column:store_time;type:varchar(100);comment:店内营业时间;NOT NULL" json:"store_time"`
	DeliveryDistance  float64 `gorm:"column:delivery_distance;type:float(10,2);default:0.00;comment:配送范围km;NOT NULL" json:"delivery_distance"`
	DeliverySet       string  `gorm:"column:delivery_set;type:varchar(150);comment:外卖配送方式;NOT NULL" json:"delivery_set"`
	StoreSet          string  `gorm:"column:store_set;type:varchar(150);comment:店内用餐方式;NOT NULL" json:"store_set"`
	MinMoney          float64 `gorm:"column:min_money;type:decimal(12,2);default:0.00;comment:最低消费;NOT NULL" json:"min_money"`
	SettleType        int     `gorm:"column:settle_type;type:tinyint(3);default:10;comment:计算模式10先结账后用餐20先用餐后结账;NOT NULL" json:"settle_type"`
	ServiceType       int     `gorm:"column:service_type;type:tinyint(1);default:0;comment:服务费类型0按就餐人数1按桌台收费;NOT NULL" json:"service_type"`
	ServiceMoney      float64 `gorm:"column:service_money;type:decimal(12,2);default:0.00;comment:服务费;NOT NULL" json:"service_money"`
	AutoClose         int     `gorm:"column:auto_close;type:tinyint(1);default:1;comment:0定时清台1立即清台;NOT NULL" json:"auto_close"`
	CloseTime         int     `gorm:"column:close_time;type:int(10);default:0;comment:0分钟清台;NOT NULL" json:"close_time"`
	CategorySet       int     `gorm:"column:category_set;type:tinyint(1);default:10;comment:商品分类设置10同步主店20分店创建;NOT NULL" json:"category_set"`
	IsDelete          int     `gorm:"column:is_delete;type:tinyint(3);default:0;comment:是否删除0，否1是" json:"is_delete"`
	CompanyId         uint    `gorm:"column:company_id;type:int(11);default:0;comment:集团id;NOT NULL" json:"company_id"`
	CreateTime        uint    `gorm:"autoCreateTime;column:create_time;type:int(11);comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime        int     `gorm:"column:update_time;type:int(11);comment:更新时间;NOT NULL" json:"update_time"`
	DeleteTime        int     `gorm:"column:delete_time;type:int(11);comment:删除时间;NOT NULL" json:"delete_time"`
}
