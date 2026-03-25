//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/utils"
)

func TestUserGrade(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	userService := UserGradeService{db: db}
	userList, err := userService.GetUserGradeList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(userList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertUserGrade(t *testing.T) {
	testConvertUserGrade()
}

func testConvertUserGrade() {
	InitializeSonyFlakeId()

	utils.InitSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}

	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		panic(err)
	}
	userService := UserGradeService{db: db, targetDB: targetDB}
	err = userService.ConvertUserGrade()
	if err != nil {
		panic(err)
	}
	fmt.Println("user_grade转换完成")
}
