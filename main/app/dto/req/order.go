package req

import (
	"errors"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"

	"github.com/shopspring/decimal"
)

var OrderReqMessage = map[string]string{
	"uuid.required":               "桌台uuid不能为空",
	"sale_bill_uuid.required":     "销售账单UUID不能为空",
	"sale_order_uuid.required":    "销售订单UUID不能为空",
	"order_product_uuid.required": "订单商品UUID不能为空",
	"price.required":              "商品价格不能为空",
	"population.required":         "人数不能为空",
	"remark.required":             "备注不能为空",
}

// OrderListReq 订单列表查询
type OrderListReq struct {
	dto.PageReq             // 分页参数
	OrderNo          string `form:"order_no"`                 // 订单编号
	DateType         int    `form:"date_type,default=-1"`     // 日期类型 -1=全都、 0=今天、 1=昨天、 2=本周
	EnableCreateTime bool   `form:"enable_create_time"`       // 启用开台时间 false-不启用，true-启用
	EnablePayTime    bool   `form:"enable_pay_time"`          // 启用支付时间 false-不启用，true-启用
	QueryStartTime   uint   `form:"query_start_time"`         // 查询开始时间戳
	QueryEndTime     uint   `form:"query_end_time"`           // 查询结束时间戳
	Status           int    `form:"status,default=-1"`        // 账单状态, -1=全都、 0=待付款、1=已完成、2=已取消
	BillType         int    `form:"bill_type,default=-1"`     // 账单类型, -1=全都、 0=Desk桌台订单、1=OrderingFood点餐订单
	DiningMethod     int    `form:"dining_method,default=-1"` // 用餐方式, -1=全都、 0-堂食 1-打包
}

// OrderInfoReq 订单信息查询
type OrderInfoReq struct {
	SaleBillUuid  uint64 `form:"sale_bill_uuid" json:"sale_bill_uuid"`   // 销售账单UUID
	SaleOrderUuid uint64 `form:"sale_order_uuid" json:"sale_order_uuid"` // 销售订单UUID 当查看子订单信息的时候才需要传
}

// OrderCancelReq 订单取消
type OrderCancelReq struct {
	SaleBillUuid    uint64 `json:"sale_bill_uuid"`    // 销售账单UUID
	CancelReason    string `json:"cancel_reason"`     // 取消原因
	Password        string `form:"password"`          // 高级密码 后台开启的时候才传
	NotNeedPassword bool   `json:"not_need_password"` // 不需要验证高级密码。仅用后端使用
}

// OrderReturnInfoReq 退款订单信息
type OrderReturnInfoReq struct {
	SaleBillUuid  uint64 `form:"sale_bill_uuid"`  // 销售账单UUID
	SaleOrderUuid uint64 `form:"sale_order_uuid"` // 销售订单UUID。退款都是针对子单进行退款
}

// OrderReturnReq 订单退款
type OrderReturnReq struct {
	SaleBillUuid        uint64               `json:"sale_bill_uuid"`         // 销售账单UUID
	SaleOrderUuid       uint64               `json:"sale_order_uuid"`        // 销售订单UUID。退款都是针对子单进行退款
	MemberSaleOrderUuid uint64               `json:"member_sale_order_uuid"` // 会员销售订单UUID。退款都是针对子单进行退款
	Products            []OrderReturnProduct `json:"products"`               // 退款商品UUID列表. 如果为空，则退款所有商品,即整单退款
	// 退款账户信息
	BankCode    string `json:"bank_code"`    // 银行编码 - 当存在QR PromptPay的时候需要传
	AccountNo   string `json:"account_no"`   // 账号 - 当存在QR PromptPay的时候需要传
	AccountName string `json:"account_name"` // 账户名称 - 当存在QR PromptPay的时候需要传
	// 手动退款积分
	Points float64 `json:"points"` // 积分。当手动退积分时，需要传积分数量。
}

// OrderReReturnReq 订单重新退款
type OrderReReturnReq struct {
	ReturnOrderUuid  uint64 `json:"return_order_uuid"`  // 退款订单UUID
	ReturnAmountUuid uint64 `json:"return_amount_uuid"` // 退款金额UUID
	// 退款账户信息
	BankCode    string `json:"bank_code"`    // 银行编码 - 当存在QR PromptPay的时候需要传
	AccountNo   string `json:"account_no"`   // 账号 - 当存在QR PromptPay的时候需要传
	AccountName string `json:"account_name"` // 账户名称 - 当存在QR PromptPay的时候需要传
}

type OrderReturnProduct struct {
	SaleOrderProductUuid uint64  `json:"sale_order_product_uuid"` // 销售订单商品UUID
	Num                  float64 `json:"num"`                     // 退款数量
}

// OrderReverseSettleInfoReq 获取反结账弹窗信息
type OrderReverseSettleInfoReq struct {
	SaleBillUuid uint64 `form:"sale_bill_uuid"` // 销售账单UUID
}

// OrderReverseSettleReq 处理反结账
type OrderReverseSettleReq struct {
	SaleBillUuid uint64 `json:"sale_bill_uuid"` // 销售账单UUID
	DeskUuid     uint64 `json:"desk_uuid"`      // 桌台UUID. 仅桌台订单才填该字段
	HideOrder    bool   `json:"hide_order"`     // 是否挂单. 仅点餐订单才填该字段
}

// OrderDeleteReq 订单删除
type OrderDeleteReq struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单UUID
	SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单UUID 传0的时候默认删除主单以及所有子单，不然只删除子单
}

// GetOrderBuffetReq 获取订单自助餐信息请求
type GetOrderBuffetReq struct {
	SaleBillUuid  uint64 `form:"sale_bill_uuid" json:"sale_bill_uuid"`    // 销售账单UUID
	SaleOrderUuid uint64 `form:"sale_order_uuid"  json:"sale_order_uuid"` // 销售订单UUID
}

// OrderShowReq 订单显示 取单
type OrderShowReq struct {
	SaleBillUuid uint64 `json:"sale_bill_uuid"` // 销售账单UUID
}

// OrderIsCellCloseReq 是否可关闭订单
type OrderIsCellCloseReq struct {
	DeskUuid     uint64 `form:"desk_uuid"`      // 桌台UUID	   二选一, 销售账单UUID权重最大
	SaleBillUuid uint64 `form:"sale_bill_uuid"` // 销售账单UUID	二选一，销售账单UUID权重最大
}

// OrderProductDeleteReq 删除订单商品
type OrderProductDeleteReq struct {
	SaleBillUuid     uint64 `json:"sale_bill_uuid" binding:"required"`     // 销售账单UUID
	SaleOrderUuid    uint64 `json:"sale_order_uuid" binding:"required"`    // 销售订单UUID
	OrderProductUuid uint64 `json:"order_product_uuid" binding:"required"` // 订单商品UUID
}

// OrderProductChangePriceReq 订单商品改价
type OrderProductChangePriceReq struct {
	SaleBillUuid     uint64  `json:"sale_bill_uuid" binding:"required"`     // 销售账单UUID
	SaleOrderUuid    uint64  `json:"sale_order_uuid" binding:"required"`    // 销售订单UUID
	OrderProductUuid uint64  `json:"order_product_uuid" binding:"required"` // 订单商品UUID
	Price            float64 `json:"price"`                                 // 改价
}

// OrderDiscountMethodReq 订单打折方式
type OrderDiscountMethodReq struct {
	DiscountMethod int     `json:"discount_method"` // 打折方式 1=改价 2=打折, 3=抹零, 4=免单
	SaleBillUuid   uint64  `json:"sale_bill_uuid"`  // 销售账单UUID
	SaleOrderUuid  uint64  `json:"sale_order_uuid"` // 销售订单UUID
	Price          float64 `json:"price"`           // 改价
	Discount       float64 `json:"discount"`        // 打折 - 0-100之间
	DiscountType   int     `json:"discount_type"`   // 打折 - 打折类型 0=百分比折扣，如八折为80% 1=百分比减免Off，如八折为20% off
	ZeroRule       int     `json:"zero_rule"`       // 抹零规则
}

// OrderAmountChangeReq 订单改价
type OrderAmountChangeReq struct {
	SaleBillUuid  uint64  `json:"sale_bill_uuid" binding:"required"`  // 销售账单UUID
	SaleOrderUuid uint64  `json:"sale_order_uuid" binding:"required"` // 销售订单UUID
	Price         float64 `json:"price"`                              // 改价
}

// Validate 验证参数
func (req OrderAmountChangeReq) Validate() error {
	if req.Price < 0 || req.Price > 1000000000000000 {
		return errors.New("请输入0-1000000000000000间的价格")
	}
	if req.SaleBillUuid == 0 || req.SaleOrderUuid == 0 {
		return errors.New("销售账单UUID或销售订单UUID不能为空")
	}
	return nil
}

// OrderDiscountReq 订单打折
type OrderDiscountReq struct {
	SaleBillUuid  uint64  `json:"sale_bill_uuid"`  // 销售账单UUID
	SaleOrderUuid uint64  `json:"sale_order_uuid"` // 销售订单UUID
	Discount      float64 `json:"discount"`        // 打折。0-100之间
	DiscountType  int     `json:"discount_type"`   // 打折类型 0=百分比折扣，如八折为80% 1=百分比减免Off，如八折为20% off
}

// Validate 验证参数
func (req OrderDiscountReq) Validate() error {
	if req.Discount < 0 || req.Discount > 100 {
		return errors.New("折扣错误")
	}
	if req.SaleBillUuid == 0 || req.SaleOrderUuid == 0 {
		return errors.New("销售账单UUID或销售订单UUID不能为空")
	}
	if req.DiscountType != constant.DiscountTypePercent && req.DiscountType != constant.DiscountTypeOff {
		return errors.New("打折类型错误")
	}
	return nil
}

// GetDiscount 获取折扣
func (req OrderDiscountReq) GetDiscount() float64 {
	if req.DiscountType == constant.DiscountTypePercent {
		// 前端传的值范围是0-100，所以需要转换为0-1
		return decimal.NewFromFloat(req.Discount).Div(decimal.NewFromInt(100)).InexactFloat64()
	}
	if req.DiscountType == constant.DiscountTypeOff {
		// 前端传的值范围是0-100，所以需要转换为0-1
		discountOff := decimal.NewFromFloat(req.Discount).Div(decimal.NewFromInt(100))
		// % off 百分比减免。需要转换为百分比打折。 1 - discountOff
		discount := decimal.NewFromFloat(1).Sub(discountOff).InexactFloat64()
		return discount
	}
	// 异常的值，拒绝打折，返回折扣值1，表示不打折
	return 1
}

// GetOffDiscount 获取百分比减免的折扣率。使用场景：1.记录订单操作日志时，“优惠折扣：折扣-80%（￥50）”
// 示例1: 80% => -20
// 示例1: 30% off => -30
func (req OrderDiscountReq) GetOffDiscount() float64 {
	discount := req.GetDiscount()
	// (1-discount) * 100
	return decimal.NewFromFloat(1).Sub(decimal.NewFromFloat(discount)).Mul(decimal.NewFromFloat(100)).InexactFloat64()
}

// GetPercentDiscount 获取百分比折扣的折扣率。使用场景：1.记录订单操作日志时，“优惠折扣：折扣-80%（￥50）”
// 示例1: 80% => 80
// 示例1: 30% off => 70
func (req OrderDiscountReq) GetPercentDiscount() float64 {
	discount := req.GetDiscount()
	// discount * 100
	return decimal.NewFromFloat(discount).Mul(decimal.NewFromFloat(100)).InexactFloat64()
}

// OrderZeroRuleReq 订单抹零规则
type OrderZeroRuleReq struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单UUID
	SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单UUID
	ZeroRule      int    `json:"zero_rule"`       // 抹零规则
}

// Validate 验证参数
func (req OrderZeroRuleReq) Validate() error {
	if req.SaleBillUuid == 0 {
		return errors.New("销售账单UUID不能为空")
	}
	if req.SaleOrderUuid == 0 {
		return errors.New("销售订单UUID不能为空")
	}
	if req.ZeroRule != constant.DiscountZeroRuleNone &&
		req.ZeroRule != constant.DiscountZeroRulePercent &&
		req.ZeroRule != constant.DiscountZeroRuleFixed &&
		req.ZeroRule != constant.DiscountZeroRuleRound &&
		req.ZeroRule != constant.DiscountZeroRuleInteger {
		return errors.New("抹零规则错误")
	}
	return nil
}

// OrderDiscountCancelReq 取消点餐订单所有优惠折扣，包括改价、打折、抹零
type OrderDiscountCancelReq struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单UUID
	SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单UUID
}

// OrderChangePopulationReq 订单人数
type OrderChangePopulationReq struct {
	SaleBillUuid uint64 `json:"sale_bill_uuid" binding:"required"` // 销售账单UUID
	Population   int    `json:"population" binding:"required"`     // 人数
}

// OrderProductRemarkReq 订单商品remark
type OrderProductRemarkReq struct {
	SaleBillUuid     uint64 `json:"sale_bill_uuid" binding:"required"`     // 销售账单UUID
	SaleOrderUuid    uint64 `json:"sale_order_uuid" binding:"required"`    // 销售订单UUID
	OrderProductUuid uint64 `json:"order_product_uuid" binding:"required"` // 订单商品UUID
	Remark           string `json:"remark"`                                // remark
}

// OrderRemarkReq 订单备注
type OrderRemarkReq struct {
	SaleBillUuid uint64   `json:"sale_bill_uuid"` // 销售账单UUID
	RemarkUuids  []uint64 `json:"remark_uuids"`   // 选择的整单备注
	Remark       string   `json:"remark"`         // 自定义的整单备注文本
}

// OrderChangeBuffetReq 订单调整自助餐
type OrderChangeBuffetReq struct {
	SaleBillUuid        uint64                   `json:"sale_bill_uuid"`        // 销售账单UUID
	BuffetUuids         []uint64                 `json:"buffet_uuids"`          // 自助餐uuid列表: 非自助餐时, 传空数组; 自助餐时, 元素数量最小为1, 最大为2
	BuffetCustomerTypes []DeskBuffetCustomerType `json:"buffet_customer_types"` // 自助餐顾客类型列表
}

func (req OrderChangeBuffetReq) Validate() error {
	if req.SaleBillUuid == 0 {
		return errors.New("销售账单UUID不能为空")
	}
	if len(req.BuffetUuids) > 2 || len(req.BuffetUuids) <= 0 {
		return errors.New("自助餐uuid列表错误")
	}
	if len(req.BuffetCustomerTypes) <= 0 {
		return errors.New("自助餐顾客类型列表错误")
	}
	for _, customerType := range req.BuffetCustomerTypes {
		if customerType.Uuid == 0 {
			return errors.New("自助餐顾客类型uuid不能为空")
		}
	}
	return nil
}

// OrderChangeBuffetClockReq 自助餐加钟
type OrderChangeBuffetClockReq struct {
	SaleBillUuid  uint64   `json:"sale_bill_uuid"`  // 销售账单UUID
	SaleOrderUuid uint64   `json:"sale_order_uuid"` // 销售订单UUID
	DelayUuids    []uint64 `json:"delay_uuids"`     // 加钟uuid列表
}

func (req OrderChangeBuffetClockReq) Validate() error {
	if req.SaleBillUuid == 0 {
		return errors.New("销售账单UUID不能为空")
	}
	if len(req.DelayUuids) <= 0 {
		return errors.New("加钟uuid列表不能为空")
	}
	for _, delayUuid := range req.DelayUuids {
		if delayUuid == 0 {
			return errors.New("加钟uuid不能为空")
		}
	}
	return nil
}

// OrderChangeBuffetProductListReq 获取桌台的自助餐商品列表
type OrderChangeBuffetProductListReq struct {
	SaleBillUuid uint64 `form:"sale_bill_uuid"` // 销售账单UUID
}

// GetProductListReq 获取桌台的(未)送厨商品列表
type GetProductListReq struct {
	SaleBillUuid uint64 `form:"sale_bill_uuid" json:"sale_bill_uuid" binding:"required"` // 销售账单UUID
}

// MemberBalanceChangeReq 会员余额变动请求
type MemberBalanceChangeReq struct {
	MemberUuid  uint64  `json:"uuid"`
	OrderNo     string  `json:"order_no"`     // 订单编号
	Money       float64 `json:"money"`        // 变动的金额。 正数为增加，负数为减少
	RelatedUuid uint64  `json:"related_uuid"` // 关联的ID。比如退款的时候，关联的是退款单金额的ID; 用餐订单反结账的时候，关联的是用餐订单的ID
}

// CashBoxBalanceChangeReq 钱箱余额变动请求
type CashBoxBalanceChangeReq struct {
	Amount      float64 `json:"amount"`       // 变动的金额。 正数为增加，负数为减少
	RelatedUuid uint64  `json:"related_uuid"` // 关联的ID。比如退款的时候，关联的是退款单金额的ID; 用餐订单反结账的时候，关联的是用餐订单的ID
	OrderNo     string  `json:"order_no"`     // 订单编号
}
