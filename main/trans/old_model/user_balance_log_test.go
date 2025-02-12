package old_model

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/database"
)

func TestUserBalanceLog(t *testing.T) {
	db, err := NewMySQLConnection(conf, dbName)
	if err != nil {
		panic(err)
	}
	userBalanceLogService := UserBalanceLogService{db: db}
	userBalanceLogList, err := userBalanceLogService.GetUserBalanceLogList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(userBalanceLogList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertUserBalanceLog(t *testing.T) {
	testConvertUserBalanceLog()
}

func testConvertUserBalanceLog() {
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
	userBalanceLogService := UserBalanceLogService{db: db, targetDB: targetDB}
	err = userBalanceLogService.ConvertUserBalanceLog()
	if err != nil {
		panic(err)
	}
	fmt.Println("user_balance_log转换完成")
}
