//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/database"
)

func TestSpec(t *testing.T) {
	db, err := database.NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	specService := SpecService{db: db}
	specList, err := specService.GetSpecList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(specList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertSpec(t *testing.T) {
	testConvertSpec()
}

func testConvertSpec() {
	InitializeSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	specService := SpecService{db: db, targetDB: targetDB}
	err = specService.ConvertSpec()
	if err != nil {
		panic(err)
	}
	fmt.Println("spec转换完成")
}
