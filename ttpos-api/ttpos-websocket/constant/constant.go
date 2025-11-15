// Package constant 定义 ttpos-websocket 服务相关的常量
package constant

// Topic 常量定义
const (
	// 用于收银机上报局域网内的打印机信息
	TopicLanPrinterReport = "lan_printer_report"
)

// GetAllTopics 获取所有 ttpos-websocket 的 Topic
func GetAllTopics() []string {
	return []string{
		TopicLanPrinterReport,
	}
}

// IsValidTopic 检查 Topic 是否有效
func IsValidTopic(topic string) bool {
	for _, t := range GetAllTopics() {
		if t == topic {
			return true
		}
	}
	return false
}
