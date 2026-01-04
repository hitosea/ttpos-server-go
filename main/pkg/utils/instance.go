package utils

import (
	"os"
)

// GetInstanceID 获取实例ID
// 优先使用环境变量 INSTANCE_ID，否则使用 hostname
func GetInstanceID() string {
	// 1. 优先使用环境变量 INSTANCE_ID
	if instanceID := os.Getenv("INSTANCE_ID"); instanceID != "" {
		return instanceID
	}

	// 2. 使用 hostname
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}

	// 3. 使用默认值
	return "unknown"
}
