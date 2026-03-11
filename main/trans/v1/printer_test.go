//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestPrinter(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	printerService := PrinterService{db: db}
	printerList, err := printerService.GetPrinterList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(printerList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertPrinter(t *testing.T) {
	testConvertPrinter()
}

func testConvertPrinter() {
	InitializeSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	printerService := PrinterService{db: db, targetDB: targetDB}
	err = printerService.ConvertPrinter()
	if err != nil {
		panic(err)
	}
}
