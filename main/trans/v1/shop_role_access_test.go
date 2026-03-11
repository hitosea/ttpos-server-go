//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/utils"
)

func TestShopRoleAccess(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	shopRoleAccessService := ShopRoleAccessService{db: db}
	shopRoleAccessList, err := shopRoleAccessService.GetShopRoleAccessList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(shopRoleAccessList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertShopRoleAccess(t *testing.T) {
	testConvertShopRoleAccess()
}

func testConvertShopRoleAccess() {
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
	shopRoleAccessService := ShopRoleAccessService{db: db, targetDB: targetDB}
	err = shopRoleAccessService.ConvertShopRoleAccess()
	if err != nil {
		panic(err)
	}
	fmt.Println("shop_role_access转换完成")
}
