//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestProductSKU(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	productSKUService := ProductSKUService{db: db}
	productSKUs, err := productSKUService.GetProductSKUService()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(productSKUs)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertProductSKU(t *testing.T) {
	testConvertProductSKU()
}

func testConvertProductSKU() {
	InitializeSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	productSKUService := ProductSKUService{db: db, targetDB: targetDB}
	err = productSKUService.ConvertProductSKU()
	if err != nil {
		panic(err)
	}
	fmt.Println("product_sku转换完成")
}
