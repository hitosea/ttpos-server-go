//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/utils"
)

func TestTableType(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	tableTypeService := TableTypeService{db: db}
	tableTypeList, err := tableTypeService.GetTableTypeList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(tableTypeList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertTableType(t *testing.T) {
	testConvertTableType()
}

func testConvertTableType() {
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
	tableTypeService := TableTypeService{db: db, targetDB: targetDB}
	err = tableTypeService.ConvertTableType()
	if err != nil {
		panic(err)
	}
	fmt.Println("table_type转换完成")
}
