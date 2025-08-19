// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// PaymentApp is the golang structure for table payment_app.
type PaymentApp struct {
	Id                   uint   `json:"id"                   orm:"id"                      description:"自增ID"`          // 自增ID
	CompanyUuid          int64  `json:"companyUuid"          orm:"company_uuid"            description:"集团ID"`          // 集团ID
	LlWhiteIp            string `json:"llWhiteIp"            orm:"ll_white_ip"             description:"白名单IP"`         // 白名单IP
	LlMerchantId         string `json:"llMerchantId"         orm:"ll_merchant_id"          description:"商户号"`           // 商户号
	LlStoreId            string `json:"llStoreId"            orm:"ll_store_id"             description:"站点ID"`          // 站点ID
	LlPublicKey          string `json:"llPublicKey"          orm:"ll_public_key"           description:"LianLianpay公钥"` // LianLianpay公钥
	LlMerchantPrivateKey string `json:"llMerchantPrivateKey" orm:"ll_merchant_private_key" description:"商户私钥"`          // 商户私钥
	LlToken              string `json:"llToken"              orm:"ll_token"                description:"Token"`         // Token
	LlSignSalt           string `json:"llSignSalt"           orm:"ll_sign_salt"            description:"签名盐"`           // 签名盐
	CreateTime           int    `json:"createTime"           orm:"create_time"             description:"创建时间（时间戳）"`     // 创建时间（时间戳）
	UpdateTime           int    `json:"updateTime"           orm:"update_time"             description:"更新时间（时间戳）"`     // 更新时间（时间戳）
	DeleteTime           int    `json:"deleteTime"           orm:"delete_time"             description:"删除时间（时间戳）"`     // 删除时间（时间戳）
}
