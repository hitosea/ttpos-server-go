package old_model

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
)

func TestFreeTag(t *testing.T) {
	db, err := NewMySQLConnection(config.DatabaseConf{
		Host:          "localhost",
		Port:          3306,
		User:          "root",
		Password:      "yourpassword",
		RootPassword:  "yourpassword",
		TablePrefix:   "jjjfood_",
		SlowQueryTime: 0,
	}, "shop_wang")
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
	database.InitSonyFlakeId()

	db, err := NewMySQLConnection(config.DatabaseConf{
		Host:          "localhost",
		Port:          3306,
		User:          "root",
		Password:      "yourpassword",
		RootPassword:  "yourpassword",
		TablePrefix:   "jjjfood_",
		SlowQueryTime: 0,
	}, "shop_wang")
	if err != nil {
		panic(err)
	}
	targetDB, err := NewMySQLConnection(config.DatabaseConf{
		Host:          "localhost",
		Port:          3306,
		User:          "root",
		Password:      "yourpassword",
		RootPassword:  "yourpassword",
		TablePrefix:   "ttpos_",
		SlowQueryTime: 0,
	}, "shop_wang")
	if err != nil {
		panic(err)
	}
	freeTagService := FreeTagService{db: db, targetDB: targetDB}
	err = freeTagService.ConvertFreeTag()
	if err != nil {
		panic(err)
	}
}
