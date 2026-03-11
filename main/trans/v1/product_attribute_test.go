//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestProductAttribute(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	productAttributeService := ProductAttributeService{db: db}
	productAttributeList, err := productAttributeService.GetProductAttributeList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(productAttributeList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertProductAttribute(t *testing.T) {
	testConvertProductAttribute()
}

func testConvertProductAttribute() {
	InitializeSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	productAttributeService := ProductAttributeService{db: db, targetDB: targetDB}
	err = productAttributeService.ConvertProductAttribute()
	if err != nil {
		panic(err)
	}
	fmt.Println("product_attribute转换完成")
}
