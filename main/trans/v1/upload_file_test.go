//go:build integration

package v1

import (
	"encoding/json"
	"fmt"
	"testing"
	"ttpos-server-go/pkg/utils"
)

func TestUploadFile(t *testing.T) {
	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		panic(err)
	}
	uploadFileService := UploadFileService{db: db}
	uploadFileList, err := uploadFileService.GetUploadFileList()
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(uploadFileList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func TestConvertUploadFile(t *testing.T) {
	testConvertUploadFile()
}

func testConvertUploadFile() {
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
	uploadFileService := UploadFileService{db: db, targetDB: targetDB}
	err = uploadFileService.ConvertUploadFile()
	if err != nil {
		panic(err)
	}
	fmt.Println("upload_file转换完成")
}
