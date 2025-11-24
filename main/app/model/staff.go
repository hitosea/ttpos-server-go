package model

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/pkg/utils"
)

// Staff 员工表 ttpos_staff
type Staff struct {
	BaseModel
	CompanyUuid         uint64 `gorm:"column:company_uuid;type:bigint(20) unsigned;default:0;comment:集团ID;NOT NULL" json:"company_uuid"`
	Username            string `gorm:"column:username;type:varchar(255);comment:用户名;NOT NULL" json:"username"`
	Password            string `gorm:"column:password;type:varchar(255);comment:登录密码;NOT NULL" json:"password"`
	PermissionPassword  string `gorm:"column:permission_password;type:varchar(255);comment:权限密码（加密存储）" json:"-"`
	Phone               string `gorm:"column:phone;type:varchar(20);comment:手机号" json:"phone"`
	PasswordChangeCount int    `gorm:"column:password_change_count;type:int(11);default:0;comment:修改密码次数" json:"password_change_count"`
	PasswordChangeTime  int64  `gorm:"column:password_change_time;type:int(10) unsigned;default:0;comment:修改密码时间;NOT NULL" json:"password_change_time"`
	RealName            string `gorm:"column:real_name;type:varchar(255);comment:姓名;NOT NULL" json:"real_name"`
	IsSuper             int    `gorm:"column:is_super;type:tinyint(3);default:0;comment:是否为超级管理员0不是,1是;NOT NULL" json:"is_super"`
	HasDataPermission   int    `gorm:"column:has_data_permission;type:tinyint(3);default:0;comment:是否有数据管理权限0否1是;NOT NULL" json:"has_data_permission"`
	UserType            int    `gorm:"column:user_type;type:tinyint(1);default:0;comment:账号类型0总台1门店;NOT NULL" json:"user_type"`
	IsDisable           int    `gorm:"column:is_disable;type:tinyint(3);default:0;comment:是否禁用1禁用,0未禁用;NOT NULL" json:"is_disable"`
	BindKey             string `gorm:"column:bind_key;type:varchar(255);comment:绑定的设备key" json:"bind_key"`
	CashierOnline       int    `gorm:"column:cashier_online;type:tinyint(1);default:0;comment:收银员当班 0-不在线 1-在线;NOT NULL" json:"cashier_online"`
	CashierLoginTime    int64  `gorm:"column:cashier_login_time;type:int(11) unsigned;default:0;comment:收银员当班登录时间;NOT NULL" json:"cashier_login_time"`
	DutyNo              string `gorm:"column:duty_no;type:varchar(64);comment:当班编号" json:"duty_no"`

	Company *Company `gorm:"foreignKey:CompanyUuid;references:Uuid"`
	Device  *Device  `gorm:"foreignKey:BindKey;references:DeviceId"`

	// 多对多关联：员工拥有多个角色
	Roles []*Role `gorm:"many2many:staff_role;foreignKey:Uuid;joinForeignKey:StaffUuid;references:Uuid;joinReferences:RoleUuid"`
}

// GetUserName 获取用户名
func (model *Staff) GetUserName() string {
	return utils.IfString(model.RealName != "", model.RealName, model.Username)
}

// StaffRole 员工角色关系表 ttpos_staff_role
type StaffRole struct {
	BaseModel
	StaffUuid int64 `gorm:"column:staff_uuid;type:bigint(20);default:0;comment:员工UUID;NOT NULL" json:"staff_uuid"`
	RoleUuid  int64 `gorm:"column:role_uuid;type:bigint(20);default:0;comment:角色UUID;NOT NULL" json:"role_uuid"`
}

// StaffShiftLog 员工交班记录表 ttpos_staff_shift_log
type StaffShiftLog struct {
	BaseModel
	StaffUuid         uint64  `gorm:"column:staff_uuid;type:bigint(20) unsigned;default:0;comment:员工ID;NOT NULL" json:"staff_uuid"`
	ShiftNo           string  `gorm:"column:shift_no;type:varchar(64);comment:交班编号;NOT NULL" json:"shift_no"`
	Status            int     `gorm:"column:status;type:int(11);default:0;comment:状态： 0未交班,1已交班;NOT NULL" json:"status"`
	PreviousShiftCash float64 `gorm:"column:previous_shift_cash;type:decimal(12,2);default:0.00;comment:上一班遗留备用金;NOT NULL" json:"previous_shift_cash"`
	CurrentCashTotal  float64 `gorm:"column:current_cash_total;type:decimal(12,2);default:0.00;comment:当前钱箱现金总计;NOT NULL" json:"current_cash_total"`
	Incomes           string  `gorm:"column:incomes;type:varchar(255);comment:收入详情" json:"incomes"`
	TotalIncome       float64 `gorm:"column:total_income;type:decimal(12,2);default:0.00;comment:总收入;NOT NULL" json:"total_income"`
	CashTakenOut      float64 `gorm:"column:cash_taken_out;type:decimal(12,2);default:0.00;comment:本班取出现金;NOT NULL" json:"cash_taken_out"`
	CashLeft          float64 `gorm:"column:cash_left;type:decimal(12,2);default:0.00;comment:本班遗留备用金;NOT NULL" json:"cash_left"`
	CashIncome        float64 `gorm:"column:cash_income;type:decimal(12,2);default:0.00;comment:本班收入现金;NOT NULL" json:"cash_income"`
	TotalBusiness     float64 `gorm:"column:total_business;type:decimal(12,2);default:0.00;comment:本班营业总额（不包含退款）;NOT NULL" json:"total_business"`
	IsPrinted         int     `gorm:"column:is_printed;type:tinyint(1);default:0;comment:是否打印 0-未打印 1-已打印;NOT NULL" json:"is_printed"`
	Remark            string  `gorm:"column:remark;type:varchar(255);comment:备注" json:"remark"`
	WithdrawCash      float64 `gorm:"column:withdraw_cash;type:decimal(12,2);default:0.00;comment:中途取出现金;NOT NULL" json:"withdraw_cash"`
	DepositCash       float64 `gorm:"column:deposit_cash;type:decimal(12,2);default:0.00;comment:中途存入现金;NOT NULL" json:"deposit_cash"`
	ExceptionRemark   string  `gorm:"column:exception_remark;type:varchar(255);comment:异常报备;NOT NULL" json:"exception_remark"`
	Abnormal          string  `gorm:"column:abnormal;type:varchar(255);comment:异常信息-json字符串" json:"abnormal"`
	ShiftStartTime    int64   `gorm:"column:shift_start_time;type:int(10);default:0;comment:当班开始时间;NOT NULL" json:"shift_start_time"`
	ShiftEndTime      int64   `gorm:"column:shift_end_time;type:int(10);default:0;comment:当班结束时间;NOT NULL" json:"shift_end_time"`

	ErpnextOpenPosEntryName  string `gorm:"column:erpnext_open_pos_entry_name;type:varchar(255);NOT NULL;default:'';comment:erpnext开账名称" json:"erpnext_open_pos_entry_name"`
	ErpnextClosePosEntryName string `gorm:"column:erpnext_close_pos_entry_name;type:varchar(255);NOT NULL;default:'';comment:erpnext结账名称" json:"erpnext_close_pos_entry_name"`
	ErpnextAsyncRecordId     string `gorm:"column:erpnext_async_record_id;type:varchar(255);NOT NULL;default:'';comment:erpnext异步记录ID" json:"erpnext_async_record_id"`

	Staff *Staff `gorm:"foreignKey:StaffUuid;references:Uuid"`
}

// StaffShiftSnapshot 员工交班快照表
type StaffShiftSnapshot struct {
	BaseModel
	ShiftLogUuid uint64 `gorm:"column:shift_log_uuid;type:bigint(20) unsigned;default:0;comment:交班记录ID;NOT NULL" json:"shift_log_uuid"`
	Content      string `gorm:"column:content;type:text;comment:快照json" json:"content"`
}

// StaffLoginLog 管理员登录记录表 `ttpos_staff_login_log`
type StaffLoginLog struct {
	BaseModel
	StaffUuid uint64 `gorm:"column:staff_uuid;type:bigint(20);default:0;comment:员工UUID;NOT NULL" json:"staff_uuid"`
	Username  string `gorm:"column:username;type:varchar(50);comment:用户名;NOT NULL" json:"username"`
	Ip        string `gorm:"column:ip;type:varchar(128);comment:登录ip;NOT NULL" json:"ip"`
	Result    string `gorm:"column:result;type:varchar(128);comment:登录结果;NOT NULL" json:"result"`
}

// StaffOperationLog 员工操作日志表 ttpos_staff_operation_log
type StaffOperationLog struct {
	BaseModel
	StaffUuid    uint64 `gorm:"column:staff_uuid;type:bigint(20) unsigned;default:0;comment:员工ID;NOT NULL" json:"staff_uuid"`
	Title        string `gorm:"column:title;type:varchar(255);comment:标题;NOT NULL" json:"title"`
	Url          string `gorm:"column:url;type:varchar(255);comment:操作URL;NOT NULL" json:"url"`
	RequestData  string `gorm:"column:request_data;type:varchar(255);comment:请求数据;NOT NULL" json:"request_data"`
	ResponseData string `gorm:"column:response_data;type:varchar(255);comment:响应数据;NOT NULL" json:"response_data"`
	Type         string `gorm:"column:type;type:varchar(255);comment:操作类型;NOT NULL" json:"type"`
	Ip           string `gorm:"column:ip;type:varchar(255);comment:操作IP;NOT NULL" json:"ip"`
	Source       string `gorm:"column:source;type:varchar(255);comment:操作来源;NOT NULL" json:"source"`
	Agent        string `gorm:"column:agent;type:varchar(255);comment:操作用户代理;NOT NULL" json:"agent"`
}

// IsHandedOver 是否已交班
func (model *StaffShiftLog) IsHandedOver() bool {
	return model.Status == constant.StaffHandedOver
}

// IsReported 是否已经报备
func (model *StaffShiftLog) IsReported() bool {
	return model.ExceptionRemark != ""
}

// StaffShiftSnapshotContent 交班快照内容
type StaffShiftSnapshotContent struct {
	ID                uint64                        `json:"id"`
	ShiftUserID       uint64                        `json:"shift_user_id"`
	ShiftNo           string                        `json:"shift_no"`
	Status            int                           `json:"status"`
	PreviousShiftCash string                        `json:"previous_shift_cash"`
	CurrentCashTotal  string                        `json:"current_cash_total"`
	Incomes           []StaffShiftSnapshotIncome    `json:"incomes"`
	TotalIncome       string                        `json:"total_income"`
	CashTakenOut      string                        `json:"cash_taken_out"`
	CashLeft          string                        `json:"cash_left"`
	CashIncome        string                        `json:"cash_income"`
	TotalBusiness     string                        `json:"total_business"`
	IsPrinted         int                           `json:"is_printed"`
	Remark            string                        `json:"remark"`
	WithdrawCash      string                        `json:"withdraw_cash"`
	DepositCash       string                        `json:"deposit_cash"`
	ExceptionRemark   string                        `json:"exception_remark"`
	AppID             uint64                        `json:"app_id"`
	ShopSupplierID    uint64                        `json:"shop_supplier_id"`
	ShiftStartTime    string                        `json:"shift_start_time"`
	ShiftEndTime      string                        `json:"shift_end_time"`
	CreateTime        string                        `json:"create_time"`
	UpdateTime        string                        `json:"update_time"`
	Abnormal          StaffShiftSnapshotAbnormal    `json:"abnormal"`
	Order             StaffShiftSnapshotOrder       `json:"order"`
	SalesInfo         []StaffShiftSnapshotSalesInfo `json:"salesInfo"`
	User              StaffShiftSnapshotUser        `json:"user"`
}

// StaffShiftSnapshotAbnormal 交班快照-异常统计
type StaffShiftSnapshotAbnormal struct {
	RefundProductTimes         int `json:"refund_product_times"`
	CancelRefundTimes          int `json:"cancel_refund_times"`
	ProductFreeTimes           int `json:"product_free_times"`
	CancelProductFreeTimes     int `json:"cancel_product_free_times"`
	RefundTime                 int `json:"refund_time"`
	ProductMoveTimes           int `json:"product_move_times"`
	ChangePriceTimes           int `json:"change_price_times"`
	ChangeOrderPriceTimes      int `json:"change_order_price_times"`
	DiscountOrderTimes         int `json:"discount_order_times"`
	RoundOrderTimes            int `json:"round_order_times"`
	FreeOrderTimes             int `json:"free_order_times"`
	ReverseSettleTimes         int `json:"reverse_settle_times"`
	RoundOrderCancelTimes      int `json:"round_order_cancel_times"`
	CheckoutRoundOrderTimes    int `json:"checkout_round_order_times"`
	RechargeRefundTimes        int `json:"recharge_refund_times"`
	RechargeReverseSettleTimes int `json:"recharge_reverse_settle_times"`
}

// StaffShiftSnapshotOrder 交班快照-订单统计
type StaffShiftSnapshotOrder struct {
	ReceivablePrice         float64                        `json:"receivable_price"`
	NotTaxTotalProductPrice float64                        `json:"not_tax_total_product_price"`
	TotalProductPrice       float64                        `json:"total_product_price"`
	TotalGiveProductPrice   float64                        `json:"total_give_product_price"`
	ServiceMoney            float64                        `json:"service_money"`
	DiscountMoney           float64                        `json:"discount_money"`
	ConsumptionTaxMoney     float64                        `json:"consumption_tax_money"`
	PayFeeMoney             float64                        `json:"pay_fee_money"`
	ReceivedBalancePrice    float64                        `json:"received_balance_price"`
	ReceivedPrice           float64                        `json:"received_price"`
	SalesPrice              float64                        `json:"sales_price"`
	ProductNum              float64                        `json:"product_num"`
	UserDiscountMoney       float64                        `json:"user_discount_money"`
	RefundMoney             float64                        `json:"refund_money"`
	RefundConsumptionTax    float64                        `json:"refund_consumption_tax"`
	FreeProductPrice        float64                        `json:"free_product_price"`
	FreeProductNum          float64                        `json:"free_product_num"`
	FreeOrderPrice          float64                        `json:"free_order_price"`
	FreeOrderNum            uint                           `json:"free_order_num"`
	TotalOrderNum           int                            `json:"total_order_num"`
	TotalTableNum           int                            `json:"total_table_num"`
	TotalPeopleNum          int                            `json:"total_people_num"`
	TotalCancelOrderNum     int                            `json:"total_cancel_order_num"`
	TotalCancelOrderAmount  float64                        `json:"total_cancel_order_amount"`
	MinOrderPrice           float64                        `json:"min_order_price"`
	MaxOrderPrice           float64                        `json:"max_order_price"`
	AvgOrderPrice           float64                        `json:"avg_order_price"`
	TableOrderNum           int                            `json:"table_order_num"`
	TablePeopleNum          int                            `json:"table_people_num"`
	TableMinOrderPrice      float64                        `json:"table_min_order_price"`
	TableMaxOrderPrice      float64                        `json:"table_max_order_price"`
	TableAvgOrderPrice      float64                        `json:"table_avg_order_price"`
	TablePeopleAvg          float64                        `json:"table_people_avg"`
	CashierOrderNum         int                            `json:"cashier_order_num"`
	CashierMinOrderPrice    float64                        `json:"cashier_min_order_price"`
	CashierMaxOrderPrice    float64                        `json:"cashier_max_order_price"`
	CashierAvgOrderPrice    float64                        `json:"cashier_avg_order_price"`
	TakeawayOrderNum        int                            `json:"takeaway_order_num"`
	TakeawayMinOrderPrice   float64                        `json:"takeaway_min_order_price"`
	TakeawayMaxOrderPrice   float64                        `json:"takeaway_max_order_price"`
	TakeawayAvgOrderPrice   float64                        `json:"takeaway_avg_order_price"`
	GiftPoints              float64                        `json:"gift_points"`
	GiftMoney               float64                        `json:"gift_money"`
	RechargeAmount          float64                        `json:"recharge_amount"`
	RechargeRefundTotal     float64                        `json:"recharge_refund_total"`
	NotSettledTotalOrderNum int                            `json:"not_settled_total_order_num"`
	NotSettledTotalPrice    float64                        `json:"not_settled_total_price"`
	DiscountRatio           string                         `json:"discount_ratio"`
	BusinessPrice           float64                        `json:"business_price"`
	UserCount               int                            `json:"user_count"`
	PeakHourList            []StaffShiftSnapshotPeakHour   `json:"peak_hour_list"`
	PercentageList          []StaffShiftSnapshotPercentage `json:"percentage_list"`
	Incomes                 []StaffShiftSnapshotIncome     `json:"incomes"`
}

// StaffShiftSnapshotOrder 交班快照-高峰期统计
type StaffShiftSnapshotPeakHour struct {
	TimePeriod string  `json:"time_period"`
	Date       string  `json:"date"`
	Hour       uint    `json:"hour"`
	Num        int     `json:"num"`
	Amount     float64 `json:"amount"`
}

// StaffShiftSnapshotPercentage 交班快照-税率统计
type StaffShiftSnapshotPercentage struct {
	TotalPrice     float64 `json:"total_price"`
	TaxRate        float64 `json:"tax_rate"`
	ConsumptionTax float64 `json:"consumption_tax"`
}

// StaffShiftSnapshotIncome 交班快照-支付统计
type StaffShiftSnapshotIncome struct {
	PayType             int     `json:"pay_type"`
	PayTypeName         string  `json:"pay_type_name"`
	PayTypeWay          string  `json:"pay_type_way"`
	Price               float64 `json:"price"`
	OrderNum            int     `json:"order_num"`
	RefundIncludedPrice float64 `json:"refund_included_price"`
}

// StaffShiftSnapshotSalesInfo 交班快照-销售统计
type StaffShiftSnapshotSalesInfo struct {
	Name     string `json:"name"`
	Sales    string `json:"sales"`
	Prices   string `json:"prices"`
	NameText string `json:"name_text"`
}

// StaffShiftSnapshotUser 交班快照-员工
type StaffShiftSnapshotUser struct {
	ShopUserID       uint64 `json:"shop_user_id"`
	UserName         string `json:"user_name"`
	Password         string `json:"password"`
	Phone            string `json:"phone"`
	PasswordChange   int    `json:"password_change"`
	RealName         string `json:"real_name"`
	IsSuper          int    `json:"is_super"`
	ShopSupplierID   uint64 `json:"shop_supplier_id"`
	IsDelete         int    `json:"is_deleted"`
	UserType         int    `json:"user_type"`
	IsStatus         int    `json:"is_status"`
	AppID            uint64 `json:"app_id"`
	BindKey          string `json:"bind_key"`
	CashierOnline    int    `json:"cashier_online"`
	CashierLoginTime int64  `json:"cashier_login_time"`
	DutyNo           string `json:"duty_no"`
	CreateTime       string `json:"create_time"`
	UpdateTime       string `json:"update_time"`
}
