package utility

import (
	"testing"

	"github.com/gogf/gf/v2/test/gtest"
)

// TestMapStatusToLineman 测试 TTPOS 状态到 Lineman 菜单状态的映射（string）
func TestMapStatusToLineman(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 测试 AVAILABLE
		status, err := MapStatusToLineman("AVAILABLE")
		t.AssertNil(err)
		t.Assert(status, "AVAILABLE")

		// 测试 UNAVAILABLE -> SUSPENDED
		status, err = MapStatusToLineman("UNAVAILABLE")
		t.AssertNil(err)
		t.Assert(status, "SUSPENDED")

		// 测试 HIDE -> SUSPENDED
		status, err = MapStatusToLineman("HIDE")
		t.AssertNil(err)
		t.Assert(status, "SUSPENDED")

		// 测试 UNAVAILABLETODAY -> SOLD_OUT_TODAY
		status, err = MapStatusToLineman("UNAVAILABLETODAY")
		t.AssertNil(err)
		t.Assert(status, "SOLD_OUT_TODAY")

		// 测试不支持的状态
		status, err = MapStatusToLineman("INVALID_STATUS")
		t.AssertNE(err, nil)
		t.Assert(status, "")
		t.AssertIN("不支持的状态", err.Error())
	})
}

// TestMapStatusToLinemanModifier 测试 TTPOS 状态到 Lineman 修饰符状态的映射（int）
func TestMapStatusToLinemanModifier(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 测试 AVAILABLE
		status, err := MapStatusToLinemanModifier("AVAILABLE")
		t.AssertNil(err)
		t.Assert(status, 1)

		// 测试 UNAVAILABLE
		status, err = MapStatusToLinemanModifier("UNAVAILABLE")
		t.AssertNil(err)
		t.Assert(status, 3)

		// 测试 UNAVAILABLETODAY -> SOLD_OUT_TODAY (2)
		status, err = MapStatusToLinemanModifier("UNAVAILABLETODAY")
		t.AssertNil(err)
		t.Assert(status, 2)

		// 测试空字符串
		status, err = MapStatusToLinemanModifier("")
		t.AssertNE(err, nil)
		t.Assert(status, 0)
		t.AssertIN("available_status 不能为空", err.Error())

		// 测试不支持的状态
		status, err = MapStatusToLinemanModifier("INVALID_STATUS")
		t.AssertNE(err, nil)
		t.Assert(status, 0)
		t.AssertIN("不支持的状态", err.Error())
	})
}
