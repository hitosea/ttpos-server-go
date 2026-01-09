// Package lineman 提供 LINE MAN 平台集成服务
package lineman

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/test/gtest"
)

// TestDataMapper_BuildNameTranslation 测试名称翻译构建
func TestDataMapper_BuildNameTranslation(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		mapper := NewDataMapper(context.Background())

		// 测试有完整翻译的情况
		trans := mapper.buildNameTranslation("中文", "ภาษาไทย", "English")
		t.Assert(trans.Thai, "ภาษาไทย")
		t.Assert(trans.English, "English")

		// 测试无翻译的情况（降级到中文）
		trans = mapper.buildNameTranslation("中文", "", "")
		t.Assert(trans.Thai, "中文")
		t.Assert(trans.English, "中文")

		// 测试部分翻译（泰语为空）
		trans = mapper.buildNameTranslation("中文", "", "English")
		t.Assert(trans.Thai, "中文")
		t.Assert(trans.English, "中文")

		// 测试部分翻译（英语为空）
		trans = mapper.buildNameTranslation("中文", "ภาษาไทย", "")
		t.Assert(trans.Thai, "中文")
		t.Assert(trans.English, "中文")

		// 测试去除空格
		trans = mapper.buildNameTranslation(" 中文 ", " ภาษาไทย ", " English ")
		t.Assert(trans.Thai, "ภาษาไทย")
		t.Assert(trans.English, "English")
	})
}

// TestDataMapper_BuildDescTranslation 测试描述翻译构建
func TestDataMapper_BuildDescTranslation(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		mapper := NewDataMapper(context.Background())

		// 测试有完整翻译的情况
		trans := mapper.buildDescTranslation("描述", "คำอธิบาย", "Description")
		t.Assert(trans.Thai, "คำอธิบาย")
		t.Assert(trans.English, "Description")

		// 测试无翻译的情况（降级到中文）
		trans = mapper.buildDescTranslation("描述", "", "")
		t.Assert(trans.Thai, "描述")
		t.Assert(trans.English, "描述")

		// 测试空描述
		trans = mapper.buildDescTranslation("", "", "")
		t.Assert(trans.Thai, "")
		t.Assert(trans.English, "")
	})
}
