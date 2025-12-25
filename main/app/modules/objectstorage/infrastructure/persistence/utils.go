package persistence

import (
	goCtx "context"
	"fmt"
	"strings"

	"ttpos-server-go/pkg/context"
)

const (
	// SystemPrefix 系统前缀，用于区分不同系统的缓存 key
	SystemPrefix = "ttpos4"
)

// BuildKey 构建 key 的辅助方法（自动从 context 提取 company UUID）
// Key 格式：{system_prefix}:{company_uuid}:{object_type}:{object_uuid}
func BuildKey(ctx goCtx.Context, objectType string, objectUuid uint64) string {
	cctx := ctx.(context.Context)
	companyUuid := cctx.GetCompanyUuid()
	return fmt.Sprintf("%s:%d:%s:%d", SystemPrefix, companyUuid, objectType, objectUuid)
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
// Key 格式：{system_prefix}:{company_uuid}:{object_type}:{object_uuid}
func extractObjectTypeFromKey(key string) string {
	parts := strings.Split(key, ":")
	if len(parts) >= 3 {
		return parts[2] // objectType 在索引 2 的位置
	}
	return ""
}
