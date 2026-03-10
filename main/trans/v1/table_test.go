//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/utils"
)

func TestTable(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	tableService := TableService{db: db}
	tableList, err := tableService.GetTableList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(tableList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertTable(t *testing.T) {
	testConvertTable()
}

func testConvertTable() {
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
	tableService := TableService{db: db, targetDB: targetDB}
	err = tableService.ConvertTable()
	if err != nil {
		panic(err)
	}
	fmt.Println("table转换完成")
}
