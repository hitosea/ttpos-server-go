package lock

import "sync"

type Lock interface {
	LockUuid(uuid uint64)
	UnlockUuid(uuid uint64)
	ClearUuidLock(uuid uint64)
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
