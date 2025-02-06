package old_model

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
)

func TestGetAttributeList(t *testing.T) {
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
	attributeRepository := AttributeRepository{db: db}
	attributes, err := attributeRepository.GetAttributeList()
	json, _ := json.Marshal(attributes)
	fmt.Println(string(json))
}

func TestConvertAttribute(t *testing.T) {
	database.InitSonyFlakeId()
	db, err := NewMySQLConnection(config.DatabaseConf{
		Host:         "localhost",
		Port:         3306,
		User:         "root",
		Password:     "yourpassword",
		RootPassword: "yourpassword",
		TablePrefix:  "jjjfood_",
	}, "shop_wang")
	if err != nil {
		panic(err)
	}
	targetDB, err := NewMySQLConnection(config.DatabaseConf{
		Host:         "localhost",
		Port:         3306,
		User:         "root",
		Password:     "yourpassword",
		RootPassword: "yourpassword",
		TablePrefix:  "ttpos_",
	}, "shop_wang")
	if err != nil {
		panic(err)
	}
	attributeRepository := AttributeRepository{db: db, targetDB: targetDB}
	err = attributeRepository.ConvertAttribute()
	if err != nil {
		panic(err)
	}
}
