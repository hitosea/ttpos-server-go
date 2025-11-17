package lock

import "sync"

const (
	DailySalesOutboundSummaryLock = 1 // 每日销售出库汇总锁
	KitchenEfficiencyAnalysisLock = 2 // 后厨效率分析锁
)

type Lock interface {
	LockUuid(uuid uint64)
	UnlockUuid(uuid uint64)
	ClearUuidLock(uuid uint64)
	// 字符串锁
	LockUuidString(uuid string)
	TryLockUuidString(uuid string) bool // 非阻塞尝试获取锁，返回是否成功获取
	UnlockUuidString(uuid string)
	ClearUuidLockString(uuid string)
}

var systemLock Lock
var once sync.Once

// NewSystemLock 创建系统锁
func NewSystemLock() Lock {
	once.Do(func() {
		//systemLock = InitLocalLock() // 本地锁，仅适用于单体应用
		systemLock = NewRedSyncLock(NewRedSync()) // 分布式锁
	})
	return systemLock
}
