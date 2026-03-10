//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/utils"
)

func TestSetting(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	settingService := SettingService{db: db}
	settingList, err := settingService.GetSettingList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(settingList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertSetting(t *testing.T) {
	testConvertSetting()
}

func testConvertSetting() {
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
	settingService := SettingService{db: db, targetDB: targetDB}
	err = settingService.ConvertSetting()
	if err != nil {
		panic(err)
	}
	fmt.Println("setting转换完成")
}
