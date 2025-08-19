// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// PaymentApp is the golang structure of table ttpos_payment_app for DAO operations like Where/Data.
type PaymentApp struct {
	g.Meta               `orm:"table:ttpos_payment_app, do:true"`
	Id                   interface{} // 自增ID
	CompanyUuid          interface{} // 集团ID
	LlWhiteIp            interface{} // 白名单IP
	LlMerchantId         interface{} // 商户号
	LlStoreId            interface{} // 站点ID
	LlPublicKey          interface{} // LianLianpay公钥
	LlMerchantPrivateKey interface{} // 商户私钥
	LlToken              interface{} // Token
	LlSignSalt           interface{} // 签名盐
	CreateTime           interface{} // 创建时间（时间戳）
	UpdateTime           interface{} // 更新时间（时间戳）
	DeleteTime           interface{} // 删除时间（时间戳）
}
