package tasks

import (
	"testing"

	"github.com/robfig/cron/v3"
)

func TestDelPrintCronExpression(t *testing.T) {
	c := cron.New(cron.WithSeconds()) // 和生产代码一样，启用6位秒级模式

	// 当前生产使用的表达式（5位）
	_, err := c.AddFunc("0 6 * * *", func() {})
	if err != nil {
		t.Logf("❌ 5位表达式 \"0 6 * * *\" 注册失败: %v", err)
	} else {
		t.Log("✅ 5位表达式 \"0 6 * * *\" 注册成功")
	}

	// 修复后的表达式（6位，每天06:00:00）
	_, err = c.AddFunc("0 0 6 * * *", func() {})
	if err != nil {
		t.Logf("❌ 6位表达式 \"0 0 6 * * *\" 注册失败: %v", err)
	} else {
		t.Log("✅ 6位表达式 \"0 0 6 * * *\" 注册成功")
	}

	// 对比：其他生产任务的表达式（都是6位，应该都成功）
	otherExprs := map[string]string{
		"0 0 * * * *":    "每小时整点",
		"0 */15 * * * *": "每15分钟",
		"15 * * * * *":   "每分钟第15秒",
		"30 * * * * *":   "每分钟第30秒",
		"0 0 0 * * *":    "每天凌晨0点",
	}
	for expr, desc := range otherExprs {
		_, err = c.AddFunc(expr, func() {})
		if err != nil {
			t.Logf("❌ \"%s\" (%s) 注册失败: %v", expr, desc, err)
		} else {
			t.Logf("✅ \"%s\" (%s) 注册成功", expr, desc)
		}
	}
}
