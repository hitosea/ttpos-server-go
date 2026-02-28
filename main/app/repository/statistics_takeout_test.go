package repository

import (
	"strconv"
	"strings"
	"testing"

	valueobject "ttpos-server-go/app/modules/takeout/domain/value_object"
)

// TestBuildDynamicTimeCondition_WithTableAlias 测试带表别名的动态时间条件
func TestBuildDynamicTimeCondition_WithTableAlias(t *testing.T) {
	timeStart := int64(1700000000)
	timeEnd := int64(1700003600)

	result := buildDynamicTimeCondition("ttpos_takeout_order", timeStart, timeEnd)

	// 验证包含已完成订单的条件
	completedState := valueobject.TakeoutOrderStateCompleted
	if !strings.Contains(result, "ttpos_takeout_order.order_state = "+strconv.Itoa(completedState)) {
		t.Errorf("Expected condition to contain completed state check, got: %s", result)
	}

	// 验证包含完成时间范围
	if !strings.Contains(result, "ttpos_takeout_order.completed_time >= "+strconv.FormatInt(timeStart, 10)) {
		t.Errorf("Expected condition to contain completed_time start, got: %s", result)
	}
	if !strings.Contains(result, "ttpos_takeout_order.completed_time <= "+strconv.FormatInt(timeEnd, 10)) {
		t.Errorf("Expected condition to contain completed_time end, got: %s", result)
	}

	// 验证包含接单时间范围
	if !strings.Contains(result, "ttpos_takeout_order.accepted_time >= "+strconv.FormatInt(timeStart, 10)) {
		t.Errorf("Expected condition to contain accepted_time start, got: %s", result)
	}
	if !strings.Contains(result, "ttpos_takeout_order.accepted_time <= "+strconv.FormatInt(timeEnd, 10)) {
		t.Errorf("Expected condition to contain accepted_time end, got: %s", result)
	}

	// 验证包含 OR 逻辑
	if !strings.Contains(result, "OR") {
		t.Errorf("Expected condition to contain OR logic, got: %s", result)
	}

	// 验证包含 completed_time > 0 边界条件检查
	if !strings.Contains(result, "ttpos_takeout_order.completed_time > 0") {
		t.Errorf("Expected condition to contain completed_time > 0 check for boundary condition, got: %s", result)
	}
}

// TestBuildDynamicTimeCondition_WithDifferentAlias 测试不同表别名
func TestBuildDynamicTimeCondition_WithDifferentAlias(t *testing.T) {
	timeStart := int64(1700000000)
	timeEnd := int64(1700003600)

	result := buildDynamicTimeCondition("to_order", timeStart, timeEnd)

	// 验证使用正确的表别名
	if !strings.Contains(result, "to_order.order_state") {
		t.Errorf("Expected condition to use 'to_order' alias, got: %s", result)
	}
	if !strings.Contains(result, "to_order.completed_time") {
		t.Errorf("Expected condition to contain to_order.completed_time, got: %s", result)
	}
	if !strings.Contains(result, "to_order.accepted_time") {
		t.Errorf("Expected condition to contain to_order.accepted_time, got: %s", result)
	}
}

// TestBuildDynamicTimeCondition_EmptyAlias 测试空表别名
func TestBuildDynamicTimeCondition_EmptyAlias(t *testing.T) {
	timeStart := int64(1700000000)
	timeEnd := int64(1700003600)

	result := buildDynamicTimeCondition("", timeStart, timeEnd)

	// 验证不包含点号前缀，字段名直接使用
	if !strings.Contains(result, "order_state = ") {
		t.Errorf("Expected condition to contain 'order_state = ' without prefix, got: %s", result)
	}
	if !strings.Contains(result, "completed_time >= ") {
		t.Errorf("Expected condition to contain 'completed_time >= ' without prefix, got: %s", result)
	}
	if !strings.Contains(result, "accepted_time >= ") {
		t.Errorf("Expected condition to contain 'accepted_time >= ' without prefix, got: %s", result)
	}
}

// TestBuildDynamicTimeCondition_CompletedStateValue 测试使用正确的完成状态值
func TestBuildDynamicTimeCondition_CompletedStateValue(t *testing.T) {
	timeStart := int64(1700000000)
	timeEnd := int64(1700003600)

	result := buildDynamicTimeCondition("t", timeStart, timeEnd)

	// 验证使用的完成状态值是 40 (TakeoutOrderStateCompleted)
	expectedState := valueobject.TakeoutOrderStateCompleted
	if expectedState != 40 {
		t.Errorf("Expected TakeoutOrderStateCompleted to be 40, got: %d", expectedState)
	}

	if !strings.Contains(result, "t.order_state = 40") {
		t.Errorf("Expected condition to use completed state 40, got: %s", result)
	}
}

// TestBuildDynamicTimeCondition_BusinessScenarios 测试业务场景
// 场景：订单 1:30 接单、2:15 完成
// - 查询 1-2 点：已完成订单不应包含（因为完成时间是 2:15）
// - 查询 2-3 点：已完成订单应包含（因为完成时间是 2:15）
// - 查询 1-2 点：未完成订单应包含（因为接单时间是 1:30）
func TestBuildDynamicTimeCondition_BusinessScenarios(t *testing.T) {
	// 时间定义（使用固定时间戳便于测试）
	// 1:00 = 3600, 1:30 = 5400, 2:00 = 7200, 2:15 = 8100, 3:00 = 10800
	hour1Start := int64(3600)
	hour1End := int64(7199) // 1:59:59
	hour2Start := int64(7200)
	hour2End := int64(10799) // 2:59:59

	// 生成 1-2 点的条件
	condition1to2 := buildDynamicTimeCondition("t", hour1Start, hour1End)

	// 验证条件结构正确
	// 已完成订单使用 completed_time
	if !strings.Contains(condition1to2, "t.order_state = 40") {
		t.Errorf("1-2 hour condition should check for completed state 40")
	}
	if !strings.Contains(condition1to2, "t.completed_time >= "+strconv.FormatInt(hour1Start, 10)) {
		t.Errorf("1-2 hour condition should filter completed_time >= %d", hour1Start)
	}
	if !strings.Contains(condition1to2, "t.completed_time <= "+strconv.FormatInt(hour1End, 10)) {
		t.Errorf("1-2 hour condition should filter completed_time <= %d", hour1End)
	}

	// 非完成订单使用 accepted_time
	if !strings.Contains(condition1to2, "t.order_state != 40") {
		t.Errorf("1-2 hour condition should check for non-completed state")
	}
	if !strings.Contains(condition1to2, "t.accepted_time >= "+strconv.FormatInt(hour1Start, 10)) {
		t.Errorf("1-2 hour condition should filter accepted_time >= %d", hour1Start)
	}
	if !strings.Contains(condition1to2, "t.accepted_time <= "+strconv.FormatInt(hour1End, 10)) {
		t.Errorf("1-2 hour condition should filter accepted_time <= %d", hour1End)
	}

	// 生成 2-3 点的条件
	condition2to3 := buildDynamicTimeCondition("t", hour2Start, hour2End)

	// 验证 2-3 点条件的时间范围
	if !strings.Contains(condition2to3, "t.completed_time >= "+strconv.FormatInt(hour2Start, 10)) {
		t.Errorf("2-3 hour condition should filter completed_time >= %d", hour2Start)
	}
	if !strings.Contains(condition2to3, "t.completed_time <= "+strconv.FormatInt(hour2End, 10)) {
		t.Errorf("2-3 hour condition should filter completed_time <= %d", hour2End)
	}
}

// TestBuildDynamicTimeCondition_CompletedTimeZeroBoundary 测试 completed_time = 0 的边界条件
// 场景：订单已完成但 completed_time 未设置（异常数据）
// 预期行为：异常数据应被显式排除，而非静默回退到 accepted_time
// 原因：静默回退会导致统计口径不一致、难以发现数据问题
func TestBuildDynamicTimeCondition_CompletedTimeZeroBoundary(t *testing.T) {
	timeStart := int64(1700000000)
	timeEnd := int64(1700003600)

	result := buildDynamicTimeCondition("t", timeStart, timeEnd)

	// 验证包含 completed_time > 0 的边界条件检查
	// 已完成订单若 completed_time = 0，则不纳入统计（显式排除异常数据）
	if !strings.Contains(result, "t.completed_time > 0") {
		t.Errorf("Expected condition to contain 'completed_time > 0' check to exclude anomalous completed orders with zero completed_time, got: %s", result)
	}
}

// TestBuildDynamicTimeCondition_TableUsedInRealQueries 测试实际使用的表别名
func TestBuildDynamicTimeCondition_TableUsedInRealQueries(t *testing.T) {
	timeStart := int64(1700000000)
	timeEnd := int64(1700003600)

	testCases := []struct {
		name       string
		tableAlias string
	}{
		{"full table name", "ttpos_takeout_order"},
		{"to_order alias", "to_order"},
		{"t alias", "t"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := buildDynamicTimeCondition(tc.tableAlias, timeStart, timeEnd)

			// 验证每个别名都正确应用
			expectedPrefix := tc.tableAlias + "."
			if !strings.Contains(result, expectedPrefix+"order_state") {
				t.Errorf("Expected %s prefix for order_state, got: %s", expectedPrefix, result)
			}
			if !strings.Contains(result, expectedPrefix+"completed_time") {
				t.Errorf("Expected %s prefix for completed_time, got: %s", expectedPrefix, result)
			}
			if !strings.Contains(result, expectedPrefix+"accepted_time") {
				t.Errorf("Expected %s prefix for accepted_time, got: %s", expectedPrefix, result)
			}
		})
	}
}

// TestBuildDynamicTimeCondition_ORLogicStructure 测试 OR 逻辑结构正确性
func TestBuildDynamicTimeCondition_ORLogicStructure(t *testing.T) {
	timeStart := int64(1700000000)
	timeEnd := int64(1700003600)

	result := buildDynamicTimeCondition("t", timeStart, timeEnd)

	// 验证条件包含两个分支的 OR 结构
	// 分支1: (order_state = 40 AND completed_time > 0 AND completed_time >= X AND completed_time <= Y)
	// 分支2: (order_state != 40 AND accepted_time >= X AND accepted_time <= Y)

	if !strings.Contains(result, "OR") {
		t.Errorf("Expected OR logic in condition, got: %s", result)
	}

	// 计算括号数量，验证结构完整性
	openParens := strings.Count(result, "(")
	closeParens := strings.Count(result, ")")
	if openParens != closeParens {
		t.Errorf("Unbalanced parentheses: %d open, %d close", openParens, closeParens)
	}

	// 验证 AND 连接词存在
	andCount := strings.Count(result, "AND")
	if andCount < 4 { // 至少应该有 4 个 AND（completed_time > 0 AND completed_time >= AND completed_time <= 和 accepted_time >= AND accepted_time <=）
		t.Errorf("Expected at least 4 AND connectors, got: %d", andCount)
	}
}
