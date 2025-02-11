package lock

import "sync"

type Lock interface {
	LockUuid(uuid uint64)
	UnlockUuid(uuid uint64)
	ClearUuidLock(uuid uint64)
}

type SystemLock struct {
	uuidLock sync.Map
}

var systemLock *SystemLock
var once sync.Once

// NewSystemLock 创建系统锁
func NewSystemLock() Lock {
	once.Do(func() {
		systemLock = &SystemLock{
			uuidLock: sync.Map{},
		}
	})
	return systemLock
}

// 获取uuid锁
func (d *SystemLock) getUuidLock(uuid uint64) *sync.Mutex {
	actual, _ := d.uuidLock.LoadOrStore(uuid, &sync.Mutex{})
	return actual.(*sync.Mutex)
}

// LockUuid 锁定uuid
func (d *SystemLock) LockUuid(uuid uint64) {
	d.getUuidLock(uuid).Lock()
}

// UnlockUuid 解锁uuid
func (d *SystemLock) UnlockUuid(uuid uint64) {
	d.getUuidLock(uuid).Unlock()
}

// ClearUuidLock 在uuid对应的资源完成或删除后，清除uuid锁
func (d *SystemLock) ClearUuidLock(uuid uint64) {
	d.uuidLock.Delete(uuid)
}
