package model

// PaymentApp 支付配置 `ttpos_payment_app`
type PaymentApp struct {
	ID                   uint   `gorm:"column:id;type:int(11) unsigned;primaryKey;autoIncrement;comment:'自增ID'"`
	CompanyUuid          uint64 `gorm:"column:company_uuid;type:bigint(20) unsigned;default:0;comment:集团ID;NOT NULL" json:"company_uuid"`
	LlWhiteIp            string `gorm:"column:ll_white_ip;type:varchar(255);default:'';comment:白名单IP;NOT NULL" json:"ll_white_ip"`
	LlMerchantId         string `gorm:"column:ll_merchant_id;type:varchar(255);default:'';comment:商户号;NOT NULL" json:"ll_merchant_id"`
	LlStoreId            string `gorm:"column:ll_store_id;type:varchar(100);default:'';comment:站点ID;NOT NULL" json:"ll_store_id"`
	LlPublicKey          string `gorm:"column:ll_public_key;type:text;default:'';comment:LianLianpay公钥;NOT NULL" json:"ll_public_key"`
	LlMerchantPrivateKey string `gorm:"column:ll_merchant_private_key;type:text;default:'';comment:商户私钥;NOT NULL" json:"ll_merchant_private_key"`
	LlToken              string `gorm:"column:ll_token;type:varchar(255);default:'';comment:Token;NOT NULL" json:"ll_token"`
	LlSignSalt           string `gorm:"column:ll_sign_salt;type:varchar(255);default:'';comment:签名盐;NOT NULL" json:"ll_sign_salt"`
	CreateTime           int64  `gorm:"autoCreateTime;column:create_time;type:int(10);comment:'创建时间(时间戳)'"`
	UpdateTime           int64  `gorm:"autoUpdateTime;column:update_time;type:int(10);comment:'更新时间(时间戳)'"`
	DeleteTime           int64  `gorm:"column:delete_time;type:int(10);default:0;comment:'删除时间(时间戳)'"`
}
