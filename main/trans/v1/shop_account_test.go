//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/utils"
)

func TestShopAccount(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	shopAccountService := ShopAccountService{db: db}
	shopAccountList, err := shopAccountService.GetShopAccountList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(shopAccountList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertShopAccount(t *testing.T) {
	testConvertShopAccount()
}

func testConvertShopAccount() {
	InitializeSonyFlakeId()

	utils.InitSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}

	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	shopAccountService := ShopAccountService{db: db, targetDB: targetDB}
	err = shopAccountService.ConvertShopAccount()
	if err != nil {
		panic(err)
	}
	fmt.Println("shop_account转换完成")
}
