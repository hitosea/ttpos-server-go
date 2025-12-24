package objectstorage

import (
	"ttpos-server-go/app/modules/objectstorage/application"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
)

// IObjectStorageAppService 对象存储应用服务接口（向后兼容）
type IObjectStorageAppService = application.IObjectStorageAppService

// NewObjectStorageAppService 创建对象存储应用服务
func NewObjectStorageAppService(
	cacheInstance cache.Cache,
	dbm *database.DBManager,
) IObjectStorageAppService {
	return application.NewObjectStorageAppService(cacheInstance, dbm)
}

