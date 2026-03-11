//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestProductAttributeGroup(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	productAttributeGroupService := ProductAttributeGroupService{db: db}
	productAttributeGroupList, err := productAttributeGroupService.GetProductAttributeGroupList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(productAttributeGroupList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertProductAttributeGroup(t *testing.T) {
	testConvertProductAttributeGroup()
}

func testConvertProductAttributeGroup() {
	InitializeSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	productAttributeGroupService := ProductAttributeGroupService{db: db, targetDB: targetDB}
	err = productAttributeGroupService.ConvertProductAttributeGroup()
	if err != nil {
		panic(err)
	}
	fmt.Println("product_attribute_group转换完成")
}
