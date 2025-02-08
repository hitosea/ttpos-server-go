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
	testConvertAttribute()
	testConvertCategory()
	testConvertFreeTag()
	testConvertProductAttributeGroup()
	testConvertProductAttribute()
	testConvertProductPrintLabel()
	testConvertProductUnit()
	testConvertReturnReason()
	testConvertSpec()
	testConvertProduct()
	testConvertTable()
	testConvertTableType()
	testConvertTableArea()
}
