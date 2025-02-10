package model

// CompanySetting 公司设置表
type CompanySetting struct {
	ID                int    `gorm:"column:id;type:int(11);primary_key;AUTO_INCREMENT;comment:自增ID" json:"id"`
	CompanyUuid       int64  `gorm:"column:company_uuid;type:bigint(20);default:0;comment:集团ID;NOT NULL" json:"company_uuid"`
	RealName          string `gorm:"column:real_name;type:varchar(50);comment:真实姓名;NOT NULL" json:"real_name"`
	LinkName          string `gorm:"column:link_name;type:varchar(50);comment:联系人;NOT NULL" json:"link_name"`
	LinkPhone         string `gorm:"column:link_phone;type:varchar(25);comment:联系电话;NOT NULL" json:"link_phone"`
	SaleStock         int    `gorm:"column:sale_stock;type:int(11);default:0;comment:进销存: 0不开启, 1开启;NOT NULL" json:"sale_stock"`
	IsOpenMember      int    `gorm:"column:is_open_member;type:int(11);default:0;comment:是否开启会员: 0不开启, 1开启;NOT NULL" json:"is_open_member"`
	IsOpenTablet      int    `gorm:"column:is_open_tablet;type:int(11);default:0;comment:是否开启平板: 0不开启, 1开启;NOT NULL" json:"is_open_tablet"`
	IsOpenScan        int    `gorm:"column:is_open_scan;type:int(11);default:0;comment:是否开启扫码H5: 0不开启, 1开启;NOT NULL" json:"is_open_scan"`
	IsOpenAssistant   int    `gorm:"column:is_open_assistant;type:int(11);default:0;comment:是否开启点餐助手: 0不开启, 1开启;NOT NULL" json:"is_open_assistant"`
	IsOpenKitchenKds  int    `gorm:"column:is_open_kitchen_kds;type:int(11);default:0;comment:是否开启后厨KDS: 0不开启, 1开启;NOT NULL" json:"is_open_kitchen_kds"`
	IsOpenBuffet      int    `gorm:"column:is_open_buffet;type:int(11);default:0;comment:是否开启自助餐: 0不开启, 1开启;NOT NULL" json:"is_open_buffet"`
	IsAcceptScanOrder int    `gorm:"column:is_accept_scan_order;type:int(11);default:0;comment:是否开启扫码点餐接单 0不开启, 1开启;NOT NULL" json:"is_accept_scan_order"`
	IsOpenLocalPrint  int    `gorm:"column:is_open_local_print;type:int(11);default:1;comment:是否开启本地打印服务 0不开启, 1开启;NOT NULL" json:"is_open_local_print"`
	CashLimit         int    `gorm:"column:cash_limit;type:int(11);default:0;comment:收银机上限;NOT NULL" json:"cash_limit"`
	KitchenLimit      int    `gorm:"column:kitchen_limit;type:int(11);default:0;comment:厨显上限;NOT NULL" json:"kitchen_limit"`
	TabletLimit       int    `gorm:"column:tablet_limit;type:int(11);default:0;comment:平板上限;NOT NULL" json:"tablet_limit"`
	AssistantLimit    int    `gorm:"column:assistant_limit;type:int(11);default:0;comment:点餐助手上限;NOT NULL" json:"assistant_limit"`
	TableLimit        int    `gorm:"column:table_limit;type:int(11);default:0;comment:桌台上限;NOT NULL" json:"table_limit"`
	PrinterLimit      int    `gorm:"column:printer_limit;type:int(11);default:0;comment:打印机上限;NOT NULL" json:"printer_limit"`
	Timezone          string `gorm:"column:timezone;type:varchar(50);default:Asia/Shanghai;comment:时区;NOT NULL" json:"timezone"`
	Languages         string `gorm:"column:languages;type:varchar(255);comment:支持语言;NOT NULL" json:"languages"`
	Address           string `gorm:"column:address;type:varchar(255);comment:联系地址;NOT NULL" json:"address"`
	CreateTime        int    `gorm:"autoCreateTime;column:create_time;type:int(10);comment:创建时间（时间戳）;NOT NULL" json:"create_time"`
	UpdateTime        int    `gorm:"autoUpdateTime;column:update_time;type:int(10);comment:更新时间（时间戳）;NOT NULL" json:"update_time"`
	DeleteTime        int    `gorm:"column:delete_time;type:int(10);default:0;comment:删除时间（时间戳）;NOT NULL" json:"delete_time"`
}
