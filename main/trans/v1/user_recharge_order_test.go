//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/utils"
)

func TestUserRechargeOrder(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	userRechargeOrderService := UserRechargeOrderService{db: db}
	userList, err := userRechargeOrderService.GetUserRechargeOrderList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(userList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertUserRechargeOrder(t *testing.T) {
	testConvertUserRechargeOrder()
}

func testConvertUserRechargeOrder() {
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
	userRechargeOrderService := UserRechargeOrderService{db: db, targetDB: targetDB}
	err = userRechargeOrderService.ConvertUserRechargeOrder()
	if err != nil {
		panic(err)
	}
	fmt.Println("user_recharge_order转换完成")
}
