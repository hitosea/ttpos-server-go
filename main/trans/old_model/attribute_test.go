package old_model

import (
	"encoding/json"
	"fmt"
	"sync"
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

var SonyFlakeIdOnce sync.Once

func InitializeSonyFlakeId() {
	SonyFlakeIdOnce.Do(func() {
		database.InitSonyFlakeId()
	})
}

func TestConvertAttribute(t *testing.T) {
	testConvertAttribute()
}

func testConvertAttribute() {
	InitializeSonyFlakeId()
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
