package utils

import (
	goid "github.com/ace-zhaoy/go-id"
	"github.com/spf13/viper"
)

var (
	idGenerator *goid.ID
)

// 初始化id生成器
func InitIdGenerator() {
	// 创建ID生成器实例
	idGenerator = goid.NewID()
	// 获取SERVER_ID并验证有效性
	serverID := uint32(viper.GetInt("SERVER_ID"))
	// 对于4位nodeBits，有效范围是1-15
	if serverID < 1 || serverID > 15 {
		// 不符合预期则使用默认值1
		serverID = 1
	}
	// 设置节点ID
	idGenerator.SetNode(serverID, 4)
}

func GetID() uint64 {
	id := goid.GenID()
	return uint64(id)
}
