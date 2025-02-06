package old_model

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
)

func TestReturnReason(t *testing.T) {
	db, err := NewMySQLConnection(config.DatabaseConf{
		Host:          "localhost",
		Port:          3306,
		User:          "root",
		Password:      "yourpassword",
		RootPassword:  "yourpassword",
		TablePrefix:   "jjjfood_",
		SlowQueryTime: 0,
	}, "shop_wang")
	if err != nil {
		panic(err)
	}
	returnReasonService := ReturnReasonService{db: db}
	returnReasonList, err := returnReasonService.GetReturnReasonList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(returnReasonList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertReturnReason(t *testing.T) {
	database.InitSonyFlakeId()

	db, err := NewMySQLConnection(config.DatabaseConf{
		Host:          "localhost",
		Port:          3306,
		User:          "root",
		Password:      "yourpassword",
		RootPassword:  "yourpassword",
		TablePrefix:   "jjjfood_",
		SlowQueryTime: 0,
	}, "shop_wang")
	if err != nil {
		panic(err)
	}
	targetDB, err := NewMySQLConnection(config.DatabaseConf{
		Host:          "localhost",
		Port:          3306,
		User:          "root",
		Password:      "yourpassword",
		RootPassword:  "yourpassword",
		TablePrefix:   "ttpos_",
		SlowQueryTime: 0,
	}, "shop_wang")
	if err != nil {
		panic(err)
	}
	returnReasonService := ReturnReasonService{db: db, targetDB: targetDB}
	err = returnReasonService.ConvertReturnReason()
	if err != nil {
		panic(err)
	}
}
