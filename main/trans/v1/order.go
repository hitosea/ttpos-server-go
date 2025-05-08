package v1

import (
	"fmt"
	"ttpos-server-go/app/errors"

	"gorm.io/gorm"
)

type Order struct {
	OrderID                           uint64  `gorm:"primaryKey;autoIncrement;not null;comment:'订单id'"`
	ParentID                          int     `gorm:"default:0;comment:'父订单ID(拆单功能)'"`
	OrderName                         string  `gorm:"type:varchar(50);default:'';comment:'订单名称'"`
	IsMustNotice                      int     `gorm:"default:0;comment:'商品必点页是否弹出 0-否 1-是'"`
	ExtraTimes                        int     `gorm:"not null;default:0;comment:'送厨次数'"`
	MergeID                           string  `gorm:"type:varchar(20);default:'';comment:'合并id'"`
	IsMerge                           int     `gorm:"default:0;comment:'合并主单 0-否 1-是'"`
	MergeParentID                     int     `gorm:"default:0;comment:'合并父ID'"`
	OrderNo                           string  `gorm:"type:varchar(20);not null;default:'';unique;comment:'订单号'"`
	DeviceID                          string  `gorm:"type:varchar(50);default:'';comment:'来源设备id'"`
	SettleDeviceID                    string  `gorm:"type:varchar(50);default:'';comment:'结算设备id'"`
	IsBuffet                          uint    `gorm:"not null;default:0;comment:'是否自助餐 0-否 1-是'"`
	TotalPrice                        float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'商品总金额'"`
	TotalProductPrice                 float64 `gorm:"type:decimal(12,2);default:0.00;comment:'订单商品总价(原价)'"`
	OrderPrice                        float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'订单总金额'"`
	StayTime                          int     `gorm:"not null;default:0;comment:'挂单时间'"`
	IsStay                            int     `gorm:"not null;default:0;comment:'挂单状态 0-否 1-是'"`
	CouponID                          uint64  `gorm:"not null;default:0;comment:'优惠券id'"`
	CouponMoney                       float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'优惠券抵扣金额'"`
	CouponIDSys                       int     `gorm:"default:0;comment:'系统优惠券'"`
	CouponMoneySys                    float64 `gorm:"type:decimal(12,2);default:0.00;comment:'平台优惠券抵扣'"`
	PointsMoney                       float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'积分抵扣金额'"`
	PointsNum                         int     `gorm:"not null;default:0;comment:'积分抵扣数量'"`
	PayPrice                          float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'实际应收金额（不包含退款）'"`
	FreePayPrice                      float64 `gorm:"type:decimal(12,2);default:0.00;comment:'免单前单pay_price'"`
	ActualPrice                       float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'客户实付金额'"`
	ChangeDue                         float64 `gorm:"type:decimal(12,2);default:0.00;comment:'找零'"`
	OriginalPrice                     float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'订单原始价格'"`
	UpdatePrice                       float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'后台修改的订单金额（差价）'"`
	BuyerRemark                       string  `gorm:"type:mediumtext;not null;comment:'买家留言'"`
	PayType                           int     `gorm:"default:0;comment:'支付方式'"`
	PaySource                         string  `gorm:"type:varchar(20);default:'';comment:'支付来源,mp,wx'"`
	PayStatus                         uint8   `gorm:"not null;default:10;comment:'付款状态(10未付款 20已付款)'"`
	IsFree                            int     `gorm:"default:0;comment:'是否免单 0-否 1-免单，计入总销售额、优惠折扣 2-免单，不计入总销售额、优惠折扣'"`
	FreeRemark                        string  `gorm:"type:varchar(500);default:'';comment:'免单备注'"`
	PayTime                           uint64  `gorm:"not null;default:0;comment:'付款时间'"`
	PayEndTime                        int     `gorm:"default:0;comment:'支付截止时间'"`
	DeliveryType                      uint8   `gorm:"not null;default:10;comment:'配送方式(10外卖配送 20上门取30打包带走40店内就餐'"`
	DeliveryStatus                    uint8   `gorm:"not null;default:10;comment:'发货状态(10未发货 20已发货)'"`
	DeliveryTime                      uint64  `gorm:"not null;default:0;comment:'发货时间'"`
	ReceiptStatus                     uint8   `gorm:"not null;default:10;comment:'收货状态(10未收货 20已收货)'"`
	ReceiptTime                       uint64  `gorm:"not null;default:0;comment:'收货时间'"`
	OrderStatus                       uint8   `gorm:"not null;default:10;comment:'订单状态10=>进行中，20=>已经取消，30=>已完成'"`
	PointsBonus                       float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'赠送的积分数量'"`
	IsSettled                         uint8   `gorm:"not null;default:0;comment:'订单是否已结算(0未结算 1已结算)'"`
	TransactionID                     string  `gorm:"type:varchar(30);not null;default:'';comment:'微信支付交易号'"`
	OrderSource                       uint8   `gorm:"not null;default:10;comment:'订单来源(10普通 20收银台)'"`
	UserID                            uint64  `gorm:"not null;default:0;comment:'用户id'"`
	ShopSupplierID                    int     `gorm:"default:0;comment:'供应商id'"`
	SupplierMoney                     float64 `gorm:"type:decimal(12,2);default:0.00;comment:'供应商结算金额,支付金额-平台结算金额'"`
	SysMoney                          float64 `gorm:"type:decimal(12,2);default:0.00;comment:'平台结算金额'"`
	RoomID                            int     `gorm:"default:0;comment:'直播间id'"`
	CancelRemark                      string  `gorm:"type:varchar(200);not null;default:'';comment:'商家取消订单备注'"`
	VirtualAuto                       uint8   `gorm:"not null;default:0;comment:'是否自动发货1自动0手动'"`
	VirtualContent                    string  `gorm:"type:varchar(200);default:'';comment:'虚拟物品内容'"`
	BagPrice                          float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'包装费'"`
	Mealtime                          string  `gorm:"type:varchar(120);not null;default:'';comment:'送餐时间'"`
	RefundMoney                       float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'退款金额'"`
	RefundConsumptionTax              float64 `gorm:"type:decimal(12,2);default:0.00;comment:'退款消费税'"`
	OrderType                         uint8   `gorm:"not null;default:0;comment:'用餐方式0外卖1店内'"`
	FinanceID                         int     `gorm:"not null;default:0;comment:'对账id'"`
	TableNo                           string  `gorm:"type:varchar(50);default:'';comment:'就餐桌号'"`
	IsDelete                          uint8   `gorm:"not null;default:0;comment:'是否删除'"`
	DeliverSource                     uint8   `gorm:"not null;default:10;comment:'10商家配送20达达30配送员'"`
	DeliverStatus                     uint8   `gorm:"not null;default:0;comment:'配送状态，待接单＝1,待取货＝2,配送中＝3,已完成＝4'"`
	DriverID                          int     `gorm:"not null;default:0;comment:'配送员id'"`
	TakeFee                           float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'骑手服务费'"`
	DiscountMoney                     float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'优惠金额'"`
	DiscountChangePrice               float64 `gorm:"type:decimal(12,2);default:0.00;comment:'改价金额'"`
	IsChangePrice                     int     `gorm:"not null;default:0;comment:'是否优惠折扣 0-否 1-是'"`
	SmallDiscountType                 int     `gorm:"default:0;comment:'抹零(优惠折扣)：1-抹分 2-抹角 3-四舍五入到角 4-四舍五入到元'"`
	SmallDiffMoney                    float64 `gorm:"type:decimal(12,2);default:0.00;comment:'抹零后与pay_price差值'"`
	SmallAuto                         uint8   `gorm:"not null;default:0;comment:'是否自动抹零'"`
	CheckoutDiffMoney                 float64 `gorm:"type:decimal(12,2);default:0.00;comment:'结账抹零后与pay_price差值'"`
	CheckoutDiscountType              int     `gorm:"default:0;comment:'结账抹零：0-实款实收 1-抹分 2-抹角 5-抹元'"`
	DiscountRatio                     int     `gorm:"not null;default:0;comment:'优惠折扣比例 如：50-百分之五十'"`
	DiscountMethod                    int     `gorm:"default:10;comment:'折扣计算方式 10-按百分比 20-直接减免'"`
	CashierID                         int     `gorm:"not null;default:0;comment:'收银员id'"`
	CallNo                            string  `gorm:"type:varchar(30);not null;default:'';comment:'取餐号'"`
	EatType                           uint8   `gorm:"not null;default:0;comment:'店内用处类型10堂食20快餐'"`
	TableID                           int     `gorm:"not null;default:0;comment:'桌位id'"`
	TableRemark                       string  `gorm:"type:varchar(100);default:'';comment:'桌台备注'"`
	MealNum                           uint    `gorm:"not null;default:0;comment:'就餐人数'"`
	ServiceMoney                      float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'服务费'"`
	ServiceType                       uint8   `gorm:"not null;default:0;comment:'服务费类型0按就餐人数1按桌台收费'"`
	SettleType                        uint8   `gorm:"not null;default:10;comment:'计算模式10先结账后用餐20先用餐后结账'"`
	AutoClose                         uint8   `gorm:"not null;default:1;comment:'0定时清台1立即清台'"`
	CloseTime                         int     `gorm:"not null;default:0;comment:'清台时间'"`
	SurplusBalance                    float64 `gorm:"type:decimal(12,2);default:0.00;comment:'会员剩余余额'"`
	Balance                           float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'余额抵扣金额'"`
	OnlineMoney                       float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'在线支付金额'"`
	TradeNo                           string  `gorm:"type:varchar(30);not null;default:'';comment:'支付订单号'"`
	AppID                             uint64  `gorm:"not null;default:0;comment:'小程序id'"`
	SettingServiceMoney               float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'门店设定服务费'"`
	ConsumptionTaxMoney               float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'门店设定消费税'"`
	OriginalConsumptionTaxMoney       float64 `gorm:"type:decimal(12,2);default:0.00;comment:'原价消费税'"`
	TotalProductConsumptionTax        float64 `gorm:"type:decimal(12,2);default:0.00;comment:'商品消费税(折后)'"`
	TotalProductServiceConsumptionTax float64 `gorm:"type:decimal(12,2);default:0.00;comment:'商品服务费消费税(折后)'"`
	TotalProductServiceFee            float64 `gorm:"type:decimal(12,2);default:0.00;comment:'商品服务费(折后)'"`
	ConsumptionTaxType                int     `gorm:"default:2;comment:'消费税类型：0关闭, 1商品已含税, 2商品未含税'"`
	UserDiscountMoney                 float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'会员折扣金额'"`
	PayFeeMoney                       float64 `gorm:"type:decimal(12,2);default:0.00;comment:'总支付手续费'"`
	IsLock                            uint    `gorm:"not null;default:0;comment:'锁定 0-否 1-是'"`
	LockTime                          int     `gorm:"default:0;comment:'锁单时间'"`
	PrintNum                          int     `gorm:"default:0;comment:'打印数量'"`
	BuffetExpiredTime                 int     `gorm:"not null;default:0;comment:'自助餐到期时间'"`
	BuffetStartTime                   int     `gorm:"not null;default:0;comment:'自助餐开始时间'"`
	LastBuffetTimeLimit               int     `gorm:"default:0;comment:'上一个自助餐限制时间'"`
	CreateTime                        int64   `gorm:"not null;default:0;comment:'创建时间'"`
	UpdateTime                        int64   `gorm:"not null;default:0;comment:'更新时间'"`
	DeleteTime                        int64   `gorm:"default:0;comment:'删除时间'"`

	OrderProducts      []OrderProduct      `gorm:"foreignKey:OrderID;references:OrderID"`
	OrderPayTypes      []OrderPayType      `gorm:"foreignKey:OrderID;references:OrderID"`
	OrderOperationLogs []OrderOperationLog `gorm:"foreignKey:OrderID;references:OrderID"`
	ShopUser           ShopUser            `gorm:"foreignKey:ShopUserID;references:CashierID"`
	SaleOrders         []Order             `gorm:"foreignKey:OrderID;references:ParentID"`
}

func (o *Order) IsSplitOrder() uint {
	if o.ParentID > 0 {
		return 1
	}
	return 0
}

func (o *Order) IsFreeOrder() uint {
	if o.IsFree > 0 {
		return 1
	}
	return 0
}

// 订单状态10=>进行中，20=>已经取消，30=>已完成
// 订单状态, 0-待付款、1-已完成、2-已取消
func (o *Order) Status() uint {
	switch o.OrderStatus {
	case 10:
		return 0
	case 20:
		return 2
	case 30:
		return 1
	}
	return 0
}

// 账单类型, 账单类型, 0-桌台订单、1-点餐订单
func (o *Order) BillType() uint {
	if o.TableID > 0 {
		return 0
	}
	return 1
}

// 店内用处类型10堂食20快餐
// 用餐方式,0-堂食 1-打包
func (o *Order) DiningMethod() uint {
	if o.EatType == 10 {
		return 0
	}
	return 1
}

type OrderRepository interface {
	GetOrderList() ([]*Order, error)
	ConvertOrder() error
}

type OrderService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *OrderService) GetOrderList() ([]*Order, error) {
	var orders []*Order
	err := s.db.Find(&orders).Error
	return orders, err
}

func (s *OrderService) ConvertOrder() error {
	orders, err := s.GetOrderList()
	if err != nil {
		return errors.WithMessage(err)
	}
	for _, order := range orders {
		fmt.Println(fmt.Sprintf("order: %+v", order))

	}
	return nil
}
