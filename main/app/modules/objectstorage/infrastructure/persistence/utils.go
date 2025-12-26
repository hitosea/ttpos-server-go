package persistence

import (
	goCtx "context"
	"fmt"
	"sort"
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
	return BuildKeyWithCompanyUuid(companyUuid, objectType, objectUuid)
}

func BuildKeyWithCompanyUuid(companyUuid uint64, objectType string, objectUuid uint64) string {
	return fmt.Sprintf("%s:%d:%s:%d", SystemPrefix, companyUuid, objectType, objectUuid)
}

// BuildAuthStaffKey 构建员工信息缓存 key
// Key 格式：{system_prefix}:{company_uuid}:staff:{staff_uuid}
func BuildAuthStaffKey(ctx goCtx.Context, staffUuid uint64) string {
	return BuildKey(ctx, "staff", staffUuid)
}

// BuildAuthDeskKey 构建桌台信息缓存 key
// Key 格式：{system_prefix}:{company_uuid}:desk:{device_uuid}
func BuildAuthDeskKey(ctx goCtx.Context, deviceUuid uint64) string {
	return BuildKey(ctx, "desk", deviceUuid)
}

// BuildApiPermissionKey 构建 API 权限缓存 key
// Key 格式：{system_prefix}:{company_uuid}:api_permission:{staff_uuid}
func BuildApiPermissionKey(companyUuid, staffUuid uint64) string {
	return BuildKeyWithCompanyUuid(companyUuid, "api_permission", staffUuid)
}

// BuildProductListCacheKey 构建商品列表缓存 key
// Key 格式：{system_prefix}:{company_uuid}:product_list:{source}:{pageNo}:{pageSize}:{isMember}:{recommendUuids}
// 参数：
//   - companyUuid: 门店 UUID
//   - source: 来源（如 cashier, assistant 等）
//   - pageNo: 页码
//   - pageSize: 每页数量
//   - isMember: 是否会员
//   - recommendUuids: 推荐商品 UUID 列表（会自动排序以确保相同 UUIDs 但顺序不同时 key 一致）
func BuildProductListCacheKey(companyUuid uint64, source string, pageNo, pageSize int, isMember bool, recommendUuids []uint64) string {
	// 对推荐商品 UUIDs 排序，确保相同 UUIDs 但顺序不同时 key 一致
	sortedUuids := make([]uint64, len(recommendUuids))
	copy(sortedUuids, recommendUuids)
	sort.Slice(sortedUuids, func(i, j int) bool {
		return sortedUuids[i] < sortedUuids[j]
	})

	// 构建 key：{system_prefix}:{company_uuid}:product_list:{source}:{pageNo}:{pageSize}:{isMember}:{recommendUuids}
	keyParts := []string{
		SystemPrefix,
		fmt.Sprintf("%d", companyUuid),
		"product_list",
		source,
		fmt.Sprintf("%d", pageNo),
		fmt.Sprintf("%d", pageSize),
		fmt.Sprintf("%v", isMember),
	}

	// 添加推荐商品 UUIDs
	if len(sortedUuids) > 0 {
		uuidStrs := make([]string, len(sortedUuids))
		for i, uuid := range sortedUuids {
			uuidStrs[i] = fmt.Sprintf("%d", uuid)
		}
		keyParts = append(keyParts, strings.Join(uuidStrs, ","))
	}

	return strings.Join(keyParts, ":")
}
