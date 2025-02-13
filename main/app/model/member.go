package model

// Member 会员信息表 `ttpos_member`
type Member struct {
	BaseModel
	MemberNo           string  `gorm:"column:member_no;type:varchar(255);comment:'会员编号';" json:"member_no"`
	Nickname           string  `gorm:"column:nickname;type:varchar(255);comment:'昵称';" json:"nickname"`
	Gender             int     `gorm:"column:gender;type:tinyint(1);default:2;comment:'性别,0-女 1-男 2-未知';" json:"gender"`
	Phone              string  `gorm:"column:phone;type:varchar(20);comment:'电话号码';" json:"phone"`
	Password           string  `gorm:"column:password;type:varchar(200);comment:'密码';" json:"password"`
	Birthday           int64   `gorm:"column:birthday;type:int(10);comment:'生日,时间戳';" json:"birthday"`
	Point              float64 `gorm:"column:point;type:decimal(12,2);default:0.00;comment:'积分';" json:"point"`
	ConsumptionAmount  float64 `gorm:"column:accumulated_consumption_amount;type:decimal(12,2);default:0.00;comment:'累计消费金额';" json:"consumption_amount"`
	ConsumptionCount   int     `gorm:"column:consumption_count;type:int(11);default:0;comment:'消费次数';" json:"consumption_count"`
	Balance            float64 `gorm:"column:balance;type:decimal(12,2);default:0.00;comment:'余额';" json:"balance"`
	RechargeAmount     float64 `gorm:"column:accumulated_recharge_amount;type:decimal(12,2);default:0.00;comment:'累计充值金额';" json:"recharge_amount"`
	GiftAccountBalance float64 `gorm:"column:gift_account_balance;type:decimal(12,2);default:0.00;comment:'赠送账户余额';" json:"gift_account_balance"`
	MemberLevelUuid    uint64  `gorm:"column:member_level_uuid;type:bigint(20) unsigned;default:0;comment:'会员等级ID';" json:"member_level_uuid"`
	MemberCardUuid     uint64  `gorm:"column:member_card_uuid;type:bigint(20) unsigned;default:0;comment:'会员卡片ID';" json:"member_card_uuid"`
}

// MemberLevel 会员等级表 `ttpos_member_level`
type MemberLevel struct {
	BaseModel
	Name          string `gorm:"column:name;type:varchar(255);comment:等级名称;"`
	OpenMoney     int    `gorm:"column:open_money;type:tinyint(3);default:0;comment:是否开放累计消费额升级，0-否 1-是"`
	UpgradeMoney  int    `gorm:"column:upgrade_money;type:int(11);default:0;comment:升级条件，累计消费额"`
	OpenPoints    int    `gorm:"column:open_points;type:tinyint(3);default:0;comment:是否开放累计积分升级，0-否 1-是"`
	UpgradePoints int    `gorm:"column:upgrade_points;type:int(11);default:0;comment:升级条件，累计积分"`
	OpenInvite    int    `gorm:"column:open_invite;type:tinyint(3);default:0;comment:是否开放邀请升级，0-否 1-是"`
	UpgradeInvite int    `gorm:"column:upgrade_invite;type:int(11);default:0;comment:升级条件，邀请人数"`
	Discount      int    `gorm:"column:discount;type:tinyint(3);default:0;comment:等级权益,百分比;"`
	Priority      int    `gorm:"column:priority;type:int(11);default:0;comment:等级权重;"`
	IsDefault     int    `gorm:"column:is_default;type:tinyint(1);default:0;comment:是否默认, 1-是 0-否;"`
	Remark        string `gorm:"column:remark;type:varchar(255);comment:备注;"`
}

// MemberCard 会员卡表 `ttpos_member_card`
type MemberCard struct {
	BaseModel
	CardTypeUuid uint64 `gorm:"column:card_type_uuid;type:bigint(20) unsigned;default:0;comment:会员卡类型ID;" json:"card_type_uuid"`
	MemberUuid   uint64 `gorm:"column:member_uuid;type:bigint(20) unsigned;default:0;comment:会员ID;" json:"member_uuid"`
	Deadline     int64  `gorm:"column:deadline;type:int(11);default:0;comment:截止日期(时间戳);" json:"deadline"`
	Discount     int    `gorm:"column:discount;type:tinyint(3);default:0;comment:折扣,单位%,不随后台改变,按领取时的折扣。后续会员卡类型折扣改变时,不改变此字段;" json:"discount"`
	Status       int    `gorm:"column:status;type:tinyint(1);default:0;comment:状态, 0-exp到期 1-valid有效 2-repeal作废,管理员点击作废按钮 3-cover覆盖,领取了新的会员卡;" json:"status"`
}

// MemberCardType 会员卡类型表 `ttpos_member_card_type`
type MemberCardType struct {
	BaseModel
	Name            string  `gorm:"column:name;type:varchar(255);comment:卡名称;"`
	Period          int     `gorm:"column:period;type:int(11);default:0;comment:有效期限,单位:月, 0为永久有效;"`
	Price           float64 `gorm:"column:price;type:decimal(12,2);default:0.00;comment:价格;"`
	Discount        int     `gorm:"column:discount;type:tinyint(3);default:0;comment:折扣,单位%;"`
	Count           int     `gorm:"column:count;type:int(11);default:0;comment:领取数量;"`
	Sort            int     `gorm:"column:sort;type:int(11);default:0;comment:排序;"`
	Status          int     `gorm:"column:status;type:tinyint(1);default:0;comment:状态, 0-开启 1-关闭;"`
	CardOpeningGift int     `gorm:"column:card_opening_gift;type:tinyint(2);default:0;comment:开卡赠送,0-point积分 1-balance余额;"`
	GiftValue       float64 `gorm:"column:gift_value;type:decimal(12,2);default:0.00;comment:赠送额;"`
	Description     string  `gorm:"column:description;type:varchar(255);comment:使用须知;"`
}

// MemberCardLog 会员卡领取记录表 `ttpos_member_card_log`
type MemberCardLog struct {
	BaseModel
	Price              float64 `gorm:"column:price;type:decimal(12,2);default:0.00;comment:价格,会员卡价格,不随后台改变,记录领取时的价格"`
	Discount           int     `gorm:"column:discount;type:tinyint(3);default:0;comment:折扣,单位%,不随后台改变,记录领取时的折扣"`
	Period             int     `gorm:"column:period;type:int(11);default:0;comment:有效期限,单位:月, 0为永久有效,不随后台改变,记录领取时的有效期限"`
	MemberName         string  `gorm:"column:member_name;type:varchar(255);default:'';comment:会员名称,不随后台改变,当无法用member_uuid获取会员信息时,用此字段"`
	MemberPhone        string  `gorm:"column:member_phone;type:varchar(255);default:'';comment:会员电话,不随后台改变,当无法用member_uuid获取会员信息时,用此字段"`
	MemberNo           string  `gorm:"column:member_no;type:varchar(255);default:'';comment:会员编号,不随后台改变,当无法用member_uuid获取会员信息时,用此字段"`
	MemberCardTypeName string  `gorm:"column:member_card_type_name;type:varchar(255);default:'';comment:会员卡类型名称,不随后台改变,当无法用member_card_type_uuid获取会员卡类型信息时,用此字段"`
	MemberCardTypeUuid uint64  `gorm:"column:member_card_type_uuid;type:bigint(20) unsigned;default:0;comment:会员卡类型ID"`
	MemberUuid         uint64  `gorm:"column:member_uuid;type:bigint(20) unsigned;default:0;comment:会员ID"`
}

// Membe	rBalanceLog 会员余额变动记录表 `ttpos_member_balance_log`
type MemberBalanceLog struct {
	BaseModel
	MemberUuid  uint64  `gorm:"column:member_uuid;type:bigint(20) unsigned;default:0;comment:会员ID"`
	Scene       int     `gorm:"column:scene;type:tinyint(2);default:0;comment:场景,10-用户充值 20-用户消费 30-管理员操作 40-订单退款 50-余额提现 60-订单反结账 70-充值反结账 80-充值退款 90-扣减"`
	Money       float64 `gorm:"column:money;type:decimal(12,2);default:0.00;comment:变动金额,负数:减余额 整数:加余额"`
	GiftMoney   float64 `gorm:"column:gift_money;type:decimal(12,2);default:0.00;comment:变动赠送金额"`
	Description string  `gorm:"column:description;type:varchar(255);default:'';comment:变动描述"`
}

// MemberPointLog 会员积分变动记录表 `ttpos_member_point_log`
type MemberPointLog struct {
	BaseModel
	MemberUuid  uint64 `gorm:"column:member_uuid;type:bigint(20) unsigned;default:0;comment:会员ID"`
	Scene       int    `gorm:"column:scene;type:tinyint(2);default:0;comment:场景,10-用户充值 20-订单赠送 30-管理员操作 40-退款扣除 60-订单反结账 70-充值赠送 80-充值反结账 90-扣减"`
	Value       int    `gorm:"column:value;type:int(11);default:0;comment:数值,负数:减积分 正数:加积分"`
	Description string `gorm:"column:description;type:varchar(255);default:'';comment:变动描述"`
}

// MemberRechargeOrder 会员充值订单表 `ttpos_member_recharge_order`
type MemberRechargeOrder struct {
	BaseModel
	Status         int     `gorm:"column:status;type:tinyint(2);default:0;comment:状态,0-pending待支付 1-paid已支付 2-canceled已取消 3-exp已过期"`
	Amount         float64 `gorm:"column:amount;type:decimal(12,2);default:0.00;comment:交易金额"`
	RechargeAmount float64 `gorm:"column:recharge_amount;type:decimal(12,2);default:0.00;comment:充值金额"`
	GiftAmount     float64 `gorm:"column:gift_amount;type:decimal(12,2);default:0.00;comment:赠送金额"`
	GiftPoint      int     `gorm:"column:gift_point;type:int(11);default:0;comment:赠送积分"`
	MemberUuid     uint64  `gorm:"column:member_uuid;type:bigint(20) unsigned;default:0;comment:会员ID"`
	StaffUuid      uint64  `gorm:"column:staff_uuid;type:bigint(20) unsigned;default:0;comment:员工ID"`
	PaymentTime    int64   `gorm:"column:payment_time;type:int(10);default:0;comment:支付时间(时间戳)"`
}

// MemberRechargeOrderOperationLog 会员充值订单操作记录表 `ttpos_member_recharge_order_operation_log`
type MemberRechargeOrderOperationLog struct {
	BaseModel
	OperatorName      string `gorm:"column:operator_name;type:varchar(50);default:'';comment:操作员姓名"`
	OperatorEmail     string `gorm:"column:operator_email;type:varchar(50);default:'';comment:操作员电子邮件"`
	Client            string `gorm:"column:client;type:varchar(50);default:'';comment:客户端信息"`
	Message           string `gorm:"column:message;type:varchar(255);default:'';comment:消息内容"`
	RechargeOrderUuid uint64 `gorm:"column:recharge_order_uuid;type:bigint(20) unsigned;default:0;comment:充值订单ID"`
}
