package old_model

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/utils"
)

func TestUser(t *testing.T) {
	db, err := NewMySQLConnection(conf, dbName)
	if err != nil {
		panic(err)
	}
	userService := UserService{db: db}
	userList, err := userService.GetUserList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(userList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertUser(t *testing.T) {
	testConvertUser()
}

func testConvertUser() {
	InitializeSonyFlakeId()

	utils.InitSonyFlakeId()

	db, err := NewMySQLConnection(conf, dbName)
	if err != nil {
		panic(err)
	}

	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	userService := UserService{db: db, targetDB: targetDB}
	err = userService.ConvertUser()
	if err != nil {
		panic(err)
	}
	fmt.Println("user转换完成")
}
