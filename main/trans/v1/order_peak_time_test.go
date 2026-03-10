//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestOrderPeakTime(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	orderPeakTimeService := OrderPeakTimeService{db: db}
	orderPeakTimeList, err := orderPeakTimeService.GetOrderPeakTimeList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(orderPeakTimeList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertOrderPeakTime(t *testing.T) {
	testConvertOrderPeakTime()
}

func testConvertOrderPeakTime() {
	InitializeSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	orderPeakTimeService := OrderPeakTimeService{db: db, targetDB: targetDB}
	err = orderPeakTimeService.ConvertOrderPeakTime()
	if err != nil {
		panic(err)
	}
}
