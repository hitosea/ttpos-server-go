//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestFreeTag(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	freeTagService := FreeTagService{db: db}
	freeTagList, err := freeTagService.GetFreeTagList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(freeTagList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertFreeTag(t *testing.T) {
	testConvertFreeTag()
}

func testConvertFreeTag() {
	InitializeSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	freeTagService := FreeTagService{db: db, targetDB: targetDB}
	err = freeTagService.ConvertFreeTag()
	if err != nil {
		panic(err)
	}
}
