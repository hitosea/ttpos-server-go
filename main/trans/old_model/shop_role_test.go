package old_model

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/database"
)

func TestShopRole(t *testing.T) {
	db, err := NewMySQLConnection(conf, dbName)
	if err != nil {
		panic(err)
	}
	shopRoleService := ShopRoleService{db: db}
	shopRoleList, err := shopRoleService.GetShopRoleList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(shopRoleList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertShopRole(t *testing.T) {
	testConvertShopRole()
}

func testConvertShopRole() {
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
	shopRoleService := ShopRoleService{db: db, targetDB: targetDB}
	err = shopRoleService.ConvertShopRole()
	if err != nil {
		panic(err)
	}
	fmt.Println("shop_role转换完成")
}
