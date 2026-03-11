//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/utils"
)

func TestUserCard(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	userCardService := UserCardService{db: db}
	userCardList, err := userCardService.GetUserCardList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(userCardList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertUserCard(t *testing.T) {
	testConvertUserCard()
}

func testConvertUserCard() {
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
	userCardService := UserCardService{db: db, targetDB: targetDB}
	err = userCardService.ConvertUserCard()
	if err != nil {
		panic(err)
	}
	fmt.Println("user_card转换完成")
}
