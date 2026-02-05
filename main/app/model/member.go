package model

import (
	"time"
	"ttpos-server-go/app/constant"

	"github.com/shopspring/decimal"
)

// Member 会员信息表 `ttpos_member`
type Member struct {
	BaseModel
	MemberNo                       string  `gorm:"column:member_no;type:varchar(255);comment:会员编号;NOT NULL" json:"member_no"`
	Nickname                       string  `gorm:"column:nickname;type:varchar(255);comment:昵称;NOT NULL" json:"nickname"`
	Gender                         int     `gorm:"column:gender;type:tinyint(3);default:0;comment:性别,0-女 1-男 2-未知;NOT NULL" json:"gender"`
	Phone                          string  `gorm:"column:phone;type:varchar(20);comment:电话号码;NOT NULL" json:"phone"`
	IsVisitor                      bool    `gorm:"column:is_visitor;type:tinyint(1);default:0;comment:是否游客,0-否 1-是;NOT NULL" json:"is_visitor"`
	DeviceId                       string  `gorm:"column:device_id;type:varchar(100);comment:设备ID,用于标识游客;NOT NULL" json:"device_id"`
	Password                       string  `gorm:"column:password;type:varchar(200);comment:密码;NOT NULL" json:"password"`
	Birthday                       int64   `gorm:"column:birthday;type:int(10);comment:生日,时间戳" json:"birthday"`
	Point                          float64 `gorm:"column:point;type:decimal(12,2);default:0.00;comment:积分;NOT NULL" json:"point"`
	FrozenPoint                    float64 `gorm:"column:frozen_point;type:decimal(12,2);default:0.00;comment:冻结积分。冻结积分不能使用，在前端显示为已扣除或已增加。冻结积分可为负数。积分余额=积分+冻结积分;NOT NULL" json:"frozen_point"`
	AccumulatedGetPoint            float64 `gorm:"column:accumulated_get_point;type:decimal(12,2);default:0.00;comment:累计获取积分;NOT NULL" json:"accumulated_get_point"`
	AccumulatedConsumptionGetPoint float64 `gorm:"column:accumulated_consumption_get_point;type:decimal(12,2);default:0.00;comment:累计消费获取积分;NOT NULL" json:"accumulated_consumption_get_point"`
	AccumulatedConsumptionAmount   float64 `gorm:"column:accumulated_consumption_amount;type:decimal(12,2);default:0.00;comment:累计消费金额;NOT NULL" json:"accumulated_consumption_amount"`
	ConsumptionCount               int     `gorm:"column:consumption_count;type:int(11);default:0;comment:消费次数;NOT NULL" json:"consumption_count"`
	Balance                        float64 `gorm:"column:balance;type:decimal(12,2);default:0.00;comment:余额;NOT NULL" json:"balance"`
	FrozenBalance                  float64 `gorm:"column:frozen_balance;type:decimal(12,2);default:0.00;comment:冻结余额。冻结余额不能使用，在前端显示为已扣除或已增加。冻结余额可为负数。会员余额=余额+冻结余额;NOT NULL" json:"frozen_balance"`
	GiftBalance                    float64 `gorm:"column:gift_balance;type:decimal(12,2);default:0.00;comment:赠送账户余额;NOT NULL" json:"gift_balance"`
	FrozenGiftBalance              float64 `gorm:"column:frozen_gift_balance;type:decimal(12,2);default:0.00;comment:冻结赠送账户余额。冻结赠送账户余额不能使用，在前端显示为已扣除或已增加。冻结赠送账户余额可为负数。赠送账户余额=赠送账户余额+冻结赠送账户余额;NOT NULL" json:"frozen_gift_balance"`
	AccumulatedRechargeAmount      float64 `gorm:"column:accumulated_recharge_amount;type:decimal(12,2);default:0.00;comment:累计充值金额;NOT NULL" json:"accumulated_recharge_amount"`
	MemberLevelUuid                uint64  `gorm:"column:member_level_uuid;type:bigint(20) unsigned;default:0;comment:会员等级ID;NOT NULL" json:"member_level_uuid"`
	MemberCardUuid                 uint64  `gorm:"column:member_card_uuid;type:bigint(20) unsigned;default:0;comment:会员卡片ID;NOT NULL" json:"member_card_uuid"`
	MemberCardNo                   string  `gorm:"column:member_card_no;type:varchar(255);comment:会员卡号;NOT NULL" json:"member_card_no"`
	ReferrerUuid                   uint64  `gorm:"column:referrer_uuid;type:bigint(20) unsigned;default:0;comment:推荐人ID;NOT NULL" json:"referrer_uuid"`
	ActivityUuid                   uint64  `gorm:"column:activity_uuid;type:bigint(20) unsigned;default:0;comment:活动ID;NOT NULL" json:"activity_uuid"`

	MemberLevel       *MemberLevel       `gorm:"foreignKey:MemberLevelUuid;references:Uuid"`
	MemberCard        *MemberCard        `gorm:"foreignKey:MemberCardUuid;references:Uuid"`
	MemberBalanceLog  []MemberBalanceLog `gorm:"foreignKey:MemberUuid;references:Uuid"`
	MarketingActivity *MarketingActivity `gorm:"foreignKey:Uuid;references:ActivityUuid"`
}

func (model *Member) SetNil() {
	model.MemberLevel = nil
	model.MemberCard = nil
}

// 是否存在推荐人和活动ID
func (model *Member) IsExistActivityAndReferrer() bool {
	return model.ReferrerUuid != 0 && model.ActivityUuid != 0
}

// 获取会员的积分余额，用于展示在前端。
// 会员积分余额=积分+冻结积分
func (model *Member) GetPoints() float64 {
	return decimal.NewFromFloat(model.Point).Add(decimal.NewFromFloat(model.FrozenPoint)).InexactFloat64()
}

// 变动积分. 扣减积分时，参数points为负数；增加积分时，参数points为正数
func (model *Member) ChangePoint(points float64) {
	model.FrozenPoint = decimal.NewFromFloat(model.FrozenPoint).Add(decimal.NewFromFloat(points)).InexactFloat64()
}

// 更改会员积分。仅用于处理积分变动记录时，修改会员积分。
// 参数changePoints为积分变动值，正数为增加积分，负数为扣减积分。
// 清零冻结的积分。表示该会员的积分变动已经处理完，无需再冻结。
func (model *Member) UpdatePoint(changePoints float64, changePointsConsumption float64) {
	model.AccumulatedGetPoint = decimal.NewFromFloat(model.AccumulatedGetPoint).Add(decimal.NewFromFloat(changePoints)).InexactFloat64()
	model.AccumulatedConsumptionGetPoint = decimal.NewFromFloat(model.AccumulatedConsumptionGetPoint).Add(decimal.NewFromFloat(changePointsConsumption)).InexactFloat64()
	model.Point = decimal.NewFromFloat(model.Point).Add(decimal.NewFromFloat(changePoints)).InexactFloat64()
	model.FrozenPoint = 0
}

// 更改会员余额。仅用于处理积分变动记录时，修改会员积分。
// 参数changeBalance为余额变动值，正数为增加余额，负数为扣减余额。
// 清零冻结的余额。表示该会员的余额变动已经处理完，无需再冻结。
func (model *Member) UpdateBalance(changeBalance, changeBalanceGift float64) {
	model.Balance = decimal.NewFromFloat(model.Balance).Add(decimal.NewFromFloat(changeBalance)).InexactFloat64()
	model.FrozenBalance = 0
	model.GiftBalance = decimal.NewFromFloat(model.GiftBalance).Add(decimal.NewFromFloat(changeBalanceGift)).InexactFloat64()
	model.FrozenGiftBalance = 0
}

// 设置会员冻结余额
// 参数balanceAmount为本销售订单使用会员余额支付的金额
// 参数deductRatioMain为主账户扣款比例，参数deductRatioGift为赠送账户扣款比例。
func (model *Member) SetFrozenBalance(balanceAmount float64, deductRatioMain float64, deductRatioGift float64) (float64, float64) {
	defer model.SetUpdate()
	// 主账户扣款
	// 当deductRatioMain为0时，balanceAmountMain1为0，balanceAmountGift1为0
	balanceAmountMain1, balanceAmountGift1 := model.getBalanceAmountMain(balanceAmount, deductRatioMain)
	// 赠送账户扣款
	// 当deductRatioGift为0时，balanceAmountMain2为0，balanceAmountGift2为0
	balanceAmountMain2, balanceAmountGift2 := model.getBalanceAmountGift(balanceAmount, deductRatioGift)

	// 主账户扣款=主账户扣款+赠送帐户余额不足时从主账户扣款。
	balanceAmountMain := balanceAmountMain1 + balanceAmountMain2
	// 赠送账户扣款=赠送账户扣款+赠送帐户余额不足时从主账户扣款。
	balanceAmountGift := balanceAmountGift1 + balanceAmountGift2

	// 冻结余额=冻结余额-主账户扣款
	model.FrozenBalance = decimal.NewFromFloat(model.FrozenBalance).Sub(decimal.NewFromFloat(balanceAmountMain)).InexactFloat64()
	// 冻结赠送账户余额=冻结赠送账户余额-赠送账户扣款
	model.FrozenGiftBalance = decimal.NewFromFloat(model.FrozenGiftBalance).Sub(decimal.NewFromFloat(balanceAmountGift)).InexactFloat64()
	// 返回主账户扣款金额和赠送账户扣款金额
	return balanceAmountMain, balanceAmountGift
}

// 获取主账户扣款金额
func (model *Member) getBalanceAmountMain(balanceAmount float64, deductRatioMain float64) (float64, float64) {
	balanceAmountMain := float64(0)
	balanceAmountGift := float64(0)
	// 主账户扣款
	balanceAmountMain = decimal.NewFromFloat(balanceAmount).Mul(decimal.NewFromFloat(deductRatioMain)).InexactFloat64()
	// 如果主账户金额不足，则不足部分从赠送账户扣款
	if balanceAmountMain > model.GetBalance() {
		balanceAmountGift = balanceAmountMain - model.GetBalance()
		balanceAmountMain = model.GetBalance()
	}
	return balanceAmountMain, balanceAmountGift
}

// 获取赠送账户扣款金额
func (model *Member) getBalanceAmountGift(balanceAmount float64, deductRatioGift float64) (float64, float64) {
	balanceAmountMain := float64(0)
	balanceAmountGift := float64(0)
	// 赠送账户扣款
	balanceAmountGift = decimal.NewFromFloat(balanceAmount).Mul(decimal.NewFromFloat(deductRatioGift)).InexactFloat64()
	// 如果赠送账户金额不足，则不足部分从主账户扣款
	if balanceAmountGift > model.GetGiftBalance() {
		balanceAmountMain = balanceAmountGift - model.GetGiftBalance()
		balanceAmountGift = model.GetGiftBalance()
	}
	return balanceAmountMain, balanceAmountGift
}

// 获取会员的余额，用于展示在前端。
// 会员余额=余额+赠送余额
func (model *Member) GetBalanceAll() float64 {
	// 会员余额=余额+赠送余额
	return decimal.NewFromFloat(model.GetBalance()).Add(decimal.NewFromFloat(model.GetGiftBalance())).InexactFloat64()
}

// 获取会员的余额，用于展示在前端。
// 会员余额=余额+冻结余额
func (model *Member) GetBalance() float64 {
	return decimal.NewFromFloat(model.Balance).Add(decimal.NewFromFloat(model.FrozenBalance)).Truncate(2).InexactFloat64()
}

// 获取会员的赠送余额，用于展示在前端。
// 会员赠送余额=赠送余额+冻结赠送余额
func (model *Member) GetGiftBalance() float64 {
	return decimal.NewFromFloat(model.GiftBalance).Add(decimal.NewFromFloat(model.FrozenGiftBalance)).Truncate(2).InexactFloat64()
}

func (model *Member) GetMemberCardName() string {
	if model.MemberCard == nil {
		return ""
	}
	if model.MemberCard.MemberCardType == nil {
		return ""
	}
	return model.MemberCard.MemberCardType.Name
}

func (model *Member) GetMemberLevelName() string {
	if model.MemberLevel == nil {
		return ""
	}
	return model.MemberLevel.Name
}

func (model *Member) GetMemberDiscountRate() float64 {
	discount := float64(1) // 默认不打折
	if model.MemberLevel != nil {
		discount = model.MemberLevel.GetDiscount()
		if model.IsVisitor {
			// 游客不享受折扣
			discount = constant.NoDiscount
		}
	}
	return discount
}

func (model *Member) GetMemberCardDiscountRate() float64 {
	discount := float64(1) // 默认不打折
	if model.MemberCard != nil {
		discount = model.MemberCard.GetDiscount()
		if model.IsVisitor {
			// 游客不享受折扣
			discount = constant.NoDiscount
		}
	}
	return discount
}

func (model *Member) HasPassword() bool {
	return model.Password != ""
}

// 获取会员的累计充值金额，用于展示在前端。
// 累计充值金额=客户端充值金额+后台充值金额-反结账金额
func (model *Member) GetRechargeMoney() float64 {
	var rechargeMoney decimal.Decimal
	for _, log := range model.MemberBalanceLog {
		if log.Scene == constant.MemberBalanceLogRecharge || log.Scene == constant.MemberBalanceLogAdmin {
			var addMoney = decimal.NewFromFloat(log.Money).Sub(decimal.NewFromFloat(log.GiftMoney))
			rechargeMoney = rechargeMoney.Add(addMoney)
		} else if log.Scene == constant.MemberBalanceLogRechargeReverse {
			var money = decimal.NewFromFloat(log.Money).Mul(decimal.NewFromFloat(-1))
			var giftMoney = decimal.NewFromFloat(log.GiftMoney).Mul(decimal.NewFromFloat(-1))
			var deductMoney = money.Sub(giftMoney)
			rechargeMoney = rechargeMoney.Sub(deductMoney)
		}
	}

	return rechargeMoney.InexactFloat64()
}

// MemberLevel 会员等级表 `ttpos_member_level`
type MemberLevel struct {
	BaseModel
	Name           string  `gorm:"column:name;type:varchar(255);comment:等级名称;NOT NULL" json:"name"`
	OpenMoney      int     `gorm:"column:open_money;type:tinyint(3);default:0;comment:是否开放累计消费额升级，0-否 1-是" json:"open_money"`
	UpgradeMoney   float64 `gorm:"column:upgrade_money;type:decimal(12,2);default:0.00;comment:升级条件，累计消费额;NOT NULL" json:"upgrade_money"`
	OpenPoint      int     `gorm:"column:open_point;type:tinyint(3);default:0;comment:是否开放累计积分升级，0-否 1-是" json:"open_point"`
	UpgradePoint   float64 `gorm:"column:upgrade_point;type:decimal(12,2);default:0.00;comment:升级条件，累计积分" json:"upgrade_point"`
	Discount       float64 `gorm:"column:discount;type:decimal(12,4);default:1;comment:等级权益,百分比折扣,单位%, 如80%为打8折，discount值为0.8;NOT NULL" json:"discount"`
	Priority       int     `gorm:"column:priority;type:int(11);default:0;comment:等级权重，越大等级越高;NOT NULL" json:"priority"`
	IsDefault      int     `gorm:"column:is_default;type:tinyint(1);default:0;comment:是否默认, 1-是 0-否;NOT NULL" json:"is_default"`
	Remark         string  `gorm:"column:remark;type:varchar(255);comment:备注;NOT NULL" json:"remark"`
	PointsRate     string  `gorm:"column:points_rate;type:varchar(255);comment:购物赠送积分按照付款金额比例赠送时的比例;NOT NULL" json:"points_rate"`
	PointsQuantity string  `gorm:"column:points_quantity;type:varchar(255);comment:购物赠送积分按照桌台人数赠送时的数量;NOT NULL" json:"points_quantity"`
}

func (model *MemberLevel) GetDiscount() float64 {
	// 兼容1-100的取值范围
	if model.Discount > 1 {
		return decimal.NewFromFloat(model.Discount).Div(decimal.NewFromUint64(100)).InexactFloat64()
	}
	// 不能为负数. 现在不能折扣为0, 所以默认100%不打折
	if model.Discount <= 0 {
		return 1
	}
	return model.Discount
}

// MemberCard 会员卡表 `ttpos_member_card`
type MemberCard struct {
	BaseModel
	CardTypeUuid uint64  `gorm:"column:card_type_uuid;type:bigint(20) unsigned;default:0;comment:会员卡类型ID;NOT NULL" json:"card_type_uuid"`
	MemberUuid   uint64  `gorm:"column:member_uuid;type:bigint(20) unsigned;default:0;comment:会员ID;NOT NULL" json:"member_uuid"`
	ExpireTime   int64   `gorm:"column:expire_time;type:int(11);default:0;comment:截止日期(时间戳);NOT NULL" json:"expire_time"`
	Discount     float64 `gorm:"column:discount;type:decimal(12, 4);default:1.0000;comment:折扣,单位%, 如80%为打8折，discount值为0.8 .不随后台改变,按领取时的折扣。后续会员卡类型折扣改变时,不改变此字段;NOT NULL" json:"discount"`

	Member         *Member         `gorm:"foreignKey:MemberUuid;references:Uuid"`
	MemberCardType *MemberCardType `gorm:"foreignKey:CardTypeUuid;references:Uuid"`
}

func (model *MemberCard) GetDiscount() float64 {
	// 如果会员卡有有效期时，会员卡过期，则折扣不生效
	if model.ExpireTime != 0 {
		if time.Now().Unix() > model.ExpireTime {
			return 1
		}
	}
	// 兼容1-100的取值范围
	if model.Discount > 1 {
		return decimal.NewFromFloat(model.Discount).Div(decimal.NewFromUint64(100)).InexactFloat64()
	}
	// 不能为负数. 现在不能折扣为0, 所以默认100%不打折
	if model.Discount <= 0 {
		return 1
	}
	return model.Discount
}

// MemberCardType 会员卡类型表 `ttpos_member_card_type`
type MemberCardType struct {
	BaseModel
	Name         string  `gorm:"column:name;type:varchar(255);comment:会员卡类型名称;NOT NULL" json:"name"`
	Expire       int64   `gorm:"column:expire;type:int(11);default:0;comment:有效期限,单位:月, 0为永久有效;NOT NULL" json:"expire"`
	Price        float64 `gorm:"column:price;type:decimal(12,2);default:0.00;comment:价格;NOT NULL" json:"price"`
	Discount     float64 `gorm:"column:discount;type:tinyint(3);default:0;comment:折扣,单位%;NOT NULL" json:"discount"`
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
	Discount           float64 `gorm:"column:discount;type:tinyint(3);default:0;comment:折扣,单位%,不随后台改变,记录领取时的折扣;NOT NULL" json:"discount"`
	Expire             int64   `gorm:"column:expire;type:int(11);default:0;comment:有效期限,单位:月, 0为永久有效,不随后台改变,记录领取时的有效期限;NOT NULL" json:"expire"`
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
	MemberUuid  uint64  `gorm:"column:member_uuid;type:bigint(20) unsigned;default:0;comment:会员ID;NOT NULL" json:"member_uuid"`
	Scene       int     `gorm:"column:scene;type:tinyint(2);default:0;comment:场景,10-用户充值 20-用户消费 30-管理员操作 40-订单退款 50-余额提现 60-订单反结账 70-充值反结账 80-充值退款 90-扣减;NOT NULL" json:"scene"`
	Money       float64 `gorm:"column:money;type:decimal(12,2);default:0.00;comment:变动金额,负数:减余额 正数:加余额。包含赠送余额;NOT NULL" json:"money"`
	GiftMoney   float64 `gorm:"column:gift_money;type:decimal(12,2);default:0.00;comment:变动赠送金额" json:"gift_money"`
	Describe    string  `gorm:"column:describe;type:varchar(255);comment:变动描述;NOT NULL" json:"describe"`
	Processed   uint64  `gorm:"column:processed;type:tinyint(1);default:0;comment:是否已处理,0-未处理 1-已处理. 用于处理会员余额变动，修改会员的余额并清0冻结的余额;NOT NULL" json:"processed"`
	RelatedUuid uint64  `gorm:"column:related_uuid;type:bigint(20) unsigned;default:0;comment:关联uuid. 表示余额变动记录关联的业务订单ID,可能是销售订单(场景20)、充值订单(场景10)、退款单(场景80)、退货单退款金额(场景40);NOT NULL" json:"related_uuid"`
}

// MemberPointLog 会员积分变动记录表 `ttpos_member_point_log`
type MemberPointLog struct {
	BaseModel
	MemberUuid  uint64  `gorm:"column:member_uuid;type:bigint(20) unsigned;default:0;comment:会员ID;NOT NULL" json:"member_uuid"`
	Scene       int     `gorm:"column:scene;type:tinyint(2);default:0;comment:场景,10-用户充值 20-订单赠送 30-管理员操作 40-退款扣除 60-订单反结账 70-充值赠送 80-充值反结账 90-扣减;NOT NULL" json:"scene"`
	Value       float64 `gorm:"column:value;type:decimal(12,2);default:0;comment:数值,负数:减积分 正数:加积分;NOT NULL" json:"value"`
	Describe    string  `gorm:"column:describe;type:varchar(255);comment:变动描述;NOT NULL" json:"describe"`
	RelatedUuid uint64  `gorm:"column:related_uuid;type:bigint(20) unsigned;default:0;comment:关联uuid. 表示积分变动记录关联的业务订单ID,可能是销售订单、充值订单、退款单、退货单退款金额;NOT NULL" json:"related_uuid"`
	Processed   uint64  `gorm:"column:processed;type:tinyint(1);default:0;comment:是否已处理,0-未处理 1-已处理. 用于处理积分变动，修改会员的积分并清0冻结的积分;NOT NULL" json:"processed"`
}

// MemberRechargeOrder 会员充值订单表 `ttpos_member_recharge_order`
type MemberRechargeOrder struct {
	BaseModel
	OrderNo          string  `gorm:"column:order_no;type:varchar(255);comment:充值订单编号;NOT NULL" json:"order_no"`
	DutyNo           string  `gorm:"column:duty_no;type:varchar(64);comment:当班编号;NOT NULL" json:"duty_no"`
	Status           int     `gorm:"column:status;type:tinyint(2);default:0;comment:状态,0-pending待付款 1-paid已完成 2-canceled已取消;NOT NULL" json:"status"`
	Amount           float64 `gorm:"column:amount;type:decimal(12,2);default:0.00;comment:交易金额=充值金额+手续费;NOT NULL" json:"amount"`
	RefundMoney      float64 `gorm:"column:refund_money;type:decimal(12,2);default:0.00;comment:退款金额，不大于amount;NOT NULL" json:"refund_money"`
	ChargeDue        float64 `gorm:"column:charge_due;type:decimal(12,2);default:0.00;comment:找零;NOT NULL" json:"charge_due"`
	RechargeAmount   float64 `gorm:"column:recharge_amount;type:decimal(12,2);default:0.00;comment:充值金额;NOT NULL" json:"recharge_amount"`
	RefundAmount     float64 `gorm:"column:refund_amount;type:decimal(12,2);default:0.00;comment:退款充值金额，不大于recharge_amount" json:"refund_amount"`
	GiftAmount       float64 `gorm:"column:gift_amount;type:decimal(12,2);default:0.00;comment:赠送金额;NOT NULL" json:"gift_amount"`
	GiftPoint        float64 `gorm:"column:gift_point;type:decimal(12,2);default:0.00;comment:赠送积分;NOT NULL" json:"gift_point"`
	MemberUuid       uint64  `gorm:"column:member_uuid;type:bigint(20) unsigned;comment:会员ID;NOT NULL" json:"member_uuid"`
	StaffUuid        uint64  `gorm:"column:staff_uuid;type:bigint(20) unsigned;comment:员工ID;NOT NULL" json:"staff_uuid"`
	PaymentTime      int64   `gorm:"column:payment_time;type:int(10) unsigned;default:0;comment:支付时间(时间戳);NOT NULL" json:"payment_time"`
	Balance          float64 `gorm:"column:balance;type:decimal(12,2);default:0.00;comment:充值前会员余额" json:"balance"`
	BalanceRecharged float64 `gorm:"column:balance_recharged;type:decimal(12,2);default:0.00;comment:充值后会员余额" json:"balance_recharged"`

	// ERP发票
	ErpProductsInvoiceName string `gorm:"column:erp_products_invoice_name;type:varchar(255);comment:商品发票名称" json:"erp_products_invoice_name"`
	// 反结账次数
	ReverseSettleCount uint `gorm:"column:reverse_settle_count;type:int(11);default:0;comment:反结账次数" json:"reverse_settle_count"`

	PaymentOrders              []PaymentOrder                    `gorm:"foreignKey:RelatedUuid;references:Uuid"`       // 一个会员充值订单关联多个支付订单
	Member                     *Member                           `gorm:"foreignKey:MemberUuid;references:Uuid"`        // 关联会员
	Staff                      *Staff                            `gorm:"foreignKey:StaffUuid;references:Uuid"`         // 关联操作员工
	ReturnOrders               []ReturnOrder                     `gorm:"foreignKey:RelatedOrderUuid;references:uuid"`  // 关联多个退货单
	RechargeOrderOperationLogs []MemberRechargeOrderOperationLog `gorm:"foreignKey:RechargeOrderUuid;references:Uuid"` // 一个充值订单关联多个操作日志
}

func (model *MemberRechargeOrder) SetNil() {
	model.PaymentOrders = nil
	model.Member = nil
	model.Staff = nil
	model.ReturnOrders = nil
	model.RechargeOrderOperationLogs = nil
}

// 获取应收金额 = 应收金额 - 减去退款
func (model *MemberRechargeOrder) GetReceivableAmount() float64 {
	payAmount := decimal.NewFromFloat(0)
	payAmount = payAmount.Add(decimal.NewFromFloat(model.Amount))
	payAmount = payAmount.Sub(decimal.NewFromFloat(model.RefundMoney))
	return payAmount.InexactFloat64()
}

// 获取手续费金额
func (model *MemberRechargeOrder) GetCommissionFee() float64 {
	return decimal.NewFromFloat(model.Amount).Sub(decimal.NewFromFloat(model.RechargeAmount)).Round(2).InexactFloat64()
}

// erp是否进行过反结账
func (model *MemberRechargeOrder) IsErpReverseSettle() bool {
	// 结账后MemberRechargeOrder中就会记录下这两个发票名称
	return model.ErpProductsInvoiceName != ""
}

// 剩余的充值金额。如充值10元，退款2元，则剩余8元
func (model *MemberRechargeOrder) GetRemainingAmount() float64 {
	return decimal.NewFromFloat(model.RechargeAmount).Sub(decimal.NewFromFloat(model.RefundAmount)).Round(2).InexactFloat64()
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

func (model *MemberRechargeOrderOperationLog) SetNil() {
}

type MemberLevelLog struct {
	BaseModel
	MemberUuid uint64 `gorm:"column:member_uuid;type:bigint(20) unsigned;default:0;comment:会员ID;NOT NULL" json:"member_uuid"`
	OldLevelId uint64 `gorm:"column:old_level_id;type:bigint(20) unsigned;default:0;comment:变更前的等级id;NOT NULL" json:"old_level_id"`
	NewLevelId uint64 `gorm:"column:new_level_id;type:bigint(20) unsigned;default:0;comment:变更后的等级id;NOT NULL" json:"new_level_id"`
	ChangeType uint   `gorm:"column:change_type;type:tinyint(3) unsigned;default:10;comment:变更类型(10后台管理员设置 20自动升级);NOT NULL" json:"change_type"`
	Remark     string `gorm:"column:remark;type:varchar(500);comment:管理员备注" json:"remark"`
}

// MemberRechargeOrderAbnormalRecord 会员充值订单异常记录 `ttpos_member_recharge_order_abnormal_record`
type MemberRechargeOrderAbnormalRecord struct {
	BaseModel
	// 基本信息
	RechargeOrderUuid uint64 `gorm:"column:recharge_order_uuid;type:bigint(20) unsigned;not null;default:0;comment:充值订单ID" json:"recharge_order_uuid"`
	DutyNo            string `gorm:"column:duty_no;type:varchar(64);not null;default:'';comment:当班编号" json:"duty_no"`
	Action            string `gorm:"column:action;type:varchar(150);not null;default:'';comment:行为" json:"action"`
	SubAction         string `gorm:"column:sub_action;type:varchar(150);not null;default:'';comment:自定义子行为" json:"sub_action"`
	Sign              string `gorm:"column:sign;type:varchar(255);not null;default:'';comment:操作签名" json:"sign"`
	Remark            string `gorm:"column:remark;type:varchar(255);not null;default:'';comment:备注" json:"remark"`
	CashierUuid       uint64 `gorm:"column:cashier_uuid;type:bigint(20) unsigned;not null;default:0;comment:收银员ID" json:"cashier_uuid"`
}
