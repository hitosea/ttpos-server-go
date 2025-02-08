package old_model

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/database"
)

func TestTable(t *testing.T) {
	db, err := NewMySQLConnection(conf, dbName)
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

	database.InitSonyFlakeId()

	db, err := NewMySQLConnection(conf, dbName)
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
