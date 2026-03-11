//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestGetAppList(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	appService := AppService{db: db}
	apps, err := appService.GetAppList()
	json, _ := json.Marshal(apps)
	fmt.Println(string(json))
}

func TestConvertApp(t *testing.T) {
	testConvertApp()
}

func testConvertApp() {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	appService := AppService{db: db, targetDB: targetDB}
	err = appService.ConvertApp()
	if err != nil {
		panic(err)
	}
}
