package uuid

import (
	"context"

	goid "github.com/ace-zhaoy/go-id"
	"github.com/gogf/gf/v2/frame/g"
)

// 应用类型定义（1-15）
// 用于区分不同应用，避免跨应用 ID 冲突
const (
	AppTypeUnknown   uint32 = 0 // 未知/默认（不推荐使用）
	AppTypeERP       uint32 = 1 // ttpos-erp
	AppTypeMessage   uint32 = 2 // ttpos-message
	AppTypeManager   uint32 = 3 // ttpos-manager
	AppTypeShop      uint32 = 4 // ttpos-shop
	AppTypeTakeout   uint32 = 5 // ttpos-takeout
	AppTypeWebsocket uint32 = 6 // ttpos-websocket
	// 预留 7-15 给未来应用
)

// 节点 ID 位数配置
const (
	appTypeBits    = 4  // 应用类型位数（支持 1-15 种应用）
	instanceIDBits = 6  // 实例 ID 位数（支持 1-63 个实例）
	totalNodeBits  = 10 // 总节点位数（支持 1-1023 个节点）
)

var (
	idGenerator *goid.ID
)

// InitIdGenerator 初始化 ID 生成器
// appType: 应用类型，使用预定义常量（AppTypeERP, AppTypeMessage 等）
// 实例 ID 从配置 SERVER_ID 读取（1-63），默认 1
//
// 节点 ID 计算公式：node_id = (appType << 6) | instanceID
// 示例：
//   - ttpos-erp 实例1：(1 << 6) | 1 = 65
//   - ttpos-message 实例1：(2 << 6) | 1 = 129
//
// 为避免多实例在同一秒生成 ID 时发生冲突，初始化时会设置随机增量
func InitIdGenerator(ctx context.Context, appType uint32) {
	// 创建 ID 生成器实例
	idGenerator = goid.NewID()

	// 验证应用类型（1-15）
	if appType < 1 || appType > 15 {
		g.Log().Warningf(ctx, "[UUID] Invalid appType %d, using default 0", appType)
		appType = AppTypeUnknown
	}

	// 获取实例 ID（1-63）
	instanceID := uint32(g.Cfg().MustGetWithEnv(ctx, "SERVER_ID", "1").Int())
	if instanceID < 1 || instanceID > 63 {
		g.Log().Warningf(ctx, "[UUID] Invalid SERVER_ID %d, using default 1", instanceID)
		instanceID = 1
	}

	// 组合节点 ID: (appType << 6) | instanceID
	nodeID := (appType << instanceIDBits) | instanceID

	// 设置节点（10 位 nodeBits）
	idGenerator.SetNode(nodeID, totalNodeBits)

}

// GetID 获取id
func GetID() (uint64, error) {
	return uint64(idGenerator.Generate()), nil
}

// MustGetID 获取唯一ID数字
func MustGetID() uint64 {
	return uint64(idGenerator.Generate())
}
