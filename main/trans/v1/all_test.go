//go:build integration

package v1

import (
	"testing"
	"ttpos-server-go/config"
)

var sourceConf = config.DatabaseConf{
	Host:          "localhost",
	Port:          25443,
	User:          "root",
	Password:      "69c1e9542d2a7f19",
	TablePrefix:   "jjjfood_",
	SlowQueryTime: 0,
}
var sourceDBName = "shop1724054088"

var targetConf = config.DatabaseConf{
	Host:          "localhost",
	Port:          13306,
	User:          "root",
	Password:      "cfeb18fa768c2d5f",
	TablePrefix:   "ttpos_",
	SlowQueryTime: 0,
}
var targetDBName = "shop4477708931072000"

func TestConvertAll(t *testing.T) {
	testConvertAttribute()             // 商品属性
	testConvertCategory()              // 商品分类
	testConvertFreeTag()               // 商品标签
	testConvertProductAttributeGroup() // 商品属性组
	// testConvertProductAttribute()      // 商品包属性
	testConvertProductPrintLabel() // 商品打印标签
	testConvertProductUnit()       // 商品单位
	testConvertReturnReason()      // 退菜原因
	testConvertSpec()              // 规格
	// testConvertProduct()               // 商品
	testConvertTable()          // 桌台
	testConvertTableType()      // 桌台类型
	testConvertTableArea()      // 桌台区域
	testConvertShopAccess()     // 权限 // 主键冲突
	testConvertShopRole()       // 角色
	testConvertShopUser()       // 用户
	testConvertShopUserRole()   // 用户角色
	testConvertShopRoleAccess() // 角色权限
	// testConvertCall()                  // 客户呼叫记录 // Table 'shop7828781666304000.customer_call' doesn't exist [recove
	testConvertSupplierPrinting() // 商品打印（档口）
	testConvertPrinter()          // 打印机
	testConvertCustomerType()     // 自助餐顾客类型 // buffet_customer_type.go:52]: Error 1062 (23000): Duplicate entry '1'
	testConvertBuffetCustomer()   // 自助餐顾客类型价格
	testConvertBuffetDelay()      // 自助餐加钟
	testConvertUserCard()         // 会员卡类型
	testConvertUser()             // 会员
	testConvertUserGrade()        // 会员等级 // Error 1062 (23000): Duplicate entry '0' for key 'unique_uuid' [recovered]
	testConvertUserCardRecord()   // 会员卡领取记录
	testConvertUserPointsLog()    // 会员积分变动记录
	testConvertUserBalanceLog()   // 会员余额变动记录
	// testConvertProductSKU()     // 商品规格
}
