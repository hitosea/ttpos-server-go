package v1

type OrderProduct struct {
	OrderProductID                       uint64  `gorm:"primaryKey;autoIncrement;not null;comment:'主键id'"`
	MainOrderProductID                   int     `gorm:"default:0;comment:'主订单商品ID'"`
	ProductID                            uint    `gorm:"not null;default:0;comment:'商品id'"`
	Delivery                             int     `gorm:"default:0;comment:'消费方式：10-外卖配送 20-上门取 30-打包带走 40-店内就餐'"`
	IsBuffetProduct                      int     `gorm:"not null;default:0;comment:'是否自助餐商品 0-否 1-是'"`
	ProductName                          string  `gorm:"type:varchar(2000);not null;default:'';comment:'产品名称'"`
	SendKitchenTime                      int     `gorm:"not null;default:0;comment:'送厨时间'"`
	IsSendKitchen                        int     `gorm:"not null;default:0;comment:'送厨状态 0-否 1-是'"`
	IsRequire                            int     `gorm:"default:0;comment:'是否必点商品 0-否 1-是'"`
	IsReturn                             int     `gorm:"default:0;comment:'是否退菜 0-否 1-是'"`
	SendKitchenSource                    int     `gorm:"default:1;comment:'送厨来源 1-收银 2-平板 3-扫码h5'"`
	AddSource                            int     `gorm:"default:1;comment:'添加来源 1-收银 2-平板 3-扫码h5'"`
	FinishNum                            int     `gorm:"not null;default:0;comment:'厨房完成数量'"`
	FinishTime                           int     `gorm:"not null;default:0;comment:'厨房完成时间'"`
	ImageID                              uint    `gorm:"not null;default:0;comment:'商品封面图id'"`
	DeductStockType                      uint8   `gorm:"not null;default:20;comment:'库存计算方式(10下单减库存 20付款减库存)'"`
	SpecType                             uint8   `gorm:"not null;default:0;comment:'规格类型(10单规格 20多规格)'"`
	SpecSkuID                            string  `gorm:"type:varchar(255);not null;default:'';comment:'商品sku标识'"`
	ProductSkuID                         uint    `gorm:"not null;default:0;comment:'商品规格id'"`
	ProductAttr                          string  `gorm:"type:longtext;default:'';comment:'商品规格信息'"`
	Content                              string  `gorm:"type:longtext;not null;comment:'商品详情'"`
	ProductNo                            string  `gorm:"type:varchar(100);not null;default:'';comment:'商品编码'"`
	ProductPrice                         float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'商品价格(单价)'"`
	ProductDiscountMoney                 float64 `gorm:"type:decimal(12,5);default:0.00000;comment:'优惠折扣后与原价总差额(包含数量)'"`
	TaxRate                              float64 `gorm:"type:decimal(12,4);default:0.0000;comment:'消费税率'"`
	ConsumptionTax                       float64 `gorm:"type:decimal(12,2);default:null"`
	TaxCalcType                          int     `gorm:"default:0;comment:'是否含税 0-关闭 1-已含税 2-未含税'"`
	ProductOriginalConsumptionTax        float64 `gorm:"type:decimal(12,2);default:0.00;comment:'商品消费税(原价)'"`
	ProductOriginalServiceConsumptionTax float64 `gorm:"type:decimal(12,2);default:0.00;comment:'商品服务费消费税(原价)'"`
	ProductOriginalServiceFee            float64 `gorm:"type:decimal(12,2);default:0.00;comment:'商品服务费(原价)'"`
	ProductConsumptionTax                float64 `gorm:"type:decimal(12,2);default:0.00;comment:'商品消费税(折后)'"`
	ProductServiceConsumptionTax         float64 `gorm:"type:decimal(12,2);default:0.00;comment:'商品服务费消费税(折后)'"`
	ProductServiceFee                    float64 `gorm:"type:decimal(12,2);default:0.00;comment:'商品服务费(折后)'"`
	ProductServiceRate                   float64 `gorm:"type:decimal(12,2);default:0.00;comment:'商品服务费率'"`
	NoFreeProductServiceConsumptionTax   float64 `gorm:"type:decimal(12,2);default:0.00;comment:'免赠前商品服务费消费税'"`
	NoFreeProductServiceFee              float64 `gorm:"type:decimal(12,2);default:0.00;comment:'免赠前商品服务费'"`
	NoFreeProductConsumptionTax          float64 `gorm:"type:decimal(12,2);default:0.00;comment:'免赠前商品消费税'"`
	IsChangePrice                        int     `gorm:"default:0;comment:'是否改价 0-否 1-是 '"`
	IsFree                               int     `gorm:"default:0;comment:'是否免单 0-否 1-免单，计入总销售额、优惠折扣 2-免单，不计入总销售额、优惠折扣'"`
	FreeRemark                           string  `gorm:"type:varchar(500);default:'';comment:'免单备注'"`
	IsMove                               int     `gorm:"default:0;comment:'是否转菜 0-否 1-是'"`
	MoveFromTableID                      int     `gorm:"default:0;comment:'转菜来源桌台ID'"`
	MoveFromOrderID                      int     `gorm:"default:0;comment:'转菜来源订单ID'"`
	LinePrice                            float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'商品划线价'"`
	ProductWeight                        float64 `gorm:"type:double;not null;default:0;comment:'商品重量(Kg)'"`
	IsUserGrade                          uint8   `gorm:"not null;default:0;comment:'是否存在会员等级折扣'"`
	GradeRatio                           uint8   `gorm:"not null;default:0;comment:'会员折扣比例(0-10)'"`
	GradeProductPrice                    float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'会员折扣的商品单价'"`
	GradeTotalMoney                      float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'会员折扣的总额差'"`
	CouponMoneySys                       float64 `gorm:"type:decimal(12,2);default:0.00;comment:'平台优惠券抵扣'"`
	CouponMoney                          float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'优惠券折扣金额'"`
	PointsMoney                          float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'积分金额'"`
	PointsNum                            uint    `gorm:"not null;default:0;comment:'积分抵扣数量'"`
	PointsBonus                          float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'赠送的积分数量'"`
	TotalNum                             uint    `gorm:"not null;default:0;comment:'购买数量'"`
	TotalPrice                           float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'商品总价(数量×单价)'"`
	FeedPrice                            float64 `gorm:"type:decimal(12,2);default:0.00;comment:'加料价格'"`
	FeedIDs                              string  `gorm:"type:varchar(255);default:'';comment:'加料ids'"`
	AttrIDs                              string  `gorm:"type:varchar(500);default:'';comment:'已选的商品属性ID'"`
	FeedUUIDs                            string  `gorm:"type:longtext;default:null;comment:'加料uuid'"`
	TotalProductPrice                    float64 `gorm:"type:decimal(12,2);default:0.00;comment:'订单商品总价(单价x数量原价)'"`
	TotalPayPrice                        float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'实际付款价(折扣和优惠后)'"`
	NoFreeTotalPayPrice                  float64 `gorm:"type:decimal(12,2);default:0.00;comment:'不受免单影响的应收'"`
	RefundMoney                          float64 `gorm:"type:decimal(12,2);default:0.00;comment:'部分退款'"`
	RefundConsumptionTax                 float64 `gorm:"type:decimal(12,2);default:0.00;comment:'退款消费税'"`
	RefundNum                            int     `gorm:"default:0;comment:'商品退款数量'"`
	SupplierMoney                        float64 `gorm:"type:decimal(12,2);default:0.00;comment:'供应商金额'"`
	SysMoney                             float64 `gorm:"type:decimal(12,2);default:0.00;comment:'平台结算金额'"`
	IsComment                            uint    `gorm:"not null;default:0;comment:'是否已评价(0否 1是)'"`
	OrderID                              uint    `gorm:"not null;default:0;comment:'订单id'"`
	SubOrderID                           int     `gorm:"default:0;comment:'子订单ID(拆单功能)'"`
	BatchNo                              string  `gorm:"type:varchar(100);default:'';comment:'扫码下单批次号'"`
	BatchTime                            int     `gorm:"default:1;comment:'扫码下单时间'"`
	IsReject                             int     `gorm:"default:0;comment:'是否拒单商品 0-否 1-是'"`
	SchemeID                             int     `gorm:"default:0;comment:'订单预设方案ID'"`
	MergeFromTableID                     int     `gorm:"default:0;comment:'并台来源桌台ID'"`
	Remark                               string  `gorm:"type:varchar(255);not null;default:'';comment:'自定义商品备注'"`
	UserID                               uint    `gorm:"not null;default:0;comment:'用户id'"`
	BagPrice                             float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'包装费'"`
	DiscountMoney                        float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'优惠金额'"`
	ExtraTimes                           int     `gorm:"not null;default:0;comment:'加餐数量'"`
	KitchenIsOpen                        int     `gorm:"default:1;comment:'厨显端是否开启，1为开启，0为关闭'"`
	AppID                                uint    `gorm:"not null;default:0;comment:'小程序id'"`
	CreateTime                           int     `gorm:"not null;default:0;comment:'创建时间'"`
	DeleteTime                           int     `gorm:"not null;default:0;comment:'删除时间'"`
}

type OrderPayType struct {
	ID             uint64  `gorm:"primaryKey;autoIncrement;not null;comment:'自增ID'"`
	SubID          int     `gorm:"default:0;comment:'子ID'"`
	OrderID        int     `gorm:"default:0;comment:'订单ID'"`
	PayStatus      int     `gorm:"default:1;comment:'支付状态 0-未支付 1-已支付'"`
	PaymentOrderID int     `gorm:"default:0;comment:'支付订单id'"`
	Value          int     `gorm:"default:0;comment:'支付方式'"`
	Price          float64 `gorm:"type:decimal(12,2);default:0.00;comment:'支付金额'"`
	Fee            int     `gorm:"default:0;comment:'支付费率0-100'"`
	FeeMoney       float64 `gorm:"type:decimal(12,2);default:0.00;comment:'单支付手续费'"`
	DisabledCancel int     `gorm:"default:0;comment:'禁止撤销 0-否 1-是'"`
	PayHash        string  `gorm:"type:varchar(100);default:'';comment:'支付唯一hash'"`
	ShopSupplierID int     `gorm:"default:0;comment:'店铺id'"`
	AppID          int     `gorm:"default:0;comment:'应用id'"`
	CreateTime     int64   `gorm:"autoCreateTime;not null;default:0;comment:'创建时间'"`
	UpdateTime     int64   `gorm:"autoUpdateTime;not null;default:0;comment:'更新时间'"`
}

type OrderOperationLog struct {
	ID             uint64 `gorm:"primaryKey;autoIncrement;not null;comment:'自增ID'"`
	OrderID        int    `gorm:"default:0;comment:'订单ID'"`
	SubOrderID     int    `gorm:"default:0;comment:'子订单ID'"`
	Source         string `gorm:"type:varchar(150);not null;default:'';comment:'来源 cashier-收银 assistant-助手 shop-商家后台'"`
	ShopUserID     int    `gorm:"default:0;comment:'操作用户id'"`
	Action         string `gorm:"type:varchar(150);not null;default:'';comment:'行为'"`
	Data           string `gorm:"type:text;default:'';comment:'数据'"`
	Remark         string `gorm:"type:varchar(255);not null;default:'';comment:'备注'"`
	ShopSupplierID int    `gorm:"default:0;comment:'门店id'"`
	AppID          int    `gorm:"default:0;comment:'应用id'"`
	CreateTime     int64  `gorm:"autoCreateTime;not null;default:0;comment:'创建时间'"`
	UpdateTime     int64  `gorm:"autoUpdateTime;not null;default:0;comment:'更新时间'"`
}
