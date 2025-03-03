package model

// Member 会员信息表 `ttpos_member`
type Member struct {
	BaseModel
	MemberNo                     string  `gorm:"column:member_no;type:varchar(255);comment:会员编号;NOT NULL" json:"member_no"`
	Nickname                     string  `gorm:"column:nickname;type:varchar(255);comment:昵称;NOT NULL" json:"nickname"`
	Gender                       int     `gorm:"column:gender;type:tinyint(3);default:2;comment:性别,0-女 1-男 2-未知;NOT NULL" json:"gender"`
	Phone                        string  `gorm:"column:phone;type:varchar(20);comment:电话号码;NOT NULL" json:"phone"`
	Password                     string  `gorm:"column:password;type:varchar(200);comment:密码;NOT NULL" json:"password"`
	Birthday                     int64   `gorm:"column:birthday;type:int(10);comment:生日,时间戳" json:"birthday"`
	Point                        float64 `gorm:"column:point;type:decimal(12,2);default:0.00;comment:积分;NOT NULL" json:"point"`
	AccumulatedConsumptionAmount float64 `gorm:"column:accumulated_consumption_amount;type:decimal(12,2);default:0.00;comment:累计消费金额;NOT NULL" json:"accumulated_consumption_amount"`
	ConsumptionCount             int     `gorm:"column:consumption_count;type:int(11);default:0;comment:消费次数;NOT NULL" json:"consumption_count"`
	Balance                      float64 `gorm:"column:balance;type:decimal(12,2);default:0.00;comment:余额;NOT NULL" json:"balance"`
	GiftBalance                  float64 `gorm:"column:gift_balance;type:decimal(12,2);default:0.00;comment:赠送账户余额;NOT NULL" json:"gift_balance"`
	AccumulatedRechargeAmount    float64 `gorm:"column:accumulated_recharge_amount;type:decimal(12,2);default:0.00;comment:累计充值金额;NOT NULL" json:"accumulated_recharge_amount"`
	MemberLevelUuid              uint64  `gorm:"column:member_level_uuid;type:bigint(20) unsigned;default:0;comment:会员等级ID;NOT NULL" json:"member_level_uuid"`
	MemberCardUuid               uint64  `gorm:"column:member_card_uuid;type:bigint(20) unsigned;default:0;comment:会员卡片ID;NOT NULL" json:"member_card_uuid"`

	MemberLevel *MemberLevel `gorm:"foreignKey:MemberLevelUuid;references:Uuid"`
	MemberCard  *MemberCard  `gorm:"foreignKey:MemberCardUuid;references:Uuid"`
}

func (model *Member) HasPassword() bool {
	return model.Password != ""
}

// MemberLevel 会员等级表 `ttpos_member_level`
type MemberLevel struct {
	BaseModel
	Uuid         uint64  `gorm:"column:uuid;type:bigint(20) unsigned;default:0;comment:会员等级ID;NOT NULL" json:"uuid"`
	Name         string  `gorm:"column:name;type:varchar(255);comment:等级名称;NOT NULL" json:"name"`
	OpenMoney    int     `gorm:"column:open_money;type:tinyint(3);default:0;comment:是否开放累计消费额升级，0-否 1-是" json:"open_money"`
	UpgradeMoney float64 `gorm:"column:upgrade_money;type:decimal(12,2);default:0.00;comment:升级条件，累计消费额;NOT NULL" json:"upgrade_money"`
	OpenPoint    int     `gorm:"column:open_point;type:tinyint(3);default:0;comment:是否开放累计积分升级，0-否 1-是" json:"open_point"`
	UpgradePoint float64 `gorm:"column:upgrade_point;type:decimal(12,2);default:0.00;comment:升级条件，累计积分" json:"upgrade_point"`
	Discount     float64 `gorm:"column:discount;type:decimal(12,2);default:0;comment:等级权益,百分比折扣,单位%, 如80%为打8折，discount值为0.8;NOT NULL" json:"discount"`
	Priority     int     `gorm:"column:priority;type:int(11);default:0;comment:等级权重，越大等级越高;NOT NULL" json:"priority"`
	IsDefault    int     `gorm:"column:is_default;type:tinyint(1);default:0;comment:是否默认, 1-是 0-否;NOT NULL" json:"is_default"`
	Remark       string  `gorm:"column:remark;type:varchar(255);comment:备注;NOT NULL" json:"remark"`
}

// MemberCard 会员卡表 `ttpos_member_card`
type MemberCard struct {
	BaseModel
	CardTypeUuid uint64  `gorm:"column:card_type_uuid;type:bigint(20) unsigned;default:0;comment:会员卡类型ID;NOT NULL" json:"card_type_uuid"`
	MemberUuid   uint64  `gorm:"column:member_uuid;type:bigint(20) unsigned;default:0;comment:会员ID;NOT NULL" json:"member_uuid"`
	ExpireTime   int64   `gorm:"column:expire_time;type:int(11);default:0;comment:截止日期(时间戳);NOT NULL" json:"expire_time"`
	Discount     float64 `gorm:"column:discount;type:decimal(12, 2);default:0;comment:折扣,单位%, 如80%为打8折，discount值为0.8 .不随后台改变,按领取时的折扣。后续会员卡类型折扣改变时,不改变此字段;NOT NULL" json:"discount"`

	Member         *Member         `gorm:"foreignKey:MemberUuid;references:Uuid"`
	MemberCardType *MemberCardType `gorm:"foreignKey:CardTypeUuid;references:Uuid"`
}

// MemberCardType 会员卡类型表 `ttpos_member_card_type`
type MemberCardType struct {
	BaseModel
	Name         string  `gorm:"column:name;type:varchar(255);comment:会员卡类型名称;NOT NULL" json:"name"`
	Expire       int     `gorm:"column:expire;type:int(11);default:0;comment:有效期限,单位:月, 0为永久有效;NOT NULL" json:"expire"`
	Price        float64 `gorm:"column:price;type:decimal(12,2);default:0.00;comment:价格;NOT NULL" json:"price"`
	Discount     int     `gorm:"column:discount;type:tinyint(3);default:0;comment:折扣,单位%;NOT NULL" json:"discount"`
	Sort         int     `gorm:"column:sort;type:int(11);default:0;comment:排序;NOT NULL" json:"sort"`
	Status       int     `gorm:"column:status;type:tinyint(1);default:0;comment:状态, 0-开启 1-关闭;NOT NULL" json:"status"`
	OpenPoint    int     `gorm:"column:open_point;type:tinyint(1);default:0;comment:开卡赠送积分,0-否 1-是;NOT NULL" json:"open_point"`
	OpenPointNum float64 `gorm:"column:open_point_num;type:decimal(12,2);default:0.00;comment:开卡赠送积分数;NOT NULL" json:"open_point_num"`
	OpenMoney    int     `gorm:"column:open_money;type:tinyint(1);default:0;comment:开卡赠送余额,0-否 1-是;NOT NULL" json:"open_money"`
	OpenMoneyNum float64 `gorm:"column:open_money_num;type:decimal(12,2);default:0.00;comment:开卡赠送余额数;NOT NULL" json:"open_money_num"`
	Describe     string  `gorm:"column:describe;type:varchar(255);comment:使用须知;NOT NULL" json:"describe"`
}

// MemberCardLog 会员卡领取记录表 `ttpos_member_card_log`
type MemberCardLog struct {
	BaseModel
	Price              float64 `gorm:"column:price;type:decimal(12,2);default:0.00;comment:价格,会员卡价格,不随后台改变,记录领取时的价格;NOT NULL" json:"price"`
	Discount           int     `gorm:"column:discount;type:tinyint(3);default:0;comment:折扣,单位%,不随后台改变,记录领取时的折扣;NOT NULL" json:"discount"`
	Expire             int     `gorm:"column:expire;type:int(11);default:0;comment:有效期限,单位:月, 0为永久有效,不随后台改变,记录领取时的有效期限;NOT NULL" json:"expire"`
	MemberName         string  `gorm:"column:member_name;type:varchar(255);comment:会员名称,不随后台改变,当无法用member_uuid获取会员信息时,用此字段;NOT NULL" json:"member_name"`
	MemberPhone        string  `gorm:"column:member_phone;type:varchar(255);comment:会员电话,不随后台改变,当无法用member_uuid获取会员信息时,用此字段;NOT NULL" json:"member_phone"`
	MemberNo           string  `gorm:"column:member_no;type:varchar(255);comment:会员编号,不随后台改变,当无法用member_uuid获取会员信息时,用此字段;NOT NULL" json:"member_no"`
	MemberCardTypeName string  `gorm:"column:member_card_type_name;type:varchar(255);comment:会员卡类型名称,不随后台改变,当无法用member_card_type_uuid获取会员卡类型信息时,用此字段;NOT NULL" json:"member_card_type_name"`
	MemberCardTypeUuid uint64  `gorm:"column:member_card_type_uuid;type:bigint(20) unsigned;default:0;comment:会员卡类型ID;NOT NULL" json:"member_card_type_uuid"`
	MemberUuid         uint64  `gorm:"column:member_uuid;type:bigint(20) unsigned;default:0;comment:会员ID;NOT NULL" json:"member_uuid"`
}

// MemberBalanceLog 会员余额变动记录表 `ttpos_member_balance_log`
type MemberBalanceLog struct {
	BaseModel
	MemberUuid uint64  `gorm:"column:member_uuid;type:bigint(20) unsigned;default:0;comment:会员ID;NOT NULL" json:"member_uuid"`
	Scene      int     `gorm:"column:scene;type:tinyint(2);default:0;comment:场景,10-用户充值 20-用户消费 30-管理员操作 40-订单退款 50-余额提现 60-订单反结账 70-充值反结账 80-充值退款 90-扣减;NOT NULL" json:"scene"`
	Money      float64 `gorm:"column:money;type:decimal(12,2);default:0.00;comment:变动金额,负数:减余额 整数:加余额;NOT NULL" json:"money"`
	GiftMoney  float64 `gorm:"column:gift_money;type:decimal(12,2);default:0.00;comment:变动赠送金额" json:"gift_money"`
	Describe   string  `gorm:"column:describe;type:varchar(255);comment:变动描述;NOT NULL" json:"describe"`
}

// MemberPointLog 会员积分变动记录表 `ttpos_member_point_log`
type MemberPointLog struct {
	BaseModel
	MemberUuid uint64  `gorm:"column:member_uuid;type:bigint(20) unsigned;default:0;comment:会员ID;NOT NULL" json:"member_uuid"`
	Scene      int     `gorm:"column:scene;type:tinyint(2);default:0;comment:场景,10-用户充值 20-订单赠送 30-管理员操作 40-退款扣除 60-订单反结账 70-充值赠送 80-充值反结账 90-扣减;NOT NULL" json:"scene"`
	Value      float64 `gorm:"column:value;type:decimal(12,2);default:0;comment:数值,负数:减积分 正数:加积分;NOT NULL" json:"value"`
	Describe   string  `gorm:"column:describe;type:varchar(255);comment:变动描述;NOT NULL" json:"describe"`
}

// MemberRechargeOrder 会员充值订单表 `ttpos_member_recharge_order`
type MemberRechargeOrder struct {
	BaseModel
	OrderNo        string  `gorm:"column:order_no;type:varchar(255);comment:订单编号;NOT NULL" json:"order_no"`
	Status         int     `gorm:"column:status;type:tinyint(2);default:0;comment:状态,0-pending待支付 1-paid已支付 2-canceled已取消;NOT NULL" json:"status"`
	Amount         float64 `gorm:"column:amount;type:decimal(12,2);default:0;comment:交易金额=充值金额+手续费;NOT NULL" json:"amount"`
	ChargeDue      float64 `gorm:"column:charge_due;type:decimal(12,2);default:0;comment:找零;NOT NULL" json:"charge_due"`
	RechargeAmount float64 `gorm:"column:recharge_amount;type:decimal(12,2);default:0;comment:充值金额;NOT NULL" json:"recharge_amount"`
	GiftAmount     float64 `gorm:"column:gift_amount;type:decimal(12,2);default:0;comment:赠送金额;NOT NULL" json:"gift_amount"`
	GiftPoint      float64 `gorm:"column:gift_point;type:decimal(12,2);default:0;comment:赠送积分;NOT NULL" json:"gift_point"`
	MemberUuid     uint64  `gorm:"column:member_uuid;type:bigint(20) unsigned;comment:会员ID;NOT NULL" json:"member_uuid"`
	StaffUuid      uint64  `gorm:"column:staff_uuid;type:bigint(20) unsigned;comment:员工ID;NOT NULL" json:"staff_uuid"`
	PaymentTime    int64   `gorm:"column:payment_time;type:int(10) unsigned;default:0;comment:支付时间(时间戳);NOT NULL" json:"payment_time"`

	PaymentOrders []PaymentOrder `gorm:"foreignKey:RelatedUuid;references:Uuid"` // 一个会员充值订单关联多个支付订单
	Member        *Member        `gorm:"foreignKey:MemberUuid;references:Uuid"`  // 关联会员
	Staff         *Staff         `gorm:"foreignKey:StaffUuid;references:Uuid"`   // 关联操作员工

	RechargeOrderOperationLogs []MemberRechargeOrderOperationLog `gorm:"foreignKey:RechargeOrderUuid;references:Uuid"` // 一个充值订单关联多个操作日志
}

// MemberRechargeOrderOperationLog 会员充值订单操作记录表 `ttpos_member_recharge_order_operation_log`
type MemberRechargeOrderOperationLog struct {
	BaseModel
	OperatorName      string `gorm:"column:operator_name;type:varchar(50);comment:操作员姓名;NOT NULL" json:"operator_name"`
	OperatorEmail     string `gorm:"column:operator_email;type:varchar(50);comment:操作员电子邮件;NOT NULL" json:"operator_email"`
	Client            string `gorm:"column:client;type:varchar(50);comment:客户端信息;NOT NULL" json:"client"`
	Message           string `gorm:"column:message;type:varchar(255);comment:消息内容;NOT NULL" json:"message"`
	Action            string `gorm:"column:action;type:varchar(255);comment:操作类型;NOT NULL" json:"action"`
	Data              string `gorm:"column:data;type:varchar(255);comment:数据;NOT NULL" json:"data"`
	RechargeOrderUuid uint64 `gorm:"column:recharge_order_uuid;type:bigint(20) unsigned;default:0;comment:充值订单ID;NOT NULL" json:"recharge_order_uuid"`
}

type MemberLevelLog struct {
	BaseModel
	MemberUuid uint64 `gorm:"column:member_uuid;type:bigint(20) unsigned;default:0;comment:会员ID;NOT NULL" json:"member_uuid"`
	OldLevelId uint64 `gorm:"column:old_level_id;type:bigint(20) unsigned;default:0;comment:变更前的等级id;NOT NULL" json:"old_level_id"`
	NewLevelId uint64 `gorm:"column:new_level_id;type:bigint(20) unsigned;default:0;comment:变更后的等级id;NOT NULL" json:"new_level_id"`
	ChangeType uint   `gorm:"column:change_type;type:tinyint(3) unsigned;default:10;comment:变更类型(10后台管理员设置 20自动升级);NOT NULL" json:"change_type"`
	Remark     string `gorm:"column:remark;type:varchar(500);comment:管理员备注" json:"remark"`
}
