//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestCustomerType(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	customerTypeService := CustomerTypeService{db: db}
	customerTypeList, err := customerTypeService.GetCustomerTypeList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(customerTypeList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertCustomerType(t *testing.T) {
	testConvertCustomerType()
}

func testConvertCustomerType() {
	InitializeSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	customerTypeService := CustomerTypeService{db: db, targetDB: targetDB}
	err = customerTypeService.ConvertCustomerType()
	if err != nil {
		panic(err)
	}
}
