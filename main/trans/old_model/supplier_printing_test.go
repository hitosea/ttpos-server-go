package old_model

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/database"
)

func TestSupplierPrinting(t *testing.T) {
	db, err := database.NewMySQLConnection(conf, dbName)
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

	db, err := NewMySQLConnection(conf, dbName)
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
