package utility

import (
	"testing"

	"github.com/gogf/gf/v2/test/gtest"
)

// TestMapStatusToLinemanModifier 测试 TTPOS 状态到 Lineman 修饰符状态的映射
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

		// 测试 SOLD_OUT_TODAY
		status, err = MapStatusToLinemanModifier("SOLD_OUT_TODAY")
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
