package old_model

import (
	"testing"
	"ttpos-server-go/config"
)

var conf = config.DatabaseConf{
	Host:          "localhost",
	Port:          13061,
	User:          "root",
	Password:      "fbe61a042f752dff",
	TablePrefix:   "jjjfood_",
	SlowQueryTime: 0,
}
var dbName = "shop1724054105"
var targetConf = config.DatabaseConf{
	Host:          "localhost",
	Port:          3306,
	User:          "root",
	Password:      "d4f3c7d055516a3b",
	TablePrefix:   "ttpos_",
	SlowQueryTime: 0,
}
var targetDBName = "shop8609817471094784"

func TestConvertAll(t *testing.T) {
	testConvertAttribute()             // 商品属性
	testConvertCategory()              // 商品分类
	testConvertFreeTag()               // 商品标签
	testConvertProductAttributeGroup() // 商品属性组
	testConvertProductAttribute()      // 商品包属性
	testConvertProductPrintLabel()     // 商品打印标签
	testConvertProductUnit()           // 商品单位
	testConvertReturnReason()          // 退菜原因
	testConvertSpec()                  // 规格
	testConvertProduct()               // 商品
	testConvertTable()                 // 桌台
	testConvertTableType()             // 桌台类型
	testConvertTableArea()             // 桌台区域
	testConvertShopAccess()            // 权限
	testConvertShopRole()              // 角色
	testConvertShopUser()              // 用户
	testConvertShopUserRole()          // 用户角色
	testConvertShopRoleAccess()        // 角色权限
	testConvertCall()                  // 客户呼叫记录
	testConvertSupplierPrinting()      // 商品打印（档口）
	testConvertPrinter()               // 打印机
	testConvertCustomerType()          // 自助餐顾客类型
	testConvertBuffetCustomer()        // 自助餐顾客类型价格
	testConvertBuffetDelay()           // 自助餐加钟
	testConvertUserCard()              // 会员卡类型
	testConvertUser()                  // 会员
	testConvertUserGrade()             // 会员等级
	testConvertUserCardRecord()        // 会员卡领取记录
	testConvertUserPointsLog()         // 会员积分变动记录
	testConvertUserBalanceLog()        // 会员余额变动记录

}
