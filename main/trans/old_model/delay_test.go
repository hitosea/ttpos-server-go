package old_model

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestBuffetDelay(t *testing.T) {
	db, err := NewMySQLConnection(conf, dbName)
	if err != nil {
		panic(err)
	}
	buffetDelayService := BuffetDelayService{db: db}
	buffetDelayList, err := buffetDelayService.GetBuffetDelayList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(buffetDelayList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertBuffetDelay(t *testing.T) {
	testConvertBuffetDelay()
}

func testConvertBuffetDelay() {
	InitializeSonyFlakeId()

	db, err := NewMySQLConnection(conf, dbName)
	if err != nil {
		panic(err)
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	buffetDelayService := BuffetDelayService{db: db, targetDB: targetDB}
	err = buffetDelayService.ConvertBuffetDelay()
	if err != nil {
		panic(err)
	}
}
