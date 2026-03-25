//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestPrinterTemplate(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	printerTemplateService := PrinterTemplateService{db: db}
	printerTemplateList, err := printerTemplateService.GetPrinterTemplateList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(printerTemplateList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertPrinterTemplate(t *testing.T) {
	testConvertPrinterTemplate()
}

func testConvertPrinterTemplate() {
	InitializeSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	printerTemplateService := PrinterTemplateService{db: db, targetDB: targetDB}
	err = printerTemplateService.ConvertPrinterTemplate()
	if err != nil {
		panic(err)
	}
}
