//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/utils"
)

func TestUploadGroup(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	uploadGroupService := UploadGroupService{db: db}
	uploadGroupList, err := uploadGroupService.GetUploadGroupList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(uploadGroupList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertUploadGroup(t *testing.T) {
	testConvertUploadGroup()
}

func testConvertUploadGroup() {
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
	uploadGroupService := UploadGroupService{db: db, targetDB: targetDB}
	err = uploadGroupService.ConvertUploadGroup()
	if err != nil {
		panic(err)
	}
	fmt.Println("upload_group转换完成")
}
