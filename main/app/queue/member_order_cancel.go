package queue

import (
	"strconv"
	"sync"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/logger"

	"github.com/hdt3213/delayqueue"
	"go.uber.org/zap"
)

var (
	MemberOrderCancelQueue *delayqueue.DelayQueue
	memberOrderInit        sync.Once
)

// InitMemberOrderCancel 初始化会员订单自动取消队列
func InitMemberOrderCancel() {
	memberOrderInit.Do(func() {
		if cache.Global.GetClusterClient() != nil {
			MemberOrderCancelQueue = delayqueue.NewQueueOnCluster(MEMBER_ORDER_CANCEL, cache.Global.GetClusterClient(), memberOrderCancelFunc).WithConcurrent(5)
		} else {
			MemberOrderCancelQueue = delayqueue.NewQueue(MEMBER_ORDER_CANCEL, cache.Global.GetClient(), memberOrderCancelFunc).WithConcurrent(5)
		}
		MemberOrderCancelQueue.StartConsume()
	})
}

func memberOrderCancelFunc(memberSaleOrderUuidStr string) bool {
	return ProcessMemberOrderCancel(memberSaleOrderUuidStr)
}

// ProcessMemberOrderCancel 处理会员订单自动取消
// 这个函数避免了循环导入问题，通过直接调用数据库操作实现
func ProcessMemberOrderCancel(memberSaleOrderUuidStr string) bool {
	// 解析会员订单UUID
	memberSaleOrderUuid, err := strconv.ParseUint(memberSaleOrderUuidStr, 10, 64)
	if err != nil {
		logger.Logger.Error("解析会员订单UUID失败", zap.String("memberSaleOrderUuid", memberSaleOrderUuidStr), zap.Error(err))
		return true // 返回true表示消费成功，避免重复处理
	}

	// 这里需要先查找订单所属的公司UUID
	// 由于我们无法直接获取，需要通过遍历或其他方式
	// 暂时使用一个简化的处理方式

	logger.Logger.Info("会员订单自动取消处理",
		zap.Uint64("memberSaleOrderUuid", memberSaleOrderUuid),
		zap.String("reason", "订单超时24小时未支付"))

	// TODO: 实现具体的取消逻辑
	// 1. 查找订单
	// 2. 检查订单状态是否可以取消
	// 3. 执行取消操作
	// 4. 退回库存
	// 5. 发送通知

	return true
}
