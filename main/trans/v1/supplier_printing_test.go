//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/database"
)

func TestSupplierPrinting(t *testing.T) {
	db, err := database.NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	supplierPrintingService := SupplierPrintingService{db: db}
	supplierPrintingList, err := supplierPrintingService.GetSupplierPrintingList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(supplierPrintingList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertSupplierPrinting(t *testing.T) {
	testConvertSupplierPrinting()
}

func testConvertSupplierPrinting() {
	InitializeSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	supplierPrintingService := SupplierPrintingService{db: db, targetDB: targetDB}
	err = supplierPrintingService.ConvertSupplierPrinting()
	if err != nil {
		panic(err)
	}
	fmt.Println("supplier_printing转换完成")
}
func TestParseRegionID(t *testing.T) {
	supplierPrintingService := SupplierPrintingService{}
	testCases := []string{
		`["2","6","4","0","8"]`,
		`["1","3","5","7","9"]`,
		`["10","20","30","40","50"]`,
		`["100","200","300","400","500"]`,
		`""`,
		``,
	}

	for _, str := range testCases {
		regionIDs, err := supplierPrintingService.parseIdList(str)
		if err != nil {
			panic(err)
		}
		fmt.Println(regionIDs)
	}
}
