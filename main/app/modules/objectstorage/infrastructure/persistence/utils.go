package persistence

import (
	"context"
	"fmt"
	"strings"

	bizctx "ttpos-server-go/pkg/context"
)

// BuildKey 构建 key 的辅助方法（自动从 context 提取 company UUID）
// Key 格式：{company_uuid}:{object_type}:{object_uuid}
func BuildKey(ctx context.Context, objectType string, objectUuid uint64) string {
	companyUuid := bizctx.GetCompanyUuid(ctx)
	return fmt.Sprintf("%d:%s:%d", companyUuid, objectType, objectUuid)
}

// deduplicate 对字符串切片去重
func deduplicate(keys []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(keys))

	for _, key := range keys {
		if !seen[key] {
			seen[key] = true
			result = append(result, key)
		}
	}

	return result
}

// extractObjectTypeFromKey 从 key 中提取 objectType
// Key 格式：{company_uuid}:{object_type}:{object_uuid}
func extractObjectTypeFromKey(key string) string {
	parts := strings.Split(key, ":")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

