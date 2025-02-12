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
