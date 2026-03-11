//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/utils"
)

func TestUserPointsLog(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	userService := UserPointsLogService{db: db}
	userList, err := userService.GetUserPointsLogList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(userList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertUserPointsLog(t *testing.T) {
	testConvertUserPointsLog()
}

func testConvertUserPointsLog() {
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
	userService := UserPointsLogService{db: db, targetDB: targetDB}
	err = userService.ConvertUserPointsLog()
	if err != nil {
		panic(err)
	}
	fmt.Println("user_points_log转换完成")
}
