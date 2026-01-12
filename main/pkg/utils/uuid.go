package utils

import (
	"fmt"
	"math/rand"
	"time"

	goid "github.com/ace-zhaoy/go-id"
	"github.com/spf13/viper"

	"github.com/sony/sonyflake"
)

// 参考：https://jicki.cn/golang-web-note-10/
var (
	sonyFlake   *sonyflake.Sonyflake
	idGenerator *goid.ID
)

const (
	totalNodeBits = 10 // 扩展为 10 位，支持 1-1023 个节点
)

// 初始化id生成器
func InitIdGenerator() {
	// 创建ID生成器实例
	idGenerator = goid.NewID()

	// 获取SERVER_ID并验证有效性
	serverID := uint32(viper.GetInt("SERVER_ID"))
	// 对于10位nodeBits，有效范围是1-1023
	if serverID < 1 || serverID > 1023 {
		// 不符合预期则使用默认值：随机值（1-1023）
		// Go 1.20+ 全局随机数生成器已自动初始化，无需手动 Seed
		serverID = uint32(rand.Intn(1023) + 1)
		fmt.Printf("[UUID] SERVER_ID未配置或超出范围，使用随机值: %d\n", serverID)
	}

	// 设置节点ID（使用10位nodeBits）
	idGenerator.SetNode(serverID, totalNodeBits)

	// 设置随机增量，避免多实例在同一秒生成 ID 时发生冲突
	// 计数器位数 = 21 - nodeBits = 21 - 10 = 11
	// 最大随机增量 = (1 << 11) - 1 = 2047
	// 使用纳秒时间戳生成随机增量，确保每个实例的增量不同
	counterBits := uint32(21 - totalNodeBits)
	maxRandomDelta := uint32((1 << counterBits) - 1)
	randomDelta := uint32((time.Now().UnixNano() % int64(maxRandomDelta)) + 2)
	if randomDelta >= maxRandomDelta {
		randomDelta = maxRandomDelta - 1
	}
	idGenerator.SetRandomDelta(randomDelta)

	fmt.Printf("[UUID] ID生成器初始化成功: SERVER_ID=%d, nodeBits=%d, randomDelta=%d\n", serverID, totalNodeBits, randomDelta)
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
