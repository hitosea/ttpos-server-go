//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/utils"
)

func TestGetProductUnitList(t *testing.T) {
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

func TestConvertProductPrintLabel(t *testing.T) {
	testConvertProductPrintLabel()
}

func testConvertProductPrintLabel() {
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
	productPrintLabelService := ProductPrintLabelService{db: db, targetDB: targetDB}
	err = productPrintLabelService.ConvertProductPrintLabel()
	if err != nil {
		panic(err)
	}
	fmt.Println("转换完成")
}
