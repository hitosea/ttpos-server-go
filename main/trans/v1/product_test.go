//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestGetProductList(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	productService := ProductService{db: db}
	products, err := productService.GetProductList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(products)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertProduct(t *testing.T) {
	testConvertProduct()
}

func testConvertProduct() {
	InitializeSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	productService := ProductService{db: db, targetDB: targetDB}
	err = productService.ConvertProduct()
	if err != nil {
		panic(err)
	}
}
