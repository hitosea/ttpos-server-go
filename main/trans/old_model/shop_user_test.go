package old_model

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/database"
)

func TestShopUser(t *testing.T) {
	db, err := NewMySQLConnection(conf, dbName)
	if err != nil {
		panic(err)
	}
	shopUserService := ShopUserService{db: db}
	shopUserList, err := shopUserService.GetShopUserList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(shopUserList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertShopUser(t *testing.T) {
	testConvertShopUser()
}

func testConvertShopUser() {
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
	shopUserService := ShopUserService{db: db, targetDB: targetDB}
	err = shopUserService.ConvertShopUser()
	if err != nil {
		panic(err)
	}
	fmt.Println("shop_user转换完成")
}
