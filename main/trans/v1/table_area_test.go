//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/utils"
)

func TestTableArea(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	tableAreaService := TableAreaService{db: db}
	tableAreaList, err := tableAreaService.GetTableAreaList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(tableAreaList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertTableArea(t *testing.T) {
	testConvertTableArea()
}

func testConvertTableArea() {
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
	tableAreaService := TableAreaService{db: db, targetDB: targetDB}
	err = tableAreaService.ConvertTableArea()
	if err != nil {
		panic(err)
	}
	fmt.Println("table_area转换完成")
}
