package old_model

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/database"
)

func TestGetAttributeList(t *testing.T) {
	db, err := NewMySQLConnection(conf, dbName)
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
	db, err := NewMySQLConnection(conf, dbName)
	if err != nil {
		panic(err)
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	attributeRepository := AttributeRepository{db: db, targetDB: targetDB}
	err = attributeRepository.ConvertAttribute()
	if err != nil {
		panic(err)
	}
}
