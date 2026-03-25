//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/utils"
)

func TestShopAccess(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	shopAccessService := ShopAccessService{db: db}
	shopAccessList, err := shopAccessService.GetShopAccessList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(shopAccessList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertShopAccess(t *testing.T) {
	testConvertShopAccess()
}

func testConvertShopAccess() {
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
	shopAccessService := ShopAccessService{db: db, targetDB: targetDB}
	err = shopAccessService.ConvertShopAccess()
	if err != nil {
		panic(err)
	}
	fmt.Println("shop_access转换完成")
}
