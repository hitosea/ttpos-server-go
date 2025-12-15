package utils

import (
	"fmt"

	goid "github.com/ace-zhaoy/go-id"
	"github.com/spf13/viper"

	"github.com/sony/sonyflake"
)

// 参考：https://jicki.cn/golang-web-note-10/
var (
	sonyFlake   *sonyflake.Sonyflake
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
	idGenerator.SetNode(13, 4)
}

// InitSonyFlakeId 初始化 sonyFlake 配置
func InitSonyFlakeId() {
	st := sonyflake.Settings{}
	sonyFlake = sonyflake.NewSonyflake(st)
}

// GetID 获取全局 Uuid 的函数
func GetID18() (id uint64, err error) {
	if sonyFlake == nil {
		err = fmt.Errorf("需要先初始化以后再执行 GetID 函数 err: %#v \n", err)
		return
	}
	return sonyFlake.NextID()
}

// 获取id
func GetID() (uint64, error) {
	return uint64(idGenerator.Generate()), nil
}

// MustGetID 获取唯一ID数字
func MustGetID() uint64 {
	return uint64(idGenerator.Generate())
}
