//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/utils"
)

func TestShopUser(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
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

	utils.InitSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}

	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	shopUserService := ShopUserService{db: db, targetDB: targetDB, targetCompanyUuid: 12122233333}
	err = shopUserService.ConvertShopUser()
	if err != nil {
		panic(err)
	}
	fmt.Println("shop_user转换完成")
}
