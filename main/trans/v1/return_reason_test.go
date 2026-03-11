//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestReturnReason(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
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
	testConvertReturnReason()
}

func testConvertReturnReason() {
	InitializeSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	returnReasonService := ReturnReasonService{db: db, targetDB: targetDB}
	err = returnReasonService.ConvertReturnReason()
	if err != nil {
		panic(err)
	}
	fmt.Println("return_reason转换完成")
}
