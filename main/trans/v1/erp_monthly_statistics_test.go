//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestErpMonthlyStatistics(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	erpMonthlyStatisticsService := ERPMonthlyStatisticsService{db: db}
	erpMonthlyStatisticsList, err := erpMonthlyStatisticsService.GetERPMonthlyStatisticsList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(erpMonthlyStatisticsList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertErpMonthlyStatistics(t *testing.T) {
	testConvertErpMonthlyStatistics()
}

func testConvertErpMonthlyStatistics() {
	InitializeSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	erpMonthlyStatisticsService := ERPMonthlyStatisticsService{db: db, targetDB: targetDB}
	err = erpMonthlyStatisticsService.ConvertERPMonthlyStatistics()
	if err != nil {
		panic(err)
	}
}
