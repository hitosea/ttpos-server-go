//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestErpSupplier(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	erpSupplierService := ErpSupplierService{db: db}
	erpSupplierList, err := erpSupplierService.GetErpSupplierList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(erpSupplierList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertErpSupplier(t *testing.T) {
	testConvertErpSupplier()
}

func testConvertErpSupplier() {
	InitializeSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	erpSupplierService := ErpSupplierService{db: db, targetDB: targetDB}
	err = erpSupplierService.ConvertErpSupplier()
	if err != nil {
		panic(err)
	}
}
