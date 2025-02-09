package database

import (
	"fmt"
	"log"
	"testing"
)

func TestGenerateUUID(t *testing.T) {
	InitSonyFlakeId()
	id, err := GetID()
	if err != nil {
		log.Fatalf("GetID Failed Err: %#v\n", err)
	}
	fmt.Println("sonyFlake 生成 Uuid: ", id)
}
