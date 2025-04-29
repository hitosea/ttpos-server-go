package utils

import (
	"net"
)

// IsInternalIP 检查IP是否为内网IP
func IsInternalIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// 检查是否为本地回环地址
	if ip.IsLoopback() {
		return true
	}

	// 内网IP地址范围:
	// A类: 10.0.0.0--10.255.255.255
	// B类: 172.16.0.0--172.31.255.255
	// C类: 192.168.0.0--192.168.255.255
	if ip4 := ip.To4(); ip4 != nil {
		// A类
		if ip4[0] == 10 {
			return true
		}
		// B类
		if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
			return true
		}
		// C类
		if ip4[0] == 192 && ip4[1] == 168 {
			return true
		}
	}

	return false
}
