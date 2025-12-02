package lock

import (
	"reflect"
	"testing"
)

// mockLock 用于测试的锁实现
type mockLock struct {
	lockedUuids []uint64
	unlockedUuids []uint64
}

func newMockLock() *mockLock {
	return &mockLock{
		lockedUuids:   make([]uint64, 0),
		unlockedUuids: make([]uint64, 0),
	}
}

func (m *mockLock) LockUuid(uuid uint64) {
	m.lockedUuids = append(m.lockedUuids, uuid)
}

func (m *mockLock) UnlockUuid(uuid uint64) {
	m.unlockedUuids = append(m.unlockedUuids, uuid)
}

func (m *mockLock) ClearUuidLock(uuid uint64) {
	// 测试中不使用
}

func (m *mockLock) LockUuidString(uuid string) {
	// 测试中不使用
}

func (m *mockLock) TryLockUuidString(uuid string) bool {
	// 测试中不使用
	return false
}

func (m *mockLock) UnlockUuidString(uuid string) {
	// 测试中不使用
}

func (m *mockLock) ClearUuidLockString(uuid string) {
	// 测试中不使用
}

// TestSortAndDeduplicateUuids 测试去重和排序逻辑（通过 LockMultipleUuids 间接测试）
func TestSortAndDeduplicateUuids(t *testing.T) {
	tests := []struct {
		name     string
		input    []uint64
		expected []uint64
	}{
		{
			name:     "空列表",
			input:    []uint64{},
			expected: []uint64{},
		},
		{
			name:     "单个 UUID",
			input:    []uint64{100},
			expected: []uint64{100},
		},
		{
			name:     "已排序的 UUID 列表",
			input:    []uint64{100, 200, 300},
			expected: []uint64{100, 200, 300},
		},
		{
			name:     "未排序的 UUID 列表",
			input:    []uint64{300, 100, 200},
			expected: []uint64{100, 200, 300},
		},
		{
			name:     "包含重复 UUID",
			input:    []uint64{200, 100, 200, 300, 100},
			expected: []uint64{100, 200, 300},
		},
		{
			name:     "包含无效 UUID（0 值）",
			input:    []uint64{200, 0, 100, 0, 300},
			expected: []uint64{100, 200, 300},
		},
		{
			name:     "全部为无效 UUID",
			input:    []uint64{0, 0, 0},
			expected: []uint64{},
		},
		{
			name:     "包含重复和无效 UUID",
			input:    []uint64{200, 0, 100, 200, 0, 300, 100},
			expected: []uint64{100, 200, 300},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockLock()
			// 通过 LockMultipleUuids 间接测试 sortAndDeduplicateUuids
			result := LockMultipleUuids(mock, tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("sortAndDeduplicateUuids() (via LockMultipleUuids) = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestLockMultipleUuids 测试锁定多个 UUID
func TestLockMultipleUuids(t *testing.T) {
	tests := []struct {
		name          string
		input         []uint64
		expectedOrder []uint64
	}{
		{
			name:          "空列表",
			input:         []uint64{},
			expectedOrder: []uint64{},
		},
		{
			name:          "单个 UUID",
			input:         []uint64{100},
			expectedOrder: []uint64{100},
		},
		{
			name:          "未排序的 UUID 列表",
			input:         []uint64{300, 100, 200},
			expectedOrder: []uint64{100, 200, 300},
		},
		{
			name:          "包含重复 UUID",
			input:         []uint64{200, 100, 200, 300},
			expectedOrder: []uint64{100, 200, 300},
		},
		{
			name:          "包含无效 UUID",
			input:         []uint64{200, 0, 100, 300},
			expectedOrder: []uint64{100, 200, 300},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockLock()
			result := LockMultipleUuids(mock, tt.input)

			// 验证返回的 UUID 列表已排序
			if !reflect.DeepEqual(result, tt.expectedOrder) {
				t.Errorf("LockMultipleUuids() returned = %v, want %v", result, tt.expectedOrder)
			}

			// 验证锁获取顺序正确（按 UUID 升序）
			if !reflect.DeepEqual(mock.lockedUuids, tt.expectedOrder) {
				t.Errorf("LockMultipleUuids() locked order = %v, want %v", mock.lockedUuids, tt.expectedOrder)
			}
		})
	}
}

// TestUnlockMultipleUuids 测试释放多个 UUID 锁
func TestUnlockMultipleUuids(t *testing.T) {
	tests := []struct {
		name            string
		input           []uint64
		expectedOrder   []uint64 // 排序后的顺序
		expectedReverse []uint64 // 释放顺序（相反顺序）
	}{
		{
			name:            "空列表",
			input:           []uint64{},
			expectedOrder:   []uint64{},
			expectedReverse: []uint64{},
		},
		{
			name:            "单个 UUID",
			input:           []uint64{100},
			expectedOrder:   []uint64{100},
			expectedReverse: []uint64{100},
		},
		{
			name:            "未排序的 UUID 列表",
			input:           []uint64{300, 100, 200},
			expectedOrder:   []uint64{100, 200, 300},
			expectedReverse: []uint64{300, 200, 100},
		},
		{
			name:            "包含重复 UUID",
			input:           []uint64{200, 100, 200, 300},
			expectedOrder:   []uint64{100, 200, 300},
			expectedReverse: []uint64{300, 200, 100},
		},
		{
			name:            "包含无效 UUID",
			input:           []uint64{200, 0, 100, 300},
			expectedOrder:   []uint64{100, 200, 300},
			expectedReverse: []uint64{300, 200, 100},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockLock()
			UnlockMultipleUuids(mock, tt.input)

			// 验证释放顺序正确（按相反顺序）
			if !reflect.DeepEqual(mock.unlockedUuids, tt.expectedReverse) {
				t.Errorf("UnlockMultipleUuids() unlocked order = %v, want %v", mock.unlockedUuids, tt.expectedReverse)
			}
		})
	}
}

// TestLockUnlockConsistency 测试 LockMultipleUuids 和 UnlockMultipleUuids 的一致性
func TestLockUnlockConsistency(t *testing.T) {
	tests := []struct {
		name  string
		input []uint64
	}{
		{
			name:  "正常 UUID 列表",
			input: []uint64{300, 100, 200},
		},
		{
			name:  "包含重复 UUID",
			input: []uint64{200, 100, 200, 300, 100},
		},
		{
			name:  "包含无效 UUID",
			input: []uint64{200, 0, 100, 0, 300},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockLock()

			// 锁定
			lockedUuids := LockMultipleUuids(mock, tt.input)

			// 验证锁定顺序
			if !reflect.DeepEqual(mock.lockedUuids, lockedUuids) {
				t.Errorf("LockMultipleUuids() locked order = %v, returned = %v", mock.lockedUuids, lockedUuids)
			}

			// 重置解锁记录
			mock.unlockedUuids = make([]uint64, 0)

			// 释放锁（使用返回的列表）
			UnlockMultipleUuids(mock, lockedUuids)

			// 验证释放顺序是锁定顺序的相反顺序
			expectedReverse := make([]uint64, len(lockedUuids))
			for i := 0; i < len(lockedUuids); i++ {
				expectedReverse[i] = lockedUuids[len(lockedUuids)-1-i]
			}

			if !reflect.DeepEqual(mock.unlockedUuids, expectedReverse) {
				t.Errorf("UnlockMultipleUuids() unlocked order = %v, want %v", mock.unlockedUuids, expectedReverse)
			}
		})
	}
}

