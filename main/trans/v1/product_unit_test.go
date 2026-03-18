//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/utils"
)

func TestProductUnit(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	productUnitService := ProductUnitService{db: db}
	productUnitList, err := productUnitService.GetProductUnitList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(productUnitList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertProductUnit(t *testing.T) {
	testConvertProductUnit()
}

func testConvertProductUnit() {
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
	productUnitService := ProductUnitService{db: db, targetDB: targetDB}
	err = productUnitService.ConvertProductUnit()
	if err != nil {
		panic(err)
	}
	fmt.Println("product_unit转换完成")
}
