//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestCall(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	callService := CallService{db: db}
	callList, err := callService.GetCallList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(callList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertCall(t *testing.T) {
	testConvertCall()
}

func testConvertCall() {
	InitializeSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	callService := CallService{db: db, targetDB: targetDB}
	err = callService.ConvertCall()
	if err != nil {
		panic(err)
	}
}
