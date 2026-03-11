//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestBuffetCustomer(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	buffetCustomerService := BuffetCustomerRepository{db: db}
	buffetCustomerList, err := buffetCustomerService.GetBuffetCustomerList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(buffetCustomerList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertBuffetCustomer(t *testing.T) {
	testConvertBuffetCustomer()
}

func testConvertBuffetCustomer() {
	InitializeSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	buffetCustomerService := BuffetCustomerRepository{db: db, targetDB: targetDB}
	err = buffetCustomerService.ConvertBuffetCustomer()
	if err != nil {
		panic(err)
	}
}
