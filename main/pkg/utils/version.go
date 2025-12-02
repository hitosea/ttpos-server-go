package utils

import (
	"github.com/Masterminds/semver"
)

// VersionOperator 版本比较操作符类型
type VersionOperator string

const (
	VersionLT  VersionOperator = "<"
	VersionGT  VersionOperator = ">"
	VersionEQ  VersionOperator = "="
	VersionGTE VersionOperator = ">="
)

// CompareVersion 比较两个版本号
// v1: 第一个版本号字符串
// op: 比较操作符
// v2: 第二个版本号字符串
// 返回: 如果比较结果为真则返回 true，否则返回 false
// 如果版本号解析失败，返回 false
func CompareVersion(v1 string, op VersionOperator, v2 string) bool {
	version1, err := semver.NewVersion(v1)
	if err != nil {
		return false
	}
	version2, err := semver.NewVersion(v2)
	if err != nil {
		return false
	}
	switch op {
	case VersionGTE:
		return version1.GreaterThan(version2) || version1.Equal(version2)
	case VersionGT:
		return version1.GreaterThan(version2)
	case VersionLT:
		return version1.LessThan(version2)
	case VersionEQ:
		return version1.Equal(version2)
	}
	return false
}
