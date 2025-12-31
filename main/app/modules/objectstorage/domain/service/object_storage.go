package service

import (
	"context"

	"ttpos-server-go/app/modules/objectstorage/domain/entity"
)

// IObjectStorage 对象存储领域服务接口
type IObjectStorage[T any] interface {

	// PreloadWithConfig 配置映射自动关联注入（推荐方式）
	PreloadWithConfig(ctx context.Context, obj interface{}, associations []entity.Association) error
}
