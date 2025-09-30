package lock

import "sync"

type LocalLock struct {
	uuidLock sync.Map
}

func InitLocalLock() *LocalLock {
	return &LocalLock{
		uuidLock: sync.Map{},
	}

}

// 获取uuid锁
func (d *LocalLock) getUuidLock(uuid uint64) *sync.Mutex {
	actual, _ := d.uuidLock.LoadOrStore(uuid, &sync.Mutex{})
	return actual.(*sync.Mutex)
}

// LockUuid 锁定uuid
func (d *LocalLock) LockUuid(uuid uint64) {
	d.getUuidLock(uuid).Lock()
}

func (d *LocalLock) getUuidLockString(uuid string) *sync.Mutex {
	actual, _ := d.uuidLock.LoadOrStore(uuid, &sync.Mutex{})
	return actual.(*sync.Mutex)
}

// UnlockUuid 解锁uuid
func (d *LocalLock) UnlockUuid(uuid uint64) {
	d.getUuidLock(uuid).Unlock()
}

// ClearUuidLock 在uuid对应的资源完成或删除后，清除uuid锁
func (d *LocalLock) ClearUuidLock(uuid uint64) {
	d.uuidLock.Delete(uuid)
}

// LockUuidString 锁定uuid
func (d *LocalLock) LockUuidString(uuid string) {
	d.getUuidLockString(uuid).Lock()
}

// TryLockUuidString 非阻塞尝试获取uuid锁
func (d *LocalLock) TryLockUuidString(uuid string) bool {
	return d.getUuidLockString(uuid).TryLock()
}

// UnlockUuidString 解锁uuid
func (d *LocalLock) UnlockUuidString(uuid string) {
	d.getUuidLockString(uuid).Unlock()
}

// ClearUuidLockString 在uuid对应的资源完成或删除后，清除uuid锁
func (d *LocalLock) ClearUuidLockString(uuid string) {
	d.uuidLock.Delete(uuid)
}
