package req

import (
	"fmt"
	"sort"
	"strings"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/pkg/utils"
)

var DeskReqMessage = map[string]string{
	"uuid.required":            "桌台uuid不能为空",
	"sale_bill_uuid.required":  "销售账单UUID不能为空",
	"sale_order_uuid.required": "销售订单UUID不能为空",
}

// DeskListReq 桌台列表查询
type DeskListReq struct {
	Status      int `form:"status,default=-1"`    // 桌台状态, -1=全都、 0=未开台、1=已开台, 2=已开台不等于待清台
	IsBuffet    int `form:"is_buffet,default=-1"` // 是否是自助餐: -1=全都、0=否、1=是
	dto.PageReq     // 分页参数
}

// DeskInfoReq 桌台信息
type DeskInfoReq struct {
	Uuid uint64 `form:"uuid" binding:"required"` // 桌台uuid
}

type DeskJsonUuidReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 桌台uuid
}

// DeskCloseReq 关闭桌台
type DeskCloseReq struct {
	Uuid     uint64 `form:"uuid" binding:"required"` // 桌台uuid
	Reason   string `form:"reason"`                  // 关闭原因
	Password string `form:"password"`                // 密码 后台开启的时候才传
}

// DeskOrderCloseReq 关闭桌台订单
type DeskOrderCloseReq struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid"  binding:"required"`  // 销售账单UUID
	SaleOrderUuid uint64 `json:"sale_order_uuid"  binding:"required"` // 销售订单UUID
	Reason        string `form:"reason"`                              // 关闭原因
	Password      string `form:"password"`                            // 密码 后台开启的时候才传
}

// DeskBuffetCustomerType 自助餐顾客类型
type DeskBuffetCustomerType struct {
	Uuid    uint64 `json:"uuid"`     // 自助餐顾客类型uuid
	MealNum uint   `json:"meal_num"` // 就餐人数
}

// DeskOrderCreateReq 桌台订单创建
type DeskOrderCreateReq struct {
	DeskUuid            uint64                   `json:"desk_uuid"`             // 桌台uuid, 必填
	MealNum             uint                     `json:"meal_num"`              // 就餐人数: 非自助餐时, 最小为0, 最大为999, 自助餐时为0
	BuffetUuids         []uint64                 `json:"buffet_uuids"`          // 自助餐uuid列表: 非自助餐时, 传空数组; 自助餐时, 元素数量最小为1, 最大为2
	BuffetCustomerTypes []DeskBuffetCustomerType `json:"buffet_customer_types"` // 自助餐顾客类型列表: 非自助餐时, 传空数组; 自助餐时, 元素数量最小为1
	Remark              string                   `json:"remark"`                // 备注: 最小空字符串,最大50字符
	OrderSourceUuid     uint64                   `json:"order_source_uuid"`     // 订单来源UUID（0=店内，>0=外卖），可选
	NationalityUuid     uint64                   `json:"nationality_uuid"`      // 国籍UUID（0=未记录），可选
}

type GetMemberOrderCheckoutInfoReq struct {
	MemberSaleOrderUuid uint64 `json:"member_sale_order_uuid" binding:"required"` // 会员端销售订单UUID
	AddressChanged      bool   // 地址是否发生变化
}

// 创建会员端订单请求参数
type CreateMemberOrderReq struct {
	MemberSaleOrderUuid uint64               `json:"member_sale_order_uuid"` // 会员端销售订单UUID. 值为0时，表示创建会员端订单。值不为0时，表示更新会员端订单的商品列表。
	Products            []OrderProductAddReq `json:"products"`               // 商品列表
}

func (req *CreateMemberOrderReq) Validate() error {
	if len(req.Products) == 0 {
		return errors.New("未选购商品，请先选购商品")
	}
	for _, product := range req.Products {
		if product.Num <= 0 {
			return errors.New("商品数量不能为0")
		}
		if product.FlavorUuid <= 0 {
			return errors.New("商品规格ID不能为0")
		}
	}
	return nil
}

type GetMemberOrderFormInfoReq struct {
	MemberSaleOrderUuid uint64 `form:"member_sale_order_uuid" binding:"required"` // 会员端销售订单UUID
}

// GetDineInOrderFormInfoReq 获取堂食订单提交表单信息请求参数
type GetDineInOrderFormInfoReq struct {
	SaleBillUuid  uint64 `form:"sale_bill_uuid" binding:"required"`  // 销售账单UUID
	SaleOrderUuid uint64 `form:"sale_order_uuid" binding:"required"` // 销售订单UUID
}

// PayDineInOrderReq 堂食订单提交支付请求参数
type PayDineInOrderReq struct {
	SaleBillUuid      uint64 `json:"sale_bill_uuid" binding:"required"`      // 销售账单UUID
	SaleOrderUuid     uint64 `json:"sale_order_uuid" binding:"required"`     // 销售订单UUID
	PaymentMethodUuid uint64 `json:"payment_method_uuid" binding:"required"` // 支付方式UUID
	Remark            string `json:"remark"`                                 // 订单备注
}

// SetDineInOrderDiningMethodReq 设置堂食订单用餐方式请求参数
type SetDineInOrderDiningMethodReq struct {
	SaleBillUuid uint64 `json:"sale_bill_uuid" binding:"required"` // 销售账单UUID
	DiningMethod uint   `json:"dining_method"`                     // 用餐方式 0:堂食 1:打包
}

// GetDineInOrderPayInfoReq 获取堂食订单支付信息请求参数
type GetDineInOrderPayInfoReq struct {
	SaleBillUuid      uint64 `form:"sale_bill_uuid" binding:"required"`  // 销售账单UUID
	SaleOrderUuid     uint64 `form:"sale_order_uuid" binding:"required"` // 销售订单UUID
	PaymentMethodUuid uint64 `form:"payment_method_uuid"`                // 支付方式UUID
}

// GetDineInOrderPayStatusReq 获取堂食订单支付状态请求参数
type GetDineInOrderPayStatusReq struct {
	SaleBillUuid  uint64 `form:"sale_bill_uuid" binding:"required"`  // 销售账单UUID
	SaleOrderUuid uint64 `form:"sale_order_uuid" binding:"required"` // 销售订单UUID
}

// CheckDineInOrderReq 会员端堂食订单结算前检查请求参数
type CheckDineInOrderReq struct {
	SaleBillUuid  uint64 `form:"sale_bill_uuid" binding:"required"`  // 销售账单UUID
	SaleOrderUuid uint64 `form:"sale_order_uuid" binding:"required"` // 销售订单UUID
}

// 创建会员端堂食订单请求参数
type CreateMemberDineInOrderReq struct {
	SaleBillUuid  uint64               `json:"sale_bill_uuid"`  // 销售账单UUID. 值为0时，表示创建新订单。值不为0时，表示向已有订单添加商品。
	SaleOrderUuid uint64               `json:"sale_order_uuid"` // 销售订单UUID. 更新订单时必填。
	Products      []OrderProductAddReq `json:"products"`        // 商品列表
}

func (req *CreateMemberDineInOrderReq) Validate() error {
	if len(req.Products) == 0 {
		return errors.New("未选购商品，请先选购商品")
	}
	for _, product := range req.Products {
		if product.Num <= 0 {
			return errors.New("商品数量不能为0")
		}
		if product.FlavorUuid <= 0 {
			return errors.New("商品规格ID不能为0")
		}
	}
	return nil
}

type OrderProductAddReq struct {
	FlavorUuid        uint64   `json:"flavor_uuid"`    // 某个规格商品ID
	SauceUuidList     []uint64 `json:"sauce_uuid"`     // 小料ID列表
	AttributeUuidList []uint64 `json:"attribute_uuid"` // 属性ID列表
	Num               float64  `json:"num"`            // 商品数量。
	Price             float64  `json:"price"`          // 商品价格单价。
}

// 商品key。格式：规格id-属性id,属性id-加料id,加料id
func (req *OrderProductAddReq) ProductKey(attributeUuidList []uint64) string {
	// 先排序
	sort.Slice(attributeUuidList, func(i, j int) bool {
		return attributeUuidList[i] < attributeUuidList[j]
	})
	sort.Slice(req.SauceUuidList, func(i, j int) bool {
		return req.SauceUuidList[i] < req.SauceUuidList[j]
	})

	// 转为字符串
	AttributeUuidListStr := make([]string, 0)
	for _, attributeUuid := range attributeUuidList {
		AttributeUuidListStr = append(AttributeUuidListStr, fmt.Sprintf("%d", attributeUuid))
	}
	SauceUuidListStr := make([]string, 0)
	for _, sauceUuid := range req.SauceUuidList {
		SauceUuidListStr = append(SauceUuidListStr, fmt.Sprintf("%d", sauceUuid))
	}
	// 按照“规格id-属性id,属性id-加料id,加料id”的格式拼接
	return fmt.Sprintf("%d-%s-%s", req.FlavorUuid, strings.Join(AttributeUuidListStr, ","), strings.Join(SauceUuidListStr, ","))
}

// 判断是不是自助餐
func (req *DeskOrderCreateReq) IsBuffet() bool {
	// 请求有自助餐套餐id的话，就肯定是自助餐
	return len(req.BuffetUuids) > 0
}

// 获取就餐人数。
// 当是非自助餐桌台请求时，就餐人数=MealNum
// 当是自助餐桌台请求时，就餐人数=每种顾客数量的累计。如老人2人、小孩3人，那么就餐人数为5人
func (req *DeskOrderCreateReq) GetMealNum() uint {
	// 当是自助餐桌台请求时，就餐人数=每种顾客数量的累计。如老人2人、小孩3人，那么就餐人数为5人
	if req.IsBuffet() {
		mealNum := uint(0)
		for _, customer := range req.BuffetCustomerTypes {
			mealNum = mealNum + customer.MealNum
		}
		return mealNum
	}
	// 当是非自助餐桌台请求时，就餐人数=MealNum
	return req.MealNum
}

func (req *DeskOrderCreateReq) ValidateCreateDeskOrderReq() error {
	if req.DeskUuid == 0 {
		return errors.New("桌台uuid不能为0")
	}
	// 非自助餐桌台的就餐人数不能为0
	if !req.IsBuffet() && req.MealNum == 0 {
		return errors.New("就餐人数不能为0")
	}
	if req.IsBuffet() {
		if req.MealNum < 1 || req.MealNum > 999 {
			return errors.New("就餐人数不能小于1或大于999")
		}
	}
	if req.IsBuffet() {
		if len(req.BuffetUuids) < 1 || len(req.BuffetUuids) > 2 {
			return errors.New("选择自助餐套餐数量不能小于1或大于2")
		}
		if len(req.BuffetCustomerTypes) == 0 {
			return errors.New("自助餐顾客类型未选择")
		}
		if req.GetMealNum() < 1 {
			return errors.WithMessage(errors.New("自助餐就餐人数不能为零"), utils.ToJson(req))
		}
	}
	return nil
}

// BindDeskReq 平板绑定桌台请求参数
type BindDeskReq struct {
	DeskUuid uint64 `json:"desk_uuid" binding:"required"` // 桌台Uuid
}

// EditSettingReq 修改设置请求参数
type EditSettingReq struct {
	DeskUuid uint64 `json:"desk_uuid" binding:"required"` // 桌台Uuid
	Remark   string `json:"remark"`                       // 机器备注
}

// ChangeDeskReq 切换桌台请求参数
type ChangeDeskReq struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid" binding:"required"`  // 销售账单UUID
	SaleOrderUuid uint64 `json:"sale_order_uuid" binding:"required"` // 销售订单UUID
	DeskUuid      uint64 `json:"desk_uuid" binding:"required"`       // 新桌台UUID
}

type MergeDeskReq struct {
	SaleBillUuid uint64   `json:"sale_bill_uuid" binding:"required"` // 销售账单UUID
	DeskUuids    []uint64 `json:"desk_uuids"`                        // 桌台UUID列表
}

func (req *MergeDeskReq) Validate() error {
	if len(req.DeskUuids) < 1 {
		return errors.New("请选择桌台")
	}
	return nil
}
