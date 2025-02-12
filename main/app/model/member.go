package model

// Member 会员信息表 ttpos_member
type Member struct {
	ID                 uint    `gorm:"column:id;primaryKey;autoIncrement;comment:'自增ID'" json:"id"`
	Uuid               uint64  `gorm:"column:uuid;type:bigint(20) unsigned;default:0;comment:'会员ID';NOT NULL" json:"uuid"`
	MemberNo           string  `gorm:"column:member_no;type:varchar(255);comment:'会员编号';NOT NULL" json:"member_no"`
	Nickname           string  `gorm:"column:nickname;type:varchar(255);comment:'昵称';NOT NULL" json:"nickname"`
	Gender             string  `gorm:"column:gender;type:varchar(10);comment:'性别';NOT NULL" json:"gender"`
	Phone              string  `gorm:"column:phone;type:varchar(20);comment:'电话号码';NOT NULL" json:"phone"`
	Password           string  `gorm:"column:password;type:varchar(20);comment:'密码';NOT NULL" json:"password"`
	Birthday           string  `gorm:"column:birthday;type:varchar(20);comment:'生日'" json:"birthday"`
	Point              float64 `gorm:"column:point;type:decimal(12,2);default:0.00;comment:'积分';NOT NULL" json:"point"`
	ConsumptionAmount  float64 `gorm:"column:accumulated_consumption_amount;type:decimal(12,2);default:0.00;comment:'累计消费金额';NOT NULL" json:"consumption_amount"`
	ConsumptionCount   int     `gorm:"column:consumption_count;type:int(11);default:0;comment:'消费次数';NOT NULL" json:"consumption_count"`
	Balance            float64 `gorm:"column:balance;type:decimal(12,2);default:0.00;comment:'余额';NOT NULL" json:"balance"`
	RechargeAmount     float64 `gorm:"column:accumulated_recharge_amount;type:decimal(12,2);default:0.00;comment:'累计充值金额';NOT NULL" json:"recharge_amount"`
	GiftAccountBalance float64 `gorm:"column:gift_account_balance;type:decimal(12,2);default:0.00;comment:'赠送账户余额';NOT NULL" json:"gift_account_balance"`
	MemberLevelUuid    uint64  `gorm:"column:member_level_uuid;type:bigint(20) unsigned;default:0;comment:'会员等级ID';NOT NULL" json:"member_level_uuid"`
	MemberCardUuid     uint64  `gorm:"column:member_card_uuid;type:bigint(20) unsigned;default:0;comment:'会员卡片ID';NOT NULL" json:"member_card_uuid"`
	CreateTime         int64   `gorm:"autoCreateTime;comment:'创建时间（时间戳）'" json:"create_time"`
	UpdateTime         int64   `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'" json:"update_time"`
	DeleteTime         int64   `gorm:"default:0;comment:'删除时间（时间戳）'" json:"delete_time"`
}

// MemberLevel 会员等级表 ttpos_member_level
type MemberLevel struct {
	ID                    uint   `gorm:"column:id;type:int(11) unsigned;primary_key;AUTO_INCREMENT;comment:自增ID" json:"id"`
	Uuid                  uint64 `gorm:"column:uuid;type:bigint(20) unsigned;default:0;comment:会员等级ID;NOT NULL" json:"uuid"`
	Name                  string `gorm:"column:name;type:varchar(255);comment:等级名称;NOT NULL" json:"name"`
	MultiLanguageNameUuid uint64 `gorm:"column:multi_language_name_uuid;type:bigint(20) unsigned;default:0;comment:多语言名称ID;NOT NULL" json:"multi_language_name_uuid"`
	OpenMoney             int    `gorm:"column:open_money;type:tinyint(3);default:0;comment:是否开放0，否1是" json:"open_money"`
	UpgradeMoney          int    `gorm:"column:upgrade_money;type:int(11);default:0;comment:升级条件;NOT NULL" json:"upgrade_money"`
	OpenPoints            int    `gorm:"column:open_points;type:tinyint(3);default:0;comment:积分是否开放0否1是" json:"open_points"`
	UpgradePoints         int    `gorm:"column:upgrade_points;type:int(11);default:0;comment:累计积分升级" json:"upgrade_points"`
	OpenInvite            int    `gorm:"column:open_invite;type:tinyint(3);default:0;comment:邀请是否开放0否1是" json:"open_invite"`
	UpgradeInvite         int    `gorm:"column:upgrade_invite;type:int(11);default:0;comment:邀请人数升级" json:"upgrade_invite"`
	Discount              int    `gorm:"column:discount;type:tinyint(3);default:0;comment:等级权益,百分比;NOT NULL" json:"discount"`
	Priority              int    `gorm:"column:priority;type:int(11);default:0;comment:等级权重;NOT NULL" json:"priority"`
	IsDefault             int    `gorm:"column:is_default;type:tinyint(1);default:0;comment:是否默认, 1-是 0-否;NOT NULL" json:"is_default"`
	Remark                string `gorm:"column:remark;type:varchar(255);comment:备注;NOT NULL" json:"remark"`
	CreateTime            int    `gorm:"autoCreateTime;column:create_time;type:int(10);comment:创建时间（时间戳）;NOT NULL" json:"create_time"`
	UpdateTime            int    `gorm:"autoUpdateTime;column:update_time;type:int(10);comment:更新时间（时间戳）;NOT NULL" json:"update_time"`
	DeleteTime            int    `gorm:"column:delete_time;type:int(10);default:0;comment:删除时间（时间戳）;NOT NULL" json:"delete_time"`
}
