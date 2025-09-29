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

// TryLockUuid 非阻塞尝试获取uuid锁
func (d *LocalLock) TryLockUuid(uuid uint64) bool {
	return d.getUuidLock(uuid).TryLock()
}

// UnlockUuid 解锁uuid
func (d *LocalLock) UnlockUuid(uuid uint64) {
	d.getUuidLock(uuid).Unlock()
}

// ClearUuidLock 在uuid对应的资源完成或删除后，清除uuid锁
func (d *LocalLock) ClearUuidLock(uuid uint64) {
	d.uuidLock.Delete(uuid)
}
