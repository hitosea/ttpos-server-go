package old_model

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
)

var conf = config.DatabaseConf{
	Host:          "localhost",
	Port:          3306,
	User:          "root",
	Password:      "5cd6a0408e9ccf92",
	RootPassword:  "5cd6a0408e9ccf92",
	TablePrefix:   "jjjfood_",
	SlowQueryTime: 0,
}
var dbName = "shop_wang"
var targetConf = config.DatabaseConf{
	Host:          "localhost",
	Port:          3306,
	User:          "root",
	Password:      "5cd6a0408e9ccf92",
	RootPassword:  "5cd6a0408e9ccf92",
	TablePrefix:   "ttpos_",
	SlowQueryTime: 0,
}
var targetDBName = "shop_1"

func TestProductUnit(t *testing.T) {
	db, err := NewMySQLConnection(conf, dbName)
	if err != nil {
		panic(err)
	}
	productUnitService := ProductUnitService{db: db}
	productUnitList, err := productUnitService.GetProductUnitList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(productUnitList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertProductUnit(t *testing.T) {
	testConvertProductUnit()
}

func testConvertProductUnit() {
	InitializeSonyFlakeId()

	database.InitSonyFlakeId()

	db, err := NewMySQLConnection(conf, dbName)
	if err != nil {
		panic(err)
	}

	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	productUnitService := ProductUnitService{db: db, targetDB: targetDB}
	err = productUnitService.ConvertProductUnit()
	if err != nil {
		panic(err)
	}
	fmt.Println("product_unit转换完成")
}
