//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/utils"
)

func TestShopUserRole(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	shopUserRoleService := ShopUserRoleService{db: db}
	shopUserRoleList, err := shopUserRoleService.GetShopUserRoleList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(shopUserRoleList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertShopUserRole(t *testing.T) {
	testConvertShopUserRole()
}

func testConvertShopUserRole() {
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
	shopUserRoleService := ShopUserRoleService{db: db, targetDB: targetDB}
	err = shopUserRoleService.ConvertShopUserRole()
	if err != nil {
		panic(err)
	}
	fmt.Println("shop_user_role转换完成")
}
