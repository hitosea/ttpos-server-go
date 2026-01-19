package lock

import (
	"sort"
)

// sortAndDeduplicateUuids 对 UUID 列表进行去重和排序
// 过滤无效 UUID（0 值），去重，并按 UUID 升序排序
// 返回排序后的唯一 UUID 列表
func sortAndDeduplicateUuids(uuids []uint64) []uint64 {
	// 去重并过滤无效 UUID
	uniqueUuids := make([]uint64, 0)
	seen := make(map[uint64]bool)
	for _, uuid := range uuids {
		if uuid != 0 && !seen[uuid] {
			uniqueUuids = append(uniqueUuids, uuid)
			seen[uuid] = true
		}
	}

	// 按 UUID 排序，确保锁获取顺序一致
	sort.Slice(uniqueUuids, func(i, j int) bool {
		return uniqueUuids[i] < uniqueUuids[j]
	})

	return uniqueUuids
}

// LockMultipleUuids 锁定多个 UUID（按 UUID 排序）
// 自动去重和过滤无效 UUID（0 值）
// 返回排序后的 UUID 列表（用于后续释放锁）
func LockMultipleUuids(lock Lock, uuids []uint64) []uint64 {
	// 使用公用方法进行去重和排序
	sortedUuids := sortAndDeduplicateUuids(uuids)

	// 依次获取锁
	for _, uuid := range sortedUuids {
		lock.LockUuid(uuid)
	}

	return sortedUuids
}

// UnlockMultipleUuids 释放多个 UUID 锁（按相反顺序）
// 使用与 LockMultipleUuids 相同的排序策略（去重、排序），然后按相反顺序释放
func UnlockMultipleUuids(lock Lock, uuids []uint64) {
	// 使用公用方法进行去重和排序（确保与 LockMultipleUuids 使用相同的策略）
	sortedUuids := sortAndDeduplicateUuids(uuids)

	// 按相反顺序释放锁
	for i := len(sortedUuids) - 1; i >= 0; i-- {
		lock.UnlockUuid(sortedUuids[i])
	}
}

