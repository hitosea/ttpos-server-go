package service

import (
	"testing"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
)

func TestGetInvalidProductList(t *testing.T) {
	dbm := database.GetDBManager(config.DatabaseConf{
		DBType:        "mysql",
		Host:          "192.168.100.58",
		Port:          3306,
		User:          "root",
		Password:      "5cd6a0408e9ccf92",
		TablePrefix:   "ttpos_",
		Database:      "shop1",
		SlowQueryTime: 0,
	})
	list, err := NewOrderProductSrv(dbm).GetInvalidProductList(1, 234)
	if err != nil {
		t.Error(err)
	}
	t.Log(list)
}
