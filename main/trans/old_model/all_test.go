package old_model

import (
	"testing"
	"ttpos-server-go/config"
)

var conf = config.DatabaseConf{
	Host:          "localhost",
	Port:          3306,
	User:          "root",
	Password:      "5cd6a0408e9ccf92",
	TablePrefix:   "jjjfood_",
	SlowQueryTime: 0,
}
var dbName = "shop_wang"
var targetConf = config.DatabaseConf{
	Host:          "localhost",
	Port:          3306,
	User:          "root",
	Password:      "5cd6a0408e9ccf92",
	TablePrefix:   "ttpos_",
	SlowQueryTime: 0,
}
var targetDBName = "shop_1"

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
}
