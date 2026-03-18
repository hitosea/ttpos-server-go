//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestPayType(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	payTypeService := PayTypeService{db: db}
	payTypeList, err := payTypeService.GetPayTypeList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(payTypeList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertPayType(t *testing.T) {
	testConvertPayType()
}

func testConvertPayType() {
	InitializeSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	payTypeService := PayTypeService{db: db, targetDB: targetDB}
	err = payTypeService.ConvertPayType()
	if err != nil {
		panic(err)
	}
}
