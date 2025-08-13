package uuid

import (
	"context"

	goid "github.com/ace-zhaoy/go-id"
	"github.com/gogf/gf/v2/frame/g"
)

/**
参考原工程实现
*/

var (
	idGenerator *goid.ID
)

// InitIdGenerator 初始化id生成器
func InitIdGenerator(ctx context.Context) {
	// 创建ID生成器实例
	idGenerator = goid.NewID()
	// 获取SERVER_ID并验证有效性
	serverID := uint32(g.Cfg().MustGetWithEnv(ctx, "SERVER_ID", "1").Int())
	// 对于4位nodeBits，有效范围是1-15
	if serverID < 1 || serverID > 15 {
		// 不符合预期则使用默认值1
		serverID = 1
	}
	// 设置节点ID
	idGenerator.SetNode(serverID, 4)
}

// GetID 获取id
func GetID() (uint64, error) {
	return uint64(idGenerator.Generate()), nil
}

// MustGetID 获取唯一ID数字
func MustGetID() uint64 {
	return uint64(idGenerator.Generate())
}
