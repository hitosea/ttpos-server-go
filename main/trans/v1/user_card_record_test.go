//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/utils"
)

func TestUserCardRecord(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	userCardRecordService := UserCardRecordService{db: db}
	userCardRecordList, err := userCardRecordService.GetUserCardRecordList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(userCardRecordList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertUserCardRecord(t *testing.T) {
	testConvertUserCardRecord()
}

func testConvertUserCardRecord() {
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
	userCardRecordService := UserCardRecordService{db: db, targetDB: targetDB}
	err = userCardRecordService.ConvertUserCardRecord()
	if err != nil {
		panic(err)
	}
	fmt.Println("user_card_record转换完成")
}
