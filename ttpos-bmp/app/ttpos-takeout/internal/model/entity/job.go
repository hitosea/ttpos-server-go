// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Job is the golang structure for table job.
type Job struct {
	Id                   int64   `json:"id"                   orm:"id"                     description:""`                                                                                                       //
	Uuid                 string  `json:"uuid"                 orm:"uuid"                   description:"外送订单UUID"`                                                                                               // 外送订单UUID
	ShopRefNo            string  `json:"shopRefNo"            orm:"shop_ref_no"            description:"餐馆订单参考，如UUID"`                                                                                           // 餐馆订单参考，如UUID
	CustomerMobile       string  `json:"customerMobile"       orm:"customer_mobile"        description:"下单客户电话"`                                                                                                 // 下单客户电话
	CustomerEmail        string  `json:"customerEmail"        orm:"customer_email"         description:"下单客户联系邮件"`                                                                                               // 下单客户联系邮件
	ProviderName         string  `json:"providerName"         orm:"provider_name"          description:"外送供应商： skootar,grab"`                                                                                    // 外送供应商： skootar,grab
	TakeoutRefNo         string  `json:"takeoutRefNo"         orm:"takeout_ref_no"         description:"外送系统订单号"`                                                                                                // 外送系统订单号
	ShopLocationUuid     string  `json:"shopLocationUuid"     orm:"shop_location_uuid"     description:"餐馆位置信息"`                                                                                                 // 餐馆位置信息
	ConsumerLocationUuid string  `json:"consumerLocationUuid" orm:"consumer_location_uuid" description:"消费者位置信息"`                                                                                                // 消费者位置信息
	JobDate              string  `json:"jobDate"              orm:"job_date"               description:"订单日期:'YYYY-MM-DD'"`                                                                                      // 订单日期:'YYYY-MM-DD'
	StartTime            string  `json:"startTime"            orm:"start_time"             description:"Start time. Format in 24 hr (00:00 to 23:59) or \"now\" for immediate job"`                              // Start time. Format in 24 hr (00:00 to 23:59) or "now" for immediate job
	FinishTime           string  `json:"finishTime"           orm:"finish_time"            description:"订单结束时间"`                                                                                                 // 订单结束时间
	PaymentType          string  `json:"paymentType"          orm:"payment_type"           description:"支付类型 Payment solution is 3 choice is \"\"invoice\"\", \"\"cash\"\", \"\"creditcard\"\",\"\"prepaid\"\""` // 支付类型 Payment solution is 3 choice is ""invoice"", ""cash"", ""creditcard"",""prepaid""
	JobStatus            string  `json:"jobStatus"            orm:"job_status"             description:"外送订单状态"`                                                                                                 // 外送订单状态
	Remark               string  `json:"remark"               orm:"remark"                 description:"订单备注"`                                                                                                   // 订单备注
	Reserved1            string  `json:"reserved1"            orm:"reserved1"              description:"保留字段1"`                                                                                                  // 保留字段1
	Reserved2            string  `json:"reserved2"            orm:"reserved2"              description:"保留字段2"`                                                                                                  // 保留字段2
	CallbackUrl          string  `json:"callbackUrl"          orm:"callback_url"           description:"订单状态更新回调"`                                                                                               // 订单状态更新回调
	SkootarId            string  `json:"skootarId"            orm:"skootar_id"             description:"骑手Id"`                                                                                                   // 骑手Id
	SkootarName          string  `json:"skootarName"          orm:"skootar_name"           description:"骑手名称"`                                                                                                   // 骑手名称
	SkootarPhone         string  `json:"skootarPhone"         orm:"skootar_phone"          description:"骑手电话"`                                                                                                   // 骑手电话
	SkootarImageUrl      string  `json:"skootarImageUrl"      orm:"skootar_image_url"      description:"骑手头像"`                                                                                                   // 骑手头像
	SkootarRating        float64 `json:"skootarRating"        orm:"skootar_rating"         description:"骑手评分"`                                                                                                   // 骑手评分
	CreatedAt            int     `json:"createdAt"            orm:"created_at"             description:"创建时间"`                                                                                                   // 创建时间
	UpdatedAt            int     `json:"updatedAt"            orm:"updated_at"             description:"更新时间"`                                                                                                   // 更新时间
	DeletedAt            int     `json:"deletedAt"            orm:"deleted_at"             description:"软删除"`                                                                                                    // 软删除
}
